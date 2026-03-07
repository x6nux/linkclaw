package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/service"
)

type contextHandler struct {
	contextSvc   *service.ContextService
	agent        *service.ContextSearchAgent
	scheduler    *service.ContextScheduler
	rateLimiter  *service.RateLimiterManager
	metrics      *service.ContextMetrics
	costBudget   *service.CostBudget
}

func newContextHandler(
	contextSvc *service.ContextService,
	agent *service.ContextSearchAgent,
	scheduler *service.ContextScheduler,
) *contextHandler {
	return &contextHandler{
		contextSvc:  contextSvc,
		agent:       agent,
		scheduler:   scheduler,
		rateLimiter: service.NewRateLimiterManager(nil),
		metrics:     service.NewContextMetrics(),
		costBudget:  service.NewCostBudget(10*1000000, 0.8, true), // $10 budget, warn at 80%, hard limit
	}
}

// GetMetrics 返回指标收集器（用于外部访问）
func (h *contextHandler) GetMetrics() *service.ContextMetrics {
	return h.metrics
}

func (h *contextHandler) listDirectories(c *gin.Context) {
	companyID := currentCompanyID(c)
	dirs, err := h.contextSvc.ListDirectories(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": dirs, "total": len(dirs)})
}

type createDirectoryRequest struct {
	Name            string `json:"name" binding:"required"`
	Path            string `json:"path" binding:"required"`
	Description     string `json:"description"`
	FilePatterns    string `json:"file_patterns"`
	ExcludePatterns string `json:"exclude_patterns"`
	MaxFileSize     int    `json:"max_file_size"`
}

func (h *contextHandler) createDirectory(c *gin.Context) {
	var req createDirectoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	companyID := currentCompanyID(c)
	d, err := h.contextSvc.CreateDirectory(c.Request.Context(), service.CreateDirectoryInput{
		CompanyID:       companyID,
		Name:            req.Name,
		Path:            req.Path,
		Description:     req.Description,
		FilePatterns:    req.FilePatterns,
		ExcludePatterns: req.ExcludePatterns,
		MaxFileSize:     req.MaxFileSize,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, d)
}

type updateDirectoryRequest struct {
	Name            *string `json:"name"`
	Path            *string `json:"path"`
	Description     *string `json:"description"`
	FilePatterns    *string `json:"file_patterns"`
	ExcludePatterns *string `json:"exclude_patterns"`
	MaxFileSize     *int    `json:"max_file_size"`
}

func (h *contextHandler) updateDirectory(c *gin.Context) {
	id := c.Param("id")
	companyID := currentCompanyID(c)
	var req updateDirectoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	existing, err := h.contextSvc.GetDirectoryByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil || existing.CompanyID != companyID {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	d, err := h.contextSvc.UpdateDirectory(c.Request.Context(), id, service.UpdateDirectoryInput{
		Name:            req.Name,
		Path:            req.Path,
		Description:     req.Description,
		FilePatterns:    req.FilePatterns,
		ExcludePatterns: req.ExcludePatterns,
		MaxFileSize:     req.MaxFileSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if d == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *contextHandler) deleteDirectory(c *gin.Context) {
	id := c.Param("id")
	companyID := currentCompanyID(c)
	existing, err := h.contextSvc.GetDirectoryByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil || existing.CompanyID != companyID {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := h.contextSvc.DeleteDirectory(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type toggleDirectoryRequest struct {
	IsActive bool `json:"is_active"`
}

func (h *contextHandler) toggleDirectory(c *gin.Context) {
	var req toggleDirectoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	d, err := h.contextSvc.ToggleDirectory(c.Request.Context(), c.Param("id"), req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if d == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, d)
}

type searchRequest struct {
	Query        string   `json:"query" binding:"required"`
	DirectoryIDs []string `json:"directory_ids"`
	// --- 新增统一参数 (v1.1) ---
	MaxResults   int     `json:"max_results,omitempty"`
	MinRelevance float64 `json:"min_relevance,omitempty"`
	TimeoutMs    int     `json:"timeout_ms,omitempty"`
	UseIndex     *bool   `json:"use_index,omitempty"`
}

func (h *contextHandler) search(c *gin.Context) {
	start := time.Now()
	var req searchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.metrics.RecordFailure("invalid_request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agent := currentAgent(c)
	companyID := currentCompanyID(c)
	agentID := ""
	if agent != nil {
		agentID = agent.ID
	}

	// 全局限流检查
	if err := h.rateLimiter.AcquireGlobal(c.Request.Context()); err != nil {
		h.metrics.RecordFailure("rate_limit_global")
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "请求过于频繁，请稍后重试"})
		return
	}
	defer h.rateLimiter.ReleaseGlobal()

	// Token 预算检查
	if !h.rateLimiter.CheckTokenBudget(100000) { // 预估每次搜索最多 100K tokens
		h.metrics.RecordFailure("token_budget_exceeded")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Token 预算已用尽，请稍后重试"})
		return
	}

	out, err := h.contextSvc.Search(c.Request.Context(), service.SearchInput{
		CompanyID:    companyID,
		AgentID:      agentID,
		Query:        req.Query,
		DirectoryIDs: req.DirectoryIDs,
		// --- 新增统一参数 ---
		MaxResults:   req.MaxResults,
		MinRelevance: req.MinRelevance,
		TimeoutMs:    req.TimeoutMs,
		UseIndex:     req.UseIndex,
	})
	if err != nil {
		h.metrics.RecordFailure("search_failed")
		h.metrics.RecordLatency(time.Since(start).Milliseconds())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 检查是否有错误
	if out.Error != nil {
		h.metrics.RecordFailure(string(out.Error.Code))
		h.metrics.RecordLatency(time.Since(start).Milliseconds())
		errorCode := http.StatusBadRequest
		switch out.Error.Code {
		case domain.ErrNoDirectories:
			errorCode = http.StatusBadRequest
		case domain.ErrTimeout:
			errorCode = http.StatusGatewayTimeout
		case domain.ErrLLMUnavailable:
			errorCode = http.StatusServiceUnavailable
		}
		c.JSON(errorCode, gin.H{
			"ok":         false,
			"request_id": uuid.NewString(), // TODO: 使用请求追踪 ID
			"error":      out.Error,
		})
		return
	}

	// 记录成功指标
	h.metrics.RecordSearch(true, out.Diagnostics != nil && out.Diagnostics.FallbackReason != "")
	h.metrics.RecordLatency(time.Since(start).Milliseconds())

	// 成本预算闭环：从 LLM client 获取实际 cost 并更新预算
	llmMetrics := h.contextSvc.GetLLMMetrics()
	if llmMetrics != nil {
		snapshot := llmMetrics.GetSnapshot()
		// 将实际 cost 记录到预算系统
		if snapshot.TotalCost > 0 {
			allowed, warned, _ := h.costBudget.CheckAndRecord(snapshot.TotalCost)
			if !allowed {
				h.metrics.RecordFailure("cost_budget_exceeded")
			} else if warned {
				h.metrics.RecordFailure("cost_budget_warning")
			}
		}
	}

	// 返回统一响应结构
	c.JSON(http.StatusOK, gin.H{
		"ok":          true,
		"request_id":  uuid.NewString(),
		"latency_ms":  out.LatencyMs,
		"results":     out.Results,
		"total":       len(out.Results),
		"diagnostics": out.Diagnostics,
	})
}

type agentSearchRequest struct {
	Query        string   `json:"query" binding:"required"`
	DirectoryIDs []string `json:"directory_ids"`
	// --- 新增统一参数 (v1.1) ---
	MaxResults   int     `json:"max_results,omitempty"`
	MinRelevance float64 `json:"min_relevance,omitempty"`
	TimeoutMs    int     `json:"timeout_ms,omitempty"`
	MaxTurns     int     `json:"max_turns,omitempty"`
}

func (h *contextHandler) agentSearch(c *gin.Context) {
	start := time.Now()
	var req agentSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.metrics.RecordFailure("invalid_request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID := currentCompanyID(c)
	ctx := c.Request.Context()

	// 全局限流检查
	if err := h.rateLimiter.AcquireGlobal(ctx); err != nil {
		h.metrics.RecordFailure("rate_limit_global")
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "请求过于频繁，请稍后重试"})
		return
	}
	defer h.rateLimiter.ReleaseGlobal()

	// Get directories to search
	var dirs []*domain.ContextDirectory
	allDirs, err := h.contextSvc.ListDirectories(ctx, companyID)
	if err != nil {
		h.metrics.RecordFailure("list_directories_failed")
		h.metrics.RecordLatency(time.Since(start).Milliseconds())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(req.DirectoryIDs) > 0 {
		dirMap := make(map[string]bool)
		for _, id := range req.DirectoryIDs {
			dirMap[id] = true
		}
		for _, d := range allDirs {
			if dirMap[d.ID] && d.IsActive {
				dirs = append(dirs, d)
			}
		}
	} else {
		for _, d := range allDirs {
			if d.IsActive {
				dirs = append(dirs, d)
			}
		}
	}

	if len(dirs) == 0 {
		h.metrics.RecordFailure("no_directories")
		h.metrics.RecordLatency(time.Since(start).Milliseconds())
		// 返回统一错误结构
		c.JSON(http.StatusOK, gin.H{
			"ok":         false,
			"request_id": uuid.NewString(),
			"error": map[string]any{
				"code":    "NO_DIRECTORIES",
				"message": "没有可搜索的目录。请先配置上下文目录。",
			},
		})
		return
	}

	// Collect directory IDs
	dirIDs := make([]string, 0, len(dirs))
	maxFileSize := 0
	for _, d := range dirs {
		dirIDs = append(dirIDs, d.ID)
		if d.MaxFileSize > maxFileSize {
			maxFileSize = d.MaxFileSize
		}
	}

	maxTurns := req.MaxTurns
	if maxTurns <= 0 {
		maxTurns = 10
	}

	out, err := h.agent.Search(ctx, service.AgentSearchInput{
		CompanyID:    companyID,
		Query:        req.Query,
		DirectoryIDs: dirIDs,
		MaxTurns:     maxTurns,
		MaxFileSize:  int64(maxFileSize),
	})
	if err != nil {
		h.metrics.RecordFailure("agent_search_failed")
		h.metrics.RecordLatency(time.Since(start).Milliseconds())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 记录成功指标
	h.metrics.RecordSearch(true, false) // agent search 不使用索引降级
	h.metrics.RecordLatency(time.Since(start).Milliseconds())

	// 返回统一响应结构
	c.JSON(http.StatusOK, gin.H{
		"ok":         true,
		"request_id": uuid.NewString(),
		"latency_ms": out.LatencyMs,
		"answer":     out.Answer,
		"files_read": out.Files,
		"diagnostics": map[string]any{
			"directories_scanned": len(dirs),
			"files_read":          len(out.Files),
		},
	})
}

// rebuildIndexRequest 重建索引请求
type rebuildIndexRequest struct {
	DirectoryIDs []string `json:"directory_ids"` // 空则重建所有
}

// rebuildIndex 手动触发重建索引
func (h *contextHandler) rebuildIndex(c *gin.Context) {
	var req rebuildIndexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID := currentCompanyID(c)
	ctx := c.Request.Context()

	// 如果指定了 directory IDs，只重建这些；否则重建所有活跃目录
	var dirIDs []string
	if len(req.DirectoryIDs) > 0 {
		// 验证这些 directory 属于当前公司
		for _, id := range req.DirectoryIDs {
			d, err := h.contextSvc.GetDirectoryByID(ctx, id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if d == nil || d.CompanyID != companyID {
				c.JSON(http.StatusNotFound, gin.H{"error": "directory not found: " + id})
				return
			}
			dirIDs = append(dirIDs, id)
		}
	} else {
		// 获取所有活跃目录
		dirs, err := h.contextSvc.ListDirectories(ctx, companyID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, d := range dirs {
			if d.IsActive {
				dirIDs = append(dirIDs, d.ID)
			}
		}
	}

	// 触发重建索引
	for _, id := range dirIDs {
		if err := h.scheduler.TriggerScan(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to trigger scan for " + id + ": " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "Index rebuild triggered",
		"count":   len(dirIDs),
	})
}
