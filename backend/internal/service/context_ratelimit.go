package service

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiter 限流器接口
type RateLimiter interface {
	Allow(ctx context.Context, key string) bool
	Wait(ctx context.Context, key string) error
}

// TokenBucket 令牌桶限流器
type TokenBucket struct {
	mu           sync.Mutex
	tokens       float64
	maxTokens    float64
	refillRate   float64 // tokens per second
	lastRefill   time.Time
}

// NewTokenBucket 创建令牌桶限流器
func NewTokenBucket(maxTokens float64, refillRate float64) *TokenBucket {
	return &TokenBucket{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow 检查是否允许请求
func (tb *TokenBucket) Allow(ctx context.Context, key string) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// 补充令牌
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens = min(tb.maxTokens, tb.tokens+elapsed*tb.refillRate)
	tb.lastRefill = now

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

// Wait 等待直到有令牌可用
func (tb *TokenBucket) Wait(ctx context.Context, key string) error {
	for {
		if tb.Allow(ctx, key) {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			// 继续循环检查
		}
	}
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// 每目录每秒请求数
	DirectoryRPS float64 `json:"directory_rps"`
	// 每目录并发请求数
	DirectoryConcurrency int `json:"directory_concurrency"`
	// 全局每秒请求数
	GlobalRPS float64 `json:"global_rps"`
	// 全局并发请求数
	GlobalConcurrency int `json:"global_concurrency"`
	// 每请求 token 预算
	TokenBudgetPerRequest int `json:"token_budget_per_request"`
	// 超时配置
	SearchTimeout       time.Duration `json:"search_timeout"`
	LLMCallTimeout      time.Duration `json:"llm_call_timeout"`
	FileReadTimeout     time.Duration `json:"file_read_timeout"`
}

// DefaultRateLimitConfig 返回默认限流配置
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		DirectoryRPS:          10,
		DirectoryConcurrency:  5,
		GlobalRPS:            100,
		GlobalConcurrency:     50,
		TokenBudgetPerRequest: 100000, // 100K tokens
		SearchTimeout:        30 * time.Second,
		LLMCallTimeout:       60 * time.Second,
		FileReadTimeout:      5 * time.Second,
	}
}

// RateLimiterManager 限流器管理器
type RateLimiterManager struct {
	mu            sync.RWMutex
	config        *RateLimitConfig
	globalBucket  *TokenBucket
	dirBuckets    map[string]*TokenBucket
	dirSemaphores map[string]chan struct{}
	globalSem     chan struct{}
}

// NewRateLimiterManager 创建限流器管理器
func NewRateLimiterManager(config *RateLimitConfig) *RateLimiterManager {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	return &RateLimiterManager{
		config:        config,
		globalBucket:  NewTokenBucket(config.GlobalRPS*2, config.GlobalRPS),
		dirBuckets:    make(map[string]*TokenBucket),
		dirSemaphores: make(map[string]chan struct{}),
		globalSem:     make(chan struct{}, config.GlobalConcurrency),
	}
}

// getDirBucket 获取或创建目录限流桶
func (m *RateLimiterManager) getDirBucket(dirID string) *TokenBucket {
	m.mu.RLock()
	bucket, ok := m.dirBuckets[dirID]
	m.mu.RUnlock()

	if ok {
		return bucket
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 双重检查
	if bucket, ok = m.dirBuckets[dirID]; ok {
		return bucket
	}

	bucket = NewTokenBucket(m.config.DirectoryRPS*2, m.config.DirectoryRPS)
	m.dirBuckets[dirID] = bucket
	return bucket
}

// getDirSemaphore 获取或创建目录信号量
func (m *RateLimiterManager) getDirSemaphore(dirID string) chan struct{} {
	m.mu.RLock()
	sem, ok := m.dirSemaphores[dirID]
	m.mu.RUnlock()

	if ok {
		return sem
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 双重检查
	if sem, ok = m.dirSemaphores[dirID]; ok {
		return sem
	}

	sem = make(chan struct{}, m.config.DirectoryConcurrency)
	m.dirSemaphores[dirID] = sem
	return sem
}

// AcquireGlobal 获取全局许可
func (m *RateLimiterManager) AcquireGlobal(ctx context.Context) error {
	if !m.globalBucket.Allow(ctx, "global") {
		return fmt.Errorf("全局请求频率超限")
	}

	select {
	case m.globalSem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ReleaseGlobal 释放全局许可
func (m *RateLimiterManager) ReleaseGlobal() {
	<-m.globalSem
}

// AcquireDirectory 获取目录许可
func (m *RateLimiterManager) AcquireDirectory(ctx context.Context, dirID string) error {
	bucket := m.getDirBucket(dirID)
	if !bucket.Allow(ctx, dirID) {
		return fmt.Errorf("目录 %s 请求频率超限", dirID)
	}

	sem := m.getDirSemaphore(dirID)
	select {
	case sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ReleaseDirectory 释放目录许可
func (m *RateLimiterManager) ReleaseDirectory(dirID string) {
	sem := m.getDirSemaphore(dirID)
	<-sem
}

// CheckTokenBudget 检查 token 预算
func (m *RateLimiterManager) CheckTokenBudget(estimatedTokens int) bool {
	return estimatedTokens <= m.config.TokenBudgetPerRequest
}

// GetTimeout 获取操作类型的超时
func (m *RateLimiterManager) GetTimeout(opType string) time.Duration {
	switch opType {
	case "search":
		return m.config.SearchTimeout
	case "llm_call":
		return m.config.LLMCallTimeout
	case "file_read":
		return m.config.FileReadTimeout
	default:
		return 30 * time.Second
	}
}

// CostBudget 成本预算控制器
type CostBudget struct {
	mu              sync.RWMutex
	budgetMicroDollars int64
	spentMicroDollars int64
	warnThreshold   float64 // 0.0-1.0
	hardLimitEnabled bool
}

// NewCostBudget 创建成本预算
func NewCostBudget(budgetMicroDollars int64, warnThreshold float64, hardLimitEnabled bool) *CostBudget {
	return &CostBudget{
		budgetMicroDollars: budgetMicroDollars,
		spentMicroDollars:  0,
		warnThreshold:      warnThreshold,
		hardLimitEnabled:   hardLimitEnabled,
	}
}

// CheckAndRecord 检查并记录成本
// 返回：是否允许，是否接近预警线，错误
func (cb *CostBudget) CheckAndRecord(cost int64) (bool, bool, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	newSpent := cb.spentMicroDollars + cost

	// 检查硬限制
	if cb.hardLimitEnabled && newSpent > cb.budgetMicroDollars {
		return false, false, fmt.Errorf("超出成本预算限制")
	}

	cb.spentMicroDollars = newSpent

	// 检查预警线
	ratio := float64(newSpent) / float64(cb.budgetMicroDollars)
	warned := ratio >= cb.warnThreshold

	return true, warned, nil
}

// GetRemaining 获取剩余预算
func (cb *CostBudget) GetRemaining() int64 {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.budgetMicroDollars - cb.spentMicroDollars
}

// GetUsageRatio 获取使用比例
func (cb *CostBudget) GetUsageRatio() float64 {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	if cb.budgetMicroDollars == 0 {
		return 0
	}
	return float64(cb.spentMicroDollars) / float64(cb.budgetMicroDollars)
}

// min 返回较小值
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
