package usecase

import (
	"fmt"
	"time"

	"github.com/alfanzaky/eraflazz/internal/domain"
	"github.com/alfanzaky/eraflazz/pkg/logger"
	"github.com/alfanzaky/eraflazz/pkg/utils"
)

type transactionUsecase struct {
	userRepo        domain.UserRepository
	productRepo     domain.ProductRepository
	supplierRepo    domain.SupplierRepository
	transactionRepo domain.TransactionRepository
	mutationRepo    domain.MutationRepository
	cacheRepo       interface{} // Will be implemented as Redis cache
	queueRepo       domain.QueueRepository
	smartRoutingUC  *smartRoutingUsecase
	adapterFactory  domain.SupplierAdapterFactory
	retryUC         *retryUsecase
}

// NewTransactionUsecase creates a new transaction use case
func NewTransactionUsecase(
	userRepo domain.UserRepository,
	productRepo domain.ProductRepository,
	supplierRepo domain.SupplierRepository,
	transactionRepo domain.TransactionRepository,
	mutationRepo domain.MutationRepository,
	smartRoutingUC *smartRoutingUsecase,
	adapterFactory domain.SupplierAdapterFactory,
	retryUC *retryUsecase,
	queueRepo domain.QueueRepository,
) domain.TransactionUsecase {
	return &transactionUsecase{
		userRepo:        userRepo,
		productRepo:     productRepo,
		supplierRepo:    supplierRepo,
		transactionRepo: transactionRepo,
		mutationRepo:    mutationRepo,
		queueRepo:       queueRepo,
		smartRoutingUC:  smartRoutingUC,
		adapterFactory:  adapterFactory,
		retryUC:         retryUC,
	}
}

