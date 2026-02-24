package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/linkclaw/backend/internal/service"
)

func (h *Handler) toolSearchKnowledge(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		Query string `json:"query"`
		Limit string `json:"limit"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.Query == "" {
		return ErrorResult("参数错误：需要 query")
	}
	limit := 10
	if p.Limit != "" {
		if l, err := strconv.Atoi(p.Limit); err == nil && l > 0 {
			limit = l
		}
	}

	docs, err := h.knowledgeSvc.Search(ctx, sess.Agent.CompanyID, p.Query, limit)
	if err != nil {
		return ErrorResult("搜索失败: " + err.Error())
	}
	if len(docs) == 0 {
		return TextResult("未找到相关文档")
	}

	var lines []string
	for _, d := range docs {
		tags := ""
		if len(d.Tags) > 0 {
			tags = " [" + strings.Join(d.Tags, ", ") + "]"
		}
		preview := d.Content
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		lines = append(lines, fmt.Sprintf("📄 %s%s\n   ID：%s\n   %s", d.Title, tags, d.ID, preview))
	}
	return TextResult(strings.Join(lines, "\n\n"))
}

func (h *Handler) toolGetDocument(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		DocID string `json:"doc_id"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.DocID == "" {
		return ErrorResult("参数错误：需要 doc_id")
	}

	doc, err := h.knowledgeSvc.GetByID(ctx, p.DocID)
	if err != nil || doc == nil {
		return ErrorResult("文档不存在")
	}

	result := fmt.Sprintf("# %s\n\nID：%s\n更新时间：%s\n\n---\n\n%s",
		doc.Title, doc.ID, doc.UpdatedAt.Format("2006-01-02 15:04"), doc.Content)
	return TextResult(result)
}

func (h *Handler) toolWriteDocument(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		DocID   string `json:"doc_id"`
		Title   string `json:"title"`
		Content string `json:"content"`
		Tags    string `json:"tags"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.Title == "" || p.Content == "" {
		return ErrorResult("参数错误：需要 title 和 content")
	}

	doc, err := h.knowledgeSvc.Write(ctx, service.WriteDocInput{
		DocID:     p.DocID,
		CompanyID: sess.Agent.CompanyID,
		AuthorID:  sess.Agent.ID,
		Title:     p.Title,
		Content:   p.Content,
		Tags:      p.Tags,
	})
	if err != nil {
		return ErrorResult("保存文档失败: " + err.Error())
	}

	action := "已创建"
	if p.DocID != "" {
		action = "已更新"
	}
	return TextResult(fmt.Sprintf("文档「%s」%s，ID：%s", doc.Title, action, doc.ID))
}
