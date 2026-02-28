package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/event"
	"github.com/linkclaw/backend/internal/repository"
)

const (
	taskUploadDirBase  = "/uploads/tasks"
	taskUploadPathBase = "/uploads/tasks"
)

type TaskService struct {
	taskRepo    repository.TaskRepo
	collabRepo  repository.TaskCollabRepo
	messageRepo repository.MessageRepo
	companyRepo repository.CompanyRepo
}

func NewTaskService(taskRepo repository.TaskRepo, collabRepo repository.TaskCollabRepo, messageRepo repository.MessageRepo, companyRepo repository.CompanyRepo) *TaskService {
	return &TaskService{taskRepo: taskRepo, collabRepo: collabRepo, messageRepo: messageRepo, companyRepo: companyRepo}
}

type CreateTaskInput struct {
	CompanyID   string
	ParentID    *string
	Title       string
	Description string
	Priority    domain.TaskPriority
	AssigneeID  *string
	CreatedBy   *string
	Tags        domain.StringList
	Attachments []TaskUploadFile
}

type TaskUploadFile struct {
	OriginalFilename string
	Size             int64
	MimeType         string
	Content          []byte
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
		Tags:        in.Tags,
	}
	if err := s.taskRepo.Create(ctx, t); err != nil {
		return nil, err
	}
	if len(in.Attachments) > 0 {
		attachments, err := s.storeAttachments(ctx, t, in.Attachments)
		if err != nil {
			return nil, err
		}
		t.Attachments = attachments
	}
	event.Global.Publish(event.NewEvent(event.TaskCreated, event.TaskCreatedPayload{
		TaskID: t.ID, CompanyID: t.CompanyID, Title: t.Title, AssigneeID: t.AssigneeID,
	}))
	return t, nil
}

func (s *TaskService) storeAttachments(ctx context.Context, task *domain.Task, files []TaskUploadFile) ([]*domain.TaskAttachment, error) {
	dirPath := filepath.Join(taskUploadDirBase, task.ID)
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		return nil, fmt.Errorf("create task upload dir: %w", err)
	}

	now := time.Now().UTC()
	attachments := make([]*domain.TaskAttachment, 0, len(files))
	writtenPaths := make([]string, 0, len(files))

	for _, file := range files {
		ext := taskAttachmentExtension(file.OriginalFilename)
		filename := uuid.NewString() + ext

		absPath := filepath.Join(dirPath, filename)
		if err := os.WriteFile(absPath, file.Content, 0o644); err != nil {
			cleanupTaskFiles(writtenPaths)
			return nil, fmt.Errorf("save task attachment: %w", err)
		}
		writtenPaths = append(writtenPaths, absPath)

		size := file.Size
		if size <= 0 {
			size = int64(len(file.Content))
		}
		mimeType := strings.TrimSpace(file.MimeType)
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		attachments = append(attachments, &domain.TaskAttachment{
			ID:               uuid.NewString(),
			TaskID:           task.ID,
			CompanyID:        task.CompanyID,
			Filename:         filename,
			OriginalFilename: file.OriginalFilename,
			FileSize:         size,
			MimeType:         mimeType,
			StoragePath:      path.Join(taskUploadPathBase, task.ID, filename),
			UploadedBy:       task.CreatedBy,
			CreatedAt:        now,
		})
	}

	if err := s.taskRepo.CreateAttachments(ctx, attachments); err != nil {
		cleanupTaskFiles(writtenPaths)
		return nil, err
	}
	return attachments, nil
}

func cleanupTaskFiles(paths []string) {
	for _, p := range paths {
		_ = os.Remove(p)
	}
}

func taskAttachmentExtension(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	if strings.HasSuffix(lower, ".tar.gz") {
		return ".tar.gz"
	}
	return strings.ToLower(filepath.Ext(lower))
}

func (s *TaskService) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	return s.taskRepo.GetByID(ctx, id)
}

func (s *TaskService) List(ctx context.Context, q repository.TaskQuery) ([]*domain.Task, int, error) {
	return s.taskRepo.List(ctx, q)
}

func (s *TaskService) UpdateTags(ctx context.Context, taskID string, tags domain.StringList) error {
	if tags == nil {
		tags = domain.StringList{}
	}
	return s.taskRepo.UpdateTags(ctx, taskID, tags)
}

func (s *TaskService) AddComment(ctx context.Context, taskID, agentID, content string) (*domain.TaskComment, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("comment content is required")
	}
	t, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, fmt.Errorf("task not found")
	}
	c := &domain.TaskComment{
		ID:        uuid.New().String(),
		TaskID:    taskID,
		CompanyID: t.CompanyID,
		AgentID:   agentID,
		Content:   content,
	}
	if err := s.collabRepo.AddComment(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *TaskService) DeleteComment(ctx context.Context, commentID, agentID, companyID string) error {
	return s.collabRepo.DeleteComment(ctx, commentID, agentID, companyID)
}

func (s *TaskService) AddDependency(ctx context.Context, taskID, dependsOnID string) (*domain.TaskDependency, error) {
	if taskID == dependsOnID {
		return nil, fmt.Errorf("task cannot depend on itself")
	}
	t, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, fmt.Errorf("task not found")
	}
	target, err := s.taskRepo.GetByID(ctx, dependsOnID)
	if err != nil {
		return nil, err
	}
	if target == nil {
		return nil, fmt.Errorf("dependency task not found")
	}
	if t.CompanyID != target.CompanyID {
		return nil, fmt.Errorf("cross-company dependency not allowed")
	}
	d := &domain.TaskDependency{
		ID:          uuid.New().String(),
		TaskID:      taskID,
		DependsOnID: dependsOnID,
		CompanyID:   t.CompanyID,
	}
	if err := s.collabRepo.AddDependency(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *TaskService) RemoveDependency(ctx context.Context, taskID, dependsOnID string) error {
	return s.collabRepo.DeleteDependency(ctx, taskID, dependsOnID)
}

func (s *TaskService) AddWatcher(ctx context.Context, taskID, agentID string) error {
	t, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if t == nil {
		return fmt.Errorf("task not found")
	}
	w := &domain.TaskWatcher{TaskID: taskID, AgentID: agentID}
	return s.collabRepo.AddWatcher(ctx, w)
}

func (s *TaskService) RemoveWatcher(ctx context.Context, taskID, agentID string) error {
	return s.collabRepo.RemoveWatcher(ctx, taskID, agentID)
}

func (s *TaskService) GetTaskDetail(ctx context.Context, id string) (*domain.Task, error) {
	t, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, fmt.Errorf("task not found")
	}
	comments, err := s.collabRepo.ListComments(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	deps, err := s.collabRepo.ListDependencies(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	watchers, err := s.collabRepo.ListWatchers(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	t.Comments = comments
	t.Dependencies = deps
	t.Watchers = watchers
	return t, nil
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

func (s *TaskService) Delete(ctx context.Context, id, companyID string) error {
	return s.taskRepo.Delete(ctx, id, companyID)
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
