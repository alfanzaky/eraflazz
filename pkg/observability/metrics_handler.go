package observability

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler provides Prometheus metrics endpoint
type MetricsHandler struct {
	registry *prometheus.Registry
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler() *MetricsHandler {
	registry := prometheus.NewRegistry()
	
	// Register default Go metrics
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	
	return &MetricsHandler{
		registry: registry,
	}
}

// RegisterMetrics registers custom metrics with the registry
func (h *MetricsHandler) RegisterMetrics() {
	// Custom metrics are auto-registered via promauto
	// but we can add additional metrics here if needed
}

// MetricsEndpoint returns the Prometheus metrics handler
func (h *MetricsHandler) MetricsEndpoint() gin.HandlerFunc {
	handler := promhttp.HandlerFor(h.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
	
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

// HealthEndpoint provides health check with metrics
func (h *MetricsHandler) HealthEndpoint() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Basic health check
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "eraflazz-api",
			"timestamp": gin.H{},
		})
	}
}

// ReadinessEndpoint provides readiness check
func (h *MetricsHandler) ReadinessEndpoint() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check dependencies here (database, redis, etc.)
		// For now, just return ready
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	}
}

// LivenessEndpoint provides liveness check
func (h *MetricsHandler) LivenessEndpoint() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Basic liveness check
		c.JSON(http.StatusOK, gin.H{
			"status": "alive",
		})
	}
}
