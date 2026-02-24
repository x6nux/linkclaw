package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/service"
)

type deploymentHandler struct {
	deployService *service.DeploymentService
	agentService  *service.AgentService
}

type deployRequest struct {
	DeployType  string `json:"deployType"  binding:"required"` // local_docker | ssh_docker | ssh_native
	AgentImage  string `json:"agentImage"  binding:"required"` // nanoclaw | openclaw
	APIKey      string `json:"apiKey"`                         // 原始 API Key，local_docker 需要
	SSHHost     string `json:"sshHost"`
	SSHPort     int    `json:"sshPort"`
	SSHUser     string `json:"sshUser"`
	SSHPassword string `json:"sshPassword"`
	SSHKey      string `json:"sshKey"`
}

func (h *deploymentHandler) deploy(c *gin.Context) {
	agentID := c.Param("id")

	agent, err := h.agentService.GetByID(c.Request.Context(), agentID)
	if err != nil || agent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	var req deployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dt := domain.DeployType(req.DeployType)
	if dt == domain.DeployTypeSSHDocker || dt == domain.DeployTypeSSHNative {
		if req.SSHHost == "" || req.SSHUser == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "sshHost and sshUser required for SSH deployments"})
			return
		}
	}

	if agent.Model == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent has no model configured, update agent first"})
		return
	}

	deployment, err := h.deployService.Deploy(c.Request.Context(), service.DeployInput{
		AgentID:     agentID,
		DeployType:  dt,
		AgentImage:  domain.AgentImage(req.AgentImage),
		APIKey:      req.APIKey,
		Model:       agent.Model,
		SSHHost:     req.SSHHost,
		SSHPort:     req.SSHPort,
		SSHUser:     req.SSHUser,
		SSHPassword: req.SSHPassword,
		SSHKey:      req.SSHKey,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deployment)
}

func (h *deploymentHandler) getDeployment(c *gin.Context) {
	agentID := c.Param("id")
	d, err := h.deployService.GetByAgentID(c.Request.Context(), agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if d == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no deployment found"})
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *deploymentHandler) stopDeployment(c *gin.Context) {
	agentID := c.Param("id")
	if err := h.deployService.Stop(c.Request.Context(), agentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *deploymentHandler) rebuild(c *gin.Context) {
	agentID := c.Param("id")
	d, newKey, err := h.deployService.Rebuild(c.Request.Context(), agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deployment": d, "newApiKey": newKey})
}
