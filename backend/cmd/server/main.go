package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/linkclaw/backend/internal/api"
	"github.com/linkclaw/backend/internal/config"
	"github.com/linkclaw/backend/internal/db"
	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/llm"
	"github.com/linkclaw/backend/internal/mcp"
	"github.com/linkclaw/backend/internal/repository"
	"github.com/linkclaw/backend/internal/service"
	webhooksub "github.com/linkclaw/backend/internal/webhook"
	"github.com/linkclaw/backend/internal/ws"
)

func main() {
	// 启动超时保护（防止无限阻塞）
	startupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := config.Load()

	// 验证环境变量配置
	if err := cfg.Validate(); err != nil {
		log.Fatalf("配置验证失败: %v", err)
	}

	gin.SetMode(cfg.Server.Mode)

	// 数据库连接（带超时）- NewPostgres 内部已包含 Ping 验证
	pg, err := db.NewPostgres(cfg.Database.DSN, cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}

	// 自动迁移（带超时）
	if err := db.RunMigrations(pg); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	// Redis 连接（带超时和优雅降级）
	rdb, err := db.NewRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Printf("redis unavailable (non-fatal): %v", err)
	} else {
		// 验证 Redis 连接
		if err := rdb.Ping(startupCtx).Err(); err != nil {
			log.Printf("redis ping failed (non-fatal): %v", err)
			rdb = nil
		}
	}

	// Repositories
	agentRepo := repository.NewAgentRepo(pg)
	companyRepo := repository.NewCompanyRepo(pg)
	taskRepo := repository.NewTaskRepo(pg)
	messageRepo := repository.NewMessageRepo(pg)
	knowledgeRepo := repository.NewKnowledgeRepo(pg)
	codeIndexRepo := repository.NewCodeIndexRepo(pg)
	memoryRepo := repository.NewMemoryRepo(pg)
	deployRepo := repository.NewDeploymentRepo(pg)
	deptRepo := repository.NewDepartmentRepo(pg)
	approvalRepo := repository.NewApprovalRepo(pg)
	obsRepo := repository.NewObservabilityRepo(pg)
	auditRepo := repository.NewAuditRepo(pg)
	collabRepo := repository.NewTaskCollabRepo(pg)
	personaRepo := repository.NewPersonaOptimizationRepo(pg)
	webhookRepo := repository.NewWebhookRepo(pg)

	// Services
	agentSvc := service.NewAgentService(agentRepo, companyRepo, deployRepo, taskRepo)
	taskSvc := service.NewTaskService(taskRepo, collabRepo, messageRepo, companyRepo)
	messageSvc := service.NewMessageService(messageRepo, companyRepo)
	knowledgeSvc := service.NewKnowledgeService(knowledgeRepo)
	deploySvc := service.NewDeploymentService(deployRepo, agentRepo, companyRepo)
	obsSvc := service.NewObservabilityService(obsRepo)
	qualitySvc := service.NewQualityScoringService(obsRepo)
	orgSvc := service.NewOrganizationService(deptRepo, agentRepo, approvalRepo)
	personaSvc := service.NewPersonaOptimizerService(personaRepo, agentRepo, taskRepo)
	webhookSvc := service.NewWebhookService(webhookRepo, 5*time.Second)
	go webhookSvc.ProcessDeliveryQueue(context.Background())

	webhookSubscriber := webhooksub.NewSubscriber(webhookSvc)
	webhookSubscriber.Start()

	budgetWatcher := service.NewBudgetWatcher(obsRepo, pg)
	budgetWatcher.Start()
	errorWatcher := service.NewErrorAlertWatcher(obsRepo, pg)
	errorWatcher.Start()

	// WebSocket Hub（实时推送）
	wsHub := ws.NewHub()
	go wsHub.Run()

	// LLM Gateway
	if cfg.LLM.EncryptKey == "" {
		log.Println("警告：LLM_ENCRYPT_KEY 未配置，LLM Gateway 功能不可用")
	}
	llmRepo := llm.NewRepository(pg)
	llmRouter := llm.NewRouter(llmRepo, cfg.LLM.EncryptKey)
	llmProxy := llm.NewProxyService(llmRepo, llmRouter, cfg.LLM.EncryptKey)
	llmHandler := llm.NewHandler(llmRepo, llmProxy, llmRouter, cfg.LLM.EncryptKey)

	// Embedding + Memory
	embeddingCli := service.NewEmbeddingClient(llmRouter)
	memorySvc := service.NewMemoryService(memoryRepo, embeddingCli)
	qdrantCfg := service.QdrantConfig{
		BaseURL: os.Getenv("QDRANT_URL"),
		APIKey:  os.Getenv("QDRANT_API_KEY"),
	}
	indexingSvc, err := service.NewIndexingService(
		codeIndexRepo,
		embeddingCli,
		qdrantCfg,
		2000,
		200,
	)
	if err != nil {
		log.Fatalf("indexing service: %v", err)
	}
	embeddingWorker := service.NewEmbeddingWorker(memoryRepo, embeddingCli)
	go embeddingWorker.Start(context.Background())

	// Prompt Service
	promptLayerRepo := repository.NewPromptLayerRepo(pg)
	promptSvc := service.NewPromptService(promptLayerRepo, companyRepo, agentRepo)

	// MCP Server
	mcpHandler := mcp.NewHandler(agentSvc, taskSvc, messageSvc, knowledgeSvc, memorySvc, indexingSvc,
		companyRepo, deploySvc, llmRepo, promptSvc, obsSvc, obsRepo, orgSvc)
	mcpServer := mcp.NewServer(agentRepo, mcpHandler, rdb)

	// HTTP Server
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// WebSocket 端点（前端，使用 JWT 或 API Key 认证）
	r.GET("/api/v1/messages/ws", func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		agent, err := validateWSToken(token, cfg.JWT.Secret, agentRepo)
		if err != nil || agent == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		ws.Upgrade(c, wsHub, agent, agentRepo, messageSvc)
	})

	// Agent WebSocket 端点（Agent 容器专用，使用 API Key 认证）
	r.GET("/api/v1/agents/me/ws", func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		agent, err := validateWSToken(token, cfg.JWT.Secret, agentRepo)
		if err != nil || agent == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		ws.UpgradeAgent(c, agent, agentRepo, messageSvc)
	})

	api.RegisterRoutes(r, agentRepo, cfg.JWT.Secret, cfg.JWT.Expiry,
		agentSvc, taskSvc, messageSvc, knowledgeSvc, memorySvc, indexingSvc,
		obsSvc, obsRepo, auditRepo, qualitySvc, mcpServer, llmHandler, &cfg.Agent, companyRepo,
		deploySvc, cfg.ResetSecret, promptSvc, orgSvc, webhookSvc, personaSvc)

	log.Printf("LinkClaw server starting on :%s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// validateWSToken 支持 JWT token 和 API Key 两种方式验证 WS 连接身份
func validateWSToken(tokenStr, secret string, agentRepo repository.AgentRepo) (*domain.Agent, error) {
	parsed, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !parsed.Valid {
		// fallback：尝试 API Key
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(tokenStr)))
		agent, err := agentRepo.GetByAPIKeyHash(context.Background(), hash)
		return agent, err
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	agentID, _ := claims["sub"].(string)
	if agentID == "" {
		return nil, fmt.Errorf("no sub claim")
	}
	return agentRepo.GetByID(context.Background(), agentID)
}
