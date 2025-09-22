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

// StockService defines the interface for stock management operations
type StockService interface {
	// CreateStockMovement creates a new stock movement
	CreateStockMovement(ctx context.Context, req *models.StockMovementCreateRequest, user *models.User) (*models.StockMovementResponse, error)

	// GetStockMovement retrieves a stock movement by ID
	GetStockMovement(ctx context.Context, id uint, user *models.User) (*models.StockMovementResponse, error)

	// UpdateStockMovement updates an existing stock movement
	UpdateStockMovement(ctx context.Context, id uint, req *models.StockMovementUpdateRequest, user *models.User) (*models.StockMovementResponse, error)

	// DeleteStockMovement deletes a stock movement
	DeleteStockMovement(ctx context.Context, id uint, user *models.User) error

	// ListStockMovements retrieves stock movements with pagination and filtering
	ListStockMovements(ctx context.Context, req *models.StockMovementListRequest, user *models.User) (*models.StockMovementListResponse, error)

	// GetMovementsByProduct retrieves stock movements for a specific product
	GetMovementsByProduct(ctx context.Context, productID uint, user *models.User) ([]models.StockMovementResponse, error)

	// GetMovementsByUser retrieves stock movements created by a specific user
	GetMovementsByUser(ctx context.Context, userID uint, requester *models.User) ([]models.StockMovementResponse, error)

	// ProcessSale processes a sale transaction
	ProcessSale(ctx context.Context, req *models.SaleRequest, user *models.User) (*models.SaleResponse, error)

	// ProcessIncomingStock processes incoming stock
	ProcessIncomingStock(ctx context.Context, req *models.StockMovementCreateRequest, user *models.User) (*models.StockMovementResponse, error)

	// ProcessStockAdjustment processes a stock adjustment
	ProcessStockAdjustment(ctx context.Context, req *models.StockMovementCreateRequest, user *models.User) (*models.StockMovementResponse, error)

	// ProcessReturn processes a product return
	ProcessReturn(ctx context.Context, req *models.StockMovementCreateRequest, user *models.User) (*models.StockMovementResponse, error)

	// GetStockMovementSummary retrieves summary statistics for stock movements
	GetStockMovementSummary(ctx context.Context, productID *uint, startDate, endDate *time.Time, user *models.User) (*models.StockMovementSummary, error)

	// GetRecentMovements retrieves recent stock movements
	GetRecentMovements(ctx context.Context, limit int, user *models.User) ([]models.StockMovementResponse, error)
}

// stockService implements the StockService interface
type stockService struct {
	stockMovementRepo repositories.StockMovementRepository
	productRepo       repositories.ProductRepository
	alertRepo         repositories.AlertRepository
}

// NewStockService creates a new StockService instance
func NewStockService(
	stockMovementRepo repositories.StockMovementRepository,
	productRepo repositories.ProductRepository,
	alertRepo repositories.AlertRepository,
) StockService {
	return &stockService{
		stockMovementRepo: stockMovementRepo,
		productRepo:       productRepo,
		alertRepo:         alertRepo,
	}
}

// CreateStockMovement creates a new stock movement
func (s *stockService) CreateStockMovement(ctx context.Context, req *models.StockMovementCreateRequest, user *models.User) (*models.StockMovementResponse, error) {
	if req == nil {
		return nil, errors.New("create request cannot be nil")
	}

	// Check permissions
	if !s.canManageInventory(user) {
		return nil, errors.New("insufficient permissions to create stock movements")
	}

	// Validate product exists
	product, err := s.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, errors.New("product not found")
	}

	// Create stock movement
	movement := &models.StockMovement{
		ProductID:    req.ProductID,
		UserID:       user.ID,
		MovementType: req.MovementType,
		Quantity:     req.Quantity,
		Reason:       req.Reason,
		Reference:    req.Reference,
		Notes:        req.Notes,
	}

	// Validate movement based on type
	if err := s.validateStockMovement(movement, product); err != nil {
		return nil, err
	}

	// Create movement
	if err := s.stockMovementRepo.Create(ctx, movement); err != nil {
		return nil, fmt.Errorf("failed to create stock movement: %w", err)
	}

	// Update product quantity
	if err := s.updateProductQuantity(ctx, product, movement); err != nil {
		return nil, fmt.Errorf("failed to update product quantity: %w", err)
	}

	// Check for low stock alerts
	if err := s.checkLowStockAlert(ctx, product); err != nil {
		// Log error but don't fail the operation
		// TODO: Add proper logging
	}

	response := movement.ToResponse()
	return &response, nil
}

