package errors

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestAppError_FormattingAndUnwrap(t *testing.T) {
	baseErr := fmt.Errorf("connection refused")
	appErr := Wrap(baseErr, CodeDatabaseError, "Failed to connect to database")

	assert.Equal(t, "[DATABASE_ERROR] Failed to connect to database: connection refused", appErr.Error())
	assert.Equal(t, baseErr, appErr.Unwrap())

	simpleErr := New(CodeNotFound, "User not found")
	assert.Equal(t, "[NOT_FOUND] User not found", simpleErr.Error())
	assert.Nil(t, simpleErr.Unwrap())
}

func TestAppError_WithDetail(t *testing.T) {
	err := New(CodeInvalidInput, "Validation failed").
		WithDetail("field", "email").
		WithDetail("reason", "invalid format")

	assert.Equal(t, "email", err.Details["field"])
	assert.Equal(t, "invalid format", err.Details["reason"])
}

func TestAppError_HTTPStatusCode(t *testing.T) {
	tests := []struct {
		code ErrorCode
		want int
	}{
		{CodeNotFound, http.StatusNotFound},
		{CodeAlreadyExists, http.StatusConflict},
		{CodeInvalidInput, http.StatusBadRequest},
		{CodeUnauthorized, http.StatusUnauthorized},
		{CodeForbidden, http.StatusForbidden},
		{CodeTimeout, http.StatusGatewayTimeout},
		{CodeServiceUnavailable, http.StatusServiceUnavailable},
		{CodeInternal, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		err := New(tt.code, "test")
		assert.Equal(t, tt.want, err.HTTPStatusCode())
	}
}

func TestAppError_GRPCCode(t *testing.T) {
	tests := []struct {
		code ErrorCode
		want codes.Code
	}{
		{CodeNotFound, codes.NotFound},
		{CodeAlreadyExists, codes.AlreadyExists},
		{CodeInvalidInput, codes.InvalidArgument},
		{CodeUnauthorized, codes.Unauthenticated},
		{CodeForbidden, codes.PermissionDenied},
		{CodeTimeout, codes.DeadlineExceeded},
		{CodeServiceUnavailable, codes.Unavailable},
		{CodeInternal, codes.Internal},
	}

	for _, tt := range tests {
		err := New(tt.code, "test")
		assert.Equal(t, tt.want, err.GRPCCode())
	}
}
