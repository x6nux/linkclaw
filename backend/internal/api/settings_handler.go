package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
)

type settingsHandler struct {
	companyRepo repository.CompanyRepo
}

func (h *settingsHandler) get(c *gin.Context) {
	companyID := currentCompanyID(c)
	company, err := h.companyRepo.GetByID(c.Request.Context(), companyID)
	if err != nil || company == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}
	c.JSON(http.StatusOK, domain.CompanySettings{
		PublicDomain:      company.PublicDomain,
		AgentWSUrl:        company.AgentWSUrl,
		MCPPublicURL:      company.MCPPublicURL,
		NanoclawImage:     company.NanoclawImage,
		OpenclawPluginURL: company.OpenclawPluginURL,
	})
}

func (h *settingsHandler) update(c *gin.Context) {
	var req domain.CompanySettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	companyID := currentCompanyID(c)
	if err := h.companyRepo.UpdateSettings(c.Request.Context(), companyID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update settings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
