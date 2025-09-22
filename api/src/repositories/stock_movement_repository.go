// Package repositories contains the repository layer implementations for the TT Stock Backend API.
// It provides data access interfaces and implementations using GORM for database operations.
package repositories

import (
	"context"
	"errors"
	"time"

	"tt-stock-api/src/models"

	"gorm.io/gorm"
)

// StockMovementRepository defines the interface for stock movement data operations
type StockMovementRepository interface {
	// Create creates a new stock movement
	Create(ctx context.Context, movement *models.StockMovement) error

	// GetByID retrieves a stock movement by ID
	GetByID(ctx context.Context, id uint) (*models.StockMovement, error)

	// Update updates an existing stock movement
	Update(ctx context.Context, movement *models.StockMovement) error

	// Delete deletes a stock movement
	Delete(ctx context.Context, id uint) error

	// List retrieves stock movements with pagination and filtering
	List(ctx context.Context, req *models.StockMovementListRequest) (*models.StockMovementListResponse, error)

	// GetMovementsByProduct retrieves stock movements for a specific product
	GetMovementsByProduct(ctx context.Context, productID uint) ([]models.StockMovement, error)

	// GetMovementsByUser retrieves stock movements created by a specific user
	GetMovementsByUser(ctx context.Context, userID uint) ([]models.StockMovement, error)

	// GetMovementsByType retrieves stock movements by type
	GetMovementsByType(ctx context.Context, movementType models.MovementType) ([]models.StockMovement, error)

	// GetMovementsByDateRange retrieves stock movements within a date range
	GetMovementsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]models.StockMovement, error)

	// GetRecentMovements retrieves recent stock movements
	GetRecentMovements(ctx context.Context, limit int) ([]models.StockMovement, error)

	// GetMovementSummary retrieves summary statistics for stock movements
	GetMovementSummary(ctx context.Context, productID *uint, startDate, endDate *time.Time) (*models.StockMovementSummary, error)

	// Count returns the total number of stock movements
	Count(ctx context.Context) (int64, error)

	// GetTotalQuantityByProduct calculates total quantity changes for a product
	GetTotalQuantityByProduct(ctx context.Context, productID uint) (int, error)
}

// stockMovementRepository implements the StockMovementRepository interface
type stockMovementRepository struct {
	db *gorm.DB
}

// NewStockMovementRepository creates a new StockMovementRepository instance
func NewStockMovementRepository(db *gorm.DB) StockMovementRepository {
	return &stockMovementRepository{
		db: db,
	}
}

// Create creates a new stock movement
func (r *stockMovementRepository) Create(ctx context.Context, movement *models.StockMovement) error {
	if movement == nil {
		return errors.New("stock movement cannot be nil")
	}

	// Create movement
	if err := r.db.WithContext(ctx).Create(movement).Error; err != nil {
		return err
	}

	return nil
}

// GetByID retrieves a stock movement by ID
func (r *stockMovementRepository) GetByID(ctx context.Context, id uint) (*models.StockMovement, error) {
	if id == 0 {
		return nil, errors.New("stock movement ID cannot be zero")
	}

	var movement models.StockMovement
	if err := r.db.WithContext(ctx).Preload("Product").Preload("User").First(&movement, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("stock movement not found")
		}
		return nil, err
	}

	return &movement, nil
}

// Update updates an existing stock movement
func (r *stockMovementRepository) Update(ctx context.Context, movement *models.StockMovement) error {
	if movement == nil {
		return errors.New("stock movement cannot be nil")
	}

	if movement.ID == 0 {
		return errors.New("stock movement ID cannot be zero")
	}

	// Check if movement exists
	_, err := r.GetByID(ctx, movement.ID)
	if err != nil {
		return err
	}

	// Update movement
	if err := r.db.WithContext(ctx).Save(movement).Error; err != nil {
		return err
	}

	return nil
}

// Delete deletes a stock movement
func (r *stockMovementRepository) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("stock movement ID cannot be zero")
	}

	// Check if movement exists
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete movement
	if err := r.db.WithContext(ctx).Delete(&models.StockMovement{}, id).Error; err != nil {
		return err
	}

	return nil
}

