package repository

import (
	"context"
	"time"

	"github.com/linkclaw/backend/internal/domain"
)

type AgentRepo interface {
	Create(ctx context.Context, a *domain.Agent) error
	GetByID(ctx context.Context, id string) (*domain.Agent, error)
	GetByAPIKeyHash(ctx context.Context, hash string) (*domain.Agent, error)
	GetByCompany(ctx context.Context, companyID string) ([]*domain.Agent, error)
	GetByName(ctx context.Context, companyID, name string) (*domain.Agent, error)
	GetByHireRequestID(ctx context.Context, requestID string) (*domain.Agent, error)
	UpdateStatus(ctx context.Context, id string, status domain.AgentStatus) error
	UpdateLastSeen(ctx context.Context, id string) error
	UpdateName(ctx context.Context, id, name string) error
	UpdateModel(ctx context.Context, id, model string) error
	MarkInitialized(ctx context.Context, id string) error
	SetPasswordHash(ctx context.Context, id, hash string) error
	UpdatePersona(ctx context.Context, id, persona string) error
	UpdateAPIKey(ctx context.Context, id, hash, prefix string) error
	UpdateDepartment(ctx context.Context, id string, departmentID *string) error
	UpdateManager(ctx context.Context, id string, managerID *string) error
	ListByDepartment(ctx context.Context, companyID, departmentID string) ([]*domain.Agent, error)
	Delete(ctx context.Context, id string) error
}

type DepartmentRepo interface {
	Create(ctx context.Context, d *domain.Department) error
	GetByID(ctx context.Context, id string) (*domain.Department, error)
	GetBySlug(ctx context.Context, companyID, slug string) (*domain.Department, error)
	List(ctx context.Context, companyID string) ([]*domain.Department, error)
	Update(ctx context.Context, d *domain.Department) error
	Delete(ctx context.Context, id string) error
	AssignAgent(ctx context.Context, agentID, departmentID string) error
}

type ApprovalRepo interface {
	Create(ctx context.Context, r *domain.ApprovalRequest) error
	GetByID(ctx context.Context, id string) (*domain.ApprovalRequest, error)
	List(ctx context.Context, q ApprovalQuery) ([]*domain.ApprovalRequest, int, error)
	UpdateStatus(ctx context.Context, id string, status domain.ApprovalStatus, decisionReason string, decidedAt *time.Time) error
}

type CompanyRepo interface {
	Create(ctx context.Context, c *domain.Company) error
	GetByID(ctx context.Context, id string) (*domain.Company, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Company, error)
	FindFirst(ctx context.Context) (*domain.Company, error)
	UpdateSystemPrompt(ctx context.Context, id, prompt string) error
	UpdateSettings(ctx context.Context, id string, s *domain.CompanySettings) error
	CreateChannel(ctx context.Context, ch *domain.Channel) error
	GetChannels(ctx context.Context, companyID string) ([]*domain.Channel, error)
	GetChannelByName(ctx context.Context, companyID, name string) (*domain.Channel, error)
}

type TaskRepo interface {
	Create(ctx context.Context, t *domain.Task) error
	CreateAttachments(ctx context.Context, attachments []*domain.TaskAttachment) error
	GetByID(ctx context.Context, id string) (*domain.Task, error)
	List(ctx context.Context, q TaskQuery) ([]*domain.Task, int, error)
	UpdateStatus(ctx context.Context, id string, status domain.TaskStatus, result, failReason *string) error
	UpdateAssignee(ctx context.Context, id, assigneeID string, status domain.TaskStatus) error
	UpdateTags(ctx context.Context, id string, tags domain.StringList) error
	Delete(ctx context.Context, id, companyID string) error
}

type TaskQuery struct {
	CompanyID  string
	AssigneeID string
	Status     domain.TaskStatus
	Priority   domain.TaskPriority
	ParentID   *string // nil = 顶层任务，"" = 所有
	Limit      int
	Offset     int
}

type TaskCollabRepo interface {
	AddComment(ctx context.Context, c *domain.TaskComment) error
	ListComments(ctx context.Context, taskID string) ([]*domain.TaskComment, error)
	DeleteComment(ctx context.Context, id, agentID, companyID string) error
	AddDependency(ctx context.Context, d *domain.TaskDependency) error
	ListDependencies(ctx context.Context, taskID string) ([]*domain.TaskDependency, error)
	DeleteDependency(ctx context.Context, taskID, dependsOnID string) error
	AddWatcher(ctx context.Context, w *domain.TaskWatcher) error
	ListWatchers(ctx context.Context, taskID string) ([]*domain.TaskWatcher, error)
	RemoveWatcher(ctx context.Context, taskID, agentID string) error
}

