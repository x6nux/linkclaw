package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
)

type PersonaOptimizerService struct {
	personaRepo repository.PersonaOptimizationRepo
	agentRepo   repository.AgentRepo
	taskRepo    repository.TaskRepo
}

func NewPersonaOptimizerService(
	personaRepo repository.PersonaOptimizationRepo,
	agentRepo repository.AgentRepo,
	taskRepo repository.TaskRepo,
) *PersonaOptimizerService {
	return &PersonaOptimizerService{
		personaRepo: personaRepo,
		agentRepo:   agentRepo,
		taskRepo:    taskRepo,
	}
}

type CreateABTestInput struct {
	CompanyID      string
	Name           string
	Description    string
	ControlAgentID string
	ControlPersona string
	VariantAgentID string
	VariantPersona string
}

func (s *PersonaOptimizerService) GenerateSuggestions(ctx context.Context, companyID, agentID string) ([]*domain.PersonaOptimizationSuggestion, error) {
	agent, err := s.agentRepo.GetByID(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("get agent: %w", err)
	}
	if agent == nil {
		return nil, fmt.Errorf("agent not found")
	}
	if agent.CompanyID != companyID {
		return nil, fmt.Errorf("agent does not belong to company")
	}

	existing, err := s.personaRepo.GetSuggestions(ctx, companyID, agentID, domain.SuggestionStatusPending)
	if err != nil {
		return nil, err
	}
	if len(existing) > 0 {
		return existing, nil
	}

	allParents := ""
	doneTasks, _, err := s.taskRepo.List(ctx, repository.TaskQuery{
		CompanyID:  companyID,
		AssigneeID: agentID,
		Status:     domain.TaskStatusDone,
		ParentID:   &allParents,
		Limit:      100,
	})
	if err != nil {
		return nil, fmt.Errorf("query done tasks: %w", err)
	}
	failedTasks, _, err := s.taskRepo.List(ctx, repository.TaskQuery{
		CompanyID:  companyID,
		AssigneeID: agentID,
		Status:     domain.TaskStatusFailed,
		ParentID:   &allParents,
		Limit:      100,
	})
	if err != nil {
		return nil, fmt.Errorf("query failed tasks: %w", err)
	}

	candidates := s.buildSuggestions(agent.Persona, len(doneTasks), len(failedTasks))
	for _, item := range candidates {
		item.ID = uuid.New().String()
		item.CompanyID = companyID
		item.AgentID = agentID
		item.CurrentPersona = agent.Persona
		item.Status = domain.SuggestionStatusPending
		now := time.Now()
		item.CreatedAt = now
		item.UpdatedAt = now

		if err := s.personaRepo.CreateSuggestion(ctx, item); err != nil {
			return nil, err
		}
	}

	return s.personaRepo.GetSuggestions(ctx, companyID, agentID, domain.SuggestionStatusPending)
}

func (s *PersonaOptimizerService) ApplySuggestion(ctx context.Context, suggestionID, agentID string) error {
	agent, err := s.agentRepo.GetByID(ctx, agentID)
	if err != nil {
		return fmt.Errorf("get agent: %w", err)
	}
	if agent == nil {
		return fmt.Errorf("agent not found")
	}

	suggestions, err := s.personaRepo.GetSuggestions(ctx, agent.CompanyID, agentID, "")
	if err != nil {
		return err
	}
	var target *domain.PersonaOptimizationSuggestion
	for _, sg := range suggestions {
		if sg.ID == suggestionID {
			target = sg
			break
		}
	}
	if target == nil {
		return fmt.Errorf("suggestion not found")
	}
	if target.Status == domain.SuggestionStatusRejected || target.Status == domain.SuggestionStatusApplied {
		return fmt.Errorf("suggestion status is %s", target.Status)
	}

	oldPersona := agent.Persona
	newPersona := applySuggestedChange(oldPersona, target.SuggestedChange)
	if newPersona == oldPersona {
		return fmt.Errorf("suggested change does not alter persona")
	}

	if err := s.agentRepo.UpdatePersona(ctx, agent.ID, newPersona); err != nil {
		return fmt.Errorf("update persona: %w", err)
	}

	history := &domain.PersonaHistory{
		ID:         uuid.New().String(),
		CompanyID:  agent.CompanyID,
		AgentID:    agent.ID,
		OldPersona: oldPersona,
		NewPersona: newPersona,
		ChangeType: domain.ChangeTypeAISuggested,
		ChangedBy:  agentID,
		CreatedAt:  time.Now(),
	}
	if err := s.personaRepo.CreateHistory(ctx, history); err != nil {
		return fmt.Errorf("record persona history: %w", err)
	}
	if err := s.personaRepo.UpdateSuggestionStatus(ctx, suggestionID, domain.SuggestionStatusApplied); err != nil {
		return fmt.Errorf("update suggestion status: %w", err)
	}
	return nil
}

func (s *PersonaOptimizerService) RecordHistory(ctx context.Context, history *domain.PersonaHistory) error {
	if history.ID == "" {
		history.ID = uuid.New().String()
	}
	if history.ChangeType == "" {
		history.ChangeType = domain.ChangeTypeManual
	}
	if history.CreatedAt.IsZero() {
		history.CreatedAt = time.Now()
	}
	return s.personaRepo.CreateHistory(ctx, history)
}

