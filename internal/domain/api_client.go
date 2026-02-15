package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// APIClient represents an H2H API client
type APIClient struct {
	ID                   string    `json:"id"`
	ClientID             string    `json:"client_id"`
	APIKey               string    `json:"api_key"`
	Secret               string    `json:"secret,omitempty"`
	IPWhitelist          []string  `json:"ip_whitelist"`
	IsActive             bool      `json:"is_active"`
	MaxRequestsPerMinute int       `json:"max_requests_per_minute"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	LastUsedAt           *time.Time `json:"last_used_at,omitempty"`
}

// H2HRequestHeaders represents required headers for H2H requests
type H2HRequestHeaders struct {
	ClientID  string `json:"client_id"`
	APIKey    string `json:"api_key"`
	Timestamp string `json:"timestamp"`
	Signature string `json:"signature"`
	Nonce     string `json:"nonce,omitempty"`
}

// ValidateSignature validates HMAC-SHA256 signature for H2H requests
func ValidateSignature(secret, timestamp, signature string, payload []byte) error {
	// Check timestamp validity (prevent replay attacks)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp format")
	}

	requestTime := time.Unix(ts, 0)
	now := time.Now()
	
	// Allow 5 minute window for timestamp
	if now.Sub(requestTime) > 5*time.Minute || requestTime.Sub(now) > 5*time.Minute {
		return fmt.Errorf("timestamp expired or too far in future")
	}

	// Create expected signature: HMAC-SHA256(secret, timestamp+payload)
	dataToSign := timestamp + string(payload)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(dataToSign))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Compare signatures securely
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// GenerateSignature generates HMAC-SHA256 signature for H2H requests
func GenerateSignature(secret, timestamp string, payload []byte) string {
	dataToSign := timestamp + string(payload)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(dataToSign))
	return hex.EncodeToString(h.Sum(nil))
}

// IsIPAllowed checks if IP address is in whitelist
func (c *APIClient) IsIPAllowed(ip string) bool {
	if len(c.IPWhitelist) == 0 {
		return true // No whitelist restriction
	}

	for _, allowedIP := range c.IPWhitelist {
		if strings.TrimSpace(allowedIP) == ip {
			return true
		}
	}

	return false
}

// UpdateLastUsed updates the last used timestamp
func (c *APIClient) UpdateLastUsed() {
	now := time.Now()
	c.LastUsedAt = &now
}
