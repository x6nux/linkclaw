package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
	"github.com/linkclaw/backend/internal/service"
)

type observabilityHandler struct {
	obsSvc     *service.ObservabilityService
	obsRepo    repository.ObservabilityRepo
	qualitySvc *service.QualityScoringService
}

func parseIntQuery(c *gin.Context, key string, defaultVal int) int {
	raw := c.Query(key)
	if raw == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return defaultVal
	}
	return v
}

func (h *observabilityHandler) overview(c *gin.Context) {
	overview, err := h.obsRepo.GetTraceOverview(c.Request.Context(), currentCompanyID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": overview})
}

func (h *observabilityHandler) listTraces(c *gin.Context) {
	q := repository.TraceRunQuery{
		CompanyID:  currentCompanyID(c),
		Status:     domain.TraceStatus(c.Query("status")),
		SourceType: domain.TraceSourceType(c.Query("source_type")),
		Limit:      parseIntQuery(c, "limit", 50),
		Offset:     parseIntQuery(c, "offset", 0),
	}
	traces, total, err := h.obsSvc.ListTraces(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": traces, "total": total})
}

func (h *observabilityHandler) getTrace(c *gin.Context) {
	tree, err := h.obsSvc.GetTraceTree(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tree == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tree})
}

func (h *observabilityHandler) scoreTrace(c *gin.Context) {
	score, err := h.qualitySvc.ScoreConversation(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": score})
}

func (h *observabilityHandler) listBudgetPolicies(c *gin.Context) {
	policies, err := h.obsRepo.ListBudgetPolicies(c.Request.Context(), currentCompanyID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": policies, "total": len(policies)})
}

type createBudgetPolicyRequest struct {
	ScopeType          string  `json:"scope_type"          binding:"required"`
	ScopeID            *string `json:"scope_id"`
	Period             string  `json:"period"             binding:"required"`
	BudgetMicrodollars int64   `json:"budget_microdollars"`
	WarnRatio          float64 `json:"warn_ratio"`
	CriticalRatio      float64 `json:"critical_ratio"`
	HardLimitEnabled   bool    `json:"hard_limit_enabled"`
	IsActive           *bool   `json:"is_active"`
}

func (h *observabilityHandler) createBudgetPolicy(c *gin.Context) {
	var req createBudgetPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	p := &domain.LLMBudgetPolicy{
		ID:                 uuid.New().String(),
		CompanyID:          currentCompanyID(c),
		ScopeType:          domain.BudgetScopeType(req.ScopeType),
		ScopeID:            req.ScopeID,
		Period:             domain.BudgetPeriod(req.Period),
		BudgetMicrodollars: req.BudgetMicrodollars,
		WarnRatio:          req.WarnRatio,
		CriticalRatio:      req.CriticalRatio,
		HardLimitEnabled:   req.HardLimitEnabled,
		IsActive:           isActive,
	}
	if err := h.obsRepo.CreateBudgetPolicy(c.Request.Context(), p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": p})
}

type updateBudgetPolicyRequest struct {
	BudgetMicrodollars int64   `json:"budget_microdollars"`
	WarnRatio          float64 `json:"warn_ratio"`
	CriticalRatio      float64 `json:"critical_ratio"`
	HardLimitEnabled   bool    `json:"hard_limit_enabled"`
	IsActive           bool    `json:"is_active"`
}

func (h *observabilityHandler) updateBudgetPolicy(c *gin.Context) {
	var req updateBudgetPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p := &domain.LLMBudgetPolicy{
		ID:                 c.Param("id"),
		BudgetMicrodollars: req.BudgetMicrodollars,
		WarnRatio:          req.WarnRatio,
		CriticalRatio:      req.CriticalRatio,
		HardLimitEnabled:   req.HardLimitEnabled,
		IsActive:           req.IsActive,
	}
	if err := h.obsRepo.UpdateBudgetPolicy(c.Request.Context(), p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": p})
}

func (h *observabilityHandler) listBudgetAlerts(c *gin.Context) {
	q := repository.BudgetAlertQuery{
		CompanyID: currentCompanyID(c),
		Status:    domain.BudgetAlertStatus(c.Query("status")),
		Level:     domain.BudgetAlertLevel(c.Query("level")),
		Limit:     parseIntQuery(c, "limit", 50),
		Offset:    parseIntQuery(c, "offset", 0),
	}
	alerts, err := h.obsRepo.ListBudgetAlerts(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": alerts, "total": len(alerts)})
}

type patchBudgetAlertRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h *observabilityHandler) patchBudgetAlert(c *gin.Context) {
	var req patchBudgetAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.obsRepo.UpdateBudgetAlert(
		c.Request.Context(),
		c.Param("id"),
		domain.BudgetAlertStatus(req.Status),
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *observabilityHandler) listErrorPolicies(c *gin.Context) {
	policies, err := h.obsRepo.ListErrorAlertPolicies(c.Request.Context(), currentCompanyID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": policies, "total": len(policies)})
}

type createErrorPolicyRequest struct {
	ScopeType          string  `json:"scope_type" binding:"required"`
	ScopeID            *string `json:"scope_id"`
	WindowMinutes      int     `json:"window_minutes"`
	MinRequests        int     `json:"min_requests"`
	ErrorRateThreshold float64 `json:"error_rate_threshold"`
	CooldownMinutes    int     `json:"cooldown_minutes"`
}

func (h *observabilityHandler) createErrorPolicy(c *gin.Context) {
	var req createErrorPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p := &domain.LLMErrorAlertPolicy{
		ID:                 uuid.New().String(),
		CompanyID:          currentCompanyID(c),
		ScopeType:          domain.ErrorAlertScopeType(req.ScopeType),
		ScopeID:            req.ScopeID,
		WindowMinutes:      req.WindowMinutes,
		MinRequests:        req.MinRequests,
		ErrorRateThreshold: req.ErrorRateThreshold,
		CooldownMinutes:    req.CooldownMinutes,
	}
	if err := h.obsRepo.CreateErrorAlertPolicy(c.Request.Context(), p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": p})
}

func (h *observabilityHandler) listQualityScores(c *gin.Context) {
	q := repository.QualityScoreQuery{
		CompanyID: currentCompanyID(c),
		Limit:     parseIntQuery(c, "limit", 50),
		Offset:    parseIntQuery(c, "offset", 0),
	}
	scores, err := h.qualitySvc.ListScores(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": scores, "total": len(scores)})
}
