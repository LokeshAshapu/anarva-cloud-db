package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInitLogger(t *testing.T) {
	log, err := InitLogger("development")
	require.NoError(t, err)
	assert.NotNil(t, log)

	logProd, err := InitLogger("production")
	require.NoError(t, err)
	assert.NotNil(t, logProd)
}

func TestLogger_WithContext(t *testing.T) {
	log, err := InitLogger("development")
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), TraceIDKey, "trace-12345")
	ctx = context.WithValue(ctx, UserIDKey, "user-67890")

	ctxLogger := log.WithContext(ctx)
	assert.NotNil(t, ctxLogger)

	// Test logging execution without panic
	ctxLogger.Info("Test context logging", zap.String("action", "test"))
}
