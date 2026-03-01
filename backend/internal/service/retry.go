package service

import (
	"context"
	"fmt"
	"math"
	"time"
)

// RetryPolicy 重试策略配置
type RetryPolicy struct {
	MaxAttempts int           // 最大重试次数
	BackoffBase int           // 指数退避基数（秒）
	BackoffMax  int           // 最大退避时间（秒）
	Timeout     time.Duration // 单次请求超时时间
}

// DefaultRetryPolicy 默认重试策略：3 次重试，指数退避 2s, 4s, 8s，最大 60s
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: 3,
		BackoffBase: 2,
		BackoffMax:  60,
		Timeout:     30 * time.Second,
	}
}

// NoRetryPolicy 不重试策略
func NoRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: 0,
		Timeout:     30 * time.Second,
	}
}

// RetryConfig 重试配置
type RetryConfig struct {
	Policy          RetryPolicy                     // 重试策略
	IsRetryable     func(error) bool                // 判断错误是否可重试（可选）
	OnRetry         func(int, error) time.Duration  // 每次重试前的回调（可选，返回自定义等待时间）
	LogFunc         func(int, error)                // 日志记录回调（可选）
	ContextOverride context.Context                 // 覆盖 context（可选，用于设置自定义超时）
}

// RetryFunc 可重试的函数类型
type RetryFunc func(ctx context.Context) error

// Retry 执行带重试的函数调用
// 返回最后一次错误，成功则返回 nil
func Retry(ctx context.Context, fn RetryFunc, config RetryConfig) error {
	policy := config.Policy
	if policy.MaxAttempts <= 0 {
		policy.MaxAttempts = 1
	}

	var lastErr error

	for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
		// 创建带有超时的 context
		reqCtx := ctx
		if config.ContextOverride != nil {
			reqCtx = config.ContextOverride
		} else if policy.Timeout > 0 {
			var cancel context.CancelFunc
			reqCtx, cancel = context.WithTimeout(ctx, policy.Timeout)
			defer cancel()
		}

		// 执行函数
		err := fn(reqCtx)
		if err == nil {
			return nil
		}

		lastErr = err

		// 记录日志
		if config.LogFunc != nil && attempt > 0 {
			config.LogFunc(attempt, err)
		}

		// 检查是否可重试
		if !isRetryableError(err, config.IsRetryable) {
			return err
		}

		// 检查是否还有重试次数
		if attempt >= policy.MaxAttempts-1 {
			return err
		}

		// 计算等待时间
		var wait time.Duration
		if config.OnRetry != nil {
			wait = config.OnRetry(attempt, err)
		} else {
			wait = calcExponentialBackoff(attempt+1, policy)
		}

		if wait > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}
	}

	return lastErr
}

// isRetryableError 判断错误是否可重试
func isRetryableError(err error, customCheck func(error) bool) bool {
	if err == nil {
		return false
	}

	if customCheck != nil {
		return customCheck(err)
	}

	// 默认重试策略：网络错误、超时、5xx 等
	errStr := err.Error()
	retryableErrors := []string{
		"timeout",
		"deadline",
		"connection",
		"dial",
		"request",
		"EOF",
		"no such host",
		"i/o timeout",
		"connection reset",
		"connection refused",
	}

	for _, s := range retryableErrors {
		if containsIgnoreCase(errStr, s) {
			return true
		}
	}

	return false
}

// calcExponentialBackoff 计算指数退避时间
func calcExponentialBackoff(attempt int, policy RetryPolicy) time.Duration {
	if attempt <= 0 {
		attempt = 1
	}
	if policy.BackoffBase <= 0 {
		policy.BackoffBase = 2
	}

	// 指数计算：base^(attempt-1)
	seconds := math.Pow(float64(policy.BackoffBase), float64(attempt-1))

	// 限制最大值
	if policy.BackoffMax > 0 && int(seconds) > policy.BackoffMax {
		seconds = float64(policy.BackoffMax)
	}

	if seconds < 1 {
		seconds = 1
	}

	return time.Duration(seconds) * time.Second
}

// containsIgnoreCase 判断字符串是否包含子串（忽略大小写）
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && containsLower(s, substr)
}

func containsLower(s, substr string) bool {
	s = toLowerASCII(s)
	substr = toLowerASCII(substr)
	return contains(s, substr)
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toLowerASCII(s string) string {
	b := []byte(s)
	for i := 0; i < len(b); i++ {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}

// WithRetry 函数包装器：为任何函数添加重试能力
func WithRetry[T any](ctx context.Context, fn func(context.Context) (T, error), config RetryConfig) (T, error) {
	var result T
	err := Retry(ctx, func(ctx context.Context) error {
		var err error
		result, err = fn(ctx)
		return err
	}, config)
	return result, err
}

// RetryableError 包装错误为可重试错误
type RetryableError struct {
	Err       error
	Retryable bool
}

func (e *RetryableError) Error() string {
	if e.Err == nil {
		return "retryable error"
	}
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// NewRetryableError 创建可重试错误
func NewRetryableError(err error) *RetryableError {
	return &RetryableError{Err: err, Retryable: true}
}

// NewNonRetryableError 创建不可重试错误
func NewNonRetryableError(err error) *RetryableError {
	return &RetryableError{Err: err, Retryable: false}
}

// HTTP 相关重试辅助

// HTTPRetryConfig 为 HTTP 请求创建默认重试配置
func HTTPRetryConfig() RetryConfig {
	return RetryConfig{
		Policy: DefaultRetryPolicy(),
		IsRetryable: func(err error) bool {
			if err == nil {
				return false
			}
			// HTTP 请求典型的可重试错误
			errStr := err.Error()
			return containsIgnoreCase(errStr, "timeout") ||
				containsIgnoreCase(errStr, "connection") ||
				containsIgnoreCase(errStr, "dial") ||
				containsIgnoreCase(errStr, "EOF") ||
				containsIgnoreCase(errStr, "503") ||
				containsIgnoreCase(errStr, "502") ||
				containsIgnoreCase(errStr, "504")
		},
	}
}

// NewHTTPStatusError 为 HTTP 状态码创建错误（标记是否可重试）
func NewHTTPStatusError(statusCode int, body string) error {
	if statusCode < 500 && statusCode != 429 {
		return fmt.Errorf("HTTP %d: %s", statusCode, body)
	}
	// 5xx 和 429 是可重试的
	return NewRetryableError(fmt.Errorf("HTTP %d: %s", statusCode, body))
}
