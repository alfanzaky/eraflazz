package logger

import (
	"sync"
	"time"
	
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	instance *zap.Logger
	once     sync.Once
)

// Init initializes the logger with the specified environment
func Init(env string) {
	once.Do(func() {
		var config zap.Config
		
		switch env {
		case "production":
			config = zap.NewProductionConfig()
			config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		case "development":
			config = zap.NewDevelopmentConfig()
			config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
			config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		default:
			config = zap.NewDevelopmentConfig()
			config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		}
		
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stderr"}
		
		var err error
		instance, err = config.Build()
		if err != nil {
			panic(err)
		}
	})
}

// GetLogger returns the logger instance
func GetLogger() *zap.Logger {
	if instance == nil {
		// Default to development if not initialized
		Init("development")
	}
	return instance
}

// Sugar returns the sugared logger for easier usage
func Sugar() *zap.SugaredLogger {
	return GetLogger().Sugar()
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// WithFields creates a logger with additional fields
func WithFields(fields ...zap.Field) *zap.Logger {
	return GetLogger().With(fields...)
}

// Sync flushes any buffered log entries
func Sync() {
	if instance != nil {
		_ = instance.Sync()
	}
}

// Close closes the logger
func Close() {
	Sync()
}

// Custom field constructors
func String(key, value string) zap.Field {
	return zap.String(key, value)
}

func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

func Int64(key string, value int64) zap.Field {
	return zap.Int64(key, value)
}

func Float64(key string, value float64) zap.Field {
	return zap.Float64(key, value)
}

func Bool(key string, value bool) zap.Field {
	return zap.Bool(key, value)
}

func Any(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

func ErrorField(err error) zap.Field {
	return zap.Error(err)
}

func Duration(key string, value interface{}) zap.Field {
	return zap.Duration(key, value.(time.Duration))
}
