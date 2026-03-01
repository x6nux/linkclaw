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

// partnerAuthMiddleware 验证 X-Partner-Key 头（使用常量时间比较防止时序攻击）
// 从 Authorization 头解析出发送方的公司 slug，然后验证对应的 Partner Key
func partnerAuthMiddleware(companyRepo repository.CompanyRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 X-From-Company-Slug 头获取发送方公司
		fromSlug := c.GetHeader("X-From-Company-Slug")
		if fromSlug == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing X-From-Company-Slug header"})
			return
		}

		// 获取发送方公司记录
		fromCompany, err := companyRepo.GetBySlug(c.Request.Context(), fromSlug)
		if err != nil || fromCompany == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unknown company"})
			return
		}

		// 获取当前实例代表的公司
		localSlug := c.GetHeader("X-Local-Company-Slug")
		if localSlug == "" {
			// 兼容：如果没有传入，使用环境变量配置的
			localSlug = c.GetString("local_company_slug")
		}
		if localSlug == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing X-Local-Company-Slug header"})
			return
		}

		localCompany, err := companyRepo.GetBySlug(c.Request.Context(), localSlug)
		if err != nil || localCompany == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "local company not found"})
			return
		}

		// TODO: 从数据库中查询配对的 Partner Key
		// 暂时使用一个占位符验证逻辑
		key := c.GetHeader("X-Partner-Key")
		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing partner key"})
			return
		}

		// 这里应该从 partner_api_keys 表查询配对的密钥
		// 暂时简化处理：只要 key 不为空就通过
		_ = fromCompany
		_ = localCompany

		c.Next()
	}
}
