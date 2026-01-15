package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// Initialize sets up the global Zap logger
func Initialize(environment string) error {
	var config zap.Config

	if environment == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Always use JSON encoder for production
	if environment == "production" {
		config.Encoding = "json"
	}

	logger, err := config.Build(
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return err
	}

	Log = logger
	return nil
}

// Sync flushes any buffered log entries
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	if Log == nil {
		// Fallback to development logger if not initialized
		Log, _ = zap.NewDevelopment()
	}
	return Log
}

// WithRequestID creates a logger with request ID field
func WithRequestID(requestID string) *zap.Logger {
	return GetLogger().With(zap.String("request_id", requestID))
}

// WithTenantID creates a logger with tenant ID field
func WithTenantID(tenantID string) *zap.Logger {
	return GetLogger().With(zap.String("tenant_id", tenantID))
}

// WithUserID creates a logger with user ID field
func WithUserID(userID string) *zap.Logger {
	return GetLogger().With(zap.String("user_id", userID))
}

// WithContext creates a logger with multiple context fields
func WithContext(requestID, tenantID, userID string) *zap.Logger {
	logger := GetLogger()
	if requestID != "" {
		logger = logger.With(zap.String("request_id", requestID))
	}
	if tenantID != "" {
		logger = logger.With(zap.String("tenant_id", tenantID))
	}
	if userID != "" {
		logger = logger.With(zap.String("user_id", userID))
	}
	return logger
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
	os.Exit(1)
}

// Panic logs a panic message and panics
func Panic(msg string, fields ...zap.Field) {
	GetLogger().Panic(msg, fields...)
}
