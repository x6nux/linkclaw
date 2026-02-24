package repository

import (
	"context"

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
	Delete(ctx context.Context, id string) error
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
	GetByID(ctx context.Context, id string) (*domain.Task, error)
	List(ctx context.Context, q TaskQuery) ([]*domain.Task, int, error)
	UpdateStatus(ctx context.Context, id string, status domain.TaskStatus, result, failReason *string) error
	UpdateAssignee(ctx context.Context, id, assigneeID string, status domain.TaskStatus) error
	Delete(ctx context.Context, id string) error
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
