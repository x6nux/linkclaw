package api

import (
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
	c.JSON(http.StatusOK, out)
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
