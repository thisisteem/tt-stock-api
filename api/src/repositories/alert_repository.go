// Package repositories contains the repository layer implementations for the TT Stock Backend API.
// It provides data access interfaces and implementations using GORM for database operations.
package repositories

import (
	"context"
	"errors"

	"tt-stock-api/src/models"

	"gorm.io/gorm"
)

// AlertRepository defines the interface for alert data operations
type AlertRepository interface {
	// Create creates a new alert
	Create(ctx context.Context, alert *models.Alert) error

	// GetByID retrieves an alert by ID
	GetByID(ctx context.Context, id uint) (*models.Alert, error)

	// Update updates an existing alert
	Update(ctx context.Context, alert *models.Alert) error

	// Delete deletes an alert
	Delete(ctx context.Context, id uint) error

	// List retrieves alerts with pagination and filtering
	List(ctx context.Context, req *models.AlertListRequest) (*models.AlertListResponse, error)

	// GetAlertsByProduct retrieves alerts for a specific product
	GetAlertsByProduct(ctx context.Context, productID uint) ([]models.Alert, error)

	// GetAlertsByUser retrieves alerts for a specific user
	GetAlertsByUser(ctx context.Context, userID uint) ([]models.Alert, error)

	// GetActiveAlerts retrieves all active alerts
	GetActiveAlerts(ctx context.Context) ([]models.Alert, error)

	// GetUnreadAlerts retrieves all unread alerts
	GetUnreadAlerts(ctx context.Context) ([]models.Alert, error)

	// GetAlertsByType retrieves alerts by type
	GetAlertsByType(ctx context.Context, alertType models.AlertType) ([]models.Alert, error)

	// GetAlertsByPriority retrieves alerts by priority
	GetAlertsByPriority(ctx context.Context, priority models.AlertPriority) ([]models.Alert, error)

	// MarkAsRead marks an alert as read
	MarkAsRead(ctx context.Context, id uint) error

	// MarkAsUnread marks an alert as unread
	MarkAsUnread(ctx context.Context, id uint) error

	// MarkAllAsRead marks all alerts as read for a user
	MarkAllAsRead(ctx context.Context, userID uint) error

	// DeactivateAlert deactivates an alert
	DeactivateAlert(ctx context.Context, id uint) error

	// GetUnreadCount returns the count of unread alerts
	GetUnreadCount(ctx context.Context, userID *uint) (int64, error)

	// Count returns the total number of alerts
	Count(ctx context.Context) (int64, error)

	// CountActiveAlerts returns the number of active alerts
	CountActiveAlerts(ctx context.Context) (int64, error)
}

// alertRepository implements the AlertRepository interface
type alertRepository struct {
	db *gorm.DB
}

// NewAlertRepository creates a new AlertRepository instance
func NewAlertRepository(db *gorm.DB) AlertRepository {
	return &alertRepository{
		db: db,
	}
}

// Create creates a new alert
func (r *alertRepository) Create(ctx context.Context, alert *models.Alert) error {
	if alert == nil {
		return errors.New("alert cannot be nil")
	}

	// Create alert
	if err := r.db.WithContext(ctx).Create(alert).Error; err != nil {
		return err
	}

	return nil
}

// GetByID retrieves an alert by ID
func (r *alertRepository) GetByID(ctx context.Context, id uint) (*models.Alert, error) {
	if id == 0 {
		return nil, errors.New("alert ID cannot be zero")
	}

	var alert models.Alert
	if err := r.db.WithContext(ctx).Preload("Product").Preload("User").First(&alert, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("alert not found")
		}
		return nil, err
	}

	return &alert, nil
}

// Update updates an existing alert
func (r *alertRepository) Update(ctx context.Context, alert *models.Alert) error {
	if alert == nil {
		return errors.New("alert cannot be nil")
	}

	if alert.ID == 0 {
		return errors.New("alert ID cannot be zero")
	}

	// Check if alert exists
	_, err := r.GetByID(ctx, alert.ID)
	if err != nil {
		return err
	}

	// Update alert
	if err := r.db.WithContext(ctx).Save(alert).Error; err != nil {
		return err
	}

	return nil
}

// Delete deletes an alert
func (r *alertRepository) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("alert ID cannot be zero")
	}

	// Check if alert exists
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete alert
	if err := r.db.WithContext(ctx).Delete(&models.Alert{}, id).Error; err != nil {
		return err
	}

	return nil
}

