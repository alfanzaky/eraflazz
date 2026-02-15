package digiflazz

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alfanzaky/eraflazz/config"
	"github.com/alfanzaky/eraflazz/internal/domain"
)

const (
	transactionEndpoint = "/transaction"
	balanceEndpoint     = "/cek-saldo"
	priceListEndpoint   = "/price-list"
)

var (
	statusSuccess = "Sukses"
	statusPending = "Pending"
)

// Adapter implements domain.SupplierAdapter for Digiflazz
// It translates domain abstraction into concrete Digiflazz HTTP calls
// while keeping signature generation, timeout, and payload structure in one place.
type Adapter struct {
	cfg        config.DigiflazzConfig
	httpClient *http.Client
	timeout    time.Duration
}

// NewAdapter creates a new Digiflazz adapter instance
func NewAdapter(cfg config.DigiflazzConfig, client *http.Client) *Adapter {
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	if client == nil {
		client = &http.Client{Timeout: timeout}
	}

	return &Adapter{
		cfg:        cfg,
		httpClient: client,
		timeout:    timeout,
	}
}

// TopUp sends a top-up request to Digiflazz
func (a *Adapter) TopUp(request *domain.SupplierRequest) (*domain.SupplierResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("supplier request is required")
	}

	payload := &topUpRequest{
		Username:     a.cfg.Username,
		BuyerSkuCode: request.ProductCode,
		CustomerNo:   request.DestinationNumber,
		RefID:        request.RefID,
		Sign:         a.generateSignature(request.RefID),
		Testing:      a.cfg.Testing,
	}

	if request.AdditionalData != nil {
		if qty, ok := request.AdditionalData["quantity"]; ok {
			payload.Quantity = qty
		}
		if msg, ok := request.AdditionalData["message"]; ok {
			payload.Message = msg
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	start := time.Now()
	var response digiflazzTransactionResponse
	if err := a.doPost(ctx, transactionEndpoint, payload, &response); err != nil {
		return nil, err
	}

	duration := time.Since(start)
	return a.mapTransactionResponse(&response, duration)
}

// CheckBalance returns current Digiflazz deposit balance
func (a *Adapter) CheckBalance() (float64, error) {
	payload := map[string]string{
		"cmd":      "deposit",
		"username": a.cfg.Username,
		"sign":     a.generateSignature("deposit"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	var response digiflazzBalanceResponse
	if err := a.doPost(ctx, balanceEndpoint, payload, &response); err != nil {
		return 0, err
	}

	if response.Data == nil {
		return 0, fmt.Errorf("digiflazz balance data is empty")
	}

	return response.Data.Deposit, nil
}

// CheckStatus fetches transaction status by reference ID
func (a *Adapter) CheckStatus(refID string) (*domain.SupplierResponse, error) {
	if strings.TrimSpace(refID) == "" {
		return nil, fmt.Errorf("ref id is required")
	}

	payload := map[string]string{
		"username": a.cfg.Username,
		"ref_id":   refID,
		"sign":     a.generateSignature(refID),
		"type":     "status",
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	start := time.Now()
	var response digiflazzTransactionResponse
	if err := a.doPost(ctx, transactionEndpoint, payload, &response); err != nil {
		return nil, err
	}

	duration := time.Since(start)
	return a.mapTransactionResponse(&response, duration)
}

// GetProductCatalog pulls Digiflazz price list
func (a *Adapter) GetProductCatalog() ([]*domain.Product, error) {
	payload := map[string]string{
		"cmd":      "prepaid",
		"username": a.cfg.Username,
		"sign":     a.generateSignature("pricelist"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	var response digiflazzPriceListResponse
	if err := a.doPost(ctx, priceListEndpoint, payload, &response); err != nil {
		return nil, err
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("digiflazz price list empty")
	}

	products := make([]*domain.Product, 0, len(response.Data))
	for _, item := range response.Data {
		products = append(products, item.toDomainProduct())
	}

	return products, nil
}

// ParseResponse converts raw JSON into SupplierResponse
func (a *Adapter) ParseResponse(raw []byte) (*domain.SupplierResponse, error) {
	var response digiflazzTransactionResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, fmt.Errorf("failed to parse digiflazz response: %w", err)
	}

	return a.mapTransactionResponse(&response, 0)
}

// Helper: perform HTTP POST and decode JSON response
type httpPayload interface{}

func (a *Adapter) doPost(ctx context.Context, path string, payload httpPayload, target interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.endpoint(path), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("digiflazz request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("digiflazz returned status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode digiflazz response: %w", err)
	}

	return nil
}

func (a *Adapter) endpoint(path string) string {
	base := strings.TrimRight(a.cfg.BaseURL, "/")
	return base + path
}

func (a *Adapter) mapTransactionResponse(resp *digiflazzTransactionResponse, duration time.Duration) (*domain.SupplierResponse, error) {
	if resp == nil {
		return nil, fmt.Errorf("empty digiflazz response")
	}

	if resp.Data == nil {
		return nil, fmt.Errorf("digiflazz response missing data: %s", resp.Message)
	}

	success := strings.EqualFold(resp.Data.Status, statusSuccess)
	statusCode := http.StatusAccepted
	switch strings.ToLower(resp.Data.Status) {
	case strings.ToLower(statusSuccess):
		statusCode = http.StatusOK
	case strings.ToLower(statusPending):
		statusCode = http.StatusAccepted
	default:
		statusCode = http.StatusBadGateway
	}

	serial := resp.Data.Sn
	if serial == "" {
		serial = resp.Data.SerialNumber
	}

	dataMap := map[string]interface{}{
		"status":           resp.Data.Status,
		"buyer_sku_code":   resp.Data.BuyerSkuCode,
		"customer_no":      resp.Data.CustomerNo,
		"price":            resp.Data.Price,
		"sell_price":       resp.Data.SellingPrice,
		"buyer_last_saldo": resp.Data.BuyerLastSaldo,
		"tele":             resp.Data.Tele,
		"rc":               resp.Data.ResponseCode,
		"message":          resp.Data.Message,
	}

	return &domain.SupplierResponse{
		Success:      success,
		Message:      resp.Message,
		TrxID:        resp.Data.RefID,
		SerialNumber: serial,
		StatusCode:   statusCode,
		ResponseTime: int(duration.Milliseconds()),
		Data:         dataMap,
	}, nil
}

func (a *Adapter) generateSignature(seed string) string {
	builder := a.cfg.Username + a.cfg.APIKey + seed
	sum := md5.Sum([]byte(builder))
	return hex.EncodeToString(sum[:])
}

// --- Digiflazz DTOs ---

type topUpRequest struct {
	Username     string            `json:"username"`
	BuyerSkuCode string            `json:"buyer_sku_code"`
	CustomerNo   string            `json:"customer_no"`
	RefID        string            `json:"ref_id"`
	Sign         string            `json:"sign"`
	Testing      bool              `json:"testing"`
	Quantity     interface{}       `json:"quantity,omitempty"`
	Message      interface{}       `json:"message,omitempty"`
	Extra        map[string]string `json:"extra,omitempty"`
}

type digiflazzTransactionResponse struct {
	Success bool                      `json:"success"`
	Message string                    `json:"message"`
	Data    *digiflazzTransactionData `json:"data"`
}

type digiflazzTransactionData struct {
	RefID          string  `json:"ref_id"`
	Status         string  `json:"status"`
	Sn             string  `json:"sn"`
	SerialNumber   string  `json:"serial_number"`
	BuyerSkuCode   string  `json:"buyer_sku_code"`
	CustomerNo     string  `json:"customer_no"`
	Price          float64 `json:"price"`
	SellingPrice   float64 `json:"selling_price"`
	BuyerLastSaldo float64 `json:"buyer_last_saldo"`
	Tele           string  `json:"tele"`
	ResponseCode   string  `json:"rc"`
	Message        string  `json:"message"`
}

type digiflazzBalanceResponse struct {
	Data *struct {
		Deposit float64 `json:"deposit"`
	} `json:"data"`
}

type digiflazzPriceListResponse struct {
	Data []*digiflazzPriceListItem `json:"data"`
}

type digiflazzPriceListItem struct {
	BuyerSkuCode string  `json:"buyer_sku_code"`
	ProductName  string  `json:"product_name"`
	Category     string  `json:"category"`
	Type         string  `json:"type"`
	SellerName   string  `json:"seller_name"`
	Brand        string  `json:"brand"`
	Price        float64 `json:"price"`
	SellerPrice  float64 `json:"seller_price"`
	Status       string  `json:"status"`
}

func (item *digiflazzPriceListItem) toDomainProduct() *domain.Product {
	product := &domain.Product{
		ID:           item.BuyerSkuCode,
		Code:         item.BuyerSkuCode,
		Name:         item.ProductName,
		Category:     strings.ToUpper(item.Category),
		Provider:     item.Brand,
		Type:         strings.ToUpper(item.Type),
		BasePrice:    item.SellerPrice,
		SellingPrice: item.Price,
		IsActive:     strings.EqualFold(item.Status, "active"),
	}

	return product
}
