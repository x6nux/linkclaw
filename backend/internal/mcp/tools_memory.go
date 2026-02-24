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

func (h *Handler) toolRemember(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		Content    string `json:"content"`
		Category   string `json:"category"`
		Tags       string `json:"tags"`
		Importance *int   `json:"importance"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.Content == "" {
		return ErrorResult("参数错误：需要 content")
	}

	importance := domain.ImportanceNormal
	if p.Importance != nil {
		importance = domain.MemoryImportance(*p.Importance)
	}

	var tags []string
	if p.Tags != "" {
		for _, t := range strings.Split(p.Tags, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tags = append(tags, t)
			}
		}
	}
	if tags == nil {
		tags = []string{}
	}

	m, err := h.memorySvc.Create(ctx, service.CreateMemoryInput{
		CompanyID:  sess.Agent.CompanyID,
		AgentID:    sess.Agent.ID,
		Content:    p.Content,
		Category:   p.Category,
		Tags:       tags,
		Importance: importance,
		Source:     domain.SourceConversation,
	})
	if err != nil {
		return ErrorResult("存储记忆失败: " + err.Error())
	}
	return TextResult(fmt.Sprintf("已记住 (id=%s, category=%s)", m.ID, m.Category))
}

func (h *Handler) toolRecall(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.Query == "" {
		return ErrorResult("参数错误：需要 query")
	}
	if p.Limit <= 0 {
		p.Limit = 5
	}

	mems, err := h.memorySvc.SemanticSearch(ctx, sess.Agent.CompanyID, sess.Agent.ID, p.Query, p.Limit)
	if err != nil {
		return ErrorResult("检索记忆失败: " + err.Error())
	}

	if len(mems) == 0 {
		return TextResult("没有找到相关记忆。")
	}

	importanceLabels := map[domain.MemoryImportance]string{
		0: "核心", 1: "重要", 2: "普通", 3: "琐碎", 4: "临时",
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("找到 %d 条相关记忆：\n\n", len(mems)))
	for i, m := range mems {
		label := importanceLabels[m.Importance]
		sb.WriteString(fmt.Sprintf("**%d. [%s] %s**\n", i+1, label, m.Category))
		sb.WriteString(m.Content)
		sb.WriteString("\n")
		if len(m.Tags) > 0 {
			sb.WriteString(fmt.Sprintf("标签: %s\n", strings.Join(m.Tags, ", ")))
		}
		sb.WriteString(fmt.Sprintf("_id: %s | 创建: %s_\n\n", m.ID, m.CreatedAt.Format("2006-01-02 15:04")))
	}
	return TextResult(sb.String())
}

func (h *Handler) toolForget(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		MemoryID string `json:"memory_id"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.MemoryID == "" {
		return ErrorResult("参数错误：需要 memory_id")
	}

	m, err := h.memorySvc.GetByID(ctx, p.MemoryID)
	if err != nil || m == nil {
		return ErrorResult("记忆不存在")
	}
	if m.AgentID != sess.Agent.ID {
		return ErrorResult("只能删除自己的记忆")
	}

	if err := h.memorySvc.Delete(ctx, p.MemoryID); err != nil {
		return ErrorResult("删除失败: " + err.Error())
	}
	return TextResult(fmt.Sprintf("已删除记忆 %s", p.MemoryID))
}

func (h *Handler) toolListMemories(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		Category string `json:"category"`
		Limit    int    `json:"limit"`
	}
	_ = json.Unmarshal(args, &p)
	if p.Limit <= 0 {
		p.Limit = 20
	}

	mems, total, err := h.memorySvc.List(ctx, repository.MemoryQuery{
		CompanyID: sess.Agent.CompanyID,
		AgentID:   sess.Agent.ID,
		Category:  p.Category,
		Limit:     p.Limit,
		OrderBy:   "created_at",
	})
	if err != nil {
		return ErrorResult("列出记忆失败: " + err.Error())
	}

	if len(mems) == 0 {
		return TextResult("暂无记忆。")
	}

	importanceLabels := map[domain.MemoryImportance]string{
		0: "核心", 1: "重要", 2: "普通", 3: "琐碎", 4: "临时",
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("共 %d 条记忆（显示 %d 条）：\n\n", total, len(mems)))
	for i, m := range mems {
		label := importanceLabels[m.Importance]
		sb.WriteString(fmt.Sprintf("%d. [%s][%s] %s",
			i+1, label, m.Category, truncate(m.Content, 80)))
		sb.WriteString(fmt.Sprintf(" _(id: %s)_\n", m.ID))
	}
	return TextResult(sb.String())
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}
