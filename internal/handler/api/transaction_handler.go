package api

import (
	"strconv"
	"time"

	"github.com/alfanzaky/eraflazz/internal/domain"
	"github.com/alfanzaky/eraflazz/pkg/logger"
	"github.com/alfanzaky/eraflazz/pkg/metrics"
	"github.com/alfanzaky/eraflazz/pkg/observability"
	"github.com/alfanzaky/eraflazz/pkg/xresponse"
	"github.com/gin-gonic/gin"
)

// TransactionHandler handles transaction-related HTTP requests
type TransactionHandler struct {
	transactionUC domain.TransactionUsecase
	roleGuard     *RoleGuard
}

// NewTransactionHandler creates a new transaction handler
func NewTransactionHandler(transactionUC domain.TransactionUsecase) *TransactionHandler {
	return &TransactionHandler{
		transactionUC: transactionUC,
		roleGuard:     NewRoleGuard(),
	}
}

// CreateTransactionRequest represents request for creating transaction
type CreateTransactionRequest struct {
	ProductCode       string  `json:"product_code" binding:"required"`
	DestinationNumber string  `json:"destination_number" binding:"required"`
	CustomerNotes     *string `json:"customer_notes,omitempty"`
}

// TransactionResponse represents response for transaction
type TransactionResponse struct {
	ID                string  `json:"id"`
	TrxCode           string  `json:"trx_code"`
	UserID            string  `json:"user_id"`
	ProductCode       string  `json:"product_code"`
	DestinationNumber string  `json:"destination_number"`
	HPP               float64 `json:"hpp"`
	SellingPrice      float64 `json:"selling_price"`
	AdminFee          float64 `json:"admin_fee"`
	Profit            float64 `json:"profit"`
	Status            string  `json:"status"`
	SerialNumber      *string `json:"serial_number,omitempty"`
	SupplierMessage   *string `json:"supplier_message,omitempty"`
	CreatedAt         string  `json:"created_at"`
	ProcessedAt       *string `json:"processed_at,omitempty"`
	CompletedAt       *string `json:"completed_at,omitempty"`
}

// CreateTransaction creates a new transaction
func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request body", logger.ErrorField(err))
		xresponse.BadRequest(c, "Invalid request format")
		return
	}

	// Check if user or H2H client is authenticated
	userID, _, _, exists := h.roleGuard.GetCurrentUser(c)
	if !exists {
		// Check if it's an H2H client
		if clientID, isH2H := GetClientIDFromContext(c); isH2H {
			userID = clientID
		} else {
			xresponse.Unauthorized(c, "Authentication required")
			return
		}
	}

	// Log the access attempt
	h.roleGuard.LogAccess(c, "create_transaction", req.ProductCode)

	// Create transaction
	transaction, err := h.transactionUC.CreateTransaction(userID, req.ProductCode, req.DestinationNumber)
	if err != nil {
		logger.Error("Failed to create transaction",
			logger.String("user_id", userID),
			logger.String("product_code", req.ProductCode),
			logger.ErrorField(err),
		)

		// Handle specific error types
		switch err.Error() {
		case "user not found":
			xresponse.UserNotFound(c, "User account not found")
		case "product not found":
			xresponse.InvalidProduct(c, "Product not found or unavailable")
		case "insufficient balance":
			xresponse.InsufficientBalance(c, "Insufficient balance for this transaction")
		case "invalid phone number format":
			xresponse.BadRequest(c, "Invalid phone number format")
		default:
			xresponse.InternalServerError(c, "Failed to create transaction")
		}
		return
	}

	// Record transaction metrics
	userRole := "anonymous"
	if role, exists := c.Get("user_role"); exists {
		if roleStr, ok := role.(string); ok {
			userRole = roleStr
		}
	} else if _, exists := c.Get("client_id"); exists {
		userRole = "h2h"
	}

	// Record transaction metrics
	metrics.RecordTransaction(
		transaction.Status,
		"unknown", // TODO: Get product category from product service
		userRole,
		transaction.SellingPrice,
	)

	// Add customer notes if provided
	if req.CustomerNotes != nil {
		observability.LogWithFields(c, "Customer notes added",
			logger.String("trx_id", transaction.ID),
			logger.String("notes", *req.CustomerNotes),
		)
	}

	response := TransactionResponse{
		ID:                transaction.ID,
		TrxCode:           transaction.TrxCode,
		UserID:            transaction.UserID,
		ProductCode:       transaction.ProductCode,
		DestinationNumber: transaction.DestinationNumber,
		HPP:               transaction.HPP,
		SellingPrice:      transaction.SellingPrice,
		AdminFee:          transaction.AdminFee,
		Profit:            transaction.CalculateProfit(),
		Status:            transaction.Status,
		CreatedAt:         transaction.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if transaction.ProcessedAt != nil {
		processedAt := transaction.ProcessedAt.Format("2006-01-02 15:04:05")
		response.ProcessedAt = &processedAt
	}

	if transaction.CompletedAt != nil {
		completedAt := transaction.CompletedAt.Format("2006-01-02 15:04:05")
		response.CompletedAt = &completedAt
	}

	logger.Info("Transaction created via API",
		logger.String("trx_id", transaction.ID),
		logger.String("trx_code", transaction.TrxCode),
		logger.String("user_id", userID),
	)

	xresponse.Created(c, "Transaction created successfully", response)
}

