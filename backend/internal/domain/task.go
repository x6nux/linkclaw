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
	CompanyID   string       `gorm:"column:company_id"  json:"company_id"`
	ParentID    *string      `gorm:"column:parent_id"   json:"parent_id"`
	Title       string       `gorm:"column:title"       json:"title"`
	Description string       `gorm:"column:description" json:"description"`
	Priority    TaskPriority `gorm:"column:priority"    json:"priority"`
	Status      TaskStatus   `gorm:"column:status"      json:"status"`
	AssigneeID  *string      `gorm:"column:assignee_id" json:"assignee_id"`
	CreatedBy   *string      `gorm:"column:created_by"  json:"created_by"`
	DueAt       *time.Time   `gorm:"column:due_at"      json:"due_at"`
	Result      *string      `gorm:"column:result"      json:"result"`
	FailReason  *string      `gorm:"column:fail_reason" json:"fail_reason"`
	Tags        StringList   `gorm:"column:tags"        json:"tags"`

	Subtasks     []*Task           `gorm:"-" json:"subtasks"`
	Comments     []*TaskComment    `gorm:"-" json:"comments,omitempty"`
	Dependencies []*TaskDependency `gorm:"-" json:"dependencies,omitempty"`
	Watchers     []*TaskWatcher    `gorm:"-" json:"watchers,omitempty"`
	Attachments  []*TaskAttachment `gorm:"-" json:"attachments,omitempty"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

type TaskAttachment struct {
	ID               string    `gorm:"column:id"                json:"id"`
	TaskID           string    `gorm:"column:task_id"           json:"task_id"`
	CompanyID        string    `gorm:"column:company_id"        json:"company_id"`
	Filename         string    `gorm:"column:filename"          json:"filename"`
	OriginalFilename string    `gorm:"column:original_filename" json:"original_filename"`
	FileSize         int64     `gorm:"column:file_size"         json:"file_size"`
	MimeType         string    `gorm:"column:mime_type"         json:"mime_type"`
	StoragePath      string    `gorm:"column:storage_path"      json:"storage_path"`
	UploadedBy       *string   `gorm:"column:uploaded_by"       json:"uploaded_by"`
	CreatedAt        time.Time `gorm:"column:created_at"        json:"created_at"`
}
