package domain

import "time"

type KnowledgeDoc struct {
	ID        string    `gorm:"column:id"         json:"id"`
	CompanyID string    `gorm:"column:company_id" json:"company_id"`
	Title     string    `gorm:"column:title"      json:"title"`
	Content   string    `gorm:"column:content"    json:"content"`
	Tags      StringList `gorm:"column:tags"      json:"tags"`
	AuthorID  *string   `gorm:"column:author_id"  json:"author_id"`
	// SearchVec 由数据库触发器自动维护，不需要应用层写入
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}