// GetTransaction retrieves a transaction by ID
func (h *TransactionHandler) GetTransaction(c *gin.Context) {
	trxID := c.Param("id")
	if trxID == "" {
		xresponse.BadRequest(c, "Transaction ID is required")
		return
	}

	// Get authenticated user or H2H client
	userID, _, _, exists := h.roleGuard.GetCurrentUser(c)
	if !exists {
		// Check if it's an H2H client
		if clientID, isH2H := GetClientIDFromContext(c); isH2H {
			userID = clientID
		} else {
			xresponse.Unauthorized(c, "Authentication required")
			return
		}
	}

	h.roleGuard.LogAccess(c, "get_transaction", trxID)

	// Get transaction
	transaction, err := h.transactionUC.GetTransaction(trxID)
	if err != nil {
		logger.Error("Failed to get transaction",
			logger.String("trx_id", trxID),
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)

		if err.Error() == "transaction not found" {
			xresponse.NotFound(c, "Transaction not found")
		} else {
			xresponse.InternalServerError(c, "Failed to retrieve transaction")
		}
		return
	}

	// Check access permissions using role guard
	if !h.roleGuard.CanAccessOwnData(c, transaction.UserID) {
		xresponse.Forbidden(c, "Access denied to this transaction")
		return
	}

	response := h.buildTransactionResponse(transaction)

	xresponse.Success(c, "Transaction retrieved successfully", response)
}

// GetTransactionByCode retrieves a transaction by transaction code
func (h *TransactionHandler) GetTransactionByCode(c *gin.Context) {
	trxCode := c.Param("code")
	if trxCode == "" {
		xresponse.BadRequest(c, "Transaction code is required")
		return
	}

	// Get authenticated user or H2H client
	userID, _, _, exists := h.roleGuard.GetCurrentUser(c)
	if !exists {
		// Check if it's an H2H client
		if clientID, isH2H := GetClientIDFromContext(c); isH2H {
			userID = clientID
		} else {
			xresponse.Unauthorized(c, "Authentication required")
			return
		}
	}

	h.roleGuard.LogAccess(c, "get_transaction_by_code", trxCode)

	// Get transaction
	transaction, err := h.transactionUC.GetTransactionByTrxCode(trxCode)
	if err != nil {
		logger.Error("Failed to get transaction by code",
			logger.String("trx_code", trxCode),
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)

		if err.Error() == "transaction not found" {
			xresponse.NotFound(c, "Transaction not found")
		} else {
			xresponse.InternalServerError(c, "Failed to retrieve transaction")
		}
		return
	}

	// Check access permissions using role guard
	if !h.roleGuard.CanAccessOwnData(c, transaction.UserID) {
		xresponse.Forbidden(c, "Access denied to this transaction")
		return
	}

	response := h.buildTransactionResponse(transaction)

	xresponse.Success(c, "Transaction retrieved successfully", response)
}

// GetUserTransactions retrieves user transactions with pagination
func (h *TransactionHandler) GetUserTransactions(c *gin.Context) {
	// Get pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	// Get authenticated user or H2H client
	userID, _, _, exists := h.roleGuard.GetCurrentUser(c)
	if !exists {
		// Check if it's an H2H client
		if clientID, isH2H := GetClientIDFromContext(c); isH2H {
			userID = clientID
		} else {
			xresponse.Unauthorized(c, "Authentication required")
			return
		}
	}

	h.roleGuard.LogAccess(c, "get_user_transactions", "own_transactions")

	// Get transactions
	transactions, err := h.transactionUC.GetUserTransactions(userID, page, limit)
	if err != nil {
		logger.Error("Failed to get user transactions",
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)
		xresponse.InternalServerError(c, "Failed to retrieve transactions")
		return
	}

	// Build response
	responses := make([]TransactionResponse, len(transactions))
	for i, trx := range transactions {
		responses[i] = h.buildTransactionResponse(trx)
	}

	xresponse.Success(c, "Transactions retrieved successfully", responses)
}

