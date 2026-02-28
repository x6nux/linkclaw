package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
)

// MemoryService 记忆业务逻辑
type MemoryService struct {
	memoryRepo   repository.MemoryRepo
	companyRepo  repository.CompanyRepo
	embeddingCli *EmbeddingClient
}

// NewMemoryService 创建 MemoryService
func NewMemoryService(memoryRepo repository.MemoryRepo, companyRepo repository.CompanyRepo, embeddingCli *EmbeddingClient) *MemoryService {
	return &MemoryService{
		memoryRepo:   memoryRepo,
		companyRepo:  companyRepo,
		embeddingCli: embeddingCli,
	}
}

// CreateInput 创建记忆输入
type CreateMemoryInput struct {
	CompanyID  string
	AgentID    string
	Content    string
	Category   string
	Tags       []string
	Importance domain.MemoryImportance
	Source     domain.MemorySource
}

// Create 创建记忆
func (s *MemoryService) Create(ctx context.Context, in CreateMemoryInput) (*domain.Memory, error) {
	if in.Content == "" {
		return nil, fmt.Errorf("content is required")
	}
	if in.Category == "" {
		in.Category = "general"
	}
	if in.Source == "" {
		in.Source = domain.SourceManual
	}
	if in.Tags == nil {
		in.Tags = []string{}
	}

	m := &domain.Memory{
		ID:         uuid.New().String(),
		CompanyID:  in.CompanyID,
		AgentID:    in.AgentID,
		Content:    in.Content,
		Category:   in.Category,
		Tags:       in.Tags,
		Importance: in.Importance,
		Source:     in.Source,
	}
	if err := s.memoryRepo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// GetByID 获取记忆详情
func (s *MemoryService) GetByID(ctx context.Context, id string) (*domain.Memory, error) {
	return s.memoryRepo.GetByID(ctx, id)
}

// UpdateMemoryInput 更新记忆输入
type UpdateMemoryInput struct {
	Content    string
	Category   string
	Tags       []string
	Importance domain.MemoryImportance
}

// Update 更新记忆
func (s *MemoryService) Update(ctx context.Context, id string, in UpdateMemoryInput) (*domain.Memory, error) {
	m, err := s.memoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, fmt.Errorf("memory not found")
	}
	m.Content = in.Content
	m.Category = in.Category
	m.Tags = in.Tags
	m.Importance = in.Importance
	if err := s.memoryRepo.Update(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// Delete 删除记忆
func (s *MemoryService) Delete(ctx context.Context, id string) error {
	return s.memoryRepo.Delete(ctx, id)
}

// List 列出记忆
func (s *MemoryService) List(ctx context.Context, q repository.MemoryQuery) ([]*domain.Memory, int, error) {
	return s.memoryRepo.List(ctx, q)
}

// SemanticSearch 语义搜索记忆
func (s *MemoryService) SemanticSearch(ctx context.Context, companyID, agentID, query string, limit int) ([]*domain.Memory, error) {
	company, err := s.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("get company: %w", err)
	}
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}

	vec, err := s.embeddingCli.Generate(ctx, company.EmbeddingBaseURL, company.EmbeddingModel, company.EmbeddingApiKey, query)
	if err != nil {
		return nil, fmt.Errorf("generate query embedding: %w", err)
	}

	mems, err := s.memoryRepo.SemanticSearch(ctx, companyID, agentID, vec, limit, 4)
	if err != nil {
		return nil, err
	}

	// 更新访问次数
	if len(mems) > 0 {
		ids := make([]string, len(mems))
		for i, m := range mems {
			ids[i] = m.ID
		}
		_ = s.memoryRepo.IncrementAccess(ctx, ids)
	}

	return mems, nil
}

// BatchDelete 批量删除
func (s *MemoryService) BatchDelete(ctx context.Context, ids []string) error {
	return s.memoryRepo.BatchDelete(ctx, ids)
}
