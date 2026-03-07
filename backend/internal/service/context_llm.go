package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/linkclaw/backend/internal/llm"
	"github.com/linkclaw/backend/internal/repository"
)

// ContextLLMClient 封装上下文服务的 LLM 调用
type ContextLLMClient struct {
	repo       *repository.ContextRepo
	llmRouter  *llm.Router
	llmRepo    *llm.Repository
	httpClient *http.Client
	encKey     string
	metrics    *ContextMetrics // 可观测性指标
}

// NewContextLLMClient 创建上下文 LLM 客户端
func NewContextLLMClient(ctxRepo *repository.ContextRepo, llmRouter *llm.Router, llmRepo *llm.Repository, encKey string) *ContextLLMClient {
	return &ContextLLMClient{
		repo:       ctxRepo,
		llmRouter:  llmRouter,
		llmRepo:    llmRepo,
		httpClient: &http.Client{Timeout: 120 * time.Second},
		encKey:     encKey,
		metrics:    NewContextMetrics(),
	}
}

// GetMetrics 返回指标收集器（用于外部访问）
func (c *ContextLLMClient) GetMetrics() *ContextMetrics {
	return c.metrics
}

// FileContent 表示待处理文件的内容
type FileContent struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
	Language string `json:"language,omitempty"`
}

// SummarizeResult 文件总结结果
type SummarizeResult struct {
	FilePath    string `json:"file_path"`
	Language    string `json:"language,omitempty"`
	Summary     string `json:"summary"`
	LineCount   int    `json:"line_count"`
	ContentHash string `json:"content_hash"`
}

// SemanticSearchRequest 语义搜索请求
type SemanticSearchRequest struct {
	Query      string       `json:"query"`
	Files      []FileContent `json:"files"`
	DirectoryID string      `json:"directory_id,omitempty"`
}

// SemanticSearchResult 语义搜索结果
type SemanticSearchResult struct {
	FilePath    string  `json:"file_path"`
	Language    string  `json:"language,omitempty"`
	Summary     string  `json:"summary"`
	Relevance   float64 `json:"relevance"`
	Reason      string  `json:"reason"`
	LineCount   int     `json:"line_count"`
	DirectoryID string  `json:"directory_id"`
}

// SummarizeFile 调用 LLM 对单个文件内容进行总结
// 返回总结文本、行数、内容哈希
func (c *ContextLLMClient) SummarizeFile(ctx context.Context, content, language string) (string, int, error) {
	lines := strings.Split(content, "\n")
	lineCount := len(lines)

	// 构建提示词
	systemPrompt := `You are a code analysis expert. Your task is to provide concise and informative summaries of code files.
Focus on:
1. What the file does (main purpose)
2. Key functions/classes and their roles
3. Important dependencies or patterns
4. Keep summary under 200 words`

	userPrompt := fmt.Sprintf(`Please summarize the following %s file:

%s

Provide a concise summary covering:
- Main purpose
- Key components (functions, classes, etc.)
- Notable patterns or dependencies`, language, truncateContent(content, 50000))

	// 调用 LLM
	summary, err := c.callLLM(ctx, systemPrompt, userPrompt, "")
	if err != nil {
		return "", 0, fmt.Errorf("summarize file: %w", err)
	}

	return strings.TrimSpace(summary), lineCount, nil
}

