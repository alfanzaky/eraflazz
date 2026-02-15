package usecase

import (
	"fmt"
	"sort"

	"github.com/alfanzaky/eraflazz/internal/domain"
	"github.com/alfanzaky/eraflazz/pkg/logger"
)

type smartRoutingUsecase struct {
	productRepo        domain.ProductRepository
	supplierRepo       domain.SupplierRepository
	productMappingRepo domain.ProductMappingRepository
}

// NewSmartRoutingUsecase creates a new smart routing use case
func NewSmartRoutingUsecase(
	productRepo domain.ProductRepository,
	supplierRepo domain.SupplierRepository,
	productMappingRepo domain.ProductMappingRepository,
) *smartRoutingUsecase {
	return &smartRoutingUsecase{
		productRepo:        productRepo,
		supplierRepo:       supplierRepo,
		productMappingRepo: productMappingRepo,
	}
}

// RoutingResult represents the result of routing decision
type RoutingResult struct {
	SelectedSupplier *domain.Supplier
	SelectedMapping  *domain.ProductMapping
	Confidence       float64 // 0.0 to 1.0
	Reason           string
	Alternatives     []*domain.Supplier // Backup suppliers
}

// RoutingCriteria defines criteria for routing decision
type RoutingCriteria struct {
	PriorityOnly   bool    // Only use priority, ignore other factors
	PreferCheapest bool    // Prefer cheapest price
	PreferFastest  bool    // Prefer fastest response time
	PreferReliable bool    // Prefer highest success rate
	MaxSuppliers   int     // Maximum number of suppliers to consider
	MinSuccessRate float64 // Minimum success rate threshold
}

// GetBestSupplier finds the best supplier for a product using smart routing
func (uc *smartRoutingUsecase) GetBestSupplier(productID string, criteria *RoutingCriteria) (*RoutingResult, error) {
	// Get product mappings for this product
	mappings, err := uc.productMappingRepo.GetActiveMappings(productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product mappings: %w", err)
	}

	if len(mappings) == 0 {
		return nil, fmt.Errorf("no active mappings found for product")
	}

	// Get supplier information for each mapping
	suppliers := make([]*domain.Supplier, 0, len(mappings))
	supplierMap := make(map[string]*domain.Supplier)

	for _, mapping := range mappings {
		supplier, err := uc.supplierRepo.GetByID(mapping.SupplierID)
		if err != nil {
			logger.Warn("Failed to get supplier for mapping",
				logger.String("supplier_id", mapping.SupplierID),
				logger.ErrorField(err),
			)
			continue
		}

		// Check if supplier is healthy
		if !supplier.IsHealthy() {
			logger.Debug("Skipping unhealthy supplier",
				logger.String("supplier_id", supplier.ID),
				logger.String("supplier_code", supplier.Code),
			)
			continue
		}

		suppliers = append(suppliers, supplier)
		supplierMap[supplier.ID] = supplier
	}

	if len(suppliers) == 0 {
		return nil, fmt.Errorf("no healthy suppliers available")
	}

	// Apply default criteria if not provided
	if criteria == nil {
		criteria = &RoutingCriteria{
			PreferCheapest: true,
			PreferReliable: true,
			MaxSuppliers:   5,
			MinSuccessRate: 50.0,
		}
	}

	// Score suppliers based on criteria
	scores := make([]*SupplierScore, 0, len(suppliers))
	for _, supplier := range suppliers {
		score := uc.calculateSupplierScore(supplier, mappings, criteria)
		scores = append(scores, score)
	}

	// Sort by score (highest first)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].TotalScore > scores[j].TotalScore
	})

	// Get the best supplier
	bestScore := scores[0]
	bestSupplier := bestScore.Supplier

	// Find the corresponding mapping
	var bestMapping *domain.ProductMapping
	for _, mapping := range mappings {
		if mapping.SupplierID == bestSupplier.ID {
			bestMapping = mapping
			break
		}
	}

	// Prepare alternatives (backup suppliers)
	alternatives := make([]*domain.Supplier, 0)
	for i := 1; i < len(scores) && i < criteria.MaxSuppliers-1; i++ {
		alternatives = append(alternatives, scores[i].Supplier)
	}

	result := &RoutingResult{
		SelectedSupplier: bestSupplier,
		SelectedMapping:  bestMapping,
		Confidence:       bestScore.Confidence,
		Reason:           bestScore.Reason,
		Alternatives:     alternatives,
	}

	logger.Info("Smart routing decision made",
		logger.String("product_id", productID),
		logger.String("selected_supplier", bestSupplier.Code),
		logger.Float64("confidence", bestScore.Confidence),
		logger.String("reason", bestScore.Reason),
		logger.Int("alternatives_count", len(alternatives)),
	)

	return result, nil
}

// SupplierScore represents the scoring result for a supplier
type SupplierScore struct {
	Supplier   *domain.Supplier
	TotalScore float64
	Confidence float64
	Reason     string
	Breakdown  map[string]float64
}

