package llm

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/linkclaw/backend/internal/domain"
)

// Handler HTTP 处理器：管理 API + 代理端点
type Handler struct {
	repo    *Repository
	proxy   *ProxyService
	router  *Router
	encKey  string
}

func NewHandler(repo *Repository, proxy *ProxyService, router *Router, encKey string) *Handler {
	return &Handler{repo: repo, proxy: proxy, router: router, encKey: encKey}
}

// ===== Provider 管理 API =====

func (h *Handler) ListProviders(c *gin.Context) {
	companyID := c.GetString("company_id")
	providers, err := h.repo.ListProviders(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	views := make([]*ProviderView, len(providers))
	for i, p := range providers {
		rawKey, _ := DecryptAPIKey(p.APIKeyEnc, h.encKey)
		views[i] = &ProviderView{
			Provider:     *p,
			Status:       h.router.GetStatus(p),
			APIKeyPrefix: APIKeyPrefix(rawKey),
		}
		views[i].APIKeyEnc = "" // 不回传加密 key
	}
	c.JSON(http.StatusOK, gin.H{"data": views, "total": len(views)})
}

type createProviderRequest struct {
	Name     string   `json:"name"     binding:"required"`
	Type     string   `json:"type"     binding:"required"`
	BaseURL  string   `json:"base_url" binding:"required"`
	APIKey   string   `json:"api_key"  binding:"required"`
	Models   []string `json:"models"`
	Weight   int      `json:"weight"`
	IsActive *bool    `json:"is_active"`
	MaxRPM   *int     `json:"max_rpm"`
}

func (h *Handler) CreateProvider(c *gin.Context) {
	var req createProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	enc, err := EncryptAPIKey(req.APIKey, h.encKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "encrypt key failed"})
		return
	}

	weight := req.Weight
	if weight <= 0 {
		weight = 100
	}
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}

	models := req.Models
	if len(models) == 0 {
		models = []string{}
	}
	p := &Provider{
		CompanyID: c.GetString("company_id"),
		Name:      req.Name,
		Type:      ProviderType(req.Type),
		BaseURL:   req.BaseURL,
		APIKeyEnc: enc,
		Models:    models,
		Weight:    weight,
		IsActive:  active,
		MaxRPM:    req.MaxRPM,
	}
	if err := h.repo.CreateProvider(c.Request.Context(), p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	p.APIKeyEnc = ""
	c.JSON(http.StatusCreated, p)
}

func (h *Handler) UpdateProvider(c *gin.Context) {
	id := c.Param("id")
	p, err := h.repo.GetProvider(c.Request.Context(), id)
	if err != nil || p == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var req createProviderRequest
	c.ShouldBindJSON(&req) //nolint:errcheck

	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Type != "" {
		p.Type = ProviderType(req.Type)
	}
	if req.BaseURL != "" {
		p.BaseURL = req.BaseURL
	}
	if req.APIKey != "" {
		enc, err := EncryptAPIKey(req.APIKey, h.encKey)
		if err == nil {
			p.APIKeyEnc = enc
		}
	}
	if len(req.Models) > 0 {
		p.Models = req.Models
	}
	if req.Weight > 0 {
		p.Weight = req.Weight
	}
	if req.IsActive != nil {
		p.IsActive = *req.IsActive
	}
	p.MaxRPM = req.MaxRPM

	if err := h.repo.UpdateProvider(c.Request.Context(), p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	p.APIKeyEnc = ""
	c.JSON(http.StatusOK, p)
}

func (h *Handler) DeleteProvider(c *gin.Context) {
	if err := h.repo.DeleteProvider(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ===== 统计 API =====

func (h *Handler) GetStats(c *gin.Context) {
	companyID := c.GetString("company_id")
	stats, err := h.repo.GetUsageStats(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	daily, err := h.repo.GetDailyUsage(c.Request.Context(), companyID)
	if err != nil {
		daily = nil
	}
	logs, _ := h.repo.GetRecentLogs(c.Request.Context(), companyID, 50)
	c.JSON(http.StatusOK, gin.H{
		"providers": stats,
		"daily":     daily,
		"recent":    logs,
		"models":    KnownModels(),
	})
}

// ===== 代理端点 =====

// ProxyAnthropic  Anthropic API 兼容代理（/v1/messages*）
func (h *Handler) ProxyAnthropic(c *gin.Context) {
	h.doProxy(c, ProviderAnthropic)
}

// ProxyOpenAI  OpenAI API 兼容代理（/v1/chat/*, /v1/completions, /v1/embeddings, /v1/models）
func (h *Handler) ProxyOpenAI(c *gin.Context) {
	h.doProxy(c, ProviderOpenAI)
}

func (h *Handler) doProxy(c *gin.Context, pt ProviderType) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "read body failed"})
		return
	}

	companyID := c.GetString("company_id")
	agentID := ""
	if a, ok := c.Get("agent"); ok {
		if agent, ok := a.(*domain.Agent); ok {
			agentID = agent.ID
		}
	}

	if err := h.proxy.ProxyRequest(
		c.Request.Context(),
		c.Writer,
		c.Request,
		body,
		companyID, agentID,
		pt,
	); err != nil {
		if !c.Writer.Written() {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		}
	}
}
