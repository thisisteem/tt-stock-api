package models

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// MovementType represents the type of stock movement
type MovementType string

const (
	// MovementTypeIncoming represents incoming stock movements
	MovementTypeIncoming MovementType = "Incoming"
	// MovementTypeOutgoing represents outgoing stock movements
	MovementTypeOutgoing MovementType = "Outgoing"
	// MovementTypeSale represents sales transactions
	MovementTypeSale MovementType = "Sale"
	// MovementTypeAdjustment represents stock adjustments
	MovementTypeAdjustment MovementType = "Adjustment"
	// MovementTypeReturn represents returned items
	MovementTypeReturn MovementType = "Return"
)

// StockMovement represents stock movement transactions
type StockMovement struct {
	ID           uint         `json:"id" gorm:"primaryKey;autoIncrement"`
	ProductID    uint         `json:"productId" gorm:"not null;index"`
	UserID       uint         `json:"userId" gorm:"not null;index"`
	MovementType MovementType `json:"movementType" gorm:"not null;type:varchar(20);index"`
	Quantity     int          `json:"quantity" gorm:"not null"`
	Reason       *string      `json:"reason,omitempty" gorm:"size:255"`
	Reference    *string      `json:"reference,omitempty" gorm:"size:100;index"`
	Notes        *string      `json:"notes,omitempty" gorm:"type:text"`
	CreatedAt    time.Time    `json:"createdAt" gorm:"autoCreateTime;index"`

	// Relationships
	Product Product `json:"product,omitempty" gorm:"foreignKey:ProductID;references:ID"`
	User    User    `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
}

// StockMovementCreateRequest represents the request payload for creating a stock movement
type StockMovementCreateRequest struct {
	ProductID    uint         `json:"productId" binding:"required"`
	MovementType MovementType `json:"movementType" binding:"required"`
	Quantity     int          `json:"quantity" binding:"required"`
	Reason       *string      `json:"reason,omitempty"`
	Reference    *string      `json:"reference,omitempty"`
	Notes        *string      `json:"notes,omitempty"`
}

// StockMovementUpdateRequest represents the request payload for updating a stock movement
type StockMovementUpdateRequest struct {
	MovementType *MovementType `json:"movementType,omitempty"`
	Quantity     *int          `json:"quantity,omitempty"`
	Reason       *string       `json:"reason,omitempty"`
	Reference    *string       `json:"reference,omitempty"`
	Notes        *string       `json:"notes,omitempty"`
}

// StockMovementResponse represents the response payload for stock movement data
type StockMovementResponse struct {
	ID           uint             `json:"id"`
	ProductID    uint             `json:"productId"`
	UserID       uint             `json:"userId"`
	MovementType MovementType     `json:"movementType"`
	Quantity     int              `json:"quantity"`
	Reason       *string          `json:"reason,omitempty"`
	Reference    *string          `json:"reference,omitempty"`
	Notes        *string          `json:"notes,omitempty"`
	CreatedAt    time.Time        `json:"createdAt"`
	Product      *ProductResponse `json:"product,omitempty"`
	User         *UserResponse    `json:"user,omitempty"`
}

// StockMovementListRequest represents the request payload for listing stock movements
type StockMovementListRequest struct {
	ProductID    *uint         `json:"productId,omitempty"`
	UserID       *uint         `json:"userId,omitempty"`
	MovementType *MovementType `json:"movementType,omitempty"`
	StartDate    *time.Time    `json:"startDate,omitempty"`
	EndDate      *time.Time    `json:"endDate,omitempty"`
	Page         int           `json:"page" binding:"min=1"`
	Limit        int           `json:"limit" binding:"min=1,max=100"`
}

// StockMovementListResponse represents the response payload for stock movement list
type StockMovementListResponse struct {
	Movements  []StockMovementResponse `json:"movements"`
	Pagination PaginationResponse      `json:"pagination"`
}

// SaleRequest represents the request payload for processing a sale
type SaleRequest struct {
	ProductID    uint    `json:"productId" binding:"required"`
	Quantity     int     `json:"quantity" binding:"required"`
	CustomerName *string `json:"customerName,omitempty"`
	Reference    *string `json:"reference,omitempty"`
	Notes        *string `json:"notes,omitempty"`
}

// SaleResponse represents the response payload for a sale
type SaleResponse struct {
	Movement       StockMovementResponse `json:"movement"`
	NewQuantity    int                   `json:"newQuantity"`
	AlertGenerated bool                  `json:"alertGenerated"`
}

// PaginationResponse represents pagination information
type PaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
	HasNext    bool  `json:"hasNext"`
	HasPrev    bool  `json:"hasPrev"`
}

// StockMovementSummary represents summary statistics for stock movements
type StockMovementSummary struct {
	TotalMovements  int64 `json:"totalMovements"`
	IncomingCount   int64 `json:"incomingCount"`
	OutgoingCount   int64 `json:"outgoingCount"`
	SaleCount       int64 `json:"saleCount"`
	AdjustmentCount int64 `json:"adjustmentCount"`
	ReturnCount     int64 `json:"returnCount"`
	NetChange       int   `json:"netChange"`
}

// BeforeCreate is a GORM hook that runs before creating a stock movement
func (sm *StockMovement) BeforeCreate(_ *gorm.DB) error {
	// Validate stock movement data
	if err := sm.Validate(); err != nil {
		return err
	}

	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a stock movement
func (sm *StockMovement) BeforeUpdate(_ *gorm.DB) error {
	// Validate stock movement data
	if err := sm.Validate(); err != nil {
		return err
	}

	return nil
}

// Validate validates the stock movement data
func (sm *StockMovement) Validate() error {
	var validationErrors []string

	// Validate product ID
	if sm.ProductID == 0 {
		validationErrors = append(validationErrors, "productId is required")
	}

	// Validate user ID
	if sm.UserID == 0 {
		validationErrors = append(validationErrors, "userId is required")
	}

	// Validate movement type
	if err := sm.ValidateMovementType(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate quantity
	if err := sm.ValidateQuantity(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate reason
	if err := sm.ValidateReason(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate reference
	if err := sm.ValidateReference(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// ValidateMovementType validates the movement type
func (sm *StockMovement) ValidateMovementType() error {
	validTypes := []MovementType{
		MovementTypeIncoming,
		MovementTypeOutgoing,
		MovementTypeSale,
		MovementTypeAdjustment,
		MovementTypeReturn,
	}

	for _, movementType := range validTypes {
		if sm.MovementType == movementType {
			return nil
		}
	}

	return errors.New("movementType must be one of: Incoming, Outgoing, Sale, Adjustment, Return")
}

// ValidateQuantity validates the quantity
func (sm *StockMovement) ValidateQuantity() error {
	if sm.Quantity == 0 {
		return errors.New("quantity cannot be zero")
	}

	// For outgoing movements, quantity should be negative
	if sm.MovementType == MovementTypeOutgoing || sm.MovementType == MovementTypeSale {
		if sm.Quantity > 0 {
			return errors.New("quantity must be negative for outgoing/sale movements")
		}
	} else {
		// For incoming movements, quantity should be positive
		if sm.Quantity < 0 {
			return errors.New("quantity must be positive for incoming/adjustment movements")
		}
	}

	return nil
}

// ValidateReason validates the reason
func (sm *StockMovement) ValidateReason() error {
	if sm.Reason != nil && len(*sm.Reason) > 255 {
		return errors.New("reason must be 255 characters or less")
	}
	return nil
}

// ValidateReference validates the reference
func (sm *StockMovement) ValidateReference() error {
	if sm.Reference != nil && len(*sm.Reference) > 100 {
		return errors.New("reference must be 100 characters or less")
	}
	return nil
}

// IsIncoming checks if the movement is incoming
func (sm *StockMovement) IsIncoming() bool {
	return sm.MovementType == MovementTypeIncoming
}

// IsOutgoing checks if the movement is outgoing
func (sm *StockMovement) IsOutgoing() bool {
	return sm.MovementType == MovementTypeOutgoing || sm.MovementType == MovementTypeSale
}

// IsSale checks if the movement is a sale
func (sm *StockMovement) IsSale() bool {
	return sm.MovementType == MovementTypeSale
}

// IsAdjustment checks if the movement is an adjustment
func (sm *StockMovement) IsAdjustment() bool {
	return sm.MovementType == MovementTypeAdjustment
}

// GetAbsoluteQuantity returns the absolute value of the quantity
func (sm *StockMovement) GetAbsoluteQuantity() int {
	if sm.Quantity < 0 {
		return -sm.Quantity
	}
	return sm.Quantity
}

// GetQuantityChange returns the quantity change (positive for incoming, negative for outgoing)
func (sm *StockMovement) GetQuantityChange() int {
	return sm.Quantity
}

// ToResponse converts a StockMovement to StockMovementResponse
func (sm *StockMovement) ToResponse() StockMovementResponse {
	response := StockMovementResponse{
		ID:           sm.ID,
		ProductID:    sm.ProductID,
		UserID:       sm.UserID,
		MovementType: sm.MovementType,
		Quantity:     sm.Quantity,
		Reason:       sm.Reason,
		Reference:    sm.Reference,
		Notes:        sm.Notes,
		CreatedAt:    sm.CreatedAt,
	}

	// Include product and user data if they are loaded
	if sm.Product.ID != 0 {
		productResponse := sm.Product.ToResponse()
		response.Product = &productResponse
	}

	if sm.User.ID != 0 {
		userResponse := sm.User.ToResponse()
		response.User = &userResponse
	}

	return response
}

// TableName returns the table name for the StockMovement model
func (StockMovement) TableName() string {
	return "stock_movements"
}
