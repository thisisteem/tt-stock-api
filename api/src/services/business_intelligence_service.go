// Package services contains the business logic layer implementations for the TT Stock Backend API.
// It provides service interfaces and implementations that orchestrate repository operations
// and implement business rules and validation logic.
package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"tt-stock-api/src/models"
	"tt-stock-api/src/repositories"
)

// BusinessIntelligenceService defines the interface for business intelligence and reporting operations
type BusinessIntelligenceService interface {
	// GetDashboardData retrieves comprehensive dashboard data
	GetDashboardData(ctx context.Context, user *models.User) (*models.DashboardData, error)

	// GetSalesReport generates sales report for a date range
	GetSalesReport(ctx context.Context, req *models.SalesReportRequest, user *models.User) (*models.SalesReport, error)

	// GetInventoryReport generates inventory report
	GetInventoryReport(ctx context.Context, req *models.InventoryReportRequest, user *models.User) (*models.InventoryReport, error)

	// GetStockMovementReport generates stock movement report
	GetStockMovementReport(ctx context.Context, req *models.StockMovementReportRequest, user *models.User) (*models.StockMovementReport, error)

	// GetLowStockReport generates low stock report
	GetLowStockReport(ctx context.Context, user *models.User) (*models.LowStockReport, error)

	// GetTopSellingProducts retrieves top selling products
	GetTopSellingProducts(ctx context.Context, req *models.TopProductsRequest, user *models.User) (*models.TopProductsReport, error)

	// GetFinancialSummary generates financial summary
	GetFinancialSummary(ctx context.Context, req *models.FinancialSummaryRequest, user *models.User) (*models.FinancialSummary, error)
}

// businessIntelligenceService implements the BusinessIntelligenceService interface
type businessIntelligenceService struct {
	productRepo       repositories.ProductRepository
	stockMovementRepo repositories.StockMovementRepository
	userRepo          repositories.UserRepository
	alertRepo         repositories.AlertRepository
}

// NewBusinessIntelligenceService creates a new BusinessIntelligenceService instance
func NewBusinessIntelligenceService(
	productRepo repositories.ProductRepository,
	stockMovementRepo repositories.StockMovementRepository,
	userRepo repositories.UserRepository,
	alertRepo repositories.AlertRepository,
) BusinessIntelligenceService {
	return &businessIntelligenceService{
		productRepo:       productRepo,
		stockMovementRepo: stockMovementRepo,
		userRepo:          userRepo,
		alertRepo:         alertRepo,
	}
}

// GetDashboardData retrieves comprehensive dashboard data
func (s *businessIntelligenceService) GetDashboardData(ctx context.Context, user *models.User) (*models.DashboardData, error) {
	// Check permissions
	if !s.canViewReports(user) {
		return nil, errors.New("insufficient permissions to view dashboard data")
	}

	// Get product statistics
	productStats, err := s.getProductStatistics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get product statistics: %w", err)
	}

	// Get sales summary
	salesSummary, err := s.getSalesSummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales summary: %w", err)
	}

	// Get stock movement summary
	stockSummary, err := s.stockMovementRepo.GetMovementSummary(ctx, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock movement summary: %w", err)
	}

	// Get recent alerts
	recentAlerts, err := s.getRecentAlerts(ctx, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent alerts: %w", err)
	}

	// Get low stock products
	lowStockProducts, err := s.productRepo.GetLowStockProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock products: %w", err)
	}

	// Get user activity summary
	userActivity, err := s.getUserActivitySummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user activity summary: %w", err)
	}

	return &models.DashboardData{
		ProductStatistics:    *productStats,
		SalesSummary:         *salesSummary,
		StockMovementSummary: *stockSummary,
		RecentAlerts:         recentAlerts,
		TopSellingProducts:   []models.ProductSalesData{}, // TODO: Implement
		LowStockProducts:     s.convertProductsToResponses(lowStockProducts),
		UserActivity:         *userActivity,
		GeneratedAt:          time.Now(),
	}, nil
}

