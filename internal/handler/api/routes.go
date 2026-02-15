package api

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/alfanzaky/eraflazz/internal/domain"
	"github.com/alfanzaky/eraflazz/internal/repository/postgres"
	authpkg "github.com/alfanzaky/eraflazz/pkg/auth"
	"github.com/alfanzaky/eraflazz/pkg/logger"
	"github.com/alfanzaky/eraflazz/pkg/xresponse"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(
	router *gin.Engine,
	transactionHandler *TransactionHandler,
	productHandler *ProductHandler,
	authService domain.AuthService,
	clientRepo *postgres.APIClientRepository,
) {
	router.GET("/health", func(c *gin.Context) {
		xresponse.Success(c, "Service is healthy", gin.H{
			"service": "eraflazz-api",
			"status":  "ok",
		})
	})

	v1 := router.Group("/api/v1")
	{
		configureTransactionRoutes(v1, transactionHandler, authService)
		configureAdminProductRoutes(v1, productHandler, authService)
		configureH2HRoutes(v1, clientRepo)
		configurePublicRoutes(v1)
	}

	logger.Info("API routes configured successfully")
}

// h2hMiddleware secures H2H routes using API key + signature verification
func h2hMiddleware(authService domain.AuthService, allowedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if authService == nil {
			xresponse.InternalServerError(c, "Auth service not available")
			c.Abort()
			return
		}

		// IP Whitelist validation
		clientIP := net.ParseIP(c.ClientIP())
		if len(allowedIPs) > 0 && (clientIP == nil || !isIPAllowed(clientIP, allowedIPs)) {
			logger.Warn("H2H access denied - IP not allowed",
				logger.String("client_ip", c.ClientIP()),
				logger.Any("allowed_ips", allowedIPs),
			)
			xresponse.Forbidden(c, "IP address not allowed")
			c.Abort()
			return
		}

		// Extract H2H authentication headers
		apiKey := strings.TrimSpace(c.GetHeader("X-API-Key"))
		signature := strings.TrimSpace(c.GetHeader("X-Signature"))
		timestamp := strings.TrimSpace(c.GetHeader("X-Timestamp"))

		// Validate required headers
		if apiKey == "" || signature == "" || timestamp == "" {
			logger.Warn("H2H authentication failed - missing headers",
				logger.String("client_ip", c.ClientIP()),
				logger.Bool("has_api_key", apiKey != ""),
				logger.Bool("has_signature", signature != ""),
				logger.Bool("has_timestamp", timestamp != ""),
			)
			xresponse.Unauthorized(c, "Missing H2H authentication headers")
			c.Abort()
			return
		}

		// Validate timestamp format and age
		reqTime, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			logger.Warn("H2H authentication failed - invalid timestamp format",
				logger.String("client_ip", c.ClientIP()),
				logger.String("timestamp", timestamp),
				logger.String("error", err.Error()),
			)
			xresponse.BadRequest(c, "Invalid timestamp format")
			c.Abort()
			return
		}

		// Check timestamp age (prevent replay attacks)
		if time.Since(reqTime) > 5*time.Minute {
			logger.Warn("H2H authentication failed - timestamp too old",
				logger.String("client_ip", c.ClientIP()),
				logger.String("timestamp", timestamp),
				logger.String("req_time", reqTime.Format(time.RFC3339)),
			)
			xresponse.Unauthorized(c, "Request timestamp too old")
			c.Abort()
			return
		}

		// Read request body for signature validation
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Error("H2H authentication failed - failed to read request body",
				logger.String("client_ip", c.ClientIP()),
				logger.String("error", err.Error()),
			)
			xresponse.InternalServerError(c, "Failed to read request body")
			c.Abort()
			return
		}
		// Restore request body for subsequent handlers
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Validate H2H signature
		if err := authService.ValidateH2HSignature(apiKey, signature, timestamp, bodyBytes); err != nil {
			logger.Warn("H2H authentication failed - invalid signature",
				logger.String("client_ip", c.ClientIP()),
				logger.String("api_key", apiKey),
				logger.String("error", err.Error()),
			)
			xresponse.Unauthorized(c, "Invalid H2H signature")
			c.Abort()
			return
		}

		// Set API key in context for handlers
		c.Set("h2h_api_key", apiKey)

		logger.Info("H2H authentication successful",
			logger.String("client_ip", c.ClientIP()),
			logger.String("api_key", apiKey),
			logger.String("endpoint", c.Request.URL.Path),
		)

		c.Next()
	}
}

