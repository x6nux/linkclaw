package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const maxRetries = 3

// ProxyService 处理 LLM API 代理请求
type ProxyService struct {
	repo   *Repository
	router *Router
	client *http.Client
	encKey string
}

func NewProxyService(repo *Repository, router *Router, encKey string) *ProxyService {
	return &ProxyService{
		repo:   repo,
		router: router,
		client: &http.Client{Timeout: 120 * time.Second},
		encKey: encKey,
	}
}

// ProxyRequest 执行代理请求（带重试 + 故障转移）
// w: 响应写入目标（gin ResponseWriter）
// body: 原始请求体（已读取）
// agentID: 发起请求的 agent
// companyID: 所属公司
func (s *ProxyService) ProxyRequest(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	body []byte,
	companyID, agentID string,
	providerType ProviderType,
) error {
	// 从请求体解析 model，用于路由和日志
	var requestedModel string
	var reqModelBody struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(body, &reqModelBody); err == nil {
		requestedModel = reqModelBody.Model
	}

	var lastErr error
	start := time.Now()

	for attempt := 0; attempt < maxRetries; attempt++ {
		provider, apiKey, err := s.router.PickProvider(ctx, companyID, providerType, requestedModel)
		if err != nil {
			return err
		}

		err = s.doProxy(ctx, w, r, body, provider, apiKey, companyID, agentID, requestedModel, int16(attempt), start)
		if err == nil {
			return nil
		}

		lastErr = err
		s.router.MarkError(ctx, provider.ID)

		// 5xx 或超时才重试
		if !isRetryable(err) {
			break
		}
	}
	return lastErr
}

func (s *ProxyService) doProxy(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	body []byte,
	provider *Provider,
	apiKey string,
	companyID, agentID string,
	requestedModel string,
	retryCount int16,
	start time.Time,
) error {
	// 构建上游请求 URL
	upstreamURL := buildUpstreamURL(provider, req.URL.Path)

	upReq, err := http.NewRequestWithContext(ctx, req.Method, upstreamURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	// 透传原始请求头，跳过逐跳头和认证头（认证头由网关注入）
	for k, vs := range req.Header {
		switch strings.ToLower(k) {
		case "host", "connection", "keep-alive", "transfer-encoding",
			"authorization", "x-api-key":
			continue
		}
		for _, v := range vs {
			upReq.Header.Add(k, v)
		}
	}

	// 注入认证头；Anthropic 特有头缺失时补默认值
	switch provider.Type {
	case ProviderAnthropic:
		upReq.Header.Set("x-api-key", apiKey)
		if upReq.Header.Get("anthropic-version") == "" {
			upReq.Header.Set("anthropic-version", "2023-06-01")
		}
		if upReq.Header.Get("anthropic-beta") == "" {
			upReq.Header.Set("anthropic-beta", "interleaved-thinking-2025-05-14,token-efficient-tools-2025-02-19")
		}
	case ProviderOpenAI:
		upReq.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := s.client.Do(upReq)
	if err != nil {
		return &proxyError{msg: err.Error(), retryable: true}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return &proxyError{msg: fmt.Sprintf("upstream %d", resp.StatusCode), retryable: true}
	}
	if resp.StatusCode == 429 {
		return &proxyError{msg: "rate limited", retryable: true}
	}

	// 转发响应头
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)

	isStream := strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream")

	var usageLog UsageLog
	usageLog.CompanyID = companyID
	usageLog.ProviderID = &provider.ID
	if agentID != "" {
		usageLog.AgentID = &agentID
	}
	usageLog.RequestModel = requestedModel
	usageLog.RetryCount = retryCount
	usageLog.Status = "success"

	if isStream {
		s.streamResponse(ctx, w, resp.Body, provider, &usageLog)
	} else {
		s.bufferedResponse(ctx, w, resp.Body, provider, &usageLog)
	}

	latency := int(time.Since(start).Milliseconds())
	usageLog.LatencyMs = &latency
	s.repo.InsertUsageLog(ctx, &usageLog) //nolint:errcheck
	s.router.MarkSuccess(ctx, provider.ID)
	return nil
}

// streamResponse 流式响应：边转发边解析 token 用量
func (s *ProxyService) streamResponse(ctx context.Context, w http.ResponseWriter, body io.Reader, p *Provider, log *UsageLog) {
	flusher, canFlush := w.(http.Flusher)
	buf := make([]byte, 4096)
	var accumulated strings.Builder

	for {
		n, err := body.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			w.Write(chunk) //nolint:errcheck
			if canFlush {
				flusher.Flush()
			}
			accumulated.Write(chunk)
		}
		if err != nil {
			break
		}
	}

	// 解析最终 usage 数据
	parseStreamUsage(accumulated.String(), p.Type, log)
}

