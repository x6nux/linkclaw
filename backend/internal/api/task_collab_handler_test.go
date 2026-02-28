package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
	"github.com/linkclaw/backend/internal/service"
)

type mockTaskRepo struct {
	getByIDFn    func(ctx context.Context, id string) (*domain.Task, error)
	updateTagsFn func(ctx context.Context, id string, tags domain.StringList) error
}

func (m *mockTaskRepo) Create(context.Context, *domain.Task) error { return nil }
func (m *mockTaskRepo) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockTaskRepo) List(context.Context, repository.TaskQuery) ([]*domain.Task, int, error) {
	return nil, 0, nil
}
func (m *mockTaskRepo) UpdateStatus(context.Context, string, domain.TaskStatus, *string, *string) error { return nil }
func (m *mockTaskRepo) UpdateAssignee(context.Context, string, string, domain.TaskStatus) error         { return nil }
func (m *mockTaskRepo) UpdateTags(ctx context.Context, id string, tags domain.StringList) error {
	if m.updateTagsFn != nil {
		return m.updateTagsFn(ctx, id, tags)
	}
	return nil
}
func (m *mockTaskRepo) Delete(context.Context, string) error { return nil }

type mockCollabRepo struct {
	addCommentFn       func(ctx context.Context, c *domain.TaskComment) error
	listCommentsFn     func(ctx context.Context, taskID string) ([]*domain.TaskComment, error)
	deleteCommentFn    func(ctx context.Context, id, agentID, companyID string) error
	addDependencyFn    func(ctx context.Context, d *domain.TaskDependency) error
	listDependencyFn   func(ctx context.Context, taskID string) ([]*domain.TaskDependency, error)
	deleteDependencyFn func(ctx context.Context, taskID, dependsOnID string) error
	addWatcherFn       func(ctx context.Context, w *domain.TaskWatcher) error
	listWatchersFn     func(ctx context.Context, taskID string) ([]*domain.TaskWatcher, error)
	removeWatcherFn    func(ctx context.Context, taskID, agentID string) error
}

func (m *mockCollabRepo) AddComment(ctx context.Context, c *domain.TaskComment) error {
	if m.addCommentFn != nil {
		return m.addCommentFn(ctx, c)
	}
	return nil
}
func (m *mockCollabRepo) ListComments(ctx context.Context, taskID string) ([]*domain.TaskComment, error) {
	if m.listCommentsFn != nil {
		return m.listCommentsFn(ctx, taskID)
	}
	return nil, nil
}
func (m *mockCollabRepo) DeleteComment(ctx context.Context, id, agentID, companyID string) error {
	if m.deleteCommentFn != nil {
		return m.deleteCommentFn(ctx, id, agentID, companyID)
	}
	return nil
}
func (m *mockCollabRepo) AddDependency(ctx context.Context, d *domain.TaskDependency) error {
	if m.addDependencyFn != nil {
		return m.addDependencyFn(ctx, d)
	}
	return nil
}
func (m *mockCollabRepo) ListDependencies(ctx context.Context, taskID string) ([]*domain.TaskDependency, error) {
	if m.listDependencyFn != nil {
		return m.listDependencyFn(ctx, taskID)
	}
	return nil, nil
}
func (m *mockCollabRepo) DeleteDependency(ctx context.Context, taskID, dependsOnID string) error {
	if m.deleteDependencyFn != nil {
		return m.deleteDependencyFn(ctx, taskID, dependsOnID)
	}
	return nil
}
func (m *mockCollabRepo) AddWatcher(ctx context.Context, w *domain.TaskWatcher) error {
	if m.addWatcherFn != nil {
		return m.addWatcherFn(ctx, w)
	}
	return nil
}
func (m *mockCollabRepo) ListWatchers(ctx context.Context, taskID string) ([]*domain.TaskWatcher, error) {
	if m.listWatchersFn != nil {
		return m.listWatchersFn(ctx, taskID)
	}
	return nil, nil
}
func (m *mockCollabRepo) RemoveWatcher(ctx context.Context, taskID, agentID string) error {
	if m.removeWatcherFn != nil {
		return m.removeWatcherFn(ctx, taskID, agentID)
	}
	return nil
}

type mockMessageRepo struct{}

func (m *mockMessageRepo) Create(context.Context, *domain.Message) error                           { return nil }
func (m *mockMessageRepo) ListByChannel(context.Context, string, int, string) ([]*domain.Message, error) {
	return nil, nil
}
func (m *mockMessageRepo) ListDM(context.Context, string, string, int, string) ([]*domain.Message, error) {
	return nil, nil
}
func (m *mockMessageRepo) MarkRead(context.Context, string, []string) error { return nil }
func (m *mockMessageRepo) ListUnreadForAgent(context.Context, string, string) ([]*domain.Message, error) {
	return nil, nil
}

type mockCompanyRepo struct{}

