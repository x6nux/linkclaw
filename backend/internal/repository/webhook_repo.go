package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/linkclaw/backend/internal/domain"
)

type webhookRepo struct {
	db *gorm.DB
}

func NewWebhookRepo(db *gorm.DB) WebhookRepo {
	return &webhookRepo{db: db}
}

func (r *webhookRepo) Create(ctx context.Context, w *domain.Webhook) error {
	res := r.db.WithContext(ctx).Exec(
		`INSERT INTO webhooks
		(id, company_id, name, url, signing_key_id, events, secret_header, is_active, timeout_seconds, retry_policy)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		w.ID, w.CompanyID, w.Name, w.URL, w.SigningKeyID, w.Events, w.SecretHeader, w.IsActive, w.TimeoutSeconds, w.RetryPolicy,
	)
	if res.Error != nil {
		return fmt.Errorf("webhook create: %w", res.Error)
	}
	return nil
}

func (r *webhookRepo) GetByID(ctx context.Context, id string) (*domain.Webhook, error) {
	var w domain.Webhook
	res := r.db.WithContext(ctx).Raw(`SELECT * FROM webhooks WHERE id = $1`, id).Scan(&w)
	if res.Error != nil {
		return nil, fmt.Errorf("webhook get: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, nil
	}

	if w.SigningKeyID != nil && *w.SigningKeyID != "" {
		k, err := r.GetSigningKeyByID(ctx, *w.SigningKeyID)
		if err != nil {
			return nil, fmt.Errorf("webhook get signing key: %w", err)
		}
		w.SigningKey = k
	}
	return &w, nil
}

func (r *webhookRepo) ListByCompany(ctx context.Context, companyID string) ([]*domain.Webhook, error) {
	var list []*domain.Webhook
	if err := r.db.WithContext(ctx).Raw(
		`SELECT * FROM webhooks WHERE company_id = $1 ORDER BY created_at DESC`, companyID,
	).Scan(&list).Error; err != nil {
		return nil, fmt.Errorf("webhook list: %w", err)
	}
	return list, nil
}

func (r *webhookRepo) ListActiveByEvent(ctx context.Context, companyID string, eventType domain.WebhookEventType) ([]*domain.Webhook, error) {
	var list []*domain.Webhook
	if err := r.db.WithContext(ctx).Raw(
		`SELECT *
		FROM webhooks
		WHERE company_id = $1 AND is_active = TRUE AND (events::jsonb ? $2)
		ORDER BY created_at DESC`,
		companyID, string(eventType),
	).Scan(&list).Error; err != nil {
		return nil, fmt.Errorf("webhook list active by event: %w", err)
	}
	return list, nil
}

func (r *webhookRepo) Update(ctx context.Context, w *domain.Webhook) error {
	res := r.db.WithContext(ctx).Exec(
		`UPDATE webhooks SET
		name = $1,
		url = $2,
		signing_key_id = $3,
		events = $4,
		secret_header = $5,
		is_active = $6,
		timeout_seconds = $7,
		retry_policy = $8,
		updated_at = NOW()
		WHERE id = $9`,
		w.Name, w.URL, w.SigningKeyID, w.Events, w.SecretHeader, w.IsActive, w.TimeoutSeconds, w.RetryPolicy, w.ID,
	)
	if res.Error != nil {
		return fmt.Errorf("webhook update: %w", res.Error)
	}
	return nil
}

func (r *webhookRepo) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Exec(`DELETE FROM webhooks WHERE id = $1`, id)
	if res.Error != nil {
		return fmt.Errorf("webhook delete: %w", res.Error)
	}
	return nil
}

func (r *webhookRepo) CreateSigningKey(ctx context.Context, k *domain.WebhookSigningKey) error {
	res := r.db.WithContext(ctx).Exec(
		`INSERT INTO webhook_signing_keys
		(id, company_id, name, key_type, public_key, secret_key_enc, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		k.ID, k.CompanyID, k.Name, string(k.KeyType), k.PublicKey, k.SecretKeyEnc, k.IsActive,
	)
	if res.Error != nil {
		return fmt.Errorf("webhook signing key create: %w", res.Error)
	}
	return nil
}

func (r *webhookRepo) GetSigningKeyByID(ctx context.Context, id string) (*domain.WebhookSigningKey, error) {
	var k domain.WebhookSigningKey
	res := r.db.WithContext(ctx).Raw(
		`SELECT id, company_id, name, key_type, COALESCE(public_key, '') AS public_key, secret_key_enc, is_active, created_at
		FROM webhook_signing_keys
		WHERE id = $1`,
		id,
	).Scan(&k)
	if res.Error != nil {
		return nil, fmt.Errorf("webhook signing key get: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, nil
	}
	return &k, nil
}

func (r *webhookRepo) ListSigningKeys(ctx context.Context, companyID string) ([]*domain.WebhookSigningKey, error) {
	var list []*domain.WebhookSigningKey
	if err := r.db.WithContext(ctx).Raw(
		`SELECT id, company_id, name, key_type, COALESCE(public_key, '') AS public_key, secret_key_enc, is_active, created_at
		FROM webhook_signing_keys
		WHERE company_id = $1
		ORDER BY created_at DESC`,
		companyID,
	).Scan(&list).Error; err != nil {
		return nil, fmt.Errorf("webhook signing key list: %w", err)
	}
	return list, nil
}

func (r *webhookRepo) DeleteSigningKey(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Exec(`DELETE FROM webhook_signing_keys WHERE id = $1`, id)
	if res.Error != nil {
		return fmt.Errorf("webhook signing key delete: %w", res.Error)
	}
	return nil
}

func (r *webhookRepo) CreateDelivery(ctx context.Context, d *domain.WebhookDelivery) error {
	var sig any
	if d.Signature != "" {
		sig = d.Signature
	}
	res := r.db.WithContext(ctx).Exec(
		`INSERT INTO webhook_deliveries
		(id, webhook_id, company_id, event_type, payload, signature, status, http_status, response_body, attempt_count, next_retry_at, delivered_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		d.ID, d.WebhookID, d.CompanyID, d.EventType, d.Payload, sig, string(d.Status), d.HTTPStatus, d.ResponseBody, d.AttemptCount, d.NextRetryAt, d.DeliveredAt,
	)
	if res.Error != nil {
		return fmt.Errorf("webhook delivery create: %w", res.Error)
	}
	return nil
}

func (r *webhookRepo) GetDeliveryByID(ctx context.Context, id string) (*domain.WebhookDelivery, error) {
	var d domain.WebhookDelivery
	res := r.db.WithContext(ctx).Raw(
		`SELECT
			id, webhook_id, company_id, event_type, payload,
			COALESCE(signature, '') AS signature,
			status, http_status, response_body, attempt_count, next_retry_at, delivered_at, created_at
		FROM webhook_deliveries
		WHERE id = $1`,
		id,
	).Scan(&d)
	if res.Error != nil {
		return nil, fmt.Errorf("webhook delivery get: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, nil
	}
	return &d, nil
}

func (r *webhookRepo) ListDeliveries(ctx context.Context, webhookID string, limit, offset int) ([]*domain.WebhookDelivery, int, error) {
	if limit <= 0 {
		limit = 20
	}

	var total int64
	if err := r.db.WithContext(ctx).Raw(
		`SELECT COUNT(*) FROM webhook_deliveries WHERE webhook_id = $1`, webhookID,
	).Scan(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("webhook delivery count: %w", err)
	}

	var list []*domain.WebhookDelivery
	if err := r.db.WithContext(ctx).Raw(
		`SELECT
			id, webhook_id, company_id, event_type, payload,
			COALESCE(signature, '') AS signature,
			status, http_status, response_body, attempt_count, next_retry_at, delivered_at, created_at
		FROM webhook_deliveries
		WHERE webhook_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		webhookID, limit, offset,
	).Scan(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("webhook delivery list: %w", err)
	}
	return list, int(total), nil
}

func (r *webhookRepo) UpdateDelivery(ctx context.Context, d *domain.WebhookDelivery) error {
	var sig any
	if d.Signature != "" {
		sig = d.Signature
	}
	res := r.db.WithContext(ctx).Exec(
		`UPDATE webhook_deliveries SET
		payload = $1,
		signature = $2,
		status = $3,
		http_status = $4,
		response_body = $5,
		attempt_count = $6,
		next_retry_at = $7,
		delivered_at = $8
		WHERE id = $9`,
		d.Payload, sig, string(d.Status), d.HTTPStatus, d.ResponseBody, d.AttemptCount, d.NextRetryAt, d.DeliveredAt, d.ID,
	)
	if res.Error != nil {
		return fmt.Errorf("webhook delivery update: %w", res.Error)
	}
	return nil
}

func (r *webhookRepo) ListPendingDeliveries(ctx context.Context, limit int) ([]*domain.WebhookDelivery, error) {
	if limit <= 0 {
		limit = 50
	}

	var list []*domain.WebhookDelivery
	if err := r.db.WithContext(ctx).Raw(
		`SELECT
			id, webhook_id, company_id, event_type, payload,
			COALESCE(signature, '') AS signature,
			status, http_status, response_body, attempt_count, next_retry_at, delivered_at, created_at
		FROM webhook_deliveries
		WHERE status IN ('pending', 'retry_later')
		  AND (next_retry_at IS NULL OR next_retry_at <= NOW())
		ORDER BY COALESCE(next_retry_at, created_at) ASC, created_at ASC
		LIMIT $1`,
		limit,
	).Scan(&list).Error; err != nil {
		return nil, fmt.Errorf("webhook delivery list pending: %w", err)
	}
	return list, nil
}
