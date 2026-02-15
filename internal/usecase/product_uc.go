package usecase

import (
	"fmt"
	"strings"
	"time"

	"github.com/alfanzaky/eraflazz/internal/domain"
	"github.com/alfanzaky/eraflazz/pkg/logger"
	"github.com/alfanzaky/eraflazz/pkg/utils"
)

type productUsecase struct {
	productRepo        domain.ProductRepository
	productMappingRepo domain.ProductMappingRepository
	supplierRepo       domain.SupplierRepository
	smartRoutingUC     *smartRoutingUsecase
}

func NewProductUsecase(
	productRepo domain.ProductRepository,
	productMappingRepo domain.ProductMappingRepository,
	supplierRepo domain.SupplierRepository,
	smartRoutingUC *smartRoutingUsecase,
) domain.ProductUsecase {
	return &productUsecase{
		productRepo:        productRepo,
		productMappingRepo: productMappingRepo,
		supplierRepo:       supplierRepo,
		smartRoutingUC:     smartRoutingUC,
	}
}

func (uc *productUsecase) CreateProduct(product *domain.Product) error {
	if product == nil {
		return fmt.Errorf("product payload is required")
	}

	if strings.TrimSpace(product.Code) == "" || strings.TrimSpace(product.Name) == "" {
		return fmt.Errorf("product code and name are required")
	}

	if !domain.IsValidCategory(product.Category) {
		return fmt.Errorf("invalid product category")
	}

	if !domain.IsValidType(product.Type) {
		return fmt.Errorf("invalid product type")
	}

	product.ID = utils.GenerateUUID()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	return uc.productRepo.Create(product)
}

func (uc *productUsecase) UpdateProduct(id string, updates *domain.Product) error {
	if updates == nil {
		return fmt.Errorf("update payload is required")
	}

	product, err := uc.productRepo.GetByID(id)
	if err != nil {
		return err
	}

	if updates.Name != "" {
		product.Name = updates.Name
	}
	if updates.Description != nil {
		product.Description = updates.Description
	}
	if updates.Category != "" {
		if !domain.IsValidCategory(updates.Category) {
			return fmt.Errorf("invalid product category")
		}
		product.Category = updates.Category
	}
	if updates.Provider != "" {
		product.Provider = updates.Provider
	}
	if updates.Type != "" {
		if !domain.IsValidType(updates.Type) {
			return fmt.Errorf("invalid product type")
		}
		product.Type = updates.Type
	}
	if updates.BasePrice > 0 {
		product.BasePrice = updates.BasePrice
	}
	if updates.SellingPrice > 0 {
		product.SellingPrice = updates.SellingPrice
	}
	if updates.MinPrice > 0 {
		product.MinPrice = updates.MinPrice
	}
	if updates.Nominal != nil {
		product.Nominal = updates.Nominal
	}
	if updates.ValidityPeriod != nil {
		product.ValidityPeriod = updates.ValidityPeriod
	}
	if updates.AllowMarkup {
		product.AllowMarkup = updates.AllowMarkup
	}
	if updates.MaxMarkupPercentage > 0 {
		product.MaxMarkupPercentage = updates.MaxMarkupPercentage
	}
	if updates.MinTransactionAmount > 0 {
		product.MinTransactionAmount = updates.MinTransactionAmount
	}
	if updates.MaxTransactionAmount > 0 {
		product.MaxTransactionAmount = updates.MaxTransactionAmount
	}

	product.UpdatedAt = time.Now()
	return uc.productRepo.Update(product)
}

func (uc *productUsecase) ListProducts(filter *domain.ProductFilter) ([]*domain.Product, int, error) {
	if filter == nil {
		filter = &domain.ProductFilter{}
	}
	products, err := uc.productRepo.List(filter)
	if err != nil {
		return nil, 0, err
	}
	total, err := uc.productRepo.Count(filter)
	if err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (uc *productUsecase) GetProduct(id string) (*domain.Product, error) {
	return uc.productRepo.GetByID(id)
}

func (uc *productUsecase) GetProductByCode(code string) (*domain.Product, error) {
	return uc.productRepo.GetByCode(code)
}

func (uc *productUsecase) GetProductsByCategory(category string) ([]*domain.Product, error) {
	return uc.productRepo.GetByCategory(category)
}

func (uc *productUsecase) GetActiveProducts() ([]*domain.Product, error) {
	return uc.productRepo.GetActiveProducts()
}

func (uc *productUsecase) SearchProducts(query string) ([]*domain.Product, error) {
	return uc.productRepo.Search(query)
}

func (uc *productUsecase) ToggleProductStatus(id string, isActive bool) error {
	return uc.productRepo.UpdateStatus(id, isActive)
}

func (uc *productUsecase) UpdateProductStock(id string, stockQuantity int, isUnlimited bool) error {
	if stockQuantity < 0 {
		return fmt.Errorf("stock quantity cannot be negative")
	}
	return uc.productRepo.UpdateStock(id, stockQuantity, isUnlimited)
}

func (uc *productUsecase) GetBestSupplier(productID string) (*domain.ProductMapping, error) {
	mappings, err := uc.productMappingRepo.GetActiveMappings(productID)
	if err != nil {
		return nil, err
	}
	if len(mappings) == 0 {
		return nil, fmt.Errorf("no active mappings for product")
	}
	return mappings[0], nil
}

func (uc *productUsecase) UpdateProductMapping(mapping *domain.ProductMapping) error {
	if mapping == nil || mapping.ID == "" {
		return fmt.Errorf("mapping payload invalid")
	}
	mapping.UpdatedAt = time.Now()
	if err := uc.productMappingRepo.Update(mapping); err != nil {
		return err
	}

	uc.refreshRoutingCache(mapping.ProductID)
	return nil
}

func (uc *productUsecase) GetProductMappings(productID string) ([]*domain.ProductMapping, error) {
	return uc.productMappingRepo.GetByProductID(productID)
}

func (uc *productUsecase) GetProductMapping(id string) (*domain.ProductMapping, error) {
	return uc.productMappingRepo.GetByID(id)
}

func (uc *productUsecase) CreateProductMapping(mapping *domain.ProductMapping) error {
	if mapping == nil {
		return fmt.Errorf("mapping payload is required")
	}
	if mapping.ProductID == "" || mapping.SupplierID == "" {
		return fmt.Errorf("product_id and supplier_id are required")
	}

	if _, err := uc.productRepo.GetByID(mapping.ProductID); err != nil {
		return err
	}
	if _, err := uc.supplierRepo.GetByID(mapping.SupplierID); err != nil {
		return err
	}

	mapping.ID = utils.GenerateUUID()
	mapping.CreatedAt = time.Now()
	mapping.UpdatedAt = time.Now()

	if err := uc.productMappingRepo.Create(mapping); err != nil {
		return err
	}

	uc.refreshRoutingCache(mapping.ProductID)
	return nil
}

func (uc *productUsecase) DeleteProductMapping(id string) error {
	mapping, err := uc.productMappingRepo.GetByID(id)
	if err != nil {
		return err
	}

	if err := uc.productMappingRepo.Delete(id); err != nil {
		return err
	}

	if mapping != nil {
		uc.refreshRoutingCache(mapping.ProductID)
	}

	return nil
}

func (uc *productUsecase) refreshRoutingCache(productID string) {
	if uc.smartRoutingUC == nil {
		return
	}
	if productID == "" {
		return
	}

	if _, err := uc.smartRoutingUC.GetBestSupplier(productID, nil); err != nil {
		logger.Warn("Smart routing refresh failed",
			logger.String("product_id", productID),
			logger.ErrorField(err),
		)
	}
}
