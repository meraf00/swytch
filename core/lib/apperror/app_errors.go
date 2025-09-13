package apperror

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ErrorType string

const (
	BadRequestError     ErrorType = "bad_request"
	NotFoundError       ErrorType = "not_found"
	ForbiddenError      ErrorType = "forbidden"
	UnauthorizedError   ErrorType = "unauthorized"
	InternalServerError ErrorType = "internal_server_error"
)

type AppError struct {
	Type    ErrorType      `json:"error"`
	Message string         `json:"message"`
	Code    string         `json:"code"`
	Meta    map[string]any `json:"meta,omitempty"`
	inner   error
}

func (e *AppError) Error() string {
	if e.inner != nil {
		return fmt.Sprintf("[%s (%s)] %s | caused by: %v", e.Type, e.Code, e.Message, e.inner)
	}
	return fmt.Sprintf("[%s (%s)] %s", e.Type, e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.inner
}

func New(errorType ErrorType, message string, code string, meta map[string]any, inner error) *AppError {
	return &AppError{
		Type:    errorType,
		Message: message,
		Code:    code,
		Meta:    meta,
		inner:   inner,
	}
}

func (e *AppError) Is(target error) bool {
	if t, ok := target.(*AppError); ok {
		return e.Type == t.Type && e.Code == t.Code
	}
	return false
}

func (e *AppError) AddMeta(key string, value any) {
	if e.Meta == nil {
		e.Meta = map[string]any{}
	}
	e.Meta[key] = value
}

// Common HTTP error helpers
func BadRequest(message, code string, meta map[string]any) *AppError {
	return New(BadRequestError, message, code, meta, nil)
}

func NotFound(message, code string, meta map[string]any) *AppError {
	return New(NotFoundError, message, code, meta, nil)
}

func Forbidden(message, code string, meta map[string]any) *AppError {
	return New(ForbiddenError, message, code, meta, nil)
}

func Unauthorized(message, code string, meta map[string]any) *AppError {
	return New(UnauthorizedError, message, code, meta, nil)
}

func InternalServer(message, code string, meta map[string]any) *AppError {
	return New(InternalServerError, message, code, meta, nil)
}

func formatFieldError(fe validator.FieldError) string {
	return fmt.Sprintf("Field validation for '%s' failed on the '%s' rule", fe.Field(), fe.Tag())
}

func ValidationError(code string, verr validator.ValidationErrors) *AppError {
	var messages []string
	for _, fieldErr := range verr {
		messages = append(messages, formatFieldError(fieldErr))
	}
	combinedMessage := strings.Join(messages, ", ")

	return New(BadRequestError, combinedMessage, code, nil, nil)
}
