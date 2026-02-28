package domain

import "time"

type TaskComment struct {
	ID        string    `gorm:"column:id"         json:"id"`
	TaskID    string    `gorm:"column:task_id"    json:"taskId"`
	CompanyID string    `gorm:"column:company_id" json:"companyId"`
	AgentID   string    `gorm:"column:agent_id"   json:"agentId"`
	Content   string    `gorm:"column:content"    json:"content"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
}

type TaskDependency struct {
	ID          string    `gorm:"column:id"            json:"id"`
	TaskID      string    `gorm:"column:task_id"       json:"taskId"`
	DependsOnID string    `gorm:"column:depends_on_id" json:"dependsOnId"`
	CompanyID   string    `gorm:"column:company_id"    json:"companyId"`
	CreatedAt   time.Time `gorm:"column:created_at"    json:"createdAt"`
}

type TaskWatcher struct {
	TaskID    string    `gorm:"column:task_id"    json:"taskId"`
	AgentID   string    `gorm:"column:agent_id"   json:"agentId"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
}
