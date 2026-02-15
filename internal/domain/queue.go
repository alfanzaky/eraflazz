package domain

// QueueRepository defines the contract for background job queues
// that transport transaction IDs to workers for processing.
type QueueRepository interface {
	EnqueueTransaction(transactionID string) error
	DequeueTransaction() (string, error)
	GetQueueLength() (int64, error)
}
