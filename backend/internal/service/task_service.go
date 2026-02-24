package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/event"
	"github.com/linkclaw/backend/internal/repository"
)

type TaskService struct {
	taskRepo    repository.TaskRepo
	messageRepo repository.MessageRepo
	companyRepo repository.CompanyRepo
}

func NewTaskService(taskRepo repository.TaskRepo, messageRepo repository.MessageRepo, companyRepo repository.CompanyRepo) *TaskService {
	return &TaskService{taskRepo: taskRepo, messageRepo: messageRepo, companyRepo: companyRepo}
}

type CreateTaskInput struct {
	CompanyID   string
	ParentID    *string
	Title       string
	Description string
	Priority    domain.TaskPriority
	AssigneeID  *string
	CreatedBy   *string
}

func (s *TaskService) Create(ctx context.Context, in CreateTaskInput) (*domain.Task, error) {
	if in.Priority == "" {
		in.Priority = domain.TaskPriorityMedium
	}
	status := domain.TaskStatusPending
	if in.AssigneeID != nil && *in.AssigneeID != "" {
		status = domain.TaskStatusAssigned
	}
	t := &domain.Task{
		ID:          uuid.New().String(),
		CompanyID:   in.CompanyID,
		ParentID:    in.ParentID,
		Title:       in.Title,
		Description: in.Description,
		Priority:    in.Priority,
		Status:      status,
		AssigneeID:  in.AssigneeID,
		CreatedBy:   in.CreatedBy,
	}
	if err := s.taskRepo.Create(ctx, t); err != nil {
		return nil, err
	}
	event.Global.Publish(event.NewEvent(event.TaskCreated, event.TaskCreatedPayload{
		TaskID: t.ID, CompanyID: t.CompanyID, Title: t.Title, AssigneeID: t.AssigneeID,
	}))
	return t, nil
}

func (s *TaskService) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	return s.taskRepo.GetByID(ctx, id)
}

func (s *TaskService) List(ctx context.Context, q repository.TaskQuery) ([]*domain.Task, int, error) {
	return s.taskRepo.List(ctx, q)
}

// Accept 将任务从 assigned 变为 in_progress，并广播 task_update 消息
func (s *TaskService) Accept(ctx context.Context, taskID, agentID string) (*domain.Task, error) {
	t, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil || t == nil {
		return nil, fmt.Errorf("task not found")
	}
	if t.AssigneeID == nil || *t.AssigneeID != agentID {
		return nil, fmt.Errorf("task not assigned to you")
	}
	if !t.Status.CanTransitionTo(domain.TaskStatusInProgress) {
		return nil, fmt.Errorf("cannot accept task in status %s", t.Status)
	}
	if err = s.taskRepo.UpdateStatus(ctx, taskID, domain.TaskStatusInProgress, nil, nil); err != nil {
		return nil, err
	}
	t.Status = domain.TaskStatusInProgress
	s.broadcastTaskUpdate(ctx, t)
	event.Global.Publish(event.NewEvent(event.TaskUpdated, event.TaskUpdatedPayload{
		TaskID: t.ID, CompanyID: t.CompanyID, Status: string(t.Status), Title: t.Title, AssigneeID: t.AssigneeID,
	}))
	return t, nil
}

// Submit 将任务标记为完成
func (s *TaskService) Submit(ctx context.Context, taskID, agentID, result string) (*domain.Task, error) {
	t, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil || t == nil {
		return nil, fmt.Errorf("task not found")
	}
	if t.AssigneeID == nil || *t.AssigneeID != agentID {
		return nil, fmt.Errorf("task not assigned to you")
	}
	if !t.Status.CanTransitionTo(domain.TaskStatusDone) {
		return nil, fmt.Errorf("cannot submit task in status %s", t.Status)
	}
	if err = s.taskRepo.UpdateStatus(ctx, taskID, domain.TaskStatusDone, &result, nil); err != nil {
		return nil, err
	}
	t.Status = domain.TaskStatusDone
	t.Result = &result
	s.broadcastTaskUpdate(ctx, t)
	event.Global.Publish(event.NewEvent(event.TaskUpdated, event.TaskUpdatedPayload{
		TaskID: t.ID, CompanyID: t.CompanyID, Status: string(t.Status), Title: t.Title, AssigneeID: t.AssigneeID,
	}))
	return t, nil
}

// Fail 将任务标记为失败
func (s *TaskService) Fail(ctx context.Context, taskID, agentID, reason string) (*domain.Task, error) {
	t, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil || t == nil {
		return nil, fmt.Errorf("task not found")
	}
	if t.AssigneeID == nil || *t.AssigneeID != agentID {
		return nil, fmt.Errorf("task not assigned to you")
	}
	if !t.Status.CanTransitionTo(domain.TaskStatusFailed) {
		return nil, fmt.Errorf("cannot fail task in status %s", t.Status)
	}
	if err = s.taskRepo.UpdateStatus(ctx, taskID, domain.TaskStatusFailed, nil, &reason); err != nil {
		return nil, err
	}
	t.Status = domain.TaskStatusFailed
	t.FailReason = &reason
	s.broadcastTaskUpdate(ctx, t)
	event.Global.Publish(event.NewEvent(event.TaskUpdated, event.TaskUpdatedPayload{
		TaskID: t.ID, CompanyID: t.CompanyID, Status: string(t.Status), Title: t.Title, AssigneeID: t.AssigneeID,
	}))
	return t, nil
}

func (s *TaskService) Delete(ctx context.Context, id string) error {
	return s.taskRepo.Delete(ctx, id)
}

// broadcastTaskUpdate 在 #general 频道发布 task_update 消息
func (s *TaskService) broadcastTaskUpdate(ctx context.Context, t *domain.Task) {
	ch, err := s.companyRepo.GetChannelByName(ctx, t.CompanyID, "general")
	if err != nil || ch == nil {
		return
	}
	meta := domain.TaskMeta{
		TaskID:     t.ID,
		Title:      t.Title,
		Status:     t.Status,
		Priority:   t.Priority,
		AssigneeID: t.AssigneeID,
		DueAt:      t.DueAt,
		Result:     t.Result,
	}
	metaJSON, _ := json.Marshal(meta)
	content := fmt.Sprintf("任务「%s」状态更新为 **%s**", t.Title, t.Status)
	chID := ch.ID
	msg := &domain.Message{
		ID:        uuid.New().String(),
		CompanyID: t.CompanyID,
		ChannelID: &chID,
		Content:   content,
		MsgType:   domain.MsgTypeTaskUpdate,
		TaskMeta:  metaJSON,
	}
	s.messageRepo.Create(ctx, msg) //nolint:errcheck
}
