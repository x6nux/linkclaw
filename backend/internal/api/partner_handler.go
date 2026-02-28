package api

import (
	"crypto/subtle"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/linkclaw/backend/internal/config"
	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
	"github.com/linkclaw/backend/internal/service"
)

type partnerHandler struct {
	messageSvc  *service.MessageService
	companyRepo repository.CompanyRepo
	agentCfg    *config.AgentConfig
}

// info 返回本公司对外公开信息（无鉴权）
func (h *partnerHandler) info(c *gin.Context) {
	if h.agentCfg.CompanySlug == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "company not configured"})
		return
	}
	company, err := h.companyRepo.GetBySlug(c.Request.Context(), h.agentCfg.CompanySlug)
	if err != nil || company == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}
	mcpURL := h.agentCfg.MCPPublicURL
	if company.MCPPublicURL != "" {
		mcpURL = company.MCPPublicURL
	}
	c.JSON(http.StatusOK, gin.H{
		"name":    company.Name,
		"slug":    company.Slug,
		"mcp_url": mcpURL,
	})
}

// receiveMessage 接收跨公司消息，注入到本地 #general 频道
func (h *partnerHandler) receiveMessage(c *gin.Context) {
	var req struct {
		FromCompany string `json:"from_company" binding:"required"`
		Content     string `json:"content"      binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.agentCfg.CompanySlug == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "company not configured"})
		return
	}
	company, err := h.companyRepo.GetBySlug(c.Request.Context(), h.agentCfg.CompanySlug)
	if err != nil || company == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}
	ch, err := h.companyRepo.GetChannelByName(c.Request.Context(), company.ID, "general")
	if err != nil || ch == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "general channel not found"})
		return
	}

	content := fmt.Sprintf("[来自 %s] %s", req.FromCompany, req.Content)
	chID := ch.ID
	msg := &domain.Message{
		ID:        uuid.New().String(),
		CompanyID: company.ID,
		ChannelID: &chID,
		Content:   content,
		MsgType:   domain.MsgTypeText,
	}
	if err := h.messageSvc.SendRaw(c.Request.Context(), msg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deliver message"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "delivered"})
}

// partnerAuthMiddleware 验证 X-Partner-Key 头（使用常量时间比较防止时序攻击）
func partnerAuthMiddleware(agentCfg *config.AgentConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if agentCfg.PartnerAPIKey == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "partner api not enabled"})
			return
		}
		key := c.GetHeader("X-Partner-Key")
		if subtle.ConstantTimeCompare([]byte(key), []byte(agentCfg.PartnerAPIKey)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid partner key"})
			return
		}
		c.Next()
	}
}
