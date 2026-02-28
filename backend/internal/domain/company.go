package domain

import "time"

type Company struct {
	ID           string    `gorm:"column:id"            json:"id"`
	Name         string    `gorm:"column:name"          json:"name"`
	Slug         string    `gorm:"column:slug"          json:"slug"`
	Description  string    `gorm:"column:description"   json:"description"`
	SystemPrompt string    `gorm:"column:system_prompt" json:"systemPrompt"`
	CreatedAt    time.Time `gorm:"column:created_at"    json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at"    json:"updatedAt"`

	// 系统设置（公司级别）
	PublicDomain      string `gorm:"column:public_domain"       json:"publicDomain"`
	AgentWSUrl        string `gorm:"column:agent_ws_url"        json:"agentWsUrl"`
	MCPPublicURL      string `gorm:"column:mcp_public_url"      json:"mcpPublicUrl"`
	NanoclawImage     string `gorm:"column:nanoclaw_image"      json:"nanoclawImage"`
	OpenclawPluginURL string `gorm:"column:openclaw_plugin_url" json:"openclawPluginUrl"`
	EmbeddingBaseURL  string `gorm:"column:embedding_base_url"  json:"embeddingBaseUrl"`
	EmbeddingModel    string `gorm:"column:embedding_model"     json:"embeddingModel"`
	EmbeddingApiKey   string `gorm:"column:embedding_api_key"   json:"embeddingApiKey"`
}

// CompanySettings 系统设置（用于 API 读写）
type CompanySettings struct {
	PublicDomain      string `json:"publicDomain"`
	AgentWSUrl        string `json:"agentWsUrl"`
	MCPPublicURL      string `json:"mcpPublicUrl"`
	NanoclawImage     string `json:"nanoclawImage"`
	OpenclawPluginURL string `json:"openclawPluginUrl"`
	EmbeddingBaseURL  string `json:"embeddingBaseUrl"`
	EmbeddingModel    string `json:"embeddingModel"`
	EmbeddingApiKey   string `json:"embeddingApiKey"`
}

type Channel struct {
	ID          string    `gorm:"column:id"          json:"id"`
	CompanyID   string    `gorm:"column:company_id"  json:"companyId"`
	Name        string    `gorm:"column:name"        json:"name"`
	Description string    `gorm:"column:description" json:"description"`
	IsDefault   bool      `gorm:"column:is_default"  json:"isDefault"`
	CreatedAt   time.Time `gorm:"column:created_at"  json:"createdAt"`
}
