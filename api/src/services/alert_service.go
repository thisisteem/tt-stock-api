// Package services contains the business logic layer implementations for the TT Stock Backend API.
// It provides service interfaces and implementations that orchestrate repository operations
// and implement business rules and validation logic.
package services

import (
	"context"
	"errors"
	"fmt"

	"tt-stock-api/src/models"
	"tt-stock-api/src/repositories"
)

// AlertService defines the interface for alert management operations
type AlertService interface {
	// CreateAlert creates a new alert
	CreateAlert(ctx context.Context, req *models.AlertCreateRequest, user *models.User) (*models.AlertResponse, error)

	// GetAlert retrieves an alert by ID
	GetAlert(ctx context.Context, id uint, user *models.User) (*models.AlertResponse, error)

	// UpdateAlert updates an existing alert
	UpdateAlert(ctx context.Context, id uint, req *models.AlertUpdateRequest, user *models.User) (*models.AlertResponse, error)

	// DeleteAlert deletes an alert
	DeleteAlert(ctx context.Context, id uint, user *models.User) error

	// ListAlerts retrieves alerts with pagination and filtering
	ListAlerts(ctx context.Context, req *models.AlertListRequest, user *models.User) (*models.AlertListResponse, error)

	// GetAlertsByProduct retrieves alerts for a specific product
	GetAlertsByProduct(ctx context.Context, productID uint, user *models.User) ([]models.AlertResponse, error)

	// GetUnreadAlerts retrieves unread alerts for a user
	GetUnreadAlerts(ctx context.Context, user *models.User) ([]models.AlertResponse, error)

	// MarkAlertAsRead marks an alert as read
	MarkAlertAsRead(ctx context.Context, id uint, user *models.User) error

	// MarkAllAlertsAsRead marks all alerts as read for a user
	MarkAllAlertsAsRead(ctx context.Context, user *models.User) error

	// GetUnreadCount returns the count of unread alerts
	GetUnreadCount(ctx context.Context, user *models.User) (int64, error)

	// CreateLowStockAlert creates a low stock alert for a product
	CreateLowStockAlert(ctx context.Context, product *models.Product) error

	// CreateOutOfStockAlert creates an out of stock alert for a product
	CreateOutOfStockAlert(ctx context.Context, product *models.Product) error

	// CreateSystemAlert creates a system-wide alert
	CreateSystemAlert(ctx context.Context, title, message string, priority models.AlertPriority) error

	// ProcessLowStockAlerts processes low stock alerts for all products
	ProcessLowStockAlerts(ctx context.Context) error

	// DeactivateAlert deactivates an alert
	DeactivateAlert(ctx context.Context, id uint, user *models.User) error
}

// alertService implements the AlertService interface
type alertService struct {
	alertRepo   repositories.AlertRepository
	productRepo repositories.ProductRepository
}

// NewAlertService creates a new AlertService instance
func NewAlertService(
	alertRepo repositories.AlertRepository,
	productRepo repositories.ProductRepository,
) AlertService {
	return &alertService{
		alertRepo:   alertRepo,
		productRepo: productRepo,
	}
}

// CreateAlert creates a new alert
func (s *alertService) CreateAlert(ctx context.Context, req *models.AlertCreateRequest, user *models.User) (*models.AlertResponse, error) {
	if req == nil {
		return nil, errors.New("create request cannot be nil")
	}

	// Check permissions
	if !s.canManageAlerts(user) {
		return nil, errors.New("insufficient permissions to create alerts")
	}

	// Validate product if specified
	if req.ProductID != nil {
		_, err := s.productRepo.GetByID(ctx, *req.ProductID)
		if err != nil {
			return nil, errors.New("product not found")
		}
	}

	// Create alert
	alert := &models.Alert{
		ProductID: req.ProductID,
		UserID:    req.UserID,
		AlertType: req.AlertType,
		Priority:  req.Priority,
		Title:     req.Title,
		Message:   req.Message,
		IsRead:    false,
		IsActive:  true,
	}

	// Set default priority if not provided
	if alert.Priority == "" {
		alert.Priority = models.AlertPriorityMedium
	}

	// Create alert
	if err := s.alertRepo.Create(ctx, alert); err != nil {
		return nil, fmt.Errorf("failed to create alert: %w", err)
	}

	response := alert.ToResponse()
	return &response, nil
}

