package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alfanzaky/eraflazz/internal/domain"
	"github.com/alfanzaky/eraflazz/pkg/logger"
	"github.com/go-redis/redis/v8"
)

type cacheRepository struct {
	client *redis.Client
}

var _ domain.QueueRepository = (*cacheRepository)(nil)

// NewCacheRepository creates a new Redis cache repository
func NewCacheRepository(client *redis.Client) *cacheRepository {
	return &cacheRepository{client: client}
}

// Cache keys
const (
	UserKeyPrefix        = "user:"
	ProductKeyPrefix     = "product:"
	SupplierKeyPrefix    = "supplier:"
	TransactionKeyPrefix = "trx:"
	BalanceKeyPrefix     = "balance:"
	ProductMappingPrefix = "mapping:"

	// TTL durations
	UserCacheTTL        = 30 * time.Minute
	ProductCacheTTL     = 60 * time.Minute
	SupplierCacheTTL    = 15 * time.Minute
	TransactionCacheTTL = 5 * time.Minute
	BalanceCacheTTL     = 1 * time.Minute
	ProductMappingTTL   = 30 * time.Minute
)

// User caching
func (r *cacheRepository) CacheUser(user *domain.User) error {
	key := UserKeyPrefix + user.ID

	data, err := json.Marshal(user)
	if err != nil {
		logger.Error("Failed to marshal user for cache",
			logger.String("user_id", user.ID),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	err = r.client.Set(context.Background(), key, data, UserCacheTTL).Err()
	if err != nil {
		logger.Error("Failed to cache user",
			logger.String("user_id", user.ID),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to cache user: %w", err)
	}

	logger.Debug("User cached successfully",
		logger.String("user_id", user.ID),
	)

	return nil
}

func (r *cacheRepository) GetUser(userID string) (*domain.User, error) {
	key := UserKeyPrefix + userID

	data, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		logger.Error("Failed to get user from cache",
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get user from cache: %w", err)
	}

	var user domain.User
	err = json.Unmarshal([]byte(data), &user)
	if err != nil {
		logger.Error("Failed to unmarshal user from cache",
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	logger.Debug("User retrieved from cache",
		logger.String("user_id", userID),
	)

	return &user, nil
}

func (r *cacheRepository) InvalidateUser(userID string) error {
	key := UserKeyPrefix + userID

	err := r.client.Del(context.Background(), key).Err()
	if err != nil {
		logger.Error("Failed to invalidate user cache",
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to invalidate user cache: %w", err)
	}

	logger.Debug("User cache invalidated",
		logger.String("user_id", userID),
	)

	return nil
}

// Product caching
func (r *cacheRepository) CacheProduct(product *domain.Product) error {
	key := ProductKeyPrefix + product.ID

	data, err := json.Marshal(product)
	if err != nil {
		logger.Error("Failed to marshal product for cache",
			logger.String("product_id", product.ID),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to marshal product: %w", err)
	}

	err = r.client.Set(context.Background(), key, data, ProductCacheTTL).Err()
	if err != nil {
		logger.Error("Failed to cache product",
			logger.String("product_id", product.ID),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to cache product: %w", err)
	}

	return nil
}

func (r *cacheRepository) GetProduct(productID string) (*domain.Product, error) {
	key := ProductKeyPrefix + productID

	data, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		logger.Error("Failed to get product from cache",
			logger.String("product_id", productID),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get product from cache: %w", err)
	}

	var product domain.Product
	err = json.Unmarshal([]byte(data), &product)
	if err != nil {
		logger.Error("Failed to unmarshal product from cache",
			logger.String("product_id", productID),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to unmarshal product: %w", err)
	}

	logger.Debug("Product retrieved from cache",
		logger.String("product_id", productID),
	)

	return &product, nil
}

func (r *cacheRepository) CacheProductByCode(code string, product *domain.Product) error {
	key := ProductKeyPrefix + "code:" + code

	data, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to marshal product: %w", err)
	}

	err = r.client.Set(context.Background(), key, data, ProductCacheTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to cache product by code: %w", err)
	}

	return nil
}

func (r *cacheRepository) GetProductByCode(code string) (*domain.Product, error) {
	key := ProductKeyPrefix + "code:" + code

	data, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get product from cache: %w", err)
	}

	var product domain.Product
	err = json.Unmarshal([]byte(data), &product)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal product: %w", err)
	}

	return &product, nil
}

// Supplier caching
func (r *cacheRepository) CacheSupplier(supplier *domain.Supplier) error {
	key := SupplierKeyPrefix + supplier.ID

	data, err := json.Marshal(supplier)
	if err != nil {
		return fmt.Errorf("failed to marshal supplier: %w", err)
	}

	err = r.client.Set(context.Background(), key, data, SupplierCacheTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to cache supplier: %w", err)
	}

	return nil
}

func (r *cacheRepository) GetSupplier(supplierID string) (*domain.Supplier, error) {
	key := SupplierKeyPrefix + supplierID

	data, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get supplier from cache: %w", err)
	}

	var supplier domain.Supplier
	err = json.Unmarshal([]byte(data), &supplier)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal supplier: %w", err)
	}

	return &supplier, nil
}

func (r *cacheRepository) CacheActiveSuppliers(suppliers []*domain.Supplier) error {
	key := SupplierKeyPrefix + "active"

	data, err := json.Marshal(suppliers)
	if err != nil {
		return fmt.Errorf("failed to marshal suppliers: %w", err)
	}

	err = r.client.Set(context.Background(), key, data, SupplierCacheTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to cache active suppliers: %w", err)
	}

	return nil
}

func (r *cacheRepository) GetActiveSuppliers() ([]*domain.Supplier, error) {
	key := SupplierKeyPrefix + "active"

	data, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get active suppliers from cache: %w", err)
	}

	var suppliers []*domain.Supplier
	err = json.Unmarshal([]byte(data), &suppliers)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal suppliers: %w", err)
	}

	return suppliers, nil
}

// Balance caching
func (r *cacheRepository) CacheUserBalance(userID string, balance float64) error {
	key := BalanceKeyPrefix + userID

	err := r.client.Set(context.Background(), key, balance, BalanceCacheTTL).Err()
	if err != nil {
		logger.Error("Failed to cache user balance",
			logger.String("user_id", userID),
			logger.Float64("balance", balance),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to cache user balance: %w", err)
	}

	return nil
}

func (r *cacheRepository) GetUserBalance(userID string) (float64, error) {
	key := BalanceKeyPrefix + userID

	balanceStr, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // Cache miss
		}
		logger.Error("Failed to get user balance from cache",
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)
		return 0, fmt.Errorf("failed to get user balance from cache: %w", err)
	}

	var balance float64
	_, err = fmt.Sscanf(balanceStr, "%f", &balance)
	if err != nil {
		logger.Error("Failed to parse balance from cache",
			logger.String("user_id", userID),
			logger.String("balance_str", balanceStr),
			logger.ErrorField(err),
		)
		return 0, fmt.Errorf("failed to parse balance: %w", err)
	}

	return balance, nil
}

func (r *cacheRepository) InvalidateUserBalance(userID string) error {
	key := BalanceKeyPrefix + userID

	err := r.client.Del(context.Background(), key).Err()
	if err != nil {
		return fmt.Errorf("failed to invalidate user balance cache: %w", err)
	}

	return nil
}

// Product mapping caching
func (r *cacheRepository) CacheProductMappings(productID string, mappings []*domain.ProductMapping) error {
	key := ProductMappingPrefix + productID

	data, err := json.Marshal(mappings)
	if err != nil {
		return fmt.Errorf("failed to marshal product mappings: %w", err)
	}

	err = r.client.Set(context.Background(), key, data, ProductMappingTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to cache product mappings: %w", err)
	}

	return nil
}

func (r *cacheRepository) GetProductMappings(productID string) ([]*domain.ProductMapping, error) {
	key := ProductMappingPrefix + productID

	data, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get product mappings from cache: %w", err)
	}

	var mappings []*domain.ProductMapping
	err = json.Unmarshal([]byte(data), &mappings)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal product mappings: %w", err)
	}

	return mappings, nil
}

