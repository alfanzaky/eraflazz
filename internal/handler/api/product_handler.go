package api

import (
	"strconv"
	"strings"

	"github.com/alfanzaky/eraflazz/internal/domain"
	"github.com/alfanzaky/eraflazz/pkg/logger"
	"github.com/alfanzaky/eraflazz/pkg/xresponse"
	"github.com/gin-gonic/gin"
)

// ProductHandler handles admin product endpoints
type ProductHandler struct {
	productUC domain.ProductUsecase
	roleGuard *RoleGuard
}

// NewProductHandler creates a new product handler
func NewProductHandler(productUC domain.ProductUsecase) *ProductHandler {
	return &ProductHandler{
		productUC: productUC,
		roleGuard: NewRoleGuard(),
	}
}

// ProductResponse represents product payload returned to clients
type ProductResponse struct {
	ID                   string   `json:"id"`
	Code                 string   `json:"code"`
	Name                 string   `json:"name"`
	Description          *string  `json:"description,omitempty"`
	Category             string   `json:"category"`
	Provider             string   `json:"provider"`
	Type                 string   `json:"type"`
	BasePrice            float64  `json:"base_price"`
	SellingPrice         float64  `json:"selling_price"`
	MinPrice             float64  `json:"min_price"`
	Nominal              *float64 `json:"nominal,omitempty"`
	ValidityPeriod       *string  `json:"validity_period,omitempty"`
	IsActive             bool     `json:"is_active"`
	IsUnlimitedStock     bool     `json:"is_unlimited_stock"`
	StockQuantity        int      `json:"stock_quantity"`
	AllowMarkup          bool     `json:"allow_markup"`
	MaxMarkupPercentage  float64  `json:"max_markup_percentage"`
	MinTransactionAmount float64  `json:"min_transaction_amount"`
	MaxTransactionAmount float64  `json:"max_transaction_amount"`
}

// CreateProductRequest payload
type CreateProductRequest struct {
	Code                 string   `json:"code" binding:"required"`
	Name                 string   `json:"name" binding:"required"`
	Description          *string  `json:"description"`
	Category             string   `json:"category" binding:"required"`
	Provider             string   `json:"provider" binding:"required"`
	Type                 string   `json:"type" binding:"required"`
	BasePrice            float64  `json:"base_price" binding:"required"`
	SellingPrice         float64  `json:"selling_price" binding:"required"`
	MinPrice             float64  `json:"min_price" binding:"required"`
	Nominal              *float64 `json:"nominal"`
	ValidityPeriod       *string  `json:"validity_period"`
	AllowMarkup          bool     `json:"allow_markup"`
	MaxMarkupPercentage  float64  `json:"max_markup_percentage"`
	MinTransactionAmount float64  `json:"min_transaction_amount"`
	MaxTransactionAmount float64  `json:"max_transaction_amount"`
}

// UpdateProductRequest payload
type UpdateProductRequest struct {
	Name                 *string  `json:"name"`
	Description          *string  `json:"description"`
	Category             *string  `json:"category"`
	Provider             *string  `json:"provider"`
	Type                 *string  `json:"type"`
	BasePrice            *float64 `json:"base_price"`
	SellingPrice         *float64 `json:"selling_price"`
	MinPrice             *float64 `json:"min_price"`
	Nominal              *float64 `json:"nominal"`
	ValidityPeriod       *string  `json:"validity_period"`
	AllowMarkup          *bool    `json:"allow_markup"`
	MaxMarkupPercentage  *float64 `json:"max_markup_percentage"`
	MinTransactionAmount *float64 `json:"min_transaction_amount"`
	MaxTransactionAmount *float64 `json:"max_transaction_amount"`
}

// ToggleStatusRequest payload
type ToggleStatusRequest struct {
	IsActive bool `json:"is_active" binding:"required"`
}

// UpdateStockRequest payload
type UpdateStockRequest struct {
	StockQuantity int  `json:"stock_quantity" binding:"required"`
	IsUnlimited   bool `json:"is_unlimited_stock"`
}

// CreateMappingRequest payload
type CreateMappingRequest struct {
	SupplierID          string  `json:"supplier_id" binding:"required"`
	SupplierProductCode string  `json:"supplier_product_code" binding:"required"`
	SupplierPrice       float64 `json:"supplier_price" binding:"required"`
	AdditionalFee       float64 `json:"additional_fee"`
	Priority            int     `json:"priority" binding:"required"`
	IsActive            bool    `json:"is_active"`
	StockStatus         string  `json:"stock_status" binding:"required"`
}

// UpdateMappingRequest payload
type UpdateMappingRequest struct {
	SupplierProductCode *string  `json:"supplier_product_code"`
	SupplierPrice       *float64 `json:"supplier_price"`
	AdditionalFee       *float64 `json:"additional_fee"`
	Priority            *int     `json:"priority"`
	IsActive            *bool    `json:"is_active"`
	StockStatus         *string  `json:"stock_status"`
}

