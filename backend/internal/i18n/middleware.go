package i18n

import (
	"github.com/gin-gonic/gin"
)

const (
	ContextKeyLocale = "locale"
)

// LocaleMiddleware extracts locale from Accept-Language header
func LocaleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := extractLocale(c.GetHeader("Accept-Language"))
		c.Set(ContextKeyLocale, locale)
		c.Next()
	}
}

// extractLocale parses Accept-Language header and returns the best matching locale
func extractLocale(acceptLanguage string) Locale {
	if acceptLanguage == "" {
		return LocaleZH // Default to Chinese
	}

	// Simple parsing: check if "en" appears before "zh"
	enIdx := findSubstr(acceptLanguage, "en")
	zhIdx := findSubstr(acceptLanguage, "zh")

	if enIdx == -1 && zhIdx == -1 {
		return LocaleZH
	}
	if enIdx == -1 {
		return LocaleZH
	}
	if zhIdx == -1 {
		return LocaleEN
	}

	// Both present, check priority (lower index = higher priority)
	if enIdx < zhIdx {
		return LocaleEN
	}
	return LocaleZH
}

func findSubstr(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// GetLocaleFromContext retrieves the locale from gin context
func GetLocaleFromContext(c *gin.Context) Locale {
	if v, exists := c.Get(ContextKeyLocale); exists {
		if locale, ok := v.(Locale); ok {
			return locale
		}
	}
	return LocaleZH // Default fallback
}

// T is a convenience function for translating messages within a request context
func T(c *gin.Context, key MessageKey, args ...interface{}) string {
	locale := GetLocaleFromContext(c)
	return GetTranslator().Translate(key, locale, args...)
}

// MustGetLocale is like GetLocaleFromContext but panics if locale not set (should not happen with middleware)
func MustGetLocale(c *gin.Context) Locale {
	v, exists := c.Get(ContextKeyLocale)
	if !exists {
		panic("locale not set in context, make sure LocaleMiddleware is registered")
	}
	locale, ok := v.(Locale)
	if !ok {
		panic("locale in context is not of type Locale")
	}
	return locale
}
