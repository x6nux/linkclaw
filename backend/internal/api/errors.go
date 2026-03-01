package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
)

// ErrorCode 统一的错误码
type ErrorCode string

const (
	// 通用错误
	ErrUnknown       ErrorCode = "UNKNOWN"
	ErrInvalidParam  ErrorCode = "INVALID_PARAM"
	ErrNotFound      ErrorCode = "NOT_FOUND"
	ErrUnauthorized  ErrorCode = "UNAUTHORIZED"
	ErrForbidden     ErrorCode = "FORBIDDEN"
	ErrInternal      ErrorCode = "INTERNAL"
	ErrTimeout       ErrorCode = "TIMEOUT"

	// 业务错误
	ErrConflict      ErrorCode = "CONFLICT"
	ErrPrecondition  ErrorCode = "PRECONDITION_FAILED"
	ErrValidation    ErrorCode = "VALIDATION_ERROR"
)

// APIError 统一的 API 错误结构
type APIError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"` // 可选的详细信息
	StatusCode int       `json:"-"`                 // HTTP 状态码（不序列化到 JSON）
	Err        error     `json:"-"`                 // 原始错误（不序列化）
}

// Error 实现 error 接口
func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 实现 errors.Unwrap 接口
func (e *APIError) Unwrap() error {
	return e.Err
}

// ErrorResponse 统一的错误响应格式
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    ErrorCode              `json:"code"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// APIResponse 统一的成功响应格式（可选使用）
type APIResponse[T any] struct {
	Data  T       `json:"data,omitempty"`
	Error string  `json:"error,omitempty"`
	Code  ErrorCode `json:"code,omitempty"`
}

// ── 快捷错误构造函数 ─────────────────────────────────────────

// NewAPIError 创建自定义 APIError
func NewAPIError(code ErrorCode, message string, statusCode int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// WrapError 包装现有 error 为 APIError
func WrapError(err error, code ErrorCode, message string, statusCode int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// NotFoundError 资源不存在
func NotFoundError(resource string) *APIError {
	return NewAPIError(ErrNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

// InvalidParamError 参数无效
func InvalidParamError(message string) *APIError {
	return NewAPIError(ErrInvalidParam, message, http.StatusBadRequest)
}

// ValidationError 参数验证失败
func ValidationError(message string) *APIError {
	return NewAPIError(ErrValidation, message, http.StatusBadRequest)
}

// UnauthorizedError 未授权
func UnauthorizedError(message string) *APIError {
	return NewAPIError(ErrUnauthorized, message, http.StatusUnauthorized)
}

// ForbiddenError 禁止访问
func ForbiddenError(message string) *APIError {
	return NewAPIError(ErrForbidden, message, http.StatusForbidden)
}

// InternalError 内部错误
func InternalError(message string) *APIError {
	return NewAPIError(ErrInternal, message, http.StatusInternalServerError)
}

// ConflictError 资源冲突
func ConflictError(message string) *APIError {
	return NewAPIError(ErrConflict, message, http.StatusConflict)
}

// TimeoutError 请求超时
func TimeoutError(message string) *APIError {
	return NewAPIError(ErrTimeout, message, http.StatusGatewayTimeout)
}

// ── JSON 响应辅助函数 ────────────────────────────────────────

// RespondError 发送错误响应（简化版，保持向后兼容）
// 如果 err 是 *APIError，使用其信息；否则使用提供的默认值
func RespondError(c *gin.Context, defaultCode ErrorCode, defaultMessage string, defaultStatus int, err error) {
	if apiErr, ok := err.(*APIError); ok {
		c.JSON(apiErr.StatusCode, ErrorResponse{
			Error: apiErr.Message,
			Code:  apiErr.Code,
		})
		return
	}

	// 非 APIError，使用默认值
	c.JSON(defaultStatus, ErrorResponse{
		Error: defaultMessage,
		Code:  defaultCode,
	})
}

// RespondSuccess 发送成功响应
func RespondSuccess[T any](c *gin.Context, data T) {
	c.JSON(http.StatusOK, APIResponse[T]{Data: data})
}

// RespondCreated 发送创建成功响应
func RespondCreated[T any](c *gin.Context, data T) {
	c.JSON(http.StatusCreated, APIResponse[T]{Data: data})
}

// RespondNoContent 发送 204 No Content
func RespondNoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// ── 全局错误处理中间件 ──────────────────────────────────────

// RecoveryMiddleware 捕获 panic 并返回统一的 500 错误响应
// 替代默认的 gin.Recovery()，提供更友好的错误格式
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 避免打印到标准输出，使用 log 包记录
				stack := make([]byte, 4096)
				length := runtime.Stack(stack, true)
				log.Printf("[PANIC RECOVERED] %v\n\n%s\n", err, stack[:length])

				c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{
					Error: "internal server error",
					Code:  ErrInternal,
				})
			}
		}()
		c.Next()
	}
}

// ErrorToResponse 将 error 转换为统一的错误响应并发送
// 用于 handler 中简化错误处理
func ErrorToResponse(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// 检查是否是 *APIError
	if apiErr, ok := err.(*APIError); ok {
		c.AbortWithStatusJSON(apiErr.StatusCode, ErrorResponse{
			Error: apiErr.Message,
			Code:  apiErr.Code,
		})
		return
	}

	// 检查是否是已知的错误类型
	switch {
	case errors.Is(err, &APIError{Code: ErrNotFound, StatusCode: http.StatusNotFound}):
		c.AbortWithStatusJSON(http.StatusNotFound, ErrorResponse{
			Error: "not found",
			Code:  ErrNotFound,
		})
	default:
		// 默认 500
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  ErrUnknown,
		})
	}
}

// ── 错误断言辅助 ────────────────────────────────────────────

// AssertNotFound 如果 err 是 "not found" 类型，发送 404 响应并返回 true
func AssertNotFound(c *gin.Context, err error) bool {
	if err != nil {
		ErrorToResponse(c, NotFoundError("resource"))
		return true
	}
	return false
}

// AssertError 如果 err 非 nil，发送响应并返回 true
func AssertError(c *gin.Context, err error) bool {
	if err != nil {
		ErrorToResponse(c, err)
		return true
	}
	return false
}

// ── 字符串工具 ──────────────────────────────────────────────

// truncateError 截断过长的错误信息，防止泄露过多内部细节
func truncateError(msg string, maxLen int) string {
	if len(msg) <= maxLen {
		return msg
	}
	return msg[:maxLen] + "..."
}

// sanitizeError 清理错误信息中的敏感内容
func sanitizeError(err string) string {
	// 移除可能的文件路径
	parts := strings.Split(err, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return err
}
