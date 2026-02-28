package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/linkclaw/backend/internal/domain"
)

type codeIndexRepo struct {
	db *gorm.DB
}

func NewCodeIndexRepo(db *gorm.DB) CodeIndexRepo {
	return &codeIndexRepo{db: db}
}

func (r *codeIndexRepo) CreateChunk(ctx context.Context, c *domain.CodeChunk) error {
	result := r.db.WithContext(ctx).Exec(
		`INSERT INTO code_chunks (id, company_id, file_path, chunk_index, content, start_line, end_line, language, symbols, embedding_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		c.ID, c.CompanyID, c.FilePath, c.ChunkIndex, c.Content, c.StartLine, c.EndLine, c.Language, c.Symbols, c.EmbeddingID,
	)
	if result.Error != nil {
		return fmt.Errorf("code chunk create: %w", result.Error)
	}
	return nil
}

func (r *codeIndexRepo) GetChunksByFile(ctx context.Context, companyID, filePath string) ([]*domain.CodeChunk, error) {
	var chunks []*domain.CodeChunk
	result := r.db.WithContext(ctx).Raw(
		`SELECT id, company_id, file_path, chunk_index, content, start_line, end_line,
		        language, symbols, embedding_id, created_at, updated_at
		FROM code_chunks WHERE company_id = $1 AND file_path = $2
		ORDER BY chunk_index ASC`,
		companyID, filePath,
	).Scan(&chunks)
	if result.Error != nil {
		return nil, fmt.Errorf("code chunk get by file: %w", result.Error)
	}
	return chunks, nil
}

func (r *codeIndexRepo) DeleteByFile(ctx context.Context, companyID, filePath string) error {
	result := r.db.WithContext(ctx).Exec(
		`DELETE FROM code_chunks WHERE company_id = $1 AND file_path = $2`,
		companyID, filePath,
	)
	if result.Error != nil {
		return fmt.Errorf("code chunk delete by file: %w", result.Error)
	}
	return nil
}

func (r *codeIndexRepo) CreateIndexTask(ctx context.Context, t *domain.IndexTask) error {
	result := r.db.WithContext(ctx).Exec(
		`INSERT INTO index_tasks (id, company_id, repository_url, branch, status, total_files, indexed_files, error_message, started_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		t.ID, t.CompanyID, t.RepositoryURL, t.Branch, t.Status, t.TotalFiles, t.IndexedFiles, t.ErrorMessage, t.StartedAt, t.CompletedAt,
	)
	if result.Error != nil {
		return fmt.Errorf("index task create: %w", result.Error)
	}
	return nil
}

func (r *codeIndexRepo) GetIndexTask(ctx context.Context, id string) (*domain.IndexTask, error) {
	var task domain.IndexTask
	result := r.db.WithContext(ctx).Raw(
		`SELECT id, company_id, repository_url, branch, status, total_files, indexed_files,
		        error_message, started_at, completed_at, created_at
		FROM index_tasks WHERE id = $1`,
		id,
	).Scan(&task)
	if result.Error != nil {
		return nil, fmt.Errorf("index task get: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &task, nil
}

func (r *codeIndexRepo) UpdateIndexTask(ctx context.Context, t *domain.IndexTask) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE index_tasks
		SET repository_url=$1, branch=$2, status=$3, total_files=$4, indexed_files=$5,
		    error_message=$6, started_at=$7, completed_at=$8
		WHERE id=$9`,
		t.RepositoryURL, t.Branch, t.Status, t.TotalFiles, t.IndexedFiles,
		t.ErrorMessage, t.StartedAt, t.CompletedAt, t.ID,
	)
	if result.Error != nil {
		return fmt.Errorf("index task update: %w", result.Error)
	}
	return nil
}

func (r *codeIndexRepo) ListIndexTasks(ctx context.Context, companyID string) ([]*domain.IndexTask, error) {
	var tasks []*domain.IndexTask
	result := r.db.WithContext(ctx).Raw(
		`SELECT id, company_id, repository_url, branch, status, total_files, indexed_files,
		        error_message, started_at, completed_at, created_at
		FROM index_tasks WHERE company_id = $1
		ORDER BY created_at DESC`,
		companyID,
	).Scan(&tasks)
	if result.Error != nil {
		return nil, fmt.Errorf("index task list: %w", result.Error)
	}
	return tasks, nil
}
