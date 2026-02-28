package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/linkclaw/backend/internal/domain"
)

type agentRepo struct {
	db *gorm.DB
}

func NewAgentRepo(db *gorm.DB) AgentRepo {
	return &agentRepo{db: db}
}

func (r *agentRepo) Create(ctx context.Context, a *domain.Agent) error {
	q := `INSERT INTO agents
		(id, company_id, name, role, role_type, position, model, is_human, permissions,
		 persona, status, api_key_hash, api_key_prefix, password_hash, hire_request_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`
	result := r.db.WithContext(ctx).Exec(q,
		a.ID, a.CompanyID, a.Name, a.Role, string(a.RoleType), string(a.Position),
		a.Model, a.IsHuman, a.Permissions, a.Persona, string(a.Status),
		a.APIKeyHash, a.APIKeyPrefix, a.PasswordHash, a.HireRequestID)
	if result.Error != nil {
		return fmt.Errorf("agent create: %w", result.Error)
	}
	return nil
}

func (r *agentRepo) GetByHireRequestID(ctx context.Context, requestID string) (*domain.Agent, error) {
	var a domain.Agent
	result := r.db.WithContext(ctx).Raw(
		`SELECT * FROM agents WHERE hire_request_id = $1`, requestID,
	).Scan(&a)
	if result.Error != nil {
		return nil, fmt.Errorf("agent get by request id: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &a, nil
}

func (r *agentRepo) GetByID(ctx context.Context, id string) (*domain.Agent, error) {
	var a domain.Agent
	result := r.db.WithContext(ctx).Raw(`SELECT * FROM agents WHERE id = $1`, id).Scan(&a)
	if result.Error != nil {
		return nil, fmt.Errorf("agent get: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &a, nil
}

func (r *agentRepo) GetByAPIKeyHash(ctx context.Context, hash string) (*domain.Agent, error) {
	var a domain.Agent
	result := r.db.WithContext(ctx).Raw(`SELECT * FROM agents WHERE api_key_hash = $1`, hash).Scan(&a)
	if result.Error != nil {
		return nil, fmt.Errorf("agent get by key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &a, nil
}

func (r *agentRepo) GetByCompany(ctx context.Context, companyID string) ([]*domain.Agent, error) {
	var agents []*domain.Agent
	result := r.db.WithContext(ctx).Raw(
		`SELECT * FROM agents WHERE company_id = $1 ORDER BY role_type, position, name`, companyID,
	).Scan(&agents)
	if result.Error != nil {
		return nil, fmt.Errorf("agent list: %w", result.Error)
	}
	return agents, nil
}

func (r *agentRepo) UpdateStatus(ctx context.Context, id string, status domain.AgentStatus) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE agents SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	return result.Error
}

func (r *agentRepo) UpdateLastSeen(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE agents SET last_seen_at = $1, status = 'online', updated_at = NOW() WHERE id = $2`,
		time.Now(), id)
	return result.Error
}

func (r *agentRepo) UpdateName(ctx context.Context, id, name string) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE agents SET name = $1, updated_at = NOW() WHERE id = $2`, name, id)
	return result.Error
}

func (r *agentRepo) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Exec(`DELETE FROM agents WHERE id = $1`, id)
	return result.Error
}

func (r *agentRepo) GetByName(ctx context.Context, companyID, name string) (*domain.Agent, error) {
	var a domain.Agent
	result := r.db.WithContext(ctx).Raw(
		`SELECT * FROM agents WHERE company_id = $1 AND name = $2 AND is_human = TRUE`,
		companyID, name,
	).Scan(&a)
	if result.Error != nil {
		return nil, fmt.Errorf("agent get by name: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &a, nil
}

func (r *agentRepo) SetPasswordHash(ctx context.Context, id, hash string) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE agents SET password_hash = $1, updated_at = NOW() WHERE id = $2`, hash, id)
	return result.Error
}

func (r *agentRepo) UpdatePersona(ctx context.Context, id, persona string) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE agents SET persona = $1, updated_at = NOW() WHERE id = $2`, persona, id)
	return result.Error
}

func (r *agentRepo) UpdateModel(ctx context.Context, id, model string) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE agents SET model = $1, updated_at = NOW() WHERE id = $2`, model, id)
	return result.Error
}

func (r *agentRepo) UpdateAPIKey(ctx context.Context, id, hash, prefix string) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE agents SET api_key_hash = $1, api_key_prefix = $2, updated_at = NOW() WHERE id = $3`,
		hash, prefix, id)
	return result.Error
}

func (r *agentRepo) UpdateDepartment(ctx context.Context, id string, departmentID *string) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE agents SET department_id = $1, updated_at = NOW() WHERE id = $2`,
		departmentID, id,
	)
	return result.Error
}

func (r *agentRepo) UpdateManager(ctx context.Context, id string, managerID *string) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE agents SET manager_id = $1, updated_at = NOW() WHERE id = $2`,
		managerID, id,
	)
	return result.Error
}

func (r *agentRepo) ListByDepartment(ctx context.Context, companyID, departmentID string) ([]*domain.Agent, error) {
	var agents []*domain.Agent
	result := r.db.WithContext(ctx).Raw(
		`SELECT * FROM agents WHERE company_id = $1 AND department_id = $2 ORDER BY created_at`,
		companyID, departmentID,
	).Scan(&agents)
	if result.Error != nil {
		return nil, fmt.Errorf("agent list by department: %w", result.Error)
	}
	return agents, nil
}

func (r *agentRepo) MarkInitialized(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE agents SET initialized = TRUE, updated_at = NOW() WHERE id = $1`, id)
	return result.Error
}
