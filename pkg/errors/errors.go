package errors

import (
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
)

// ErrorCode represents a domain error classification type.
type ErrorCode string

const (
	CodeNotFound           ErrorCode = "NOT_FOUND"
	CodeAlreadyExists      ErrorCode = "ALREADY_EXISTS"
	CodeInvalidInput       ErrorCode = "INVALID_INPUT"
	CodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	CodeForbidden          ErrorCode = "FORBIDDEN"
	CodeInternal           ErrorCode = "INTERNAL_ERROR"
	CodeConflict           ErrorCode = "CONFLICT"
	CodeTimeout            ErrorCode = "TIMEOUT"
	CodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	CodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	CodeQuotaExceeded      ErrorCode = "QUOTA_EXCEEDED"

	// Normalized Operational Error Intelligence Categories (Phase 46)
	CodeAuthenticationError ErrorCode = "AUTHENTICATION_ERROR"
	CodeAuthorizationError  ErrorCode = "AUTHORIZATION_ERROR"
	CodeValidationError     ErrorCode = "VALIDATION_ERROR"
	CodeConflictError       ErrorCode = "CONFLICT"
	CodeDependencyFailure   ErrorCode = "DEPENDENCY_FAILURE"
	CodeProviderFailure     ErrorCode = "PROVIDER_FAILURE"
	CodeTimeoutError        ErrorCode = "TIMEOUT"
	CodeDatabaseFailure     ErrorCode = "DATABASE_FAILURE"
	CodeInternalError       ErrorCode = "INTERNAL_ERROR"
)

// AppError is the standard error struct for the entire platform.
type AppError struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Err     error                  `json:"-"`
	Details map[string]interface{} `json:"details,omitempty"`
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

// New creates a new AppError without an underlying error.
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error into an AppError with a domain code.
func Wrap(err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return nil
	}
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
		Details: make(map[string]interface{}),
	}
}

// WithDetail adds structured context details to the error.
func (e *AppError) WithDetail(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// HTTPStatusCode maps domain ErrorCode to standard HTTP status codes.
func (e *AppError) HTTPStatusCode() int {
	switch e.Code {
	case CodeNotFound:
		return http.StatusNotFound
	case CodeAlreadyExists, CodeConflict:
		return http.StatusConflict
	case CodeInvalidInput:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeTimeout:
		return http.StatusGatewayTimeout
	case CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// GRPCCode maps domain ErrorCode to gRPC status codes.
func (e *AppError) GRPCCode() codes.Code {
	switch e.Code {
	case CodeNotFound:
		return codes.NotFound
	case CodeAlreadyExists, CodeConflict:
		return codes.AlreadyExists
	case CodeInvalidInput:
		return codes.InvalidArgument
	case CodeUnauthorized:
		return codes.Unauthenticated
	case CodeForbidden:
		return codes.PermissionDenied
	case CodeTimeout:
		return codes.DeadlineExceeded
	case CodeServiceUnavailable:
		return codes.Unavailable
	default:
		return codes.Internal
	}
}
