package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
	"github.com/linkclaw/backend/internal/service"
)

type mockDeptRepo struct {
	createFn  func(ctx context.Context, d *domain.Department) error
	getByIDFn func(ctx context.Context, id string) (*domain.Department, error)
	listFn    func(ctx context.Context, companyID string) ([]*domain.Department, error)
}

func (m *mockDeptRepo) Create(ctx context.Context, d *domain.Department) error {
	if m.createFn != nil {
		return m.createFn(ctx, d)
	}
	return nil
}
func (m *mockDeptRepo) GetByID(ctx context.Context, id string) (*domain.Department, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockDeptRepo) GetBySlug(context.Context, string, string) (*domain.Department, error) { return nil, nil }
func (m *mockDeptRepo) List(ctx context.Context, companyID string) ([]*domain.Department, error) {
	if m.listFn != nil {
		return m.listFn(ctx, companyID)
	}
	return nil, nil
}
func (m *mockDeptRepo) Update(context.Context, *domain.Department) error { return nil }
func (m *mockDeptRepo) Delete(context.Context, string) error              { return nil }
func (m *mockDeptRepo) AssignAgent(context.Context, string, string) error { return nil }

type mockAgentRepo struct {
	getByIDFn      func(ctx context.Context, id string) (*domain.Agent, error)
	getByCompanyFn func(ctx context.Context, companyID string) ([]*domain.Agent, error)
}

func (m *mockAgentRepo) Create(context.Context, *domain.Agent) error { return nil }
func (m *mockAgentRepo) GetByID(ctx context.Context, id string) (*domain.Agent, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockAgentRepo) GetByAPIKeyHash(context.Context, string) (*domain.Agent, error) { return nil, nil }
func (m *mockAgentRepo) GetByCompany(ctx context.Context, companyID string) ([]*domain.Agent, error) {
	if m.getByCompanyFn != nil {
		return m.getByCompanyFn(ctx, companyID)
	}
	return nil, nil
}
func (m *mockAgentRepo) GetByName(context.Context, string, string) (*domain.Agent, error) { return nil, nil }
func (m *mockAgentRepo) GetByHireRequestID(context.Context, string) (*domain.Agent, error) { return nil, nil }
func (m *mockAgentRepo) UpdateStatus(context.Context, string, domain.AgentStatus) error     { return nil }
func (m *mockAgentRepo) UpdateLastSeen(context.Context, string) error                        { return nil }
func (m *mockAgentRepo) UpdateName(context.Context, string, string) error                    { return nil }
func (m *mockAgentRepo) UpdateModel(context.Context, string, string) error                   { return nil }
func (m *mockAgentRepo) MarkInitialized(context.Context, string) error                        { return nil }
func (m *mockAgentRepo) SetPasswordHash(context.Context, string, string) error               { return nil }
func (m *mockAgentRepo) UpdatePersona(context.Context, string, string) error                 { return nil }
func (m *mockAgentRepo) UpdateAPIKey(context.Context, string, string, string) error          { return nil }
func (m *mockAgentRepo) UpdateDepartment(context.Context, string, *string) error             { return nil }
func (m *mockAgentRepo) UpdateManager(context.Context, string, *string) error                { return nil }
func (m *mockAgentRepo) ListByDepartment(context.Context, string, string) ([]*domain.Agent, error) {
	return nil, nil
}
func (m *mockAgentRepo) Delete(context.Context, string) error { return nil }

type mockApprovalRepo struct {
	createFn       func(ctx context.Context, r *domain.ApprovalRequest) error
	getByIDFn      func(ctx context.Context, id string) (*domain.ApprovalRequest, error)
	listFn         func(ctx context.Context, q repository.ApprovalQuery) ([]*domain.ApprovalRequest, int, error)
	updateStatusFn func(ctx context.Context, id string, status domain.ApprovalStatus, decisionReason string, decidedAt *time.Time) error
}

func (m *mockApprovalRepo) Create(ctx context.Context, r *domain.ApprovalRequest) error {
	if m.createFn != nil {
		return m.createFn(ctx, r)
	}
	return nil
}
func (m *mockApprovalRepo) GetByID(ctx context.Context, id string) (*domain.ApprovalRequest, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockApprovalRepo) List(ctx context.Context, q repository.ApprovalQuery) ([]*domain.ApprovalRequest, int, error) {
	if m.listFn != nil {
		return m.listFn(ctx, q)
	}
	return nil, 0, nil
}
func (m *mockApprovalRepo) UpdateStatus(ctx context.Context, id string, status domain.ApprovalStatus, decisionReason string, decidedAt *time.Time) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status, decisionReason, decidedAt)
	}
	return nil
}

func newOrgHandler(dept *mockDeptRepo, agent *mockAgentRepo, approval *mockApprovalRepo) *organizationHandler {
	return &organizationHandler{orgSvc: service.NewOrganizationService(dept, agent, approval)}
}