// GetAlert retrieves an alert by ID
func (s *alertService) GetAlert(ctx context.Context, id uint, user *models.User) (*models.AlertResponse, error) {
	if id == 0 {
		return nil, errors.New("alert ID cannot be zero")
	}

	// Check permissions
	if !s.canViewAlerts(user) {
		return nil, errors.New("insufficient permissions to view alerts")
	}

	// Get alert
	alert, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user can view this specific alert
	if !s.canViewSpecificAlert(user, alert) {
		return nil, errors.New("insufficient permissions to view this alert")
	}

	response := alert.ToResponse()
	return &response, nil
}

// UpdateAlert updates an existing alert
func (s *alertService) UpdateAlert(ctx context.Context, id uint, req *models.AlertUpdateRequest, user *models.User) (*models.AlertResponse, error) {
	if id == 0 {
		return nil, errors.New("alert ID cannot be zero")
	}

	if req == nil {
		return nil, errors.New("update request cannot be nil")
	}

	// Check permissions
	if !s.canManageAlerts(user) {
		return nil, errors.New("insufficient permissions to update alerts")
	}

	// Get existing alert
	alert, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user can manage this specific alert
	if !s.canManageSpecificAlert(user, alert) {
		return nil, errors.New("insufficient permissions to update this alert")
	}

	// Update fields
	if req.Priority != nil {
		alert.Priority = *req.Priority
	}
	if req.Title != nil {
		alert.Title = *req.Title
	}
	if req.Message != nil {
		alert.Message = *req.Message
	}
	if req.IsRead != nil {
		if *req.IsRead {
			alert.MarkAsRead()
		} else {
			alert.MarkAsUnread()
		}
	}
	if req.IsActive != nil {
		if *req.IsActive {
			alert.Activate()
		} else {
			alert.Deactivate()
		}
	}

	// Update alert
	if err := s.alertRepo.Update(ctx, alert); err != nil {
		return nil, fmt.Errorf("failed to update alert: %w", err)
	}

	response := alert.ToResponse()
	return &response, nil
}

// DeleteAlert deletes an alert
func (s *alertService) DeleteAlert(ctx context.Context, id uint, user *models.User) error {
	if id == 0 {
		return errors.New("alert ID cannot be zero")
	}

	// Check permissions
	if !s.canManageAlerts(user) {
		return errors.New("insufficient permissions to delete alerts")
	}

	// Get alert to check permissions
	alert, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if user can manage this specific alert
	if !s.canManageSpecificAlert(user, alert) {
		return errors.New("insufficient permissions to delete this alert")
	}

	// Delete alert
	return s.alertRepo.Delete(ctx, id)
}

// ListAlerts retrieves alerts with pagination and filtering
func (s *alertService) ListAlerts(ctx context.Context, req *models.AlertListRequest, user *models.User) (*models.AlertListResponse, error) {
	if req == nil {
		return nil, errors.New("list request cannot be nil")
	}

	// Check permissions
	if !s.canViewAlerts(user) {
		return nil, errors.New("insufficient permissions to list alerts")
	}

	// Apply user-based filtering
	if err := s.applyUserFiltering(req, user); err != nil {
		return nil, err
	}

	// List alerts
	return s.alertRepo.List(ctx, req)
}

// GetAlertsByProduct retrieves alerts for a specific product
func (s *alertService) GetAlertsByProduct(ctx context.Context, productID uint, user *models.User) ([]models.AlertResponse, error) {
	if productID == 0 {
		return nil, errors.New("product ID cannot be zero")
	}

	// Check permissions
	if !s.canViewAlerts(user) {
		return nil, errors.New("insufficient permissions to view product alerts")
	}

	// Get alerts
	alerts, err := s.alertRepo.GetAlertsByProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]models.AlertResponse, len(alerts))
	for i, alert := range alerts {
		responses[i] = alert.ToResponse()
	}

	return responses, nil
}

