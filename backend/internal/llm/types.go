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
	CompanyID   string       `gorm:"column:company_id"    json:"companyId"`
	Name        string       `gorm:"column:name"          json:"name"`
	Type        ProviderType `gorm:"column:provider_type" json:"type"`
	BaseURL     string       `gorm:"column:base_url"      json:"baseUrl"`
	APIKeyEnc   string            `gorm:"column:api_key_enc"   json:"-"`
	Models      domain.StringList `gorm:"column:models"        json:"models"`
	Weight      int               `gorm:"column:weight"        json:"weight"`
	IsActive    bool         `gorm:"column:is_active"     json:"isActive"`
	ErrorCount  int          `gorm:"column:error_count"   json:"errorCount"`
	LastErrorAt *time.Time   `gorm:"column:last_error_at" json:"lastErrorAt"`
	LastUsedAt  *time.Time   `gorm:"column:last_used_at"  json:"lastUsedAt"`
	MaxRPM      *int         `gorm:"column:max_rpm"       json:"maxRpm"`
	CreatedAt   time.Time    `gorm:"column:created_at"    json:"createdAt"`
	UpdatedAt   time.Time    `gorm:"column:updated_at"    json:"updatedAt"`
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
	CompanyID           string    `gorm:"column:company_id"            json:"companyId"`
	ProviderID          *string   `gorm:"column:provider_id"           json:"providerId"`
	AgentID             *string   `gorm:"column:agent_id"              json:"agentId"`
	RequestModel        string    `gorm:"column:request_model"         json:"requestModel"`
	InputTokens         int       `gorm:"column:input_tokens"          json:"inputTokens"`
	OutputTokens        int       `gorm:"column:output_tokens"         json:"outputTokens"`
	CacheCreationTokens int       `gorm:"column:cache_creation_tokens" json:"cacheCreationTokens"`
	CacheReadTokens     int       `gorm:"column:cache_read_tokens"     json:"cacheReadTokens"`
	CachedPromptTokens  int       `gorm:"column:cached_prompt_tokens"  json:"cachedPromptTokens"`
	CostMicrodollars    int64     `gorm:"column:cost_microdollars"     json:"costMicrodollars"`
	Status              string    `gorm:"column:status"                json:"status"`
	LatencyMs           *int      `gorm:"column:latency_ms"            json:"latencyMs"`
	RetryCount          int16     `gorm:"column:retry_count"           json:"retryCount"`
	ErrorMsg            *string   `gorm:"column:error_msg"             json:"errorMsg"`
	CreatedAt           time.Time `gorm:"column:created_at"            json:"createdAt"`
}

// UsageStats 聚合统计（用于前端图表）
type UsageStats struct {
	ProviderID          string  `gorm:"column:provider_id"           json:"providerId"`
	ProviderName        string  `gorm:"column:name"                  json:"providerName"`
	TotalRequests       int     `gorm:"column:total_requests"        json:"totalRequests"`
	SuccessRequests     int     `gorm:"column:success_requests"      json:"successRequests"`
	InputTokens         int64   `gorm:"column:input_tokens"          json:"inputTokens"`
	OutputTokens        int64   `gorm:"column:output_tokens"         json:"outputTokens"`
	CacheCreationTokens int64   `gorm:"column:cache_creation_tokens" json:"cacheCreationTokens"`
	CacheReadTokens     int64   `gorm:"column:cache_read_tokens"     json:"cacheReadTokens"`
	TotalCostUSD        float64 `gorm:"column:total_cost_usd"        json:"totalCostUsd"`
}

// DailyUsage 按天聚合（折线图数据）
type DailyUsage struct {
	Date         string  `gorm:"column:date"          json:"date"`
	InputTokens  int64   `gorm:"column:input_tokens"  json:"inputTokens"`
	OutputTokens int64   `gorm:"column:output_tokens" json:"outputTokens"`
	CostUSD      float64 `gorm:"column:cost_usd"      json:"costUsd"`
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
