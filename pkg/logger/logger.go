package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

// Init initializes the global logger with the specified level and format
func Init(level string, format string) error {
	// Parse log level
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		return fmt.Errorf("invalid log level %q: %w", level, err)
	}

	// Create config based on format
	var config zap.Config
	if format == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	// Set log level
	config.Level = zap.NewAtomicLevelAt(zapLevel)

	// Set output to stdout
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	// Build logger
	logger, err := config.Build()
	if err != nil {
		return fmt.Errorf("failed to build logger: %w", err)
	}

	log = logger
	return nil
}

// Info logs an info-level message
func Info(msg string, fields ...zap.Field) {
	if log != nil {
		log.Info(msg, fields...)
	}
}

// Warn logs a warning-level message
func Warn(msg string, fields ...zap.Field) {
	if log != nil {
		log.Warn(msg, fields...)
	}
}

// Error logs an error-level message
func Error(msg string, fields ...zap.Field) {
	if log != nil {
		log.Error(msg, fields...)
	}
}

// Debug logs a debug-level message
func Debug(msg string, fields ...zap.Field) {
	if log != nil {
		log.Debug(msg, fields...)
	}
}

// Fatal logs a fatal-level message and exits the application
func Fatal(msg string, fields ...zap.Field) {
	if log != nil {
		log.Fatal(msg, fields...)
	}
}

// WithRequestID creates a new logger with the request ID field
func WithRequestID(requestID string) *zap.Logger {
	if log == nil {
		return nil
	}
	return log.With(zap.String("request_id", requestID))
}

// Sync flushes any buffered log entries
// Should be called before application shutdown
func Sync() error {
	if log != nil {
		return log.Sync()
	}
	return nil
}

// GetLogger returns the underlying zap logger instance
// Useful for advanced use cases
func GetLogger() *zap.Logger {
	return log
}
