package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/linkclaw/backend/internal/service"
)

type knowledgeHandler struct {
	knowledgeSvc *service.KnowledgeService
}

func (h *knowledgeHandler) list(c *gin.Context) {
	agent := currentAgent(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	docs, total, err := h.knowledgeSvc.List(c.Request.Context(), agent.CompanyID, limit, offset)
	if err != nil {
		ErrorToResponse(c, InternalError("failed to list knowledge documents"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": docs, "total": total})
}

func (h *knowledgeHandler) search(c *gin.Context) {
	agent := currentAgent(c)
	query := c.Query("q")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if query == "" {
		ErrorToResponse(c, InvalidParamError("query parameter 'q' is required"))
		return
	}
	docs, err := h.knowledgeSvc.Search(c.Request.Context(), agent.CompanyID, query, limit)
	if err != nil {
		ErrorToResponse(c, InternalError("failed to search knowledge documents"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": docs, "total": len(docs)})
}

func (h *knowledgeHandler) get(c *gin.Context) {
	doc, err := h.knowledgeSvc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil || doc == nil {
		ErrorToResponse(c, NotFoundError("knowledge document"))
		return
	}
	c.JSON(http.StatusOK, doc)
}

type writeDocRequest struct {
	Title   string `json:"title"   binding:"required"`
	Content string `json:"content" binding:"required"`
	Tags    string `json:"tags"`
}

func (h *knowledgeHandler) write(c *gin.Context) {
	var req writeDocRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorToResponse(c, ValidationError("invalid request body: "+err.Error()))
		return
	}
	agent := currentAgent(c)
	doc, err := h.knowledgeSvc.Write(c.Request.Context(), service.WriteDocInput{
		DocID:     c.Param("id"),
		CompanyID: agent.CompanyID,
		AuthorID:  agent.ID,
		Title:     req.Title,
		Content:   req.Content,
		Tags:      req.Tags,
	})
	if err != nil {
		ErrorToResponse(c, InvalidParamError(err.Error()))
		return
	}
	status := http.StatusCreated
	if c.Param("id") != "" {
		status = http.StatusOK
	}
	c.JSON(status, doc)
}

func (h *knowledgeHandler) delete(c *gin.Context) {
	if err := h.knowledgeSvc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		ErrorToResponse(c, InternalError("failed to delete knowledge document"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