// calculateSupplierScore calculates a comprehensive score for a supplier
func (uc *smartRoutingUsecase) calculateSupplierScore(
	supplier *domain.Supplier,
	mappings []*domain.ProductMapping,
	criteria *RoutingCriteria,
) *SupplierScore {
	score := &SupplierScore{
		Supplier:  supplier,
		Breakdown: make(map[string]float64),
	}

	// Find mapping for this supplier
	var mapping *domain.ProductMapping
	for _, m := range mappings {
		if m.SupplierID == supplier.ID {
			mapping = m
			break
		}
	}

	if mapping == nil {
		score.Reason = "No mapping found"
		return score
	}

	// Priority score (lower priority number = higher score)
	priorityScore := 1.0 / float64(supplier.Priority)
	score.Breakdown["priority"] = priorityScore

	// Success rate score
	successRateScore := supplier.SuccessRate / 100.0
	score.Breakdown["success_rate"] = successRateScore

	// Response time score (inverse - faster is better)
	responseTimeScore := 1.0
	if supplier.AvgResponseTimeMs > 0 {
		responseTimeScore = 10000.0 / float64(supplier.AvgResponseTimeMs)
	}
	score.Breakdown["response_time"] = responseTimeScore

	// Price score (lower price = higher score)
	priceScore := 1.0
	if criteria.PreferCheapest && mapping.SupplierPrice > 0 {
		// Find the minimum price among all mappings
		minPrice := mapping.SupplierPrice
		for _, m := range mappings {
			if m.SupplierPrice < minPrice && m.IsActive {
				minPrice = m.SupplierPrice
			}
		}
		priceScore = minPrice / mapping.SupplierPrice
	}
	score.Breakdown["price"] = priceScore

	// Stock availability score
	stockScore := 1.0
	if mapping.StockStatus == domain.StockStatusOutOfStock {
		stockScore = 0.0
	} else if mapping.StockStatus == domain.StockStatusUnknown {
		stockScore = 0.5
	}
	score.Breakdown["stock"] = stockScore

	// Recent performance score (based on recent success/failure)
	recentPerformanceScore := uc.calculateRecentPerformanceScore(mapping)
	score.Breakdown["recent_performance"] = recentPerformanceScore

	// Calculate weighted total score
	var weights map[string]float64
	if criteria.PriorityOnly {
		weights = map[string]float64{
			"priority": 1.0,
		}
	} else {
		weights = map[string]float64{
			"priority":           0.3,
			"success_rate":       0.3,
			"response_time":      0.2,
			"price":              0.1,
			"stock":              0.05,
			"recent_performance": 0.05,
		}

		// Adjust weights based on preferences
		if criteria.PreferCheapest {
			weights["price"] = 0.3
			weights["priority"] = 0.2
		}
		if criteria.PreferFastest {
			weights["response_time"] = 0.4
			weights["priority"] = 0.2
		}
		if criteria.PreferReliable {
			weights["success_rate"] = 0.5
			weights["priority"] = 0.2
		}
	}

	totalScore := 0.0
	for factor, factorScore := range score.Breakdown {
		weight, exists := weights[factor]
		if exists {
			totalScore += factorScore * weight
		}
	}

	score.TotalScore = totalScore

	// Calculate confidence based on data availability and consistency
	confidence := uc.calculateConfidence(score, supplier, mapping)
	score.Confidence = confidence

	// Generate reason
	score.Reason = uc.generateReason(score, criteria)

	return score
}

// calculateRecentPerformanceScore calculates performance based on recent transactions
func (uc *smartRoutingUsecase) calculateRecentPerformanceScore(mapping *domain.ProductMapping) float64 {
	totalAttempts := mapping.SuccessCount + mapping.FailureCount
	if totalAttempts == 0 {
		return 0.5 // Neutral score for no data
	}

	// Weight recent performance more heavily
	recentSuccessRate := float64(mapping.SuccessCount) / float64(totalAttempts)

	// Apply bonus for consistent success
	if totalAttempts >= 10 && recentSuccessRate >= 0.95 {
		return 1.0
	}

	return recentSuccessRate
}

