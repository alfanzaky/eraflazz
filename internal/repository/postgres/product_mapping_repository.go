package postgres

import (
    "fmt"

    "github.com/jmoiron/sqlx"

    "github.com/alfanzaky/eraflazz/internal/domain"
    "github.com/alfanzaky/eraflazz/pkg/logger"
)

type productMappingRepository struct {
    db *sqlx.DB
}

// NewProductMappingRepository creates a new repository instance
func NewProductMappingRepository(db *sqlx.DB) domain.ProductMappingRepository {
    return &productMappingRepository{db: db}
}

func (r *productMappingRepository) Create(mapping *domain.ProductMapping) error {
    query := `
        INSERT INTO product_mappings (
            id, product_id, supplier_id, supplier_product_code,
            supplier_price, additional_fee, priority, is_active,
            stock_status, success_count, failure_count,
            last_success_at, last_failure_at, last_stock_check,
            created_at, updated_at
        ) VALUES (
            :id, :product_id, :supplier_id, :supplier_product_code,
            :supplier_price, :additional_fee, :priority, :is_active,
            :stock_status, :success_count, :failure_count,
            :last_success_at, :last_failure_at, :last_stock_check,
            NOW(), NOW()
        )`

    _, err := r.db.NamedExec(query, mapping)
    if err != nil {
        logger.Error("Failed to create product mapping", logger.ErrorField(err))
        return fmt.Errorf("failed to create product mapping: %w", err)
    }
    return nil
}

func (r *productMappingRepository) GetByID(id string) (*domain.ProductMapping, error) {
    query := `SELECT * FROM product_mappings WHERE id = $1`
    var mapping domain.ProductMapping
    if err := r.db.Get(&mapping, query, id); err != nil {
        return nil, fmt.Errorf("failed to get product mapping: %w", err)
    }
    return &mapping, nil
}

func (r *productMappingRepository) GetByProductAndSupplier(productID, supplierID string) (*domain.ProductMapping, error) {
    query := `SELECT * FROM product_mappings WHERE product_id = $1 AND supplier_id = $2`
    var mapping domain.ProductMapping
    if err := r.db.Get(&mapping, query, productID, supplierID); err != nil {
        return nil, fmt.Errorf("failed to get product mapping: %w", err)
    }
    return &mapping, nil
}

func (r *productMappingRepository) GetByProductID(productID string) ([]*domain.ProductMapping, error) {
    query := `SELECT * FROM product_mappings WHERE product_id = $1`
    var mappings []*domain.ProductMapping
    if err := r.db.Select(&mappings, query, productID); err != nil {
        return nil, fmt.Errorf("failed to get product mappings by product: %w", err)
    }
    return mappings, nil
}

func (r *productMappingRepository) GetActiveMappings(productID string) ([]*domain.ProductMapping, error) {
    query := `
        SELECT * FROM product_mappings 
        WHERE product_id = $1 AND is_active = TRUE
        ORDER BY priority ASC, supplier_price ASC`
    var mappings []*domain.ProductMapping
    if err := r.db.Select(&mappings, query, productID); err != nil {
        return nil, fmt.Errorf("failed to get active product mappings: %w", err)
    }
    return mappings, nil
}

func (r *productMappingRepository) Update(mapping *domain.ProductMapping) error {
    query := `
        UPDATE product_mappings SET
            supplier_product_code = :supplier_product_code,
            supplier_price = :supplier_price,
            additional_fee = :additional_fee,
            priority = :priority,
            is_active = :is_active,
            stock_status = :stock_status,
            success_count = :success_count,
            failure_count = :failure_count,
            last_success_at = :last_success_at,
            last_failure_at = :last_failure_at,
            last_stock_check = :last_stock_check,
            updated_at = NOW()
        WHERE id = :id`

    _, err := r.db.NamedExec(query, mapping)
    if err != nil {
        logger.Error("Failed to update product mapping", logger.ErrorField(err))
        return fmt.Errorf("failed to update product mapping: %w", err)
    }
    return nil
}

func (r *productMappingRepository) Delete(id string) error {
    query := `DELETE FROM product_mappings WHERE id = $1`
    if _, err := r.db.Exec(query, id); err != nil {
        logger.Error("Failed to delete product mapping", logger.ErrorField(err))
        return fmt.Errorf("failed to delete product mapping: %w", err)
    }
    return nil
}

func (r *productMappingRepository) GetBySupplierID(supplierID string) ([]*domain.ProductMapping, error) {
    query := `SELECT * FROM product_mappings WHERE supplier_id = $1`
    var mappings []*domain.ProductMapping
    if err := r.db.Select(&mappings, query, supplierID); err != nil {
        return nil, fmt.Errorf("failed to get product mappings by supplier: %w", err)
    }
    return mappings, nil
}
