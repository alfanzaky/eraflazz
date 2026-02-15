package xresponse

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response represents standard API response format
type Response struct {
	Code      int         `json:"code"`
	Status    string      `json:"status"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// ErrorResponse represents error response format
type ErrorResponse struct {
	Code      int         `json:"code"`
	Status    string      `json:"status"`
	ErrorCode string      `json:"error_code"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// PaginatedResponse represents paginated response
type PaginatedResponse struct {
	Code       int           `json:"code"`
	Status     string        `json:"status"`
	Message    string        `json:"message"`
	Data       interface{}   `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
	Timestamp  int64         `json:"timestamp"`
}

// Common error codes
const (
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeUnauthorized     = "UNAUTHORIZED"
	ErrCodeForbidden        = "FORBIDDEN"
	ErrCodeConflict         = "CONFLICT"
	ErrCodeInternalError    = "INTERNAL_ERROR"
	ErrCodeInsufficientBalance = "INSUFFICIENT_BALANCE"
	ErrCodeInvalidProduct   = "INVALID_PRODUCT"
	ErrCodeSupplierError    = "SUPPLIER_ERROR"
	ErrCodeTransactionFailed = "TRANSACTION_FAILED"
	ErrCodeUserNotFound     = "USER_NOT_FOUND"
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeAccountLocked    = "ACCOUNT_LOCKED"
	ErrCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
)

// Success sends success response
func Success(c *gin.Context, message string, data interface{}) {
	response := Response{
		Code:      http.StatusOK,
		Status:    "success",
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
	c.JSON(http.StatusOK, response)
}

// SuccessWithCode sends success response with custom status code
func SuccessWithCode(c *gin.Context, statusCode int, message string, data interface{}) {
	response := Response{
		Code:      statusCode,
		Status:    "success",
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
	c.JSON(statusCode, response)
}

// Created sends created response (201)
func Created(c *gin.Context, message string, data interface{}) {
	response := Response{
		Code:      http.StatusCreated,
		Status:    "success",
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
	c.JSON(http.StatusCreated, response)
}

// Error sends error response
func Error(c *gin.Context, statusCode int, errorCode, message string) {
	response := ErrorResponse{
		Code:      statusCode,
		Status:    "error",
		ErrorCode: errorCode,
		Message:   message,
		Timestamp: time.Now().Unix(),
	}
	c.JSON(statusCode, response)
}

// ErrorWithDetails sends error response with details
func ErrorWithDetails(c *gin.Context, statusCode int, errorCode, message string, details interface{}) {
	response := ErrorResponse{
		Code:      statusCode,
		Status:    "error",
		ErrorCode: errorCode,
		Message:   message,
		Details:   details,
		Timestamp: time.Now().Unix(),
	}
	c.JSON(statusCode, response)
}

// BadRequest sends 400 Bad Request response
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, ErrCodeValidationFailed, message)
}

// BadRequestWithCode sends 400 Bad Request response with custom error code
func BadRequestWithCode(c *gin.Context, errorCode, message string) {
	Error(c, http.StatusBadRequest, errorCode, message)
}

// Unauthorized sends 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, ErrCodeUnauthorized, message)
}

// Forbidden sends 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, ErrCodeForbidden, message)
}

// NotFound sends 404 Not Found response
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, ErrCodeNotFound, message)
}

// Conflict sends 409 Conflict response
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, ErrCodeConflict, message)
}

// InternalServerError sends 500 Internal Server Error response
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, ErrCodeInternalError, message)
}

// InsufficientBalance sends 400 Insufficient Balance error response
func InsufficientBalance(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, ErrCodeInsufficientBalance, message)
}

// InvalidProduct sends 400 Invalid Product error response
func InvalidProduct(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, ErrCodeInvalidProduct, message)
}

// SupplierError sends 502 Supplier Error response
func SupplierError(c *gin.Context, message string) {
	Error(c, http.StatusBadGateway, ErrCodeSupplierError, message)
}

// TransactionFailed sends 400 Transaction Failed error response
func TransactionFailed(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, ErrCodeTransactionFailed, message)
}

// UserNotFound sends 404 User Not Found error response
func UserNotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, ErrCodeUserNotFound, message)
}

// InvalidCredentials sends 401 Invalid Credentials error response
func InvalidCredentials(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, ErrCodeInvalidCredentials, message)
}

// AccountLocked sends 423 Account Locked error response
func AccountLocked(c *gin.Context, message string) {
	Error(c, http.StatusLocked, ErrCodeAccountLocked, message)
}

// RateLimitExceeded sends 429 Rate Limit Exceeded error response
func RateLimitExceeded(c *gin.Context, message string) {
	Error(c, http.StatusTooManyRequests, ErrCodeRateLimitExceeded, message)
}

// Paginated sends paginated response
func Paginated(c *gin.Context, message string, data interface{}, page, limit, total int) {
	totalPages := (total + limit - 1) / limit
	
	response := PaginatedResponse{
		Code:    http.StatusOK,
		Status:  "success",
		Message: message,
		Data:    data,
		Pagination: PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
		Timestamp: time.Now().Unix(),
	}
	c.JSON(http.StatusOK, response)
}

// ValidationError sends validation error response with field details
func ValidationError(c *gin.Context, details interface{}) {
	ErrorWithDetails(c, http.StatusBadRequest, ErrCodeValidationFailed, "Validation failed", details)
}

// Helper function to get status from code
func GetStatusFromCode(code int) string {
	if code >= 200 && code < 300 {
		return "success"
	}
	return "error"
}

// Helper function to create standard response
func NewResponse(code int, message string, data interface{}) Response {
	return Response{
		Code:      code,
		Status:    GetStatusFromCode(code),
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}

// Helper function to create error response
func NewErrorResponse(code int, errorCode, message string, details interface{}) ErrorResponse {
	return ErrorResponse{
		Code:      code,
		Status:    "error",
		ErrorCode: errorCode,
		Message:   message,
		Details:   details,
		Timestamp: time.Now().Unix(),
	}
}
