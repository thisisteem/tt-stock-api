// Package models contains the domain models for the TT Stock Backend API.
package models

import (
	"time"
)

// DateRange represents a date range for reporting
type DateRange struct {
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

// DashboardData represents comprehensive dashboard data
type DashboardData struct {
	ProductStatistics    ProductStatistics    `json:"productStatistics"`
	SalesSummary         SalesSummary         `json:"salesSummary"`
	StockMovementSummary StockMovementSummary `json:"stockMovementSummary"`
	RecentAlerts         []AlertResponse      `json:"recentAlerts"`
	TopSellingProducts   []ProductSalesData   `json:"topSellingProducts"`
	LowStockProducts     []ProductResponse    `json:"lowStockProducts"`
	UserActivity         UserActivitySummary  `json:"userActivity"`
	GeneratedAt          time.Time            `json:"generatedAt"`
}

// SalesSummary represents sales summary data
type SalesSummary struct {
	TotalSales        float64 `json:"totalSales"`
	TotalQuantity     int     `json:"totalQuantity"`
	AverageOrderValue float64 `json:"averageOrderValue"`
	SalesCount        int64   `json:"salesCount"`
}

// ProductSalesData represents product sales data
type ProductSalesData struct {
	Product      ProductResponse `json:"product"`
	QuantitySold int             `json:"quantitySold"`
	TotalRevenue float64         `json:"totalRevenue"`
	SalesCount   int64           `json:"salesCount"`
}

// UserActivitySummary represents user activity summary
type UserActivitySummary struct {
	TotalUsers         int64   `json:"totalUsers"`
	ActiveUsers        int64   `json:"activeUsers"`
	TotalLogins        int64   `json:"totalLogins"`
	AverageSessionTime float64 `json:"averageSessionTime"`
}

// SalesReportRequest represents the request payload for sales report
type SalesReportRequest struct {
	StartDate time.Time `json:"startDate" binding:"required"`
	EndDate   time.Time `json:"endDate" binding:"required"`
	ProductID *uint     `json:"productId,omitempty"`
	UserID    *uint     `json:"userId,omitempty"`
}

// SalesReport represents a sales report
type SalesReport struct {
	Period         DateRange          `json:"period"`
	SalesMetrics   SalesMetrics       `json:"salesMetrics"`
	SalesByProduct []ProductSalesData `json:"salesByProduct"`
	SalesByUser    []UserSalesData    `json:"salesByUser"`
	GeneratedAt    time.Time          `json:"generatedAt"`
}

// SalesMetrics represents sales metrics
type SalesMetrics struct {
	TotalRevenue      float64          `json:"totalRevenue"`
	TotalQuantity     int              `json:"totalQuantity"`
	AverageOrderValue float64          `json:"averageOrderValue"`
	SalesCount        int64            `json:"salesCount"`
	TopSellingProduct *ProductResponse `json:"topSellingProduct,omitempty"`
}

// UserSalesData represents user sales data
type UserSalesData struct {
	User         UserResponse `json:"user"`
	QuantitySold int          `json:"quantitySold"`
	TotalRevenue float64      `json:"totalRevenue"`
	SalesCount   int64        `json:"salesCount"`
}

// InventoryReportRequest represents the request payload for inventory report
type InventoryReportRequest struct {
	ProductType *ProductType `json:"productType,omitempty"`
	Brand       *string      `json:"brand,omitempty"`
	StockStatus *StockStatus `json:"stockStatus,omitempty"`
}

// InventoryReport represents an inventory report
type InventoryReport struct {
	InventoryMetrics        InventoryMetrics        `json:"inventoryMetrics"`
	InventoryByCategory     []CategoryInventoryData `json:"inventoryByCategory"`
	InventoryByBrand        []BrandInventoryData    `json:"inventoryByBrand"`
	StockStatusDistribution StockStatusDistribution `json:"stockStatusDistribution"`
	GeneratedAt             time.Time               `json:"generatedAt"`
}

// InventoryMetrics represents inventory metrics
type InventoryMetrics struct {
	TotalProducts     int64   `json:"totalProducts"`
	TotalValue        float64 `json:"totalValue"`
	AverageStockValue float64 `json:"averageStockValue"`
	LowStockCount     int64   `json:"lowStockCount"`
	OutOfStockCount   int64   `json:"outOfStockCount"`
}

// CategoryInventoryData represents inventory data by category
type CategoryInventoryData struct {
	Category      ProductType `json:"category"`
	ProductCount  int64       `json:"productCount"`
	TotalValue    float64     `json:"totalValue"`
	TotalQuantity int         `json:"totalQuantity"`
}

// BrandInventoryData represents inventory data by brand
type BrandInventoryData struct {
	Brand         string  `json:"brand"`
	ProductCount  int64   `json:"productCount"`
	TotalValue    float64 `json:"totalValue"`
	TotalQuantity int     `json:"totalQuantity"`
}

// StockStatusDistribution represents stock status distribution
type StockStatusDistribution struct {
	Available  int64 `json:"available"`
	LowStock   int64 `json:"lowStock"`
	OutOfStock int64 `json:"outOfStock"`
}

// StockMovementReportRequest represents the request payload for stock movement report
type StockMovementReportRequest struct {
	StartDate    time.Time     `json:"startDate" binding:"required"`
	EndDate      time.Time     `json:"endDate" binding:"required"`
	ProductID    *uint         `json:"productId,omitempty"`
	UserID       *uint         `json:"userId,omitempty"`
	MovementType *MovementType `json:"movementType,omitempty"`
}

// StockMovementReport represents a stock movement report
type StockMovementReport struct {
	Period             DateRange             `json:"period"`
	MovementMetrics    MovementMetrics       `json:"movementMetrics"`
	MovementsByType    []MovementTypeData    `json:"movementsByType"`
	MovementsByProduct []ProductMovementData `json:"movementsByProduct"`
	GeneratedAt        time.Time             `json:"generatedAt"`
}

// MovementMetrics represents movement metrics
type MovementMetrics struct {
	TotalMovements  int64 `json:"totalMovements"`
	IncomingCount   int64 `json:"incomingCount"`
	OutgoingCount   int64 `json:"outgoingCount"`
	SaleCount       int64 `json:"saleCount"`
	AdjustmentCount int64 `json:"adjustmentCount"`
	ReturnCount     int64 `json:"returnCount"`
	NetChange       int   `json:"netChange"`
}

// MovementTypeData represents movement data by type
type MovementTypeData struct {
	MovementType  MovementType `json:"movementType"`
	Count         int64        `json:"count"`
	TotalQuantity int          `json:"totalQuantity"`
}

// ProductMovementData represents movement data by product
type ProductMovementData struct {
	Product       ProductResponse `json:"product"`
	MovementCount int64           `json:"movementCount"`
	NetChange     int             `json:"netChange"`
	IncomingQty   int             `json:"incomingQty"`
	OutgoingQty   int             `json:"outgoingQty"`
}

// ProductPerformanceRequest represents the request payload for product performance report
type ProductPerformanceRequest struct {
	StartDate time.Time `json:"startDate" binding:"required"`
	EndDate   time.Time `json:"endDate" binding:"required"`
	Limit     int       `json:"limit" binding:"min=1,max=100"`
}

// ProductPerformanceReport represents a product performance report
type ProductPerformanceReport struct {
	Period             DateRange                `json:"period"`
	ProductPerformance ProductPerformanceData   `json:"productPerformance"`
	TopPerformers      []ProductPerformanceItem `json:"topPerformers"`
	Underperformers    []ProductPerformanceItem `json:"underperformers"`
	GeneratedAt        time.Time                `json:"generatedAt"`
}

// ProductPerformanceData represents product performance data
type ProductPerformanceData struct {
	TotalProducts       int64   `json:"totalProducts"`
	ActiveProducts      int64   `json:"activeProducts"`
	AverageSales        float64 `json:"averageSales"`
	AverageProfitMargin float64 `json:"averageProfitMargin"`
}

// ProductPerformanceItem represents a product performance item
type ProductPerformanceItem struct {
	Product          ProductResponse `json:"product"`
	SalesCount       int64           `json:"salesCount"`
	QuantitySold     int             `json:"quantitySold"`
	TotalRevenue     float64         `json:"totalRevenue"`
	ProfitMargin     float64         `json:"profitMargin"`
	PerformanceScore float64         `json:"performanceScore"`
}

// UserActivityRequest represents the request payload for user activity report
type UserActivityRequest struct {
	StartDate time.Time `json:"startDate" binding:"required"`
	EndDate   time.Time `json:"endDate" binding:"required"`
	Limit     int       `json:"limit" binding:"min=1,max=100"`
}

// UserActivityReport represents a user activity report
type UserActivityReport struct {
	Period          DateRange          `json:"period"`
	UserActivity    UserActivityData   `json:"userActivity"`
	MostActiveUsers []UserActivityItem `json:"mostActiveUsers"`
	LoginStatistics LoginStatistics    `json:"loginStatistics"`
	GeneratedAt     time.Time          `json:"generatedAt"`
}

// UserActivityData represents user activity data
type UserActivityData struct {
	TotalUsers         int64   `json:"totalUsers"`
	ActiveUsers        int64   `json:"activeUsers"`
	TotalLogins        int64   `json:"totalLogins"`
	AverageSessionTime float64 `json:"averageSessionTime"`
	TotalMovements     int64   `json:"totalMovements"`
}

// UserActivityItem represents a user activity item
type UserActivityItem struct {
	User          UserResponse `json:"user"`
	LoginCount    int64        `json:"loginCount"`
	MovementCount int64        `json:"movementCount"`
	LastLoginAt   *time.Time   `json:"lastLoginAt,omitempty"`
	ActivityScore float64      `json:"activityScore"`
}

// LoginStatistics represents login statistics
type LoginStatistics struct {
	TotalLogins          int64   `json:"totalLogins"`
	UniqueUsers          int64   `json:"uniqueUsers"`
	AverageLoginsPerUser float64 `json:"averageLoginsPerUser"`
	MostActiveDay        string  `json:"mostActiveDay"`
}

// FinancialSummaryRequest represents the request payload for financial summary
type FinancialSummaryRequest struct {
	StartDate time.Time `json:"startDate" binding:"required"`
	EndDate   time.Time `json:"endDate" binding:"required"`
}

// FinancialSummary represents a financial summary
type FinancialSummary struct {
	Period           DateRange              `json:"period"`
	FinancialData    FinancialData          `json:"financialData"`
	RevenueBreakdown []RevenueBreakdownItem `json:"revenueBreakdown"`
	ProfitMargins    []ProfitMarginItem     `json:"profitMargins"`
	GeneratedAt      time.Time              `json:"generatedAt"`
}

// FinancialData represents financial data
type FinancialData struct {
	TotalRevenue      float64 `json:"totalRevenue"`
	TotalCost         float64 `json:"totalCost"`
	TotalProfit       float64 `json:"totalProfit"`
	ProfitMargin      float64 `json:"profitMargin"`
	AverageOrderValue float64 `json:"averageOrderValue"`
}

// RevenueBreakdownItem represents revenue breakdown item
type RevenueBreakdownItem struct {
	Category     string  `json:"category"`
	Revenue      float64 `json:"revenue"`
	Percentage   float64 `json:"percentage"`
	ProductCount int64   `json:"productCount"`
}

// ProfitMarginItem represents profit margin item
type ProfitMarginItem struct {
	Product      ProductResponse `json:"product"`
	CostPrice    float64         `json:"costPrice"`
	SellingPrice float64         `json:"sellingPrice"`
	ProfitMargin float64         `json:"profitMargin"`
	QuantitySold int             `json:"quantitySold"`
}

// LowStockReport represents a low stock report
type LowStockReport struct {
	LowStockMetrics    LowStockMetrics        `json:"lowStockMetrics"`
	LowStockProducts   []ProductResponse      `json:"lowStockProducts"`
	OutOfStockProducts []ProductResponse      `json:"outOfStockProducts"`
	LowStockByCategory []CategoryLowStockData `json:"lowStockByCategory"`
	LowStockByBrand    []BrandLowStockData    `json:"lowStockByBrand"`
	GeneratedAt        time.Time              `json:"generatedAt"`
}

// LowStockMetrics represents low stock metrics
type LowStockMetrics struct {
	LowStockCount     int64   `json:"lowStockCount"`
	OutOfStockCount   int64   `json:"outOfStockCount"`
	TotalValueAtRisk  float64 `json:"totalValueAtRisk"`
	AverageStockLevel float64 `json:"averageStockLevel"`
}

// CategoryLowStockData represents low stock data by category
type CategoryLowStockData struct {
	Category        ProductType `json:"category"`
	LowStockCount   int64       `json:"lowStockCount"`
	OutOfStockCount int64       `json:"outOfStockCount"`
	ValueAtRisk     float64     `json:"valueAtRisk"`
}

// BrandLowStockData represents low stock data by brand
type BrandLowStockData struct {
	Brand           string  `json:"brand"`
	LowStockCount   int64   `json:"lowStockCount"`
	OutOfStockCount int64   `json:"outOfStockCount"`
	ValueAtRisk     float64 `json:"valueAtRisk"`
}

// TopProductsRequest represents the request payload for top products report
type TopProductsRequest struct {
	StartDate time.Time `json:"startDate" binding:"required"`
	EndDate   time.Time `json:"endDate" binding:"required"`
	Limit     int       `json:"limit" binding:"min=1,max=100"`
}

// TopProductsReport represents a top products report
type TopProductsReport struct {
	Period       DateRange          `json:"period"`
	TopProducts  []ProductSalesData `json:"topProducts"`
	SalesMetrics SalesMetrics       `json:"salesMetrics"`
	GeneratedAt  time.Time          `json:"generatedAt"`
}

// SlowMovingRequest represents the request payload for slow moving products report
type SlowMovingRequest struct {
	StartDate time.Time `json:"startDate" binding:"required"`
	EndDate   time.Time `json:"endDate" binding:"required"`
	Limit     int       `json:"limit" binding:"min=1,max=100"`
}

// SlowMovingReport represents a slow moving products report
type SlowMovingReport struct {
	Period             DateRange               `json:"period"`
	SlowMovingProducts []SlowMovingProductItem `json:"slowMovingProducts"`
	SlowMovingMetrics  SlowMovingMetrics       `json:"slowMovingMetrics"`
	GeneratedAt        time.Time               `json:"generatedAt"`
}

// SlowMovingProductItem represents a slow moving product item
type SlowMovingProductItem struct {
	Product           ProductResponse `json:"product"`
	DaysSinceLastSale int             `json:"daysSinceLastSale"`
	QuantitySold      int             `json:"quantitySold"`
	StockTurnover     float64         `json:"stockTurnover"`
	ValueAtRisk       float64         `json:"valueAtRisk"`
}

// SlowMovingMetrics represents slow moving metrics
type SlowMovingMetrics struct {
	TotalSlowMovingProducts int64   `json:"totalSlowMovingProducts"`
	TotalValueAtRisk        float64 `json:"totalValueAtRisk"`
	AverageDaysSinceSale    float64 `json:"averageDaysSinceSale"`
}

// ProfitabilityRequest represents the request payload for profitability analysis
type ProfitabilityRequest struct {
	StartDate time.Time `json:"startDate" binding:"required"`
	EndDate   time.Time `json:"endDate" binding:"required"`
	Limit     int       `json:"limit" binding:"min=1,max=100"`
}

// ProfitabilityReport represents a profitability report
type ProfitabilityReport struct {
	Period            DateRange           `json:"period"`
	ProfitabilityData ProfitabilityData   `json:"profitabilityData"`
	MostProfitable    []ProfitabilityItem `json:"mostProfitable"`
	LeastProfitable   []ProfitabilityItem `json:"leastProfitable"`
	GeneratedAt       time.Time           `json:"generatedAt"`
}

// ProfitabilityData represents profitability data
type ProfitabilityData struct {
	TotalProfit   float64 `json:"totalProfit"`
	TotalRevenue  float64 `json:"totalRevenue"`
	TotalCost     float64 `json:"totalCost"`
	OverallMargin float64 `json:"overallMargin"`
	AverageMargin float64 `json:"averageMargin"`
}

// ProfitabilityItem represents a profitability item
type ProfitabilityItem struct {
	Product      ProductResponse `json:"product"`
	Revenue      float64         `json:"revenue"`
	Cost         float64         `json:"cost"`
	Profit       float64         `json:"profit"`
	ProfitMargin float64         `json:"profitMargin"`
	QuantitySold int             `json:"quantitySold"`
}

// TrendAnalysisRequest represents the request payload for trend analysis
type TrendAnalysisRequest struct {
	StartDate   time.Time `json:"startDate" binding:"required"`
	EndDate     time.Time `json:"endDate" binding:"required"`
	Granularity string    `json:"granularity" binding:"required,oneof=daily weekly monthly"`
}

// TrendAnalysisReport represents a trend analysis report
type TrendAnalysisReport struct {
	Period          DateRange        `json:"period"`
	Granularity     string           `json:"granularity"`
	TrendData       TrendData        `json:"trendData"`
	SalesTrends     []TrendDataPoint `json:"salesTrends"`
	InventoryTrends []TrendDataPoint `json:"inventoryTrends"`
	GeneratedAt     time.Time        `json:"generatedAt"`
}

// TrendData represents trend data
type TrendData struct {
	TotalDataPoints int     `json:"totalDataPoints"`
	TrendDirection  string  `json:"trendDirection"`
	GrowthRate      float64 `json:"growthRate"`
	Volatility      float64 `json:"volatility"`
}

// TrendDataPoint represents a trend data point
type TrendDataPoint struct {
	Date          time.Time `json:"date"`
	Value         float64   `json:"value"`
	Change        float64   `json:"change"`
	ChangePercent float64   `json:"changePercent"`
}
