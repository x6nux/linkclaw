package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/service"
)

type personaHandler struct {
	personaSvc *service.PersonaOptimizerService
}

type createABTestRequest struct {
	Name           string `json:"name" binding:"required"`
	Description    string `json:"description"`
	ControlAgentID string `json:"controlAgentId" binding:"required"`
	ControlPersona string `json:"controlPersona"`
	VariantAgentID string `json:"variantAgentId" binding:"required"`
	VariantPersona string `json:"variantPersona"`
}

func (h *personaHandler) listSuggestions(c *gin.Context) {
	agentID := c.Query("agentId")
	if agentID == "" {
		if agent := currentAgent(c); agent != nil {
			agentID = agent.ID
		}
	}
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agentId is required"})
		return
	}

	suggestions, err := h.personaSvc.ListSuggestions(
		c.Request.Context(),
		currentCompanyID(c),
		agentID,
		domain.SuggestionStatus(c.Query("status")),
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": suggestions, "total": len(suggestions)})
}

func (h *personaHandler) applySuggestion(c *gin.Context) {
	agentID := c.Query("agentId")
	if agentID == "" {
		if agent := currentAgent(c); agent != nil {
			agentID = agent.ID
		}
	}
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agentId is required"})
		return
	}

	if err := h.personaSvc.ApplySuggestion(c.Request.Context(), c.Param("id"), agentID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *personaHandler) getHistory(c *gin.Context) {
	history, err := h.personaSvc.GetHistory(c.Request.Context(), currentCompanyID(c), c.Param("agentId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": history, "total": len(history)})
}

func (h *personaHandler) createABTest(c *gin.Context) {
	var req createABTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	abTest, err := h.personaSvc.CreateABTest(c.Request.Context(), service.CreateABTestInput{
		CompanyID:      currentCompanyID(c),
		Name:           req.Name,
		Description:    req.Description,
		ControlAgentID: req.ControlAgentID,
		ControlPersona: req.ControlPersona,
		VariantAgentID: req.VariantAgentID,
		VariantPersona: req.VariantPersona,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": abTest})
}

func (h *personaHandler) listABTests(c *gin.Context) {
	tests, err := h.personaSvc.ListABTests(c.Request.Context(), currentCompanyID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tests, "total": len(tests)})
}