type PersonaOptimizationRepo interface {
	CreateSuggestion(ctx context.Context, s *domain.PersonaOptimizationSuggestion) error
	GetSuggestions(ctx context.Context, companyID, agentID string, status domain.SuggestionStatus) ([]*domain.PersonaOptimizationSuggestion, error)
	UpdateSuggestionStatus(ctx context.Context, id string, status domain.SuggestionStatus) error
	CreateHistory(ctx context.Context, h *domain.PersonaHistory) error
	GetHistory(ctx context.Context, companyID, agentID string, limit int) ([]*domain.PersonaHistory, error)
	CreateABTest(ctx context.Context, t *domain.ABTestPersona) error
	GetABTest(ctx context.Context, id string) (*domain.ABTestPersona, error)
	ListABTests(ctx context.Context, companyID string) ([]*domain.ABTestPersona, error)
	UpdateABTest(ctx context.Context, t *domain.ABTestPersona) error
}

type ApprovalQuery struct {
	CompanyID   string
	RequesterID string
	ApproverID  string
	Status      domain.ApprovalStatus
	RequestType domain.ApprovalRequestType
	Limit       int
	Offset      int
}

type AuditQuery struct {
	CompanyID    string
	ResourceType string
	Action       string
	AgentID      string
	Limit        int
	Offset       int
}

type AuditRepo interface {
	CreateAuditLog(ctx context.Context, log *domain.AuditLog) error
	CreatePartnerAPICall(ctx context.Context, call *domain.PartnerAPICall) error
	CreateSensitiveAccessLog(ctx context.Context, log *domain.SensitiveAccessLog) error
	ListAuditLogs(ctx context.Context, q AuditQuery) ([]*domain.AuditLog, int, error)
}

type MessageRepo interface {
	Create(ctx context.Context, m *domain.Message) error
	ListByChannel(ctx context.Context, channelID string, limit int, beforeID string) ([]*domain.Message, error)
	ListDM(ctx context.Context, agentA, agentB string, limit int, beforeID string) ([]*domain.Message, error)
	MarkRead(ctx context.Context, agentID string, messageIDs []string) error
	ListUnreadForAgent(ctx context.Context, agentID, companyID string) ([]*domain.Message, error)
}

type DeploymentRepo interface {
	Create(ctx context.Context, d *domain.AgentDeployment) error
	GetByAgentID(ctx context.Context, agentID string) (*domain.AgentDeployment, error)
	UpdateStatus(ctx context.Context, id string, status domain.DeployStatus, errMsg string) error
	Delete(ctx context.Context, id string) error
}

type KnowledgeRepo interface {
	Create(ctx context.Context, d *domain.KnowledgeDoc) error
	GetByID(ctx context.Context, id string) (*domain.KnowledgeDoc, error)
	Update(ctx context.Context, d *domain.KnowledgeDoc) error
	Search(ctx context.Context, companyID, query string, limit int) ([]*domain.KnowledgeDoc, error)
	List(ctx context.Context, companyID string, limit, offset int) ([]*domain.KnowledgeDoc, int, error)
	Delete(ctx context.Context, id string) error
}

type CodeIndexRepo interface {
	CreateChunk(ctx context.Context, c *domain.CodeChunk) error
	GetChunksByFile(ctx context.Context, companyID, filePath string) ([]*domain.CodeChunk, error)
	DeleteByFile(ctx context.Context, companyID, filePath string) error
	DeleteAllChunks(ctx context.Context) error
	CreateIndexTask(ctx context.Context, t *domain.IndexTask) error
	GetIndexTask(ctx context.Context, id string) (*domain.IndexTask, error)
	UpdateIndexTask(ctx context.Context, t *domain.IndexTask) error
	ListIndexTasks(ctx context.Context, companyID string) ([]*domain.IndexTask, error)
}

type MemoryRepo interface {
	Create(ctx context.Context, m *domain.Memory) error
	GetByID(ctx context.Context, id string) (*domain.Memory, error)
	Update(ctx context.Context, m *domain.Memory) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, q MemoryQuery) ([]*domain.Memory, int, error)
	SemanticSearch(ctx context.Context, companyID, agentID string, embedding []float32, limit, minImportance int) ([]*domain.Memory, error)
	ListPendingEmbedding(ctx context.Context, limit int) ([]*domain.Memory, error)
	UpdateEmbedding(ctx context.Context, id string, embedding []float32) error
	IncrementAccess(ctx context.Context, ids []string) error
	BatchDelete(ctx context.Context, ids []string) error
}

type WebhookRepo interface {
	// Webhooks
	Create(ctx context.Context, w *domain.Webhook) error
	GetByID(ctx context.Context, id string) (*domain.Webhook, error)
	ListByCompany(ctx context.Context, companyID string) ([]*domain.Webhook, error)
	ListActiveByEvent(ctx context.Context, companyID string, eventType domain.WebhookEventType) ([]*domain.Webhook, error)
	Update(ctx context.Context, w *domain.Webhook) error
	Delete(ctx context.Context, id string) error

	// Signing Keys
	CreateSigningKey(ctx context.Context, k *domain.WebhookSigningKey) error
	GetSigningKeyByID(ctx context.Context, id string) (*domain.WebhookSigningKey, error)
	ListSigningKeys(ctx context.Context, companyID string) ([]*domain.WebhookSigningKey, error)
	DeleteSigningKey(ctx context.Context, id string) error

	// Deliveries
	CreateDelivery(ctx context.Context, d *domain.WebhookDelivery) error
	GetDeliveryByID(ctx context.Context, id string) (*domain.WebhookDelivery, error)
	ListDeliveries(ctx context.Context, webhookID string, limit, offset int) ([]*domain.WebhookDelivery, int, error)
	UpdateDelivery(ctx context.Context, d *domain.WebhookDelivery) error
	ListPendingDeliveries(ctx context.Context, limit int) ([]*domain.WebhookDelivery, error)
}

