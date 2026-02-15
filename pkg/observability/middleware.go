package observability

import (
	"context"
	"strconv"
	"time"

	"github.com/alfanzaky/eraflazz/pkg/logger"
	"github.com/alfanzaky/eraflazz/pkg/metrics"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TraceIDKey is the context key for trace ID
type TraceIDKey string

const (
	// TraceIDHeader is the HTTP header for trace ID
	TraceIDHeader = "X-Trace-ID"
	// TraceIDContextKey is the context key for trace ID
	TraceIDContextKey TraceIDKey = "trace_id"
)

// ObservabilityMiddleware provides trace ID generation and metrics collection
func ObservabilityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Generate or extract trace ID
		traceID := c.GetHeader(TraceIDHeader)
		if traceID == "" {
			traceID = generateTraceID()
		}

		// Set trace ID in response header and context
		c.Header(TraceIDHeader, traceID)
		c.Set(string(TraceIDContextKey), traceID)

		// Add trace ID to logger context
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), TraceIDContextKey, traceID))

		// Get user role from context if available
		userRole := "anonymous"
		if role, exists := c.Get("user_role"); exists {
			if roleStr, ok := role.(string); ok {
				userRole = roleStr
			}
		} else if clientID, exists := c.Get("client_id"); exists {
			if clientStr, ok := clientID.(string); ok {
				userRole = "h2h_" + clientStr
			}
		}

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())

		metrics.RecordHTTPRequest(
			c.Request.Method,
			c.FullPath(),
			statusCode,
			userRole,
			duration,
		)

		// Log request completion with trace ID
		logger.Info("Request completed",
			logger.String("trace_id", traceID),
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.String("status", statusCode),
			logger.Float64("duration_ms", duration*1000),
			logger.String("user_role", userRole),
			logger.String("client_ip", c.ClientIP()),
		)
	}
}

// generateTraceID generates a new trace ID
func generateTraceID() string {
	return uuid.New().String()
}

// GetTraceID extracts trace ID from context
func GetTraceID(c *gin.Context) string {
	if traceID, exists := c.Get(string(TraceIDContextKey)); exists {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

// GetTraceIDFromContext extracts trace ID from context.Context
func GetTraceIDFromContext(ctx context.Context) string {
	if traceID := ctx.Value(TraceIDContextKey); traceID != nil {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

// WithTraceID adds trace ID to context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDContextKey, traceID)
}

// LogWithError logs error with trace ID
func LogWithError(c *gin.Context, err error, message string) {
	traceID := GetTraceID(c)
	logger.Error(message,
		logger.String("trace_id", traceID),
	)
}

// LogWithFields logs with trace ID and custom fields
func LogWithFields(c *gin.Context, message string, fields ...zap.Field) {
	traceID := GetTraceID(c)
	allFields := append([]zap.Field{
		zap.String("trace_id", traceID),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("client_ip", c.ClientIP()),
	}, fields...)

	logger.Info(message, allFields...)
}

// RecordSystemError records system error with metrics and logging
func RecordSystemError(c *gin.Context, errorType, component string, err error) {
	traceID := GetTraceID(c)

	// Record metrics
	metrics.RecordSystemError(errorType, component)

	// Log with trace ID
	logger.Error("System error occurred",
		logger.String("trace_id", traceID),
		logger.String("error_type", errorType),
		logger.String("component", component),
		logger.ErrorField(err),
		logger.String("method", c.Request.Method),
		logger.String("path", c.Request.URL.Path),
		logger.String("client_ip", c.ClientIP()),
	)
}