// bufferedResponse 非流式响应：全量读取解析
func (s *ProxyService) bufferedResponse(ctx context.Context, w http.ResponseWriter, body io.Reader, p *Provider, log *UsageLog) {
	data, _ := io.ReadAll(body)
	w.Write(data) //nolint:errcheck
	parseBufferedUsage(data, p.Type, log)
}

// parseStreamUsage 从 SSE 流中提取 usage 信息
func parseStreamUsage(stream string, pt ProviderType, log *UsageLog) {
	lines := strings.Split(stream, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			continue
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal([]byte(data), &m); err != nil {
			continue
		}
		if usage, ok := m["usage"]; ok {
			applyUsage(usage, pt, log)
		}
		// Anthropic stream: message_delta 包含 usage
		if msgType, ok := m["type"]; ok {
			var t string
			json.Unmarshal(msgType, &t)
			if t == "message_delta" || t == "message_start" {
				if msg, ok := m["message"]; ok {
					var inner map[string]json.RawMessage
					if json.Unmarshal(msg, &inner) == nil {
						if usage, ok := inner["usage"]; ok {
							applyUsage(usage, pt, log)
						}
					}
				}
				if usage, ok := m["usage"]; ok {
					applyUsage(usage, pt, log)
				}
			}
		}
	}
	if pt == ProviderAnthropic {
		log.CostMicrodollars = CalcAnthropicCost(log.RequestModel, AnthropicUsage{
			InputTokens:              log.InputTokens,
			OutputTokens:             log.OutputTokens,
			CacheCreationInputTokens: log.CacheCreationTokens,
			CacheReadInputTokens:     log.CacheReadTokens,
		})
	}
}

// parseBufferedUsage 从完整响应体中提取 usage
func parseBufferedUsage(data []byte, pt ProviderType, log *UsageLog) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return
	}
	if usage, ok := m["usage"]; ok {
		applyUsage(usage, pt, log)
	}
	if pt == ProviderAnthropic {
		log.CostMicrodollars = CalcAnthropicCost(log.RequestModel, AnthropicUsage{
			InputTokens:              log.InputTokens,
			OutputTokens:             log.OutputTokens,
			CacheCreationInputTokens: log.CacheCreationTokens,
			CacheReadInputTokens:     log.CacheReadTokens,
		})
	} else {
		var usage OpenAIUsage
		if uRaw, ok := m["usage"]; ok {
			json.Unmarshal(uRaw, &usage)
			log.CostMicrodollars = CalcOpenAICost(log.RequestModel, usage)
		}
	}
}

func applyUsage(raw json.RawMessage, pt ProviderType, log *UsageLog) {
	switch pt {
	case ProviderAnthropic:
		var u AnthropicUsage
		if json.Unmarshal(raw, &u) == nil {
			if u.InputTokens > 0 {
				log.InputTokens = u.InputTokens
			}
			if u.OutputTokens > 0 {
				log.OutputTokens = u.OutputTokens
			}
			if u.CacheCreationInputTokens > 0 {
				log.CacheCreationTokens = u.CacheCreationInputTokens
			}
			if u.CacheReadInputTokens > 0 {
				log.CacheReadTokens = u.CacheReadInputTokens
			}
		}
	case ProviderOpenAI:
		var u OpenAIUsage
		if json.Unmarshal(raw, &u) == nil {
			log.InputTokens = u.PromptTokens
			log.OutputTokens = u.CompletionTokens
			if u.PromptTokensDetails != nil {
				log.CachedPromptTokens = u.PromptTokensDetails.CachedTokens
			}
		}
	}
}

func buildUpstreamURL(p *Provider, path string) string {
	base := strings.TrimRight(p.BaseURL, "/")
	// 去除内部路由前缀，保留 API 路径
	path = strings.TrimPrefix(path, "/llm")
	path = strings.TrimPrefix(path, "/api/v1/llm")
	return base + path
}

// isRetryable 判断错误是否值得重试
func isRetryable(err error) bool {
	if pe, ok := err.(*proxyError); ok {
		return pe.retryable
	}
	return false
}

type proxyError struct {
	msg       string
	retryable bool
}

func (e *proxyError) Error() string { return e.msg }
