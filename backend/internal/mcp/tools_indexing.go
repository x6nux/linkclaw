package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func (h *Handler) toolSearchCode(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	if h.indexingSvc == nil {
		return ErrorResult("索引服务未配置")
	}

	var p struct {
		Query string `json:"query"`
		Limit any    `json:"limit"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.Query == "" {
		return ErrorResult("参数错误：需要 query")
	}

	limit := 10
	switch v := p.Limit.(type) {
	case string:
		if l, err := strconv.Atoi(v); err == nil && l > 0 {
			limit = l
		}
	case float64:
		if v > 0 {
			limit = int(v)
		}
	}

	results, err := h.indexingSvc.SearchCode(ctx, sess.Agent.CompanyID, p.Query, limit)
	if err != nil {
		return ErrorResult("搜索失败: " + err.Error())
	}
	if len(results) == 0 {
		return TextResult("未找到相关代码")
	}

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("找到 %d 条相关代码：\n\n", len(results)))
	for i, r := range results {
		buf.WriteString(fmt.Sprintf("%d. [%v] %v:%v-%v (相似度: %.2f)\n",
			i+1,
			r.Payload["language"],
			r.Payload["file_path"],
			r.Payload["start_line"],
			r.Payload["end_line"],
			r.Score,
		))
		if content, ok := r.Payload["content"]; ok {
			buf.WriteString(fmt.Sprintf("%v", content))
			buf.WriteString("\n\n")
		}
	}

	return TextResult(buf.String())
}

func (h *Handler) toolIndexRepository(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	if h.indexingSvc == nil {
		return ErrorResult("索引服务未配置")
	}

	var p struct {
		RepoURL string `json:"repository_url"`
		Branch  string `json:"branch"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.RepoURL == "" {
		return ErrorResult("参数错误：需要 repository_url")
	}
	if p.Branch == "" {
		p.Branch = "main"
	}

	task, err := h.indexingSvc.IndexRepository(ctx, sess.Agent.CompanyID, p.RepoURL, p.Branch)
	if err != nil {
		return ErrorResult("创建索引任务失败: " + err.Error())
	}
	return TextResult(fmt.Sprintf("索引任务已创建，ID: %s，状态: %s", task.ID, task.Status))
}

func (h *Handler) toolGetIndexStatus(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	if h.indexingSvc == nil {
		return ErrorResult("索引服务未配置")
	}

	var p struct {
		TaskID string `json:"task_id"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.TaskID == "" {
		return ErrorResult("参数错误：需要 task_id")
	}

	task, err := h.indexingSvc.GetIndexStatus(ctx, p.TaskID)
	if err != nil {
		return ErrorResult("获取索引状态失败: " + err.Error())
	}
	if task == nil {
		return ErrorResult("索引任务不存在")
	}
	// 权限检查：确保任务属于当前公司
	if task.CompanyID != sess.Agent.CompanyID {
		return ErrorResult("索引任务不存在")
	}

	return TextResult(fmt.Sprintf("索引任务 %s: %s (%d/%d 文件)",
		task.ID, task.Status, task.IndexedFiles, task.TotalFiles,
	))
}
