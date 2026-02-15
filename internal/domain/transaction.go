package domain

import (
	"time"
)

// Transaction represents a transaction in the system
type Transaction struct {
	ID         string  `json:"id" db:"id"`
	TrxCode    string  `json:"trx_code" db:"trx_code"`
	UserID     string  `json:"user_id" db:"user_id"`
	ProductID  string  `json:"product_id" db:"product_id"`
	SupplierID *string `json:"supplier_id" db:"supplier_id"`

	// Transaction details
	DestinationNumber string `json:"destination_number" db:"destination_number"`
	ProductCode       string `json:"product_code" db:"product_code"`

	// Pricing information (snapshot)
	HPP          float64 `json:"hpp" db:"hpp"`
	SellingPrice float64 `json:"selling_price" db:"selling_price"`
	AdminFee     float64 `json:"admin_fee" db:"admin_fee"`
	Profit       float64 `json:"profit" db:"profit"`

	// Status
	Status string `json:"status" db:"status"`

	// Supplier response
	SerialNumber    *string `json:"serial_number" db:"serial_number"`
	SupplierMessage *string `json:"supplier_message" db:"supplier_message"`
	SupplierTrxID   *string `json:"supplier_trx_id" db:"supplier_trx_id"`

	// Routing information
	RoutingAttempts int     `json:"routing_attempts" db:"routing_attempts"`
	FinalSupplierID *string `json:"final_supplier_id" db:"final_supplier_id"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	ProcessedAt *time.Time `json:"processed_at" db:"processed_at"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`

	// Metadata
	UserIP      *string `json:"user_ip" db:"user_ip"`
	UserAgent   *string `json:"user_agent" db:"user_agent"`
	APIEndpoint *string `json:"api_endpoint" db:"api_endpoint"`
	Notes       *string `json:"notes" db:"notes"`
}

// Mutation represents a balance mutation (double-entry accounting)
type Mutation struct {
	ID            string  `json:"id" db:"id"`
	UserID        string  `json:"user_id" db:"user_id"`
	Type          string  `json:"type" db:"type"`
	Amount        float64 `json:"amount" db:"amount"`
	BalanceBefore float64 `json:"balance_before" db:"balance_before"`
	BalanceAfter  float64 `json:"balance_after" db:"balance_after"`

	// Reference information
	ReferenceType *string `json:"reference_type" db:"reference_type"`
	ReferenceID   *string `json:"reference_id" db:"reference_id"`

	// Description
	Description string  `json:"description" db:"description"`
	Notes       *string `json:"notes" db:"notes"`

	// System information
	CreatedBy *string `json:"created_by" db:"created_by"`
	IPAddress *string `json:"ip_address" db:"ip_address"`
	UserAgent *string `json:"user_agent" db:"user_agent"`

	// Timestamp
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// TransactionRepository defines operations for transaction data access
type TransactionRepository interface {
	Create(transaction *Transaction) error
	GetByID(id string) (*Transaction, error)
	GetByTrxCode(trxCode string) (*Transaction, error)
	Update(transaction *Transaction) error
	GetByUserID(userID string, limit, offset int) ([]*Transaction, error)
	GetByStatus(status string) ([]*Transaction, error)
	GetPendingTransactions() ([]*Transaction, error)
	UpdateStatus(id, status string) error
	UpdateSupplierInfo(id, supplierID, supplierTrxID string) error
	GetTransactionsByDateRange(startDate, endDate time.Time) ([]*Transaction, error)
}

// MutationRepository defines operations for mutation data access
type MutationRepository interface {
	Create(mutation *Mutation) error
	GetByID(id string) (*Mutation, error)
	GetByUserID(userID string, limit, offset int) ([]*Mutation, error)
	GetByReference(referenceType, referenceID string) ([]*Mutation, error)
	GetBalanceHistory(userID string, limit, offset int) ([]*Mutation, error)
	GetCurrentBalance(userID string) (float64, error)
}

// TransactionUsecase defines business logic operations for transactions
type TransactionUsecase interface {
	CreateTransaction(userID, productCode, destinationNumber string) (*Transaction, error)
	ProcessTransaction(transactionID string) error
	ProcessPendingTransactions() error
	RetryFailedTransaction(transactionID string) error
	GetTransaction(id string) (*Transaction, error)
	GetUserTransactions(userID string, page, limit int) ([]*Transaction, error)
	GetTransactionByTrxCode(trxCode string) (*Transaction, error)
	CancelTransaction(transactionID string) error
	RefundTransaction(transactionID string) error
	GetTransactionStats(userID string, startDate, endDate time.Time) (*TransactionStats, error)
}

// TransactionUsecase defines business logic operations for mutations
type MutationUsecase interface {
	CreateMutation(userID, mutationType string, amount, balanceBefore, balanceAfter float64, description string, referenceType, referenceID *string) error
	GetUserMutations(userID string, page, limit int) ([]*Mutation, error)
	GetBalanceHistory(userID string, startDate, endDate time.Time) ([]*Mutation, error)
	GetCurrentBalance(userID string) (float64, error)
	ValidateBalance(userID string, requiredAmount float64) error
}

// TransactionStats represents transaction statistics
type TransactionStats struct {
	TotalTransactions int     `json:"total_transactions"`
	SuccessCount      int     `json:"success_count"`
	FailedCount       int     `json:"failed_count"`
	PendingCount      int     `json:"pending_count"`
	TotalRevenue      float64 `json:"total_revenue"`
	TotalProfit       float64 `json:"total_profit"`
	AverageAmount     float64 `json:"average_amount"`
}

// Transaction validation constants
const (
	StatusPending    = "PENDING"
	StatusProcessing = "PROCESSING"
	StatusSuccess    = "SUCCESS"
	StatusFailed     = "FAILED"
	StatusRefund     = "REFUND"
	StatusTimeout    = "TIMEOUT"

	MutationTypeDebit  = "DEBIT"  // Money in
	MutationTypeCredit = "CREDIT" // Money out

	ReferenceTypeTransaction = "TRANSACTION"
	ReferenceTypeDeposit     = "DEPOSIT"
	ReferenceTypeWithdrawal  = "WITHDRAWAL"
	ReferenceTypeCommission  = "COMMISSION"
	ReferenceTypePenalty     = "PENALTY"
)

// IsValidStatus checks if the transaction status is valid
func IsValidStatus(status string) bool {
	validStatuses := []string{
		StatusPending, StatusProcessing, StatusSuccess,
		StatusFailed, StatusRefund, StatusTimeout,
	}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// IsValidMutationType checks if the mutation type is valid
func IsValidMutationType(mutationType string) bool {
	return mutationType == MutationTypeDebit || mutationType == MutationTypeCredit
}

// IsFinalStatus checks if the transaction status is final (no more processing)
func (t *Transaction) IsFinalStatus() bool {
	return t.Status == StatusSuccess || t.Status == StatusFailed || t.Status == StatusRefund || t.Status == StatusTimeout
}

// CanRetry checks if the transaction can be retried
func (t *Transaction) CanRetry() bool {
	return t.Status == StatusFailed && t.RoutingAttempts < 3
}

// GetDuration returns the duration of the transaction
func (t *Transaction) GetDuration() *time.Duration {
	if t.ProcessedAt == nil || t.CompletedAt == nil {
		return nil
	}
	duration := t.CompletedAt.Sub(*t.ProcessedAt)
	return &duration
}

// CalculateProfit returns the profit for this transaction
func (t *Transaction) CalculateProfit() float64 {
	return t.SellingPrice - t.HPP - t.AdminFee
}

// IsExpired checks if the transaction is expired (for timeout handling)
func (t *Transaction) IsExpired(timeoutMinutes int) bool {
	if t.Status != StatusPending && t.Status != StatusProcessing {
		return false
	}
	expiryTime := t.CreatedAt.Add(time.Duration(timeoutMinutes) * time.Minute)
	return time.Now().After(expiryTime)
}