// SemanticSearchFromSummaries 基于文件摘要进行语义搜索（轻量级）
// 与 SemanticSearch 不同，这个方法只分析摘要内容，不读取完整文件
func (c *ContextLLMClient) SemanticSearchFromSummaries(ctx context.Context, query string, summaries []FileContent) ([]*SemanticSearchResult, error) {
	if len(summaries) == 0 {
		return []*SemanticSearchResult{}, nil
	}

	// 构建摘要上下文
	var summariesContext strings.Builder
	for i, s := range summaries {
		summariesContext.WriteString(fmt.Sprintf("=== File %d: %s (%s) ===\n摘要：%s\n", i+1, s.FilePath, s.Language, s.Content))
	}

	systemPrompt := `You are a semantic search expert. Analyze the given query and file summaries to determine relevance.
For each file, provide:
1. A relevance score from 0.0 to 1.0
2. A brief explanation of why it's relevant (or not)
Focus on semantic meaning, not just keyword matching.`

	userPrompt := fmt.Sprintf(`Query: %s

File summaries to analyze:
%s

For EACH file, output a JSON object with:
- file_path: the file path
- relevance: score from 0.0 to 1.0
- reason: brief explanation (1-2 sentences)

Output ONLY valid JSON array, no other text.

Example output format:
[
  {"file_path": "path/to/file.go", "relevance": 0.85, "reason": "Contains the main search logic implementation"},
  {"file_path": "other/file.go", "relevance": 0.3, "reason": "Only mentions search in passing"}
]`, query, summariesContext.String())

	response, err := c.callLLM(ctx, systemPrompt, userPrompt, "")
	if err != nil {
		return nil, fmt.Errorf("semantic search LLM: %w", err)
	}

	// 解析 LLM 返回的 JSON 结果
	var llmResults []struct {
		FilePath  string  `json:"file_path"`
		Relevance float64 `json:"relevance"`
		Reason    string  `json:"reason"`
	}
	if err := json.Unmarshal([]byte(response), &llmResults); err != nil {
		return nil, fmt.Errorf("parse LLM response: %w", err)
	}

	// 构建最终结果
	results := make([]*SemanticSearchResult, 0, len(llmResults))
	summaryMap := make(map[string]FileContent)
	for _, s := range summaries {
		summaryMap[s.FilePath] = s
	}

	for _, llmRes := range llmResults {
		file, ok := summaryMap[llmRes.FilePath]
		if !ok {
			continue
		}

		results = append(results, &SemanticSearchResult{
			FilePath:  llmRes.FilePath,
			Language:  file.Language,
			Summary:   file.Content, // 摘要内容作为返回的 summary
			Relevance: clampRelevance(llmRes.Relevance),
			Reason:    llmRes.Reason,
		})
	}

	// 按相关性分数降序排序
	sortByRelevance(results)

	return results, nil
}

// SemanticSearch 对多个文件内容进行语义搜索
// 返回相关性排序的搜索结果
func (c *ContextLLMClient) SemanticSearch(ctx context.Context, query string, files []FileContent, directoryID string) ([]*SemanticSearchResult, error) {
	if len(files) == 0 {
		return []*SemanticSearchResult{}, nil
	}

	// 构建文件内容摘要（用于 LLM 分析）
	var filesContext strings.Builder
	for i, f := range files {
		filesContext.WriteString(fmt.Sprintf("=== File %d: %s (%s) ===\n", i+1, f.FilePath, f.Language))
		filesContext.WriteString(truncateContent(f.Content, 10000))
		filesContext.WriteString("\n\n")
	}

	systemPrompt := `You are a semantic search expert. Analyze the given query and file contents to determine relevance.
For each file, provide:
1. A relevance score from 0.0 to 1.0
2. A brief explanation of why it's relevant (or not)
Focus on semantic meaning, not just keyword matching.`

	userPrompt := fmt.Sprintf(`Query: %s

Files to analyze:
%s

For EACH file, output a JSON object with:
- file_path: the file path
- relevance: score from 0.0 to 1.0
- reason: brief explanation (1-2 sentences)

Output ONLY valid JSON array, no other text.

Example output format:
[
  {"file_path": "path/to/file.go", "relevance": 0.85, "reason": "Contains the main search logic implementation"},
  {"file_path": "other/file.go", "relevance": 0.3, "reason": "Only mentions search in passing"}
]`, query, filesContext.String())

	response, err := c.callLLM(ctx, systemPrompt, userPrompt, "")
	if err != nil {
		return nil, fmt.Errorf("semantic search LLM: %w", err)
	}

	// 解析 LLM 返回的 JSON 结果
	var llmResults []struct {
		FilePath  string  `json:"file_path"`
		Relevance float64 `json:"relevance"`
		Reason    string  `json:"reason"`
	}
	if err := json.Unmarshal([]byte(response), &llmResults); err != nil {
		return nil, fmt.Errorf("parse LLM response: %w", err)
	}

	// 构建最终结果
	results := make([]*SemanticSearchResult, 0, len(llmResults))
	fileMap := make(map[string]FileContent)
	for _, f := range files {
		fileMap[f.FilePath] = f
	}

	for _, llmRes := range llmResults {
		file, ok := fileMap[llmRes.FilePath]
		if !ok {
			continue
		}

		lines := strings.Split(file.Content, "\n")
		lineCount := len(lines)

		// 获取或创建文件总结
		summary := c.getOrCreateSummary(ctx, file)

		results = append(results, &SemanticSearchResult{
			FilePath:    llmRes.FilePath,
			Language:    file.Language,
			Summary:     summary,
			Relevance:   clampRelevance(llmRes.Relevance),
			Reason:      llmRes.Reason,
			LineCount:   lineCount,
			DirectoryID: directoryID,
		})
	}

	// 按相关性分数降序排序
	sortByRelevance(results)

	return results, nil
}

