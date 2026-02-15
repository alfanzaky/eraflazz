package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/alfanzaky/eraflazz/internal/domain"
	"github.com/alfanzaky/eraflazz/internal/repository/postgres"
	"github.com/alfanzaky/eraflazz/pkg/logger"
	"github.com/alfanzaky/eraflazz/pkg/xresponse"
	"github.com/gin-gonic/gin"
)

type APIClientHandler struct {
	clientRepo *postgres.APIClientRepository
}

func NewAPIClientHandler(clientRepo *postgres.APIClientRepository) *APIClientHandler {
	return &APIClientHandler{
		clientRepo: clientRepo,
	}
}

// CreateAPIClient creates a new API client for H2H integration
func (h *APIClientHandler) CreateAPIClient(c *gin.Context) {
	var request struct {
		ClientID             string   `json:"client_id" binding:"required"`
		IPWhitelist          []string `json:"ip_whitelist"`
		MaxRequestsPerMinute int      `json:"max_requests_per_minute"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		xresponse.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	// Generate API key and secret
	apiKey := generateRandomString(32)
	secret := generateRandomString(64)

	// Set default values
	if request.MaxRequestsPerMinute == 0 {
		request.MaxRequestsPerMinute = 60
	}

	client := &domain.APIClient{
		ClientID:             request.ClientID,
		APIKey:               apiKey,
		Secret:               secret,
		IPWhitelist:          request.IPWhitelist,
		IsActive:             true,
		MaxRequestsPerMinute: request.MaxRequestsPerMinute,
	}

	if err := h.clientRepo.Create(c.Request.Context(), client); err != nil {
		logger.Error("Failed to create API client",
			logger.String("client_id", request.ClientID),
			logger.String("error", err.Error()),
		)
		xresponse.InternalServerError(c, "Failed to create API client")
		return
	}

	// Don't return secret in response
	client.Secret = ""

	logger.Info("API client created successfully",
		logger.String("client_id", client.ClientID),
		logger.String("api_key", client.APIKey),
	)

	xresponse.Created(c, "API client created successfully", client)
}

// GetAPIClient retrieves API client information
func (h *APIClientHandler) GetAPIClient(c *gin.Context) {
	clientID := c.Param("client_id")
	if clientID == "" {
		xresponse.BadRequest(c, "Client ID is required")
		return
	}

	client, err := h.clientRepo.FindByClientID(c.Request.Context(), clientID)
	if err != nil {
		logger.Warn("API client not found",
			logger.String("client_id", clientID),
			logger.String("error", err.Error()),
		)
		xresponse.NotFound(c, "API client not found")
		return
	}

	// Don't return secret in response
	client.Secret = ""

	xresponse.Success(c, "API client retrieved successfully", client)
}

// ListAPIClients lists all active API clients (admin only)
func (h *APIClientHandler) ListAPIClients(c *gin.Context) {
	// TODO: Implement pagination and filtering
	// For now, this is a placeholder implementation
	xresponse.Success(c, "API clients list", gin.H{
		"clients": []interface{}{},
		"total":   0,
		"page":    1,
		"limit":   10,
	})
}

// RegenerateSecret regenerates API client secret
func (h *APIClientHandler) RegenerateSecret(c *gin.Context) {
	clientID := c.Param("client_id")
	if clientID == "" {
		xresponse.BadRequest(c, "Client ID is required")
		return
	}

	// Get existing client
	client, err := h.clientRepo.FindByClientID(c.Request.Context(), clientID)
	if err != nil {
		xresponse.NotFound(c, "API client not found")
		return
	}

	// Generate new secret
	newSecret := generateRandomString(64)
	client.Secret = newSecret

	// TODO: Update client in database
	// For now, return the new secret
	logger.Info("API client secret regenerated",
		logger.String("client_id", clientID),
	)

	xresponse.Success(c, "Secret regenerated successfully", gin.H{
		"client_id": clientID,
		"secret":    newSecret,
		"warning":   "Please save this secret securely. It won't be shown again.",
	})
}

// generateRandomString generates a random hex string
func generateRandomString(length int) string {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to less secure method if crypto/rand fails
		return fmt.Sprintf("%x", length)
	}
	return hex.EncodeToString(bytes)
}
