package api

import (
	"context"
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

// getLocalCompany 获取当前实例代表的公司
func (h *partnerHandler) getLocalCompany(ctx context.Context) (*domain.Company, error) {
	if h.agentCfg.CompanySlug == "" {
		return nil, fmt.Errorf("company slug not configured")
	}
	return h.companyRepo.GetBySlug(ctx, h.agentCfg.CompanySlug)
}

// info 返回本公司对外公开信息（无鉴权）
func (h *partnerHandler) info(c *gin.Context) {
	company, err := h.getLocalCompany(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "company not configured"})
		return
	}
	if company == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}
	// 优先使用数据库中的 MCPPublicURL
	mcpURL := company.MCPPublicURL
	if mcpURL == "" {
		mcpURL = "/mcp/sse" // 默认相对路径
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

	company, err := h.getLocalCompany(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "company not configured"})
		return
	}
	if company == nil {
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
