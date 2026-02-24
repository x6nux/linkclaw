package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/service"
)

type agentHandler struct {
	agentSvc *service.AgentService
}

func (h *agentHandler) list(c *gin.Context) {
	companyID := currentCompanyID(c)
	agents, err := h.agentSvc.ListByCompany(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": agents, "total": len(agents)})
}

func (h *agentHandler) get(c *gin.Context) {
	agent, err := h.agentSvc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil || agent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, agent)
}

type createAgentRequest struct {
	Name     string `json:"name"`
	Position string `json:"position" binding:"required"`
	Persona  string `json:"persona"`
	Model    string `json:"model"`    // LLM 模型名
	Password string `json:"password"` // 仅 chairman 创建人类账户时使用
	IsHuman  bool   `json:"is_human"`
}

func (h *agentHandler) create(c *gin.Context) {
	var req createAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	companyID := currentCompanyID(c)
	out, err := h.agentSvc.Create(c.Request.Context(), service.CreateAgentInput{
		CompanyID: companyID,
		Name:      req.Name,
		Position:  domain.Position(req.Position),
		Persona:   req.Persona,
		Model:     req.Model,
		IsHuman:   req.IsHuman,
		Password:  req.Password,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := gin.H{"agent": out.Agent}
	if out.APIKey != "" {
		resp["api_key"] = out.APIKey // 仅创建时返回
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *agentHandler) delete(c *gin.Context) {
	if err := h.agentSvc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type updateAgentRequest struct {
	Name    *string `json:"name"`
	Model   *string `json:"model"`
	Persona *string `json:"persona"`
}

func (h *agentHandler) update(c *gin.Context) {
	id := c.Param("id")
	var req updateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := c.Request.Context()
	if req.Name != nil {
		if err := h.agentSvc.UpdateName(ctx, id, *req.Name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if req.Model != nil {
		if err := h.agentSvc.UpdateModel(ctx, id, *req.Model); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if req.Persona != nil {
		if err := h.agentSvc.UpdatePersona(ctx, id, *req.Persona); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	agent, err := h.agentSvc.GetByID(ctx, id)
	if err != nil || agent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, agent)
}