// calculateConfidence calculates confidence in the routing decision
func (uc *smartRoutingUsecase) calculateConfidence(score *SupplierScore, supplier *domain.Supplier, mapping *domain.ProductMapping) float64 {
	confidence := 0.5 // Base confidence

	// Increase confidence based on supplier reliability
	if supplier.SuccessRate >= 95 {
		confidence += 0.2
	} else if supplier.SuccessRate >= 90 {
		confidence += 0.1
	}

	// Increase confidence based on data volume
	totalTransactions := supplier.TotalTransactions
	if totalTransactions >= 1000 {
		confidence += 0.2
	} else if totalTransactions >= 100 {
		confidence += 0.1
	}

	// Increase confidence based on recent performance
	if score.Breakdown["recent_performance"] >= 0.9 {
		confidence += 0.1
	}

	// Decrease confidence for unknown stock status
	if mapping.StockStatus == domain.StockStatusUnknown {
		confidence -= 0.1
	}

	// Ensure confidence is within bounds
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// generateReason generates a human-readable reason for the routing decision
func (uc *smartRoutingUsecase) generateReason(score *SupplierScore, criteria *RoutingCriteria) string {
	reasons := []string{}

	// Priority
	if score.Breakdown["priority"] >= 0.8 {
		reasons = append(reasons, "highest priority")
	}

	// Success rate
	if score.Breakdown["success_rate"] >= 0.9 {
		reasons = append(reasons, "excellent success rate")
	} else if score.Breakdown["success_rate"] >= 0.8 {
		reasons = append(reasons, "good success rate")
	}

	// Response time
	if score.Breakdown["response_time"] >= 0.8 {
		reasons = append(reasons, "fast response")
	}

	// Price
	if criteria.PreferCheapest && score.Breakdown["price"] >= 0.9 {
		reasons = append(reasons, "best price")
	}

	// Stock
	if score.Breakdown["stock"] >= 0.8 {
		reasons = append(reasons, "stock available")
	}

	if len(reasons) == 0 {
		return "selected by algorithm"
	}

	// Combine reasons (limit to 3 for readability)
	if len(reasons) > 3 {
		reasons = reasons[:3]
	}

	result := reasons[0]
	for i := 1; i < len(reasons); i++ {
		if i == len(reasons)-1 {
			result += " and " + reasons[i]
		} else {
			result += ", " + reasons[i]
		}
	}

	return result
}

// GetRoutingStats returns statistics about routing decisions
func (uc *smartRoutingUsecase) GetRoutingStats() (*RoutingStats, error) {
	// Get all active suppliers
	suppliers, err := uc.supplierRepo.GetActiveSuppliers()
	if err != nil {
		return nil, fmt.Errorf("failed to get suppliers: %w", err)
	}

	stats := &RoutingStats{
		TotalSuppliers:    len(suppliers),
		HealthySuppliers:  0,
		AvgSuccessRate:    0,
		AvgResponseTime:   0,
		SupplierBreakdown: make(map[string]*SupplierStats),
	}

	totalSuccessRate := 0.0
	totalResponseTime := 0.0

	for _, supplier := range suppliers {
		if supplier.IsHealthy() {
			stats.HealthySuppliers++
		}

		totalSuccessRate += supplier.SuccessRate
		totalResponseTime += float64(supplier.AvgResponseTimeMs)

		stats.SupplierBreakdown[supplier.Code] = &SupplierStats{
			Code:              supplier.Code,
			SuccessRate:       supplier.SuccessRate,
			ResponseTimeMs:    supplier.AvgResponseTimeMs,
			TotalTransactions: supplier.TotalTransactions,
			IsHealthy:         supplier.IsHealthy(),
		}
	}

	if len(suppliers) > 0 {
		stats.AvgSuccessRate = totalSuccessRate / float64(len(suppliers))
		stats.AvgResponseTime = totalResponseTime / float64(len(suppliers))
	}

	return stats, nil
}

// RoutingStats represents routing statistics
type RoutingStats struct {
	TotalSuppliers    int
	HealthySuppliers  int
	AvgSuccessRate    float64
	AvgResponseTime   float64
	SupplierBreakdown map[string]*SupplierStats
}

// SupplierStats represents individual supplier statistics
type SupplierStats struct {
	Code              string
	SuccessRate       float64
	ResponseTimeMs    int
	TotalTransactions int
	IsHealthy         bool
}

// UpdateSupplierMetrics updates supplier metrics after a transaction
func (uc *smartRoutingUsecase) UpdateSupplierMetrics(supplierID string, success bool, responseTimeMs int) error {
	return uc.supplierRepo.UpdateMetrics(supplierID, success, responseTimeMs)
}

// GetFallbackSuppliers returns a list of fallback suppliers for a product
func (uc *smartRoutingUsecase) GetFallbackSuppliers(productID string, excludeSupplierID string, maxCount int) ([]*domain.Supplier, error) {
	result, err := uc.GetBestSupplier(productID, &RoutingCriteria{
		MaxSuppliers: maxCount + 1, // +1 to account for excluded supplier
	})
	if err != nil {
		return nil, err
	}

	fallbacks := make([]*domain.Supplier, 0)

	// Add the selected supplier if it's not the excluded one
	if result.SelectedSupplier.ID != excludeSupplierID {
		fallbacks = append(fallbacks, result.SelectedSupplier)
	}

	// Add alternatives, excluding the specified supplier
	for _, alt := range result.Alternatives {
		if alt.ID != excludeSupplierID && len(fallbacks) < maxCount {
			fallbacks = append(fallbacks, alt)
		}
	}

	return fallbacks, nil
}