// callLLM 调用 LLM API 的通用方法
func (c *ContextLLMClient) callLLM(ctx context.Context, systemPrompt, userPrompt, model string) (string, error) {
	start := time.Now()

	// 从仓库获取活跃的 LLM provider
	providers, err := c.llmRepo.ListActiveProviders(ctx, "")
	if err != nil {
		c.metrics.RecordFailure("list_providers")
		c.metrics.RecordLatency(time.Since(start).Milliseconds())
		return "", fmt.Errorf("list providers: %w", err)
	}
	if len(providers) == 0 {
		c.metrics.RecordFailure("no_providers")
		c.metrics.RecordLatency(time.Since(start).Milliseconds())
		return "", fmt.Errorf("no active LLM providers configured")
	}

	// 选择一个 provider（优先选 Anthropic，其次 OpenAI）
	var provider *llm.Provider
	for _, p := range providers {
		if p.Type == llm.ProviderAnthropic {
			provider = p
			break
		}
	}
	if provider == nil {
		for _, p := range providers {
			if p.Type == llm.ProviderOpenAI {
				provider = p
				break
			}
		}
	}
	if provider == nil {
		c.metrics.RecordFailure("unsupported_provider")
		c.metrics.RecordLatency(time.Since(start).Milliseconds())
		return "", fmt.Errorf("no supported LLM provider type")
	}

	// 解密 API Key
	apiKey, err := llm.DecryptAPIKey(provider.APIKeyEnc, c.encKey)
	if err != nil {
		c.metrics.RecordFailure("decrypt_api_key")
		c.metrics.RecordLatency(time.Since(start).Milliseconds())
		return "", fmt.Errorf("decrypt api key: %w", err)
	}

	// 根据 provider 类型调用相应 API
	var result string
	switch provider.Type {
	case llm.ProviderAnthropic:
		result, err = c.callAnthropic(ctx, provider.BaseURL, apiKey, systemPrompt, userPrompt, model)
	case llm.ProviderOpenAI:
		result, err = c.callOpenAI(ctx, provider.BaseURL, apiKey, systemPrompt, userPrompt, model)
	default:
		c.metrics.RecordFailure("unsupported_provider_type")
		c.metrics.RecordLatency(time.Since(start).Milliseconds())
		return "", fmt.Errorf("unsupported provider type: %s", provider.Type)
	}

	// 记录指标
	c.metrics.RecordLatency(time.Since(start).Milliseconds())
	if err != nil {
		c.metrics.RecordFailure("llm_call_failed")
	}

	return result, err
}