// GetSalesReport generates sales report for a date range
func (s *businessIntelligenceService) GetSalesReport(ctx context.Context, req *models.SalesReportRequest, user *models.User) (*models.SalesReport, error) {
	if req == nil {
		return nil, errors.New("sales report request cannot be nil")
	}

	// Check permissions
	if !s.canViewReports(user) {
		return nil, errors.New("insufficient permissions to view sales reports")
	}

	// Validate date range
	if err := s.validateDateRange(req.StartDate, req.EndDate); err != nil {
		return nil, err
	}

	// Get sales metrics
	salesMetrics, err := s.getSalesMetrics(ctx, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales metrics: %w", err)
	}

	return &models.SalesReport{
		Period:         models.DateRange{StartDate: req.StartDate, EndDate: req.EndDate},
		SalesMetrics:   *salesMetrics,
		SalesByProduct: []models.ProductSalesData{}, // TODO: Implement
		SalesByUser:    []models.UserSalesData{},    // TODO: Implement
		GeneratedAt:    time.Now(),
	}, nil
}

// GetInventoryReport generates inventory report
func (s *businessIntelligenceService) GetInventoryReport(ctx context.Context, req *models.InventoryReportRequest, user *models.User) (*models.InventoryReport, error) {
	if req == nil {
		return nil, errors.New("inventory report request cannot be nil")
	}

	// Check permissions
	if !s.canViewReports(user) {
		return nil, errors.New("insufficient permissions to view inventory reports")
	}

	// Get all products
	products, err := s.productRepo.GetActiveProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	// Calculate inventory metrics
	inventoryMetrics := s.calculateInventoryMetrics(products)

	return &models.InventoryReport{
		InventoryMetrics:        *inventoryMetrics,
		InventoryByCategory:     []models.CategoryInventoryData{}, // TODO: Implement
		InventoryByBrand:        []models.BrandInventoryData{},    // TODO: Implement
		StockStatusDistribution: s.calculateStockStatusDistribution(products),
		GeneratedAt:             time.Now(),
	}, nil
}

// GetStockMovementReport generates stock movement report
func (s *businessIntelligenceService) GetStockMovementReport(ctx context.Context, req *models.StockMovementReportRequest, user *models.User) (*models.StockMovementReport, error) {
	if req == nil {
		return nil, errors.New("stock movement report request cannot be nil")
	}

	// Check permissions
	if !s.canViewReports(user) {
		return nil, errors.New("insufficient permissions to view stock movement reports")
	}

	// Validate date range
	if err := s.validateDateRange(req.StartDate, req.EndDate); err != nil {
		return nil, err
	}

	// Get movement summary
	movementSummary, err := s.stockMovementRepo.GetMovementSummary(ctx, req.ProductID, &req.StartDate, &req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get movement summary: %w", err)
	}

	return &models.StockMovementReport{
		Period:             models.DateRange{StartDate: req.StartDate, EndDate: req.EndDate},
		MovementMetrics:    models.MovementMetrics(*movementSummary),
		MovementsByType:    []models.MovementTypeData{},    // TODO: Implement
		MovementsByProduct: []models.ProductMovementData{}, // TODO: Implement
		GeneratedAt:        time.Now(),
	}, nil
}

// GetLowStockReport generates low stock report
func (s *businessIntelligenceService) GetLowStockReport(ctx context.Context, user *models.User) (*models.LowStockReport, error) {
	// Check permissions
	if !s.canViewReports(user) {
		return nil, errors.New("insufficient permissions to view low stock report")
	}

	// Get low stock products
	lowStockProducts, err := s.productRepo.GetLowStockProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock products: %w", err)
	}

	// Get out of stock products
	outOfStockProducts, err := s.productRepo.GetOutOfStockProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get out of stock products: %w", err)
	}

	// Calculate low stock metrics
	lowStockMetrics := s.calculateLowStockMetrics(lowStockProducts, outOfStockProducts)

	return &models.LowStockReport{
		LowStockMetrics:    *lowStockMetrics,
		LowStockProducts:   s.convertProductsToResponses(lowStockProducts),
		OutOfStockProducts: s.convertProductsToResponses(outOfStockProducts),
		LowStockByCategory: []models.CategoryLowStockData{}, // TODO: Implement
		LowStockByBrand:    []models.BrandLowStockData{},    // TODO: Implement
		GeneratedAt:        time.Now(),
	}, nil
}

