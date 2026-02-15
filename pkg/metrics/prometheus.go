package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP request metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code", "user_role"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status_code"},
	)

	// Business metrics
	transactionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transactions_total",
			Help: "Total number of transactions",
		},
		[]string{"status", "product_category", "user_role"},
	)

	transactionAmount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "transaction_amount_rupiah",
			Help:    "Transaction amount in Rupiah",
			Buckets: []float64{1000, 5000, 10000, 25000, 50000, 100000, 250000, 500000, 1000000, 2500000, 5000000, 10000000},
		},
		[]string{"product_category", "user_role"},
	)

	// Database metrics
	dbConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// Redis metrics
	redisConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "redis_connections_active",
			Help: "Number of active Redis connections",
		},
	)

	redisOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_operations_total",
			Help: "Total number of Redis operations",
		},
		[]string{"operation", "status"},
	)

	// Queue metrics
	queueSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "queue_size",
			Help: "Current queue size",
		},
		[]string{"queue_name"},
	)

	queueProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "queue_processing_duration_seconds",
			Help:    "Queue processing duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"queue_name", "status"},
	)

	// Supplier adapter metrics
	supplierRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "supplier_requests_total",
			Help: "Total number of supplier requests",
		},
		[]string{"supplier", "operation", "status"},
	)

	supplierRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "supplier_request_duration_seconds",
			Help:    "Supplier request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"supplier", "operation"},
	)

	// Authentication metrics
	authAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"method", "status"},
	)

	// Application metrics
	activeUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users_total",
			Help: "Number of active users",
		},
	)

	systemErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "system_errors_total",
			Help: "Total number of system errors",
		},
		[]string{"error_type", "component"},
	)
)

// HTTP Metrics
func RecordHTTPRequest(method, endpoint, statusCode, userRole string, duration float64) {
	httpRequestsTotal.WithLabelValues(method, endpoint, statusCode, userRole).Inc()
	httpRequestDuration.WithLabelValues(method, endpoint, statusCode).Observe(duration)
}

// Transaction Metrics
func RecordTransaction(status, productCategory, userRole string, amount float64) {
	transactionsTotal.WithLabelValues(status, productCategory, userRole).Inc()
	transactionAmount.WithLabelValues(productCategory, userRole).Observe(amount)
}

// Database Metrics
func SetDBConnectionsActive(count float64) {
	dbConnectionsActive.Set(count)
}

func RecordDBQuery(operation, table string, duration float64) {
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration)
}

// Redis Metrics
func SetRedisConnectionsActive(count float64) {
	redisConnectionsActive.Set(count)
}

func RecordRedisOperation(operation, status string) {
	redisOperationsTotal.WithLabelValues(operation, status).Inc()
}

// Queue Metrics
func SetQueueSize(queueName string, size float64) {
	queueSize.WithLabelValues(queueName).Set(size)
}

func RecordQueueProcessing(queueName, status string, duration float64) {
	queueProcessingDuration.WithLabelValues(queueName, status).Observe(duration)
}

// Supplier Metrics
func RecordSupplierRequest(supplier, operation, status string, duration float64) {
	supplierRequestsTotal.WithLabelValues(supplier, operation, status).Inc()
	supplierRequestDuration.WithLabelValues(supplier, operation).Observe(duration)
}

// Authentication Metrics
func RecordAuthAttempt(method, status string) {
	authAttemptsTotal.WithLabelValues(method, status).Inc()
}

// Application Metrics
func SetActiveUsers(count float64) {
	activeUsers.Set(count)
}

func RecordSystemError(errorType, component string) {
	systemErrorsTotal.WithLabelValues(errorType, component).Inc()
}