func TestOrgHandler_ListDepartments(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		mockFn     func(*mockDeptRepo)
		wantStatus int
	}{
		{"success", func(d *mockDeptRepo) {
			d.listFn = func(context.Context, string) ([]*domain.Department, error) { return []*domain.Department{{ID: "d1"}}, nil }
		}, http.StatusOK},
		{"error", func(d *mockDeptRepo) {
			d.listFn = func(context.Context, string) ([]*domain.Department, error) { return nil, errors.New("db error") }
		}, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dept := &mockDeptRepo{}
			tt.mockFn(dept)
			r := gin.New()
			h := newOrgHandler(dept, &mockAgentRepo{}, &mockApprovalRepo{})
			r.GET("/organization/departments", injectAgent(&domain.Agent{ID: "chair-1", CompanyID: "c1", RoleType: domain.RoleChairman}), h.listDepartments)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/organization/departments", nil))
			if w.Code != tt.wantStatus {
				t.Errorf("want %d got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestOrgHandler_CreateDepartment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		body       string
		mockFn     func(*mockDeptRepo)
		wantStatus int
	}{
		{"success", `{"name":"Engineering","slug":"eng","description":"x"}`, func(d *mockDeptRepo) { d.createFn = func(context.Context, *domain.Department) error { return nil } }, http.StatusCreated},
		{"bad json", `{"name":`, nil, http.StatusBadRequest},
		{"service error", `{"name":"Engineering","slug":"eng"}`, func(d *mockDeptRepo) { d.createFn = func(context.Context, *domain.Department) error { return errors.New("create failed") } }, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dept := &mockDeptRepo{}
			if tt.mockFn != nil {
				tt.mockFn(dept)
			}
			r := gin.New()
			h := newOrgHandler(dept, &mockAgentRepo{}, &mockApprovalRepo{})
			r.POST("/organization/departments", injectAgent(&domain.Agent{ID: "chair-1", CompanyID: "c1", RoleType: domain.RoleChairman}), h.createDepartment)
			req := httptest.NewRequest(http.MethodPost, "/organization/departments", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("want %d got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestOrgHandler_ListApprovals_AsChairman(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var got repository.ApprovalQuery
	approval := &mockApprovalRepo{
		listFn: func(_ context.Context, q repository.ApprovalQuery) ([]*domain.ApprovalRequest, int, error) {
			got = q
			return []*domain.ApprovalRequest{{ID: "ap1"}}, 1, nil
		},
	}
	r := gin.New()
	h := newOrgHandler(&mockDeptRepo{}, &mockAgentRepo{}, approval)
	r.GET("/organization/approvals", injectAgent(&domain.Agent{ID: "chair-1", CompanyID: "c1", RoleType: domain.RoleChairman}), h.listApprovals)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/organization/approvals", nil))
	if w.Code != http.StatusOK {
		t.Errorf("want %d got %d", http.StatusOK, w.Code)
	}
	if got.RequesterID != "" {
		t.Errorf("chairman should not be filtered, got requester_id=%q", got.RequesterID)
	}
}

func TestOrgHandler_ListApprovals_AsEmployee(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var got repository.ApprovalQuery
	approval := &mockApprovalRepo{
		listFn: func(_ context.Context, q repository.ApprovalQuery) ([]*domain.ApprovalRequest, int, error) {
			got = q
			return []*domain.ApprovalRequest{{ID: "ap1"}}, 1, nil
		},
	}
	employee := &domain.Agent{ID: "agent-1", CompanyID: "c1", RoleType: domain.RoleEmployee}
	r := gin.New()
	h := newOrgHandler(&mockDeptRepo{}, &mockAgentRepo{}, approval)
	r.GET("/organization/approvals", injectAgent(employee), h.listApprovals)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/organization/approvals", nil))
	if w.Code != http.StatusOK {
		t.Errorf("want %d got %d", http.StatusOK, w.Code)
	}
	if got.RequesterID != employee.ID {
		t.Errorf("want requester_id=%q got %q", employee.ID, got.RequesterID)
	}
}

func TestOrgHandler_CreateApproval(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		body       string
		setup      func(*mockAgentRepo, *mockApprovalRepo, *bool)
		wantStatus int
		wantCreate bool
	}{
		{"success", `{"request_type":"custom","reason":"need this","payload":{"x":1}}`, func(a *mockAgentRepo, ap *mockApprovalRepo, created *bool) {
			a.getByIDFn = func(_ context.Context, id string) (*domain.Agent, error) {
				return &domain.Agent{ID: id, CompanyID: "c1", Position: domain.PositionBackendDev}, nil
			}
			a.getByCompanyFn = func(context.Context, string) ([]*domain.Agent, error) { return []*domain.Agent{}, nil }
			ap.createFn = func(context.Context, *domain.ApprovalRequest) error { *created = true; return nil }
		}, http.StatusCreated, true},
		{"bad json", `{"request_type":`, nil, http.StatusBadRequest, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &mockAgentRepo{}
			approval := &mockApprovalRepo{}
			created := false
			if tt.setup != nil {
				tt.setup(agent, approval, &created)
			}
			r := gin.New()
			h := newOrgHandler(&mockDeptRepo{}, agent, approval)
			r.POST("/organization/approvals", injectAgent(&domain.Agent{ID: "agent-1", CompanyID: "c1", RoleType: domain.RoleEmployee}), h.createApproval)
			req := httptest.NewRequest(http.MethodPost, "/organization/approvals", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("want %d got %d", tt.wantStatus, w.Code)
			}
			if tt.wantCreate && !created {
				t.Errorf("expected approvalRepo.Create to be called")
			}
		})
	}
}

func TestOrgHandler_ChairmanRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := newOrgHandler(&mockDeptRepo{}, &mockAgentRepo{}, &mockApprovalRepo{})
	r.GET("/organization/departments", injectAgent(&domain.Agent{ID: "agent-1", CompanyID: "c1", RoleType: domain.RoleEmployee}), ChairmanOnly(), h.listDepartments)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/organization/departments", nil))
	if w.Code != http.StatusForbidden {
		t.Errorf("want %d got %d", http.StatusForbidden, w.Code)
	}
}