// CreateProduct handles creating a new product
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	// Check if user has admin privileges
	userID, role, _, exists := h.roleGuard.GetCurrentUser(c)
	if !exists || role != domain.RoleAdmin {
		logger.Warn("Access denied - insufficient privileges for create product",
			logger.String("user_id", userID),
			logger.String("role", role),
			logger.String("ip", c.ClientIP()),
		)
		xresponse.Forbidden(c, "Admin access required to create products")
		return
	}

	h.roleGuard.LogAccess(c, "create_product", "admin")

	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		xresponse.ValidationError(c, err.Error())
		return
	}

	product := &domain.Product{
		Code:                 strings.ToUpper(req.Code),
		Name:                 req.Name,
		Description:          req.Description,
		Category:             strings.ToUpper(req.Category),
		Provider:             req.Provider,
		Type:                 strings.ToUpper(req.Type),
		BasePrice:            req.BasePrice,
		SellingPrice:         req.SellingPrice,
		MinPrice:             req.MinPrice,
		Nominal:              req.Nominal,
		ValidityPeriod:       req.ValidityPeriod,
		AllowMarkup:          req.AllowMarkup,
		MaxMarkupPercentage:  req.MaxMarkupPercentage,
		MinTransactionAmount: req.MinTransactionAmount,
		MaxTransactionAmount: req.MaxTransactionAmount,
		IsActive:             true,
	}

	if err := h.productUC.CreateProduct(product); err != nil {
		logger.Error("Failed to create product", logger.ErrorField(err))
		xresponse.BadRequest(c, err.Error())
		return
	}

	xresponse.Created(c, "Product created", h.toProductResponse(product))
}

// ListProducts returns products using filters
func (h *ProductHandler) ListProducts(c *gin.Context) {
	filter := &domain.ProductFilter{}

	if v := c.Query("category"); v != "" {
		filter.Category = &v
	}
	if v := c.Query("provider"); v != "" {
		filter.Provider = &v
	}
	if v := c.Query("query"); v != "" {
		filter.Query = &v
	}
	if v := c.Query("is_active"); v != "" {
		if isActive, err := strconv.ParseBool(v); err == nil {
			filter.IsActive = &isActive
		}
	}
	if v := c.Query("page"); v != "" {
		if page, err := strconv.Atoi(v); err == nil && page > 0 {
			filter.Page = page
		}
	}
	if v := c.Query("page_size"); v != "" {
		if size, err := strconv.Atoi(v); err == nil && size > 0 {
			filter.PageSize = size
		}
	}

	products, total, err := h.productUC.ListProducts(filter)
	if err != nil {
		logger.Error("Failed to list products", logger.ErrorField(err))
		xresponse.InternalServerError(c, "Failed to list products")
		return
	}

	responses := make([]*ProductResponse, 0, len(products))
	for _, p := range products {
		responses = append(responses, h.toProductResponse(p))
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	limit := filter.PageSize
	if limit <= 0 {
		limit = 50
	}

	xresponse.Paginated(c, "Products fetched", responses, page, limit, total)
}

// GetProduct returns a product by ID
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		xresponse.BadRequest(c, "product id is required")
		return
	}

	product, err := h.productUC.GetProduct(id)
	if err != nil {
		xresponse.NotFound(c, err.Error())
		return
	}

	xresponse.Success(c, "Product fetched", h.toProductResponse(product))
}

// UpdateProduct updates a product
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		xresponse.BadRequest(c, "product id is required")
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		xresponse.ValidationError(c, err.Error())
		return
	}

	updates := &domain.Product{}
	if req.Name != nil {
		updates.Name = *req.Name
	}
	updates.Description = req.Description
	if req.Category != nil {
		updates.Category = strings.ToUpper(*req.Category)
	}
	if req.Provider != nil {
		updates.Provider = *req.Provider
	}
	if req.Type != nil {
		updates.Type = strings.ToUpper(*req.Type)
	}
	if req.BasePrice != nil {
		updates.BasePrice = *req.BasePrice
	}
	if req.SellingPrice != nil {
		updates.SellingPrice = *req.SellingPrice
	}
	if req.MinPrice != nil {
		updates.MinPrice = *req.MinPrice
	}
	updates.Nominal = req.Nominal
	updates.ValidityPeriod = req.ValidityPeriod
	if req.AllowMarkup != nil {
		updates.AllowMarkup = *req.AllowMarkup
	}
	if req.MaxMarkupPercentage != nil {
		updates.MaxMarkupPercentage = *req.MaxMarkupPercentage
	}
	if req.MinTransactionAmount != nil {
		updates.MinTransactionAmount = *req.MinTransactionAmount
	}
	if req.MaxTransactionAmount != nil {
		updates.MaxTransactionAmount = *req.MaxTransactionAmount
	}

	if err := h.productUC.UpdateProduct(id, updates); err != nil {
		xresponse.BadRequest(c, err.Error())
		return
	}

	product, _ := h.productUC.GetProduct(id)
	xresponse.Success(c, "Product updated", h.toProductResponse(product))
}

