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
	CompanyID   string    `gorm:"column:company_id"   json:"companyId"`
	FilePath    string    `gorm:"column:file_path"    json:"filePath"`
	ChunkIndex  int       `gorm:"column:chunk_index"  json:"chunkIndex"`
	Content     string    `gorm:"column:content"      json:"content"`
	StartLine   int       `gorm:"column:start_line"   json:"startLine"`
	EndLine     int       `gorm:"column:end_line"     json:"endLine"`
	Language    string    `gorm:"column:language"     json:"language"`
	Symbols     string    `gorm:"column:symbols"      json:"symbols"`
	EmbeddingID string    `gorm:"column:embedding_id" json:"embeddingId"`
	CreatedAt   time.Time `gorm:"column:created_at"   json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at"   json:"updatedAt"`
}

// IndexTask 索引任务
type IndexTask struct {
	ID            string      `gorm:"column:id"             json:"id"`
	CompanyID     string      `gorm:"column:company_id"     json:"companyId"`
	RepositoryURL string      `gorm:"column:repository_url" json:"repositoryUrl"`
	Branch        string      `gorm:"column:branch"         json:"branch"`
	Status        IndexStatus `gorm:"column:status"         json:"status"`
	TotalFiles    int         `gorm:"column:total_files"    json:"totalFiles"`
	IndexedFiles  int         `gorm:"column:indexed_files"  json:"indexedFiles"`
	ErrorMessage  string      `gorm:"column:error_message"  json:"errorMessage"`
	StartedAt     *time.Time  `gorm:"column:started_at"     json:"startedAt"`
	CompletedAt   *time.Time  `gorm:"column:completed_at"   json:"completedAt"`
	CreatedAt     time.Time   `gorm:"column:created_at"     json:"createdAt"`
}