// GetStockMovement retrieves a stock movement by ID
func (s *stockService) GetStockMovement(ctx context.Context, id uint, user *models.User) (*models.StockMovementResponse, error) {
	if id == 0 {
		return nil, errors.New("stock movement ID cannot be zero")
	}

	// Check permissions
	if !s.canViewInventory(user) {
		return nil, errors.New("insufficient permissions to view stock movements")
	}

	// Get stock movement
	movement, err := s.stockMovementRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := movement.ToResponse()
	return &response, nil
}

// UpdateStockMovement updates an existing stock movement
func (s *stockService) UpdateStockMovement(ctx context.Context, id uint, req *models.StockMovementUpdateRequest, user *models.User) (*models.StockMovementResponse, error) {
	if id == 0 {
		return nil, errors.New("stock movement ID cannot be zero")
	}

	if req == nil {
		return nil, errors.New("update request cannot be nil")
	}

	// Check permissions
	if !s.canManageInventory(user) {
		return nil, errors.New("insufficient permissions to update stock movements")
	}

	// Get existing movement
	movement, err := s.stockMovementRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store original quantity for rollback
	originalQuantity := movement.Quantity

	// Update fields
	if req.MovementType != nil {
		movement.MovementType = *req.MovementType
	}
	if req.Quantity != nil {
		movement.Quantity = *req.Quantity
	}
	if req.Reason != nil {
		movement.Reason = req.Reason
	}
	if req.Reference != nil {
		movement.Reference = req.Reference
	}
	if req.Notes != nil {
		movement.Notes = req.Notes
	}

	// Get product for validation
	product, err := s.productRepo.GetByID(ctx, movement.ProductID)
	if err != nil {
		return nil, err
	}

	// Validate updated movement
	if err := s.validateStockMovement(movement, product); err != nil {
		return nil, err
	}

	// Calculate quantity difference
	quantityDiff := movement.Quantity - originalQuantity

	// Update movement
	if err := s.stockMovementRepo.Update(ctx, movement); err != nil {
		return nil, fmt.Errorf("failed to update stock movement: %w", err)
	}

	// Update product quantity
	if quantityDiff != 0 {
		if err := s.productRepo.AdjustQuantity(ctx, product.ID, quantityDiff); err != nil {
			// Rollback movement update
			movement.Quantity = originalQuantity
			s.stockMovementRepo.Update(ctx, movement)
			return nil, fmt.Errorf("failed to update product quantity: %w", err)
		}
	}

	response := movement.ToResponse()
	return &response, nil
}

// DeleteStockMovement deletes a stock movement
func (s *stockService) DeleteStockMovement(ctx context.Context, id uint, user *models.User) error {
	if id == 0 {
		return errors.New("stock movement ID cannot be zero")
	}

	// Check permissions
	if !s.canManageInventory(user) {
		return errors.New("insufficient permissions to delete stock movements")
	}

	// Get movement to check if it exists
	movement, err := s.stockMovementRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Get product
	product, err := s.productRepo.GetByID(ctx, movement.ProductID)
	if err != nil {
		return err
	}

	// Reverse the quantity change
	reverseQuantity := -movement.Quantity
	if err := s.productRepo.AdjustQuantity(ctx, product.ID, reverseQuantity); err != nil {
		return fmt.Errorf("failed to reverse quantity change: %w", err)
	}

	// Delete movement
	return s.stockMovementRepo.Delete(ctx, id)
}

// ListStockMovements retrieves stock movements with pagination and filtering
func (s *stockService) ListStockMovements(ctx context.Context, req *models.StockMovementListRequest, user *models.User) (*models.StockMovementListResponse, error) {
	if req == nil {
		return nil, errors.New("list request cannot be nil")
	}

	// Check permissions
	if !s.canViewInventory(user) {
		return nil, errors.New("insufficient permissions to list stock movements")
	}

	// List stock movements
	return s.stockMovementRepo.List(ctx, req)
}

