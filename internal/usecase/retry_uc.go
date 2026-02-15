package usecase

import (
	"fmt"
	"time"

	"github.com/alfanzaky/eraflazz/internal/domain"
	"github.com/alfanzaky/eraflazz/pkg/logger"
	"github.com/alfanzaky/eraflazz/pkg/utils"
)

type retryUsecase struct {
	transactionRepo domain.TransactionRepository
	supplierRepo    domain.SupplierRepository
	smartRoutingUC  *smartRoutingUsecase
}

// NewRetryUsecase creates a new retry use case
func NewRetryUsecase(
	transactionRepo domain.TransactionRepository,
	supplierRepo domain.SupplierRepository,
	smartRoutingUC *smartRoutingUsecase,
) *retryUsecase {
	return &retryUsecase{
		transactionRepo: transactionRepo,
		supplierRepo:    supplierRepo,
		smartRoutingUC:  smartRoutingUC,
	}
}

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxAttempts       int           // Maximum retry attempts
	InitialDelay      time.Duration // Initial delay between retries
	MaxDelay          time.Duration // Maximum delay between retries
	BackoffMultiplier float64       // Multiplier for exponential backoff
	TimeoutPerAttempt time.Duration // Timeout for each attempt
	EnableJitter      bool          // Add random jitter to prevent thundering herd
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:       3,
		InitialDelay:      2 * time.Second,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
		TimeoutPerAttempt: 30 * time.Second,
		EnableJitter:      true,
	}
}

// RetryResult represents the result of a retry operation
type RetryResult struct {
	Success         bool
	AttemptsMade    int
	TotalDuration   time.Duration
	FinalSupplierID string
	FinalError      error
	AttemptHistory  []*RetryAttempt
	RefundIssued    bool
	RefundAmount    float64
}

// RetryAttempt represents a single retry attempt
type RetryAttempt struct {
	AttemptNumber  int
	SupplierID     string
	SupplierCode   string
	StartTime      time.Time
	EndTime        time.Time
	Duration       time.Duration
	Success        bool
	Error          error
	ResponseTimeMs int
	Reason         string
}