// GetUnreadAlerts retrieves unread alerts for a user
func (s *alertService) GetUnreadAlerts(ctx context.Context, user *models.User) ([]models.AlertResponse, error) {
	// Check permissions
	if !s.canViewAlerts(user) {
		return nil, errors.New("insufficient permissions to view unread alerts")
	}

	// Get unread alerts
	alerts, err := s.alertRepo.GetUnreadAlerts(ctx)
	if err != nil {
		return nil, err
	}

	// Filter alerts based on user permissions
	var filteredAlerts []models.Alert
	for _, alert := range alerts {
		if s.canViewSpecificAlert(user, &alert) {
			filteredAlerts = append(filteredAlerts, alert)
		}
	}

	// Convert to response format
	responses := make([]models.AlertResponse, len(filteredAlerts))
	for i, alert := range filteredAlerts {
		responses[i] = alert.ToResponse()
	}

	return responses, nil
}

// MarkAlertAsRead marks an alert as read
func (s *alertService) MarkAlertAsRead(ctx context.Context, id uint, user *models.User) error {
	if id == 0 {
		return errors.New("alert ID cannot be zero")
	}

	// Check permissions
	if !s.canViewAlerts(user) {
		return errors.New("insufficient permissions to mark alerts as read")
	}

	// Get alert to check permissions
	alert, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if user can view this specific alert
	if !s.canViewSpecificAlert(user, alert) {
		return errors.New("insufficient permissions to mark this alert as read")
	}

	// Mark as read
	return s.alertRepo.MarkAsRead(ctx, id)
}

// MarkAllAlertsAsRead marks all alerts as read for a user
func (s *alertService) MarkAllAlertsAsRead(ctx context.Context, user *models.User) error {
	// Check permissions
	if !s.canViewAlerts(user) {
		return errors.New("insufficient permissions to mark alerts as read")
	}

	// Mark all user alerts as read
	return s.alertRepo.MarkAllAsRead(ctx, user.ID)
}

// GetUnreadCount returns the count of unread alerts
func (s *alertService) GetUnreadCount(ctx context.Context, user *models.User) (int64, error) {
	// Check permissions
	if !s.canViewAlerts(user) {
		return 0, errors.New("insufficient permissions to view unread count")
	}

	// Get unread count for user
	return s.alertRepo.GetUnreadCount(ctx, &user.ID)
}

// CreateLowStockAlert creates a low stock alert for a product
func (s *alertService) CreateLowStockAlert(ctx context.Context, product *models.Product) error {
	if product == nil {
		return errors.New("product cannot be nil")
	}

	// Check if alert already exists
	existingAlerts, err := s.alertRepo.GetAlertsByProduct(ctx, product.ID)
	if err != nil {
		return err
	}

	// Check if there's already an active low stock alert
	for _, alert := range existingAlerts {
		if alert.AlertType == models.AlertTypeLowStock && alert.IsActive {
			return nil // Alert already exists
		}
	}

	// Create low stock alert
	alert := &models.Alert{
		ProductID: &product.ID,
		AlertType: models.AlertTypeLowStock,
		Priority:  models.AlertPriorityMedium,
		Title:     "Low Stock Alert",
		Message:   fmt.Sprintf("Product %s (%s) is running low on stock (%d remaining)", product.SKU, product.Model, product.QuantityOnHand),
		IsRead:    false,
		IsActive:  true,
	}

	return s.alertRepo.Create(ctx, alert)
}

// CreateOutOfStockAlert creates an out of stock alert for a product
func (s *alertService) CreateOutOfStockAlert(ctx context.Context, product *models.Product) error {
	if product == nil {
		return errors.New("product cannot be nil")
	}

	// Check if alert already exists
	existingAlerts, err := s.alertRepo.GetAlertsByProduct(ctx, product.ID)
	if err != nil {
		return err
	}

	// Check if there's already an active out of stock alert
	for _, alert := range existingAlerts {
		if alert.AlertType == models.AlertTypeOutOfStock && alert.IsActive {
			return nil // Alert already exists
		}
	}

	// Create out of stock alert
	alert := &models.Alert{
		ProductID: &product.ID,
		AlertType: models.AlertTypeOutOfStock,
		Priority:  models.AlertPriorityHigh,
		Title:     "Out of Stock Alert",
		Message:   fmt.Sprintf("Product %s (%s) is out of stock", product.SKU, product.Model),
		IsRead:    false,
		IsActive:  true,
	}

	return s.alertRepo.Create(ctx, alert)
}

