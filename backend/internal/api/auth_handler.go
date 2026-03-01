package api

import (
	"crypto/subtle"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/linkclaw/backend/internal/i18n"
	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
)

type authHandler struct {
	agentRepo   repository.AgentRepo
	companyRepo repository.CompanyRepo
	jwtSecret   string
	jwtExpiry   int
	resetSecret string
}

type loginRequest struct {
	Name     string `json:"name"     binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *authHandler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(c, i18n.ErrBadRequest)})
		return
	}

	company, err := h.companyRepo.FindFirst(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, i18n.ErrInternalServerError)})
		return
	}
	if company == nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": i18n.T(c, i18n.ErrNotInitialized)})
		return
	}

	agent, err := h.agentRepo.GetByName(c.Request.Context(), company.ID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, i18n.ErrInternalServerError)})
		return
	}

	// 防止用户名枚举时序攻击：即使用户不存在也执行 bcrypt 比较
	// 预计算的 dummy bcrypt hash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	var passwordHash string
	if agent != nil && agent.PasswordHash != nil {
		passwordHash = *agent.PasswordHash
	} else {
		// 使用一个预计算的 bcrypt hash 作为 dummy，防止时序攻击
		passwordHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(c, i18n.ErrUnauthorized)})
		return
	}

	// 如果使用 dummy hash，说明用户不存在
	if agent == nil || agent.PasswordHash == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(c, i18n.ErrUnauthorized)})
		return
	}

	// 登录即在线
	ctx := c.Request.Context()
	_ = h.agentRepo.UpdateStatus(ctx, agent.ID, domain.StatusOnline)
	_ = h.agentRepo.UpdateLastSeen(ctx, agent.ID)
	agent.Status = "online"

	token, err := generateJWT(agent.ID, h.jwtSecret, h.jwtExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, i18n.ErrInternalServerError)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "agent": agent})
}

func generateJWT(agentID, secret string, expiryHours int) (string, error) {
	claims := jwt.MapClaims{
		"sub": agentID,
		"exp": time.Now().Add(time.Duration(expiryHours) * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func (h *authHandler) logout(c *gin.Context) {
	agent := currentAgent(c)
	if agent != nil {
		_ = h.agentRepo.UpdateStatus(c.Request.Context(), agent.ID, domain.StatusOffline)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type resetPasswordRequest struct {
	Name        string `json:"name"        binding:"required"`
	ResetSecret string `json:"reset_secret" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func (h *authHandler) resetPassword(c *gin.Context) {
	if h.resetSecret == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "password reset not configured"})
		return
	}

	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(c, i18n.ErrBadRequest)})
		return
	}

	if subtle.ConstantTimeCompare([]byte(req.ResetSecret), []byte(h.resetSecret)) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(c, i18n.ErrUnauthorized)})
		return
	}

	ctx := c.Request.Context()
	company, err := h.companyRepo.FindFirst(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, i18n.ErrInternalServerError)})
		return
	}
	if company == nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": i18n.T(c, i18n.ErrNotInitialized)})
		return
	}

	agent, err := h.agentRepo.GetByName(ctx, company.ID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, i18n.ErrInternalServerError)})
		return
	}
	if agent == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(c, i18n.ErrUnauthorized)})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, i18n.ErrInternalServerError)})
		return
	}
	if err := h.agentRepo.SetPasswordHash(ctx, agent.ID, string(hash)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, i18n.ErrInternalServerError)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password"     binding:"required,min=8"`
}

func (h *authHandler) changePassword(c *gin.Context) {
	agent := currentAgent(c)
	if agent == nil || agent.PasswordHash == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(c, i18n.ErrUnauthorized)})
		return
	}

	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(c, i18n.ErrBadRequest)})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*agent.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(c, i18n.ErrUnauthorized)})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, i18n.ErrInternalServerError)})
		return
	}
	if err := h.agentRepo.SetPasswordHash(c.Request.Context(), agent.ID, string(hash)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, i18n.ErrInternalServerError)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
