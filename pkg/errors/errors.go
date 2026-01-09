package errors

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/middleware"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
)

// AppError 统一应用错误结构体
// 包含错误码、消息、详情、追踪ID和时间戳
type AppError struct {
	// Code 错误码
	Code int `json:"code"`
	// Message 错误消息
	Message string `json:"message"`
	// Details 错误详情（可选）
	Details []string `json:"details,omitempty"`
	// TraceID 请求追踪ID
	TraceID string `json:"traceId,omitempty"`
	// Cause 原始错误（不序列化到JSON）
	Cause error `json:"-"`
	// Timestamp 错误发生时间
	Timestamp time.Time `json:"timestamp"`
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	return e.Message
}

// Unwrap 实现 errors.Unwrap 接口，支持错误链路追踪
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewAppError 从 Code 对象创建 AppError
func NewAppError(c *code.Code, cause error) *AppError {
	return &AppError{
		Code:      c.Code(),
		Message:   c.Msg(),
		Details:   c.Details(),
		Cause:     cause,
		Timestamp: time.Now(),
	}
}

// NewAppErrorWithMessage 创建带自定义消息的 AppError
func NewAppErrorWithMessage(errorCode int, message string, cause error) *AppError {
	return &AppError{
		Code:      errorCode,
		Message:   message,
		Cause:     cause,
		Timestamp: time.Now(),
	}
}

// WithTraceID 设置 TraceID 并返回自身（链式调用）
func (e *AppError) WithTraceID(traceID string) *AppError {
	e.TraceID = traceID
	return e
}

// WithDetails 设置详情并返回自身（链式调用）
func (e *AppError) WithDetails(details ...string) *AppError {
	e.Details = details
	return e
}

// ErrorResponse 统一错误响应处理
// 从 gin.Context 获取 TraceID，将错误转换为 AppError 并返回 JSON 响应
func ErrorResponse(c *gin.Context, err error) {
	traceID := middleware.GetTraceIDFromGin(c)

	var appErr *AppError
	if errors.As(err, &appErr) {
		// 已经是 AppError，设置 TraceID
		appErr.TraceID = traceID
		c.JSON(http.StatusOK, appErr)
		return
	}

	// 检查是否是 Code 类型错误
	var codeErr *code.Code
	if errors.As(err, &codeErr) {
		response := &AppError{
			Code:      codeErr.Code(),
			Message:   codeErr.Msg(),
			Details:   codeErr.Details(),
			TraceID:   traceID,
			Timestamp: time.Now(),
		}
		c.JSON(http.StatusOK, response)
		return
	}

	// 未知错误，返回内部错误
	c.JSON(http.StatusOK, &AppError{
		Code:      500,
		Message:   "Internal Server Error",
		TraceID:   traceID,
		Timestamp: time.Now(),
	})
}

// ErrorResponseWithCode 使用指定的 Code 对象返回错误响应
func ErrorResponseWithCode(c *gin.Context, codeErr *code.Code, cause error) {
	traceID := middleware.GetTraceIDFromGin(c)

	response := &AppError{
		Code:      codeErr.Code(),
		Message:   codeErr.Msg(),
		Details:   codeErr.Details(),
		TraceID:   traceID,
		Cause:     cause,
		Timestamp: time.Now(),
	}
	c.JSON(http.StatusOK, response)
}

// IsAppError 检查错误是否为 AppError 类型
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError 从错误链中获取 AppError
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}
