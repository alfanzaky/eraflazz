package domain

import (
	"time"
)

// Product represents a product in the system
type Product struct {
	ID          string  `json:"id" db:"id"`
	Code        string  `json:"code" db:"code"`
	Name        string  `json:"name" db:"name"`
	Description *string `json:"description" db:"description"`

	// Categorization
	Category string `json:"category" db:"category"`
	Provider string `json:"provider" db:"provider"`
	Type     string `json:"type" db:"type"`

	// Pricing
	BasePrice    float64 `json:"base_price" db:"base_price"`
	SellingPrice float64 `json:"selling_price" db:"selling_price"`
	MinPrice     float64 `json:"min_price" db:"min_price"`

	// Specifications
	Nominal        *float64 `json:"nominal" db:"nominal"`
	ValidityPeriod *string  `json:"validity_period" db:"validity_period"`

	// Status and availability
	IsActive         bool `json:"is_active" db:"is_active"`
	IsUnlimitedStock bool `json:"is_unlimited_stock" db:"is_unlimited_stock"`
	StockQuantity    int  `json:"stock_quantity" db:"stock_quantity"`

	// Business rules
	AllowMarkup          bool    `json:"allow_markup" db:"allow_markup"`
	MaxMarkupPercentage  float64 `json:"max_markup_percentage" db:"max_markup_percentage"`
	MinTransactionAmount float64 `json:"min_transaction_amount" db:"min_transaction_amount"`
	MaxTransactionAmount float64 `json:"max_transaction_amount" db:"max_transaction_amount"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ProductMapping represents mapping between product and supplier
type ProductMapping struct {
	ID                  string `json:"id" db:"id"`
	ProductID           string `json:"product_id" db:"product_id"`
	SupplierID          string `json:"supplier_id" db:"supplier_id"`
	SupplierProductCode string `json:"supplier_product_code" db:"supplier_product_code"`

	// Supplier-specific pricing
	SupplierPrice float64 `json:"supplier_price" db:"supplier_price"`
	AdditionalFee float64 `json:"additional_fee" db:"additional_fee"`

	// Priority and availability
	Priority    int    `json:"priority" db:"priority"`
	IsActive    bool   `json:"is_active" db:"is_active"`
	StockStatus string `json:"stock_status" db:"stock_status"`

	// Performance metrics
	SuccessCount   int        `json:"success_count" db:"success_count"`
	FailureCount   int        `json:"failure_count" db:"failure_count"`
	LastSuccessAt  *time.Time `json:"last_success_at" db:"last_success_at"`
	LastFailureAt  *time.Time `json:"last_failure_at" db:"last_failure_at"`
	LastStockCheck *time.Time `json:"last_stock_check" db:"last_stock_check"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ProductRepository defines operations for product data access
type ProductRepository interface {
	Create(product *Product) error
	GetByID(id string) (*Product, error)
	GetByCode(code string) (*Product, error)
	Update(product *Product) error
	Delete(id string) error
	GetByCategory(category string) ([]*Product, error)
	GetByProvider(provider string) ([]*Product, error)
	GetActiveProducts() ([]*Product, error)
	Search(query string) ([]*Product, error)
	List(filter *ProductFilter) ([]*Product, error)
	Count(filter *ProductFilter) (int, error)
	UpdateStatus(id string, isActive bool) error
	UpdateStock(id string, stockQuantity int, isUnlimited bool) error
}

// ProductMappingRepository defines operations for product mapping data access
type ProductMappingRepository interface {
	Create(mapping *ProductMapping) error
	GetByID(id string) (*ProductMapping, error)
	GetByProductAndSupplier(productID, supplierID string) (*ProductMapping, error)
	GetByProductID(productID string) ([]*ProductMapping, error)
	GetActiveMappings(productID string) ([]*ProductMapping, error)
	Update(mapping *ProductMapping) error
	Delete(id string) error
	GetBySupplierID(supplierID string) ([]*ProductMapping, error)
}

// ProductUsecase defines business logic operations for products
type ProductUsecase interface {
	CreateProduct(product *Product) error
	UpdateProduct(id string, updates *Product) error
	ListProducts(filter *ProductFilter) ([]*Product, int, error)
	GetProduct(id string) (*Product, error)
	GetProductByCode(code string) (*Product, error)
	GetProductsByCategory(category string) ([]*Product, error)
	GetActiveProducts() ([]*Product, error)
	SearchProducts(query string) ([]*Product, error)
	ToggleProductStatus(id string, isActive bool) error
	UpdateProductStock(id string, stockQuantity int, isUnlimited bool) error
	GetBestSupplier(productID string) (*ProductMapping, error)
	UpdateProductMapping(mapping *ProductMapping) error
	GetProductMappings(productID string) ([]*ProductMapping, error)
	GetProductMapping(id string) (*ProductMapping, error)
	CreateProductMapping(mapping *ProductMapping) error
	DeleteProductMapping(id string) error
}

// ProductFilter represents filter criteria for listing products
type ProductFilter struct {
	Category *string
	Provider *string
	Query    *string
	IsActive *bool
	Page     int
	PageSize int
}

// Product validation constants
const (
	CategoryPulsa   = "PULSA"
	CategoryData    = "DATA"
	CategoryPLN     = "PLN"
	CategoryPDAM    = "PDAM"
	CategoryBPJS    = "BPJS"
	CategoryGame    = "GAME"
	CategoryVoucher = "VOUCHER"

	TypePrepaid  = "PREPAID"
	TypePostpaid = "POSTPAID"
	TypeVoucher  = "VOUCHER"

	StockStatusAvailable  = "AVAILABLE"
	StockStatusOutOfStock = "OUT_OF_STOCK"
	StockStatusUnknown    = "UNKNOWN"
)

// IsValidCategory checks if the category is valid
func IsValidCategory(category string) bool {
	validCategories := []string{
		CategoryPulsa, CategoryData, CategoryPLN, CategoryPDAM,
		CategoryBPJS, CategoryGame, CategoryVoucher,
	}
	for _, c := range validCategories {
		if c == category {
			return true
		}
	}
	return false
}

// IsValidType checks if the product type is valid
func IsValidType(productType string) bool {
	validTypes := []string{TypePrepaid, TypePostpaid, TypeVoucher}
	for _, t := range validTypes {
		if t == productType {
			return true
		}
	}
	return false
}

// GetSuccessRate calculates success rate percentage
func (pm *ProductMapping) GetSuccessRate() float64 {
	total := pm.SuccessCount + pm.FailureCount
	if total == 0 {
		return 100.0 // Default to 100% if no transactions
	}
	return float64(pm.SuccessCount) / float64(total) * 100
}

// GetEffectivePrice calculates the total price including additional fees
func (pm *ProductMapping) GetEffectivePrice() float64 {
	return pm.SupplierPrice + pm.AdditionalFee
}

// IsAvailable checks if the product mapping is available for use
func (pm *ProductMapping) IsAvailable() bool {
	return pm.IsActive && pm.StockStatus == StockStatusAvailable
}
