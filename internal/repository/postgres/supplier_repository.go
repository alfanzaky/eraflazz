package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/alfanzaky/eraflazz/internal/domain"
	"github.com/alfanzaky/eraflazz/pkg/logger"
)

type supplierRepository struct {
	db *sqlx.DB
}

// NewSupplierRepository creates a new supplier repository
func NewSupplierRepository(db *sqlx.DB) domain.SupplierRepository {
	return &supplierRepository{db: db}
}

// Create creates a new supplier
func (r *supplierRepository) Create(supplier *domain.Supplier) error {
	query := `
		INSERT INTO suppliers (id, name, code, api_url, api_key, api_secret, api_username, api_password,
			is_active, priority, timeout_seconds, retry_attempts, balance, min_balance_threshold,
			success_rate, avg_response_time_ms, total_transactions, failed_transactions)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`

	_, err := r.db.Exec(query,
		supplier.ID, supplier.Name, supplier.Code, supplier.APIURL, supplier.APIKey,
		supplier.APISecret, supplier.APIUsername, supplier.APIPassword, supplier.IsActive,
		supplier.Priority, supplier.TimeoutSeconds, supplier.RetryAttempts, supplier.Balance,
		supplier.MinBalanceThreshold, supplier.SuccessRate, supplier.AvgResponseTimeMs,
		supplier.TotalTransactions, supplier.FailedTransactions,
	)

	if err != nil {
		logger.Error("Failed to create supplier", 
			logger.String("code", supplier.Code),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to create supplier: %w", err)
	}

	logger.Info("Supplier created successfully", 
		logger.String("supplier_id", supplier.ID),
		logger.String("code", supplier.Code),
	)

	return nil
}

// GetByID retrieves a supplier by ID
func (r *supplierRepository) GetByID(id string) (*domain.Supplier, error) {
	query := `
		SELECT id, name, code, api_url, api_key, api_secret, api_username, api_password,
			is_active, priority, timeout_seconds, retry_attempts, balance, min_balance_threshold,
			success_rate, avg_response_time_ms, total_transactions, failed_transactions,
			created_at, updated_at, last_checked_at, last_success_at
		FROM suppliers WHERE id = $1
	`

	var supplier domain.Supplier
	err := r.db.Get(&supplier, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("supplier not found")
		}
		logger.Error("Failed to get supplier by ID", 
			logger.String("supplier_id", id),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	return &supplier, nil
}

// GetByCode retrieves a supplier by code
func (r *supplierRepository) GetByCode(code string) (*domain.Supplier, error) {
	query := `
		SELECT id, name, code, api_url, api_key, api_secret, api_username, api_password,
			is_active, priority, timeout_seconds, retry_attempts, balance, min_balance_threshold,
			success_rate, avg_response_time_ms, total_transactions, failed_transactions,
			created_at, updated_at, last_checked_at, last_success_at
		FROM suppliers WHERE code = $1
	`

	var supplier domain.Supplier
	err := r.db.Get(&supplier, query, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("supplier not found")
		}
		logger.Error("Failed to get supplier by code", 
			logger.String("code", code),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	return &supplier, nil
}

// Update updates a supplier
func (r *supplierRepository) Update(supplier *domain.Supplier) error {
	query := `
		UPDATE suppliers SET 
			name = $2, code = $3, api_url = $4, api_key = $5, api_secret = $6, 
			api_username = $7, api_password = $8, is_active = $9, priority = $10,
			timeout_seconds = $11, retry_attempts = $12, balance = $13, 
			min_balance_threshold = $14, success_rate = $15, avg_response_time_ms = $16,
			total_transactions = $17, failed_transactions = $18, last_checked_at = $19, last_success_at = $20
		WHERE id = $1
	`

	result, err := r.db.Exec(query,
		supplier.ID, supplier.Name, supplier.Code, supplier.APIURL, supplier.APIKey,
		supplier.APISecret, supplier.APIUsername, supplier.APIPassword, supplier.IsActive,
		supplier.Priority, supplier.TimeoutSeconds, supplier.RetryAttempts, supplier.Balance,
		supplier.MinBalanceThreshold, supplier.SuccessRate, supplier.AvgResponseTimeMs,
		supplier.TotalTransactions, supplier.FailedTransactions, supplier.LastCheckedAt,
		supplier.LastSuccessAt,
	)

	if err != nil {
		logger.Error("Failed to update supplier", 
			logger.String("supplier_id", supplier.ID),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to update supplier: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("supplier not found")
	}

	logger.Info("Supplier updated successfully", 
		logger.String("supplier_id", supplier.ID),
		logger.String("code", supplier.Code),
	)

	return nil
}

// Delete deletes a supplier
func (r *supplierRepository) Delete(id string) error {
	query := `DELETE FROM suppliers WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		logger.Error("Failed to delete supplier", 
			logger.String("supplier_id", id),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to delete supplier: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("supplier not found")
	}

	logger.Info("Supplier deleted successfully", 
		logger.String("supplier_id", id),
	)

	return nil
}

// GetActiveSuppliers retrieves all active suppliers
func (r *supplierRepository) GetActiveSuppliers() ([]*domain.Supplier, error) {
	query := `
		SELECT id, name, code, api_url, api_key, api_secret, api_username, api_password,
			is_active, priority, timeout_seconds, retry_attempts, balance, min_balance_threshold,
			success_rate, avg_response_time_ms, total_transactions, failed_transactions,
			created_at, updated_at, last_checked_at, last_success_at
		FROM suppliers WHERE is_active = true ORDER BY priority ASC, success_rate DESC
	`

	var suppliers []*domain.Supplier
	err := r.db.Select(&suppliers, query)
	if err != nil {
		logger.Error("Failed to get active suppliers", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get active suppliers: %w", err)
	}

	return suppliers, nil
}

// GetSuppliersByPriority retrieves suppliers ordered by priority
func (r *supplierRepository) GetSuppliersByPriority() ([]*domain.Supplier, error) {
	query := `
		SELECT id, name, code, api_url, api_key, api_secret, api_username, api_password,
			is_active, priority, timeout_seconds, retry_attempts, balance, min_balance_threshold,
			success_rate, avg_response_time_ms, total_transactions, failed_transactions,
			created_at, updated_at, last_checked_at, last_success_at
		FROM suppliers ORDER BY priority ASC, success_rate DESC
	`

	var suppliers []*domain.Supplier
	err := r.db.Select(&suppliers, query)
	if err != nil {
		logger.Error("Failed to get suppliers by priority", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get suppliers by priority: %w", err)
	}

	return suppliers, nil
}

// UpdateMetrics updates supplier performance metrics
func (r *supplierRepository) UpdateMetrics(id string, success bool, responseTimeMs int) error {
	query := `
		UPDATE suppliers SET 
			total_transactions = total_transactions + 1,
			failed_transactions = CASE WHEN $2 THEN failed_transactions ELSE failed_transactions + 1 END,
			success_rate = CASE 
				WHEN total_transactions + 1 > 0 
				THEN ((total_transactions + 1 - CASE WHEN $2 THEN failed_transactions ELSE failed_transactions + 1 END) * 100.0 / (total_transactions + 1))
				ELSE 100.0 
			END,
			avg_response_time_ms = CASE 
				WHEN avg_response_time_ms = 0 THEN $3
				ELSE (avg_response_time_ms * 0.7 + $3 * 0.3)::integer
			END,
			last_success_at = CASE WHEN $2 THEN $4 ELSE last_success_at END,
			last_checked_at = $4
		WHERE id = $1
	`
	
	now := time.Now()
	
	result, err := r.db.Exec(query, id, success, responseTimeMs, now)
	if err != nil {
		logger.Error("Failed to update supplier metrics", 
			logger.String("supplier_id", id),
			logger.Bool("success", success),
			logger.Int("response_time_ms", responseTimeMs),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to update supplier metrics: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("supplier not found")
	}

	return nil
}

// GetBalance retrieves supplier balance
func (r *supplierRepository) GetBalance(id string) (float64, error) {
	query := `SELECT balance FROM suppliers WHERE id = $1`

	var balance float64
	err := r.db.Get(&balance, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("supplier not found")
		}
		logger.Error("Failed to get supplier balance", 
			logger.String("supplier_id", id),
			logger.ErrorField(err),
		)
		return 0, fmt.Errorf("failed to get supplier balance: %w", err)
	}

	return balance, nil
}

// UpdateBalance updates supplier balance
func (r *supplierRepository) UpdateBalance(id string, newBalance float64) error {
	query := `UPDATE suppliers SET balance = $2, updated_at = $3 WHERE id = $1`
	now := time.Now()

	result, err := r.db.Exec(query, id, newBalance, now)
	if err != nil {
		logger.Error("Failed to update supplier balance", 
			logger.String("supplier_id", id),
			logger.Float64("new_balance", newBalance),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to update supplier balance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("supplier not found")
	}

	logger.Info("Supplier balance updated", 
		logger.String("supplier_id", id),
		logger.Float64("new_balance", newBalance),
	)

	return nil
}

// GetHealthySuppliers retrieves suppliers that are healthy (active, good success rate, sufficient balance)
func (r *supplierRepository) GetHealthySuppliers() ([]*domain.Supplier, error) {
	query := `
		SELECT id, name, code, api_url, api_key, api_secret, api_username, api_password,
			is_active, priority, timeout_seconds, retry_attempts, balance, min_balance_threshold,
			success_rate, avg_response_time_ms, total_transactions, failed_transactions,
			created_at, updated_at, last_checked_at, last_success_at
		FROM suppliers 
		WHERE is_active = true 
		AND success_rate >= 50.0 
		AND balance >= min_balance_threshold
		ORDER BY priority ASC, success_rate DESC
	`

	var suppliers []*domain.Supplier
	err := r.db.Select(&suppliers, query)
	if err != nil {
		logger.Error("Failed to get healthy suppliers", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get healthy suppliers: %w", err)
	}

	return suppliers, nil
}

// UpdateLastChecked updates the last checked timestamp
func (r *supplierRepository) UpdateLastChecked(id string) error {
	query := `UPDATE suppliers SET last_checked_at = $2 WHERE id = $1`
	now := time.Now()

	result, err := r.db.Exec(query, id, now)
	if err != nil {
		logger.Error("Failed to update last checked", 
			logger.String("supplier_id", id),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to update last checked: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("supplier not found")
	}

	return nil
}

// GetSuppliersNeedingCheck retrieves suppliers that need health check
func (r *supplierRepository) GetSuppliersNeedingCheck(checkIntervalMinutes int) ([]*domain.Supplier, error) {
	query := `
		SELECT id, name, code, api_url, api_key, api_secret, api_username, api_password,
			is_active, priority, timeout_seconds, retry_attempts, balance, min_balance_threshold,
			success_rate, avg_response_time_ms, total_transactions, failed_transactions,
			created_at, updated_at, last_checked_at, last_success_at
		FROM suppliers 
		WHERE is_active = true 
		AND (last_checked_at IS NULL OR last_checked_at < $1)
		ORDER BY priority ASC
	`

	checkTime := time.Now().Add(-time.Duration(checkIntervalMinutes) * time.Minute)
	var suppliers []*domain.Supplier
	err := r.db.Select(&suppliers, query, checkTime)
	if err != nil {
		logger.Error("Failed to get suppliers needing check", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get suppliers needing check: %w", err)
	}

	return suppliers, nil
}
