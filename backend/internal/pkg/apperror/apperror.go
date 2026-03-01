package apperror

import (
	"fmt"
	"net/http"
)

// Code represents a standard application error code.
type Code string

const (
	CodeNotFound         Code = "NOT_FOUND"
	CodeForbidden        Code = "FORBIDDEN"
	CodeUnauthorized     Code = "UNAUTHORIZED"
	CodeValidationError  Code = "VALIDATION_ERROR"
	CodeConflict         Code = "CONFLICT"
	CodeInternal         Code = "INTERNAL_ERROR"
	CodeBadRequest       Code = "BAD_REQUEST"
	CodeRateLimited      Code = "RATE_LIMITED"
	CodeUnprocessable    Code = "UNPROCESSABLE_ENTITY"
	CodeUnsupportedMedia Code = "UNSUPPORTED_MEDIA_TYPE"
)

// AppError is a structured application error with a code, message, and optional details.
type AppError struct {
	Code    Code                   `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Status  int                    `json:"-"`
	Err     error                  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// HTTPStatus returns the appropriate HTTP status code.
func (e *AppError) HTTPStatus() int {
	if e.Status != 0 {
		return e.Status
	}
	return codeToStatus(e.Code)
}

// ErrorResponse is the JSON response structure for errors.
type ErrorResponse struct {
	Error *AppError `json:"error"`
}

// New creates a new AppError with the given code and message.
func New(code Code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  codeToStatus(code),
	}
}

// Wrap creates a new AppError wrapping an existing error.
func Wrap(code Code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  codeToStatus(code),
		Err:     err,
	}
}

// WithDetails adds details to the error.
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}

// Convenience constructors

func NotFound(resource string) *AppError {
	return New(CodeNotFound, fmt.Sprintf("%s not found", resource))
}

func Forbidden(message string) *AppError {
	return New(CodeForbidden, message)
}

func Unauthorized(message string) *AppError {
	return New(CodeUnauthorized, message)
}

func ValidationError(message string, details map[string]interface{}) *AppError {
	return New(CodeValidationError, message).WithDetails(details)
}

func Conflict(message string) *AppError {
	return New(CodeConflict, message)
}

func Internal(message string, err error) *AppError {
	return Wrap(CodeInternal, message, err)
}

func BadRequest(message string) *AppError {
	return New(CodeBadRequest, message)
}

func RateLimited() *AppError {
	return New(CodeRateLimited, "rate limit exceeded, please try again later")
}

func UnprocessableEntity(message string) *AppError {
	return New(CodeUnprocessable, message)
}

func codeToStatus(code Code) int {
	switch code {
	case CodeNotFound:
		return http.StatusNotFound
	case CodeForbidden:
		return http.StatusForbidden
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeValidationError:
		return http.StatusBadRequest
	case CodeConflict:
		return http.StatusConflict
	case CodeBadRequest:
		return http.StatusBadRequest
	case CodeRateLimited:
		return http.StatusTooManyRequests
	case CodeUnprocessable:
		return http.StatusUnprocessableEntity
	case CodeUnsupportedMedia:
		return http.StatusUnsupportedMediaType
	case CodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
