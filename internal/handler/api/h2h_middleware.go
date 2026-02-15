package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/alfanzaky/eraflazz/internal/domain"
	"github.com/alfanzaky/eraflazz/internal/repository/postgres"
	"github.com/gin-gonic/gin"
)

type H2HMiddleware struct {
	clientRepo *postgres.APIClientRepository
}

func NewH2HMiddleware(clientRepo *postgres.APIClientRepository) *H2HMiddleware {
	return &H2HMiddleware{
		clientRepo: clientRepo,
	}
}

// H2HAuth middleware validates H2H API requests
func (m *H2HMiddleware) H2HAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract required headers
		headers := &domain.H2HRequestHeaders{
			ClientID:  c.GetHeader("X-Client-ID"),
			APIKey:    c.GetHeader("X-API-Key"),
			Timestamp: c.GetHeader("X-Timestamp"),
			Signature: c.GetHeader("X-Signature"),
			Nonce:     c.GetHeader("X-Nonce"),
		}

		// Validate required headers
		if headers.ClientID == "" || headers.APIKey == "" || headers.Timestamp == "" || headers.Signature == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing required H2H headers",
				"code":  "MISSING_HEADERS",
			})
			c.Abort()
			return
		}

		// Find API client by ClientID
		client, err := m.clientRepo.FindByClientID(c.Request.Context(), headers.ClientID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid client credentials",
				"code":  "INVALID_CLIENT",
			})
			c.Abort()
			return
		}

		// Validate API Key
		if client.APIKey != headers.APIKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
				"code":  "INVALID_API_KEY",
			})
			c.Abort()
			return
		}

		// Check IP whitelist
		clientIP := c.ClientIP()
		if !client.IsIPAllowed(clientIP) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "IP address not allowed",
				"code":  "IP_NOT_ALLOWED",
			})
			c.Abort()
			return
		}

		// Read request body for signature validation
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Validate signature
		err = domain.ValidateSignature(client.Secret, headers.Timestamp, headers.Signature, bodyBytes)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid signature: " + err.Error(),
				"code":  "INVALID_SIGNATURE",
			})
			c.Abort()
			return
		}

		// Update last used timestamp (async, don't block request)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			m.clientRepo.UpdateLastUsed(ctx, headers.ClientID)
		}()

		// Set client info in context
		c.Set("client_id", headers.ClientID)
		c.Set("client_info", client)

		c.Next()
	}
}

// OptionalH2HAuth middleware applies H2H auth only if headers are present
func (m *H2HMiddleware) OptionalH2HAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := c.GetHeader("X-Client-ID")
		apiKey := c.GetHeader("X-API-Key")

		// If H2H headers are present, validate them
		if clientID != "" || apiKey != "" {
			m.H2HAuth()(c)
			return
		}

		// Otherwise, continue without H2H validation
		c.Next()
	}
}

// GetClientFromContext retrieves client info from gin context
func GetClientFromContext(c *gin.Context) (*domain.APIClient, bool) {
	if client, exists := c.Get("client_info"); exists {
		if apiClient, ok := client.(*domain.APIClient); ok {
			return apiClient, true
		}
	}
	return nil, false
}

// GetClientIDFromContext retrieves client ID from gin context
func GetClientIDFromContext(c *gin.Context) (string, bool) {
	if clientID, exists := c.Get("client_id"); exists {
		if id, ok := clientID.(string); ok {
			return id, true
		}
	}
	return "", false
}
