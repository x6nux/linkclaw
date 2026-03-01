package api

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/linkclaw/backend/internal/repository"
	"github.com/linkclaw/backend/internal/service"
)

type partnerSettingsHandler struct {
	partnerSvc   *service.PartnerService
	partnerKeyRepo repository.PartnerAPIKeyRepo
}

// PartnerSettingsRequest 生成/更新配对密钥请求
type PartnerSettingsRequest struct {
	PartnerSlug string `json:"partner_slug" binding:"required"`
	Name        string `json:"name"`
}

// PartnerSettingsResponse 配对密钥响应
type PartnerSettingsResponse struct {
	ID           string  `json:"id"`
	PartnerSlug  string  `json:"partner_slug"`
	PartnerName  string  `json:"partner_name"`
	KeyPrefix    string  `json:"key_prefix"`
	RawKey       *string `json:"raw_key,omitempty"` // 仅在生成时返回
	IsActive     bool    `json:"is_active"`
	LastUsedAt   *string `json:"last_used_at,omitempty"`
}

// NewPartnerSettingsHandler 创建配对设置 Handler
func NewPartnerSettingsHandler(partnerSvc *service.PartnerService, partnerKeyRepo repository.PartnerAPIKeyRepo) *partnerSettingsHandler {
	return &partnerSettingsHandler{
		partnerSvc:     partnerSvc,
		partnerKeyRepo: partnerKeyRepo,
	}
}

// createOrUpdateKey 生成或更新配对密钥
// POST /api/v1/partner/settings
func (h *partnerSettingsHandler) createOrUpdateKey(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req PartnerSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name == "" {
		req.Name = "Partner API Key"
	}

	rawKey, err := h.partnerSvc.GenerateKey(c.Request.Context(), companyID, req.PartnerSlug, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取生成的密钥记录
	k, err := h.partnerSvc.GetKey(c.Request.Context(), companyID, req.PartnerSlug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取配对公司信息
	partnerName := req.PartnerSlug
	// 这里可以进一步获取配对公司详细信息

	resp := PartnerSettingsResponse{
		ID:          k.ID,
		PartnerSlug: k.PartnerSlug,
		PartnerName: partnerName,
		KeyPrefix:   k.KeyPrefix,
		RawKey:      &rawKey, // 仅显示一次
		IsActive:    k.IsActive,
	}

	c.JSON(http.StatusOK, resp)
}

// getSettings 查看配对配置
// GET /api/v1/partner/settings/:partner_slug
func (h *partnerSettingsHandler) getSettings(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	partnerSlug := c.Param("partner_slug")
	if partnerSlug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing partner_slug"})
		return
	}

	k, err := h.partnerSvc.GetKey(c.Request.Context(), companyID, partnerSlug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if k == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "partner key not found"})
		return
	}

	resp := PartnerSettingsResponse{
		ID:          k.ID,
		PartnerSlug: k.PartnerSlug,
		PartnerName: partnerSlug,
		KeyPrefix:   k.KeyPrefix,
		IsActive:    k.IsActive,
	}
	if k.LastUsedAt != nil {
		s := k.LastUsedAt.Format("2006-01-02T15:04:05Z")
		resp.LastUsedAt = &s
	}

	c.JSON(http.StatusOK, resp)
}

// revokeKey 撤销配对
// DELETE /api/v1/partner/settings/:partner_slug
func (h *partnerSettingsHandler) revokeKey(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	partnerSlug := c.Param("partner_slug")
	if partnerSlug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing partner_slug"})
		return
	}

	if err := h.partnerSvc.RevokeKey(c.Request.Context(), companyID, partnerSlug); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "revoked"})
}

// partnerAuthMiddleware 验证 X-Partner-Key 头（使用常量时间比较防止时序攻击）
func partnerAuthMiddleware(companyRepo repository.CompanyRepo, partnerKeyRepo repository.PartnerAPIKeyRepo) gin.HandlerFunc {
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

		// 获取当前实例代表的公司（接收方）
		localSlug := c.GetHeader("X-Local-Company-Slug")
		if localSlug == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing X-Local-Company-Slug header"})
			return
		}

		localCompany, err := companyRepo.GetBySlug(c.Request.Context(), localSlug)
		if err != nil || localCompany == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "local company not found"})
			return
		}

		// 获取并验证 Partner Key
		key := c.GetHeader("X-Partner-Key")
		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing X-Partner-Key header"})
			return
		}

		// 计算 key 的 hash
		hash := sha256.Sum256([]byte(key))
		keyHash := hex.EncodeToString(hash[:])

		// 从数据库查询匹配的密钥
		k, err := partnerKeyRepo.GetByKeyHash(c.Request.Context(), localCompany.ID, keyHash)
		if err != nil || k == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid partner key"})
			return
		}

		// 验证配对关系（发送方必须是密钥持有者）
		if k.CompanyID != fromCompany.ID {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "partner key mismatch"})
			return
		}

		// 验证通过，将发送方公司信息存入上下文
		c.Set("from_company", fromCompany)
		c.Set("local_company", localCompany)
		c.Set("partner_key_id", k.ID)

		c.Next()
	}
}
