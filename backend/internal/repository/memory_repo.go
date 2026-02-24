package repository

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/linkclaw/backend/internal/domain"
)

type memoryRepo struct {
	db *gorm.DB
}

// NewMemoryRepo 创建 MemoryRepo 实例
func NewMemoryRepo(db *gorm.DB) MemoryRepo {
	return &memoryRepo{db: db}
}

func (r *memoryRepo) Create(ctx context.Context, m *domain.Memory) error {
	result := r.db.WithContext(ctx).Exec(
		`INSERT INTO agent_memories (id, company_id, agent_id, content, category, tags, importance, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		m.ID, m.CompanyID, m.AgentID, m.Content, m.Category, m.Tags, m.Importance, m.Source)
	if result.Error != nil {
		return fmt.Errorf("memory create: %w", result.Error)
	}
	return nil
}

func (r *memoryRepo) GetByID(ctx context.Context, id string) (*domain.Memory, error) {
	var m domain.Memory
	result := r.db.WithContext(ctx).Raw(
		`SELECT id, company_id, agent_id, content, category, tags, importance, source,
		        access_count, last_accessed_at, created_at, updated_at
		FROM agent_memories WHERE id = $1`, id,
	).Scan(&m)
	if result.Error != nil {
		return nil, fmt.Errorf("memory get: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &m, nil
}

func (r *memoryRepo) Update(ctx context.Context, m *domain.Memory) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE agent_memories SET content=$1, category=$2, tags=$3, importance=$4, embedding=NULL
		WHERE id=$5`,
		m.Content, m.Category, m.Tags, m.Importance, m.ID)
	return result.Error
}

func (r *memoryRepo) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Exec(`DELETE FROM agent_memories WHERE id = $1`, id)
	return result.Error
}

func (r *memoryRepo) List(ctx context.Context, q MemoryQuery) ([]*domain.Memory, int, error) {
	if q.Limit <= 0 {
		q.Limit = 20
	}

	where, args := r.buildWhere(q)

	var total int64
	if err := r.db.WithContext(ctx).Raw(
		`SELECT COUNT(*) FROM agent_memories`+where, args...,
	).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	orderBy := "created_at DESC"
	switch q.OrderBy {
	case "importance":
		orderBy = "importance ASC, created_at DESC"
	case "access_count":
		orderBy = "access_count DESC, created_at DESC"
	}

	var mems []*domain.Memory
	if err := r.db.WithContext(ctx).Raw(
		`SELECT id, company_id, agent_id, content, category, tags, importance, source,
		        access_count, last_accessed_at, created_at, updated_at
		FROM agent_memories`+where+
			` ORDER BY `+orderBy+
			` LIMIT $`+fmt.Sprint(len(args)+1)+` OFFSET $`+fmt.Sprint(len(args)+2),
		append(args, q.Limit, q.Offset)...,
	).Scan(&mems).Error; err != nil {
		return nil, 0, fmt.Errorf("memory list: %w", err)
	}
	return mems, int(total), nil
}

func (r *memoryRepo) buildWhere(q MemoryQuery) (string, []any) {
	var conds []string
	var args []any
	n := 0

	if q.CompanyID != "" {
		n++
		conds = append(conds, fmt.Sprintf("company_id=$%d", n))
		args = append(args, q.CompanyID)
	}
	if q.AgentID != "" {
		n++
		conds = append(conds, fmt.Sprintf("agent_id=$%d", n))
		args = append(args, q.AgentID)
	}
	if q.Category != "" {
		n++
		conds = append(conds, fmt.Sprintf("category=$%d", n))
		args = append(args, q.Category)
	}
	if q.Importance != nil {
		n++
		conds = append(conds, fmt.Sprintf("importance=$%d", n))
		args = append(args, *q.Importance)
	}
	if len(conds) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(conds, " AND "), args
}

func (r *memoryRepo) SemanticSearch(ctx context.Context, companyID, agentID string, embedding []float32, limit, minImportance int) ([]*domain.Memory, error) {
	if limit <= 0 {
		limit = 10
	}
	vecStr := float32SliceToVec(embedding)

	var mems []*domain.Memory
	if err := r.db.WithContext(ctx).Raw(
		`SELECT id, company_id, agent_id, content, category, tags, importance, source,
		        access_count, last_accessed_at, created_at, updated_at
		FROM agent_memories
		WHERE company_id=$1 AND agent_id=$2 AND embedding IS NOT NULL AND importance<=$3
		ORDER BY
			(1 - (embedding <=> $4::vector))
			* (1.0 + (4 - importance) * 0.15)
			* (1.0 / (1 + EXTRACT(EPOCH FROM NOW()-created_at) / 86400 * 0.01))
			DESC
		LIMIT $5`,
		companyID, agentID, minImportance, vecStr, limit,
	).Scan(&mems).Error; err != nil {
		return nil, fmt.Errorf("memory semantic search: %w", err)
	}
	return mems, nil
}

func (r *memoryRepo) ListPendingEmbedding(ctx context.Context, limit int) ([]*domain.Memory, error) {
	if limit <= 0 {
		limit = 10
	}
	var mems []*domain.Memory
	if err := r.db.WithContext(ctx).Raw(
		`SELECT id, company_id, agent_id, content, category, tags, importance, source,
		        access_count, last_accessed_at, created_at, updated_at
		FROM agent_memories WHERE embedding IS NULL
		ORDER BY created_at ASC LIMIT $1`, limit,
	).Scan(&mems).Error; err != nil {
		return nil, fmt.Errorf("memory list pending: %w", err)
	}
	return mems, nil
}

func (r *memoryRepo) UpdateEmbedding(ctx context.Context, id string, embedding []float32) error {
	vecStr := float32SliceToVec(embedding)
	result := r.db.WithContext(ctx).Exec(
		`UPDATE agent_memories SET embedding=$1::vector WHERE id=$2`,
		vecStr, id)
	return result.Error
}

func (r *memoryRepo) IncrementAccess(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	result := r.db.WithContext(ctx).Exec(
		`UPDATE agent_memories SET access_count = access_count + 1, last_accessed_at = NOW()
		WHERE id IN (`+strings.Join(placeholders, ",")+`)`, args...)
	return result.Error
}

func (r *memoryRepo) BatchDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	result := r.db.WithContext(ctx).Exec(
		`DELETE FROM agent_memories WHERE id IN (`+strings.Join(placeholders, ",")+`)`, args...)
	return result.Error
}

// float32SliceToVec 将 []float32 转为 pgvector 字符串 "[0.1,0.2,...]"
func float32SliceToVec(v []float32) string {
	var b strings.Builder
	b.WriteByte('[')
	for i, f := range v {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(fmt.Sprintf("%g", f))
	}
	b.WriteByte(']')
	return b.String()
}
