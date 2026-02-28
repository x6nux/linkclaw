package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
	"github.com/linkclaw/backend/internal/service"
)

type mockObsRepo struct {
	getTraceOverviewFn        func(ctx context.Context, companyID string) (*repository.TraceOverview, error)
	listTraceRunsFn           func(ctx context.Context, q repository.TraceRunQuery) ([]*domain.TraceRun, int, error)
	getTraceRunByIDFn         func(ctx context.Context, id string) (*domain.TraceRun, error)
	listTraceSpansByTraceIDFn func(ctx context.Context, traceID string) ([]*domain.TraceSpan, error)
	listBudgetPoliciesFn      func(ctx context.Context, companyID string) ([]*domain.LLMBudgetPolicy, error)
	createBudgetPolicyFn      func(ctx context.Context, p *domain.LLMBudgetPolicy) error
	listBudgetAlertsFn        func(ctx context.Context, q repository.BudgetAlertQuery) ([]*domain.LLMBudgetAlert, error)
	listErrorPoliciesFn       func(ctx context.Context, companyID string) ([]*domain.LLMErrorAlertPolicy, error)
	listQualityScoresFn       func(ctx context.Context, q repository.QualityScoreQuery) ([]*domain.ConversationQualityScore, error)
}

func (m *mockObsRepo) CreateTraceRun(context.Context, *domain.TraceRun) error { return nil }
func (m *mockObsRepo) GetTraceRunByID(ctx context.Context, id string) (*domain.TraceRun, error) {
	if m.getTraceRunByIDFn != nil {
		return m.getTraceRunByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockObsRepo) ListTraceRuns(ctx context.Context, q repository.TraceRunQuery) ([]*domain.TraceRun, int, error) {
	if m.listTraceRunsFn != nil {
		return m.listTraceRunsFn(ctx, q)
	}
	return nil, 0, nil
}
func (m *mockObsRepo) UpdateTraceRunStatus(context.Context, string, domain.TraceStatus, *time.Time, *int, *string) error {
	return nil
}
func (m *mockObsRepo) UpdateTraceRunTotals(context.Context, string, int64, int, int) error { return nil }
func (m *mockObsRepo) CreateTraceSpan(context.Context, *domain.TraceSpan) error             { return nil }
func (m *mockObsRepo) GetTraceSpanByID(context.Context, string) (*domain.TraceSpan, error)  { return nil, nil }
func (m *mockObsRepo) ListTraceSpansByTraceID(ctx context.Context, traceID string) ([]*domain.TraceSpan, error) {
	if m.listTraceSpansByTraceIDFn != nil {
		return m.listTraceSpansByTraceIDFn(ctx, traceID)
	}
	return nil, nil
}
func (m *mockObsRepo) UpdateTraceSpan(context.Context, string, domain.TraceStatus, *time.Time, *int, *int, *int, *int64, *string) error {
	return nil
}
func (m *mockObsRepo) CreateTraceReplay(context.Context, *domain.TraceReplay) error { return nil }
func (m *mockObsRepo) GetTraceReplayBySpanID(context.Context, string) (*domain.TraceReplay, error) {
	return nil, nil
}
func (m *mockObsRepo) CreateBudgetPolicy(ctx context.Context, p *domain.LLMBudgetPolicy) error {
	if m.createBudgetPolicyFn != nil {
		return m.createBudgetPolicyFn(ctx, p)
	}
	return nil
}
func (m *mockObsRepo) UpdateBudgetPolicy(context.Context, *domain.LLMBudgetPolicy) error { return nil }
func (m *mockObsRepo) GetBudgetPolicyByID(context.Context, string) (*domain.LLMBudgetPolicy, error) {
	return nil, nil
}
func (m *mockObsRepo) ListBudgetPolicies(ctx context.Context, companyID string) ([]*domain.LLMBudgetPolicy, error) {
	if m.listBudgetPoliciesFn != nil {
		return m.listBudgetPoliciesFn(ctx, companyID)
	}
	return nil, nil
}
func (m *mockObsRepo) ListActiveBudgetPolicies(context.Context, string) ([]*domain.LLMBudgetPolicy, error) {
	return nil, nil
}
func (m *mockObsRepo) CreateBudgetAlert(context.Context, *domain.LLMBudgetAlert) error { return nil }
func (m *mockObsRepo) UpdateBudgetAlert(context.Context, string, domain.BudgetAlertStatus) error {
	return nil
}
func (m *mockObsRepo) ListBudgetAlerts(ctx context.Context, q repository.BudgetAlertQuery) ([]*domain.LLMBudgetAlert, error) {
	if m.listBudgetAlertsFn != nil {
		return m.listBudgetAlertsFn(ctx, q)
	}
	return nil, nil
}
func (m *mockObsRepo) CreateErrorAlertPolicy(context.Context, *domain.LLMErrorAlertPolicy) error { return nil }
func (m *mockObsRepo) UpdateErrorAlertPolicy(context.Context, *domain.LLMErrorAlertPolicy) error {
	return nil
}
func (m *mockObsRepo) ListErrorAlertPolicies(ctx context.Context, companyID string) ([]*domain.LLMErrorAlertPolicy, error) {
	if m.listErrorPoliciesFn != nil {
		return m.listErrorPoliciesFn(ctx, companyID)
	}
	return nil, nil
}
func (m *mockObsRepo) CreateQualityScore(context.Context, *domain.ConversationQualityScore) error { return nil }
func (m *mockObsRepo) ListQualityScores(ctx context.Context, q repository.QualityScoreQuery) ([]*domain.ConversationQualityScore, error) {
	if m.listQualityScoresFn != nil {
		return m.listQualityScoresFn(ctx, q)
	}
	return nil, nil
}
func (m *mockObsRepo) GetQualityScoreByTraceID(context.Context, string) (*domain.ConversationQualityScore, error) {
	return nil, nil
}
func (m *mockObsRepo) GetTraceOverview(ctx context.Context, companyID string) (*repository.TraceOverview, error) {
	if m.getTraceOverviewFn != nil {
		return m.getTraceOverviewFn(ctx, companyID)
	}
	return nil, nil
}

func injectAgent(agent *domain.Agent) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ctxAgent, agent)
		c.Set(ctxCompanyID, agent.CompanyID)
		c.Next()
	}
}

