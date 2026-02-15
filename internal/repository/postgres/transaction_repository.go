package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/alfanzaky/eraflazz/internal/domain"
	"github.com/alfanzaky/eraflazz/pkg/logger"
)

type transactionRepository struct {
	db *sqlx.DB
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *sqlx.DB) domain.TransactionRepository {
	return &transactionRepository{db: db}
}

// Create creates a new transaction
func (r *transactionRepository) Create(transaction *domain.Transaction) error {
	query := `
		INSERT INTO transactions (id, trx_code, user_id, product_id, supplier_id,
			destination_number, product_code, hpp, selling_price, admin_fee,
			status, user_ip, user_agent, api_endpoint, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := r.db.Exec(query,
		transaction.ID, transaction.TrxCode, transaction.UserID, transaction.ProductID,
		transaction.SupplierID, transaction.DestinationNumber, transaction.ProductCode,
		transaction.HPP, transaction.SellingPrice, transaction.AdminFee,
		transaction.Status, transaction.UserIP, transaction.UserAgent,
		transaction.APIEndpoint, transaction.Notes,
	)

	if err != nil {
		logger.Error("Failed to create transaction", 
			logger.String("trx_code", transaction.TrxCode),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	logger.Info("Transaction created successfully", 
		logger.String("trx_id", transaction.ID),
		logger.String("trx_code", transaction.TrxCode),
	)

	return nil
}

// GetByID retrieves a transaction by ID
func (r *transactionRepository) GetByID(id string) (*domain.Transaction, error) {
	query := `
		SELECT id, trx_code, user_id, product_id, supplier_id,
			destination_number, product_code, hpp, selling_price, admin_fee, profit,
			status, serial_number, supplier_message, supplier_trx_id,
			routing_attempts, final_supplier_id,
			created_at, updated_at, processed_at, completed_at,
			user_ip, user_agent, api_endpoint, notes
		FROM transactions WHERE id = $1
	`

	var transaction domain.Transaction
	err := r.db.Get(&transaction, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found")
		}
		logger.Error("Failed to get transaction by ID", 
			logger.String("trx_id", id),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &transaction, nil
}

// GetByTrxCode retrieves a transaction by transaction code
func (r *transactionRepository) GetByTrxCode(trxCode string) (*domain.Transaction, error) {
	query := `
		SELECT id, trx_code, user_id, product_id, supplier_id,
			destination_number, product_code, hpp, selling_price, admin_fee, profit,
			status, serial_number, supplier_message, supplier_trx_id,
			routing_attempts, final_supplier_id,
			created_at, updated_at, processed_at, completed_at,
			user_ip, user_agent, api_endpoint, notes
		FROM transactions WHERE trx_code = $1
	`

	var transaction domain.Transaction
	err := r.db.Get(&transaction, query, trxCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found")
		}
		logger.Error("Failed to get transaction by code", 
			logger.String("trx_code", trxCode),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &transaction, nil
}

// Update updates a transaction
func (r *transactionRepository) Update(transaction *domain.Transaction) error {
	query := `
		UPDATE transactions SET 
			supplier_id = $2, status = $3, serial_number = $4, supplier_message = $5,
			supplier_trx_id = $6, routing_attempts = $7, final_supplier_id = $8,
			processed_at = $9, completed_at = $10, notes = $11
		WHERE id = $1
	`

	result, err := r.db.Exec(query,
		transaction.ID, transaction.SupplierID, transaction.Status,
		transaction.SerialNumber, transaction.SupplierMessage,
		transaction.SupplierTrxID, transaction.RoutingAttempts,
		transaction.FinalSupplierID, transaction.ProcessedAt,
		transaction.CompletedAt, transaction.Notes,
	)

	if err != nil {
		logger.Error("Failed to update transaction", 
			logger.String("trx_id", transaction.ID),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transaction not found")
	}

	logger.Info("Transaction updated successfully", 
		logger.String("trx_id", transaction.ID),
		logger.String("status", transaction.Status),
	)

	return nil
}

// GetByUserID retrieves transactions by user ID with pagination
func (r *transactionRepository) GetByUserID(userID string, limit, offset int) ([]*domain.Transaction, error) {
	query := `
		SELECT id, trx_code, user_id, product_id, supplier_id,
			destination_number, product_code, hpp, selling_price, admin_fee, profit,
			status, serial_number, supplier_message, supplier_trx_id,
			routing_attempts, final_supplier_id,
			created_at, updated_at, processed_at, completed_at,
			user_ip, user_agent, api_endpoint, notes
		FROM transactions 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`

	var transactions []*domain.Transaction
	err := r.db.Select(&transactions, query, userID, limit, offset)
	if err != nil {
		logger.Error("Failed to get transactions by user ID", 
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get transactions by user ID: %w", err)
	}

	return transactions, nil
}

// GetByStatus retrieves transactions by status
func (r *transactionRepository) GetByStatus(status string) ([]*domain.Transaction, error) {
	query := `
		SELECT id, trx_code, user_id, product_id, supplier_id,
			destination_number, product_code, hpp, selling_price, admin_fee, profit,
			status, serial_number, supplier_message, supplier_trx_id,
			routing_attempts, final_supplier_id,
			created_at, updated_at, processed_at, completed_at,
			user_ip, user_agent, api_endpoint, notes
		FROM transactions 
		WHERE status = $1 
		ORDER BY created_at ASC
	`

	var transactions []*domain.Transaction
	err := r.db.Select(&transactions, query, status)
	if err != nil {
		logger.Error("Failed to get transactions by status", 
			logger.String("status", status),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get transactions by status: %w", err)
	}

	return transactions, nil
}

// GetPendingTransactions retrieves all pending transactions
func (r *transactionRepository) GetPendingTransactions() ([]*domain.Transaction, error) {
	return r.GetByStatus(domain.StatusPending)
}

// UpdateStatus updates transaction status
func (r *transactionRepository) UpdateStatus(id, status string) error {
	query := `UPDATE transactions SET status = $2, updated_at = $3 WHERE id = $1`
	now := time.Now()

	result, err := r.db.Exec(query, id, status, now)
	if err != nil {
		logger.Error("Failed to update transaction status", 
			logger.String("trx_id", id),
			logger.String("status", status),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transaction not found")
	}

	logger.Info("Transaction status updated", 
		logger.String("trx_id", id),
		logger.String("status", status),
	)

	return nil
}

// UpdateSupplierInfo updates supplier information for a transaction
func (r *transactionRepository) UpdateSupplierInfo(id, supplierID, supplierTrxID string) error {
	query := `
		UPDATE transactions SET 
			supplier_id = $2, supplier_trx_id = $3, updated_at = $4
		WHERE id = $1
	`
	now := time.Now()

	result, err := r.db.Exec(query, id, supplierID, supplierTrxID, now)
	if err != nil {
		logger.Error("Failed to update supplier info", 
			logger.String("trx_id", id),
			logger.String("supplier_id", supplierID),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to update supplier info: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transaction not found")
	}

	return nil
}

// GetTransactionsByDateRange retrieves transactions within date range
func (r *transactionRepository) GetTransactionsByDateRange(startDate, endDate time.Time) ([]*domain.Transaction, error) {
	query := `
		SELECT id, trx_code, user_id, product_id, supplier_id,
			destination_number, product_code, hpp, selling_price, admin_fee, profit,
			status, serial_number, supplier_message, supplier_trx_id,
			routing_attempts, final_supplier_id,
			created_at, updated_at, processed_at, completed_at,
			user_ip, user_agent, api_endpoint, notes
		FROM transactions 
		WHERE created_at BETWEEN $1 AND $2 
		ORDER BY created_at DESC
	`

	var transactions []*domain.Transaction
	err := r.db.Select(&transactions, query, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get transactions by date range", 
			logger.String("start_date", startDate.Format(time.RFC3339)),
			logger.String("end_date", endDate.Format(time.RFC3339)),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get transactions by date range: %w", err)
	}

	return transactions, nil
}

// UpdateProcessingInfo updates processing information
func (r *transactionRepository) UpdateProcessingInfo(id string) error {
	query := `UPDATE transactions SET processed_at = $2, status = $3 WHERE id = $1`
	now := time.Now()

	result, err := r.db.Exec(query, id, now, domain.StatusProcessing)
	if err != nil {
		logger.Error("Failed to update processing info", 
			logger.String("trx_id", id),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to update processing info: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transaction not found")
	}

	return nil
}

// UpdateCompletionInfo updates completion information
func (r *transactionRepository) UpdateCompletionInfo(id, status, serialNumber, supplierMessage string) error {
	query := `
		UPDATE transactions SET 
			status = $2, serial_number = $3, supplier_message = $4, completed_at = $5
		WHERE id = $1
	`
	now := time.Now()

	result, err := r.db.Exec(query, id, status, serialNumber, supplierMessage, now)
	if err != nil {
		logger.Error("Failed to update completion info", 
			logger.String("trx_id", id),
			logger.String("status", status),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to update completion info: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transaction not found")
	}

	return nil
}

// IncrementRoutingAttempts increments routing attempts counter
func (r *transactionRepository) IncrementRoutingAttempts(id string) error {
	query := `
		UPDATE transactions SET 
			routing_attempts = routing_attempts + 1, updated_at = $2
		WHERE id = $1
	`
	now := time.Now()

	result, err := r.db.Exec(query, id, now)
	if err != nil {
		logger.Error("Failed to increment routing attempts", 
			logger.String("trx_id", id),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to increment routing attempts: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transaction not found")
	}

	return nil
}

// GetTransactionCountByStatus gets count of transactions by status
func (r *transactionRepository) GetTransactionCountByStatus(status string) (int, error) {
	query := `SELECT COUNT(*) FROM transactions WHERE status = $1`

	var count int
	err := r.db.Get(&count, query, status)
	if err != nil {
		logger.Error("Failed to get transaction count by status", 
			logger.String("status", status),
			logger.ErrorField(err),
		)
		return 0, fmt.Errorf("failed to get transaction count: %w", err)
	}

	return count, nil
}

// GetExpiredTransactions retrieves transactions that have expired
func (r *transactionRepository) GetExpiredTransactions(timeoutMinutes int) ([]*domain.Transaction, error) {
	query := `
		SELECT id, trx_code, user_id, product_id, supplier_id,
			destination_number, product_code, hpp, selling_price, admin_fee, profit,
			status, serial_number, supplier_message, supplier_trx_id,
			routing_attempts, final_supplier_id,
			created_at, updated_at, processed_at, completed_at,
			user_ip, user_agent, api_endpoint, notes
		FROM transactions 
		WHERE status IN ($1, $2) 
		AND created_at < $3
		ORDER BY created_at ASC
	`

	expiryTime := time.Now().Add(-time.Duration(timeoutMinutes) * time.Minute)
	var transactions []*domain.Transaction
	err := r.db.Select(&transactions, query, domain.StatusPending, domain.StatusProcessing, expiryTime)
	if err != nil {
		logger.Error("Failed to get expired transactions", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get expired transactions: %w", err)
	}

	return transactions, nil
}