func (m *mockCompanyRepo) Create(context.Context, *domain.Company) error { return nil }
func (m *mockCompanyRepo) GetByID(context.Context, string) (*domain.Company, error) {
	return nil, nil
}
func (m *mockCompanyRepo) GetBySlug(context.Context, string) (*domain.Company, error) { return nil, nil }
func (m *mockCompanyRepo) FindFirst(context.Context) (*domain.Company, error)          { return nil, nil }
func (m *mockCompanyRepo) UpdateSystemPrompt(context.Context, string, string) error    { return nil }
func (m *mockCompanyRepo) UpdateSettings(context.Context, string, *domain.CompanySettings) error {
	return nil
}
func (m *mockCompanyRepo) CreateChannel(context.Context, *domain.Channel) error { return nil }
func (m *mockCompanyRepo) GetChannels(context.Context, string) ([]*domain.Channel, error) {
	return nil, nil
}
func (m *mockCompanyRepo) GetChannelByName(context.Context, string, string) (*domain.Channel, error) {
	return nil, nil
}

func newTaskHandler(task *mockTaskRepo, collab *mockCollabRepo) *taskHandler {
	return &taskHandler{taskSvc: service.NewTaskService(task, collab, &mockMessageRepo{}, &mockCompanyRepo{})}
}

func TestTaskHandler_Detail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct{ name string; setup func(*mockTaskRepo, *mockCollabRepo); want int }{
		{"found", func(task *mockTaskRepo, collab *mockCollabRepo) {
			task.getByIDFn = func(context.Context, string) (*domain.Task, error) { return &domain.Task{ID: "task-1", CompanyID: "company-1"}, nil }
			collab.listCommentsFn = func(context.Context, string) ([]*domain.TaskComment, error) { return []*domain.TaskComment{{ID: "c1"}}, nil }
			collab.listDependencyFn = func(context.Context, string) ([]*domain.TaskDependency, error) { return []*domain.TaskDependency{{ID: "d1"}}, nil }
			collab.listWatchersFn = func(context.Context, string) ([]*domain.TaskWatcher, error) { return []*domain.TaskWatcher{{TaskID: "task-1", AgentID: "agent-1"}}, nil }
		}, http.StatusOK},
		{"not found", func(task *mockTaskRepo, _ *mockCollabRepo) {
			task.getByIDFn = func(context.Context, string) (*domain.Task, error) { return nil, nil }
		}, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, collab := &mockTaskRepo{}, &mockCollabRepo{}
			tt.setup(task, collab)
			r := gin.New()
			r.GET("/tasks/:id/detail", injectAgent(&domain.Agent{ID: "agent-1", CompanyID: "company-1", RoleType: domain.RoleEmployee}), newTaskHandler(task, collab).detail)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/tasks/task-1/detail", nil))
			if w.Code != tt.want {
				t.Errorf("want %d got %d", tt.want, w.Code)
			}
		})
	}
}

func TestTaskHandler_AddComment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct{ name, body string; setup func(*mockTaskRepo, *mockCollabRepo); want int }{
		{"success", `{"content":"hello"}`, func(task *mockTaskRepo, collab *mockCollabRepo) {
			task.getByIDFn = func(context.Context, string) (*domain.Task, error) { return &domain.Task{ID: "task-1", CompanyID: "company-1"}, nil }
			collab.addCommentFn = func(context.Context, *domain.TaskComment) error { return nil }
		}, http.StatusCreated},
		{"bad json", `{"content":`, nil, http.StatusBadRequest},
		{"service error", `{"content":"hello"}`, func(task *mockTaskRepo, _ *mockCollabRepo) {
			task.getByIDFn = func(context.Context, string) (*domain.Task, error) { return nil, nil }
		}, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, collab := &mockTaskRepo{}, &mockCollabRepo{}
			if tt.setup != nil {
				tt.setup(task, collab)
			}
			r := gin.New()
			r.POST("/tasks/:id/comments", injectAgent(&domain.Agent{ID: "agent-1", CompanyID: "company-1", RoleType: domain.RoleEmployee}), newTaskHandler(task, collab).addComment)
			req := httptest.NewRequest(http.MethodPost, "/tasks/task-1/comments", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.want {
				t.Errorf("want %d got %d", tt.want, w.Code)
			}
		})
	}
}

func TestTaskHandler_DeleteComment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name string
		fn   func(context.Context, string, string, string) error
		want int
	}{
		{"success", func(_ context.Context, id, agentID, companyID string) error {
			if id != "comment-1" || agentID != "agent-1" || companyID != "company-1" {
				return errors.New("ownership mismatch")
			}
			return nil
		}, http.StatusOK},
		{"error", func(_ context.Context, _, _, _ string) error {
			return errors.New("delete failed")
		}, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collab := &mockCollabRepo{deleteCommentFn: tt.fn}
			r := gin.New()
			r.DELETE("/tasks/:id/comments/:commentId", injectAgent(&domain.Agent{ID: "agent-1", CompanyID: "company-1", RoleType: domain.RoleEmployee}), newTaskHandler(&mockTaskRepo{}, collab).deleteComment)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/tasks/task-1/comments/comment-1", nil))
			if w.Code != tt.want {
				t.Errorf("want %d got %d", tt.want, w.Code)
			}
		})
	}
}

