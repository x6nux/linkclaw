package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
	"github.com/linkclaw/backend/internal/service"
)

type memoryHandler struct {
	memorySvc *service.MemoryService
}

func (h *memoryHandler) list(c *gin.Context) {
	agent := currentAgent(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	q := repository.MemoryQuery{
		CompanyID: agent.CompanyID,
		Limit:     limit,
		Offset:    offset,
		OrderBy:   c.DefaultQuery("order_by", "created_at"),
	}
	if aid := c.Query("agent_id"); aid != "" {
		q.AgentID = aid
	}
	if cat := c.Query("category"); cat != "" {
		q.Category = cat
	}
	if imp := c.Query("importance"); imp != "" {
		v, err := strconv.Atoi(imp)
		if err == nil {
			q.Importance = &v
		}
	}

	mems, total, err := h.memorySvc.List(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": mems, "total": total})
}

func (h *memoryHandler) get(c *gin.Context) {
	m, err := h.memorySvc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil || m == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, m)
}

type createMemoryRequest struct {
	AgentID    string   `json:"agent_id" binding:"required"`
	Content    string   `json:"content" binding:"required"`
	Category   string   `json:"category"`
	Tags       []string `json:"tags"`
	Importance int      `json:"importance"`
	Source     string   `json:"source"`
}

func (h *memoryHandler) create(c *gin.Context) {
	var req createMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	agent := currentAgent(c)
	source := domain.MemorySource(req.Source)
	if source == "" {
		source = domain.SourceManual
	}

	m, err := h.memorySvc.Create(c.Request.Context(), service.CreateMemoryInput{
		CompanyID:  agent.CompanyID,
		AgentID:    req.AgentID,
		Content:    req.Content,
		Category:   req.Category,
		Tags:       req.Tags,
		Importance: domain.MemoryImportance(req.Importance),
		Source:     source,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, m)
}

type updateMemoryRequest struct {
	Content    string   `json:"content" binding:"required"`
	Category   string   `json:"category"`
	Tags       []string `json:"tags"`
	Importance int      `json:"importance"`
}

func (h *memoryHandler) update(c *gin.Context) {
	var req updateMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m, err := h.memorySvc.Update(c.Request.Context(), c.Param("id"), service.UpdateMemoryInput{
		Content:    req.Content,
		Category:   req.Category,
		Tags:       req.Tags,
		Importance: domain.MemoryImportance(req.Importance),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *memoryHandler) delete(c *gin.Context) {
	if err := h.memorySvc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type searchMemoryRequest struct {
	Query string `json:"query" binding:"required"`
	Limit int    `json:"limit"`
}

func (h *memoryHandler) search(c *gin.Context) {
	var req searchMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	agent := currentAgent(c)
	agentID := c.Query("agent_id")
	if agentID == "" {
		agentID = agent.ID
	}

	mems, err := h.memorySvc.SemanticSearch(c.Request.Context(), agent.CompanyID, agentID, req.Query, req.Limit)
	if err != nil {
		if strings.Contains(err.Error(), "no active") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "未配置 OpenAI Provider，无法使用语义搜索"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": mems, "total": len(mems)})
}

type batchDeleteRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

func (h *memoryHandler) batchDelete(c *gin.Context) {
	var req batchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.memorySvc.BatchDelete(c.Request.Context(), req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "deleted": len(req.IDs)})
}