// Transaction queue operations
func (r *cacheRepository) EnqueueTransaction(transactionID string) error {
	queueKey := "transaction_queue"

	err := r.client.LPush(context.Background(), queueKey, transactionID).Err()
	if err != nil {
		logger.Error("Failed to enqueue transaction",
			logger.String("transaction_id", transactionID),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to enqueue transaction: %w", err)
	}

	logger.Debug("Transaction enqueued",
		logger.String("transaction_id", transactionID),
	)

	return nil
}

func (r *cacheRepository) DequeueTransaction() (string, error) {
	queueKey := "transaction_queue"

	result, err := r.client.BRPop(context.Background(), 5*time.Second, queueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // No items in queue
		}
		logger.Error("Failed to dequeue transaction", logger.ErrorField(err))
		return "", fmt.Errorf("failed to dequeue transaction: %w", err)
	}

	if len(result) < 2 {
		return "", fmt.Errorf("unexpected queue result format")
	}

	transactionID := result[1]
	logger.Debug("Transaction dequeued",
		logger.String("transaction_id", transactionID),
	)

	return transactionID, nil
}

func (r *cacheRepository) GetQueueLength() (int64, error) {
	queueKey := "transaction_queue"

	length, err := r.client.LLen(context.Background(), queueKey).Result()
	if err != nil {
		logger.Error("Failed to get queue length", logger.ErrorField(err))
		return 0, fmt.Errorf("failed to get queue length: %w", err)
	}

	return length, nil
}

// Health check
func (r *cacheRepository) Ping() error {
	return r.client.Ping(context.Background()).Err()
}

// Clear cache (for testing)
func (r *cacheRepository) ClearAll() error {
	return r.client.FlushDB(context.Background()).Err()
}
