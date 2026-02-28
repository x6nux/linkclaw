package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/service"
)

type webhookHandler struct {
	webhookSvc *service.WebhookService
}

type createWebhookRequest struct {
	Name           string                    `json:"name" binding:"required"`
	URL            string                    `json:"url" binding:"required"`
	SigningKeyID   *string                   `json:"signing_key_id"`
	Events         []domain.WebhookEventType `json:"events" binding:"required"`
	SecretHeader   string                    `json:"secret_header"`
	IsActive       *bool                     `json:"is_active"`
	TimeoutSeconds int                       `json:"timeout_seconds"`
	RetryPolicy    *domain.RetryPolicy       `json:"retry_policy"`
}

type updateWebhookRequest struct {
	Name           string                    `json:"name"`
	URL            string                    `json:"url"`
	SigningKeyID   *string                   `json:"signing_key_id"`
	Events         []domain.WebhookEventType `json:"events"`
	SecretHeader   string                    `json:"secret_header"`
	IsActive       *bool                     `json:"is_active"`
	TimeoutSeconds int                       `json:"timeout_seconds"`
	RetryPolicy    *domain.RetryPolicy       `json:"retry_policy"`
}

func (h *webhookHandler) createWebhook(c *gin.Context) {
	var req createWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	out, err := h.webhookSvc.CreateWebhook(c.Request.Context(), service.CreateWebhookInput{
		CompanyID:      currentCompanyID(c),
		Name:           req.Name,
		URL:            req.URL,
		SigningKeyID:   req.SigningKeyID,
		Events:         req.Events,
		SecretHeader:   req.SecretHeader,
		IsActive:       req.IsActive,
		TimeoutSeconds: req.TimeoutSeconds,
		RetryPolicy:    req.RetryPolicy,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *webhookHandler) listWebhooks(c *gin.Context) {
	list, err := h.webhookSvc.ListWebhooks(c.Request.Context(), currentCompanyID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *webhookHandler) getWebhook(c *gin.Context) {
	out, err := h.webhookSvc.GetWebhook(c.Request.Context(), currentCompanyID(c), c.Param("id"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	if out == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *webhookHandler) updateWebhook(c *gin.Context) {
	var req updateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	out, err := h.webhookSvc.UpdateWebhook(c.Request.Context(), c.Param("id"), service.UpdateWebhookInput{
		CompanyID:      currentCompanyID(c),
		Name:           req.Name,
		URL:            req.URL,
		SigningKeyID:   req.SigningKeyID,
		Events:         req.Events,
		SecretHeader:   req.SecretHeader,
		IsActive:       req.IsActive,
		TimeoutSeconds: req.TimeoutSeconds,
		RetryPolicy:    req.RetryPolicy,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *webhookHandler) deleteWebhook(c *gin.Context) {
	if err := h.webhookSvc.DeleteWebhook(c.Request.Context(), currentCompanyID(c), c.Param("id")); err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type createSigningKeyRequest struct {
	Name      string                       `json:"name" binding:"required"`
	KeyType   domain.WebhookSigningKeyType `json:"key_type"`
	PublicKey string                       `json:"public_key"`
	SecretKey string                       `json:"secret_key"`
	IsActive  *bool                        `json:"is_active"`
}

func (h *webhookHandler) createSigningKey(c *gin.Context) {
	var req createSigningKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	out, err := h.webhookSvc.CreateSigningKey(c.Request.Context(), service.CreateSigningKeyInput{
		CompanyID: currentCompanyID(c),
		Name:      req.Name,
		KeyType:   req.KeyType,
		PublicKey: req.PublicKey,
		SecretKey: req.SecretKey,
		IsActive:  req.IsActive,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *webhookHandler) listSigningKeys(c *gin.Context) {
	list, err := h.webhookSvc.ListSigningKeys(c.Request.Context(), currentCompanyID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *webhookHandler) deleteSigningKey(c *gin.Context) {
	if err := h.webhookSvc.DeleteSigningKey(c.Request.Context(), currentCompanyID(c), c.Param("id")); err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *webhookHandler) getDeliveryStatus(c *gin.Context) {
	out, err := h.webhookSvc.GetDelivery(c.Request.Context(), currentCompanyID(c), c.Param("id"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	if out == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *webhookHandler) writeError(c *gin.Context, err error) {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "not found"):
		c.JSON(http.StatusNotFound, gin.H{"error": msg})
	case strings.Contains(msg, "forbidden"):
		c.JSON(http.StatusForbidden, gin.H{"error": msg})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
	}
}