type PromptLayerRepo interface {
	Upsert(ctx context.Context, layer *domain.PromptLayer) error
	Delete(ctx context.Context, companyID, layerType, key string) error
	ListByCompany(ctx context.Context, companyID string) ([]*domain.PromptLayer, error)
	Get(ctx context.Context, companyID, layerType, key string) (*domain.PromptLayer, error)
}

type MemoryQuery struct {
	CompanyID  string
	AgentID    string
	Category   string
	Importance *int
	Limit      int
	Offset     int
	OrderBy    string // "created_at" | "importance" | "access_count"
}

type WebhookQuery struct {
	CompanyID string
	WebhookID string
	Status    domain.WebhookDeliveryStatus
	Limit     int
	Offset    int
}

type TraceOverview struct {
	Total                 int64   `gorm:"column:total"                   json:"total"`
	SuccessCount          int64   `gorm:"column:success_count"           json:"success_count"`
	AvgLatencyMs          float64 `gorm:"column:avg_latency_ms"          json:"avg_latency_ms"`
	TotalCostMicrodollars int64   `gorm:"column:total_cost_microdollars" json:"total_cost_microdollars"`
}

type TraceRunQuery struct {
	CompanyID  string
	Status     domain.TraceStatus
	SourceType domain.TraceSourceType
	Limit      int
	Offset     int
}

type BudgetAlertQuery struct {
	CompanyID string
	Status    domain.BudgetAlertStatus
	Level     domain.BudgetAlertLevel
	Limit     int
	Offset    int
}

type QualityScoreQuery struct {
	CompanyID string
	Limit     int
	Offset    int
}

type ObservabilityRepo interface {
	CreateTraceRun(ctx context.Context, t *domain.TraceRun) error
	GetTraceRunByID(ctx context.Context, id string) (*domain.TraceRun, error)
	ListTraceRuns(ctx context.Context, q TraceRunQuery) ([]*domain.TraceRun, int, error)
	UpdateTraceRunStatus(ctx context.Context, id string, status domain.TraceStatus, endedAt *time.Time, durationMs *int, errorMsg *string) error
	UpdateTraceRunTotals(ctx context.Context, id string, cost int64, inputTokens, outputTokens int) error

	CreateTraceSpan(ctx context.Context, s *domain.TraceSpan) error
	GetTraceSpanByID(ctx context.Context, id string) (*domain.TraceSpan, error)
	ListTraceSpansByTraceID(ctx context.Context, traceID string) ([]*domain.TraceSpan, error)
	UpdateTraceSpan(ctx context.Context, id string, status domain.TraceStatus, endedAt *time.Time, durationMs *int, inputTokens, outputTokens *int, cost *int64, errorMsg *string) error

	CreateTraceReplay(ctx context.Context, r *domain.TraceReplay) error
	GetTraceReplayBySpanID(ctx context.Context, spanID string) (*domain.TraceReplay, error)

	CreateBudgetPolicy(ctx context.Context, p *domain.LLMBudgetPolicy) error
	UpdateBudgetPolicy(ctx context.Context, p *domain.LLMBudgetPolicy) error
	GetBudgetPolicyByID(ctx context.Context, id string) (*domain.LLMBudgetPolicy, error)
	ListBudgetPolicies(ctx context.Context, companyID string) ([]*domain.LLMBudgetPolicy, error)
	ListActiveBudgetPolicies(ctx context.Context, companyID string) ([]*domain.LLMBudgetPolicy, error)

	CreateBudgetAlert(ctx context.Context, a *domain.LLMBudgetAlert) error
	UpdateBudgetAlert(ctx context.Context, id string, status domain.BudgetAlertStatus) error
	ListBudgetAlerts(ctx context.Context, q BudgetAlertQuery) ([]*domain.LLMBudgetAlert, error)

	CreateErrorAlertPolicy(ctx context.Context, p *domain.LLMErrorAlertPolicy) error
	UpdateErrorAlertPolicy(ctx context.Context, p *domain.LLMErrorAlertPolicy) error
	ListErrorAlertPolicies(ctx context.Context, companyID string) ([]*domain.LLMErrorAlertPolicy, error)

	CreateQualityScore(ctx context.Context, s *domain.ConversationQualityScore) error
	ListQualityScores(ctx context.Context, q QualityScoreQuery) ([]*domain.ConversationQualityScore, error)
	GetQualityScoreByTraceID(ctx context.Context, traceID string) (*domain.ConversationQualityScore, error)

	GetTraceOverview(ctx context.Context, companyID string) (*TraceOverview, error)
}
