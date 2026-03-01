package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/linkclaw/backend/internal/domain"
)

type partnerAPIKeyRepo struct {
	db *gorm.DB
}

func NewPartnerAPIKeyRepo(db *gorm.DB) PartnerAPIKeyRepo {
	return &partnerAPIKeyRepo{db: db}
}

func (r *partnerAPIKeyRepo) Create(ctx context.Context, k *domain.PartnerApiKey) error {
	q := `INSERT INTO partner_api_keys
		(id, company_id, partner_slug, partner_id, name, key_hash, key_prefix, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (company_id, partner_slug) DO UPDATE SET
			key_hash = EXCLUDED.key_hash,
			key_prefix = EXCLUDED.key_prefix,
			is_active = TRUE,
			updated_at = NOW()`
	result := r.db.WithContext(ctx).Exec(q,
		k.ID, k.CompanyID, k.PartnerSlug, k.PartnerID, k.Name,
		k.KeyHash, k.KeyPrefix, k.IsActive)
	if result.Error != nil {
		return fmt.Errorf("partner api key create: %w", result.Error)
	}
	return nil
}

func (r *partnerAPIKeyRepo) GetByCompanyAndPartner(ctx context.Context, companyID, partnerSlug string) (*domain.PartnerApiKey, error) {
	var k domain.PartnerApiKey
	result := r.db.WithContext(ctx).Raw(
		`SELECT * FROM partner_api_keys
		 WHERE company_id = $1 AND partner_slug = $2 AND is_active = TRUE`,
		companyID, partnerSlug,
	).Scan(&k)
	if result.Error != nil {
		return nil, fmt.Errorf("partner api key get: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &k, nil
}

func (r *partnerAPIKeyRepo) GetByKeyHash(ctx context.Context, companyID, keyHash string) (*domain.PartnerApiKey, error) {
	var k domain.PartnerApiKey
	result := r.db.WithContext(ctx).Raw(
		`SELECT * FROM partner_api_keys
		 WHERE company_id = $1 AND key_hash = $2 AND is_active = TRUE`,
		companyID, keyHash,
	).Scan(&k)
	if result.Error != nil {
		return nil, fmt.Errorf("partner api key get by hash: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &k, nil
}

func (r *partnerAPIKeyRepo) Deactivate(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE partner_api_keys SET is_active = FALSE, updated_at = NOW() WHERE id = $1`, id)
	if result.Error != nil {
		return fmt.Errorf("partner api key deactivate: %w", result.Error)
	}
	return nil
}

func (r *partnerAPIKeyRepo) Regenerate(ctx context.Context, id, newHash, newPrefix string) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE partner_api_keys
		 SET key_hash = $1, key_prefix = $2, updated_at = NOW()
		 WHERE id = $3`,
		newHash, newPrefix, id)
	if result.Error != nil {
		return fmt.Errorf("partner api key regenerate: %w", result.Error)
	}
	return nil
}

func (r *partnerAPIKeyRepo) UpdateLastUsed(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE partner_api_keys SET last_used_at = NOW(), updated_at = NOW() WHERE id = $1`, id)
	if result.Error != nil {
		return fmt.Errorf("partner api key update last used: %w", result.Error)
	}
	return nil
}
