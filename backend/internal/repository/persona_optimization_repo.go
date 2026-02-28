package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/linkclaw/backend/internal/domain"
)

type personaOptimizationRepo struct {
	db *gorm.DB
}

func NewPersonaOptimizationRepo(db *gorm.DB) PersonaOptimizationRepo {
	return &personaOptimizationRepo{db: db}
}

func (r *personaOptimizationRepo) CreateSuggestion(ctx context.Context, s *domain.PersonaOptimizationSuggestion) error {
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now()
	}
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = s.CreatedAt
	}
	if s.Priority == "" {
		s.Priority = domain.SuggestionPriorityMedium
	}
	result := r.db.WithContext(ctx).Exec(
		`INSERT INTO persona_optimization_suggestions
		(id, company_id, agent_id, suggestion_type, priority, current_persona, suggested_change, reason, confidence, status, applied_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		s.ID, s.CompanyID, s.AgentID, string(s.SuggestionType), string(s.Priority), s.CurrentPersona,
		s.SuggestedChange, s.Reason, s.Confidence, string(s.Status), s.AppliedAt, s.CreatedAt, s.UpdatedAt,
	)
	if result.Error != nil {
		return fmt.Errorf("persona suggestion create: %w", result.Error)
	}
	return nil
}

func (r *personaOptimizationRepo) GetSuggestions(ctx context.Context, companyID, agentID string, status domain.SuggestionStatus) ([]*domain.PersonaOptimizationSuggestion, error) {
	conds := []string{"company_id = $1"}
	args := []any{companyID}
	idx := 2

	if agentID != "" {
		conds = append(conds, fmt.Sprintf("agent_id = $%d", idx))
		args = append(args, agentID)
		idx++
	}
	if status != "" {
		conds = append(conds, fmt.Sprintf("status = $%d", idx))
		args = append(args, string(status))
	}

	var suggestions []*domain.PersonaOptimizationSuggestion
	if err := r.db.WithContext(ctx).Raw(
		`SELECT * FROM persona_optimization_suggestions WHERE `+strings.Join(conds, " AND ")+` ORDER BY created_at DESC`,
		args...,
	).Scan(&suggestions).Error; err != nil {
		return nil, fmt.Errorf("persona suggestion list: %w", err)
	}
	return suggestions, nil
}

func (r *personaOptimizationRepo) UpdateSuggestionStatus(ctx context.Context, id string, status domain.SuggestionStatus) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE persona_optimization_suggestions SET status = $1, updated_at = NOW() WHERE id = $2`,
		string(status), id,
	)
	if result.Error != nil {
		return fmt.Errorf("persona suggestion status update: %w", result.Error)
	}
	return nil
}

func (r *personaOptimizationRepo) CreateHistory(ctx context.Context, h *domain.PersonaHistory) error {
	if h.CreatedAt.IsZero() {
		h.CreatedAt = time.Now()
	}
	result := r.db.WithContext(ctx).Exec(
		`INSERT INTO persona_history
		(id, company_id, agent_id, old_persona, new_persona, change_reason, suggestion_id, change_type, changed_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		h.ID, h.CompanyID, h.AgentID, h.OldPersona, h.NewPersona, h.ChangeReason, h.SuggestionID, string(h.ChangeType), h.ChangedBy, h.CreatedAt,
	)
	if result.Error != nil {
		return fmt.Errorf("persona history create: %w", result.Error)
	}
	return nil
}

func (r *personaOptimizationRepo) GetHistory(ctx context.Context, companyID, agentID string, limit int) ([]*domain.PersonaHistory, error) {
	if limit <= 0 {
		limit = 20
	}
	conds := []string{"company_id = $1"}
	args := []any{companyID}
	idx := 2

	if agentID != "" {
		conds = append(conds, fmt.Sprintf("agent_id = $%d", idx))
		args = append(args, agentID)
		idx++
	}

	q := `SELECT * FROM persona_history WHERE ` + strings.Join(conds, " AND ") +
		` ORDER BY created_at DESC LIMIT $` + fmt.Sprint(idx)
	args = append(args, limit)

	var list []*domain.PersonaHistory
	if err := r.db.WithContext(ctx).Raw(q, args...).Scan(&list).Error; err != nil {
		return nil, fmt.Errorf("persona history list: %w", err)
	}
	return list, nil
}

func (r *personaOptimizationRepo) CreateABTest(ctx context.Context, t *domain.ABTestPersona) error {
	if t.StartTime.IsZero() {
		t.StartTime = time.Now()
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = t.CreatedAt
	}
	result := r.db.WithContext(ctx).Exec(
		`INSERT INTO ab_test_personas
		(id, company_id, name, description, control_agent_id, control_persona, variant_agent_id, variant_persona, status, start_time, end_time,
		 control_tasks_completed, variant_tasks_completed, control_avg_duration, variant_avg_duration, winner, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`,
		t.ID, t.CompanyID, t.Name, t.Description, t.ControlAgentID, t.ControlPersona,
		t.VariantAgentID, t.VariantPersona, string(t.Status), t.StartTime, t.EndTime,
		t.ControlTasksCompleted, t.VariantTasksCompleted, t.ControlAvgDuration, t.VariantAvgDuration, t.Winner,
		t.CreatedAt, t.UpdatedAt,
	)
	if result.Error != nil {
		return fmt.Errorf("ab test create: %w", result.Error)
	}
	return nil
}

func (r *personaOptimizationRepo) GetABTest(ctx context.Context, id string) (*domain.ABTestPersona, error) {
	var t domain.ABTestPersona
	result := r.db.WithContext(ctx).Raw(
		`SELECT * FROM ab_test_personas WHERE id = $1`,
		id,
	).Scan(&t)
	if result.Error != nil {
		return nil, fmt.Errorf("ab test get: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &t, nil
}

func (r *personaOptimizationRepo) ListABTests(ctx context.Context, companyID string) ([]*domain.ABTestPersona, error) {
	var list []*domain.ABTestPersona
	if err := r.db.WithContext(ctx).Raw(
		`SELECT * FROM ab_test_personas WHERE company_id = $1 ORDER BY start_time DESC`,
		companyID,
	).Scan(&list).Error; err != nil {
		return nil, fmt.Errorf("ab test list: %w", err)
	}
	return list, nil
}

func (r *personaOptimizationRepo) UpdateABTest(ctx context.Context, t *domain.ABTestPersona) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE ab_test_personas
		SET name = $1,
		    description = $2,
		    control_agent_id = $3,
		    control_persona = $4,
		    variant_agent_id = $5,
		    variant_persona = $6,
		    status = $7,
		    start_time = $8,
		    end_time = $9,
		    control_tasks_completed = $10,
		    variant_tasks_completed = $11,
		    control_avg_duration = $12,
		    variant_avg_duration = $13,
		    winner = $14,
		    updated_at = NOW()
		WHERE id = $15`,
		t.Name, t.Description, t.ControlAgentID, t.ControlPersona, t.VariantAgentID, t.VariantPersona,
		string(t.Status), t.StartTime, t.EndTime, t.ControlTasksCompleted, t.VariantTasksCompleted,
		t.ControlAvgDuration, t.VariantAvgDuration, t.Winner, t.ID,
	)
	if result.Error != nil {
		return fmt.Errorf("ab test update: %w", result.Error)
	}
	return nil
}
