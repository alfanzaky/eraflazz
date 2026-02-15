package domain

import (
	"fmt"
	"time"
)

// Inbox represents incoming messages
type Inbox struct {
	ID              string  `json:"id" db:"id"`
	Source          string  `json:"source" db:"source"`
	SenderNumber    string  `json:"sender_number" db:"sender_number"`
	SenderName      *string `json:"sender_name" db:"sender_name"`
	Message         string  `json:"message" db:"message"`
	OriginalMessage *string `json:"original_message" db:"original_message"`

	// Processing information
	UserID        *string    `json:"user_id" db:"user_id"`
	TransactionID *string    `json:"transaction_id" db:"transaction_id"`
	Status        string     `json:"status" db:"status"`
	ProcessedAt   *time.Time `json:"processed_at" db:"processed_at"`

	// Response
	ResponseMessage *string    `json:"response_message" db:"response_message"`
	ResponseSentAt  *time.Time `json:"response_sent_at" db:"response_sent_at"`

	// Metadata
	IPAddress  *string `json:"ip_address" db:"ip_address"`
	DeviceInfo *string `json:"device_info" db:"device_info"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Outbox represents outgoing messages
type Outbox struct {
	ID              string  `json:"id" db:"id"`
	Destination     string  `json:"destination" db:"destination"`
	RecipientNumber string  `json:"recipient_number" db:"recipient_number"`
	RecipientName   *string `json:"recipient_name" db:"recipient_name"`
	Message         string  `json:"message" db:"message"`
	MessageType     string  `json:"message_type" db:"message_type"`

	// Related entities
	UserID        *string `json:"user_id" db:"user_id"`
	TransactionID *string `json:"transaction_id" db:"transaction_id"`

	// Sending status
	Status         string     `json:"status" db:"status"`
	RetryCount     int        `json:"retry_count" db:"retry_count"`
	MaxRetries     int        `json:"max_retries" db:"max_retries"`
	SentAt         *time.Time `json:"sent_at" db:"sent_at"`
	DeliveryReport *string    `json:"delivery_report" db:"delivery_report"`
	ExternalID     *string    `json:"external_id" db:"external_id"`

	// Scheduling
	ScheduledAt time.Time  `json:"scheduled_at" db:"scheduled_at"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`

	// Metadata
	Priority  int     `json:"priority" db:"priority"`
	CreatedBy *string `json:"created_by" db:"created_by"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// InboxRepository defines operations for inbox data access
type InboxRepository interface {
	Create(inbox *Inbox) error
	GetByID(id string) (*Inbox, error)
	Update(inbox *Inbox) error
	GetBySenderNumber(senderNumber string) ([]*Inbox, error)
	GetByStatus(status string) ([]*Inbox, error)
	GetPendingMessages() ([]*Inbox, error)
	GetUnprocessedMessages() ([]*Inbox, error)
	MarkAsProcessed(id string, responseMessage string) error
}

// OutboxRepository defines operations for outbox data access
type OutboxRepository interface {
	Create(outbox *Outbox) error
	GetByID(id string) (*Outbox, error)
	Update(outbox *Outbox) error
	GetByStatus(status string) ([]*Outbox, error)
	GetPendingMessages() ([]*Outbox, error)
	GetScheduledMessages() ([]*Outbox, error)
	GetExpiredMessages() ([]*Outbox, error)
	MarkAsSent(id string, externalID string) error
	MarkAsFailed(id string, deliveryReport string) error
	IncrementRetryCount(id string) error
}

// MessageUsecase defines business logic operations for messages
type MessageUsecase interface {
	ProcessIncomingMessage(source, senderNumber, message string) error
	SendMessage(destination, recipientNumber, message string, messageType string) error
	SendTransactionNotification(userID, transactionID string) error
	SendBalanceNotification(userID string, amount float64, mutationType string) error
	BroadcastMessage(userIDs []string, message string) error
	ProcessPendingOutbox() error
	ProcessPendingInbox() error
	GetMessageHistory(userID string, limit, offset int) ([]*Inbox, []*Outbox, error)
}

// Message validation constants
const (
	// Message sources
	SourceWhatsApp = "WHATSAPP"
	SourceTelegram = "TELEGRAM"
	SourceSMS      = "SMS"
	SourceAPI      = "API"

	// Message statuses
	MessageStatusPending    = "PENDING"
	MessageStatusProcessing = "PROCESSING"
	MessageStatusProcessed  = "PROCESSED"
	MessageStatusFailed     = "FAILED"
	MessageStatusIgnored    = "IGNORED"
	MessageStatusSending    = "SENDING"
	MessageStatusSent       = "SENT"
	MessageStatusCancelled  = "CANCELLED"

	// Message types
	MessageTypeNotification = "NOTIFICATION"
	MessageTypeTransaction  = "TRANSACTION"
	MessageTypeAlert        = "ALERT"
	MessageTypeMarketing    = "MARKETING"

	// Message priorities
	PriorityHigh   = 1
	PriorityNormal = 2
	PriorityLow    = 3
)

// IsValidMessageStatus checks if the message status is valid
func IsValidMessageStatus(status string) bool {
	validStatuses := []string{
		MessageStatusPending, MessageStatusProcessing, MessageStatusProcessed,
		MessageStatusFailed, MessageStatusIgnored, MessageStatusSending,
		MessageStatusSent, MessageStatusCancelled,
	}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// IsValidSource checks if the message source is valid
func IsValidSource(source string) bool {
	validSources := []string{SourceWhatsApp, SourceTelegram, SourceSMS, SourceAPI}
	for _, s := range validSources {
		if s == source {
			return true
		}
	}
	return false
}

// IsValidMessageType checks if the message type is valid
func IsValidMessageType(messageType string) bool {
	validTypes := []string{MessageTypeNotification, MessageTypeTransaction, MessageTypeAlert, MessageTypeMarketing}
	for _, t := range validTypes {
		if t == messageType {
			return true
		}
	}
	return false
}

// IsExpired checks if the outbox message is expired
func (o *Outbox) IsExpired() bool {
	if o.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*o.ExpiresAt)
}

// CanRetry checks if the message can be retried
func (o *Outbox) CanRetry() bool {
	return o.Status == MessageStatusFailed && o.RetryCount < o.MaxRetries && !o.IsExpired()
}

// IsReadyToSend checks if the message is ready to be sent
func (o *Outbox) IsReadyToSend() bool {
	return (o.Status == MessageStatusPending || o.CanRetry()) &&
		time.Now().After(o.ScheduledAt) &&
		!o.IsExpired()
}

// ParseTransactionCommand parses transaction command from message (e.g., "T10.08123456789.1234")
func ParseTransactionCommand(message string) (productCode, destination, pin string, isValid bool) {
	// Simple parsing logic - can be enhanced
	parts := []string{}
	current := ""

	for _, char := range message {
		if char == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	if len(parts) >= 3 {
		productCode = parts[0]
		destination = parts[1]
		pin = parts[2]
		isValid = true
	}

	return
}

// FormatBalance formats balance amount for display
func FormatBalance(amount float64) string {
	// Simple formatting - can be enhanced with proper currency formatting
	return fmt.Sprintf("Rp %.2f", amount)
}

// GenerateTransactionResponse generates response message for transaction
func GenerateTransactionResponse(transaction *Transaction) string {
	switch transaction.Status {
	case StatusSuccess:
		return fmt.Sprintf("Transaksi BERHASIL! %s -> %s. SN: %s",
			transaction.ProductCode, transaction.DestinationNumber,
			*transaction.SerialNumber)
	case StatusFailed:
		return fmt.Sprintf("Transaksi GAGAL! %s -> %s. %s",
			transaction.ProductCode, transaction.DestinationNumber,
			func() string {
				if transaction.SupplierMessage != nil {
					return *transaction.SupplierMessage
				}
				return "Silakan coba beberapa saat lagi."
			}())
	case StatusPending:
		return fmt.Sprintf("Transaksi DIPROSES! %s -> %s. Mohon ditunggu...",
			transaction.ProductCode, transaction.DestinationNumber)
	default:
		return fmt.Sprintf("Status transaksi: %s", transaction.Status)
	}
}