// callAnthropic 调用 Anthropic API
func (c *ContextLLMClient) callAnthropic(ctx context.Context, baseURL, apiKey, systemPrompt, userPrompt, model string) (string, error) {
	if model == "" {
		model = "claude-sonnet-4-5"
	}

	reqBody := llm.AnthropicRequest{
		Model:     model,
		MaxTokens: 4096,
		System:    systemPrompt,
		Stream:    false,
		Messages: []llm.AnthropicMessage{
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := strings.TrimRight(baseURL, "/") + "/v1/messages"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if isRetryableError(err, nil) {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			return "", fmt.Errorf("anthropic request: %w", err)
		}

		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("anthropic %d: %s", resp.StatusCode, string(respBody))
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("read response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("anthropic %d: %s", resp.StatusCode, string(respBody))
		}

		// 解析响应
		var result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
			Usage *struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(respBody, &result); err != nil {
			return "", fmt.Errorf("parse response: %w", err)
		}

		if result.Error != nil {
			return "", fmt.Errorf("anthropic error: %s", result.Error.Message)
		}

		if len(result.Content) == 0 {
			return "", fmt.Errorf("empty response from anthropic")
		}

		// 记录 token 使用
		if result.Usage != nil {
			c.metrics.RecordTokens(result.Usage.InputTokens, result.Usage.OutputTokens)
			// 简单成本估算 (假设 $3/1M input, $15/1M output)
			costMicrodollars := int64(result.Usage.InputTokens*3 + result.Usage.OutputTokens*15)
			c.metrics.RecordCost(costMicrodollars)
		}

		return result.Content[0].Text, nil
	}

	return "", lastErr
}

// callOpenAI 调用 OpenAI API
func (c *ContextLLMClient) callOpenAI(ctx context.Context, baseURL, apiKey, systemPrompt, userPrompt, model string) (string, error) {
	if model == "" {
		model = "gpt-4o"
	}

	reqBody := struct {
		Model     string          `json:"model"`
		Messages  []openaiMessage `json:"messages"`
		MaxTokens int             `json:"max_tokens"`
	}{
		Model:     model,
		MaxTokens: 4096,
		Messages: []openaiMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := strings.TrimRight(baseURL, "/") + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if isRetryableError(err, nil) {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			return "", fmt.Errorf("openai request: %w", err)
		}

		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("openai %d: %s", resp.StatusCode, string(respBody))
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("read response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("openai %d: %s", resp.StatusCode, string(respBody))
		}

		// 解析响应
		var result struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
			Usage *struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
			} `json:"usage"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(respBody, &result); err != nil {
			return "", fmt.Errorf("parse response: %w", err)
		}

		if result.Error != nil {
			return "", fmt.Errorf("openai error: %s", result.Error.Message)
		}

		if len(result.Choices) == 0 {
			return "", fmt.Errorf("empty response from openai")
		}

		// 记录 token 使用
		if result.Usage != nil {
			c.metrics.RecordTokens(result.Usage.PromptTokens, result.Usage.CompletionTokens)
			// 简单成本估算 (假设 $2.5/1M input, $10/1M output)
			costMicrodollars := int64(float64(result.Usage.PromptTokens)*2.5 + float64(result.Usage.CompletionTokens)*10)
			c.metrics.RecordCost(costMicrodollars)
	 }

		return result.Choices[0].Message.Content, nil
	}

	return "", lastErr
}

// openaiMessage OpenAI 消息结构
type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// getOrCreateSummary 调用 LLM 对文件进行总结，失败则返回文件路径作为 fallback
func (c *ContextLLMClient) getOrCreateSummary(ctx context.Context, file FileContent) string {
	summary, _, err := c.SummarizeFile(ctx, file.Content, file.Language)
	if err != nil || summary == "" {
		return file.FilePath
	}
	return summary
}

// truncateContent 截断内容到最大长度
func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "\n... [truncated]"
}

// clampRelevance 限制相关性分数在 0-1 之间
func clampRelevance(score float64) float64 {
	if score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
}

// sortByRelevance 按相关性分数降序排序
func sortByRelevance(results []*SemanticSearchResult) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].Relevance > results[j].Relevance
	})
}

// CallWithTools calls LLM with tool support for agentic workflows
// Returns the response text and any tool calls made by the model
func (c *ContextLLMClient) CallWithTools(ctx context.Context, systemPrompt string, messages []AgentMessage, tools []AgentTool) (string, []AgentToolCall, error) {
	providers, err := c.llmRepo.ListActiveProviders(ctx, "")
	if err != nil {
		return "", nil, fmt.Errorf("list providers: %w", err)
	}
	if len(providers) == 0 {
		return "", nil, fmt.Errorf("no active LLM providers configured")
	}

	var provider *llm.Provider
	for _, p := range providers {
		if p.Type == llm.ProviderAnthropic {
			provider = p
			break
		}
	}
	if provider == nil {
		for _, p := range providers {
			if p.Type == llm.ProviderOpenAI {
				provider = p
				break
			}
		}
	}
	if provider == nil {
		return "", nil, fmt.Errorf("no supported LLM provider type")
	}

	apiKey, err := llm.DecryptAPIKey(provider.APIKeyEnc, c.encKey)
	if err != nil {
		return "", nil, fmt.Errorf("decrypt api key: %w", err)
	}

	switch provider.Type {
	case llm.ProviderAnthropic:
		return c.callAnthropicWithTools(ctx, provider.BaseURL, apiKey, systemPrompt, messages, tools)
	case llm.ProviderOpenAI:
		return c.callOpenAIWithTools(ctx, provider.BaseURL, apiKey, systemPrompt, messages, tools)
	default:
		return "", nil, fmt.Errorf("unsupported provider type: %s", provider.Type)
	}
}

func (c *ContextLLMClient) callAnthropicWithTools(ctx context.Context, baseURL, apiKey, systemPrompt string, messages []AgentMessage, tools []AgentTool) (string, []AgentToolCall, error) {
	// Convert messages to Anthropic format
	anthropicMsgs := make([]map[string]any, 0, len(messages))
	for _, m := range messages {
		msg := map[string]any{"role": m.Role}
		if m.Content != "" {
			msg["content"] = m.Content
		}
		if len(m.Tools) > 0 {
			content := make([]map[string]any, 0)
			if m.Content != "" {
				content = append(content, map[string]any{"type": "text", "text": m.Content})
			}
			for _, tc := range m.Tools {
				content = append(content, map[string]any{
					"type":       "tool_use",
					"id":         tc.ID,
					"name":       tc.Name,
					"input":      tc.Arguments,
				})
			}
			msg["content"] = content
		}
		anthropicMsgs = append(anthropicMsgs, msg)
	}

	// Convert tools to Anthropic format
	anthropicTools := make([]map[string]any, 0, len(tools))
	for _, t := range tools {
		anthropicTools = append(anthropicTools, map[string]any{
			"name":        t.Name,
			"description": t.Description,
			"input_schema": t.InputSchema,
		})
	}

	reqBody := map[string]any{
		"model":      "claude-sonnet-4-5",
		"max_tokens": 4096,
		"system":     systemPrompt,
		"messages":   anthropicMsgs,
		"tools":      anthropicTools,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("marshal request: %w", err)
	}

	url := strings.TrimRight(baseURL, "/") + "/v1/messages"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("anthropic request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("anthropic %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Content []struct {
			Type  string          `json:"type"`
			Text  string          `json:"text,omitempty"`
			ID    string          `json:"id,omitempty"`
			Name  string          `json:"name,omitempty"`
			Input json.RawMessage `json:"input,omitempty"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
		Error      *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", nil, fmt.Errorf("parse response: %w", err)
	}

	if result.Error != nil {
		return "", nil, fmt.Errorf("anthropic error: %s", result.Error.Message)
	}

	var textContent string
	var toolCalls []AgentToolCall

	for _, c := range result.Content {
		if c.Type == "text" {
			textContent += c.Text
		} else if c.Type == "tool_use" {
			toolCalls = append(toolCalls, AgentToolCall{
				ID:        c.ID,
				Name:      c.Name,
				Arguments: c.Input,
			})
		}
	}

	return textContent, toolCalls, nil
}

