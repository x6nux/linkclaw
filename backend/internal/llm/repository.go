package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateProvider(ctx context.Context, p *Provider) error {
	p.ID = uuid.New().String()
	result := r.db.WithContext(ctx).Exec(
		`INSERT INTO llm_providers
		(id, company_id, name, provider_type, base_url, api_key_enc, models, weight, is_active, max_rpm)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		p.ID, p.CompanyID, p.Name, string(p.Type), p.BaseURL,
		p.APIKeyEnc, p.Models, p.Weight, p.IsActive, p.MaxRPM)
	return result.Error
}

func (r *Repository) GetProvider(ctx context.Context, id string) (*Provider, error) {
	var p Provider
	result := r.db.WithContext(ctx).Raw(`SELECT * FROM llm_providers WHERE id = $1`, id).Scan(&p)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &p, nil
}

func (r *Repository) ListProviders(ctx context.Context, companyID string) ([]*Provider, error) {
	var providers []*Provider
	result := r.db.WithContext(ctx).Raw(
		`SELECT * FROM llm_providers WHERE company_id = $1 ORDER BY weight DESC, name`, companyID,
	).Scan(&providers)
	return providers, result.Error
}

func (r *Repository) ListActiveProviders(ctx context.Context, companyID string) ([]*Provider, error) {
	var providers []*Provider
	result := r.db.WithContext(ctx).Raw(
		`SELECT * FROM llm_providers WHERE company_id = $1 AND is_active = TRUE ORDER BY weight DESC`, companyID,
	).Scan(&providers)
	return providers, result.Error
}

func (r *Repository) UpdateProvider(ctx context.Context, p *Provider) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE llm_providers SET
		name=$1, provider_type=$2, base_url=$3, api_key_enc=$4, models=$5,
		weight=$6, is_active=$7, max_rpm=$8, updated_at=NOW()
		WHERE id=$9`,
		p.Name, string(p.Type), p.BaseURL, p.APIKeyEnc, p.Models,
		p.Weight, p.IsActive, p.MaxRPM, p.ID)
	return result.Error
}

func (r *Repository) DeleteProvider(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Exec(`DELETE FROM llm_providers WHERE id = $1`, id)
	return result.Error
}

func (r *Repository) MarkProviderError(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE llm_providers SET error_count = error_count+1, last_error_at = $1 WHERE id = $2`,
		time.Now(), id)
	return result.Error
}

func (r *Repository) MarkProviderUsed(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE llm_providers SET last_used_at = $1, error_count = 0 WHERE id = $2`,
		time.Now(), id)
	return result.Error
}

func (r *Repository) InsertUsageLog(ctx context.Context, log *UsageLog) error {
	log.ID = uuid.New().String()
	result := r.db.WithContext(ctx).Exec(
		`INSERT INTO llm_usage_logs
		(id, company_id, provider_id, agent_id, request_model,
		 input_tokens, output_tokens, cache_creation_tokens, cache_read_tokens, cached_prompt_tokens,
		 cost_microdollars, status, latency_ms, retry_count, error_msg)
		VALUES
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		log.ID, log.CompanyID, log.ProviderID, log.AgentID, log.RequestModel,
		log.InputTokens, log.OutputTokens, log.CacheCreationTokens, log.CacheReadTokens, log.CachedPromptTokens,
		log.CostMicrodollars, log.Status, log.LatencyMs, log.RetryCount, log.ErrorMsg)
	return result.Error
}

func (r *Repository) GetUsageStats(ctx context.Context, companyID string) ([]*UsageStats, error) {
	var stats []*UsageStats
	result := r.db.WithContext(ctx).Raw(`
		SELECT
			l.provider_id,
			COALESCE(p.name, '已删除') AS name,
			COUNT(*)                        AS total_requests,
			SUM(CASE WHEN l.status='success' THEN 1 ELSE 0 END) AS success_requests,
			SUM(l.input_tokens)             AS input_tokens,
			SUM(l.output_tokens)            AS output_tokens,
			SUM(l.cache_creation_tokens)    AS cache_creation_tokens,
			SUM(l.cache_read_tokens)        AS cache_read_tokens,
			SUM(l.cost_microdollars) / 1000000.0 AS total_cost_usd
		FROM llm_usage_logs l
		LEFT JOIN llm_providers p ON l.provider_id = p.id
		WHERE l.company_id = $1
		GROUP BY l.provider_id, p.name
		ORDER BY total_cost_usd DESC`, companyID).Scan(&stats)
	return stats, result.Error
}

func (r *Repository) GetDailyUsage(ctx context.Context, companyID string) ([]*DailyUsage, error) {
	var daily []*DailyUsage
	result := r.db.WithContext(ctx).Raw(`
		SELECT
			to_char(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS date,
			SUM(input_tokens)   AS input_tokens,
			SUM(output_tokens)  AS output_tokens,
			SUM(cost_microdollars) / 1000000.0 AS cost_usd,
			COUNT(*) AS requests
		FROM llm_usage_logs
		WHERE company_id = $1
		  AND created_at >= NOW() - INTERVAL '30 days'
		GROUP BY 1
		ORDER BY 1`, companyID).Scan(&daily)
	if result.Error != nil {
		return nil, fmt.Errorf("GetDailyUsage: %w", result.Error)
	}
	return daily, nil
}

func (r *Repository) GetRecentLogs(ctx context.Context, companyID string, limit int) ([]*UsageLog, error) {
	if limit <= 0 {
		limit = 50
	}
	var logs []*UsageLog
	result := r.db.WithContext(ctx).Raw(
		`SELECT * FROM llm_usage_logs WHERE company_id = $1 ORDER BY created_at DESC LIMIT $2`,
		companyID, limit,
	).Scan(&logs)
	return logs, result.Error
}
