package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/alfanzaky/eraflazz/internal/domain"
)

type APIClientRepository struct {
	db *sql.DB
}

func NewAPIClientRepository(db *sql.DB) *APIClientRepository {
	return &APIClientRepository{db: db}
}

// FindByClientID finds an API client by client_id
func (r *APIClientRepository) FindByClientID(ctx context.Context, clientID string) (*domain.APIClient, error) {
	query := `
		SELECT id, client_id, api_key, secret, ip_whitelist, is_active, 
			   max_requests_per_minute, created_at, updated_at, last_used_at
		FROM api_clients 
		WHERE client_id = $1 AND is_active = true`

	var client domain.APIClient
	var ipWhitelistJSON []byte
	var lastUsedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, clientID).Scan(
		&client.ID,
		&client.ClientID,
		&client.APIKey,
		&client.Secret,
		&ipWhitelistJSON,
		&client.IsActive,
		&client.MaxRequestsPerMinute,
		&client.CreatedAt,
		&client.UpdatedAt,
		&lastUsedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("api client not found")
		}
		return nil, err
	}

	// Parse IP whitelist JSON array
	if len(ipWhitelistJSON) > 0 {
		if err := json.Unmarshal(ipWhitelistJSON, &client.IPWhitelist); err != nil {
			return nil, fmt.Errorf("failed to parse ip_whitelist: %w", err)
		}
	}

	if lastUsedAt.Valid {
		client.LastUsedAt = &lastUsedAt.Time
	}

	return &client, nil
}

// FindByAPIKey finds an API client by api_key
func (r *APIClientRepository) FindByAPIKey(ctx context.Context, apiKey string) (*domain.APIClient, error) {
	query := `
		SELECT id, client_id, api_key, secret, ip_whitelist, is_active, 
			   max_requests_per_minute, created_at, updated_at, last_used_at
		FROM api_clients 
		WHERE api_key = $1 AND is_active = true`

	var client domain.APIClient
	var ipWhitelistJSON []byte
	var lastUsedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, apiKey).Scan(
		&client.ID,
		&client.ClientID,
		&client.APIKey,
		&client.Secret,
		&ipWhitelistJSON,
		&client.IsActive,
		&client.MaxRequestsPerMinute,
		&client.CreatedAt,
		&client.UpdatedAt,
		&lastUsedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("api client not found")
		}
		return nil, err
	}

	// Parse IP whitelist JSON array
	if len(ipWhitelistJSON) > 0 {
		if err := json.Unmarshal(ipWhitelistJSON, &client.IPWhitelist); err != nil {
			return nil, fmt.Errorf("failed to parse ip_whitelist: %w", err)
		}
	}

	if lastUsedAt.Valid {
		client.LastUsedAt = &lastUsedAt.Time
	}

	return &client, nil
}

// UpdateLastUsed updates the last_used_at timestamp for a client
func (r *APIClientRepository) UpdateLastUsed(ctx context.Context, clientID string) error {
	query := `UPDATE api_clients SET last_used_at = NOW() WHERE client_id = $1`

	_, err := r.db.ExecContext(ctx, query, clientID)
	return err
}

// Create creates a new API client
func (r *APIClientRepository) Create(ctx context.Context, client *domain.APIClient) error {
	query := `
		INSERT INTO api_clients (client_id, api_key, secret, ip_whitelist, is_active, max_requests_per_minute)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	ipWhitelistJSON, err := json.Marshal(client.IPWhitelist)
	if err != nil {
		return fmt.Errorf("failed to marshal ip_whitelist: %w", err)
	}

	err = r.db.QueryRowContext(ctx, query,
		client.ClientID,
		client.APIKey,
		client.Secret,
		ipWhitelistJSON,
		client.IsActive,
		client.MaxRequestsPerMinute,
	).Scan(&client.ID, &client.CreatedAt, &client.UpdatedAt)

	return err
}

// FindByID finds an API client by ID
func (r *APIClientRepository) FindByID(ctx context.Context, id string) (*domain.APIClient, error) {
	query := `
		SELECT id, client_id, api_key, secret, ip_whitelist, is_active, 
			   max_requests_per_minute, created_at, updated_at, last_used_at
		FROM api_clients 
		WHERE id = $1`

	var client domain.APIClient
	var ipWhitelistJSON []byte
	var lastUsedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&client.ID,
		&client.ClientID,
		&client.APIKey,
		&client.Secret,
		&ipWhitelistJSON,
		&client.IsActive,
		&client.MaxRequestsPerMinute,
		&client.CreatedAt,
		&client.UpdatedAt,
		&lastUsedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("api client not found")
		}
		return nil, err
	}

	// Parse IP whitelist JSON array
	if len(ipWhitelistJSON) > 0 {
		if err := json.Unmarshal(ipWhitelistJSON, &client.IPWhitelist); err != nil {
			return nil, fmt.Errorf("failed to parse ip_whitelist: %w", err)
		}
	}

	if lastUsedAt.Valid {
		client.LastUsedAt = &lastUsedAt.Time
	}

	return &client, nil
}