// List retrieves alerts with pagination and filtering
func (r *alertRepository) List(ctx context.Context, req *models.AlertListRequest) (*models.AlertListResponse, error) {
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

	query := r.db.WithContext(ctx).Model(&models.Alert{}).Preload("Product").Preload("User")

	// Apply filters
	if req.ProductID != nil {
		query = query.Where("product_id = ?", *req.ProductID)
	}
	if req.UserID != nil {
		query = query.Where("user_id = ?", *req.UserID)
	}
	if req.AlertType != nil {
		query = query.Where("alert_type = ?", *req.AlertType)
	}
	if req.Priority != nil {
		query = query.Where("priority = ?", *req.Priority)
	}
	if req.IsRead != nil {
		query = query.Where("is_read = ?", *req.IsRead)
	}
	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Count unread alerts
	var unreadCount int64
	unreadQuery := r.db.WithContext(ctx).Model(&models.Alert{})
	if req.UserID != nil {
		unreadQuery = unreadQuery.Where("user_id = ?", *req.UserID)
	}
	if req.IsActive != nil {
		unreadQuery = unreadQuery.Where("is_active = ?", *req.IsActive)
	}
	unreadQuery = unreadQuery.Where("is_read = ?", false)
	if err := unreadQuery.Count(&unreadCount).Error; err != nil {
		return nil, err
	}

	// Calculate pagination
	offset := (req.Page - 1) * req.Limit
	totalPages := (total + int64(req.Limit) - 1) / int64(req.Limit)

	// Apply pagination and ordering
	var alerts []models.Alert
	if err := query.Order("created_at DESC").Offset(offset).Limit(req.Limit).Find(&alerts).Error; err != nil {
		return nil, err
	}

	// Convert to response format
	alertResponses := make([]models.AlertResponse, len(alerts))
	for i, alert := range alerts {
		alertResponses[i] = alert.ToResponse()
	}

	return &models.AlertListResponse{
		Alerts:      alertResponses,
		UnreadCount: unreadCount,
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

// GetAlertsByProduct retrieves alerts for a specific product
func (r *alertRepository) GetAlertsByProduct(ctx context.Context, productID uint) ([]models.Alert, error) {
	if productID == 0 {
		return nil, errors.New("product ID cannot be zero")
	}

	var alerts []models.Alert
	if err := r.db.WithContext(ctx).Preload("Product").Preload("User").Where("product_id = ?", productID).Order("created_at DESC").Find(&alerts).Error; err != nil {
		return nil, err
	}

	return alerts, nil
}

// GetAlertsByUser retrieves alerts for a specific user
func (r *alertRepository) GetAlertsByUser(ctx context.Context, userID uint) ([]models.Alert, error) {
	if userID == 0 {
		return nil, errors.New("user ID cannot be zero")
	}

	var alerts []models.Alert
	if err := r.db.WithContext(ctx).Preload("Product").Preload("User").Where("user_id = ?", userID).Order("created_at DESC").Find(&alerts).Error; err != nil {
		return nil, err
	}

	return alerts, nil
}

// GetActiveAlerts retrieves all active alerts
func (r *alertRepository) GetActiveAlerts(ctx context.Context) ([]models.Alert, error) {
	var alerts []models.Alert
	if err := r.db.WithContext(ctx).Preload("Product").Preload("User").Where("is_active = ?", true).Order("created_at DESC").Find(&alerts).Error; err != nil {
		return nil, err
	}

	return alerts, nil
}

// GetUnreadAlerts retrieves all unread alerts
func (r *alertRepository) GetUnreadAlerts(ctx context.Context) ([]models.Alert, error) {
	var alerts []models.Alert
	if err := r.db.WithContext(ctx).Preload("Product").Preload("User").Where("is_read = ? AND is_active = ?", false, true).Order("created_at DESC").Find(&alerts).Error; err != nil {
		return nil, err
	}

	return alerts, nil
}

// GetAlertsByType retrieves alerts by type
func (r *alertRepository) GetAlertsByType(ctx context.Context, alertType models.AlertType) ([]models.Alert, error) {
	var alerts []models.Alert
	if err := r.db.WithContext(ctx).Preload("Product").Preload("User").Where("alert_type = ? AND is_active = ?", alertType, true).Order("created_at DESC").Find(&alerts).Error; err != nil {
		return nil, err
	}

	return alerts, nil
}

// GetAlertsByPriority retrieves alerts by priority
func (r *alertRepository) GetAlertsByPriority(ctx context.Context, priority models.AlertPriority) ([]models.Alert, error) {
	var alerts []models.Alert
	if err := r.db.WithContext(ctx).Preload("Product").Preload("User").Where("priority = ? AND is_active = ?", priority, true).Order("created_at DESC").Find(&alerts).Error; err != nil {
		return nil, err
	}

	return alerts, nil
}

// MarkAsRead marks an alert as read
func (r *alertRepository) MarkAsRead(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("alert ID cannot be zero")
	}

	// Check if alert exists
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Update alert to read
	if err := r.db.WithContext(ctx).Model(&models.Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_read": true,
		"read_at": "NOW()",
	}).Error; err != nil {
		return err
	}

	return nil
}

// MarkAsUnread marks an alert as unread
func (r *alertRepository) MarkAsUnread(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("alert ID cannot be zero")
	}

	// Check if alert exists
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Update alert to unread
	if err := r.db.WithContext(ctx).Model(&models.Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_read": false,
		"read_at": nil,
	}).Error; err != nil {
		return err
	}

	return nil
}

// MarkAllAsRead marks all alerts as read for a user
func (r *alertRepository) MarkAllAsRead(ctx context.Context, userID uint) error {
	if userID == 0 {
		return errors.New("user ID cannot be zero")
	}

	// Update all user alerts to read
	if err := r.db.WithContext(ctx).Model(&models.Alert{}).Where("user_id = ?", userID).Updates(map[string]interface{}{
		"is_read": true,
		"read_at": "NOW()",
	}).Error; err != nil {
		return err
	}

	return nil
}

// DeactivateAlert deactivates an alert
func (r *alertRepository) DeactivateAlert(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("alert ID cannot be zero")
	}

	// Check if alert exists
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Update alert to inactive
	if err := r.db.WithContext(ctx).Model(&models.Alert{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
		return err
	}

	return nil
}

// GetUnreadCount returns the count of unread alerts
func (r *alertRepository) GetUnreadCount(ctx context.Context, userID *uint) (int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Alert{}).Where("is_read = ? AND is_active = ?", false, true)

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// Count returns the total number of alerts
func (r *alertRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Alert{}).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// CountActiveAlerts returns the number of active alerts
func (r *alertRepository) CountActiveAlerts(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Alert{}).Where("is_active = ?", true).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