// RetryTransaction implements intelligent retry logic with failover
func (uc *retryUsecase) RetryTransaction(transactionID string, config *RetryConfig) (*RetryResult, error) {
	if config == nil {
		config = DefaultRetryConfig()
	}

	// Get transaction
	transaction, err := uc.transactionRepo.GetByID(transactionID)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	// Check if transaction can be retried
	if !uc.canRetryTransaction(transaction, config) {
		return &RetryResult{
			Success:      false,
			AttemptsMade: transaction.RoutingAttempts,
			FinalError:   fmt.Errorf("transaction cannot be retried"),
			RefundIssued: transaction.Status == domain.StatusRefund,
		}, nil
	}

	logger.Info("Starting retry process",
		logger.String("trx_id", transactionID),
		logger.String("trx_code", transaction.TrxCode),
		logger.Int("max_attempts", config.MaxAttempts),
	)

	startTime := time.Now()
	result := &RetryResult{
		AttemptHistory: make([]*RetryAttempt, 0),
		RefundAmount:   transaction.SellingPrice,
	}

	// Get available suppliers for failover
	suppliers, err := uc.getFailoverSuppliers(transaction.ProductID, config.MaxAttempts)
	if err != nil {
		logger.Error("Failed to get failover suppliers",
			logger.String("trx_id", transactionID),
			logger.ErrorField(err),
		)
		result.FinalError = fmt.Errorf("failed to get suppliers: %w", err)
		return result, nil
	}

	if len(suppliers) == 0 {
		result.FinalError = fmt.Errorf("no suppliers available for retry")
		return result, nil
	}

	// Execute retry attempts
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		if attempt > len(suppliers) {
			logger.Warn("No more suppliers available for retry",
				logger.String("trx_id", transactionID),
				logger.Int("attempt", attempt),
			)
			break
		}

		supplier := suppliers[attempt-1]
		attemptResult := uc.executeRetryAttempt(transaction, supplier, attempt, config)
		result.AttemptHistory = append(result.AttemptHistory, attemptResult)
		result.AttemptsMade = attempt

		// Update transaction routing attempts
		transaction.RoutingAttempts = attempt
		err = uc.transactionRepo.Update(transaction)
		if err != nil {
			logger.Error("Failed to update transaction attempts", logger.ErrorField(err))
		}

		if attemptResult.Success {
			// Success! Update transaction and return
			result.Success = true
			result.FinalSupplierID = supplier.ID
			result.TotalDuration = time.Since(startTime)

			logger.Info("Retry successful",
				logger.String("trx_id", transactionID),
				logger.String("supplier_code", supplier.Code),
				logger.Int("attempt", attempt),
				logger.Duration("total_duration", result.TotalDuration),
			)

			return result, nil
		}

		// Update supplier metrics
		uc.smartRoutingUC.UpdateSupplierMetrics(
			supplier.ID,
			false,
			int(attemptResult.Duration.Milliseconds()),
		)

		// If this is not the last attempt, wait before retrying
		if attempt < config.MaxAttempts {
			delay := uc.calculateRetryDelay(attempt, config)
			logger.Debug("Waiting before retry",
				logger.String("trx_id", transactionID),
				logger.Int("attempt", attempt),
				logger.Duration("delay", delay),
			)
			time.Sleep(delay)
		}
	}

	// All attempts failed - issue refund
	result.TotalDuration = time.Since(startTime)
	result.FinalError = fmt.Errorf("all retry attempts failed")

	refundErr := uc.issueRefund(transaction)
	if refundErr != nil {
		logger.Error("Failed to issue refund",
			logger.String("trx_id", transactionID),
			logger.ErrorField(refundErr),
		)
	} else {
		result.RefundIssued = true
	}

	logger.Warn("All retry attempts failed",
		logger.String("trx_id", transactionID),
		logger.Int("attempts_made", result.AttemptsMade),
		logger.Duration("total_duration", result.TotalDuration),
		logger.Bool("refund_issued", result.RefundIssued),
	)

	return result, nil
}

// canRetryTransaction checks if a transaction can be retried
func (uc *retryUsecase) canRetryTransaction(transaction *domain.Transaction, config *RetryConfig) bool {
	// Check if transaction is in a retryable state
	if transaction.Status != domain.StatusFailed && transaction.Status != domain.StatusTimeout {
		return false
	}

	// Check if we haven't exceeded max attempts
	if transaction.RoutingAttempts >= config.MaxAttempts {
		return false
	}

	// Check if transaction is not too old (optional - can be configured)
	maxAge := 24 * time.Hour // Default: don't retry transactions older than 24 hours
	if time.Since(transaction.CreatedAt) > maxAge {
		return false
	}

	return true
}

// getFailoverSuppliers gets suppliers for failover, excluding previously tried ones
func (uc *retryUsecase) getFailoverSuppliers(productID string, maxCount int) ([]*domain.Supplier, error) {
	// Get best suppliers using smart routing
	result, err := uc.smartRoutingUC.GetBestSupplier(productID, &RoutingCriteria{
		MaxSuppliers:   maxCount,
		PreferReliable: true,
		MinSuccessRate: 50.0,
	})
	if err != nil {
		return nil, err
	}

	// Combine selected supplier with alternatives
	suppliers := make([]*domain.Supplier, 0, len(result.Alternatives)+1)
	suppliers = append(suppliers, result.SelectedSupplier)
	suppliers = append(suppliers, result.Alternatives...)

	return suppliers, nil
}

