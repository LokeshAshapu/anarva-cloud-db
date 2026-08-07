package logger

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const (
	TraceIDKey contextKey = "trace_id"
	UserIDKey  contextKey = "user_id"
)

var (
	globalLogger *Logger
	once         sync.Once
)

type Logger struct {
	zapLogger *zap.Logger
}

// InitLogger initializes the global structured logger.
func InitLogger(env string) (*Logger, error) {
	var config zap.Config

	if env == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	zLogger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	l := &Logger{zapLogger: zLogger}
	once.Do(func() {
		globalLogger = l
	})

	return l, nil
}

func GetLogger() *Logger {
	if globalLogger == nil {
		l, _ := InitLogger("development")
		return l
	}
	return globalLogger
}

func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{zapLogger: l.zapLogger.With(fields...)}
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	if ctx == nil {
		return l
	}

	var fields []zap.Field

	if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	if len(fields) == 0 {
		return l
	}

	return &Logger{zapLogger: l.zapLogger.With(fields...)}
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zapLogger.Debug(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zapLogger.Info(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zapLogger.Warn(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zapLogger.Error(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.zapLogger.Fatal(msg, fields...)
}

func (l *Logger) Sync() error {
	return l.zapLogger.Sync()
}

func (l *Logger) ZapLogger() *zap.Logger {
	return l.zapLogger
}

// Package-level helper functions using global logger
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

func Context(ctx context.Context) *Logger {
	return GetLogger().WithContext(ctx)
}
