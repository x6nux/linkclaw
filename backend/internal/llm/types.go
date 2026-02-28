package llm

import (
	"time"

	"github.com/linkclaw/backend/internal/domain"
)

// ProviderType LLM 服务商类型
type ProviderType string

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
)

// Provider 配置模型（对应 llm_providers 表）
type Provider struct {
	ID          string       `gorm:"column:id"            json:"id"`
	CompanyID   string       `gorm:"column:company_id"    json:"company_id"`
	Name        string       `gorm:"column:name"          json:"name"`
	Type        ProviderType `gorm:"column:provider_type" json:"type"`
	BaseURL     string       `gorm:"column:base_url"      json:"base_url"`
	APIKeyEnc   string            `gorm:"column:api_key_enc"   json:"-"`
	Models      domain.StringList `gorm:"column:models"        json:"models"`
	Weight      int               `gorm:"column:weight"        json:"weight"`
	IsActive    bool         `gorm:"column:is_active"     json:"is_active"`
	ErrorCount  int          `gorm:"column:error_count"   json:"error_count"`
	LastErrorAt *time.Time   `gorm:"column:last_error_at" json:"last_error_at"`
	LastUsedAt  *time.Time   `gorm:"column:last_used_at"  json:"last_used_at"`
	MaxRPM      *int         `gorm:"column:max_rpm"       json:"max_rpm"`
	CreatedAt   time.Time    `gorm:"column:created_at"    json:"created_at"`
	UpdatedAt   time.Time    `gorm:"column:updated_at"    json:"updated_at"`
}

// ProviderStatus 健康状态（运行时计算，不存数据库）
type ProviderStatus string

const (
	StatusHealthy  ProviderStatus = "healthy"
	StatusDegraded ProviderStatus = "degraded" // 近期有错误，仍可用
	StatusDown     ProviderStatus = "down"     // 已冷却，暂停使用
)

// ProviderView 带运行时状态的 Provider（返回给前端）
type ProviderView struct {
	Provider
	Status       ProviderStatus `json:"status"`
	APIKeyPrefix string         `json:"api_key_prefix"` // 显示前10字符
	APIKeyEnc    string         `json:"-"`              // 隐藏
}

// UsageLog 对应 llm_usage_logs 表
type UsageLog struct {
	ID                  string    `gorm:"column:id"                    json:"id"`
	CompanyID           string    `gorm:"column:company_id"            json:"company_id"`
	ProviderID          *string   `gorm:"column:provider_id"           json:"provider_id"`
	AgentID             *string   `gorm:"column:agent_id"              json:"agent_id"`
	RequestModel        string    `gorm:"column:request_model"         json:"request_model"`
	InputTokens         int       `gorm:"column:input_tokens"          json:"input_tokens"`
	OutputTokens        int       `gorm:"column:output_tokens"         json:"output_tokens"`
	CacheCreationTokens int       `gorm:"column:cache_creation_tokens" json:"cache_creation_tokens"`
	CacheReadTokens     int       `gorm:"column:cache_read_tokens"     json:"cache_read_tokens"`
	CachedPromptTokens  int       `gorm:"column:cached_prompt_tokens"  json:"cached_prompt_tokens"`
	CostMicrodollars    int64     `gorm:"column:cost_microdollars"     json:"cost_microdollars"`
	Status              string    `gorm:"column:status"                json:"status"`
	LatencyMs           *int      `gorm:"column:latency_ms"            json:"latency_ms"`
	RetryCount          int16     `gorm:"column:retry_count"           json:"retry_count"`
	ErrorMsg            *string   `gorm:"column:error_msg"             json:"error_msg"`
	CreatedAt           time.Time `gorm:"column:created_at"            json:"created_at"`
}

// UsageStats 聚合统计（用于前端图表）
type UsageStats struct {
	ProviderID          string  `gorm:"column:provider_id"           json:"provider_id"`
	ProviderName        string  `gorm:"column:name"                  json:"provider_name"`
	TotalRequests       int     `gorm:"column:total_requests"        json:"total_requests"`
	SuccessRequests     int     `gorm:"column:success_requests"      json:"success_requests"`
	InputTokens         int64   `gorm:"column:input_tokens"          json:"input_tokens"`
	OutputTokens        int64   `gorm:"column:output_tokens"         json:"output_tokens"`
	CacheCreationTokens int64   `gorm:"column:cache_creation_tokens" json:"cache_creation_tokens"`
	CacheReadTokens     int64   `gorm:"column:cache_read_tokens"     json:"cache_read_tokens"`
	TotalCostUSD        float64 `gorm:"column:total_cost_usd"        json:"total_cost_usd"`
}

// DailyUsage 按天聚合（折线图数据）
type DailyUsage struct {
	Date         string  `gorm:"column:date"          json:"date"`
	InputTokens  int64   `gorm:"column:input_tokens"  json:"input_tokens"`
	OutputTokens int64   `gorm:"column:output_tokens" json:"output_tokens"`
	CostUSD      float64 `gorm:"column:cost_usd"      json:"cost_usd"`
	Requests     int     `gorm:"column:requests"      json:"requests"`
}

// ===== Anthropic API 结构 =====

// AnthropicRequest /v1/messages 请求体
type AnthropicRequest struct {
	Model     string             `json:"model"`
	Messages  []AnthropicMessage `json:"messages"`
	MaxTokens int                `json:"max_tokens"`
	System    interface{}        `json:"system,omitempty"` // string or []ContentBlock
	Stream    bool               `json:"stream,omitempty"`
	// 扩展字段（thinking、tools 等）直接透传
}

type AnthropicMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // string or []ContentBlock
}

// AnthropicUsage 最新计费字段
type AnthropicUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

// ===== OpenAI API 结构 =====

// OpenAIUsage 包含 cached_tokens
type OpenAIUsage struct {
	PromptTokens            int                      `json:"prompt_tokens"`
	CompletionTokens        int                      `json:"completion_tokens"`
	TotalTokens             int                      `json:"total_tokens"`
	PromptTokensDetails     *OpenAIPromptDetails     `json:"prompt_tokens_details,omitempty"`
	CompletionTokensDetails *OpenAICompletionDetails `json:"completion_tokens_details,omitempty"`
}

type OpenAIPromptDetails struct {
	CachedTokens int `json:"cached_tokens"`
	AudioTokens  int `json:"audio_tokens"`
}

type OpenAICompletionDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
	AudioTokens     int `json:"audio_tokens"`
}
