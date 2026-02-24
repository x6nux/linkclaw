package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"

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
	"github.com/linkclaw/backend/internal/ws"
)

func main() {
	cfg := config.Load()
	gin.SetMode(cfg.Server.Mode)

	// 数据库连接
	pg, err := db.NewPostgres(cfg.Database.DSN, cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}

	// 自动迁移
	if err := db.RunMigrations(pg); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	rdb, err := db.NewRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Printf("redis unavailable (non-fatal): %v", err)
	}

	// Repositories
	agentRepo     := repository.NewAgentRepo(pg)
	companyRepo   := repository.NewCompanyRepo(pg)
	taskRepo      := repository.NewTaskRepo(pg)
	messageRepo   := repository.NewMessageRepo(pg)
	knowledgeRepo := repository.NewKnowledgeRepo(pg)
	memoryRepo    := repository.NewMemoryRepo(pg)
	deployRepo    := repository.NewDeploymentRepo(pg)

	// Services
	agentSvc     := service.NewAgentService(agentRepo, companyRepo, deployRepo, taskRepo)
	taskSvc      := service.NewTaskService(taskRepo, messageRepo, companyRepo)
	messageSvc   := service.NewMessageService(messageRepo, companyRepo)
	knowledgeSvc := service.NewKnowledgeService(knowledgeRepo)
	deploySvc    := service.NewDeploymentService(deployRepo, agentRepo, companyRepo)

	// WebSocket Hub（实时推送）
	wsHub := ws.NewHub()
	go wsHub.Run()

	// LLM Gateway
	if cfg.LLM.EncryptKey == "" {
		log.Println("警告：LLM_ENCRYPT_KEY 未配置，LLM Gateway 功能不可用")
	}
	llmRepo    := llm.NewRepository(pg)
	llmRouter  := llm.NewRouter(llmRepo, cfg.LLM.EncryptKey)
	llmProxy   := llm.NewProxyService(llmRepo, llmRouter, cfg.LLM.EncryptKey)
	llmHandler := llm.NewHandler(llmRepo, llmProxy, llmRouter, cfg.LLM.EncryptKey)

	// Embedding + Memory
	embeddingCli    := service.NewEmbeddingClient(llmRouter)
	memorySvc       := service.NewMemoryService(memoryRepo, embeddingCli)
	embeddingWorker := service.NewEmbeddingWorker(memoryRepo, embeddingCli)
	go embeddingWorker.Start(context.Background())

	// Prompt Service
	promptLayerRepo := repository.NewPromptLayerRepo(pg)
	promptSvc := service.NewPromptService(promptLayerRepo, companyRepo, agentRepo)

	// MCP Server
	mcpHandler := mcp.NewHandler(agentSvc, taskSvc, messageSvc, knowledgeSvc, memorySvc, companyRepo, deploySvc, llmRepo, promptSvc)
	mcpServer  := mcp.NewServer(agentRepo, mcpHandler, rdb)

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
		agentSvc, taskSvc, messageSvc, knowledgeSvc, memorySvc,
		mcpServer, llmHandler, &cfg.Agent, companyRepo,
		deploySvc, cfg.ResetSecret, promptSvc)

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
