package mcp

import (
	"context"
	"encoding/json"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
)

func (h *Handler) toolGetOrgChart(ctx context.Context, sess *Session, _ json.RawMessage) ToolCallResult {
	chart, err := h.orgSvc.BuildOrgChart(ctx, sess.Agent.CompanyID)
	if err != nil {
		return ErrorResult("获取组织架构失败: " + err.Error())
	}
	return okResult(chart)
}

func (h *Handler) toolListMyApprovals(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		Status string `json:"status"`
		Limit  int    `json:"limit"`
	}
	if len(args) > 0 {
		if err := json.Unmarshal(args, &p); err != nil {
			return ErrorResult("参数错误")
		}
	}
	if p.Limit <= 0 {
		p.Limit = 20
	}
	if p.Limit > 100 {
		p.Limit = 100
	}

	items, total, err := h.orgSvc.ListApprovals(ctx, repository.ApprovalQuery{
		CompanyID:   sess.Agent.CompanyID,
		RequesterID: sess.Agent.ID,
		Status:      domain.ApprovalStatus(p.Status),
		Limit:       p.Limit,
	})
	if err != nil {
		return ErrorResult("查询审批请求失败: " + err.Error())
	}
	return okResult(map[string]any{"data": items, "total": total})
}

func (h *Handler) toolSubmitApproval(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		RequestType string `json:"request_type"`
		Reason      string `json:"reason"`
		Payload     string `json:"payload"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.RequestType == "" || p.Reason == "" {
		return ErrorResult("参数错误：需要 request_type 和 reason")
	}

	reqType := domain.ApprovalRequestType(p.RequestType)
	switch reqType {
	case domain.ApprovalHire, domain.ApprovalFire, domain.ApprovalBudgetOverride, domain.ApprovalTaskEscalation, domain.ApprovalCustom:
	default:
		return ErrorResult("参数错误：request_type 无效")
	}

	payload := json.RawMessage("{}")
	if p.Payload != "" {
		if !json.Valid([]byte(p.Payload)) {
			return ErrorResult("参数错误：payload 必须是合法 JSON 字符串")
		}
		payload = json.RawMessage(p.Payload)
	}

	req, err := h.orgSvc.CreateApproval(
		ctx,
		sess.Agent.CompanyID,
		sess.Agent.ID,
		reqType,
		payload,
		p.Reason,
	)
	if err != nil {
		return ErrorResult("提交审批请求失败: " + err.Error())
	}
	return okResult(req)
}

func (h *Handler) toolViewDepartment(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		DepartmentID string `json:"department_id"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.DepartmentID == "" {
		return ErrorResult("参数错误：需要 department_id")
	}

	dept, err := h.orgSvc.GetDepartment(ctx, p.DepartmentID)
	if err != nil {
		return ErrorResult("获取部门信息失败: " + err.Error())
	}
	if dept == nil || dept.CompanyID != sess.Agent.CompanyID {
		return ErrorResult("部门不存在")
	}
	return okResult(dept)
}