// GetMovementsByProduct retrieves stock movements for a specific product
func (s *stockService) GetMovementsByProduct(ctx context.Context, productID uint, user *models.User) ([]models.StockMovementResponse, error) {
	if productID == 0 {
		return nil, errors.New("product ID cannot be zero")
	}

	// Check permissions
	if !s.canViewInventory(user) {
		return nil, errors.New("insufficient permissions to view product movements")
	}

	// Get movements
	movements, err := s.stockMovementRepo.GetMovementsByProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]models.StockMovementResponse, len(movements))
	for i, movement := range movements {
		responses[i] = movement.ToResponse()
	}

	return responses, nil
}

// GetMovementsByUser retrieves stock movements created by a specific user
func (s *stockService) GetMovementsByUser(ctx context.Context, userID uint, requester *models.User) ([]models.StockMovementResponse, error) {
	if userID == 0 {
		return nil, errors.New("user ID cannot be zero")
	}

	// Check permissions
	if !s.canViewUserMovements(requester, userID) {
		return nil, errors.New("insufficient permissions to view user movements")
	}

	// Get movements
	movements, err := s.stockMovementRepo.GetMovementsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]models.StockMovementResponse, len(movements))
	for i, movement := range movements {
		responses[i] = movement.ToResponse()
	}

	return responses, nil
}

// ProcessSale processes a sale transaction
func (s *stockService) ProcessSale(ctx context.Context, req *models.SaleRequest, user *models.User) (*models.SaleResponse, error) {
	if req == nil {
		return nil, errors.New("sale request cannot be nil")
	}

	// Check permissions
	if !s.canManageInventory(user) {
		return nil, errors.New("insufficient permissions to process sales")
	}

	// Get product
	product, err := s.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, errors.New("product not found")
	}

	// Check if product can be sold
	if !product.CanSell(req.Quantity) {
		return nil, errors.New("insufficient stock for sale")
	}

	// Create sale movement
	movement := &models.StockMovement{
		ProductID:    req.ProductID,
		UserID:       user.ID,
		MovementType: models.MovementTypeSale,
		Quantity:     -req.Quantity, // Negative for outgoing
		Reason:       req.CustomerName,
		Reference:    req.Reference,
		Notes:        req.Notes,
	}

	// Create movement
	if err := s.stockMovementRepo.Create(ctx, movement); err != nil {
		return nil, fmt.Errorf("failed to create sale movement: %w", err)
	}

	// Update product quantity
	if err := s.productRepo.AdjustQuantity(ctx, product.ID, -req.Quantity); err != nil {
		return nil, fmt.Errorf("failed to update product quantity: %w", err)
	}

	// Get updated product
	updatedProduct, err := s.productRepo.GetByID(ctx, product.ID)
	if err != nil {
		return nil, err
	}

	// Check for low stock alert
	alertGenerated := false
	if err := s.checkLowStockAlert(ctx, updatedProduct); err == nil {
		alertGenerated = true
	}

	return &models.SaleResponse{
		Movement:       movement.ToResponse(),
		NewQuantity:    updatedProduct.QuantityOnHand,
		AlertGenerated: alertGenerated,
	}, nil
}

// ProcessIncomingStock processes incoming stock
func (s *stockService) ProcessIncomingStock(ctx context.Context, req *models.StockMovementCreateRequest, user *models.User) (*models.StockMovementResponse, error) {
	if req == nil {
		return nil, errors.New("incoming stock request cannot be nil")
	}

	// Set movement type
	req.MovementType = models.MovementTypeIncoming

	// Create movement
	return s.CreateStockMovement(ctx, req, user)
}

// ProcessStockAdjustment processes a stock adjustment
func (s *stockService) ProcessStockAdjustment(ctx context.Context, req *models.StockMovementCreateRequest, user *models.User) (*models.StockMovementResponse, error) {
	if req == nil {
		return nil, errors.New("adjustment request cannot be nil")
	}

	// Set movement type
	req.MovementType = models.MovementTypeAdjustment

	// Create movement
	return s.CreateStockMovement(ctx, req, user)
}

