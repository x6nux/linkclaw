package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/linkclaw/backend/internal/domain"
)

type departmentRepo struct {
	db *gorm.DB
}

func NewDepartmentRepo(db *gorm.DB) DepartmentRepo {
	return &departmentRepo{db: db}
}

func (r *departmentRepo) Create(ctx context.Context, d *domain.Department) error {
	q := `INSERT INTO departments
		(id, company_id, name, slug, description, director_agent_id, parent_dept_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	result := r.db.WithContext(ctx).Exec(q,
		d.ID, d.CompanyID, d.Name, d.Slug, d.Description, d.DirectorAgentID, d.ParentDeptID, d.CreatedAt,
	)
	if result.Error != nil {
		return fmt.Errorf("department create: %w", result.Error)
	}
	return nil
}

func (r *departmentRepo) GetByID(ctx context.Context, id string) (*domain.Department, error) {
	var d domain.Department
	result := r.db.WithContext(ctx).Raw(`SELECT * FROM departments WHERE id = $1`, id).Scan(&d)
	if result.Error != nil {
		return nil, fmt.Errorf("department get: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &d, nil
}

func (r *departmentRepo) GetBySlug(ctx context.Context, companyID, slug string) (*domain.Department, error) {
	var d domain.Department
	result := r.db.WithContext(ctx).Raw(
		`SELECT * FROM departments WHERE company_id = $1 AND slug = $2`,
		companyID, slug,
	).Scan(&d)
	if result.Error != nil {
		return nil, fmt.Errorf("department get by slug: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &d, nil
}

func (r *departmentRepo) List(ctx context.Context, companyID string) ([]*domain.Department, error) {
	var departments []*domain.Department
	result := r.db.WithContext(ctx).Raw(
		`SELECT * FROM departments WHERE company_id = $1 ORDER BY created_at`,
		companyID,
	).Scan(&departments)
	if result.Error != nil {
		return nil, fmt.Errorf("department list: %w", result.Error)
	}
	return departments, nil
}

func (r *departmentRepo) Update(ctx context.Context, d *domain.Department) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE departments
		SET name = $1, description = $2, slug = $3, director_agent_id = $4, parent_dept_id = $5
		WHERE id = $6`,
		d.Name, d.Description, d.Slug, d.DirectorAgentID, d.ParentDeptID, d.ID,
	)
	if result.Error != nil {
		return fmt.Errorf("department update: %w", result.Error)
	}
	return nil
}

func (r *departmentRepo) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Exec(`DELETE FROM departments WHERE id = $1`, id)
	if result.Error != nil {
		return fmt.Errorf("department delete: %w", result.Error)
	}
	return nil
}

func (r *departmentRepo) AssignAgent(ctx context.Context, agentID, departmentID string) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE agents SET department_id = $1 WHERE id = $2`,
		departmentID, agentID,
	)
	if result.Error != nil {
		return fmt.Errorf("department assign agent: %w", result.Error)
	}
	return nil
}
