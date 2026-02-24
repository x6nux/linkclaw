package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/linkclaw/backend/internal/service"
)

type promptHandler struct {
	promptSvc *service.PromptService
	agentSvc  *service.AgentService
}

// list 列出所有提示词层
func (h *promptHandler) list(c *gin.Context) {
	companyID := currentCompanyID(c)
	result, err := h.promptSvc.ListAll(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

type upsertPromptRequest struct {
	Content string `json:"content"`
}

// upsert 创建/更新提示词层
func (h *promptHandler) upsert(c *gin.Context) {
	layerType := c.Param("type")
	key := c.Param("key")
	companyID := currentCompanyID(c)

	var req upsertPromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.promptSvc.Upsert(c.Request.Context(), companyID, layerType, key, req.Content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// remove 删除提示词层
func (h *promptHandler) remove(c *gin.Context) {
	layerType := c.Param("type")
	key := c.Param("key")
	companyID := currentCompanyID(c)

	if err := h.promptSvc.Delete(c.Request.Context(), companyID, layerType, key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// preview 预览 Agent 的完整拼接提示词
func (h *promptHandler) preview(c *gin.Context) {
	agentID := c.Param("agentId")
	agent, err := h.agentSvc.GetByID(c.Request.Context(), agentID)
	if err != nil || agent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent 不存在"})
		return
	}

	assembled := h.promptSvc.AssembleForAgent(c.Request.Context(), agent)
	c.JSON(http.StatusOK, gin.H{"prompt": assembled})
}