func isIPAllowed(ip net.IP, allowed []string) bool {
	for _, entry := range allowed {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		if strings.Contains(entry, "/") {
			_, cidr, err := net.ParseCIDR(entry)
			if err == nil && cidr.Contains(ip) {
				return true
			}
			continue
		}

		if ip.Equal(net.ParseIP(entry)) {
			return true
		}
	}

	return false
}

func configureTransactionRoutes(group *gin.RouterGroup, transactionHandler *TransactionHandler, authService domain.AuthService) {
	routes := group.Group("/transactions")
	routes.Use(authMiddleware(authService))
	{
		routes.POST("", transactionHandler.CreateTransaction)
		routes.GET("/:id", transactionHandler.GetTransaction)
		routes.GET("/code/:code", transactionHandler.GetTransactionByCode)
		routes.GET("/user", transactionHandler.GetUserTransactions)
		routes.DELETE("/:id", transactionHandler.CancelTransaction)
		routes.GET("/stats", transactionHandler.GetTransactionStats)
	}
}

func configureAdminProductRoutes(group *gin.RouterGroup, productHandler *ProductHandler, authService domain.AuthService) {
	adminRoutes := group.Group("/admin")
	adminRoutes.Use(authMiddleware(authService), adminMiddleware())
	{
		products := adminRoutes.Group("/products")
		{
			products.POST("", productHandler.CreateProduct)
			products.GET("", productHandler.ListProducts)
			products.GET("/:id", productHandler.GetProduct)
			products.PUT("/:id", productHandler.UpdateProduct)
			products.PATCH("/:id/status", productHandler.ToggleProductStatus)
			products.PATCH("/:id/stock", productHandler.UpdateProductStock)
			products.GET("/:id/mappings", productHandler.ListProductMappings)
			products.POST("/:id/mappings", productHandler.CreateProductMapping)
		}

		mappings := adminRoutes.Group("/product-mappings")
		{
			mappings.PUT("/:id", productHandler.UpdateProductMapping)
			mappings.DELETE("/:id", productHandler.DeleteProductMapping)
		}
	}
}

func configureH2HRoutes(group *gin.RouterGroup, clientRepo *postgres.APIClientRepository) {
	h2hMiddleware := NewH2HMiddleware(clientRepo)
	h2hRoutes := group.Group("/h2h")
	h2hRoutes.Use(h2hMiddleware.H2HAuth())
	{
		// H2H callback endpoint for supplier notifications
		h2hRoutes.POST("/callback", func(c *gin.Context) {
			clientID, exists := GetClientIDFromContext(c)
			if !exists {
				xresponse.Unauthorized(c, "Client not authenticated")
				return
			}

			logger.Info("H2H callback received",
				logger.String("client_id", clientID),
				logger.String("client_ip", c.ClientIP()),
			)
			xresponse.Success(c, "H2H callback accepted", gin.H{
				"status":    "ok",
				"client_id": clientID,
			})
		})

		// TODO: Add H2H inquiry endpoint when ready
		// h2hRoutes.POST("/inquiry", transactionHandler.H2HInquiry)

		// TODO: Add H2H payment endpoint when ready
		// h2hRoutes.POST("/payment", transactionHandler.H2HPayment)

		// TODO: Add H2H status check endpoint when ready
		// h2hRoutes.POST("/status", transactionHandler.H2HStatus)
	}
}

