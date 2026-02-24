package domain

import "time"

// PromptLayerType 提示词层类型
type PromptLayerType string

const (
	PromptDepartment PromptLayerType = "department"
	PromptPosition   PromptLayerType = "position"
)

// PromptLayer 分层提示词（部门/职位层）
type PromptLayer struct {
	ID        string          `gorm:"column:id"         json:"id"`
	CompanyID string          `gorm:"column:company_id" json:"companyId"`
	Type      PromptLayerType `gorm:"column:type"       json:"type"`
	Key       string          `gorm:"column:key"        json:"key"`
	Content   string          `gorm:"column:content"    json:"content"`
	UpdatedAt time.Time       `gorm:"column:updated_at" json:"updatedAt"`
}
