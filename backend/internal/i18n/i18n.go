package i18n

import (
	"sync"
)

// Locale represents a supported language
type Locale string

const (
	LocaleEN Locale = "en"
	LocaleZH Locale = "zh"
)

// MessageKey represents a translation key
type MessageKey string

// Error message keys
const (
	ErrUnauthorized        MessageKey = "err.unauthorized"
	ErrForbidden           MessageKey = "err.forbidden"
	ErrNotFound            MessageKey = "err.not_found"
	ErrInternalServerError MessageKey = "err.internal_server_error"
	ErrBadRequest          MessageKey = "err.bad_request"
	ErrConflict            MessageKey = "err.conflict"
	ErrTimeout             MessageKey = "err.timeout"
	ErrRateLimit           MessageKey = "err.rate_limit"
	ErrInvalidInput        MessageKey = "err.invalid_input"
	ErrMissingField        MessageKey = "err.missing_field"
	ErrInvalidFormat       MessageKey = "err.invalid_format"
	ErrAlreadyExists       MessageKey = "err.already_exists"
	ErrNotInitialized      MessageKey = "err.not_initialized"
)

// messages contains translations for all supported locales
var messages = map[Locale]map[MessageKey]string{
	LocaleEN: {
		ErrUnauthorized:        "Unauthorized",
		ErrForbidden:           "Access denied",
		ErrNotFound:            "Resource not found",
		ErrInternalServerError: "Internal server error",
		ErrBadRequest:          "Invalid request",
		ErrConflict:            "Resource already exists",
		ErrTimeout:             "Request timeout",
		ErrRateLimit:           "Too many requests, please try again later",
		ErrInvalidInput:        "Invalid input",
		ErrMissingField:        "Missing required field: %s",
		ErrInvalidFormat:       "Invalid format: %s",
		ErrAlreadyExists:       "%s already exists",
		ErrNotInitialized:      "System not initialized, please complete setup first",
	},
	LocaleZH: {
		ErrUnauthorized:        "未授权，请先登录",
		ErrForbidden:           "无权访问",
		ErrNotFound:            "资源不存在",
		ErrInternalServerError: "服务器内部错误",
		ErrBadRequest:          "请求无效",
		ErrConflict:            "资源已存在",
		ErrTimeout:             "请求超时",
		ErrRateLimit:           "请求过于频繁，请稍后重试",
		ErrInvalidInput:        "输入无效",
		ErrMissingField:        "缺少必填字段：%s",
		ErrInvalidFormat:       "格式错误：%s",
		ErrAlreadyExists:       "%s 已存在",
		ErrNotInitialized:      "系统未初始化，请先完成初始化设置",
	},
}

// Translator handles message translation
type Translator struct {
	mu sync.RWMutex
}

var (
	translatorInstance *Translator
	once               sync.Once
)

// GetTranslator returns the singleton translator instance
func GetTranslator() *Translator {
	once.Do(func() {
		translatorInstance = &Translator{}
	})
	return translatorInstance
}

// Translate returns the translated message for the given key and locale
func (t *Translator) Translate(key MessageKey, locale Locale, args ...interface{}) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Fallback to English if locale not found
	msgs, ok := messages[locale]
	if !ok {
		msgs = messages[LocaleEN]
	}

	msg, ok := msgs[key]
	if !ok {
		// Return key as fallback
		return string(key)
	}

	// Format with args if provided
	if len(args) > 0 {
		return formatString(msg, args...)
	}

	return msg
}

// formatString is a simple string formatter
func formatString(format string, args ...interface{}) string {
	// Simple implementation - in production, use fmt.Sprintf
	result := format
	for _, arg := range args {
		result = replaceFirst(result, "%s", toString(arg))
	}
	return result
}

func replaceFirst(s, old, new string) string {
	idx := findSubstring(s, old)
	if idx == -1 {
		return s
	}
	return s[:idx] + new + s[idx+len(old):]
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	// Simple conversion for common types
	switch val := v.(type) {
	case int:
		return itoa(val)
	case int64:
		return i64toa(val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return "?"
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	buf := make([]byte, 0, 20)
	for i > 0 {
		buf = append(buf, byte('0'+(i%10)))
		i /= 10
	}
	if neg {
		buf = append(buf, '-')
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}

func i64toa(i int64) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	buf := make([]byte, 0, 20)
	for i > 0 {
		buf = append(buf, byte('0'+(i%10)))
		i /= 10
	}
	if neg {
		buf = append(buf, '-')
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
