package domain

import "time"

type Company struct {
	ID           string    `gorm:"column:id"            json:"id"`
	Name         string    `gorm:"column:name"          json:"name"`
	Slug         string    `gorm:"column:slug"          json:"slug"`
	Description  string    `gorm:"column:description"   json:"description"`
	SystemPrompt string    `gorm:"column:system_prompt" json:"system_prompt"`
	CreatedAt    time.Time `gorm:"column:created_at"    json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"    json:"updated_at"`

	// 系统设置（公司级别）
	PublicDomain      string `gorm:"column:public_domain"       json:"public_domain"`
	AgentWSUrl        string `gorm:"column:agent_ws_url"        json:"agent_ws_url"`
	MCPPublicURL      string `gorm:"column:mcp_public_url"      json:"mcp_public_url"`
	MCPPrivateURL     string `gorm:"column:mcp_private_url"     json:"mcp_private_url"`
	NanoclawImage     string `gorm:"column:nanoclaw_image"      json:"nanoclaw_image"`
	OpenclawPluginURL string `gorm:"column:openclaw_plugin_url" json:"openclaw_plugin_url"`
	EmbeddingBaseURL  string `gorm:"column:embedding_base_url"  json:"embedding_base_url"`
	EmbeddingModel    string `gorm:"column:embedding_model"     json:"embedding_model"`
	EmbeddingApiKey   string `gorm:"column:embedding_api_key"   json:"embedding_api_key"`
}

// CompanySettings 系统设置（用于 API 读写）
type CompanySettings struct {
	PublicDomain      string `json:"public_domain"`
	AgentWSUrl        string `json:"agent_ws_url"`
	MCPPublicURL      string `json:"mcp_public_url"`
	MCPPrivateURL     string `json:"mcp_private_url"`
	NanoclawImage     string `json:"nanoclaw_image"`
	OpenclawPluginURL string `json:"openclaw_plugin_url"`
	EmbeddingBaseURL  string `json:"embedding_base_url"`
	EmbeddingModel    string `json:"embedding_model"`
	EmbeddingApiKey   string `json:"embedding_api_key"`
}

type Channel struct {
	ID          string    `gorm:"column:id"          json:"id"`
	CompanyID   string    `gorm:"column:company_id"  json:"company_id"`
	Name        string    `gorm:"column:name"        json:"name"`
	Description string    `gorm:"column:description" json:"description"`
	IsDefault   bool      `gorm:"column:is_default"  json:"is_default"`
	CreatedAt   time.Time `gorm:"column:created_at"  json:"created_at"`
}

// PartnerApiKey 公司间配对 API 密钥
type PartnerApiKey struct {
	ID           string     `gorm:"column:id"            json:"id"`
	CompanyID    string     `gorm:"column:company_id"    json:"company_id"`
	PartnerSlug  string     `gorm:"column:partner_slug"  json:"partner_slug"`
	PartnerID    *string    `gorm:"column:partner_id"    json:"partner_id"`
	Name         *string    `gorm:"column:name"          json:"name"`
	KeyHash      string     `gorm:"column:key_hash"      json:"-"`
	KeyPrefix    string     `gorm:"column:key_prefix"    json:"key_prefix"`
	IsActive     bool       `gorm:"column:is_active"     json:"is_active"`
	LastUsedAt   *time.Time `gorm:"column:last_used_at"  json:"last_used_at"`
	CreatedAt    time.Time  `gorm:"column:created_at"    json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"    json:"updated_at"`
}
