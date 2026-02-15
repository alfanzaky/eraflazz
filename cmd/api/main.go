package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/alfanzaky/eraflazz/config"
	digiflazzadapter "github.com/alfanzaky/eraflazz/internal/adapter/digiflazz"
	adapterfactory "github.com/alfanzaky/eraflazz/internal/adapter/factory"
	"github.com/alfanzaky/eraflazz/internal/domain"
	apihandler "github.com/alfanzaky/eraflazz/internal/handler/api"
	"github.com/alfanzaky/eraflazz/internal/repository/postgres"
	redisrepo "github.com/alfanzaky/eraflazz/internal/repository/redis"
	"github.com/alfanzaky/eraflazz/internal/usecase"
	"github.com/alfanzaky/eraflazz/internal/worker"
	"github.com/alfanzaky/eraflazz/pkg/auth"
	"github.com/alfanzaky/eraflazz/pkg/logger"
	"github.com/alfanzaky/eraflazz/pkg/observability"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Initialize logger
	logger.Init(cfg.App.Environment)
	defer logger.Close()

	// Print configuration in development mode
	if cfg.App.IsDevelopment() {
		cfg.Print()
	}

	// Initialize database connection
	db, err := sqlx.Connect("postgres", cfg.Database.GetDSN())
	if err != nil {
		logger.Fatal("Failed to connect to database", logger.ErrorField(err))
	}
	defer db.Close()

	// Initialize Redis connection
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.GetRedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	// Test Redis connection
	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		logger.Fatal("Failed to connect to Redis", logger.ErrorField(err))
	}
	defer rdb.Close()

	logger.Info("Database and Redis connections established")

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	productRepo := postgres.NewProductRepository(db)
	supplierRepo := postgres.NewSupplierRepository(db)
	transactionRepo := postgres.NewTransactionRepository(db)
	mutationRepo := postgres.NewMutationRepository(db)
	productMappingRepo := postgres.NewProductMappingRepository(db)
	apiClientRepo := postgres.NewAPIClientRepository(db.DB)

	// Initialize smart routing
	smartRoutingUC := usecase.NewSmartRoutingUsecase(productRepo, supplierRepo, productMappingRepo)

	// Initialize product use case
	productUC := usecase.NewProductUsecase(productRepo, productMappingRepo, supplierRepo, smartRoutingUC)

	// Initialize retry use case
	retryUC := usecase.NewRetryUsecase(transactionRepo, supplierRepo, smartRoutingUC)

	// Initialize supplier adapters
	adapterFactory := adapterfactory.NewSupplierAdapterFactory()
	digiflazzAdapter := digiflazzadapter.NewAdapter(cfg.Suppliers.Digiflazz, nil)
	adapterFactory.RegisterAdapter(domain.SupplierCodeDigiflazz, digiflazzAdapter)

	// Initialize repositories that depend on Redis
	queueRepo := redisrepo.NewCacheRepository(rdb)

	// Initialize use cases
	transactionUC := usecase.NewTransactionUsecase(
		userRepo,
		productRepo,
		supplierRepo,
		transactionRepo,
		mutationRepo,
		smartRoutingUC,
		adapterFactory,
		retryUC,
		queueRepo,
	)

	// Initialize handlers
	transactionHandler := apihandler.NewTransactionHandler(transactionUC)
	productHandler := apihandler.NewProductHandler(productUC)

	// Start background transaction worker
	transactionWorker := worker.NewTransactionWorker(queueRepo, transactionUC, worker.TransactionWorkerConfig{})
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	go transactionWorker.Start(workerCtx)

	// Set Gin mode
	if cfg.App.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize auth service
	authService := auth.NewJWTAuthService(cfg.Auth)

	// Initialize metrics handler
	metricsHandler := observability.NewMetricsHandler()
	metricsHandler.RegisterMetrics()

	// Create Gin router
	router := gin.New()

	// Add middleware
	router.Use(observability.ObservabilityMiddleware())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Setup metrics and health endpoints
	router.GET("/metrics", metricsHandler.MetricsEndpoint())
	router.GET("/health", metricsHandler.HealthEndpoint())
	router.GET("/ready", metricsHandler.ReadinessEndpoint())
	router.GET("/live", metricsHandler.LivenessEndpoint())

	// Setup API routes
	apihandler.SetupRoutes(router, transactionHandler, productHandler, authService, apiClientRepo)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.API.TimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.API.TimeoutSeconds) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server",
			logger.String("port", cfg.App.Port),
			logger.String("environment", cfg.App.Environment),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", logger.ErrorField(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	workerCancel()

	logger.Info("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", logger.ErrorField(err))
	}

	logger.Info("Server exited")
}

// corsMiddleware handles CORS
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
