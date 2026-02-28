package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type WebhookEventType string

const (
	WebhookEventAgentOnline   WebhookEventType = "agent.online"
	WebhookEventAgentOffline  WebhookEventType = "agent.offline"
	WebhookEventTaskCreated   WebhookEventType = "task.created"
	WebhookEventTaskUpdated   WebhookEventType = "task.updated"
	WebhookEventTaskCompleted WebhookEventType = "task.completed"
	WebhookEventMessageNew    WebhookEventType = "message.new"
	WebhookEventApprovalEvent WebhookEventType = "approval.event"
	WebhookEventBudgetAlert   WebhookEventType = "budget.alert.created"
	WebhookEventErrorAlert    WebhookEventType = "error_alert.created"
)

type WebhookSigningKeyType string

const (
	WebhookSigningKeyHMACSHA256 WebhookSigningKeyType = "hmac_sha256"
	WebhookSigningKeyEd25519    WebhookSigningKeyType = "ed25519"
)

type WebhookDeliveryStatus string

const (
	WebhookDeliveryStatusPending    WebhookDeliveryStatus = "pending"
	WebhookDeliveryStatusDelivering WebhookDeliveryStatus = "delivering"
	WebhookDeliveryStatusSuccess    WebhookDeliveryStatus = "success"
	WebhookDeliveryStatusFailed     WebhookDeliveryStatus = "failed"
	WebhookDeliveryStatusRetryLater WebhookDeliveryStatus = "retry_later"
)

type JSONMap map[string]any

func (m *JSONMap) Scan(value any) error {
	if m == nil {
		return errors.New("JSONMap.Scan: nil receiver")
	}
	if value == nil {
		*m = JSONMap{}
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return errors.New("JSONMap.Scan: unsupported type")
	}

	str := strings.TrimSpace(string(data))
	if str == "" || str == "null" {
		*m = JSONMap{}
		return nil
	}

	out := JSONMap{}
	if err := json.Unmarshal(data, &out); err != nil {
		*m = JSONMap{}
		return nil
	}
	*m = out
	return nil
}

func (m JSONMap) Value() (driver.Value, error) {
	if len(m) == 0 {
		return "{}", nil
	}
	b, err := json.Marshal(map[string]any(m))
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

type WebhookSigningKey struct {
	ID           string                `gorm:"column:id"             json:"id"`
	CompanyID    string                `gorm:"column:company_id"     json:"company_id"`
	Name         string                `gorm:"column:name"           json:"name"`
	KeyType      WebhookSigningKeyType `gorm:"column:key_type"       json:"key_type"`
	PublicKey    string                `gorm:"column:public_key"     json:"public_key"`
	SecretKeyEnc []byte                `gorm:"column:secret_key_enc" json:"-"`
	IsActive     bool                  `gorm:"column:is_active"       json:"is_active"`
	CreatedAt    time.Time             `gorm:"column:created_at"      json:"created_at"`
}

type Webhook struct {
	ID             string             `gorm:"column:id"              json:"id"`
	CompanyID      string             `gorm:"column:company_id"      json:"company_id"`
	Name           string             `gorm:"column:name"            json:"name"`
	URL            string             `gorm:"column:url"             json:"url"`
	SigningKeyID   *string            `gorm:"column:signing_key_id"  json:"signing_key_id"`
	Events         StringList         `gorm:"column:events"          json:"events"`
	SecretHeader   string             `gorm:"column:secret_header"   json:"secret_header"`
	IsActive       bool               `gorm:"column:is_active"       json:"is_active"`
	TimeoutSeconds int                `gorm:"column:timeout_seconds" json:"timeout_seconds"`
	RetryPolicy    JSONMap            `gorm:"column:retry_policy"    json:"retry_policy"`
	CreatedAt      time.Time          `gorm:"column:created_at"      json:"created_at"`
	UpdatedAt      time.Time          `gorm:"column:updated_at"      json:"updated_at"`
	SigningKey     *WebhookSigningKey `gorm:"-"                      json:"signing_key,omitempty"`
}

type WebhookDelivery struct {
	ID           string                `gorm:"column:id"            json:"id"`
	WebhookID    string                `gorm:"column:webhook_id"    json:"webhook_id"`
	CompanyID    string                `gorm:"column:company_id"    json:"company_id"`
	EventType    string                `gorm:"column:event_type"    json:"event_type"`
	Payload      JSONMap               `gorm:"column:payload"       json:"payload"`
	Signature    string                `gorm:"column:signature"     json:"signature"`
	Status       WebhookDeliveryStatus `gorm:"column:status"        json:"status"`
	HTTPStatus   *int                  `gorm:"column:http_status"   json:"http_status"`
	ResponseBody *string               `gorm:"column:response_body" json:"response_body"`
	AttemptCount int                   `gorm:"column:attempt_count" json:"attempt_count"`
	NextRetryAt  *time.Time            `gorm:"column:next_retry_at" json:"next_retry_at"`
	DeliveredAt  *time.Time            `gorm:"column:delivered_at"  json:"delivered_at"`
	CreatedAt    time.Time             `gorm:"column:created_at"    json:"created_at"`
}

type RetryPolicy struct {
	MaxAttempts int `json:"max_attempts"`
	BackoffBase int `json:"backoff_base"`
	BackoffMax  int `json:"backoff_max"`
}
