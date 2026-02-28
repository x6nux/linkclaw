package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/service"
)

type indexingHandler struct {
	indexingSvc *service.IndexingService
}

type createIndexTaskRequest struct {
	RepositoryURL string `json:"repository_url" binding:"required"`
	Branch        string `json:"branch"`
}

type searchCodeRequest struct {
	Query string `json:"query" binding:"required"`
	Limit int    `json:"limit"`
}

type searchTaskCodeRequest struct {
	Query string `form:"query" binding:"required"`
	Limit int    `form:"limit"`
}

type grantTaskAccessRequest struct {
	AgentID string `json:"agent_id" binding:"required"`
}

type indexTaskResponse struct {
	ID            string             `json:"id"`
	RepositoryURL string             `json:"repository_url"`
	Branch        string             `json:"branch"`
	Status        domain.IndexStatus `json:"status"`
	TotalFiles    int                `json:"total_files"`
	IndexedFiles  int                `json:"indexed_files"`
	ErrorMessage  string             `json:"error_message,omitempty"`
	StartedAt     *time.Time         `json:"started_at,omitempty"`
	CompletedAt   *time.Time         `json:"completed_at,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
}

func (h *indexingHandler) createTask(c *gin.Context) {
	if h.indexingSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "indexing service not configured"})
		return
	}

	var req createIndexTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	repoURL := strings.TrimSpace(req.RepositoryURL)
	if repoURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "repository_url is required"})
		return
	}

	task, err := h.indexingSvc.IndexRepository(c.Request.Context(), currentCompanyID(c), repoURL, strings.TrimSpace(req.Branch))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	agent := currentAgent(c)
	if agent != nil {
		if err := h.indexingSvc.GrantTaskAccess(c.Request.Context(), currentCompanyID(c), task.ID, agent.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, toIndexTaskResponse(task))
}

func (h *indexingHandler) listTasks(c *gin.Context) {
	if h.indexingSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "indexing service not configured"})
		return
	}

	tasks, err := h.indexingSvc.ListIndexTasks(c.Request.Context(), currentCompanyID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := make([]indexTaskResponse, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, toIndexTaskResponse(t))
	}
	c.JSON(http.StatusOK, out)
}

func (h *indexingHandler) getStatus(c *gin.Context) {
	if h.indexingSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "indexing service not configured"})
		return
	}

	taskID := strings.TrimSpace(c.Param("id"))
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task id is required"})
		return
	}

	task, err := h.indexingSvc.GetIndexStatus(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if task == nil || task.CompanyID != currentCompanyID(c) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, toIndexTaskResponse(task))
}

func (h *indexingHandler) search(c *gin.Context) {
	if h.indexingSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "indexing service not configured"})
		return
	}

	var req searchCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := strings.TrimSpace(req.Query)
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	results, err := h.indexingSvc.SearchCode(c.Request.Context(), currentCompanyID(c), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toSearchCodeResponse(results))
}

func (h *indexingHandler) grantAccess(c *gin.Context) {
	if h.indexingSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "indexing service not configured"})
		return
	}

	taskID := strings.TrimSpace(c.Param("id"))
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task id is required"})
		return
	}

	var req grantTaskAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.indexingSvc.GrantTaskAccess(c.Request.Context(), currentCompanyID(c), taskID, strings.TrimSpace(req.AgentID)); err != nil {
		if handled := writeIndexTaskServiceError(c, err); handled {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

func (h *indexingHandler) revokeAccess(c *gin.Context) {
	if h.indexingSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "indexing service not configured"})
		return
	}

	taskID := strings.TrimSpace(c.Param("id"))
	agentID := strings.TrimSpace(c.Param("agent_id"))
	if taskID == "" || agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task id and agent id are required"})
		return
	}

	if err := h.indexingSvc.RevokeTaskAccess(c.Request.Context(), currentCompanyID(c), taskID, agentID); err != nil {
		if handled := writeIndexTaskServiceError(c, err); handled {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *indexingHandler) listAuthorizedAgents(c *gin.Context) {
	if h.indexingSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "indexing service not configured"})
		return
	}

	taskID := strings.TrimSpace(c.Param("id"))
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task id is required"})
		return
	}

	agents, err := h.indexingSvc.ListTaskAgents(c.Request.Context(), currentCompanyID(c), taskID)
	if err != nil {
		if handled := writeIndexTaskServiceError(c, err); handled {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, agents)
}

func (h *indexingHandler) searchTask(c *gin.Context) {
	if h.indexingSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "indexing service not configured"})
		return
	}

	taskID := strings.TrimSpace(c.Param("id"))
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task id is required"})
		return
	}

	var req searchTaskCodeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := strings.TrimSpace(req.Query)
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}

	agent := currentAgent(c)
	if agent == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	results, err := h.indexingSvc.SearchTaskCode(c.Request.Context(), currentCompanyID(c), taskID, agent.ID, query, limit)
	if err != nil {
		if handled := writeIndexTaskServiceError(c, err); handled {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toSearchCodeResponse(results))
}

func (h *indexingHandler) retryTask(c *gin.Context) {
	if h.indexingSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "indexing service not configured"})
		return
	}

	taskID := strings.TrimSpace(c.Param("id"))
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task id is required"})
		return
	}

	task, err := h.indexingSvc.RetryIndexTask(c.Request.Context(), currentCompanyID(c), taskID)
	if err != nil {
		if handled := writeIndexTaskServiceError(c, err); handled {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toIndexTaskResponse(task))
}

func toSearchCodeResponse(results []*service.SearchResult) []gin.H {
	out := make([]gin.H, 0, len(results))
	for _, r := range results {
		out = append(out, gin.H{
			"id":         r.ID,
			"file_path":  r.Payload["file_path"],
			"content":    r.Payload["content"],
			"score":      r.Score,
			"start_line": r.Payload["start_line"],
			"end_line":   r.Payload["end_line"],
			"language":   r.Payload["language"],
			"symbols":    r.Payload["symbols"],
		})
	}
	return out
}

func writeIndexTaskServiceError(c *gin.Context, err error) bool {
	switch {
	case errors.Is(err, service.ErrIndexTaskNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return true
	case errors.Is(err, service.ErrIndexTaskAccessDenied):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return true
	case strings.Contains(err.Error(), "is required"):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return true
	default:
		return false
	}
}

func toIndexTaskResponse(task *domain.IndexTask) indexTaskResponse {
	return indexTaskResponse{
		ID:            task.ID,
		RepositoryURL: task.RepositoryURL,
		Branch:        task.Branch,
		Status:        task.Status,
		TotalFiles:    task.TotalFiles,
		IndexedFiles:  task.IndexedFiles,
		ErrorMessage:  task.ErrorMessage,
		StartedAt:     task.StartedAt,
		CompletedAt:   task.CompletedAt,
		CreatedAt:     task.CreatedAt,
	}
}