func (s *PersonaOptimizerService) CreateABTest(ctx context.Context, in CreateABTestInput) (*domain.ABTestPersona, error) {
	if in.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if in.ControlAgentID == "" || in.VariantAgentID == "" {
		return nil, fmt.Errorf("controlAgentId and variantAgentId are required")
	}

	controlAgent, err := s.agentRepo.GetByID(ctx, in.ControlAgentID)
	if err != nil {
		return nil, fmt.Errorf("get control agent: %w", err)
	}
	if controlAgent == nil {
		return nil, fmt.Errorf("control agent not found")
	}
	variantAgent, err := s.agentRepo.GetByID(ctx, in.VariantAgentID)
	if err != nil {
		return nil, fmt.Errorf("get variant agent: %w", err)
	}
	if variantAgent == nil {
		return nil, fmt.Errorf("variant agent not found")
	}
	if controlAgent.CompanyID != in.CompanyID || variantAgent.CompanyID != in.CompanyID {
		return nil, fmt.Errorf("agents do not belong to company")
	}

	controlPersona := strings.TrimSpace(in.ControlPersona)
	if controlPersona == "" {
		controlPersona = controlAgent.Persona
	}
	variantPersona := strings.TrimSpace(in.VariantPersona)
	if variantPersona == "" {
		variantPersona = variantAgent.Persona
	}
	now := time.Now()
	t := &domain.ABTestPersona{
		ID:             uuid.New().String(),
		CompanyID:      in.CompanyID,
		Name:           in.Name,
		Description:    in.Description,
		ControlAgentID: in.ControlAgentID,
		ControlPersona: controlPersona,
		VariantAgentID: in.VariantAgentID,
		VariantPersona: variantPersona,
		Status:         domain.ABTestStatusRunning,
		StartTime:      now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.personaRepo.CreateABTest(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *PersonaOptimizerService) GetHistory(ctx context.Context, companyID, agentID string) ([]*domain.PersonaHistory, error) {
	return s.personaRepo.GetHistory(ctx, companyID, agentID, 50)
}

func (s *PersonaOptimizerService) ListSuggestions(
	ctx context.Context,
	companyID, agentID string,
	status domain.SuggestionStatus,
) ([]*domain.PersonaOptimizationSuggestion, error) {
	list, err := s.personaRepo.GetSuggestions(ctx, companyID, agentID, status)
	if err != nil {
		return nil, err
	}
	if len(list) > 0 || status == domain.SuggestionStatusApplied || status == domain.SuggestionStatusRejected {
		return list, nil
	}
	return s.GenerateSuggestions(ctx, companyID, agentID)
}

func (s *PersonaOptimizerService) ListABTests(ctx context.Context, companyID string) ([]*domain.ABTestPersona, error) {
	return s.personaRepo.ListABTests(ctx, companyID)
}

func (s *PersonaOptimizerService) buildSuggestions(persona string, doneCount, failedCount int) []*domain.PersonaOptimizationSuggestion {
	var suggestions []*domain.PersonaOptimizationSuggestion

	if len(persona) > 700 {
		suggestions = append(suggestions, &domain.PersonaOptimizationSuggestion{
			SuggestionType:  domain.SuggestionTypeLength,
			Priority:        domain.SuggestionPriorityHigh,
			SuggestedChange: "将 persona 压缩为 4-6 条关键职责，减少冗余描述，保留可执行约束。",
			Reason:          "当前 persona 过长，可能影响响应速度与指令聚焦。",
			Confidence:      0.68,
		})
	}
	if doneCount < 5 {
		suggestions = append(suggestions, &domain.PersonaOptimizationSuggestion{
			SuggestionType:  domain.SuggestionTypeStructure,
			Priority:        domain.SuggestionPriorityMedium,
			SuggestedChange: "按\"目标 / 输入 / 输出 / 约束 / 失败处理\"结构重写 persona，明确交付标准。",
			Reason:          "已完成任务数量偏少，建议强化结构化执行框架。",
			Confidence:      0.72,
		})
	}
	if failedCount > doneCount/2 && failedCount > 0 {
		suggestions = append(suggestions, &domain.PersonaOptimizationSuggestion{
			SuggestionType:  domain.SuggestionTypeContent,
			Priority:        domain.SuggestionPriorityHigh,
			SuggestedChange: "补充边界条件与澄清流程：信息不完整时先提出问题，再执行任务。",
			Reason:          "失败任务占比较高，建议增加澄清和风险控制策略。",
			Confidence:      0.75,
		})
	}
	if len(suggestions) == 0 {
		suggestions = append(suggestions, &domain.PersonaOptimizationSuggestion{
			SuggestionType:  domain.SuggestionTypeTone,
			Priority:        domain.SuggestionPriorityLow,
			SuggestedChange: "保持专业且简洁的沟通语气，并在回复中优先给出可执行下一步。",
			Reason:          "当前表现稳定，建议微调沟通风格以提升协作效率。",
			Confidence:      0.61,
		})
	}
	return suggestions
}

func applySuggestedChange(currentPersona, suggestedChange string) string {
	base := strings.TrimSpace(currentPersona)
	change := strings.TrimSpace(suggestedChange)
	if change == "" {
		return base
	}
	if len(base) > 0 && len(change) >= len(base)/2 {
		return change
	}
	if base == "" {
		return change
	}
	if strings.Contains(base, change) {
		return base
	}
	return base + "\n\n" + change
}
