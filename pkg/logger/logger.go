// pkg/logger/logger.go
package pkg_logger

import (
	"mikrobill/pkg/filelog"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log        *zap.Logger
	currentEnv string
)

func InitLogger(environment string) error {
	currentEnv = environment
	var config zap.Config

	if environment == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	logger, err := config.Build()
	if err != nil {
		return err
	}

	Log = logger

	// Also initialize file logger
	// We ignore error here as it might be already initialized or we want to proceed anyway
	_ = filelog.Init()

	return nil
}

func zapFieldsToMap(fields []zap.Field) map[string]interface{} {
	enc := zapcore.NewMapObjectEncoder()
	for _, f := range fields {
		f.AddTo(enc)
	}
	return enc.Fields
}

func Info(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Info(msg, fields...)
	}

	filelog.Info(msg, zapFieldsToMap(fields))
}

func Error(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Error(msg, fields...)
	}
	// Errors are always important, log to file regardless of environment
	filelog.Error(msg, zapFieldsToMap(fields))
}

func Debug(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Debug(msg, fields...)
	}
	// Debug logs never go to file in production
	if currentEnv != "production" {
		filelog.Debug(msg, zapFieldsToMap(fields))
	}
}

func Warn(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Warn(msg, fields...)
	}
	// Warnings are always important, log to file regardless of environment
	filelog.Warn(msg, zapFieldsToMap(fields))
}

func Fatal(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Fatal(msg, fields...)
	}
	// Fatal logs are always critical, log to file regardless of environment
	filelog.Fatal(msg, zapFieldsToMap(fields))
}

func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
