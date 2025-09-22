package models

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// AlertType represents the type of alert
type AlertType string

const (
	// AlertTypeLowStock represents low stock alerts
	AlertTypeLowStock AlertType = "LowStock"
	// AlertTypeOutOfStock represents out of stock alerts
	AlertTypeOutOfStock AlertType = "OutOfStock"
	// AlertTypeExpired represents expired product alerts
	AlertTypeExpired AlertType = "Expired"
	// AlertTypeMaintenance represents maintenance alerts
	AlertTypeMaintenance AlertType = "Maintenance"
	// AlertTypeSystem represents system alerts
	AlertTypeSystem AlertType = "System"
)

// AlertPriority represents the priority level of an alert
type AlertPriority string

const (
	// AlertPriorityLow represents low priority alerts
	AlertPriorityLow AlertPriority = "Low"
	// AlertPriorityMedium represents medium priority alerts
	AlertPriorityMedium AlertPriority = "Medium"
	// AlertPriorityHigh represents high priority alerts
	AlertPriorityHigh AlertPriority = "High"
	// AlertPriorityCritical represents critical priority alerts
	AlertPriorityCritical AlertPriority = "Critical"
)

// Alert represents system alerts and notifications
type Alert struct {
	ID        uint          `json:"id" gorm:"primaryKey;autoIncrement"`
	ProductID *uint         `json:"productId,omitempty" gorm:"index"` // Nullable for system alerts
	UserID    *uint         `json:"userId,omitempty" gorm:"index"`    // Nullable for system-wide alerts
	AlertType AlertType     `json:"alertType" gorm:"not null;type:varchar(20);index"`
	Priority  AlertPriority `json:"priority" gorm:"not null;type:varchar(20);default:'Medium';index"`
	Title     string        `json:"title" gorm:"not null;size:255"`
	Message   string        `json:"message" gorm:"not null;type:text"`
	IsRead    bool          `json:"isRead" gorm:"default:false;index"`
	IsActive  bool          `json:"isActive" gorm:"default:true;index"`
	CreatedAt time.Time     `json:"createdAt" gorm:"autoCreateTime;index"`
	UpdatedAt time.Time     `json:"updatedAt" gorm:"autoUpdateTime"`
	ReadAt    *time.Time    `json:"readAt,omitempty" gorm:"index"`

	// Relationships
	Product *Product `json:"product,omitempty" gorm:"foreignKey:ProductID;references:ID"`
	User    *User    `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
}

// AlertCreateRequest represents the request payload for creating an alert
type AlertCreateRequest struct {
	ProductID *uint         `json:"productId,omitempty"`
	UserID    *uint         `json:"userId,omitempty"`
	AlertType AlertType     `json:"alertType" binding:"required"`
	Priority  AlertPriority `json:"priority"`
	Title     string        `json:"title" binding:"required"`
	Message   string        `json:"message" binding:"required"`
}

// AlertUpdateRequest represents the request payload for updating an alert
type AlertUpdateRequest struct {
	Priority *AlertPriority `json:"priority,omitempty"`
	Title    *string        `json:"title,omitempty"`
	Message  *string        `json:"message,omitempty"`
	IsRead   *bool          `json:"isRead,omitempty"`
	IsActive *bool          `json:"isActive,omitempty"`
}

// AlertResponse represents the response payload for alert data
type AlertResponse struct {
	ID        uint             `json:"id"`
	ProductID *uint            `json:"productId,omitempty"`
	UserID    *uint            `json:"userId,omitempty"`
	AlertType AlertType        `json:"alertType"`
	Priority  AlertPriority    `json:"priority"`
	Title     string           `json:"title"`
	Message   string           `json:"message"`
	IsRead    bool             `json:"isRead"`
	IsActive  bool             `json:"isActive"`
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
	ReadAt    *time.Time       `json:"readAt,omitempty"`
	Product   *ProductResponse `json:"product,omitempty"`
	User      *UserResponse    `json:"user,omitempty"`
}

// AlertListRequest represents the request payload for listing alerts
type AlertListRequest struct {
	ProductID *uint          `json:"productId,omitempty"`
	UserID    *uint          `json:"userId,omitempty"`
	AlertType *AlertType     `json:"alertType,omitempty"`
	Priority  *AlertPriority `json:"priority,omitempty"`
	IsRead    *bool          `json:"isRead,omitempty"`
	IsActive  *bool          `json:"isActive,omitempty"`
	Page      int            `json:"page" binding:"min=1"`
	Limit     int            `json:"limit" binding:"min=1,max=100"`
}

// AlertListResponse represents the response payload for alert list
type AlertListResponse struct {
	Alerts      []AlertResponse    `json:"alerts"`
	UnreadCount int64              `json:"unreadCount"`
	Pagination  PaginationResponse `json:"pagination"`
}

// BeforeCreate is a GORM hook that runs before creating an alert
func (a *Alert) BeforeCreate(_ *gorm.DB) error {
	// Validate alert data
	if err := a.Validate(); err != nil {
		return err
	}

	// Set default priority if not provided
	if a.Priority == "" {
		a.Priority = AlertPriorityMedium
	}

	return nil
}

// BeforeUpdate is a GORM hook that runs before updating an alert
func (a *Alert) BeforeUpdate(tx *gorm.DB) error {
	// Validate alert data
	if err := a.Validate(); err != nil {
		return err
	}

	// Update readAt timestamp when marking as read
	if tx.Statement.Changed("IsRead") && a.IsRead && a.ReadAt == nil {
		now := time.Now()
		a.ReadAt = &now
	}

	return nil
}

// Validate validates the alert data
func (a *Alert) Validate() error {
	var validationErrors []string

	// Validate alert type
	if err := a.ValidateAlertType(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate priority
	if err := a.ValidatePriority(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate title
	if err := a.ValidateTitle(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate message
	if err := a.ValidateMessage(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// ValidateAlertType validates the alert type
func (a *Alert) ValidateAlertType() error {
	validTypes := []AlertType{
		AlertTypeLowStock,
		AlertTypeOutOfStock,
		AlertTypeExpired,
		AlertTypeMaintenance,
		AlertTypeSystem,
	}

	for _, alertType := range validTypes {
		if a.AlertType == alertType {
			return nil
		}
	}

	return errors.New("alertType must be one of: LowStock, OutOfStock, Expired, Maintenance, System")
}

// ValidatePriority validates the alert priority
func (a *Alert) ValidatePriority() error {
	if a.Priority == "" {
		return nil // Will be set to default in BeforeCreate
	}

	validPriorities := []AlertPriority{
		AlertPriorityLow,
		AlertPriorityMedium,
		AlertPriorityHigh,
		AlertPriorityCritical,
	}

	for _, priority := range validPriorities {
		if a.Priority == priority {
			return nil
		}
	}

	return errors.New("priority must be one of: Low, Medium, High, Critical")
}

// ValidateTitle validates the alert title
func (a *Alert) ValidateTitle() error {
	if a.Title == "" {
		return errors.New("title is required")
	}

	const maxTitleLength = 255
	if len(a.Title) > maxTitleLength {
		return errors.New("title must be 255 characters or less")
	}

	return nil
}

// ValidateMessage validates the alert message
func (a *Alert) ValidateMessage() error {
	if a.Message == "" {
		return errors.New("message is required")
	}

	return nil
}

// MarkAsRead marks the alert as read
func (a *Alert) MarkAsRead() {
	a.IsRead = true
	now := time.Now()
	a.ReadAt = &now
}

// MarkAsUnread marks the alert as unread
func (a *Alert) MarkAsUnread() {
	a.IsRead = false
	a.ReadAt = nil
}

// Deactivate deactivates the alert
func (a *Alert) Deactivate() {
	a.IsActive = false
}

// Activate activates the alert
func (a *Alert) Activate() {
	a.IsActive = true
}

// IsHighPriority checks if the alert is high priority
func (a *Alert) IsHighPriority() bool {
	return a.Priority == AlertPriorityHigh || a.Priority == AlertPriorityCritical
}

// IsCritical checks if the alert is critical
func (a *Alert) IsCritical() bool {
	return a.Priority == AlertPriorityCritical
}

// IsProductAlert checks if the alert is related to a product
func (a *Alert) IsProductAlert() bool {
	return a.ProductID != nil
}

// IsUserAlert checks if the alert is related to a specific user
func (a *Alert) IsUserAlert() bool {
	return a.UserID != nil
}

// IsSystemAlert checks if the alert is a system-wide alert
func (a *Alert) IsSystemAlert() bool {
	return a.AlertType == AlertTypeSystem
}

// GetAge returns the age of the alert
func (a *Alert) GetAge() time.Duration {
	return time.Since(a.CreatedAt)
}

// IsOld checks if the alert is older than the given duration
func (a *Alert) IsOld(duration time.Duration) bool {
	return a.GetAge() > duration
}

// ToResponse converts an Alert to AlertResponse
func (a *Alert) ToResponse() AlertResponse {
	response := AlertResponse{
		ID:        a.ID,
		ProductID: a.ProductID,
		UserID:    a.UserID,
		AlertType: a.AlertType,
		Priority:  a.Priority,
		Title:     a.Title,
		Message:   a.Message,
		IsRead:    a.IsRead,
		IsActive:  a.IsActive,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
		ReadAt:    a.ReadAt,
	}

	// Include product data if it is loaded
	if a.Product != nil && a.Product.ID != 0 {
		productResponse := a.Product.ToResponse()
		response.Product = &productResponse
	}

	// Include user data if it is loaded
	if a.User != nil && a.User.ID != 0 {
		userResponse := a.User.ToResponse()
		response.User = &userResponse
	}

	return response
}

// TableName returns the table name for the Alert model
func (Alert) TableName() string {
	return "alerts"
}