// ProcessReturn processes a product return
func (s *stockService) ProcessReturn(ctx context.Context, req *models.StockMovementCreateRequest, user *models.User) (*models.StockMovementResponse, error) {
	if req == nil {
		return nil, errors.New("return request cannot be nil")
	}

	// Set movement type
	req.MovementType = models.MovementTypeReturn

	// Create movement
	return s.CreateStockMovement(ctx, req, user)
}

// GetStockMovementSummary retrieves summary statistics for stock movements
func (s *stockService) GetStockMovementSummary(ctx context.Context, productID *uint, startDate, endDate *time.Time, user *models.User) (*models.StockMovementSummary, error) {
	// Check permissions
	if !s.canViewInventory(user) {
		return nil, errors.New("insufficient permissions to view movement summary")
	}

	// Get summary
	return s.stockMovementRepo.GetMovementSummary(ctx, productID, startDate, endDate)
}

// GetRecentMovements retrieves recent stock movements
func (s *stockService) GetRecentMovements(ctx context.Context, limit int, user *models.User) ([]models.StockMovementResponse, error) {
	// Check permissions
	if !s.canViewInventory(user) {
		return nil, errors.New("insufficient permissions to view recent movements")
	}

	// Get recent movements
	movements, err := s.stockMovementRepo.GetRecentMovements(ctx, limit)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]models.StockMovementResponse, len(movements))
	for i, movement := range movements {
		responses[i] = movement.ToResponse()
	}

	return responses, nil
}

// Helper methods

// canViewInventory checks if user can view inventory
func (s *stockService) canViewInventory(user *models.User) bool {
	return user != nil && user.CanManageInventory()
}

// canManageInventory checks if user can manage inventory
func (s *stockService) canManageInventory(user *models.User) bool {
	return user != nil && user.CanManageInventory()
}

// canViewUserMovements checks if user can view movements by another user
func (s *stockService) canViewUserMovements(requester *models.User, targetUserID uint) bool {
	if requester == nil {
		return false
	}
	// Users can view their own movements, admins and owners can view all
	return requester.ID == targetUserID || requester.IsAdmin() || requester.IsOwner()
}

// validateStockMovement validates a stock movement
func (s *stockService) validateStockMovement(movement *models.StockMovement, product *models.Product) error {
	// Validate movement type and quantity relationship
	switch movement.MovementType {
	case models.MovementTypeIncoming, models.MovementTypeReturn:
		if movement.Quantity <= 0 {
			return errors.New("incoming and return movements must have positive quantities")
		}
	case models.MovementTypeOutgoing, models.MovementTypeSale:
		if movement.Quantity >= 0 {
			return errors.New("outgoing and sale movements must have negative quantities")
		}
		// Check if there's enough stock
		if product.QuantityOnHand < -movement.Quantity {
			return errors.New("insufficient stock for this movement")
		}
	case models.MovementTypeAdjustment:
		// Adjustments can be positive or negative
		if movement.Quantity == 0 {
			return errors.New("adjustment quantity cannot be zero")
		}
		// Check if adjustment would result in negative stock
		if product.QuantityOnHand+movement.Quantity < 0 {
			return errors.New("adjustment would result in negative stock")
		}
	default:
		return errors.New("invalid movement type")
	}

	return nil
}

// updateProductQuantity updates product quantity based on movement
func (s *stockService) updateProductQuantity(ctx context.Context, product *models.Product, movement *models.StockMovement) error {
	return s.productRepo.AdjustQuantity(ctx, product.ID, movement.Quantity)
}

// checkLowStockAlert checks if a low stock alert should be created
func (s *stockService) checkLowStockAlert(ctx context.Context, product *models.Product) error {
	if product.IsLowStock() {
		// Create low stock alert
		alert := &models.Alert{
			ProductID: &product.ID,
			AlertType: models.AlertTypeLowStock,
			Priority:  models.AlertPriorityMedium,
			Title:     "Low Stock Alert",
			Message:   fmt.Sprintf("Product %s is running low on stock (%d remaining)", product.SKU, product.QuantityOnHand),
			IsRead:    false,
			IsActive:  true,
		}

		return s.alertRepo.Create(ctx, alert)
	}

	return nil
}
