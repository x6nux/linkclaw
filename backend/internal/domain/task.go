package domain

import "time"

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusAssigned   TaskStatus = "assigned"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent TaskPriority = "urgent"
)

// ValidTransitions 任务状态机合法跃迁
var ValidTransitions = map[TaskStatus][]TaskStatus{
	TaskStatusPending:    {TaskStatusAssigned, TaskStatusCancelled},
	TaskStatusAssigned:   {TaskStatusInProgress, TaskStatusCancelled},
	TaskStatusInProgress: {TaskStatusDone, TaskStatusFailed, TaskStatusCancelled},
}

func (s TaskStatus) CanTransitionTo(next TaskStatus) bool {
	for _, allowed := range ValidTransitions[s] {
		if allowed == next {
			return true
		}
	}
	return false
}

type Task struct {
	ID          string       `gorm:"column:id"          json:"id"`
	CompanyID   string       `gorm:"column:company_id"  json:"companyId"`
	ParentID    *string      `gorm:"column:parent_id"   json:"parentId"`
	Title       string       `gorm:"column:title"       json:"title"`
	Description string       `gorm:"column:description" json:"description"`
	Priority    TaskPriority `gorm:"column:priority"    json:"priority"`
	Status      TaskStatus   `gorm:"column:status"      json:"status"`
	AssigneeID  *string      `gorm:"column:assignee_id" json:"assigneeId"`
	CreatedBy   *string      `gorm:"column:created_by"  json:"createdBy"`
	DueAt       *time.Time   `gorm:"column:due_at"      json:"dueAt"`
	Result      *string      `gorm:"column:result"      json:"result"`
	FailReason   *string           `gorm:"column:fail_reason" json:"failReason"`
	Tags         StringList        `gorm:"column:tags"        json:"tags"`
	Subtasks     []*Task           `gorm:"-"                  json:"subtasks"`
	Comments     []*TaskComment    `gorm:"-"                  json:"comments,omitempty"`
	Dependencies []*TaskDependency `gorm:"-"                  json:"dependencies,omitempty"`
	Watchers     []*TaskWatcher    `gorm:"-"                  json:"watchers,omitempty"`
	CreatedAt    time.Time         `gorm:"column:created_at"  json:"createdAt"`
	UpdatedAt    time.Time         `gorm:"column:updated_at"  json:"updatedAt"`
}