// ToggleProductStatus updates product activation state
func (h *ProductHandler) ToggleProductStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		xresponse.BadRequest(c, "product id is required")
		return
	}

	var req ToggleStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		xresponse.ValidationError(c, err.Error())
		return
	}

	if err := h.productUC.ToggleProductStatus(id, req.IsActive); err != nil {
		xresponse.BadRequest(c, err.Error())
		return
	}

	xresponse.Success(c, "Product status updated", gin.H{"product_id": id, "is_active": req.IsActive})
}

// UpdateProductStock updates stock info
func (h *ProductHandler) UpdateProductStock(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		xresponse.BadRequest(c, "product id is required")
		return
	}

	var req UpdateStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		xresponse.ValidationError(c, err.Error())
		return
	}

	if err := h.productUC.UpdateProductStock(id, req.StockQuantity, req.IsUnlimited); err != nil {
		xresponse.BadRequest(c, err.Error())
		return
	}

	xresponse.Success(c, "Product stock updated", gin.H{
		"product_id":     id,
		"stock_quantity": req.StockQuantity,
		"is_unlimited":   req.IsUnlimited,
	})
}

// ListProductMappings returns mappings for a product
func (h *ProductHandler) ListProductMappings(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		xresponse.BadRequest(c, "product id is required")
		return
	}

	mappings, err := h.productUC.GetProductMappings(productID)
	if err != nil {
		xresponse.BadRequest(c, err.Error())
		return
	}

	xresponse.Success(c, "Product mappings fetched", mappings)
}

// CreateProductMapping adds a mapping for a product
func (h *ProductHandler) CreateProductMapping(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		xresponse.BadRequest(c, "product id is required")
		return
	}

	var req CreateMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		xresponse.ValidationError(c, err.Error())
		return
	}

	mapping := &domain.ProductMapping{
		ProductID:           productID,
		SupplierID:          req.SupplierID,
		SupplierProductCode: req.SupplierProductCode,
		SupplierPrice:       req.SupplierPrice,
		AdditionalFee:       req.AdditionalFee,
		Priority:            req.Priority,
		IsActive:            req.IsActive,
		StockStatus:         strings.ToUpper(req.StockStatus),
	}

	if err := h.productUC.CreateProductMapping(mapping); err != nil {
		xresponse.BadRequest(c, err.Error())
		return
	}

	xresponse.Created(c, "Product mapping created", mapping)
}

// UpdateProductMapping updates mapping fields
func (h *ProductHandler) UpdateProductMapping(c *gin.Context) {
	mappingID := c.Param("id")
	if mappingID == "" {
		xresponse.BadRequest(c, "mapping id is required")
		return
	}

	var req UpdateMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		xresponse.ValidationError(c, err.Error())
		return
	}

	mapping := &domain.ProductMapping{ID: mappingID}
	if req.SupplierProductCode != nil {
		mapping.SupplierProductCode = *req.SupplierProductCode
	}
	if req.SupplierPrice != nil {
		mapping.SupplierPrice = *req.SupplierPrice
	}
	if req.AdditionalFee != nil {
		mapping.AdditionalFee = *req.AdditionalFee
	}
	if req.Priority != nil {
		mapping.Priority = *req.Priority
	}
	if req.IsActive != nil {
		mapping.IsActive = *req.IsActive
	}
	if req.StockStatus != nil {
		mapping.StockStatus = strings.ToUpper(*req.StockStatus)
	}

	if err := h.productUC.UpdateProductMapping(mapping); err != nil {
		xresponse.BadRequest(c, err.Error())
		return
	}

	xresponse.Success(c, "Product mapping updated", mapping)
}

// DeleteProductMapping removes mapping
func (h *ProductHandler) DeleteProductMapping(c *gin.Context) {
	mappingID := c.Param("id")
	if mappingID == "" {
		xresponse.BadRequest(c, "mapping id is required")
		return
	}

	if err := h.productUC.DeleteProductMapping(mappingID); err != nil {
		xresponse.BadRequest(c, err.Error())
		return
	}

	xresponse.Success(c, "Product mapping deleted", gin.H{"mapping_id": mappingID})
}

func (h *ProductHandler) toProductResponse(product *domain.Product) *ProductResponse {
	return &ProductResponse{
		ID:                   product.ID,
		Code:                 product.Code,
		Name:                 product.Name,
		Description:          product.Description,
		Category:             product.Category,
		Provider:             product.Provider,
		Type:                 product.Type,
		BasePrice:            product.BasePrice,
		SellingPrice:         product.SellingPrice,
		MinPrice:             product.MinPrice,
		Nominal:              product.Nominal,
		ValidityPeriod:       product.ValidityPeriod,
		IsActive:             product.IsActive,
		IsUnlimitedStock:     product.IsUnlimitedStock,
		StockQuantity:        product.StockQuantity,
		AllowMarkup:          product.AllowMarkup,
		MaxMarkupPercentage:  product.MaxMarkupPercentage,
		MinTransactionAmount: product.MinTransactionAmount,
		MaxTransactionAmount: product.MaxTransactionAmount,
	}
}
