package domain

import (
	"time"
)

// Supplier represents a supplier in the system
type Supplier struct {
	ID   string `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	Code string `json:"code" db:"code"`

	// API Configuration
	APIURL      string  `json:"api_url" db:"api_url"`
	APIKey      *string `json:"api_key" db:"api_key"`
	APISecret   *string `json:"api_secret" db:"api_secret"`
	APIUsername *string `json:"api_username" db:"api_username"`
	APIPassword *string `json:"api_password" db:"api_password"`

	// Supplier status and settings
	IsActive       bool `json:"is_active" db:"is_active"`
	Priority       int  `json:"priority" db:"priority"`
	TimeoutSeconds int  `json:"timeout_seconds" db:"timeout_seconds"`
	RetryAttempts  int  `json:"retry_attempts" db:"retry_attempts"`

	// Financial information
	Balance             float64 `json:"balance" db:"balance"`
	MinBalanceThreshold float64 `json:"min_balance_threshold" db:"min_balance_threshold"`

	// Performance metrics
	SuccessRate        float64 `json:"success_rate" db:"success_rate"`
	AvgResponseTimeMs  int     `json:"avg_response_time_ms" db:"avg_response_time_ms"`
	TotalTransactions  int     `json:"total_transactions" db:"total_transactions"`
	FailedTransactions int     `json:"failed_transactions" db:"failed_transactions"`

	// Timestamps
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	LastCheckedAt *time.Time `json:"last_checked_at" db:"last_checked_at"`
	LastSuccessAt *time.Time `json:"last_success_at" db:"last_success_at"`
}

// SupplierRepository defines operations for supplier data access
type SupplierRepository interface {
	Create(supplier *Supplier) error
	GetByID(id string) (*Supplier, error)
	GetByCode(code string) (*Supplier, error)
	Update(supplier *Supplier) error
	Delete(id string) error
	GetActiveSuppliers() ([]*Supplier, error)
	GetSuppliersByPriority() ([]*Supplier, error)
	UpdateMetrics(id string, success bool, responseTimeMs int) error
	GetBalance(id string) (float64, error)
	UpdateBalance(id string, newBalance float64) error
}

// SupplierUsecase defines business logic operations for suppliers
type SupplierUsecase interface {
	CreateSupplier(supplier *Supplier) error
	UpdateSupplier(id string, updates *Supplier) error
	GetSupplier(id string) (*Supplier, error)
	GetSupplierByCode(code string) (*Supplier, error)
	GetActiveSuppliers() ([]*Supplier, error)
	GetSuppliersForProduct(productID string) ([]*Supplier, error)
	DeactivateSupplier(id string) error
	UpdateSupplierMetrics(id string, success bool, responseTimeMs int) error
	CheckSupplierHealth(id string) error
	GetBestSupplier(productID string) (*Supplier, error)
}

// SupplierRequest represents a request to supplier API
type SupplierRequest struct {
	ProductCode       string            `json:"product_code"`
	DestinationNumber string            `json:"destination_number"`
	RefID             string            `json:"ref_id"`
	AdditionalData    map[string]string `json:"additional_data,omitempty"`
}

// SupplierResponse represents a response from supplier API
type SupplierResponse struct {
	Success      bool                   `json:"success"`
	Message      string                 `json:"message"`
	TrxID        string                 `json:"trx_id"`
	SerialNumber string                 `json:"serial_number"`
	StatusCode   int                    `json:"status_code"`
	ResponseTime int                    `json:"response_time_ms"`
	Data         map[string]interface{} `json:"data,omitempty"`
}

// SupplierAdapter defines the interface for supplier integrations
type SupplierAdapter interface {
	TopUp(request *SupplierRequest) (*SupplierResponse, error)
	CheckBalance() (float64, error)
	CheckStatus(trxID string) (*SupplierResponse, error)
	GetProductCatalog() ([]*Product, error)
	ParseResponse(response []byte) (*SupplierResponse, error)
}

// SupplierAdapterFactory resolves supplier adapters by supplier code
type SupplierAdapterFactory interface {
	RegisterAdapter(code string, adapter SupplierAdapter)
	GetAdapter(code string) (SupplierAdapter, error)
}

// Supplier validation constants
const (
	SupplierCodeDigiflazz = "DIGIFLAZZ"
	SupplierCodeVIP       = "VIP"
	SupplierCodeOtomax    = "OTOMAX"
	SupplierCodeTokopedia = "TOKOPEDIA"

	DefaultTimeoutSeconds   = 30
	DefaultRetryAttempts    = 3
	DefaultPriority         = 1
	MinSuccessRateThreshold = 50.0 // Minimum success rate to consider supplier reliable
)

// IsValidSupplierCode checks if the supplier code is valid
func IsValidSupplierCode(code string) bool {
	validCodes := []string{
		SupplierCodeDigiflazz, SupplierCodeVIP, SupplierCodeOtomax, SupplierCodeTokopedia,
	}
	for _, c := range validCodes {
		if c == code {
			return true
		}
	}
	return false
}

// IsHealthy checks if the supplier is healthy based on metrics
func (s *Supplier) IsHealthy() bool {
	if !s.IsActive {
		return false
	}
	if s.SuccessRate < MinSuccessRateThreshold {
		return false
	}
	if s.Balance < s.MinBalanceThreshold {
		return false
	}
	return true
}

// UpdatePerformanceMetrics updates the supplier's performance metrics
func (s *Supplier) UpdatePerformanceMetrics(success bool, responseTimeMs int) {
	s.TotalTransactions++
	if !success {
		s.FailedTransactions++
	}

	// Update success rate
	if s.TotalTransactions > 0 {
		s.SuccessRate = float64(s.TotalTransactions-s.FailedTransactions) / float64(s.TotalTransactions) * 100
	}

	// Update average response time (simple moving average)
	if s.AvgResponseTimeMs == 0 {
		s.AvgResponseTimeMs = responseTimeMs
	} else {
		// Weighted average: 70% old, 30% new
		s.AvgResponseTimeMs = int(float64(s.AvgResponseTimeMs)*0.7 + float64(responseTimeMs)*0.3)
	}

	if success {
		now := time.Now()
		s.LastSuccessAt = &now
	}

	now := time.Now()
	s.LastCheckedAt = &now
}

// GetFailureRate returns the failure rate percentage
func (s *Supplier) GetFailureRate() float64 {
	if s.TotalTransactions == 0 {
		return 0.0
	}
	return float64(s.FailedTransactions) / float64(s.TotalTransactions) * 100
}

// ShouldRetry determines if a request should be retried based on supplier metrics
func (s *Supplier) ShouldRetry(attemptCount int) bool {
	if attemptCount >= s.RetryAttempts {
		return false
	}
	if !s.IsHealthy() {
		return false
	}
	return true
}

// GetPriorityWeight returns a weight for supplier selection based on priority and performance
func (s *Supplier) GetPriorityWeight() float64 {
	if !s.IsHealthy() {
		return 0.0
	}

	// Lower priority number = higher weight
	priorityWeight := 1.0 / float64(s.Priority)

	// Success rate weight (0-1)
	successRateWeight := s.SuccessRate / 100.0

	// Response time weight (inverse - faster is better)
	responseTimeWeight := 1.0
	if s.AvgResponseTimeMs > 0 {
		responseTimeWeight = 10000.0 / float64(s.AvgResponseTimeMs) // Normalize
	}

	// Combined weight (adjust weights as needed)
	return priorityWeight*0.4 + successRateWeight*0.4 + responseTimeWeight*0.2
}
