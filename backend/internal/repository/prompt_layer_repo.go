package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/linkclaw/backend/internal/domain"
)

type promptLayerRepo struct {
	db *gorm.DB
}

func NewPromptLayerRepo(db *gorm.DB) PromptLayerRepo {
	return &promptLayerRepo{db: db}
}

func (r *promptLayerRepo) Upsert(ctx context.Context, layer *domain.PromptLayer) error {
	result := r.db.WithContext(ctx).Exec(
		`INSERT INTO prompt_layers (id, company_id, type, key, content, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, NOW())
		ON CONFLICT (company_id, type, key)
		DO UPDATE SET content = $4, updated_at = NOW()`,
		layer.CompanyID, layer.Type, layer.Key, layer.Content)
	if result.Error != nil {
		return fmt.Errorf("prompt layer upsert: %w", result.Error)
	}
	return nil
}

func (r *promptLayerRepo) Delete(ctx context.Context, companyID, layerType, key string) error {
	result := r.db.WithContext(ctx).Exec(
		`DELETE FROM prompt_layers WHERE company_id = $1 AND type = $2 AND key = $3`,
		companyID, layerType, key)
	return result.Error
}

func (r *promptLayerRepo) ListByCompany(ctx context.Context, companyID string) ([]*domain.PromptLayer, error) {
	var layers []*domain.PromptLayer
	result := r.db.WithContext(ctx).Raw(
		`SELECT * FROM prompt_layers WHERE company_id = $1 ORDER BY type, key`,
		companyID,
	).Scan(&layers)
	if result.Error != nil {
		return nil, fmt.Errorf("prompt layer list: %w", result.Error)
	}
	return layers, nil
}

func (r *promptLayerRepo) Get(ctx context.Context, companyID, layerType, key string) (*domain.PromptLayer, error) {
	var layer domain.PromptLayer
	result := r.db.WithContext(ctx).Raw(
		`SELECT * FROM prompt_layers WHERE company_id = $1 AND type = $2 AND key = $3`,
		companyID, layerType, key,
	).Scan(&layer)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &layer, nil
}