// executeRetryAttempt executes a single retry attempt
func (uc *retryUsecase) executeRetryAttempt(
	transaction *domain.Transaction,
	supplier *domain.Supplier,
	attemptNumber int,
	config *RetryConfig,
) *RetryAttempt {
	startTime := time.Now()

	attempt := &RetryAttempt{
		AttemptNumber: attemptNumber,
		SupplierID:    supplier.ID,
		SupplierCode:  supplier.Code,
		StartTime:     startTime,
		Reason:        fmt.Sprintf("Retry attempt %d", attemptNumber),
	}

	logger.Info("Executing retry attempt",
		logger.String("trx_id", transaction.ID),
		logger.String("supplier_code", supplier.Code),
		logger.Int("attempt", attemptNumber),
	)

	// Update transaction with current supplier
	transaction.SupplierID = &supplier.ID
	transaction.Status = domain.StatusProcessing
	now := time.Now()
	transaction.ProcessedAt = &now

	err := uc.transactionRepo.Update(transaction)
	if err != nil {
		attempt.Error = fmt.Errorf("failed to update transaction: %w", err)
		attempt.EndTime = time.Now()
		attempt.Duration = attempt.EndTime.Sub(attempt.StartTime)
		return attempt
	}

	// Simulate supplier call (replace with actual supplier adapter)
	success, responseTimeMs, err := uc.simulateSupplierCall(supplier, transaction, config.TimeoutPerAttempt)

	attempt.EndTime = time.Now()
	attempt.Duration = attempt.EndTime.Sub(attempt.StartTime)
	attempt.ResponseTimeMs = responseTimeMs
	attempt.Success = success
	attempt.Error = err

	if success {
		// Update transaction to success
		serialNumber := utils.GenerateRandomString(12)
		sn := serialNumber
		msg := "Transaction successful via retry"

		transaction.Status = domain.StatusSuccess
		transaction.SerialNumber = &sn
		transaction.SupplierMessage = &msg
		transaction.FinalSupplierID = &supplier.ID
		completedAt := time.Now()
		transaction.CompletedAt = &completedAt

		err = uc.transactionRepo.Update(transaction)
		if err != nil {
			logger.Error("Failed to update successful transaction", logger.ErrorField(err))
		}

		// Update supplier metrics
		uc.smartRoutingUC.UpdateSupplierMetrics(supplier.ID, true, attempt.ResponseTimeMs)
	} else {
		// Update transaction to failed
		msg := fmt.Sprintf("Retry attempt %d failed: %v", attemptNumber, err)
		transaction.Status = domain.StatusFailed
		transaction.SupplierMessage = &msg
		completedAt := time.Now()
		transaction.CompletedAt = &completedAt

		err = uc.transactionRepo.Update(transaction)
		if err != nil {
			logger.Error("Failed to update failed transaction", logger.ErrorField(err))
		}
	}

	return attempt
}

// simulateSupplierCall simulates a supplier API call (replace with actual implementation)
func (uc *retryUsecase) simulateSupplierCall(supplier *domain.Supplier, transaction *domain.Transaction, timeout time.Duration) (bool, int, error) {
	// Simulate network delay
	delay := time.Duration(supplier.AvgResponseTimeMs) * time.Millisecond
	if delay == 0 {
		delay = 2 * time.Second
	}

	// Add some randomness
	delay += time.Duration(utils.GenerateRandomString(1)[0]) * 100 * time.Millisecond

	if delay > timeout {
		return false, int(timeout.Milliseconds()), fmt.Errorf("timeout")
	}

	time.Sleep(delay)

	// Simulate success rate based on supplier's actual success rate
	successChance := supplier.SuccessRate / 100.0
	random := float64(utils.GenerateRandomString(1)[0]) / 255.0

	if random < successChance {
		return true, int(delay.Milliseconds()), nil
	}

	return false, int(delay.Milliseconds()), fmt.Errorf("supplier error")
}

