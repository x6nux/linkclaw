package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
)

// PromptService 分层提示词管理
type PromptService struct {
	repo        repository.PromptLayerRepo
	companyRepo repository.CompanyRepo
	agentRepo   repository.AgentRepo
}

func NewPromptService(repo repository.PromptLayerRepo, companyRepo repository.CompanyRepo, agentRepo repository.AgentRepo) *PromptService {
	return &PromptService{repo: repo, companyRepo: companyRepo, agentRepo: agentRepo}
}

// PromptListResult 列出所有提示词层
type PromptListResult struct {
	Global      string              `json:"global"`
	Departments map[string]string   `json:"departments"`
	Positions   map[string]string   `json:"positions"`
	Agents      []PromptAgentBrief  `json:"agents"`
}

// PromptAgentBrief Agent 简要信息，用于前端导航
type PromptAgentBrief struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Position string `json:"position"`
	Persona  string `json:"persona"`
}

// ListAll 列出公司所有提示词层
func (s *PromptService) ListAll(ctx context.Context, companyID string) (*PromptListResult, error) {
	company, err := s.companyRepo.GetByID(ctx, companyID)
	if err != nil || company == nil {
		return nil, fmt.Errorf("公司不存在")
	}

	layers, err := s.repo.ListByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	result := &PromptListResult{
		Global:      company.SystemPrompt,
		Departments: make(map[string]string),
		Positions:   make(map[string]string),
	}

	for _, l := range layers {
		switch l.Type {
		case domain.PromptDepartment:
			result.Departments[l.Key] = l.Content
		case domain.PromptPosition:
			result.Positions[l.Key] = l.Content
		}
	}

	// 获取所有非 Human Agent
	agents, _ := s.agentRepo.GetByCompany(ctx, companyID)
	for _, a := range agents {
		if a.IsHuman {
			continue
		}
		result.Agents = append(result.Agents, PromptAgentBrief{
			ID:       a.ID,
			Name:     a.Name,
			Position: string(a.Position),
			Persona:  a.Persona,
		})
	}

	return result, nil
}

// Upsert 创建或更新提示词层
func (s *PromptService) Upsert(ctx context.Context, companyID, layerType, key, content string) error {
	switch layerType {
	case "global":
		return s.companyRepo.UpdateSystemPrompt(ctx, companyID, content)
	case "agent":
		return s.agentRepo.UpdatePersona(ctx, key, content)
	case "department", "position":
		return s.repo.Upsert(ctx, &domain.PromptLayer{
			CompanyID: companyID,
			Type:      domain.PromptLayerType(layerType),
			Key:       key,
			Content:   content,
		})
	default:
		return fmt.Errorf("未知的提示词层类型: %s", layerType)
	}
}

// Delete 删除提示词层（清空内容）
func (s *PromptService) Delete(ctx context.Context, companyID, layerType, key string) error {
	switch layerType {
	case "global":
		return s.companyRepo.UpdateSystemPrompt(ctx, companyID, "")
	case "agent":
		return s.agentRepo.UpdatePersona(ctx, key, "")
	case "department", "position":
		return s.repo.Delete(ctx, companyID, layerType, key)
	default:
		return fmt.Errorf("未知的提示词层类型: %s", layerType)
	}
}

// AssembleForAgent 拼接 4 层提示词
func (s *PromptService) AssembleForAgent(ctx context.Context, agent *domain.Agent) string {
	var parts []string

	// 1. 全局提示词
	company, _ := s.companyRepo.GetByID(ctx, agent.CompanyID)
	if company != nil && company.SystemPrompt != "" {
		parts = append(parts, company.SystemPrompt)
	}

	// 2. 部门提示词
	dept := domain.DepartmentOf(agent.Position)
	if dept != "" && dept != "高管" {
		layer, _ := s.repo.Get(ctx, agent.CompanyID, "department", dept)
		if layer != nil && layer.Content != "" {
			parts = append(parts, layer.Content)
		}
	}

	// 3. 职位提示词
	posLayer, _ := s.repo.Get(ctx, agent.CompanyID, "position", string(agent.Position))
	if posLayer != nil && posLayer.Content != "" {
		parts = append(parts, posLayer.Content)
	}

	// 4. Agent 专属提示词
	if agent.Persona != "" {
		parts = append(parts, agent.Persona)
	}

	return strings.Join(parts, "\n\n---\n\n")
}
