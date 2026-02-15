package factory

import (
	"fmt"
	"strings"
	"sync"

	"github.com/alfanzaky/eraflazz/internal/domain"
)

// supplierAdapterFactory is a thread-safe registry for supplier adapters
// ensuring each supplier code resolves to a concrete adapter implementation.
type supplierAdapterFactory struct {
	mu       sync.RWMutex
	adapters map[string]domain.SupplierAdapter
}

// NewSupplierAdapterFactory creates a new supplier adapter registry instance.
func NewSupplierAdapterFactory() domain.SupplierAdapterFactory {
	return &supplierAdapterFactory{
		adapters: make(map[string]domain.SupplierAdapter),
	}
}

// RegisterAdapter registers an adapter under the given supplier code.
func (f *supplierAdapterFactory) RegisterAdapter(code string, adapter domain.SupplierAdapter) {
	if adapter == nil {
		return
	}

	normalized := strings.ToUpper(strings.TrimSpace(code))
	if normalized == "" {
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	f.adapters[normalized] = adapter
}

// GetAdapter returns the adapter implementation for a supplier code.
func (f *supplierAdapterFactory) GetAdapter(code string) (domain.SupplierAdapter, error) {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	if normalized == "" {
		return nil, fmt.Errorf("supplier code is required")
	}

	f.mu.RLock()
	adapter, ok := f.adapters[normalized]
	f.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("supplier adapter for %s not found", normalized)
	}

	return adapter, nil
}