// GetTopSellingProducts retrieves top selling products
func (s *businessIntelligenceService) GetTopSellingProducts(ctx context.Context, req *models.TopProductsRequest, user *models.User) (*models.TopProductsReport, error) {
	if req == nil {
		return nil, errors.New("top products request cannot be nil")
	}

	// Check permissions
	if !s.canViewReports(user) {
		return nil, errors.New("insufficient permissions to view top selling products")
	}

	// Validate date range
	if err := s.validateDateRange(req.StartDate, req.EndDate); err != nil {
		return nil, err
	}

	// Get sales metrics
	salesMetrics, err := s.getSalesMetrics(ctx, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales metrics: %w", err)
	}

	return &models.TopProductsReport{
		Period:       models.DateRange{StartDate: req.StartDate, EndDate: req.EndDate},
		TopProducts:  []models.ProductSalesData{}, // TODO: Implement
		SalesMetrics: *salesMetrics,
		GeneratedAt:  time.Now(),
	}, nil
}

// GetFinancialSummary generates financial summary
func (s *businessIntelligenceService) GetFinancialSummary(ctx context.Context, req *models.FinancialSummaryRequest, user *models.User) (*models.FinancialSummary, error) {
	if req == nil {
		return nil, errors.New("financial summary request cannot be nil")
	}

	// Check permissions
	if !s.canViewReports(user) {
		return nil, errors.New("insufficient permissions to view financial summary")
	}

	// Validate date range
	if err := s.validateDateRange(req.StartDate, req.EndDate); err != nil {
		return nil, err
	}

	// Get financial data
	financialData, err := s.getFinancialData(ctx, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get financial data: %w", err)
	}

	return &models.FinancialSummary{
		Period:           models.DateRange{StartDate: req.StartDate, EndDate: req.EndDate},
		FinancialData:    *financialData,
		RevenueBreakdown: []models.RevenueBreakdownItem{}, // TODO: Implement
		ProfitMargins:    []models.ProfitMarginItem{},     // TODO: Implement
		GeneratedAt:      time.Now(),
	}, nil
}

// Helper methods

// canViewReports checks if user can view reports
func (s *businessIntelligenceService) canViewReports(user *models.User) bool {
	return user != nil && user.CanViewReports()
}

// validateDateRange validates a date range
func (s *businessIntelligenceService) validateDateRange(startDate, endDate time.Time) error {
	if startDate.IsZero() || endDate.IsZero() {
		return errors.New("start date and end date are required")
	}
	if startDate.After(endDate) {
		return errors.New("start date cannot be after end date")
	}
	if endDate.After(time.Now()) {
		return errors.New("end date cannot be in the future")
	}
	return nil
}

// convertProductsToResponses converts products to response format
func (s *businessIntelligenceService) convertProductsToResponses(products []models.Product) []models.ProductResponse {
	responses := make([]models.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = product.ToResponse()
	}
	return responses
}

// getProductStatistics gets basic product statistics
func (s *businessIntelligenceService) getProductStatistics(ctx context.Context) (*models.ProductStatistics, error) {
	// Get active products
	products, err := s.productRepo.GetActiveProducts(ctx)
	if err != nil {
		return nil, err
	}

	// Calculate basic statistics
	var totalValue float64
	var lowStockCount, outOfStockCount int64

	for _, product := range products {
		totalValue += product.GetTotalValue()
		if product.IsLowStock() {
			lowStockCount++
		}
		if product.IsOutOfStock() {
			outOfStockCount++
		}
	}

	var averageStockValue float64
	if len(products) > 0 {
		averageStockValue = totalValue / float64(len(products))
	}

	return &models.ProductStatistics{
		TotalProducts:     int64(len(products)),
		TotalValue:        totalValue,
		LowStockCount:     lowStockCount,
		OutOfStockCount:   outOfStockCount,
		AverageStockValue: averageStockValue,
	}, nil
}

// getSalesSummary gets basic sales summary
func (s *businessIntelligenceService) getSalesSummary(ctx context.Context) (*models.SalesSummary, error) {
	// TODO: Implement actual sales summary calculation
	// For now, return empty summary
	return &models.SalesSummary{
		TotalSales:        0.0,
		TotalQuantity:     0,
		AverageOrderValue: 0.0,
		SalesCount:        0,
	}, nil
}

