package domain

import "time"

type IndexStatus string

const (
	IndexStatusPending   IndexStatus = "pending"
	IndexStatusRunning   IndexStatus = "running"
	IndexStatusCompleted IndexStatus = "completed"
	IndexStatusFailed    IndexStatus = "failed"
)

// CodeChunk 代码块
type CodeChunk struct {
	ID          string    `gorm:"column:id"           json:"id"`
	CompanyID   string    `gorm:"column:company_id"   json:"company_id"`
	FilePath    string    `gorm:"column:file_path"    json:"file_path"`
	ChunkIndex  int       `gorm:"column:chunk_index"  json:"chunk_index"`
	Content     string    `gorm:"column:content"      json:"content"`
	StartLine   int       `gorm:"column:start_line"   json:"start_line"`
	EndLine     int       `gorm:"column:end_line"     json:"end_line"`
	Language    string    `gorm:"column:language"     json:"language"`
	Symbols     string    `gorm:"column:symbols"      json:"symbols"`
	EmbeddingID string    `gorm:"column:embedding_id" json:"embedding_id"`
	CreatedAt   time.Time `gorm:"column:created_at"   json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"   json:"updated_at"`
}

// IndexTask 索引任务
type IndexTask struct {
	ID            string      `gorm:"column:id"             json:"id"`
	CompanyID     string      `gorm:"column:company_id"     json:"company_id"`
	RepositoryURL string      `gorm:"column:repository_url" json:"repository_url"`
	Branch        string      `gorm:"column:branch"         json:"branch"`
	Status        IndexStatus `gorm:"column:status"         json:"status"`
	TotalFiles    int         `gorm:"column:total_files"    json:"total_files"`
	IndexedFiles  int         `gorm:"column:indexed_files"  json:"indexed_files"`
	ErrorMessage  string      `gorm:"column:error_message"  json:"error_message"`
	StartedAt     *time.Time  `gorm:"column:started_at"     json:"started_at"`
	CompletedAt   *time.Time  `gorm:"column:completed_at"   json:"completed_at"`
	CreatedBy     *string     `gorm:"column:created_by"     json:"created_by,omitempty"`
	CreatedAt     time.Time   `gorm:"column:created_at"     json:"created_at"`

	AuthorizedAgents []*IndexTaskAgent `gorm:"-" json:"authorized_agents,omitempty"`
}

// IndexTaskAgent 索引任务授权 Agent
type IndexTaskAgent struct {
	ID          string    `gorm:"column:id"            json:"id"`
	IndexTaskID string    `gorm:"column:index_task_id" json:"index_task_id"`
	AgentID     string    `gorm:"column:agent_id"      json:"agent_id"`
	CompanyID   string    `gorm:"column:company_id"    json:"company_id"`
	CreatedAt   time.Time `gorm:"column:created_at"    json:"created_at"`
}
