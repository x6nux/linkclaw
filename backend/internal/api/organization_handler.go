package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
	"github.com/linkclaw/backend/internal/service"
)

type organizationHandler struct {
	orgSvc *service.OrganizationService
}

type departmentPayload struct {
	Name            string  `json:"name" binding:"required"`
	Slug            string  `json:"slug" binding:"required"`
	Description     string  `json:"description"`
	DirectorAgentID *string `json:"director_agent_id"`
	ParentDeptID    *string `json:"parent_dept_id"`
}

type assignAgentPayload struct {
	AgentID string `json:"agent_id" binding:"required"`
}

type setManagerPayload struct {
	ManagerID *string `json:"manager_id"`
}

type createApprovalPayload struct {
	RequestType string          `json:"request_type" binding:"required"`
	Payload     json.RawMessage `json:"payload"`
	Reason      string          `json:"reason" binding:"required"`
}

type approvalDecisionPayload struct {
	DecisionReason string `json:"decision_reason" binding:"required"`
}

func normalizeOptionalString(v *string) *string {
	if v == nil || *v == "" {
		return nil
	}
	return v
}

func (h *organizationHandler) listDepartments(c *gin.Context) {
	departments, err := h.orgSvc.ListDepartments(c.Request.Context(), currentCompanyID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": departments, "total": len(departments)})
}

func (h *organizationHandler) createDepartment(c *gin.Context) {
	var req departmentPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dept, err := h.orgSvc.CreateDepartment(
		c.Request.Context(),
		currentCompanyID(c),
		req.Name,
		req.Slug,
		req.Description,
		normalizeOptionalString(req.DirectorAgentID),
		normalizeOptionalString(req.ParentDeptID),
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": dept})
}

func (h *organizationHandler) updateDepartment(c *gin.Context) {
	var req departmentPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.orgSvc.UpdateDepartment(
		c.Request.Context(),
		c.Param("id"),
		req.Name,
		req.Slug,
		req.Description,
		normalizeOptionalString(req.DirectorAgentID),
		normalizeOptionalString(req.ParentDeptID),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *organizationHandler) deleteDepartment(c *gin.Context) {
	if err := h.orgSvc.DeleteDepartment(c.Request.Context(), c.Param("id"), currentCompanyID(c)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *organizationHandler) assignAgent(c *gin.Context) {
	var req assignAgentPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.orgSvc.AssignAgentToDepartment(c.Request.Context(), req.AgentID, c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *organizationHandler) setManager(c *gin.Context) {
	var req setManagerPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.orgSvc.SetManager(c.Request.Context(), c.Param("id"), normalizeOptionalString(req.ManagerID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *organizationHandler) orgChart(c *gin.Context) {
	chart, err := h.orgSvc.BuildOrgChart(c.Request.Context(), currentCompanyID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": chart})
}

func (h *organizationHandler) listApprovals(c *gin.Context) {
	agent := currentAgent(c)
	q := repository.ApprovalQuery{
		CompanyID:   currentCompanyID(c),
		Status:      domain.ApprovalStatus(c.Query("status")),
		RequestType: domain.ApprovalRequestType(c.Query("request_type")),
		Limit:       parseIntQuery(c, "limit", 20),
		Offset:      parseIntQuery(c, "offset", 0),
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}
	if q.Limit > 200 {
		q.Limit = 200
	}
	if q.Offset < 0 {
		q.Offset = 0
	}
	if agent == nil || agent.RoleType != domain.RoleChairman {
		q.RequesterID = currentAgent(c).ID
	}

	approvals, total, err := h.orgSvc.ListApprovals(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": approvals, "total": total})
}

func (h *organizationHandler) createApproval(c *gin.Context) {
	var req createApprovalPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	approval, err := h.orgSvc.CreateApproval(
		c.Request.Context(),
		currentCompanyID(c),
		currentAgent(c).ID,
		domain.ApprovalRequestType(req.RequestType),
		req.Payload,
		req.Reason,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": approval})
}

func (h *organizationHandler) approveRequest(c *gin.Context) {
	var req approvalDecisionPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.orgSvc.ApproveRequest(c.Request.Context(), c.Param("id"), req.DecisionReason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *organizationHandler) rejectRequest(c *gin.Context) {
	var req approvalDecisionPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.orgSvc.RejectRequest(c.Request.Context(), c.Param("id"), req.DecisionReason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