func configurePublicRoutes(group *gin.RouterGroup) {
	public := group.Group("/public")
	{
		public.GET("/ping", func(c *gin.Context) {
			xresponse.Success(c, "pong", nil)
		})
		public.GET("/health", func(c *gin.Context) {
			xresponse.Success(c, "API is healthy", gin.H{
				"version": "1.0.0",
				"status":  "ok",
			})
		})
	}
}

// authMiddleware validates JWT token and sets user context
func authMiddleware(authService domain.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if authService == nil {
			xresponse.InternalServerError(c, "Auth service not available")
			c.Abort()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			xresponse.Unauthorized(c, "Authorization header with Bearer token required")
			c.Abort()
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if token == "" {
			xresponse.Unauthorized(c, "Token is empty")
			c.Abort()
			return
		}

		claims, err := authService.ValidateToken(token)
		if err != nil {
			switch {
			case errors.Is(err, authpkg.ErrExpiredToken):
				xresponse.Unauthorized(c, "Token expired")
			case errors.Is(err, authpkg.ErrInvalidToken):
				xresponse.Unauthorized(c, "Invalid token")
			case errors.Is(err, authpkg.ErrSignatureInvalid):
				xresponse.Unauthorized(c, "Invalid signature")
			default:
				xresponse.InternalServerError(c, "Failed to validate token")
			}
			c.Abort()
			return
		}

		userID := strings.TrimSpace(claims.UserID)
		if userID == "" {
			xresponse.Unauthorized(c, "Invalid token payload")
			c.Abort()
			return
		}

		role := strings.ToUpper(strings.TrimSpace(claims.Role))
		level := domain.MapRoleToLevel(role)

		// Set all context values for handlers
		c.Set("user_id", userID)
		c.Set("user_role", role)
		c.Set("user_level", level)
		c.Set("token_issued_at", claims.IssuedAt)
		c.Set("token_expires_at", claims.ExpiresAt)

		// Log successful authentication with TTL info
		ttl := time.Until(claims.ExpiresAt)
		logger.Debug("User authenticated via middleware",
			logger.String("user_id", userID),
			logger.String("role", role),
			logger.String("level", fmt.Sprintf("%d", level)),
			logger.String("token_ttl", ttl.String()),
			logger.String("token_expires_at", claims.ExpiresAt.Format(time.RFC3339)),
		)

		c.Next()
	}
}

// adminMiddleware restricts access to admin users only
func adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("user_role")
		if !exists {
			xresponse.Unauthorized(c, "User not authenticated")
			c.Abort()
			return
		}

		role, _ := roleVal.(string)
		if strings.ToUpper(role) != domain.RoleAdmin {
			logger.Warn("Admin access denied",
				logger.String("user_role", role),
				logger.String("required_role", domain.RoleAdmin),
				logger.String("ip", c.ClientIP()),
			)
			xresponse.Forbidden(c, "Admin access required")
			c.Abort()
			return
		}

		logger.Debug("Admin access granted",
			logger.String("user_id", c.GetString("user_id")),
			logger.String("ip", c.ClientIP()),
		)

		c.Next()
	}
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

// rateLimitMiddleware implements basic rate limiting
func rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement proper rate limiting with Redis
		// For now, we'll just log the request
		logger.Debug("API request",
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.String("ip", c.ClientIP()),
		)
		c.Next()
	}
}

// loggingMiddleware logs API requests
func loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Custom log format
		return fmt.Sprintf("[%s] %s %s %d %s %s\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
		)
	})
}

// recoveryMiddleware handles panics
func recoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.Error("Panic recovered",
			logger.String("error", fmt.Sprintf("%v", recovered)),
			logger.String("path", c.Request.URL.Path),
			logger.String("method", c.Request.Method),
		)

		xresponse.InternalServerError(c, "Internal server error")
		c.Abort()
	})
}
