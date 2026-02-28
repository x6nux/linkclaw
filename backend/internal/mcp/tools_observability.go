package mcp

import (
	"context"
	"encoding/json"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
)

func okResult(payload any) ToolCallResult {
	data, err := json.Marshal(payload)
	if err != nil {
		return ErrorResult("结果序列化失败: " + err.Error())
	}
	return TextResult(string(data))
}

func (h *Handler) toolGetMyTraceHistory(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		Limit  int    `json:"limit"`
		Status string `json:"status"`
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

	fetchLimit := p.Limit * 5
	if fetchLimit < 50 {
		fetchLimit = 50
	}
	if fetchLimit > 500 {
		fetchLimit = 500
	}

	traces, _, err := h.obsRepo.ListTraceRuns(ctx, repository.TraceRunQuery{
		CompanyID: sess.Agent.CompanyID,
		Status:    domain.TraceStatus(p.Status),
		Limit:     fetchLimit,
	})
	if err != nil {
		return ErrorResult("查询 Trace 历史失败: " + err.Error())
	}

	result := make([]*domain.TraceRun, 0, p.Limit)
	for _, tr := range traces {
		if tr.RootAgentID == nil || *tr.RootAgentID != sess.Agent.ID {
			continue
		}
		result = append(result, tr)
		if len(result) >= p.Limit {
			break
		}
	}

	return okResult(map[string]any{"data": result, "total": len(result)})
}

func (h *Handler) toolGetCostStatus(ctx context.Context, sess *Session, _ json.RawMessage) ToolCallResult {
	overview, err := h.obsRepo.GetTraceOverview(ctx, sess.Agent.CompanyID)
	if err != nil {
		return ErrorResult("获取成本概览失败: " + err.Error())
	}
	alerts, err := h.obsRepo.ListBudgetAlerts(ctx, repository.BudgetAlertQuery{
		CompanyID: sess.Agent.CompanyID,
		Status:    domain.AlertStatusOpen,
		Limit:     50,
	})
	if err != nil {
		return ErrorResult("获取预算告警失败: " + err.Error())
	}

	return okResult(map[string]any{
		"overview":         overview,
		"activeAlerts":     alerts,
		"activeAlertCount": len(alerts),
	})
}

func (h *Handler) toolListObsAlerts(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		Status string `json:"status"`
		Level  string `json:"level"`
		Limit  int    `json:"limit"`
	}
	if len(args) > 0 {
		if err := json.Unmarshal(args, &p); err != nil {
			return ErrorResult("参数错误")
		}
	}
	if p.Limit <= 0 {
		p.Limit = 50
	}
	if p.Limit > 200 {
		p.Limit = 200
	}

	alerts, err := h.obsRepo.ListBudgetAlerts(ctx, repository.BudgetAlertQuery{
		CompanyID: sess.Agent.CompanyID,
		Status:    domain.BudgetAlertStatus(p.Status),
		Level:     domain.BudgetAlertLevel(p.Level),
		Limit:     p.Limit,
	})
	if err != nil {
		return ErrorResult("查询预算告警失败: " + err.Error())
	}
	return okResult(map[string]any{"data": alerts, "total": len(alerts)})
}

func (h *Handler) toolReplayTrace(ctx context.Context, _ *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		TraceID string `json:"trace_id"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.TraceID == "" {
		return ErrorResult("参数错误：需要 trace_id")
	}

	tree, err := h.obsSvc.GetTraceTree(ctx, p.TraceID)
	if err != nil {
		return ErrorResult("获取 Trace 回放失败: " + err.Error())
	}
	if tree == nil {
		return ErrorResult("trace 不存在")
	}
	return okResult(tree)
}