// getRecentAlerts gets recent alerts
func (s *businessIntelligenceService) getRecentAlerts(ctx context.Context, limit int) ([]models.AlertResponse, error) {
	// TODO: Implement actual recent alerts retrieval
	// For now, return empty list
	return []models.AlertResponse{}, nil
}

// getUserActivitySummary gets user activity summary
func (s *businessIntelligenceService) getUserActivitySummary(ctx context.Context) (*models.UserActivitySummary, error) {
	// TODO: Implement actual user activity summary
	// For now, return empty summary
	return &models.UserActivitySummary{
		TotalUsers:         0,
		ActiveUsers:        0,
		TotalLogins:        0,
		AverageSessionTime: 0.0,
	}, nil
}

// getSalesMetrics gets sales metrics for a date range
func (s *businessIntelligenceService) getSalesMetrics(ctx context.Context, startDate, endDate time.Time) (*models.SalesMetrics, error) {
	// TODO: Implement actual sales metrics calculation
	// For now, return empty metrics
	return &models.SalesMetrics{
		TotalRevenue:      0.0,
		TotalQuantity:     0,
		AverageOrderValue: 0.0,
		SalesCount:        0,
	}, nil
}

// calculateInventoryMetrics calculates inventory metrics
func (s *businessIntelligenceService) calculateInventoryMetrics(products []models.Product) *models.InventoryMetrics {
	var totalValue float64
	var lowStockCount, outOfStockCount int64

	for _, product := range products {
		totalValue += product.GetTotalValue()
		if product.IsLowStock() {
			lowStockCount++
		}
		if product.IsOutOfStock() {
			outOfStockCount++
		}
	}

	var averageStockValue float64
	if len(products) > 0 {
		averageStockValue = totalValue / float64(len(products))
	}

	return &models.InventoryMetrics{
		TotalProducts:     int64(len(products)),
		TotalValue:        totalValue,
		AverageStockValue: averageStockValue,
		LowStockCount:     lowStockCount,
		OutOfStockCount:   outOfStockCount,
	}
}

// calculateStockStatusDistribution calculates stock status distribution
func (s *businessIntelligenceService) calculateStockStatusDistribution(products []models.Product) models.StockStatusDistribution {
	var available, lowStock, outOfStock int64

	for _, product := range products {
		switch product.GetStockStatus() {
		case models.StockStatusAvailable:
			available++
		case models.StockStatusLowStock:
			lowStock++
		case models.StockStatusOutOfStock:
			outOfStock++
		}
	}

	return models.StockStatusDistribution{
		Available:  available,
		LowStock:   lowStock,
		OutOfStock: outOfStock,
	}
}

// calculateLowStockMetrics calculates low stock metrics
func (s *businessIntelligenceService) calculateLowStockMetrics(lowStockProducts, outOfStockProducts []models.Product) *models.LowStockMetrics {
	var totalValueAtRisk float64
	var totalQuantity int

	// Calculate value at risk for low stock products
	for _, product := range lowStockProducts {
		totalValueAtRisk += product.GetTotalValue()
		totalQuantity += product.QuantityOnHand
	}

	// Calculate value at risk for out of stock products
	for _, product := range outOfStockProducts {
		totalValueAtRisk += product.GetTotalValue()
	}

	var averageStockLevel float64
	totalProducts := len(lowStockProducts) + len(outOfStockProducts)
	if totalProducts > 0 {
		averageStockLevel = float64(totalQuantity) / float64(totalProducts)
	}

	return &models.LowStockMetrics{
		LowStockCount:     int64(len(lowStockProducts)),
		OutOfStockCount:   int64(len(outOfStockProducts)),
		TotalValueAtRisk:  totalValueAtRisk,
		AverageStockLevel: averageStockLevel,
	}
}

// getFinancialData gets financial data for a date range
func (s *businessIntelligenceService) getFinancialData(ctx context.Context, startDate, endDate time.Time) (*models.FinancialData, error) {
	// TODO: Implement actual financial data calculation
	// For now, return empty financial data
	return &models.FinancialData{
		TotalRevenue:      0.0,
		TotalCost:         0.0,
		TotalProfit:       0.0,
		ProfitMargin:      0.0,
		AverageOrderValue: 0.0,
	}, nil
}