// calculateRetryDelay calculates delay between retry attempts with exponential backoff
func (uc *retryUsecase) calculateRetryDelay(attempt int, config *RetryConfig) time.Duration {
	multiplier := 1 << (attempt - 1) // This is integer bit shift
	delay := time.Duration(float64(config.InitialDelay) *
		float64(multiplier) * config.BackoffMultiplier)

	// Cap at max delay
	if delay > config.MaxDelay {
		delay = config.MaxDelay
	}

	// Add jitter if enabled
	if config.EnableJitter {
		jitter := time.Duration(float64(delay) * 0.1 * (float64(utils.GenerateRandomString(1)[0]) / 255.0))
		delay += jitter
	}

	return delay
}

// issueRefund issues a refund for a failed transaction
func (uc *retryUsecase) issueRefund(transaction *domain.Transaction) error {
	// Update transaction status to refund
	msg := "Auto refund after retry failure"
	transaction.Status = domain.StatusRefund
	transaction.SupplierMessage = &msg
	now := time.Now()
	transaction.CompletedAt = &now

	err := uc.transactionRepo.Update(transaction)
	if err != nil {
		return fmt.Errorf("failed to update transaction for refund: %w", err)
	}

	// TODO: Implement actual balance refund logic
	// This should create a mutation and update user balance

	logger.Info("Refund issued for failed transaction",
		logger.String("trx_id", transaction.ID),
		logger.String("trx_code", transaction.TrxCode),
		logger.Float64("amount", transaction.SellingPrice),
	)

	return nil
}

// GetRetryStatistics returns statistics about retry operations
func (uc *retryUsecase) GetRetryStatistics(startDate, endDate time.Time) (*RetryStatistics, error) {
	// Get failed transactions in date range
	failedTransactions, err := uc.transactionRepo.GetTransactionsByDateRange(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	stats := &RetryStatistics{
		TotalFailedTransactions: 0,
		RetriedTransactions:     0,
		SuccessfulRetries:       0,
		AverageRetryAttempts:    0,
		TotalRefundAmount:       0,
	}

	totalRetryAttempts := 0

	for _, trx := range failedTransactions {
		if trx.Status == domain.StatusFailed || trx.Status == domain.StatusTimeout {
			stats.TotalFailedTransactions++

			if trx.RoutingAttempts > 1 {
				stats.RetriedTransactions++
				totalRetryAttempts += trx.RoutingAttempts

				if trx.Status == domain.StatusSuccess {
					stats.SuccessfulRetries++
				} else if trx.Status == domain.StatusRefund {
					stats.TotalRefundAmount += trx.SellingPrice
				}
			}
		}
	}

	if stats.RetriedTransactions > 0 {
		stats.AverageRetryAttempts = float64(totalRetryAttempts) / float64(stats.RetriedTransactions)
		stats.RetrySuccessRate = float64(stats.SuccessfulRetries) / float64(stats.RetriedTransactions) * 100
	}

	return stats, nil
}

// RetryStatistics represents retry operation statistics
type RetryStatistics struct {
	TotalFailedTransactions int
	RetriedTransactions     int
	SuccessfulRetries       int
	AverageRetryAttempts    float64
	TotalRefundAmount       float64
	RetrySuccessRate        float64
}

// ProcessFailedTransactions processes all failed transactions that are eligible for retry
func (uc *retryUsecase) ProcessFailedTransactions(config *RetryConfig) ([]*RetryResult, error) {
	// Get all failed transactions
	failedTransactions, err := uc.transactionRepo.GetByStatus(domain.StatusFailed)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed transactions: %w", err)
	}

	results := make([]*RetryResult, 0)

	for _, transaction := range failedTransactions {
		if uc.canRetryTransaction(transaction, config) {
			result, err := uc.RetryTransaction(transaction.ID, config)
			if err != nil {
				logger.Error("Failed to retry transaction",
					logger.String("trx_id", transaction.ID),
					logger.ErrorField(err),
				)
				continue
			}
			results = append(results, result)
		}
	}

	logger.Info("Processed failed transactions for retry",
		logger.Int("total_failed", len(failedTransactions)),
		logger.Int("retried", len(results)),
	)

	return results, nil
}
