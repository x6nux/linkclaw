package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/linkclaw/backend/internal/domain"
)

type knowledgeRepo struct {
	db *gorm.DB
}

func NewKnowledgeRepo(db *gorm.DB) KnowledgeRepo {
	return &knowledgeRepo{db: db}
}

func (r *knowledgeRepo) Create(ctx context.Context, d *domain.KnowledgeDoc) error {
	result := r.db.WithContext(ctx).Exec(
		`INSERT INTO knowledge_docs (id, company_id, title, content, tags, author_id)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		d.ID, d.CompanyID, d.Title, d.Content, d.Tags, d.AuthorID)
	if result.Error != nil {
		return fmt.Errorf("knowledge create: %w", result.Error)
	}
	return nil
}

func (r *knowledgeRepo) GetByID(ctx context.Context, id string) (*domain.KnowledgeDoc, error) {
	var doc domain.KnowledgeDoc
	result := r.db.WithContext(ctx).Raw(
		`SELECT * FROM knowledge_docs WHERE id = $1`, id,
	).Scan(&doc)
	if result.Error != nil {
		return nil, fmt.Errorf("knowledge get: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &doc, nil
}

func (r *knowledgeRepo) Update(ctx context.Context, d *domain.KnowledgeDoc) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE knowledge_docs SET title=$1, content=$2, tags=$3, updated_at=NOW() WHERE id=$4`,
		d.Title, d.Content, d.Tags, d.ID)
	return result.Error
}

func (r *knowledgeRepo) Search(ctx context.Context, companyID, query string, limit int) ([]*domain.KnowledgeDoc, error) {
	if limit <= 0 {
		limit = 20
	}
	var docs []*domain.KnowledgeDoc
	result := r.db.WithContext(ctx).Raw(
		`SELECT * FROM knowledge_docs
		WHERE company_id = $1
		  AND search_vec @@ plainto_tsquery('simple', $2)
		ORDER BY ts_rank(search_vec, plainto_tsquery('simple', $2)) DESC
		LIMIT $3`,
		companyID, query, limit,
	).Scan(&docs)
	if result.Error != nil {
		return nil, fmt.Errorf("knowledge search: %w", result.Error)
	}
	return docs, nil
}

func (r *knowledgeRepo) List(ctx context.Context, companyID string, limit, offset int) ([]*domain.KnowledgeDoc, int, error) {
	if limit <= 0 {
		limit = 20
	}
	var total int64
	if err := r.db.WithContext(ctx).Raw(
		`SELECT COUNT(*) FROM knowledge_docs WHERE company_id = $1`, companyID,
	).Scan(&total).Error; err != nil {
		return nil, 0, err
	}
	var docs []*domain.KnowledgeDoc
	if err := r.db.WithContext(ctx).Raw(
		`SELECT * FROM knowledge_docs WHERE company_id = $1
		ORDER BY updated_at DESC LIMIT $2 OFFSET $3`,
		companyID, limit, offset,
	).Scan(&docs).Error; err != nil {
		return nil, 0, fmt.Errorf("knowledge list: %w", err)
	}
	return docs, int(total), nil
}

func (r *knowledgeRepo) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Exec(`DELETE FROM knowledge_docs WHERE id = $1`, id)
	return result.Error
}