// CreateSystemAlert creates a system-wide alert
func (s *alertService) CreateSystemAlert(ctx context.Context, title, message string, priority models.AlertPriority) error {
	if title == "" {
		return errors.New("title cannot be empty")
	}
	if message == "" {
		return errors.New("message cannot be empty")
	}

	// Create system alert
	alert := &models.Alert{
		ProductID: nil, // System-wide alert
		UserID:    nil, // System-wide alert
		AlertType: models.AlertTypeSystem,
		Priority:  priority,
		Title:     title,
		Message:   message,
		IsRead:    false,
		IsActive:  true,
	}

	return s.alertRepo.Create(ctx, alert)
}

// ProcessLowStockAlerts processes low stock alerts for all products
func (s *alertService) ProcessLowStockAlerts(ctx context.Context) error {
	// Get all active products
	products, err := s.productRepo.GetActiveProducts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active products: %w", err)
	}

	// Process each product
	for _, product := range products {
		if product.IsOutOfStock() {
			// Create out of stock alert
			if err := s.CreateOutOfStockAlert(ctx, &product); err != nil {
				// Log error but continue processing
				// TODO: Add proper logging
			}
		} else if product.IsLowStock() {
			// Create low stock alert
			if err := s.CreateLowStockAlert(ctx, &product); err != nil {
				// Log error but continue processing
				// TODO: Add proper logging
			}
		}
	}

	return nil
}

// DeactivateAlert deactivates an alert
func (s *alertService) DeactivateAlert(ctx context.Context, id uint, user *models.User) error {
	if id == 0 {
		return errors.New("alert ID cannot be zero")
	}

	// Check permissions
	if !s.canManageAlerts(user) {
		return errors.New("insufficient permissions to deactivate alerts")
	}

	// Get alert to check permissions
	alert, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if user can manage this specific alert
	if !s.canManageSpecificAlert(user, alert) {
		return errors.New("insufficient permissions to deactivate this alert")
	}

	// Deactivate alert
	return s.alertRepo.DeactivateAlert(ctx, id)
}

// Helper methods

// canViewAlerts checks if user can view alerts
func (s *alertService) canViewAlerts(user *models.User) bool {
	return user != nil && user.CanManageInventory()
}

// canManageAlerts checks if user can manage alerts
func (s *alertService) canManageAlerts(user *models.User) bool {
	return user != nil && (user.IsAdmin() || user.IsOwner())
}

// canViewSpecificAlert checks if user can view a specific alert
func (s *alertService) canViewSpecificAlert(user *models.User, alert *models.Alert) bool {
	if user == nil || alert == nil {
		return false
	}

	// System alerts can be viewed by all users
	if alert.IsSystemAlert() {
		return true
	}

	// Product alerts can be viewed by users with inventory access
	if alert.IsProductAlert() {
		return user.CanManageInventory()
	}

	// User-specific alerts can only be viewed by the target user or admins/owners
	if alert.IsUserAlert() {
		return user.ID == *alert.UserID || user.IsAdmin() || user.IsOwner()
	}

	return false
}

// canManageSpecificAlert checks if user can manage a specific alert
func (s *alertService) canManageSpecificAlert(user *models.User, alert *models.Alert) bool {
	if user == nil || alert == nil {
		return false
	}

	// Only admins and owners can manage alerts
	if !user.IsAdmin() && !user.IsOwner() {
		return false
	}

	// System alerts can be managed by admins and owners
	if alert.IsSystemAlert() {
		return true
	}

	// Product alerts can be managed by admins and owners
	if alert.IsProductAlert() {
		return true
	}

	// User-specific alerts can be managed by admins and owners
	if alert.IsUserAlert() {
		return true
	}

	return false
}

// applyUserFiltering applies user-based filtering to alert list requests
func (s *alertService) applyUserFiltering(req *models.AlertListRequest, user *models.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	// Staff users can only see alerts for products they have access to
	// This is handled at the repository level by filtering by user permissions
	// For now, we don't apply additional filtering here

	return nil
}