func newObsHandler(repo *mockObsRepo) *observabilityHandler {
	return &observabilityHandler{
		obsSvc:     service.NewObservabilityService(repo),
		obsRepo:    repo,
		qualitySvc: service.NewQualityScoringService(repo),
	}
}

func TestObsHandler_Overview(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		mockFn     func(*mockObsRepo)
		wantStatus int
	}{
		{name: "success", mockFn: func(m *mockObsRepo) {
			m.getTraceOverviewFn = func(context.Context, string) (*repository.TraceOverview, error) {
				return &repository.TraceOverview{Total: 5}, nil
			}
		}, wantStatus: http.StatusOK},
		{name: "repo error", mockFn: func(m *mockObsRepo) {
			m.getTraceOverviewFn = func(context.Context, string) (*repository.TraceOverview, error) { return nil, errors.New("db error") }
		}, wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockObsRepo{}
			tt.mockFn(mock)
			r := gin.New()
			r.GET("/observability/overview", injectAgent(&domain.Agent{ID: "a1", CompanyID: "c1", RoleType: domain.RoleChairman}), newObsHandler(mock).overview)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/observability/overview", nil))
			if w.Code != tt.wantStatus {
				t.Errorf("want %d got %d", tt.wantStatus, w.Code)
			}
			if !json.Valid(w.Body.Bytes()) {
				t.Fatalf("invalid json response")
			}
		})
	}
}

func TestObsHandler_ListTraces(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct{ name string; mockFn func(*mockObsRepo); want int }{
		{"success", func(m *mockObsRepo) { m.listTraceRunsFn = func(context.Context, repository.TraceRunQuery) ([]*domain.TraceRun, int, error) { return []*domain.TraceRun{{ID: "t1"}}, 1, nil } }, http.StatusOK},
		{"repo error", func(m *mockObsRepo) { m.listTraceRunsFn = func(context.Context, repository.TraceRunQuery) ([]*domain.TraceRun, int, error) { return nil, 0, errors.New("db error") } }, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockObsRepo{}
			tt.mockFn(mock)
			r := gin.New()
			r.GET("/observability/traces", injectAgent(&domain.Agent{ID: "a1", CompanyID: "c1", RoleType: domain.RoleChairman}), newObsHandler(mock).listTraces)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/observability/traces", nil))
			if w.Code != tt.want {
				t.Errorf("want %d got %d", tt.want, w.Code)
			}
		})
	}
}

func TestObsHandler_GetTrace(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct{ name string; mockFn func(*mockObsRepo); want int }{
		{"found", func(m *mockObsRepo) {
			m.getTraceRunByIDFn = func(context.Context, string) (*domain.TraceRun, error) { return &domain.TraceRun{ID: "t1"}, nil }
			m.listTraceSpansByTraceIDFn = func(context.Context, string) ([]*domain.TraceSpan, error) { return []*domain.TraceSpan{}, nil }
		}, http.StatusOK},
		{"not found", func(m *mockObsRepo) {
			m.getTraceRunByIDFn = func(context.Context, string) (*domain.TraceRun, error) { return nil, nil }
		}, http.StatusNotFound},
		{"repo error", func(m *mockObsRepo) {
			m.getTraceRunByIDFn = func(context.Context, string) (*domain.TraceRun, error) { return nil, errors.New("db error") }
		}, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockObsRepo{}
			tt.mockFn(mock)
			r := gin.New()
			r.GET("/observability/traces/:id", injectAgent(&domain.Agent{ID: "a1", CompanyID: "c1", RoleType: domain.RoleChairman}), newObsHandler(mock).getTrace)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/observability/traces/t1", nil))
			if w.Code != tt.want {
				t.Errorf("want %d got %d", tt.want, w.Code)
			}
		})
	}
}

