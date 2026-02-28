package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/linkclaw/backend/internal/domain"
)

type taskCollabRepo struct {
	db *gorm.DB
}

func NewTaskCollabRepo(db *gorm.DB) TaskCollabRepo {
	return &taskCollabRepo{db: db}
}

func (r *taskCollabRepo) AddComment(ctx context.Context, c *domain.TaskComment) error {
	res := r.db.WithContext(ctx).Exec(
		`INSERT INTO task_comments (id, task_id, company_id, agent_id, content) VALUES ($1, $2, $3, $4, $5)`,
		c.ID, c.TaskID, c.CompanyID, c.AgentID, c.Content,
	)
	if res.Error != nil {
		return fmt.Errorf("task comment add: %w", res.Error)
	}
	return nil
}

func (r *taskCollabRepo) ListComments(ctx context.Context, taskID string) ([]*domain.TaskComment, error) {
	var comments []*domain.TaskComment
	if err := r.db.WithContext(ctx).Raw(
		`SELECT * FROM task_comments WHERE task_id = $1 ORDER BY created_at ASC`, taskID,
	).Scan(&comments).Error; err != nil {
		return nil, fmt.Errorf("task comment list: %w", err)
	}
	return comments, nil
}

func (r *taskCollabRepo) DeleteComment(ctx context.Context, id, agentID, companyID string) error {
	res := r.db.WithContext(ctx).Exec(
		`DELETE FROM task_comments WHERE id = $1 AND agent_id = $2 AND company_id = $3`,
		id, agentID, companyID,
	)
	if res.Error != nil {
		return fmt.Errorf("task comment delete: %w", res.Error)
	}
	return nil
}

func (r *taskCollabRepo) AddDependency(ctx context.Context, d *domain.TaskDependency) error {
	res := r.db.WithContext(ctx).Exec(
		`INSERT INTO task_dependencies (id, task_id, depends_on_id, company_id) VALUES ($1, $2, $3, $4)`,
		d.ID, d.TaskID, d.DependsOnID, d.CompanyID,
	)
	if res.Error != nil {
		return fmt.Errorf("task dependency add: %w", res.Error)
	}
	return nil
}

func (r *taskCollabRepo) ListDependencies(ctx context.Context, taskID string) ([]*domain.TaskDependency, error) {
	var deps []*domain.TaskDependency
	if err := r.db.WithContext(ctx).Raw(
		`SELECT * FROM task_dependencies WHERE task_id = $1 ORDER BY created_at ASC`, taskID,
	).Scan(&deps).Error; err != nil {
		return nil, fmt.Errorf("task dependency list: %w", err)
	}
	return deps, nil
}

func (r *taskCollabRepo) DeleteDependency(ctx context.Context, taskID, dependsOnID string) error {
	res := r.db.WithContext(ctx).Exec(
		`DELETE FROM task_dependencies WHERE task_id = $1 AND depends_on_id = $2`, taskID, dependsOnID,
	)
	if res.Error != nil {
		return fmt.Errorf("task dependency delete: %w", res.Error)
	}
	return nil
}

func (r *taskCollabRepo) AddWatcher(ctx context.Context, w *domain.TaskWatcher) error {
	res := r.db.WithContext(ctx).Exec(
		`INSERT INTO task_watchers (task_id, agent_id) VALUES ($1, $2)
		 ON CONFLICT (task_id, agent_id) DO NOTHING`,
		w.TaskID, w.AgentID,
	)
	if res.Error != nil {
		return fmt.Errorf("task watcher add: %w", res.Error)
	}
	return nil
}

func (r *taskCollabRepo) ListWatchers(ctx context.Context, taskID string) ([]*domain.TaskWatcher, error) {
	var watchers []*domain.TaskWatcher
	if err := r.db.WithContext(ctx).Raw(
		`SELECT * FROM task_watchers WHERE task_id = $1 ORDER BY created_at ASC`, taskID,
	).Scan(&watchers).Error; err != nil {
		return nil, fmt.Errorf("task watcher list: %w", err)
	}
	return watchers, nil
}

func (r *taskCollabRepo) RemoveWatcher(ctx context.Context, taskID, agentID string) error {
	res := r.db.WithContext(ctx).Exec(
		`DELETE FROM task_watchers WHERE task_id = $1 AND agent_id = $2`, taskID, agentID,
	)
	if res.Error != nil {
		return fmt.Errorf("task watcher remove: %w", res.Error)
	}
	return nil
}
