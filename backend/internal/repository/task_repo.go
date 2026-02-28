package repository

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/linkclaw/backend/internal/domain"
)

type taskRepo struct {
	db *gorm.DB
}

func NewTaskRepo(db *gorm.DB) TaskRepo {
	return &taskRepo{db: db}
}

func (r *taskRepo) Create(ctx context.Context, t *domain.Task) error {
	q := `INSERT INTO tasks
		(id, company_id, parent_id, title, description, priority, status, assignee_id, created_by, due_at, tags)
		VALUES
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	result := r.db.WithContext(ctx).Exec(q,
		t.ID, t.CompanyID, t.ParentID, t.Title, t.Description,
		string(t.Priority), string(t.Status), t.AssigneeID, t.CreatedBy, t.DueAt, t.Tags)
	if result.Error != nil {
		return fmt.Errorf("task create: %w", result.Error)
	}
	return nil
}

func (r *taskRepo) CreateAttachments(ctx context.Context, attachments []*domain.TaskAttachment) error {
	if len(attachments) == 0 {
		return nil
	}

	q := `INSERT INTO task_attachments
		(id, task_id, company_id, filename, original_filename, file_size, mime_type, storage_path, uploaded_by, created_at)
		VALUES
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	for _, a := range attachments {
		if err := r.db.WithContext(ctx).Exec(
			q,
			a.ID,
			a.TaskID,
			a.CompanyID,
			a.Filename,
			a.OriginalFilename,
			a.FileSize,
			a.MimeType,
			a.StoragePath,
			a.UploadedBy,
			a.CreatedAt,
		).Error; err != nil {
			return fmt.Errorf("task attachment create: %w", err)
		}
	}

	return nil
}

func (r *taskRepo) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	var t domain.Task
	result := r.db.WithContext(ctx).Raw(`SELECT * FROM tasks WHERE id = $1`, id).Scan(&t)
	if result.Error != nil {
		return nil, fmt.Errorf("task get: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	var subtasks []*domain.Task
	r.db.WithContext(ctx).Raw(
		`SELECT * FROM tasks WHERE parent_id = $1 ORDER BY created_at`, id,
	).Scan(&subtasks)
	t.Subtasks = subtasks

	var attachments []*domain.TaskAttachment
	if err := r.db.WithContext(ctx).Raw(
		`SELECT * FROM task_attachments WHERE task_id = $1 ORDER BY created_at`, id,
	).Scan(&attachments).Error; err != nil {
		return nil, fmt.Errorf("task attachments get: %w", err)
	}
	t.Attachments = attachments

	return &t, nil
}

func (r *taskRepo) List(ctx context.Context, q TaskQuery) ([]*domain.Task, int, error) {
	where := []string{"company_id = $1"}
	args := []interface{}{q.CompanyID}
	idx := 2

	if q.AssigneeID != "" {
		where = append(where, fmt.Sprintf("assignee_id = $%d", idx))
		args = append(args, q.AssigneeID)
		idx++
	}
	if q.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", idx))
		args = append(args, string(q.Status))
		idx++
	}
	if q.Priority != "" {
		where = append(where, fmt.Sprintf("priority = $%d", idx))
		args = append(args, string(q.Priority))
		idx++
	}
	if q.ParentID == nil {
		where = append(where, "parent_id IS NULL")
	}

	whereClause := strings.Join(where, " AND ")
	var total int64
	if err := r.db.WithContext(ctx).Raw(
		fmt.Sprintf("SELECT COUNT(*) FROM tasks WHERE %s", whereClause), args...,
	).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	listArgs := append(args, limit, q.Offset)
	listQ := fmt.Sprintf(
		"SELECT * FROM tasks WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		whereClause, idx, idx+1,
	)

	var tasks []*domain.Task
	if err := r.db.WithContext(ctx).Raw(listQ, listArgs...).Scan(&tasks).Error; err != nil {
		return nil, 0, fmt.Errorf("task list: %w", err)
	}
	return tasks, int(total), nil
}

func (r *taskRepo) UpdateStatus(ctx context.Context, id string, status domain.TaskStatus, result, failReason *string) error {
	res := r.db.WithContext(ctx).Exec(
		`UPDATE tasks SET status = $1, result = $2, fail_reason = $3, updated_at = NOW() WHERE id = $4`,
		status, result, failReason, id)
	return res.Error
}

func (r *taskRepo) UpdateAssignee(ctx context.Context, id, assigneeID string, status domain.TaskStatus) error {
	res := r.db.WithContext(ctx).Exec(
		`UPDATE tasks SET assignee_id = $1, status = $2, updated_at = NOW() WHERE id = $3`,
		assigneeID, status, id)
	return res.Error
}

func (r *taskRepo) UpdateTags(ctx context.Context, id string, tags domain.StringList) error {
	res := r.db.WithContext(ctx).Exec(
		`UPDATE tasks SET tags = $1, updated_at = NOW() WHERE id = $2`, tags, id)
	return res.Error
}

func (r *taskRepo) Delete(ctx context.Context, id, companyID string) error {
	res := r.db.WithContext(ctx).Exec(`DELETE FROM tasks WHERE id = $1 AND company_id = $2`, id, companyID)
	return res.Error
}
