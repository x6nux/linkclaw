package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/linkclaw/backend/internal/service"
)

func (h *Handler) toolSearchContext(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var params struct {
		Query        string   `json:"query"`
		DirectoryIDs []string `json:"directory_ids"`
		MaxResults   int      `json:"max_results,omitempty"`
		MinRelevance float64  `json:"min_relevance,omitempty"`
		TimeoutMs    int      `json:"timeout_ms,omitempty"`
		UseIndex     *bool    `json:"use_index,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return ErrorResult("invalid parameters")
	}

	if params.Query == "" {
		return ErrorResult("query is required")
	}

	out, err := h.contextSvc.Search(ctx, service.SearchInput{
		CompanyID:    sess.Agent.CompanyID,
		AgentID:      sess.Agent.ID,
		Query:        params.Query,
		DirectoryIDs: params.DirectoryIDs,
		MaxResults:   params.MaxResults,
		MinRelevance: params.MinRelevance,
		TimeoutMs:    params.TimeoutMs,
		UseIndex:     params.UseIndex,
	})
	if err != nil {
		return ErrorResult("search failed: " + err.Error())
	}

	// 检查是否有错误
	if out.Error != nil {
		return ErrorResult(fmt.Sprintf("[%s] %s", out.Error.Code, out.Error.Message))
	}

	// 返回格式化文本结果（Agent 可读）
	var content string
	if len(out.Results) == 0 {
		content = "未找到相关文件"
	} else {
		content = fmt.Sprintf("找到 %d 个相关文件:\n\n", len(out.Results))
		for i, r := range out.Results {
			content += fmt.Sprintf("%d. %s (相关性：%.2f)\n", i+1, r.FilePath, r.Relevance)
			content += fmt.Sprintf("   摘要：%s\n", r.Summary)
			if r.Reason != "" {
				content += fmt.Sprintf("   原因：%s\n", r.Reason)
			}
			content += "\n"
		}
	}

	return TextResult(content)
}
