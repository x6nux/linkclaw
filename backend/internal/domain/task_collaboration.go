package domain

import "time"

type TaskComment struct {
	ID        string    `gorm:"column:id"         json:"id"`
	TaskID    string    `gorm:"column:task_id"    json:"task_id"`
	CompanyID string    `gorm:"column:company_id" json:"company_id"`
	AgentID   string    `gorm:"column:agent_id"   json:"agent_id"`
	Content   string    `gorm:"column:content"    json:"content"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

type TaskDependency struct {
	ID          string    `gorm:"column:id"            json:"id"`
	TaskID      string    `gorm:"column:task_id"       json:"task_id"`
	DependsOnID string    `gorm:"column:depends_on_id" json:"depends_on_id"`
	CompanyID   string    `gorm:"column:company_id"    json:"company_id"`
	CreatedAt   time.Time `gorm:"column:created_at"    json:"created_at"`
}

type TaskWatcher struct {
	TaskID    string    `gorm:"column:task_id"    json:"task_id"`
	AgentID   string    `gorm:"column:agent_id"   json:"agent_id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}
