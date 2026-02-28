package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/linkclaw/backend/internal/repository"
)

type auditHandler struct {
	auditRepo repository.AuditRepo
}

func (h *auditHandler) listAuditLogs(c *gin.Context) {
	q := repository.AuditQuery{
		CompanyID:    currentCompanyID(c),
		ResourceType: c.Query("resource_type"),
		Action:       c.Query("action"),
		AgentID:      c.Query("agent_id"),
		Limit:        50,
		Offset:       0,
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			q.Limit = l
		}
	}
	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			q.Offset = o
		}
	}

	logs, total, err := h.auditRepo.ListAuditLogs(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": logs, "total": total})
}
