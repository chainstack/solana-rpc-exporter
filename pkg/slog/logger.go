package slog

import (
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

// Init initializes the logger with custom configuration
func Init() {
	config := zap.NewProductionConfig()

	// Configure encoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	config.EncoderConfig.StacktraceKey = "" // Disable stacktrace
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// Set log level from environment
	config.Level = zap.NewAtomicLevelAt(getEnvLogLevel())

	// Add development mode settings if enabled
	if isDevelopment() {
		config.Development = true
		config.Encoding = "console"
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Build logger
	logger, err := config.Build()
	if err != nil {
		panic(fmt.Errorf("error initializing logger: %v", err))
	}
	
	log = logger.Sugar()
}

// Get returns the global logger instance
func Get() *zap.SugaredLogger {
	return log
}

// Sync flushes any buffered log entries
func Sync() error {
	return log.Sync()
}

// getEnvLogLevel gets the log level from environment variable
func getEnvLogLevel() zapcore.Level {
	level, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		return zapcore.InfoLevel
	}

	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		fmt.Printf("Unrecognized LOG_LEVEL '%s', using 'info'\n", level)
		return zapcore.InfoLevel
	}
}

// isDevelopment checks if we're in development mode
func isDevelopment() bool {
	return strings.ToLower(os.Getenv("ENV")) == "development"
}

// Helper functions for common logging patterns
func WithError(err error) zap.Field {
	return zap.Error(err)
}

func WithField(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}
