package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/linkclaw/backend/internal/i18n"
	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
	"github.com/linkclaw/backend/internal/service"
)

type setupHandler struct {
	companyRepo repository.CompanyRepo
	agentRepo   repository.AgentRepo
	agentSvc    *service.AgentService
	jwtSecret   string
	jwtExpiry   int
}

func (h *setupHandler) status(c *gin.Context) {
	company, err := h.companyRepo.FindFirst(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, i18n.ErrInternalServerError)})
		return
	}
	if company == nil {
		c.JSON(http.StatusOK, gin.H{"initialized": false, "companySlug": ""})
		return
	}
	c.JSON(http.StatusOK, gin.H{"initialized": true, "companySlug": company.Slug})
}

type initRequest struct {
	CompanyName string `json:"company_name" binding:"required"`
	CompanySlug string `json:"company_slug" binding:"required"`
	AdminName   string `json:"admin_name"   binding:"required"`
	Password    string `json:"password"    binding:"required,min=8"`
}

func (h *setupHandler) initialize(c *gin.Context) {
	var req initRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(c, i18n.ErrBadRequest)})
		return
	}

	ctx := c.Request.Context()

	existing, err := h.companyRepo.FindFirst(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, i18n.ErrInternalServerError)})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": i18n.T(c, i18n.ErrConflict)})
		return
	}

	company := &domain.Company{
		ID:   uuid.New().String(),
		Name: req.CompanyName,
		Slug: req.CompanySlug,
	}
	if err := h.companyRepo.Create(ctx, company); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, i18n.ErrInternalServerError)})
		return
	}

	for _, ch := range domain.DefaultChannels {
		channel := &domain.Channel{
			ID:          uuid.New().String(),
			CompanyID:   company.ID,
			Name:        ch.Name,
			Description: ch.Description,
			IsDefault:   ch.IsDefault,
		}
		_ = h.companyRepo.CreateChannel(ctx, channel)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, i18n.ErrInternalServerError)})
		return
	}

	meta := domain.PositionMetaByPosition[domain.PositionChairman]
	out, err := h.agentSvc.Create(ctx, service.CreateAgentInput{
		CompanyID: company.ID,
		Name:      req.AdminName,
		RoleType:  domain.RoleChairman,
		Position:  domain.PositionChairman,
		IsHuman:   true,
		Persona:   meta.DefaultPersona,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, i18n.ErrInternalServerError)})
		return
	}

	if err := h.agentRepo.SetPasswordHash(ctx, out.Agent.ID, string(hash)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, i18n.ErrInternalServerError)})
		return
	}

	// 初始化即在线
	_ = h.agentRepo.UpdateStatus(ctx, out.Agent.ID, domain.StatusOnline)
	_ = h.agentRepo.UpdateLastSeen(ctx, out.Agent.ID)
	out.Agent.Status = domain.StatusOnline

	token, err := generateJWT(out.Agent.ID, h.jwtSecret, h.jwtExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, i18n.ErrInternalServerError)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"agent":   out.Agent,
		"company": company,
	})
}
