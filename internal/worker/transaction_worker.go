package worker

import (
    "context"
    "time"

    "github.com/alfanzaky/eraflazz/internal/domain"
    "github.com/alfanzaky/eraflazz/pkg/logger"
)

// TransactionWorker continuously consumes transaction IDs from QueueRepository
// and delegates processing to TransactionUsecase. Callers should manage lifecycle
// by controlling the provided context (cancel on shutdown).
type TransactionWorker struct {
    queueRepo domain.QueueRepository
    trxUC     domain.TransactionUsecase
    interval  time.Duration
}

// TransactionWorkerConfig defines runtime options for the worker.
type TransactionWorkerConfig struct {
    PollingInterval time.Duration
}

// NewTransactionWorker builds a new transaction worker instance.
func NewTransactionWorker(queueRepo domain.QueueRepository, trxUC domain.TransactionUsecase, cfg TransactionWorkerConfig) *TransactionWorker {
    interval := cfg.PollingInterval
    if interval <= 0 {
        interval = 500 * time.Millisecond
    }

    return &TransactionWorker{
        queueRepo: queueRepo,
        trxUC:     trxUC,
        interval:  interval,
    }
}

// Start launches the worker loop. It blocks until context cancellation.
func (w *TransactionWorker) Start(ctx context.Context) {
    logger.Info("Transaction worker started")
    ticker := time.NewTicker(w.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            logger.Info("Transaction worker stopping", logger.ErrorField(ctx.Err()))
            return
        case <-ticker.C:
            w.processNext(ctx)
        }
    }
}

func (w *TransactionWorker) processNext(ctx context.Context) {
    if w.queueRepo == nil || w.trxUC == nil {
        logger.Warn("Transaction worker missing dependencies")
        return
    }

    trxID, err := w.queueRepo.DequeueTransaction()
    if err != nil {
        logger.Error("Failed to dequeue transaction", logger.ErrorField(err))
        return
    }

    if trxID == "" {
        // No items available
        return
    }

    start := time.Now()
    err = w.trxUC.ProcessTransaction(trxID)
    duration := time.Since(start)

    if err != nil {
        logger.Error("Failed to process queued transaction",
            logger.String("trx_id", trxID),
            logger.Duration("duration", duration),
            logger.ErrorField(err),
        )
        return
    }

    logger.Info("Queued transaction processed",
        logger.String("trx_id", trxID),
        logger.Duration("duration", duration),
    )
}
