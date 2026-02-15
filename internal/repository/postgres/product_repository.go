package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/alfanzaky/eraflazz/internal/domain"
	"github.com/alfanzaky/eraflazz/pkg/logger"
	"github.com/jmoiron/sqlx"
)

type productRepository struct {
	db *sqlx.DB
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *sqlx.DB) domain.ProductRepository {
	return &productRepository{db: db}
}

// Create creates a new product
func (r *productRepository) Create(product *domain.Product) error {
	query := `
		INSERT INTO products (id, code, name, description, category, provider, type,
			base_price, selling_price, min_price, nominal, validity_period,
			is_active, is_unlimited_stock, stock_quantity, allow_markup,
			max_markup_percentage, min_transaction_amount, max_transaction_amount)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`

	_, err := r.db.Exec(query,
		product.ID, product.Code, product.Name, product.Description,
		product.Category, product.Provider, product.Type, product.BasePrice,
		product.SellingPrice, product.MinPrice, product.Nominal, product.ValidityPeriod,
		product.IsActive, product.IsUnlimitedStock, product.StockQuantity,
		product.AllowMarkup, product.MaxMarkupPercentage, product.MinTransactionAmount,
		product.MaxTransactionAmount,
	)

	if err != nil {
		logger.Error("Failed to create product",
			logger.String("code", product.Code),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to create product: %w", err)
	}

	logger.Info("Product created successfully",
		logger.String("product_id", product.ID),
		logger.String("code", product.Code),
	)

	return nil
}

// GetByID retrieves a product by ID
func (r *productRepository) GetByID(id string) (*domain.Product, error) {
	query := `
		SELECT id, code, name, description, category, provider, type,
			base_price, selling_price, min_price, nominal, validity_period,
			is_active, is_unlimited_stock, stock_quantity, allow_markup,
			max_markup_percentage, min_transaction_amount, max_transaction_amount,
			created_at, updated_at
		FROM products WHERE id = $1
	`

	var product domain.Product
	err := r.db.Get(&product, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found")
		}
		logger.Error("Failed to get product by ID",
			logger.String("product_id", id),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

// GetByCode retrieves a product by code
func (r *productRepository) GetByCode(code string) (*domain.Product, error) {
	query := `
		SELECT id, code, name, description, category, provider, type,
			base_price, selling_price, min_price, nominal, validity_period,
			is_active, is_unlimited_stock, stock_quantity, allow_markup,
			max_markup_percentage, min_transaction_amount, max_transaction_amount,
			created_at, updated_at
		FROM products WHERE code = $1
	`

	var product domain.Product
	err := r.db.Get(&product, query, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found")
		}
		logger.Error("Failed to get product by code",
			logger.String("code", code),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

// Update updates a product
func (r *productRepository) Update(product *domain.Product) error {
	query := `
		UPDATE products SET 
			code = $2, name = $3, description = $4, category = $5, provider = $6, type = $7,
			base_price = $8, selling_price = $9, min_price = $10, nominal = $11, validity_period = $12,
			is_active = $13, is_unlimited_stock = $14, stock_quantity = $15, allow_markup = $16,
			max_markup_percentage = $17, min_transaction_amount = $18, max_transaction_amount = $19
		WHERE id = $1
	`

	result, err := r.db.Exec(query,
		product.ID, product.Code, product.Name, product.Description,
		product.Category, product.Provider, product.Type, product.BasePrice,
		product.SellingPrice, product.MinPrice, product.Nominal, product.ValidityPeriod,
		product.IsActive, product.IsUnlimitedStock, product.StockQuantity,
		product.AllowMarkup, product.MaxMarkupPercentage, product.MinTransactionAmount,
		product.MaxTransactionAmount,
	)

	if err != nil {
		logger.Error("Failed to update product",
			logger.String("product_id", product.ID),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to update product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	logger.Info("Product updated successfully",
		logger.String("product_id", product.ID),
		logger.String("code", product.Code),
	)

	return nil
}

// Delete deletes a product
func (r *productRepository) Delete(id string) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		logger.Error("Failed to delete product",
			logger.String("product_id", id),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	logger.Info("Product deleted successfully",
		logger.String("product_id", id),
	)

	return nil
}

// GetByCategory retrieves products by category
func (r *productRepository) GetByCategory(category string) ([]*domain.Product, error) {
	query := `
		SELECT id, code, name, description, category, provider, type,
			base_price, selling_price, min_price, nominal, validity_period,
			is_active, is_unlimited_stock, stock_quantity, allow_markup,
			max_markup_percentage, min_transaction_amount, max_transaction_amount,
			created_at, updated_at
		FROM products WHERE category = $1 ORDER BY code ASC
	`

	var products []*domain.Product
	err := r.db.Select(&products, query, category)
	if err != nil {
		logger.Error("Failed to get products by category",
			logger.String("category", category),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get products by category: %w", err)
	}

	return products, nil
}

// Count returns total products for a given filter
func (r *productRepository) Count(filter *domain.ProductFilter) (int, error) {
	query := `SELECT COUNT(*) FROM products WHERE 1=1`
	var args []interface{}
	var conditions []string

	if filter != nil {
		if filter.Category != nil {
			conditions = append(conditions, fmt.Sprintf("category = $%d", len(args)+1))
			args = append(args, *filter.Category)
		}
		if filter.Provider != nil {
			conditions = append(conditions, fmt.Sprintf("provider = $%d", len(args)+1))
			args = append(args, *filter.Provider)
		}
		if filter.IsActive != nil {
			conditions = append(conditions, fmt.Sprintf("is_active = $%d", len(args)+1))
			args = append(args, *filter.IsActive)
		}
		if filter.Query != nil && strings.TrimSpace(*filter.Query) != "" {
			conditions = append(conditions, fmt.Sprintf("(code ILIKE $%d OR name ILIKE $%d)", len(args)+1, len(args)+1))
			args = append(args, "%"+strings.TrimSpace(*filter.Query)+"%")
		}
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	var total int
	if err := r.db.Get(&total, query, args...); err != nil {
		return 0, fmt.Errorf("failed to count products: %w", err)
	}

	return total, nil
}

// GetByProvider retrieves products by provider
func (r *productRepository) GetByProvider(provider string) ([]*domain.Product, error) {
	query := `
		SELECT id, code, name, description, category, provider, type,
			base_price, selling_price, min_price, nominal, validity_period,
			is_active, is_unlimited_stock, stock_quantity, allow_markup,
			max_markup_percentage, min_transaction_amount, max_transaction_amount,
			created_at, updated_at
		FROM products WHERE provider = $1 ORDER BY code ASC
	`

	var products []*domain.Product
	err := r.db.Select(&products, query, provider)
	if err != nil {
		logger.Error("Failed to get products by provider",
			logger.String("provider", provider),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get products by provider: %w", err)
	}

	return products, nil
}

// GetActiveProducts retrieves all active products
func (r *productRepository) GetActiveProducts() ([]*domain.Product, error) {
	query := `
		SELECT id, code, name, description, category, provider, type,
			base_price, selling_price, min_price, nominal, validity_period,
			is_active, is_unlimited_stock, stock_quantity, allow_markup,
			max_markup_percentage, min_transaction_amount, max_transaction_amount,
			created_at, updated_at
		FROM products WHERE is_active = true ORDER BY category, code ASC
	`

	var products []*domain.Product
	err := r.db.Select(&products, query)
	if err != nil {
		logger.Error("Failed to get active products", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get active products: %w", err)
	}

	return products, nil
}

// Search searches products by name or code
func (r *productRepository) Search(query string) ([]*domain.Product, error) {
	searchQuery := `%` + query + `%`
	sql := `
		SELECT id, code, name, description, category, provider, type,
			base_price, selling_price, min_price, nominal, validity_period,
			is_active, is_unlimited_stock, stock_quantity, allow_markup,
			max_markup_percentage, min_transaction_amount, max_transaction_amount,
			created_at, updated_at
		FROM products 
		WHERE (code ILIKE $1 OR name ILIKE $1) AND is_active = true
		ORDER BY code ASC
		LIMIT 50
	`

	var products []*domain.Product
	err := r.db.Select(&products, sql, searchQuery)
	if err != nil {
		logger.Error("Failed to search products",
			logger.String("query", query),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	return products, nil
}

// GetProductsByType retrieves products by type
func (r *productRepository) GetProductsByType(productType string) ([]*domain.Product, error) {
	query := `
		SELECT id, code, name, description, category, provider, type,
			base_price, selling_price, min_price, nominal, validity_period,
			is_active, is_unlimited_stock, stock_quantity, allow_markup,
			max_markup_percentage, min_transaction_amount, max_transaction_amount,
			created_at, updated_at
		FROM products WHERE type = $1 AND is_active = true ORDER BY code ASC
	`

	var products []*domain.Product
	err := r.db.Select(&products, query, productType)
	if err != nil {
		logger.Error("Failed to get products by type",
			logger.String("type", productType),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get products by type: %w", err)
	}

	return products, nil
}

// List returns products using flexible filters
func (r *productRepository) List(filter *domain.ProductFilter) ([]*domain.Product, error) {
	baseQuery := `
		SELECT id, code, name, description, category, provider, type,
			base_price, selling_price, min_price, nominal, validity_period,
			is_active, is_unlimited_stock, stock_quantity, allow_markup,
			max_markup_percentage, min_transaction_amount, max_transaction_amount,
			created_at, updated_at
		FROM products
		WHERE 1=1`

	var args []interface{}
	var conditions []string

	if filter != nil {
		if filter.Category != nil {
			conditions = append(conditions, fmt.Sprintf("category = $%d", len(args)+1))
			args = append(args, *filter.Category)
		}
		if filter.Provider != nil {
			conditions = append(conditions, fmt.Sprintf("provider = $%d", len(args)+1))
			args = append(args, *filter.Provider)
		}
		if filter.IsActive != nil {
			conditions = append(conditions, fmt.Sprintf("is_active = $%d", len(args)+1))
			args = append(args, *filter.IsActive)
		}
		if filter.Query != nil && strings.TrimSpace(*filter.Query) != "" {
			conditions = append(conditions, fmt.Sprintf("(code ILIKE $%d OR name ILIKE $%d)", len(args)+1, len(args)+1))
			args = append(args, "%"+strings.TrimSpace(*filter.Query)+"%")
		}
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	baseQuery += " ORDER BY category, code ASC"

	limit := 50
	offset := 0
	if filter != nil {
		if filter.PageSize > 0 {
			limit = filter.PageSize
		}
		page := filter.Page
		if page <= 0 {
			page = 1
		}
		offset = (page - 1) * limit
	}

	baseQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	var products []*domain.Product
	if err := r.db.Select(&products, baseQuery, args...); err != nil {
		logger.Error("Failed to list products", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	return products, nil
}

// UpdateStatus updates product active status
func (r *productRepository) UpdateStatus(id string, isActive bool) error {
	query := `UPDATE products SET is_active = $2, updated_at = NOW() WHERE id = $1`
	result, err := r.db.Exec(query, id, isActive)
	if err != nil {
		logger.Error("Failed to update product status",
			logger.String("product_id", id),
			logger.Bool("is_active", isActive),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to update product status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// UpdateStock updates product stock quantity and unlimited flag
func (r *productRepository) UpdateStock(id string, quantity int, isUnlimited bool) error {
	query := `UPDATE products SET stock_quantity = $2, is_unlimited_stock = $3, updated_at = NOW() WHERE id = $1`

	result, err := r.db.Exec(query, id, quantity, isUnlimited)
	if err != nil {
		logger.Error("Failed to update stock",
			logger.String("product_id", id),
			logger.Int("quantity", quantity),
			logger.Bool("is_unlimited", isUnlimited),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to update stock: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}
