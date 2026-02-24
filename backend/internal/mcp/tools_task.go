package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
	"github.com/linkclaw/backend/internal/service"
)

func (h *Handler) toolListTasks(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		Scope    string `json:"scope"`
		Status   string `json:"status"`
		Priority string `json:"priority"`
	}
	json.Unmarshal(args, &p) //nolint:errcheck

	q := repository.TaskQuery{
		CompanyID: sess.Agent.CompanyID,
		Status:    domain.TaskStatus(p.Status),
		Priority:  domain.TaskPriority(p.Priority),
	}
	if p.Scope != "all" {
		q.AssigneeID = sess.Agent.ID
	}

	tasks, total, err := h.taskSvc.List(ctx, q)
	if err != nil {
		return ErrorResult("查询任务失败: " + err.Error())
	}

	if len(tasks) == 0 {
		return TextResult("暂无任务")
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("共 %d 个任务：\n", total))
	for _, t := range tasks {
		assignee := "未分配"
		if t.AssigneeID != nil {
			assignee = *t.AssigneeID
		}
		lines = append(lines, fmt.Sprintf(
			"[%s] %s\n  ID: %s | 优先级: %s | 负责人: %s",
			t.Status, t.Title, t.ID, t.Priority, assignee,
		))
	}
	return TextResult(strings.Join(lines, "\n"))
}

func (h *Handler) toolGetTask(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		TaskID string `json:"task_id"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.TaskID == "" {
		return ErrorResult("参数错误：需要 task_id")
	}
	t, err := h.taskSvc.GetByID(ctx, p.TaskID)
	if err != nil || t == nil {
		return ErrorResult("任务不存在")
	}

	result := fmt.Sprintf(
		"任务：%s\nID：%s\n状态：%s\n优先级：%s\n描述：%s",
		t.Title, t.ID, t.Status, t.Priority, t.Description,
	)
	if len(t.Subtasks) > 0 {
		result += fmt.Sprintf("\n子任务：%d 个", len(t.Subtasks))
		for _, sub := range t.Subtasks {
			result += fmt.Sprintf("\n  [%s] %s（%s）", sub.Status, sub.Title, sub.ID)
		}
	}
	return TextResult(result)
}

func (h *Handler) toolAcceptTask(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		TaskID string `json:"task_id"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.TaskID == "" {
		return ErrorResult("参数错误：需要 task_id")
	}
	t, err := h.taskSvc.Accept(ctx, p.TaskID, sess.Agent.ID)
	if err != nil {
		return ErrorResult(err.Error())
	}
	return TextResult(fmt.Sprintf("已接受任务「%s」，状态变更为 in_progress。请开始工作！", t.Title))
}

func (h *Handler) toolSubmitTaskResult(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		TaskID string `json:"task_id"`
		Result string `json:"result"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.TaskID == "" || p.Result == "" {
		return ErrorResult("参数错误：需要 task_id 和 result")
	}
	t, err := h.taskSvc.Submit(ctx, p.TaskID, sess.Agent.ID, p.Result)
	if err != nil {
		return ErrorResult(err.Error())
	}
	return TextResult(fmt.Sprintf("任务「%s」已完成！结果已记录。", t.Title))
}

func (h *Handler) toolFailTask(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		TaskID string `json:"task_id"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.TaskID == "" || p.Reason == "" {
		return ErrorResult("参数错误：需要 task_id 和 reason")
	}
	t, err := h.taskSvc.Fail(ctx, p.TaskID, sess.Agent.ID, p.Reason)
	if err != nil {
		return ErrorResult(err.Error())
	}
	return TextResult(fmt.Sprintf("任务「%s」已标记为失败。失败原因已记录。", t.Title))
}

func (h *Handler) toolCreateTask(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		AssigneeID  string `json:"assignee_id"`
		Department  string `json:"department"`
		Priority    string `json:"priority"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.Title == "" {
		return ErrorResult("参数错误：需要 title")
	}

	agentID := sess.Agent.ID
	in := service.CreateTaskInput{
		CompanyID:   sess.Agent.CompanyID,
		Title:       p.Title,
		Description: p.Description,
		Priority:    domain.TaskPriority(p.Priority),
		CreatedBy:   &agentID,
	}

	if p.AssigneeID != "" {
		in.AssigneeID = &p.AssigneeID
	} else if p.Department != "" {
		dirPos, ok := domain.DepartmentDirectors[p.Department]
		if !ok {
			return ErrorResult("未知部门: " + p.Department + "（可选：人力资源、产品、工程、商务、市场、财务）")
		}
		colleagues, _ := h.agentSvc.ListByCompany(ctx, sess.Agent.CompanyID)
		var found bool
		for _, c := range colleagues {
			if c.Position == dirPos {
				in.AssigneeID = &c.ID
				found = true
				break
			}
		}
		if !found {
			return ErrorResult(fmt.Sprintf("部门「%s」暂无总监，请直接指定 assignee_id", p.Department))
		}
	}

	t, err := h.taskSvc.Create(ctx, in)
	if err != nil {
		return ErrorResult("创建任务失败: " + err.Error())
	}

	assigneeInfo := "未分配"
	if t.AssigneeID != nil {
		assigneeInfo = *t.AssigneeID
	}
	return TextResult(fmt.Sprintf("任务「%s」已创建\nID: %s\n负责人: %s\n优先级: %s", t.Title, t.ID, assigneeInfo, t.Priority))
}

func (h *Handler) toolCreateSubtask(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		ParentTaskID string `json:"parent_task_id"`
		Title        string `json:"title"`
		Description  string `json:"description"`
		AssigneeID   string `json:"assignee_id"`
		Priority     string `json:"priority"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.ParentTaskID == "" || p.Title == "" {
		return ErrorResult("参数错误：需要 parent_task_id 和 title")
	}

	agentID := sess.Agent.ID
	in := service.CreateTaskInput{
		CompanyID:   sess.Agent.CompanyID,
		ParentID:    &p.ParentTaskID,
		Title:       p.Title,
		Description: p.Description,
		Priority:    domain.TaskPriority(p.Priority),
		CreatedBy:   &agentID,
	}
	if p.AssigneeID != "" {
		in.AssigneeID = &p.AssigneeID
	}

	t, err := h.taskSvc.Create(ctx, in)
	if err != nil {
		return ErrorResult("创建子任务失败: " + err.Error())
	}
	return TextResult(fmt.Sprintf("子任务「%s」已创建，ID：%s", t.Title, t.ID))
}
