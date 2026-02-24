package llm

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	errorCooldown   = 60 * time.Second // 出错后冷却时间
	maxErrorCount   = 5                // 超过此错误次数进入 Down 状态
	healthInterval  = 30 * time.Second
)

// Router 负载均衡 + 故障转移
type Router struct {
	repo      *Repository
	encKey    string

	mu        sync.RWMutex
	cooldowns map[string]time.Time // provider_id → 冷却到期时间
}

func NewRouter(repo *Repository, encKey string) *Router {
	r := &Router{
		repo:      repo,
		encKey:    encKey,
		cooldowns: make(map[string]time.Time),
	}
	return r
}

// PickProvider 为指定公司和类型选择 provider
// preferModel: 请求中指定的模型名，优先选包含该模型的 provider；为空则不过滤
// 返回 provider 和已解密的 API Key
func (r *Router) PickProvider(ctx context.Context, companyID string, pt ProviderType, preferModel string) (*Provider, string, error) {
	providers, err := r.repo.ListActiveProviders(ctx, companyID)
	if err != nil {
		return nil, "", fmt.Errorf("list providers: %w", err)
	}

	// 按类型过滤
	var typed []*Provider
	for _, p := range providers {
		if p.Type == pt {
			typed = append(typed, p)
		}
	}

	if len(typed) == 0 {
		return nil, "", fmt.Errorf("no active %s provider configured for this company", pt)
	}

	// 若指定了 preferModel，优先选支持该模型的 provider
	if preferModel != "" {
		var matched []*Provider
		for _, p := range typed {
			for _, m := range p.Models {
				if m == preferModel {
					matched = append(matched, p)
					break
				}
			}
		}
		if len(matched) > 0 {
			typed = matched
		}
		// 无匹配则 fallback 到全部同类 provider
	}

	// 冷却过滤（若全部冷却则兜底使用第一个）
	r.mu.RLock()
	available := make([]*Provider, 0, len(typed))
	for _, p := range typed {
		if exp, cooling := r.cooldowns[p.ID]; cooling && time.Now().Before(exp) {
			continue
		}
		available = append(available, p)
	}
	r.mu.RUnlock()

	if len(available) == 0 {
		r.clearOldestCooldown()
		available = typed
	}

	var picked *Provider
	if len(available) == 1 {
		picked = available[0]
	} else {
		picked = weightedPick(available)
	}

	apiKey, err := DecryptAPIKey(picked.APIKeyEnc, r.encKey)
	if err != nil {
		return nil, "", fmt.Errorf("decrypt api key: %w", err)
	}
	return picked, apiKey, nil
}

// MarkError 标记 provider 出错并设置冷却
func (r *Router) MarkError(ctx context.Context, providerID string) {
	r.repo.MarkProviderError(ctx, providerID) //nolint:errcheck
	r.mu.Lock()
	r.cooldowns[providerID] = time.Now().Add(errorCooldown)
	r.mu.Unlock()
}

// MarkSuccess 清除 provider 冷却
func (r *Router) MarkSuccess(ctx context.Context, providerID string) {
	r.repo.MarkProviderUsed(ctx, providerID) //nolint:errcheck
	r.mu.Lock()
	delete(r.cooldowns, providerID)
	r.mu.Unlock()
}

// GetStatus 获取 provider 运行时状态
func (r *Router) GetStatus(p *Provider) ProviderStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if exp, cooling := r.cooldowns[p.ID]; cooling && time.Now().Before(exp) {
		return StatusDown
	}
	if p.ErrorCount >= maxErrorCount {
		return StatusDegraded
	}
	return StatusHealthy
}

func (r *Router) clearOldestCooldown() {
	r.mu.Lock()
	defer r.mu.Unlock()
	var oldest string
	var oldestTime time.Time
	for id, exp := range r.cooldowns {
		if oldest == "" || exp.Before(oldestTime) {
			oldest = id
			oldestTime = exp
		}
	}
	if oldest != "" {
		delete(r.cooldowns, oldest)
	}
}

// weightedPick 加权随机选择：权重越高越容易被选中
func weightedPick(providers []*Provider) *Provider {
	totalWeight := 0
	for _, p := range providers {
		totalWeight += p.Weight
	}
	if totalWeight == 0 {
		return providers[0]
	}
	// 使用时间纳秒作为简单随机源（避免引入 math/rand）
	n := int(time.Now().UnixNano() % int64(totalWeight))
	for _, p := range providers {
		n -= p.Weight
		if n < 0 {
			return p
		}
	}
	return providers[len(providers)-1]
}