func TestObsHandler_ListBudgetPolicies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct{ name string; mockFn func(*mockObsRepo); want int }{
		{"success", func(m *mockObsRepo) { m.listBudgetPoliciesFn = func(context.Context, string) ([]*domain.LLMBudgetPolicy, error) { return []*domain.LLMBudgetPolicy{{ID: "p1"}}, nil } }, http.StatusOK},
		{"error", func(m *mockObsRepo) { m.listBudgetPoliciesFn = func(context.Context, string) ([]*domain.LLMBudgetPolicy, error) { return nil, errors.New("db error") } }, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockObsRepo{}
			tt.mockFn(mock)
			r := gin.New()
			r.GET("/observability/budget-policies", injectAgent(&domain.Agent{ID: "a1", CompanyID: "c1", RoleType: domain.RoleChairman}), newObsHandler(mock).listBudgetPolicies)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/observability/budget-policies", nil))
			if w.Code != tt.want {
				t.Errorf("want %d got %d", tt.want, w.Code)
			}
		})
	}
}

func TestObsHandler_ListBudgetAlerts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct{ name string; mockFn func(*mockObsRepo); want int }{
		{"success", func(m *mockObsRepo) { m.listBudgetAlertsFn = func(context.Context, repository.BudgetAlertQuery) ([]*domain.LLMBudgetAlert, error) { return []*domain.LLMBudgetAlert{{ID: "a1"}}, nil } }, http.StatusOK},
		{"error", func(m *mockObsRepo) { m.listBudgetAlertsFn = func(context.Context, repository.BudgetAlertQuery) ([]*domain.LLMBudgetAlert, error) { return nil, errors.New("db error") } }, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockObsRepo{}
			tt.mockFn(mock)
			r := gin.New()
			r.GET("/observability/budget-alerts", injectAgent(&domain.Agent{ID: "a1", CompanyID: "c1", RoleType: domain.RoleChairman}), newObsHandler(mock).listBudgetAlerts)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/observability/budget-alerts", nil))
			if w.Code != tt.want {
				t.Errorf("want %d got %d", tt.want, w.Code)
			}
		})
	}
}

func TestObsHandler_ListErrorPolicies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct{ name string; mockFn func(*mockObsRepo); want int }{
		{"success", func(m *mockObsRepo) { m.listErrorPoliciesFn = func(context.Context, string) ([]*domain.LLMErrorAlertPolicy, error) { return []*domain.LLMErrorAlertPolicy{{ID: "e1"}}, nil } }, http.StatusOK},
		{"error", func(m *mockObsRepo) { m.listErrorPoliciesFn = func(context.Context, string) ([]*domain.LLMErrorAlertPolicy, error) { return nil, errors.New("db error") } }, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockObsRepo{}
			tt.mockFn(mock)
			r := gin.New()
			r.GET("/observability/error-policies", injectAgent(&domain.Agent{ID: "a1", CompanyID: "c1", RoleType: domain.RoleChairman}), newObsHandler(mock).listErrorPolicies)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/observability/error-policies", nil))
			if w.Code != tt.want {
				t.Errorf("want %d got %d", tt.want, w.Code)
			}
		})
	}
}

func TestObsHandler_ListQualityScores(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct{ name string; mockFn func(*mockObsRepo); want int }{
		{"success", func(m *mockObsRepo) { m.listQualityScoresFn = func(context.Context, repository.QualityScoreQuery) ([]*domain.ConversationQualityScore, error) { return []*domain.ConversationQualityScore{{ID: "q1"}}, nil } }, http.StatusOK},
		{"error", func(m *mockObsRepo) { m.listQualityScoresFn = func(context.Context, repository.QualityScoreQuery) ([]*domain.ConversationQualityScore, error) { return nil, errors.New("db error") } }, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockObsRepo{}
			tt.mockFn(mock)
			r := gin.New()
			r.GET("/observability/quality-scores", injectAgent(&domain.Agent{ID: "a1", CompanyID: "c1", RoleType: domain.RoleChairman}), newObsHandler(mock).listQualityScores)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/observability/quality-scores", nil))
			if w.Code != tt.want {
				t.Errorf("want %d got %d", tt.want, w.Code)
			}
		})
	}
}

func TestObsHandler_ChairmanOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/observability/overview", injectAgent(&domain.Agent{ID: "a1", CompanyID: "c1", RoleType: domain.RoleEmployee}), ChairmanOnly(), newObsHandler(&mockObsRepo{}).overview)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/observability/overview", nil))
	if w.Code != http.StatusForbidden {
		t.Errorf("want %d got %d", http.StatusForbidden, w.Code)
	}
}
