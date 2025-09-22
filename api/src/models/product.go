package models

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ProductType represents the type of product
type ProductType string

const (
	// ProductTypeTire represents a tire product
	ProductTypeTire ProductType = "Tire"
	// ProductTypeWheel represents a wheel product
	ProductTypeWheel ProductType = "Wheel"
)

// StockStatus represents the stock status of a product
type StockStatus string

const (
	// StockStatusAvailable represents products with sufficient stock
	StockStatusAvailable StockStatus = "available"
	// StockStatusLowStock represents products with low stock levels
	StockStatusLowStock StockStatus = "lowStock"
	// StockStatusOutOfStock represents products with no stock
	StockStatusOutOfStock StockStatus = "outOfStock"
)

// Product represents individual tire or wheel items with complete specifications
type Product struct {
	ID                uint        `json:"id" gorm:"primaryKey;autoIncrement"`
	Type              ProductType `json:"type" gorm:"not null;type:varchar(20)"`
	Brand             string      `json:"brand" gorm:"not null;size:100;index"`
	Model             string      `json:"model" gorm:"not null;size:100;index"`
	SKU               string      `json:"sku" gorm:"uniqueIndex;not null;size:50"`
	Description       *string     `json:"description,omitempty" gorm:"type:text"`
	ImageBase64       *string     `json:"imageBase64,omitempty" gorm:"type:text"`
	CostPrice         float64     `json:"costPrice" gorm:"not null;type:decimal(10,2)"`
	SellingPrice      float64     `json:"sellingPrice" gorm:"not null;type:decimal(10,2)"`
	QuantityOnHand    int         `json:"quantityOnHand" gorm:"not null;default:0;index"`
	LowStockThreshold int         `json:"lowStockThreshold" gorm:"not null;default:5"`
	IsActive          bool        `json:"isActive" gorm:"default:true;index"`
	CreatedAt         time.Time   `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt         time.Time   `json:"updatedAt" gorm:"autoUpdateTime"`

	// Relationships
	Specifications []ProductSpecification `json:"specifications,omitempty" gorm:"foreignKey:ProductID;references:ID"`
	StockMovements []StockMovement        `json:"-" gorm:"foreignKey:ProductID;references:ID"`
	Alerts         []Alert                `json:"-" gorm:"foreignKey:ProductID;references:ID"`
}

// ProductCreateRequest represents the request payload for creating a product
type ProductCreateRequest struct {
	Type              ProductType                         `json:"type" binding:"required"`
	Brand             string                              `json:"brand" binding:"required"`
	Model             string                              `json:"model" binding:"required"`
	SKU               string                              `json:"sku" binding:"required"`
	Description       *string                             `json:"description,omitempty"`
	ImageBase64       *string                             `json:"imageBase64,omitempty"`
	CostPrice         float64                             `json:"costPrice" binding:"required"`
	SellingPrice      float64                             `json:"sellingPrice" binding:"required"`
	QuantityOnHand    int                                 `json:"quantityOnHand"`
	LowStockThreshold int                                 `json:"lowStockThreshold"`
	Specifications    []ProductSpecificationCreateRequest `json:"specifications,omitempty"`
}

// ProductUpdateRequest represents the request payload for updating a product
type ProductUpdateRequest struct {
	Type              *ProductType `json:"type,omitempty"`
	Brand             *string      `json:"brand,omitempty"`
	Model             *string      `json:"model,omitempty"`
	SKU               *string      `json:"sku,omitempty"`
	Description       *string      `json:"description,omitempty"`
	ImageBase64       *string      `json:"imageBase64,omitempty"`
	CostPrice         *float64     `json:"costPrice,omitempty"`
	SellingPrice      *float64     `json:"sellingPrice,omitempty"`
	QuantityOnHand    *int         `json:"quantityOnHand,omitempty"`
	LowStockThreshold *int         `json:"lowStockThreshold,omitempty"`
	IsActive          *bool        `json:"isActive,omitempty"`
}

// ProductResponse represents the response payload for product data
type ProductResponse struct {
	ID                uint                           `json:"id"`
	Type              ProductType                    `json:"type"`
	Brand             string                         `json:"brand"`
	Model             string                         `json:"model"`
	SKU               string                         `json:"sku"`
	Description       *string                        `json:"description,omitempty"`
	ImageBase64       *string                        `json:"imageBase64,omitempty"`
	CostPrice         float64                        `json:"costPrice"`
	SellingPrice      float64                        `json:"sellingPrice"`
	QuantityOnHand    int                            `json:"quantityOnHand"`
	LowStockThreshold int                            `json:"lowStockThreshold"`
	StockStatus       StockStatus                    `json:"stockStatus"`
	IsActive          bool                           `json:"isActive"`
	CreatedAt         time.Time                      `json:"createdAt"`
	UpdatedAt         time.Time                      `json:"updatedAt"`
	Specifications    []ProductSpecificationResponse `json:"specifications,omitempty"`
}

// ProductSearchRequest represents the request payload for searching products
type ProductSearchRequest struct {
	Query       *string      `json:"query,omitempty"`
	Type        *ProductType `json:"type,omitempty"`
	Brand       *string      `json:"brand,omitempty"`
	StockStatus *StockStatus `json:"stockStatus,omitempty"`
	MinPrice    *float64     `json:"minPrice,omitempty"`
	MaxPrice    *float64     `json:"maxPrice,omitempty"`
	Page        int          `json:"page" binding:"min=1"`
	Limit       int          `json:"limit" binding:"min=1,max=100"`
}

// ProductSearchResponse represents the response payload for product search
type ProductSearchResponse struct {
	Products       []ProductResponse    `json:"products"`
	TotalCount     int64                `json:"totalCount"`
	SearchCriteria ProductSearchRequest `json:"searchCriteria"`
	Pagination     PaginationResponse   `json:"pagination"`
}

// ProductStatistics represents product statistics
type ProductStatistics struct {
	TotalProducts     int64   `json:"totalProducts"`
	TotalValue        float64 `json:"totalValue"`
	LowStockCount     int64   `json:"lowStockCount"`
	OutOfStockCount   int64   `json:"outOfStockCount"`
	AverageStockValue float64 `json:"averageStockValue"`
}

// BeforeCreate is a GORM hook that runs before creating a product
func (p *Product) BeforeCreate(_ *gorm.DB) error {
	// Validate product data
	if err := p.Validate(); err != nil {
		return err
	}

	// Set default values
	if p.QuantityOnHand < 0 {
		p.QuantityOnHand = 0
	}
	if p.LowStockThreshold <= 0 {
		p.LowStockThreshold = 5
	}

	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a product
func (p *Product) BeforeUpdate(tx *gorm.DB) error {
	// Validate product data
	if err := p.Validate(); err != nil {
		return err
	}

	return nil
}

// Validate validates the product data
func (p *Product) Validate() error {
	var validationErrors []string

	// Validate type
	if err := p.ValidateType(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate brand
	if err := p.ValidateBrand(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate model
	if err := p.ValidateModel(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate SKU
	if err := p.ValidateSKU(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate prices
	if err := p.ValidatePrices(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate quantity
	if err := p.ValidateQuantity(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// ValidateType validates the product type
func (p *Product) ValidateType() error {
	validTypes := []ProductType{ProductTypeTire, ProductTypeWheel}

	for _, productType := range validTypes {
		if p.Type == productType {
			return nil
		}
	}

	return errors.New("type must be one of: Tire, Wheel")
}

// ValidateBrand validates the product brand
func (p *Product) ValidateBrand() error {
	if p.Brand == "" {
		return errors.New("brand is required")
	}

	if len(p.Brand) < 1 || len(p.Brand) > 100 {
		return errors.New("brand must be 1-100 characters")
	}

	return nil
}

// ValidateModel validates the product model
func (p *Product) ValidateModel() error {
	if p.Model == "" {
		return errors.New("model is required")
	}

	if len(p.Model) < 1 || len(p.Model) > 100 {
		return errors.New("model must be 1-100 characters")
	}

	return nil
}

// ValidateSKU validates the product SKU
func (p *Product) ValidateSKU() error {
	if p.SKU == "" {
		return errors.New("SKU is required")
	}

	if len(p.SKU) < 3 || len(p.SKU) > 50 {
		return errors.New("SKU must be 3-50 characters")
	}

	return nil
}

// ValidatePrices validates the product prices
func (p *Product) ValidatePrices() error {
	if p.CostPrice <= 0 {
		return errors.New("cost price must be positive")
	}

	if p.SellingPrice <= 0 {
		return errors.New("selling price must be positive")
	}

	return nil
}

// ValidateQuantity validates the product quantity
func (p *Product) ValidateQuantity() error {
	if p.QuantityOnHand < 0 {
		return errors.New("quantity on hand cannot be negative")
	}

	if p.LowStockThreshold < 0 {
		return errors.New("low stock threshold cannot be negative")
	}

	return nil
}

// GetStockStatus returns the current stock status based on quantity
func (p *Product) GetStockStatus() StockStatus {
	if p.QuantityOnHand == 0 {
		return StockStatusOutOfStock
	} else if p.QuantityOnHand <= p.LowStockThreshold {
		return StockStatusLowStock
	}
	return StockStatusAvailable
}

// IsLowStock checks if the product is low on stock
func (p *Product) IsLowStock() bool {
	return p.QuantityOnHand <= p.LowStockThreshold && p.QuantityOnHand > 0
}

// IsOutOfStock checks if the product is out of stock
func (p *Product) IsOutOfStock() bool {
	return p.QuantityOnHand == 0
}

// IsAvailable checks if the product is available for sale
func (p *Product) IsAvailable() bool {
	return p.QuantityOnHand > p.LowStockThreshold
}

// CanSell checks if the product can be sold in the given quantity
func (p *Product) CanSell(quantity int) bool {
	return p.IsActive && p.QuantityOnHand >= quantity
}

// UpdateQuantity updates the quantity on hand
func (p *Product) UpdateQuantity(quantity int) error {
	if quantity < 0 {
		return errors.New("quantity cannot be negative")
	}
	p.QuantityOnHand = quantity
	return nil
}

// AdjustQuantity adjusts the quantity by the given amount
func (p *Product) AdjustQuantity(adjustment int) error {
	newQuantity := p.QuantityOnHand + adjustment
	if newQuantity < 0 {
		return errors.New("insufficient stock for this operation")
	}
	p.QuantityOnHand = newQuantity
	return nil
}

// GetTotalValue calculates the total value of the product in stock
func (p *Product) GetTotalValue() float64 {
	return float64(p.QuantityOnHand) * p.CostPrice
}

// GetProfitMargin calculates the profit margin percentage
func (p *Product) GetProfitMargin() float64 {
	if p.CostPrice == 0 {
		return 0
	}
	const percentageMultiplier = 100
	return ((p.SellingPrice - p.CostPrice) / p.CostPrice) * percentageMultiplier
}

// ToResponse converts a Product to ProductResponse
func (p *Product) ToResponse() ProductResponse {
	response := ProductResponse{
		ID:                p.ID,
		Type:              p.Type,
		Brand:             p.Brand,
		Model:             p.Model,
		SKU:               p.SKU,
		Description:       p.Description,
		ImageBase64:       p.ImageBase64,
		CostPrice:         p.CostPrice,
		SellingPrice:      p.SellingPrice,
		QuantityOnHand:    p.QuantityOnHand,
		LowStockThreshold: p.LowStockThreshold,
		StockStatus:       p.GetStockStatus(),
		IsActive:          p.IsActive,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}

	// Convert specifications if they exist
	if len(p.Specifications) > 0 {
		response.Specifications = make([]ProductSpecificationResponse, len(p.Specifications))
		for i := range p.Specifications {
			response.Specifications[i] = p.Specifications[i].ToResponse()
		}
	}

	return response
}

// TableName returns the table name for the Product model
func (Product) TableName() string {
	return "products"
}