func (c *ContextLLMClient) callOpenAIWithTools(ctx context.Context, baseURL, apiKey, systemPrompt string, messages []AgentMessage, tools []AgentTool) (string, []AgentToolCall, error) {
	// Convert messages to OpenAI format
	openaiMsgs := make([]map[string]any, 0, len(messages)+1)
	openaiMsgs = append(openaiMsgs, map[string]any{"role": "system", "content": systemPrompt})

	for _, m := range messages {
		msg := map[string]any{"role": m.Role}
		if len(m.Tools) > 0 {
			// Assistant message with tool calls
			msg["content"] = m.Content
			tcs := make([]map[string]any, 0, len(m.Tools))
			for _, tc := range m.Tools {
				tcs = append(tcs, map[string]any{
					"id":        tc.ID,
					"type":      "function",
					"function": map[string]any{
						"name":      tc.Name,
						"arguments": string(tc.Arguments),
					},
				})
			}
			msg["tool_calls"] = tcs
		} else {
			msg["content"] = m.Content
		}
		openaiMsgs = append(openaiMsgs, msg)
	}

	// Convert tools to OpenAI format
	openaiTools := make([]map[string]any, 0, len(tools))
	for _, t := range tools {
		openaiTools = append(openaiTools, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        t.Name,
				"description": t.Description,
				"parameters":  t.InputSchema,
			},
		})
	}

	reqBody := map[string]any{
		"model":      "gpt-4o",
		"max_tokens": 4096,
		"messages":   openaiMsgs,
		"tools":      openaiTools,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("marshal request: %w", err)
	}

	url := strings.TrimRight(baseURL, "/") + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("openai request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("openai %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content   string `json:"content"`
				ToolCalls []struct {
					ID   string `json:"id"`
					Type string `json:"type"`
					Function struct {
						Name      string          `json:"name"`
						Arguments json.RawMessage `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", nil, fmt.Errorf("parse response: %w", err)
	}

	if result.Error != nil {
		return "", nil, fmt.Errorf("openai error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", nil, fmt.Errorf("empty response from openai")
	}

	choice := result.Choices[0]
	var toolCalls []AgentToolCall

	for _, tc := range choice.Message.ToolCalls {
		toolCalls = append(toolCalls, AgentToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}

	return choice.Message.Content, toolCalls, nil
}
