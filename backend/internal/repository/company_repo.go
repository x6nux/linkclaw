package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/linkclaw/backend/internal/domain"
)

type companyRepo struct {
	db *gorm.DB
}

func NewCompanyRepo(db *gorm.DB) CompanyRepo {
	return &companyRepo{db: db}
}

func (r *companyRepo) Create(ctx context.Context, c *domain.Company) error {
	result := r.db.WithContext(ctx).Exec(
		`INSERT INTO companies (id, name, slug, description, system_prompt) VALUES ($1, $2, $3, $4, $5)`,
		c.ID, c.Name, c.Slug, c.Description, c.SystemPrompt)
	if result.Error != nil {
		return fmt.Errorf("company create: %w", result.Error)
	}
	return nil
}

func (r *companyRepo) GetByID(ctx context.Context, id string) (*domain.Company, error) {
	var c domain.Company
	result := r.db.WithContext(ctx).Raw(`SELECT * FROM companies WHERE id = $1`, id).Scan(&c)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &c, nil
}

func (r *companyRepo) GetBySlug(ctx context.Context, slug string) (*domain.Company, error) {
	var c domain.Company
	result := r.db.WithContext(ctx).Raw(`SELECT * FROM companies WHERE slug = $1`, slug).Scan(&c)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &c, nil
}

func (r *companyRepo) UpdateSystemPrompt(ctx context.Context, id, prompt string) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE companies SET system_prompt = $1, updated_at = NOW() WHERE id = $2`, prompt, id)
	return result.Error
}

func (r *companyRepo) FindFirst(ctx context.Context) (*domain.Company, error) {
	var c domain.Company
	result := r.db.WithContext(ctx).Raw(`SELECT * FROM companies LIMIT 1`).Scan(&c)
	if result.Error != nil {
		return nil, fmt.Errorf("company find first: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &c, nil
}

func (r *companyRepo) UpdateSettings(ctx context.Context, id string, s *domain.CompanySettings) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE companies
		 SET public_domain = $1, agent_ws_url = $2, mcp_public_url = $3,
		     nanoclaw_image = $4, openclaw_plugin_url = $5, updated_at = NOW()
		 WHERE id = $6`,
		s.PublicDomain, s.AgentWSUrl, s.MCPPublicURL,
		s.NanoclawImage, s.OpenclawPluginURL, id)
	return result.Error
}

func (r *companyRepo) CreateChannel(ctx context.Context, ch *domain.Channel) error {
	result := r.db.WithContext(ctx).Exec(
		`INSERT INTO channels (id, company_id, name, description, is_default)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (company_id, name) DO NOTHING`,
		ch.ID, ch.CompanyID, ch.Name, ch.Description, ch.IsDefault)
	return result.Error
}

func (r *companyRepo) GetChannels(ctx context.Context, companyID string) ([]*domain.Channel, error) {
	var channels []*domain.Channel
	result := r.db.WithContext(ctx).Raw(
		`SELECT * FROM channels WHERE company_id = $1 ORDER BY is_default DESC, name`, companyID,
	).Scan(&channels)
	return channels, result.Error
}

func (r *companyRepo) GetChannelByName(ctx context.Context, companyID, name string) (*domain.Channel, error) {
	var ch domain.Channel
	result := r.db.WithContext(ctx).Raw(
		`SELECT * FROM channels WHERE company_id = $1 AND name = $2`, companyID, name,
	).Scan(&ch)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &ch, nil
}
