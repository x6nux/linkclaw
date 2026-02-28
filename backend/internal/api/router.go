package api

import (
	"github.com/gin-gonic/gin"

	"github.com/linkclaw/backend/internal/config"
	"github.com/linkclaw/backend/internal/llm"
	"github.com/linkclaw/backend/internal/mcp"
	"github.com/linkclaw/backend/internal/repository"
	"github.com/linkclaw/backend/internal/service"
)

// RegisterRoutes 注册所有 HTTP 路由
func RegisterRoutes(
	r *gin.Engine,
	agentRepo repository.AgentRepo,
	jwtSecret string,
	jwtExpiry int,
	agentSvc *service.AgentService,
	taskSvc *service.TaskService,
	messageSvc *service.MessageService,
	knowledgeSvc *service.KnowledgeService,
	memorySvc *service.MemoryService,
	indexingSvc *service.IndexingService,
	obsSvc *service.ObservabilityService,
	obsRepo repository.ObservabilityRepo,
	auditRepo repository.AuditRepo,
	qualitySvc *service.QualityScoringService,
	mcpServer *mcp.Server,
	llmHandler *llm.Handler,
	agentCfg *config.AgentConfig,
	companyRepo repository.CompanyRepo,
	deploySvc *service.DeploymentService,
	resetSecret string,
	promptSvc *service.PromptService,
	orgSvc *service.OrganizationService,
	webhookSvc *service.WebhookService,
	personaSvc *service.PersonaOptimizerService,
) {
	// MCP 端点（自带 Bearer 认证，不走统一 AuthMiddleware）
	mcpServer.RegisterRoutes(r)

	// Public 路由（不走 AuthMiddleware）
	pub := r.Group("/api/v1")
	authH := &authHandler{agentRepo: agentRepo, companyRepo: companyRepo, jwtSecret: jwtSecret, jwtExpiry: jwtExpiry, resetSecret: resetSecret}
	pub.POST("/auth/login", authH.login)
	pub.POST("/auth/reset-password", authH.resetPassword)

	setupH := &setupHandler{companyRepo: companyRepo, agentRepo: agentRepo, agentSvc: agentSvc, jwtSecret: jwtSecret, jwtExpiry: jwtExpiry}
	pub.GET("/setup/status", setupH.status)
	pub.POST("/setup", setupH.initialize)

	// REST API（需要认证）
	auth := r.Group(
		"/api/v1",
		AuthMiddleware(agentRepo, jwtSecret),
		AuditMiddleware(auditRepo),
		SensitiveAccessMiddleware(auditRepo),
	)
	auth.POST("/auth/logout", authH.logout)
	auth.POST("/auth/change-password", authH.changePassword)
	auditH := &auditHandler{auditRepo: auditRepo}
	auth.GET("/audit/logs", auditH.listAuditLogs)

	// Agent
	ah := &agentHandler{agentSvc: agentSvc}
	auth.GET("/agents", ah.list)
	auth.POST("/agents", ah.create)
	auth.GET("/agents/:id", ah.get)
	auth.PATCH("/agents/:id", ah.update)
	auth.DELETE("/agents/:id", ChairmanOnly(), ah.delete)

	// Agent 部署
	dh := &deploymentHandler{deployService: deploySvc, agentService: agentSvc}
	auth.POST("/agents/:id/deploy", ChairmanOnly(), dh.deploy)
	auth.GET("/agents/:id/deployment", dh.getDeployment)
	auth.DELETE("/agents/:id/deployment", ChairmanOnly(), dh.stopDeployment)
	auth.POST("/agents/:id/deployment/rebuild", ChairmanOnly(), dh.rebuild)

	// Task
	th := &taskHandler{taskSvc: taskSvc}
	auth.GET("/tasks", th.list)
	auth.POST("/tasks", th.create)
	auth.GET("/tasks/:id", th.get)
	auth.DELETE("/tasks/:id", th.delete)
	auth.GET("/tasks/:id/detail", th.detail)
	auth.POST("/tasks/:id/comments", th.addComment)
	auth.DELETE("/tasks/:id/comments/:commentId", th.deleteComment)
	auth.POST("/tasks/:id/dependencies", th.addDependency)
	auth.DELETE("/tasks/:id/dependencies/:depId", th.removeDependency)
	auth.POST("/tasks/:id/watchers", th.addWatcher)
	auth.DELETE("/tasks/:id/watchers", th.removeWatcher)
	auth.PUT("/tasks/:id/tags", th.updateTags)

	// Message
	mh := &messageHandler{messageSvc: messageSvc}
	auth.GET("/messages", mh.list)
	auth.POST("/messages", mh.send)
	auth.GET("/channels", mh.listChannels)

	// Knowledge
	kh := &knowledgeHandler{knowledgeSvc: knowledgeSvc}
	auth.GET("/knowledge", kh.list)
	auth.GET("/knowledge/search", kh.search)
	auth.POST("/knowledge", kh.write)
	auth.GET("/knowledge/:id", kh.get)
	auth.PUT("/knowledge/:id", kh.write)
	auth.DELETE("/knowledge/:id", kh.delete)

	// Memory
	memH := &memoryHandler{memorySvc: memorySvc}
	auth.GET("/memories", memH.list)
	auth.GET("/memories/:id", memH.get)
	auth.POST("/memories", memH.create)
	auth.PUT("/memories/:id", memH.update)
	auth.DELETE("/memories/:id", memH.delete)
	auth.POST("/memories/search", memH.search)
	auth.POST("/memories/batch-delete", memH.batchDelete)

	// Webhook（Chairman only）
	wh := &webhookHandler{webhookSvc: webhookSvc}
	webhookAdmin := auth.Group("/webhooks", ChairmanOnly())
	webhookAdmin.POST("", wh.createWebhook)
	webhookAdmin.GET("", wh.listWebhooks)
	webhookAdmin.POST("/signing-keys", wh.createSigningKey)
	webhookAdmin.GET("/signing-keys", wh.listSigningKeys)
	webhookAdmin.DELETE("/signing-keys/:id", wh.deleteSigningKey)
	webhookAdmin.GET("/deliveries/:id", wh.getDeliveryStatus)
	webhookAdmin.GET("/:id", wh.getWebhook)
	webhookAdmin.PUT("/:id", wh.updateWebhook)
	webhookAdmin.DELETE("/:id", wh.deleteWebhook)

	// Organization（审批：全员可用；组织管理：仅董事长）
	orgH := &organizationHandler{orgSvc: orgSvc}
	orgRoutes := auth.Group("/organization")
	orgRoutes.GET("/approvals", orgH.listApprovals)
	orgRoutes.POST("/approvals", orgH.createApproval)

	orgAdmin := auth.Group("/organization", ChairmanOnly())
	orgAdmin.GET("/departments", orgH.listDepartments)
	orgAdmin.POST("/departments", orgH.createDepartment)
	orgAdmin.PUT("/departments/:id", orgH.updateDepartment)
	orgAdmin.DELETE("/departments/:id", orgH.deleteDepartment)
	orgAdmin.POST("/departments/:id/assign", orgH.assignAgent)
	orgAdmin.PUT("/agents/:id/manager", orgH.setManager)
	orgAdmin.GET("/chart", orgH.orgChart)
	orgAdmin.POST("/approvals/:id/approve", orgH.approveRequest)
	orgAdmin.POST("/approvals/:id/reject", orgH.rejectRequest)

	// Persona 优化（需要认证）
	personaH := &personaHandler{personaSvc: personaSvc}
	personaRoutes := auth.Group("/persona")
	personaRoutes.GET("/suggestions", personaH.listSuggestions)
	personaRoutes.POST("/suggestions/:id/apply", personaH.applySuggestion)
	personaRoutes.GET("/history/:agentId", personaH.getHistory)
	personaRoutes.POST("/ab-test", personaH.createABTest)
	personaRoutes.GET("/ab-tests", personaH.listABTests)

	// Prompt 分层提示词管理（Chairman only）
	promptH := &promptHandler{promptSvc: promptSvc, agentSvc: agentSvc}
	promptAdmin := auth.Group("/prompts", ChairmanOnly())
	promptAdmin.GET("", promptH.list)
	promptAdmin.PUT("/:type/:key", promptH.upsert)
	promptAdmin.DELETE("/:type/:key", promptH.remove)
	promptAdmin.GET("/preview/:agentId", promptH.preview)

	// 系统设置（Chairman only）
	settingsH := &settingsHandler{companyRepo: companyRepo}
	settingsAdmin := auth.Group("/settings", ChairmanOnly())
	settingsAdmin.GET("", settingsH.get)
	settingsAdmin.PUT("", settingsH.update)

	// Context Indexing（Chairman only）
	indexH := &indexingHandler{indexingSvc: indexingSvc}
	indexAdmin := auth.Group("/indexing", ChairmanOnly())
	indexAdmin.POST("/tasks", indexH.createTask)
	indexAdmin.GET("/tasks", indexH.listTasks)
	indexAdmin.GET("/tasks/:id", indexH.getStatus)
	indexAdmin.POST("/search", indexH.search)

	// Observability 管理（Chairman only）
	obsH := &observabilityHandler{obsSvc: obsSvc, obsRepo: obsRepo, qualitySvc: qualitySvc}
	obsAdmin := auth.Group("/observability", ChairmanOnly())
	obsAdmin.GET("/overview", obsH.overview)
	obsAdmin.GET("/traces", obsH.listTraces)
	obsAdmin.GET("/traces/:id", obsH.getTrace)
	obsAdmin.POST("/traces/:id/score", obsH.scoreTrace)
	obsAdmin.GET("/budget-policies", obsH.listBudgetPolicies)
	obsAdmin.POST("/budget-policies", obsH.createBudgetPolicy)
	obsAdmin.PUT("/budget-policies/:id", obsH.updateBudgetPolicy)
	obsAdmin.GET("/budget-alerts", obsH.listBudgetAlerts)
	obsAdmin.PATCH("/budget-alerts/:id", obsH.patchBudgetAlert)
	obsAdmin.GET("/error-policies", obsH.listErrorPolicies)
	obsAdmin.POST("/error-policies", obsH.createErrorPolicy)
	obsAdmin.GET("/quality-scores", obsH.listQualityScores)

	// LLM Gateway 管理 API（Chairman only）
	llmAdmin := auth.Group("/llm", ChairmanOnly())
	llmAdmin.GET("/providers", llmHandler.ListProviders)
	llmAdmin.POST("/providers", llmHandler.CreateProvider)
	llmAdmin.PUT("/providers/:id", llmHandler.UpdateProvider)
	llmAdmin.DELETE("/providers/:id", llmHandler.DeleteProvider)
	llmAdmin.GET("/stats", llmHandler.GetStats)

	// LLM 代理端点（所有认证用户/Agent 可调用）
	// 支持完整 Anthropic API（/v1/messages、/v1/messages/batches、/v1/messages/count_tokens 等）
	// 支持完整 OpenAI API（/v1/chat/completions、/v1/completions、/v1/embeddings、/v1/models 等）
	registerLLMProxy(
		r,
		llmHandler,
		AuthMiddleware(agentRepo, jwtSecret),
		AuditMiddleware(auditRepo),
		SensitiveAccessMiddleware(auditRepo),
	)

	// 跨公司 Partner API
	ph := &partnerHandler{messageSvc: messageSvc, companyRepo: companyRepo, agentCfg: agentCfg}
	r.GET("/api/v1/partner/info", ph.info)
	r.POST("/api/v1/partner/message", partnerAuthMiddleware(agentCfg), ph.receiveMessage)

	// 健康检查
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
}