// CreateTransaction creates a new transaction
func (uc *transactionUsecase) CreateTransaction(userID, productCode, destinationNumber string) (*domain.Transaction, error) {
	// Validate input
	if userID == "" || productCode == "" || destinationNumber == "" {
		return nil, fmt.Errorf("missing required fields")
	}

	// Validate phone number
	if !utils.ValidatePhoneNumber(destinationNumber) {
		return nil, fmt.Errorf("invalid phone number format")
	}

	// Get user
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		logger.Error("Failed to get user for transaction",
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is not active")
	}

	// Get product
	product, err := uc.productRepo.GetByCode(productCode)
	if err != nil {
		logger.Error("Failed to get product for transaction",
			logger.String("product_code", productCode),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Check if product is active
	if !product.IsActive {
		return nil, fmt.Errorf("product is not available")
	}

	// Calculate pricing
	basePrice := product.BasePrice
	sellingPrice := user.GetEffectivePrice(basePrice)

	// Check transaction limits
	if sellingPrice < product.MinPrice || sellingPrice > product.MaxTransactionAmount {
		return nil, fmt.Errorf("price out of allowed range")
	}

	// Check user balance
	if !user.HasSufficientBalance(sellingPrice) {
		return nil, fmt.Errorf("insufficient balance")
	}

	// Create transaction
	transaction := &domain.Transaction{
		ID:                utils.GenerateUUID(),
		TrxCode:           utils.GenerateTrxCode(),
		UserID:            userID,
		ProductID:         product.ID,
		DestinationNumber: utils.ParsePhoneNumber(destinationNumber),
		ProductCode:       productCode,
		HPP:               basePrice,
		SellingPrice:      sellingPrice,
		AdminFee:          0, // Can be calculated based on business rules
		Status:            domain.StatusPending,
		RoutingAttempts:   0,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Save transaction
	err = uc.transactionRepo.Create(transaction)
	if err != nil {
		logger.Error("Failed to create transaction",
			logger.String("trx_code", transaction.TrxCode),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Enqueue transaction for processing
	if uc.queueRepo != nil {
		err = uc.queueRepo.EnqueueTransaction(transaction.ID)
		if err != nil {
			logger.Error("Failed to enqueue transaction",
				logger.String("trx_id", transaction.ID),
				logger.String("trace_id", transaction.TrxCode),
				logger.ErrorField(err),
			)
		} else {
			logger.Debug("Transaction queued for processing",
				logger.String("trx_id", transaction.ID),
				logger.String("trace_id", transaction.TrxCode),
			)
		}
	} else {
		logger.Warn("Queue repository is not configured; transaction will not be auto-processed",
			logger.String("trx_id", transaction.ID),
			logger.String("trace_id", transaction.TrxCode),
		)
	}

	logger.Info("Transaction created successfully",
		logger.String("trace_id", transaction.TrxCode),
		logger.String("trx_id", transaction.ID),
		logger.String("user_id", userID),
		logger.String("product_code", productCode),
		logger.Float64("amount", sellingPrice),
	)

	return transaction, nil
}

// ProcessTransaction processes a pending transaction
func (uc *transactionUsecase) ProcessTransaction(transactionID string) error {
	// Get transaction
	transaction, err := uc.transactionRepo.GetByID(transactionID)
	if err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}

	// Check if transaction is in pending status
	if transaction.Status != domain.StatusPending {
		return fmt.Errorf("transaction is not in pending status")
	}

	// Update status to processing
	now := time.Now()
	transaction.ProcessedAt = &now
	err = uc.transactionRepo.UpdateStatus(transactionID, domain.StatusProcessing)
	if err != nil {
		return fmt.Errorf("failed to update processing status: %w", err)
	}

	logger.Info("Processing transaction",
		logger.String("trace_id", transaction.TrxCode),
		logger.String("trx_id", transaction.ID),
		logger.Float64("amount", transaction.SellingPrice),
	)

	// Get user for balance check
	user, err := uc.userRepo.GetByID(transaction.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Check balance again (in case it changed)
	if !user.HasSufficientBalance(transaction.SellingPrice) {
		// Update transaction to failed due to insufficient balance
		msg := "Insufficient balance"
		transaction.Status = domain.StatusFailed
		transaction.SupplierMessage = &msg
		err = uc.transactionRepo.Update(transaction)
		if err != nil {
			logger.Error("Failed to update transaction status", logger.ErrorField(err))
		}
		return fmt.Errorf("insufficient balance")
	}

	selectedSupplier, selectedMapping, err := uc.selectSupplier(transaction)
	if err != nil {
		logger.Error("Failed to select supplier",
			logger.String("trx_id", transaction.ID),
			logger.String("trace_id", transaction.TrxCode),
			logger.ErrorField(err),
		)
		return uc.handleSupplierFailure(transaction, fmt.Sprintf("routing error: %v", err))
	}

	logger.Info("Supplier selected",
		logger.String("trace_id", transaction.TrxCode),
		logger.String("trx_id", transaction.ID),
		logger.String("supplier_code", selectedSupplier.Code),
		logger.String("mapping_code", selectedMapping.SupplierProductCode),
	)

	supplierID := selectedSupplier.ID
	transaction.SupplierID = &supplierID

	// Deduct balance (create mutation)
	refType := domain.ReferenceTypeTransaction
	err = uc.createBalanceMutation(
		user.ID,
		domain.MutationTypeCredit, // Credit = money out
		transaction.SellingPrice,
		user.Balance,
		user.Balance-transaction.SellingPrice,
		fmt.Sprintf("Pembelian %s %s", transaction.ProductCode, transaction.DestinationNumber),
		&refType,
		&transaction.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to create balance mutation: %w", err)
	}

	// Update user balance
	newBalance := user.Balance - transaction.SellingPrice
	err = uc.userRepo.UpdateBalance(user.ID, newBalance)
	if err != nil {
		logger.Error("Failed to update user balance", logger.ErrorField(err))
		// Continue processing even if balance update fails
		// Will be handled by reconciliation
	}

	return uc.executeSupplierTransaction(transaction, selectedSupplier, selectedMapping)
}

// ProcessPendingTransactions processes all pending transactions
func (uc *transactionUsecase) ProcessPendingTransactions() error {
	// Get all pending transactions
	pendingTransactions, err := uc.transactionRepo.GetPendingTransactions()
	if err != nil {
		return fmt.Errorf("failed to get pending transactions: %w", err)
	}

	logger.Info("Processing pending transactions",
		logger.Int("count", len(pendingTransactions)),
	)

	// Process each transaction
	for _, transaction := range pendingTransactions {
		err := uc.ProcessTransaction(transaction.ID)
		if err != nil {
			logger.Error("Failed to process transaction",
				logger.String("trx_id", transaction.ID),
				logger.ErrorField(err),
			)
			// Continue processing other transactions
		}
	}

	return nil
}

func (uc *transactionUsecase) selectSupplier(transaction *domain.Transaction) (*domain.Supplier, *domain.ProductMapping, error) {
	if uc.smartRoutingUC == nil {
		return nil, nil, fmt.Errorf("smart routing is not configured")
	}

	result, err := uc.smartRoutingUC.GetBestSupplier(transaction.ProductID, nil)
	if err != nil {
		return nil, nil, err
	}

	if result == nil || result.SelectedSupplier == nil || result.SelectedMapping == nil {
		return nil, nil, fmt.Errorf("no supplier available for product %s", transaction.ProductID)
	}

	return result.SelectedSupplier, result.SelectedMapping, nil
}

func (uc *transactionUsecase) executeSupplierTransaction(
	transaction *domain.Transaction,
	supplier *domain.Supplier,
	mapping *domain.ProductMapping,
) error {
	if uc.adapterFactory == nil {
		return uc.handleSupplierFailure(transaction, "supplier adapter factory not configured")
	}

	adapter, err := uc.adapterFactory.GetAdapter(supplier.Code)
	if err != nil {
		return uc.handleSupplierFailure(transaction, fmt.Sprintf("adapter for %s not found: %v", supplier.Code, err))
	}

	request := &domain.SupplierRequest{
		ProductCode:       mapping.SupplierProductCode,
		DestinationNumber: transaction.DestinationNumber,
		RefID:             transaction.TrxCode,
	}

	logger.Info("Calling supplier",
		logger.String("trace_id", transaction.TrxCode),
		logger.String("trx_id", transaction.ID),
		logger.String("supplier_code", supplier.Code),
		logger.String("product_code", mapping.SupplierProductCode),
	)

	start := time.Now()
	response, err := adapter.TopUp(request)
	duration := time.Since(start)

	success := err == nil && response != nil && response.Success
	responseTime := int(duration.Milliseconds())
	if response != nil && response.ResponseTime > 0 {
		responseTime = response.ResponseTime
	}

	if uc.smartRoutingUC != nil {
		if updateErr := uc.smartRoutingUC.UpdateSupplierMetrics(supplier.ID, success, responseTime); updateErr != nil {
			logger.Warn("Failed to update supplier metrics",
				logger.String("supplier_id", supplier.ID),
				logger.ErrorField(updateErr),
			)
		}
	}

	if err != nil {
		return uc.handleSupplierFailure(transaction, fmt.Sprintf("supplier error: %v", err))
	}

	if !response.Success {
		msg := response.Message
		if msg == "" {
			msg = "supplier returned failure"
		}
		return uc.handleSupplierFailure(transaction, msg)
	}

	serial := response.SerialNumber
	if serial == "" {
		serial = response.TrxID
	}
	if serial != "" {
		transaction.SerialNumber = &serial
	}

	msg := response.Message
	if msg != "" {
		transaction.SupplierMessage = &msg
	}

	if response.TrxID != "" {
		supplierTrxID := response.TrxID
		transaction.SupplierTrxID = &supplierTrxID
	}

	transaction.Status = domain.StatusSuccess
	transaction.FinalSupplierID = &supplier.ID
	now := time.Now()
	transaction.CompletedAt = &now

	if err := uc.transactionRepo.Update(transaction); err != nil {
		return fmt.Errorf("failed to update successful transaction: %w", err)
	}

	logger.Info("Transaction completed via supplier",
		logger.String("trace_id", transaction.TrxCode),
		logger.String("trx_id", transaction.ID),
		logger.String("supplier_code", supplier.Code),
		logger.Duration("duration", duration),
		logger.Int("response_time_ms", responseTime),
	)

	return nil
}

func (uc *transactionUsecase) handleSupplierFailure(transaction *domain.Transaction, reason string) error {
	msg := reason
	transaction.Status = domain.StatusFailed
	transaction.SupplierMessage = &msg
	now := time.Now()
	transaction.CompletedAt = &now

	if err := uc.transactionRepo.Update(transaction); err != nil {
		logger.Error("Failed to update failed transaction", logger.ErrorField(err))
	}

	logger.Warn("Supplier failure",
		logger.String("trace_id", transaction.TrxCode),
		logger.String("trx_id", transaction.ID),
		logger.String("reason", reason),
	)

	if uc.retryUC != nil {
		result, err := uc.retryUC.RetryTransaction(transaction.ID, nil)
		if err == nil {
			if result != nil {
				if result.Success {
					return nil
				}
				if result.RefundIssued {
					return nil
				}
			}
		} else {
			logger.Error("Retry transaction failed", logger.ErrorField(err))
		}
	}

	if err := uc.refundTransaction(transaction); err != nil {
		return fmt.Errorf("failed to refund transaction after supplier failure: %w", err)
	}

	return fmt.Errorf("supplier failure: %s", reason)
}

// RetryFailedTransaction retries a failed transaction
func (uc *transactionUsecase) RetryFailedTransaction(transactionID string) error {
	// Get transaction
	transaction, err := uc.transactionRepo.GetByID(transactionID)
	if err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}

	// Check if transaction can be retried
	if !transaction.CanRetry() {
		return fmt.Errorf("transaction cannot be retried")
	}

	// Increment routing attempts
	transaction.RoutingAttempts++
	err = uc.transactionRepo.Update(transaction)
	if err != nil {
		return fmt.Errorf("failed to increment routing attempts: %w", err)
	}

	// Reset status to pending
	err = uc.transactionRepo.UpdateStatus(transactionID, domain.StatusPending)
	if err != nil {
		return fmt.Errorf("failed to reset transaction status: %w", err)
	}

	// Process transaction again
	return uc.ProcessTransaction(transactionID)
}

// GetTransaction retrieves a transaction by ID
func (uc *transactionUsecase) GetTransaction(id string) (*domain.Transaction, error) {
	return uc.transactionRepo.GetByID(id)
}

// GetUserTransactions retrieves user transactions with pagination
func (uc *transactionUsecase) GetUserTransactions(userID string, page, limit int) ([]*domain.Transaction, error) {
	offset := (page - 1) * limit
	return uc.transactionRepo.GetByUserID(userID, limit, offset)
}

// GetTransactionByTrxCode retrieves a transaction by transaction code
func (uc *transactionUsecase) GetTransactionByTrxCode(trxCode string) (*domain.Transaction, error) {
	return uc.transactionRepo.GetByTrxCode(trxCode)
}

// CancelTransaction cancels a transaction
func (uc *transactionUsecase) CancelTransaction(transactionID string) error {
	// Get transaction
	transaction, err := uc.transactionRepo.GetByID(transactionID)
	if err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}

	// Can only cancel pending transactions
	if transaction.Status != domain.StatusPending {
		return fmt.Errorf("cannot cancel transaction in %s status", transaction.Status)
	}

	// Update status to failed
	msg := "Transaction cancelled by user"
	transaction.Status = domain.StatusFailed
	transaction.SupplierMessage = &msg
	err = uc.transactionRepo.Update(transaction)
	if err != nil {
		return fmt.Errorf("failed to cancel transaction: %w", err)
	}

	// Refund balance if already deducted
	if transaction.Status == domain.StatusProcessing {
		err = uc.refundTransaction(transaction)
		if err != nil {
			logger.Error("Failed to refund cancelled transaction", logger.ErrorField(err))
		}
	}

	return nil
}

// RefundTransaction refunds a failed transaction
func (uc *transactionUsecase) RefundTransaction(transactionID string) error {
	// Get transaction
	transaction, err := uc.transactionRepo.GetByID(transactionID)
	if err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}

	return uc.refundTransaction(transaction)
}

// GetTransactionStats gets transaction statistics for a user
func (uc *transactionUsecase) GetTransactionStats(userID string, startDate, endDate time.Time) (*domain.TransactionStats, error) {
	// Get transactions in date range
	transactions, err := uc.transactionRepo.GetTransactionsByDateRange(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	// Filter by user and calculate stats
	stats := &domain.TransactionStats{}
	var totalAmount float64

	for _, trx := range transactions {
		if trx.UserID == userID {
			stats.TotalTransactions++
			totalAmount += trx.SellingPrice

			switch trx.Status {
			case domain.StatusSuccess:
				stats.SuccessCount++
				stats.TotalRevenue += trx.SellingPrice
				stats.TotalProfit += trx.Profit
			case domain.StatusFailed:
				stats.FailedCount++
			case domain.StatusPending:
				stats.PendingCount++
			}
		}
	}

	// Calculate averages
	if stats.TotalTransactions > 0 {
		stats.AverageAmount = totalAmount / float64(stats.TotalTransactions)
	}

	return stats, nil
}

// Helper functions

func (uc *transactionUsecase) createBalanceMutation(
	userID, mutationType string, amount, balanceBefore, balanceAfter float64,
	description string, referenceType *string, referenceID *string,
) error {
	if uc.mutationRepo == nil {
		return fmt.Errorf("mutation repository is not configured")
	}

	mutation := &domain.Mutation{
		ID:            utils.GenerateUUID(),
		UserID:        userID,
		Type:          mutationType,
		Amount:        amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		Description:   description,
		ReferenceType: referenceType,
		ReferenceID:   referenceID,
		CreatedAt:     time.Now(),
	}

	if err := uc.mutationRepo.Create(mutation); err != nil {
		return fmt.Errorf("failed to create mutation: %w", err)
	}

	logger.Debug("Balance mutation persisted",
		logger.String("user_id", userID),
		logger.String("type", mutationType),
		logger.Float64("amount", amount),
	)

	return nil
}

func (uc *transactionUsecase) refundTransaction(transaction *domain.Transaction) error {
	// Get user
	user, err := uc.userRepo.GetByID(transaction.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user for refund: %w", err)
	}

	// Create refund mutation
	refType := domain.ReferenceTypeTransaction
	err = uc.createBalanceMutation(
		user.ID,
		domain.MutationTypeDebit, // Debit = money in (refund)
		transaction.SellingPrice,
		user.Balance,
		user.Balance+transaction.SellingPrice,
		fmt.Sprintf("Refund transaksi gagal %s", transaction.TrxCode),
		&refType,
		&transaction.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to create refund mutation: %w", err)
	}

	// Update user balance
	newBalance := user.Balance + transaction.SellingPrice
	err = uc.userRepo.UpdateBalance(user.ID, newBalance)
	if err != nil {
		logger.Error("Failed to update user balance for refund", logger.ErrorField(err))
	}

	// Update transaction status
	msg := "Transaction refunded due to failure"
	transaction.Status = domain.StatusRefund
	transaction.SupplierMessage = &msg
	now := time.Now()
	transaction.CompletedAt = &now
	err = uc.transactionRepo.Update(transaction)
	if err != nil {
		logger.Error("Failed to update transaction status for refund", logger.ErrorField(err))
	}

	logger.Info("Transaction refunded successfully",
		logger.String("trx_id", transaction.ID),
		logger.String("trx_code", transaction.TrxCode),
		logger.Float64("amount", transaction.SellingPrice),
	)

	return nil
}

func (uc *transactionUsecase) simulateSupplierCall(transaction *domain.Transaction) error {
	// Simulate API call delay
	time.Sleep(2 * time.Second)

	// Simulate success (90% success rate)
	if time.Now().UnixNano()%10 < 9 {
		// Success
		serialNumber := utils.GenerateRandomString(12)
		sn := serialNumber
		msg := "Transaction successful"
		now := time.Now()

		transaction.Status = domain.StatusSuccess
		transaction.SerialNumber = &sn
		transaction.SupplierMessage = &msg
		transaction.CompletedAt = &now

		err := uc.transactionRepo.Update(transaction)
		if err != nil {
			return fmt.Errorf("failed to update successful transaction: %w", err)
		}

		logger.Info("Transaction completed successfully",
			logger.String("trx_id", transaction.ID),
			logger.String("serial_number", serialNumber),
		)

		return nil
	} else {
		// Failure
		msg := "Supplier error: timeout"
		transaction.Status = domain.StatusFailed
		transaction.SupplierMessage = &msg
		now := time.Now()
		transaction.CompletedAt = &now

		err := uc.transactionRepo.Update(transaction)
		if err != nil {
			return fmt.Errorf("failed to update failed transaction: %w", err)
		}

		return fmt.Errorf("supplier call failed")
	}
}
