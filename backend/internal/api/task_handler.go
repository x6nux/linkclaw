package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
	"github.com/linkclaw/backend/internal/service"
)

type taskHandler struct {
	taskSvc *service.TaskService
}

func (h *taskHandler) list(c *gin.Context) {
	companyID := currentCompanyID(c)
	q := repository.TaskQuery{
		CompanyID: companyID,
		Status:    domain.TaskStatus(c.Query("status")),
		Priority:  domain.TaskPriority(c.Query("priority")),
	}
	if aid := c.Query("assignee_id"); aid != "" {
		q.AssigneeID = aid
	}
	tasks, total, err := h.taskSvc.List(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tasks, "total": total})
}

func (h *taskHandler) get(c *gin.Context) {
	task, err := h.taskSvc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil || task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

type createTaskRequest struct {
	Title       string `json:"title"       binding:"required"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	AssigneeID  string `json:"assignee_id"`
	ParentID    string `json:"parent_id"`
}

func (h *taskHandler) create(c *gin.Context) {
	var req createTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	agent := currentAgent(c)
	agentID := agent.ID
	in := service.CreateTaskInput{
		CompanyID:   agent.CompanyID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    domain.TaskPriority(req.Priority),
		CreatedBy:   &agentID,
	}
	if req.AssigneeID != "" {
		in.AssigneeID = &req.AssigneeID
	}
	if req.ParentID != "" {
		in.ParentID = &req.ParentID
	}
	task, err := h.taskSvc.Create(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, task)
}

func (h *taskHandler) delete(c *gin.Context) {
	if err := h.taskSvc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