// registerLLMProxy 注册 LLM 代理路由，同时支持 Anthropic 和 OpenAI 完整 API 路径
func registerLLMProxy(r *gin.Engine, h *llm.Handler, middlewares ...gin.HandlerFunc) {
	anthropic := append(append([]gin.HandlerFunc{}, middlewares...), h.ProxyAnthropic)
	openai := append(append([]gin.HandlerFunc{}, middlewares...), h.ProxyOpenAI)

	// ── Anthropic API ─────────────────────────────────────────
	// POST /v1/messages              — 创建消息
	// POST /v1/messages/count_tokens — 计算 token 数量
	// POST /v1/messages/batches      — 创建批处理
	// GET  /v1/messages/batches/:id  — 查询批处理
	r.POST("/v1/messages", anthropic...)
	r.POST("/v1/messages/*path", anthropic...)
	r.GET("/v1/messages/*path", anthropic...)

	// ── OpenAI API ────────────────────────────────────────────
	// POST /v1/chat/completions      — 聊天补全
	// POST /v1/completions           — 文本补全（legacy）
	// POST /v1/embeddings            — 向量嵌入
	// GET  /v1/models                — 模型列表
	// GET  /v1/models/:id            — 模型详情
	r.POST("/v1/chat/completions", openai...)
	r.POST("/v1/completions", openai...)
	r.POST("/v1/embeddings", openai...)
	r.GET("/v1/models", openai...)
	r.GET("/v1/models/*path", openai...)
}
