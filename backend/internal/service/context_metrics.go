package service

import (
	"context"
	"sync"
	"time"
)

// ContextMetrics 上下文服务指标收集器
type ContextMetrics struct {
	mu sync.RWMutex

	// tool call 统计
	toolCalls map[string]*ToolCallStats

	// 失败原因统计
	failures map[string]int

	// latency 直方图 buckets (ms)
	latencyBuckets []int64
	latencyCounts  map[int64]int64

	// token 统计
	inputTokens  int64
	outputTokens int64

	// cost 统计 (microdollars)
	totalCost int64

	// 搜索统计
	searchCount      int64
	searchSuccess    int64
	searchFallback   int64 // 降级到全文搜索的次数
}

// ToolCallStats tool call 统计
type ToolCallStats struct {
	Name      string `json:"name"`
	Count     int64  `json:"count"`
	Success   int64  `json:"success"`
	Error     int64  `json:"error"`
	TotalMs   int64  `json:"total_ms"`
	MinMs     int64  `json:"min_ms"`
	MaxMs     int64  `json:"max_ms"`
}

// NewContextMetrics 创建指标收集器
func NewContextMetrics() *ContextMetrics {
	return &ContextMetrics{
		toolCalls: make(map[string]*ToolCallStats),
		failures:  make(map[string]int),
		latencyBuckets: []int64{10, 50, 100, 250, 500, 1000, 2500, 5000, 10000},
		latencyCounts:  make(map[int64]int64),
	}
}

// RecordToolCall 记录 tool call 指标
func (m *ContextMetrics) RecordToolCall(name string, durationMs int64, isError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats, ok := m.toolCalls[name]
	if !ok {
		stats = &ToolCallStats{
			Name:    name,
			MinMs:   durationMs,
			MaxMs:   durationMs,
		}
		m.toolCalls[name] = stats
	}

	stats.Count++
	stats.TotalMs += durationMs

	if durationMs < stats.MinMs {
		stats.MinMs = durationMs
	}
	if durationMs > stats.MaxMs {
		stats.MaxMs = durationMs
	}

	if isError {
		stats.Error++
	} else {
		stats.Success++
	}
}

// RecordFailure 记录失败原因
func (m *ContextMetrics) RecordFailure(reason string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failures[reason]++
}

// RecordLatency 记录延迟到直方图
func (m *ContextMetrics) RecordLatency(ms int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 找到对应的 bucket
	for _, bucket := range m.latencyBuckets {
		if ms <= bucket {
			m.latencyCounts[bucket]++
			return
		}
	}
	// 超出最大 bucket
	m.latencyCounts[10000]++
}

// RecordTokens 记录 token 使用
func (m *ContextMetrics) RecordTokens(input, output int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inputTokens += int64(input)
	m.outputTokens += int64(output)
}

// RecordCost 记录成本
func (m *ContextMetrics) RecordCost(cost int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalCost += cost
}

// RecordSearch 记录搜索指标
func (m *ContextMetrics) RecordSearch(isSuccess, isFallback bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.searchCount++
	if isSuccess {
		m.searchSuccess++
	}
	if isFallback {
		m.searchFallback++
	}
}

// GetSnapshot 获取指标快照
func (m *ContextMetrics) GetSnapshot() *MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	toolCallList := make([]*ToolCallStats, 0, len(m.toolCalls))
	for _, stats := range m.toolCalls {
		toolCallList = append(toolCallList, stats)
	}

	latencySnapshot := make(map[int64]int64)
	for k, v := range m.latencyCounts {
		latencySnapshot[k] = v
	}

	failureList := make(map[string]int)
	for k, v := range m.failures {
		failureList[k] = v
	}

	return &MetricsSnapshot{
		ToolCalls:      toolCallList,
		Failures:       failureList,
		LatencyCounts:  latencySnapshot,
		InputTokens:    m.inputTokens,
		OutputTokens:   m.outputTokens,
		TotalCost:      m.totalCost,
		SearchCount:    m.searchCount,
		SearchSuccess:  m.searchSuccess,
		SearchFallback: m.searchFallback,
	}
}

// MetricsSnapshot 指标快照
type MetricsSnapshot struct {
	ToolCalls      []*ToolCallStats   `json:"tool_calls"`
	Failures       map[string]int     `json:"failures"`
	LatencyCounts  map[int64]int64    `json:"latency_counts"`
	InputTokens    int64              `json:"input_tokens"`
	OutputTokens   int64              `json:"output_tokens"`
	TotalCost      int64              `json:"total_cost"`
	SearchCount    int64              `json:"search_count"`
	SearchSuccess  int64              `json:"search_success"`
	SearchFallback int64              `json:"search_fallback"`
}

// GetLatencyPercentile 获取延迟百分位数
func (m *ContextMetrics) GetLatencyPercentile(p float64) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := int64(0)
	for _, count := range m.latencyCounts {
		total += count
	}

	if total == 0 {
		return 0
	}

	target := int64(float64(total) * p)
	cumulative := int64(0)

	for _, bucket := range m.latencyBuckets {
		cumulative += m.latencyCounts[bucket]
		if cumulative >= target {
			return bucket
		}
	}

	return 10000
}

// Reset 重置所有指标（用于定期清空）
func (m *ContextMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.toolCalls = make(map[string]*ToolCallStats)
	m.failures = make(map[string]int)
	m.latencyCounts = make(map[int64]int64)
	m.inputTokens = 0
	m.outputTokens = 0
	m.totalCost = 0
	m.searchCount = 0
	m.searchSuccess = 0
	m.searchFallback = 0
}

// SearchMetrics 搜索操作指标上下文
type SearchMetrics struct {
	StartTime   time.Time
	DirectoryID string
	Query       string
	ToolCalls   int
	InputTokens int
	OutputTokens int
	Cost        int64
	Error       error
}

// MetricsCollector 指标收集器接口
type MetricsCollector interface {
	StartSearch(ctx context.Context, query string, directoryIDs []string) *SearchMetrics
	EndSearch(metrics *SearchMetrics, err error)
	RecordToolCall(name string, durationMs int64, isError bool)
	RecordFailure(reason string)
}