func TestTaskHandler_AddDependency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct{ name, body string; setup func(*mockTaskRepo, *mockCollabRepo); want int }{
		{"success", `{"depends_on_id":"task-2"}`, func(task *mockTaskRepo, collab *mockCollabRepo) {
			task.getByIDFn = func(_ context.Context, id string) (*domain.Task, error) {
				if id == "task-1" || id == "task-2" { return &domain.Task{ID: id, CompanyID: "company-1"}, nil }
				return nil, nil
			}
			collab.addDependencyFn = func(context.Context, *domain.TaskDependency) error { return nil }
		}, http.StatusCreated},
		{"bad json", `{"depends_on_id":`, nil, http.StatusBadRequest},
		{"self dependency error", `{"depends_on_id":"task-1"}`, nil, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, collab := &mockTaskRepo{}, &mockCollabRepo{}
			if tt.setup != nil { tt.setup(task, collab) }
			r := gin.New()
			r.POST("/tasks/:id/dependencies", injectAgent(&domain.Agent{ID: "agent-1", CompanyID: "company-1", RoleType: domain.RoleEmployee}), newTaskHandler(task, collab).addDependency)
			req := httptest.NewRequest(http.MethodPost, "/tasks/task-1/dependencies", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.want { t.Errorf("want %d got %d", tt.want, w.Code) }
		})
	}
}

func TestTaskHandler_RemoveDependency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct{ name string; fn func(context.Context, string, string) error; want int }{
		{"success", func(context.Context, string, string) error { return nil }, http.StatusOK},
		{"error", func(context.Context, string, string) error { return errors.New("remove failed") }, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collab := &mockCollabRepo{deleteDependencyFn: tt.fn}
			r := gin.New()
			r.DELETE("/tasks/:id/dependencies/:depId", injectAgent(&domain.Agent{ID: "agent-1", CompanyID: "company-1", RoleType: domain.RoleEmployee}), newTaskHandler(&mockTaskRepo{}, collab).removeDependency)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/tasks/task-1/dependencies/task-2", nil))
			if w.Code != tt.want { t.Errorf("want %d got %d", tt.want, w.Code) }
		})
	}
}

func TestTaskHandler_AddWatcher(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct{ name string; setup func(*mockTaskRepo, *mockCollabRepo); want int }{
		{"success", func(task *mockTaskRepo, collab *mockCollabRepo) {
			task.getByIDFn = func(context.Context, string) (*domain.Task, error) { return &domain.Task{ID: "task-1", CompanyID: "company-1"}, nil }
			collab.addWatcherFn = func(context.Context, *domain.TaskWatcher) error { return nil }
		}, http.StatusOK},
		{"task not found", func(task *mockTaskRepo, _ *mockCollabRepo) {
			task.getByIDFn = func(context.Context, string) (*domain.Task, error) { return nil, nil }
		}, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, collab := &mockTaskRepo{}, &mockCollabRepo{}
			tt.setup(task, collab)
			r := gin.New()
			r.POST("/tasks/:id/watchers", injectAgent(&domain.Agent{ID: "agent-1", CompanyID: "company-1", RoleType: domain.RoleEmployee}), newTaskHandler(task, collab).addWatcher)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/tasks/task-1/watchers", nil))
			if w.Code != tt.want { t.Errorf("want %d got %d", tt.want, w.Code) }
		})
	}
}

func TestTaskHandler_RemoveWatcher(t *testing.T) {
	gin.SetMode(gin.TestMode)
	collab := &mockCollabRepo{removeWatcherFn: func(context.Context, string, string) error { return nil }}
	r := gin.New()
	r.DELETE("/tasks/:id/watchers", injectAgent(&domain.Agent{ID: "agent-1", CompanyID: "company-1", RoleType: domain.RoleEmployee}), newTaskHandler(&mockTaskRepo{}, collab).removeWatcher)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/tasks/task-1/watchers", nil))
	if w.Code != http.StatusOK {
		t.Errorf("want %d got %d", http.StatusOK, w.Code)
	}
}

func TestTaskHandler_UpdateTags(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct{ name, body string; fn func(context.Context, string, domain.StringList) error; want int }{
		{"success", `{"tags":["a","b"]}`, func(context.Context, string, domain.StringList) error { return nil }, http.StatusOK},
		{"bad json", `{"tags":`, nil, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &mockTaskRepo{updateTagsFn: tt.fn}
			r := gin.New()
			r.PUT("/tasks/:id/tags", injectAgent(&domain.Agent{ID: "agent-1", CompanyID: "company-1", RoleType: domain.RoleEmployee}), newTaskHandler(task, &mockCollabRepo{}).updateTags)
			req := httptest.NewRequest(http.MethodPut, "/tasks/task-1/tags", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.want { t.Errorf("want %d got %d", tt.want, w.Code) }
		})
	}
}