// CancelTransaction cancels a pending transaction
func (h *TransactionHandler) CancelTransaction(c *gin.Context) {
	trxID := c.Param("id")
	if trxID == "" {
		xresponse.BadRequest(c, "Transaction ID is required")
		return
	}

	// Get authenticated user or H2H client
	userID, _, _, exists := h.roleGuard.GetCurrentUser(c)
	if !exists {
		// Check if it's an H2H client
		if clientID, isH2H := GetClientIDFromContext(c); isH2H {
			userID = clientID
		} else {
			xresponse.Unauthorized(c, "Authentication required")
			return
		}
	}

	h.roleGuard.LogAccess(c, "cancel_transaction", trxID)

	// Get transaction first to check ownership
	transaction, err := h.transactionUC.GetTransaction(trxID)
	if err != nil {
		if err.Error() == "transaction not found" {
			xresponse.NotFound(c, "Transaction not found")
		} else {
			xresponse.InternalServerError(c, "Failed to retrieve transaction")
		}
		return
	}

	// Check access permissions using role guard
	if !h.roleGuard.CanAccessOwnData(c, transaction.UserID) {
		xresponse.Forbidden(c, "Access denied to this transaction")
		return
	}

	// Cancel transaction
	err = h.transactionUC.CancelTransaction(trxID)
	if err != nil {
		logger.Error("Failed to cancel transaction",
			logger.String("trx_id", trxID),
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)

		if err.Error() == "cannot cancel transaction in "+transaction.Status {
			xresponse.BadRequest(c, "Cannot cancel transaction in "+transaction.Status+" status")
		} else {
			xresponse.InternalServerError(c, "Failed to cancel transaction")
		}
		return
	}

	logger.Info("Transaction cancelled via API",
		logger.String("trx_id", trxID),
		logger.String("user_id", userID),
	)

	xresponse.Success(c, "Transaction cancelled successfully", nil)
}

// GetTransactionStats retrieves transaction statistics for the user
func (h *TransactionHandler) GetTransactionStats(c *gin.Context) {
	// Get date range parameters
	startDateStr := c.DefaultQuery("start_date", "")
	endDateStr := c.DefaultQuery("end_date", "")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			xresponse.BadRequest(c, "Invalid start_date format. Use YYYY-MM-DD")
			return
		}
	} else {
		startDate = time.Now().AddDate(0, -1, 0) // Default to 1 month ago
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			xresponse.BadRequest(c, "Invalid end_date format. Use YYYY-MM-DD")
			return
		}
		// Set end date to end of day
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	} else {
		endDate = time.Now()
	}

	// Get authenticated user or H2H client
	userID, _, _, exists := h.roleGuard.GetCurrentUser(c)
	if !exists {
		// Check if it's an H2H client
		if clientID, isH2H := GetClientIDFromContext(c); isH2H {
			userID = clientID
		} else {
			xresponse.Unauthorized(c, "Authentication required")
			return
		}
	}

	h.roleGuard.LogAccess(c, "get_transaction_stats", "own_stats")

	// Get statistics
	stats, err := h.transactionUC.GetTransactionStats(userID, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get transaction stats",
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)
		xresponse.InternalServerError(c, "Failed to retrieve statistics")
		return
	}

	xresponse.Success(c, "Statistics retrieved successfully", stats)
}

// buildTransactionResponse builds transaction response from domain model
func (h *TransactionHandler) buildTransactionResponse(trx *domain.Transaction) TransactionResponse {
	response := TransactionResponse{
		ID:                trx.ID,
		TrxCode:           trx.TrxCode,
		UserID:            trx.UserID,
		ProductCode:       trx.ProductCode,
		DestinationNumber: trx.DestinationNumber,
		HPP:               trx.HPP,
		SellingPrice:      trx.SellingPrice,
		AdminFee:          trx.AdminFee,
		Profit:            trx.CalculateProfit(),
		Status:            trx.Status,
		SerialNumber:      trx.SerialNumber,
		SupplierMessage:   trx.SupplierMessage,
		CreatedAt:         trx.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if trx.ProcessedAt != nil {
		processedAt := trx.ProcessedAt.Format("2006-01-02 15:04:05")
		response.ProcessedAt = &processedAt
	}

	if trx.CompletedAt != nil {
		completedAt := trx.CompletedAt.Format("2006-01-02 15:04:05")
		response.CompletedAt = &completedAt
	}

	return response
}