// List retrieves stock movements with pagination and filtering
func (r *stockMovementRepository) List(ctx context.Context, req *models.StockMovementListRequest) (*models.StockMovementListResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	// Set default values
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	query := r.db.WithContext(ctx).Model(&models.StockMovement{}).Preload("Product").Preload("User")

	// Apply filters
	if req.ProductID != nil {
		query = query.Where("product_id = ?", *req.ProductID)
	}
	if req.UserID != nil {
		query = query.Where("user_id = ?", *req.UserID)
	}
	if req.MovementType != nil {
		query = query.Where("movement_type = ?", *req.MovementType)
	}
	if req.StartDate != nil {
		query = query.Where("created_at >= ?", *req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("created_at <= ?", *req.EndDate)
	}

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Calculate pagination
	offset := (req.Page - 1) * req.Limit
	totalPages := (total + int64(req.Limit) - 1) / int64(req.Limit)

	// Apply pagination and ordering
	var movements []models.StockMovement
	if err := query.Order("created_at DESC").Offset(offset).Limit(req.Limit).Find(&movements).Error; err != nil {
		return nil, err
	}

	// Convert to response format
	movementResponses := make([]models.StockMovementResponse, len(movements))
	for i, movement := range movements {
		movementResponses[i] = movement.ToResponse()
	}

	return &models.StockMovementListResponse{
		Movements: movementResponses,
		Pagination: models.PaginationResponse{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: int(totalPages),
			HasNext:    req.Page < int(totalPages),
			HasPrev:    req.Page > 1,
		},
	}, nil
}

// GetMovementsByProduct retrieves stock movements for a specific product
func (r *stockMovementRepository) GetMovementsByProduct(ctx context.Context, productID uint) ([]models.StockMovement, error) {
	if productID == 0 {
		return nil, errors.New("product ID cannot be zero")
	}

	var movements []models.StockMovement
	if err := r.db.WithContext(ctx).Preload("Product").Preload("User").Where("product_id = ?", productID).Order("created_at DESC").Find(&movements).Error; err != nil {
		return nil, err
	}

	return movements, nil
}

// GetMovementsByUser retrieves stock movements created by a specific user
func (r *stockMovementRepository) GetMovementsByUser(ctx context.Context, userID uint) ([]models.StockMovement, error) {
	if userID == 0 {
		return nil, errors.New("user ID cannot be zero")
	}

	var movements []models.StockMovement
	if err := r.db.WithContext(ctx).Preload("Product").Preload("User").Where("user_id = ?", userID).Order("created_at DESC").Find(&movements).Error; err != nil {
		return nil, err
	}

	return movements, nil
}

// GetMovementsByType retrieves stock movements by type
func (r *stockMovementRepository) GetMovementsByType(ctx context.Context, movementType models.MovementType) ([]models.StockMovement, error) {
	var movements []models.StockMovement
	if err := r.db.WithContext(ctx).Preload("Product").Preload("User").Where("movement_type = ?", movementType).Order("created_at DESC").Find(&movements).Error; err != nil {
		return nil, err
	}

	return movements, nil
}

// GetMovementsByDateRange retrieves stock movements within a date range
func (r *stockMovementRepository) GetMovementsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]models.StockMovement, error) {
	var movements []models.StockMovement
	if err := r.db.WithContext(ctx).Preload("Product").Preload("User").Where("created_at BETWEEN ? AND ?", startDate, endDate).Order("created_at DESC").Find(&movements).Error; err != nil {
		return nil, err
	}

	return movements, nil
}

// GetRecentMovements retrieves recent stock movements
func (r *stockMovementRepository) GetRecentMovements(ctx context.Context, limit int) ([]models.StockMovement, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	var movements []models.StockMovement
	if err := r.db.WithContext(ctx).Preload("Product").Preload("User").Order("created_at DESC").Limit(limit).Find(&movements).Error; err != nil {
		return nil, err
	}

	return movements, nil
}

// GetMovementSummary retrieves summary statistics for stock movements
func (r *stockMovementRepository) GetMovementSummary(ctx context.Context, productID *uint, startDate, endDate *time.Time) (*models.StockMovementSummary, error) {
	query := r.db.WithContext(ctx).Model(&models.StockMovement{})

	// Apply filters
	if productID != nil {
		query = query.Where("product_id = ?", *productID)
	}
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	// Get summary data
	var summary models.StockMovementSummary
	if err := query.Select(`
		COUNT(*) as total_movements,
		SUM(CASE WHEN movement_type = 'Incoming' THEN 1 ELSE 0 END) as incoming_count,
		SUM(CASE WHEN movement_type = 'Outgoing' THEN 1 ELSE 0 END) as outgoing_count,
		SUM(CASE WHEN movement_type = 'Sale' THEN 1 ELSE 0 END) as sale_count,
		SUM(CASE WHEN movement_type = 'Adjustment' THEN 1 ELSE 0 END) as adjustment_count,
		SUM(CASE WHEN movement_type = 'Return' THEN 1 ELSE 0 END) as return_count,
		SUM(quantity) as net_change
	`).Scan(&summary).Error; err != nil {
		return nil, err
	}

	return &summary, nil
}

// Count returns the total number of stock movements
func (r *stockMovementRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.StockMovement{}).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// GetTotalQuantityByProduct calculates total quantity changes for a product
func (r *stockMovementRepository) GetTotalQuantityByProduct(ctx context.Context, productID uint) (int, error) {
	if productID == 0 {
		return 0, errors.New("product ID cannot be zero")
	}

	var total int
	if err := r.db.WithContext(ctx).Model(&models.StockMovement{}).Where("product_id = ?", productID).Select("COALESCE(SUM(quantity), 0)").Scan(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}
