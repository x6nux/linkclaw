package domain

import "time"

// MemoryImportance 记忆重要性等级
type MemoryImportance int

const (
	ImportanceCore      MemoryImportance = 0 // 核心，永不遗忘
	ImportanceImportant MemoryImportance = 1 // 重要
	ImportanceNormal    MemoryImportance = 2 // 普通
	ImportanceTrivial   MemoryImportance = 3 // 琐碎
	ImportanceEphemeral MemoryImportance = 4 // 临时
)

// MemorySource 记忆来源
type MemorySource string

const (
	SourceConversation MemorySource = "conversation"
	SourceManual       MemorySource = "manual"
	SourceSystem       MemorySource = "system"
)

// Memory Agent 记忆
type Memory struct {
	ID             string           `gorm:"column:id"               json:"id"`
	CompanyID      string           `gorm:"column:company_id"       json:"company_id"`
	AgentID        string           `gorm:"column:agent_id"         json:"agent_id"`
	Content        string           `gorm:"column:content"          json:"content"`
	Category       string           `gorm:"column:category"         json:"category"`
	Tags           StringList       `gorm:"column:tags"             json:"tags"`
	Importance     MemoryImportance `gorm:"column:importance"       json:"importance"`
	Source         MemorySource     `gorm:"column:source"           json:"source"`
	AccessCount    int              `gorm:"column:access_count"     json:"access_count"`
	LastAccessedAt *time.Time       `gorm:"column:last_accessed_at" json:"last_accessed_at"`
	CreatedAt      time.Time        `gorm:"column:created_at"       json:"created_at"`
	UpdatedAt      time.Time        `gorm:"column:updated_at"       json:"updated_at"`
}
