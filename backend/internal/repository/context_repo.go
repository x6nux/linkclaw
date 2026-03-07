package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/linkclaw/backend/internal/domain"
)

type contextRepo struct {
	db *gorm.DB
}

func NewContextRepo(db *gorm.DB) ContextRepo {
	return &contextRepo{db: db}
}

// ── 目录管理 ─────────────────────────────────────────

func (r *contextRepo) CreateDirectory(ctx context.Context, d *domain.ContextDirectory) error {
	result := r.db.WithContext(ctx).Exec(
		`INSERT INTO context_directories (id, company_id, name, path, description, is_active, file_patterns, exclude_patterns, max_file_size, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		d.ID, d.CompanyID, d.Name, d.Path, d.Description, d.IsActive, d.FilePatterns, d.ExcludePatterns, d.MaxFileSize, d.CreatedAt, d.UpdatedAt)
	return result.Error
}

func (r *contextRepo) GetDirectoryByID(ctx context.Context, id string) (*domain.ContextDirectory, error) {
	var d domain.ContextDirectory
	result := r.db.WithContext(ctx).Raw(
		`SELECT id, company_id, name, path, description, is_active, file_patterns, exclude_patterns, max_file_size, last_indexed_at, file_count, created_at, updated_at
		FROM context_directories WHERE id = $1`, id).Scan(&d)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &d, nil
}

func (r *contextRepo) ListDirectories(ctx context.Context, companyID string) ([]*domain.ContextDirectory, error) {
	var dirs []*domain.ContextDirectory
	result := r.db.WithContext(ctx).Raw(
		`SELECT id, company_id, name, path, description, is_active, file_patterns, exclude_patterns, max_file_size, last_indexed_at, file_count, created_at, updated_at
		FROM context_directories WHERE company_id = $1 ORDER BY name`, companyID).Scan(&dirs)
	return dirs, result.Error
}

func (r *contextRepo) ListActiveDirectories(ctx context.Context, companyID string) ([]*domain.ContextDirectory, error) {
	var dirs []*domain.ContextDirectory
	result := r.db.WithContext(ctx).Raw(
		`SELECT id, company_id, name, path, description, is_active, file_patterns, exclude_patterns, max_file_size, last_indexed_at, file_count, created_at, updated_at
		FROM context_directories WHERE company_id = $1 AND is_active = TRUE ORDER BY name`, companyID).Scan(&dirs)
	return dirs, result.Error
}

// ListAllActiveDirectories 获取所有公司的活跃目录（用于后台任务）
func (r *contextRepo) ListAllActiveDirectories(ctx context.Context) ([]*domain.ContextDirectory, error) {
	var dirs []*domain.ContextDirectory
	result := r.db.WithContext(ctx).Raw(
		`SELECT id, company_id, name, path, description, is_active, file_patterns, exclude_patterns, max_file_size, last_indexed_at, file_count, created_at, updated_at
		FROM context_directories WHERE is_active = TRUE ORDER BY company_id, name`).Scan(&dirs)
	return dirs, result.Error
}

func (r *contextRepo) UpdateDirectory(ctx context.Context, d *domain.ContextDirectory) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE context_directories SET name=$1, path=$2, description=$3, is_active=$4, file_patterns=$5, exclude_patterns=$6, max_file_size=$7, last_indexed_at=$8, file_count=$9, updated_at=$10 WHERE id=$11`,
		d.Name, d.Path, d.Description, d.IsActive, d.FilePatterns, d.ExcludePatterns, d.MaxFileSize, d.LastIndexedAt, d.FileCount, d.UpdatedAt, d.ID)
	return result.Error
}

func (r *contextRepo) DeleteDirectory(ctx context.Context, id string) error {
	// 先删除关联的文件总结
	if err := r.DeleteFileSummariesByDirectory(ctx, id); err != nil {
		return fmt.Errorf("delete file summaries: %w", err)
	}
	result := r.db.WithContext(ctx).Exec(`DELETE FROM context_directories WHERE id = $1`, id)
	return result.Error
}

// ── 文件总结 ─────────────────────────────────────────

func (r *contextRepo) CreateFileSummary(ctx context.Context, s *domain.ContextFileSummary) error {
	result := r.db.WithContext(ctx).Exec(
		`INSERT INTO context_file_summaries (id, directory_id, file_path, content_hash, summary, language, line_count, summarized_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		s.ID, s.DirectoryID, s.FilePath, s.ContentHash, s.Summary, s.Language, s.LineCount, s.SummarizedAt)
	return result.Error
}

func (r *contextRepo) GetFileSummary(ctx context.Context, directoryID, filePath string) (*domain.ContextFileSummary, error) {
	var s domain.ContextFileSummary
	result := r.db.WithContext(ctx).Raw(
		`SELECT id, directory_id, file_path, content_hash, summary, language, line_count, summarized_at
		FROM context_file_summaries WHERE directory_id = $1 AND file_path = $2`, directoryID, filePath).Scan(&s)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &s, nil
}

func (r *contextRepo) GetFileSummaryByHash(ctx context.Context, contentHash string) (*domain.ContextFileSummary, error) {
	var s domain.ContextFileSummary
	result := r.db.WithContext(ctx).Raw(
		`SELECT id, directory_id, file_path, content_hash, summary, language, line_count, summarized_at
		FROM context_file_summaries WHERE content_hash = $1 LIMIT 1`, contentHash).Scan(&s)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &s, nil
}

func (r *contextRepo) ListFileSummaries(ctx context.Context, directoryID string) ([]*domain.ContextFileSummary, error) {
	var summaries []*domain.ContextFileSummary
	result := r.db.WithContext(ctx).Raw(
		`SELECT id, directory_id, file_path, content_hash, summary, language, line_count, summarized_at
		FROM context_file_summaries WHERE directory_id = $1 ORDER BY file_path`, directoryID).Scan(&summaries)
	return summaries, result.Error
}

func (r *contextRepo) DeleteFileSummary(ctx context.Context, directoryID, filePath string) error {
	result := r.db.WithContext(ctx).Exec(
		`DELETE FROM context_file_summaries WHERE directory_id = $1 AND file_path = $2`, directoryID, filePath)
	return result.Error
}

func (r *contextRepo) DeleteFileSummariesByDirectory(ctx context.Context, directoryID string) error {
	result := r.db.WithContext(ctx).Exec(
		`DELETE FROM context_file_summaries WHERE directory_id = $1`, directoryID)
	return result.Error
}

// ── 搜索日志 ─────────────────────────────────────────

func (r *contextRepo) CreateSearchLog(ctx context.Context, log *domain.ContextSearchLog) error {
	result := r.db.WithContext(ctx).Exec(
		`INSERT INTO context_search_logs (id, company_id, agent_id, query, directory_ids, results_count, latency_ms, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		log.ID, log.CompanyID, log.AgentID, log.Query, log.DirectoryIDs, log.ResultsCount, log.LatencyMs, log.CreatedAt)
	return result.Error
}
