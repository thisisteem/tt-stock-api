// Package validators contains input validation logic for the TT Stock Backend API.
// It provides validation structs, rules, and functions to ensure data integrity
// and security at the API boundary layer.
package validators

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"tt-stock-api/src/models"
)

// ProductValidator handles product-related input validation
type ProductValidator struct{}

// NewProductValidator creates a new ProductValidator instance
func NewProductValidator() *ProductValidator {
	return &ProductValidator{}
}

// ValidateProductCreateRequest validates product creation request data
func (v *ProductValidator) ValidateProductCreateRequest(req *models.ProductCreateRequest) error {
	if req == nil {
		return errors.New("product creation request cannot be nil")
	}

	var validationErrors []string

	// Validate product type
	if err := v.validateProductType(req.Type); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("type: %s", err.Error()))
	}

	// Validate brand
	if err := v.validateBrand(req.Brand); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("brand: %s", err.Error()))
	}

	// Validate model
	if err := v.validateModel(req.Model); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("model: %s", err.Error()))
	}

	// Validate SKU
	if err := v.validateSKU(req.SKU); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("SKU: %s", err.Error()))
	}

	// Validate description if provided
	if req.Description != nil {
		if err := v.validateDescription(*req.Description); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("description: %s", err.Error()))
		}
	}

	// Validate image if provided
	if req.ImageBase64 != nil {
		if err := v.validateImageBase64(*req.ImageBase64); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("image: %s", err.Error()))
		}
	}

	// Validate cost price
	if err := v.validatePrice(req.CostPrice, "cost price"); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("cost price: %s", err.Error()))
	}

	// Validate selling price
	if err := v.validatePrice(req.SellingPrice, "selling price"); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("selling price: %s", err.Error()))
	}

	// Validate quantity on hand
	if err := v.validateQuantity(req.QuantityOnHand, "quantity on hand"); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("quantity on hand: %s", err.Error()))
	}

	// Validate low stock threshold
	if err := v.validateQuantity(req.LowStockThreshold, "low stock threshold"); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("low stock threshold: %s", err.Error()))
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// ValidateProductUpdateRequest validates product update request data
func (v *ProductValidator) ValidateProductUpdateRequest(req *models.ProductUpdateRequest) error {
	if req == nil {
		return errors.New("product update request cannot be nil")
	}

	var validationErrors []string

	// Validate product type if provided
	if req.Type != nil {
		if err := v.validateProductType(*req.Type); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("type: %s", err.Error()))
		}
	}

	// Validate brand if provided
	if req.Brand != nil {
		if err := v.validateBrand(*req.Brand); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("brand: %s", err.Error()))
		}
	}

	// Validate model if provided
	if req.Model != nil {
		if err := v.validateModel(*req.Model); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("model: %s", err.Error()))
		}
	}

	// Validate SKU if provided
	if req.SKU != nil {
		if err := v.validateSKU(*req.SKU); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("SKU: %s", err.Error()))
		}
	}

	// Validate description if provided
	if req.Description != nil {
		if err := v.validateDescription(*req.Description); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("description: %s", err.Error()))
		}
	}

	// Validate image if provided
	if req.ImageBase64 != nil {
		if err := v.validateImageBase64(*req.ImageBase64); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("image: %s", err.Error()))
		}
	}

	// Validate cost price if provided
	if req.CostPrice != nil {
		if err := v.validatePrice(*req.CostPrice, "cost price"); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("cost price: %s", err.Error()))
		}
	}

	// Validate selling price if provided
	if req.SellingPrice != nil {
		if err := v.validatePrice(*req.SellingPrice, "selling price"); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("selling price: %s", err.Error()))
		}
	}

	// Validate quantity on hand if provided
	if req.QuantityOnHand != nil {
		if err := v.validateQuantity(*req.QuantityOnHand, "quantity on hand"); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("quantity on hand: %s", err.Error()))
		}
	}

	// Validate low stock threshold if provided
	if req.LowStockThreshold != nil {
		if err := v.validateQuantity(*req.LowStockThreshold, "low stock threshold"); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("low stock threshold: %s", err.Error()))
		}
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// ValidateProductSearchRequest validates product search request data
func (v *ProductValidator) ValidateProductSearchRequest(req *models.ProductSearchRequest) error {
	if req == nil {
		return errors.New("product search request cannot be nil")
	}

	var validationErrors []string

	// Validate pagination
	if err := v.validatePagination(req.Page, req.Limit); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("pagination: %s", err.Error()))
	}

	// Validate query if provided
	if req.Query != nil {
		if err := v.validateSearchQuery(*req.Query); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("query: %s", err.Error()))
		}
	}

	// Validate product type if provided
	if req.Type != nil {
		if err := v.validateProductType(*req.Type); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("type: %s", err.Error()))
		}
	}

	// Validate brand if provided
	if req.Brand != nil {
		if err := v.validateBrand(*req.Brand); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("brand: %s", err.Error()))
		}
	}

	// Validate stock status if provided
	if req.StockStatus != nil {
		if err := v.validateStockStatus(*req.StockStatus); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("stock status: %s", err.Error()))
		}
	}

	// Validate price range if provided
	if req.MinPrice != nil && req.MaxPrice != nil {
		if *req.MinPrice > *req.MaxPrice {
			validationErrors = append(validationErrors, "minimum price cannot be greater than maximum price")
		}
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// ValidateProductSpecificationRequest validates product specification request data
func (v *ProductValidator) ValidateProductSpecificationRequest(req *models.ProductSpecificationCreateRequest) error {
	if req == nil {
		return errors.New("product specification request cannot be nil")
	}

	var validationErrors []string

	// Validate specification type
	if err := v.validateSpecificationType(req.SpecType); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("specType: %s", err.Error()))
	}

	// Validate spec data
	if err := v.validateSpecificationData(req.SpecData); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("specData: %s", err.Error()))
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// Helper validation methods

// validateProductType validates product type
func (v *ProductValidator) validateProductType(productType models.ProductType) error {
	switch productType {
	case models.ProductTypeTire, models.ProductTypeWheel:
		return nil
	default:
		return errors.New("product type must be one of: Tire, Wheel")
	}
}

// validateBrand validates product brand
func (v *ProductValidator) validateBrand(brand string) error {
	if strings.TrimSpace(brand) == "" {
		return errors.New("brand is required")
	}

	trimmedBrand := strings.TrimSpace(brand)
	if len(trimmedBrand) < 1 {
		return errors.New("brand must be at least 1 character")
	}

	if len(trimmedBrand) > 100 {
		return errors.New("brand must not exceed 100 characters")
	}

	return nil
}

// validateModel validates product model
func (v *ProductValidator) validateModel(model string) error {
	if strings.TrimSpace(model) == "" {
		return errors.New("model is required")
	}

	trimmedModel := strings.TrimSpace(model)
	if len(trimmedModel) < 1 {
		return errors.New("model must be at least 1 character")
	}

	if len(trimmedModel) > 100 {
		return errors.New("model must not exceed 100 characters")
	}

	return nil
}

// validateSKU validates product SKU
func (v *ProductValidator) validateSKU(sku string) error {
	if strings.TrimSpace(sku) == "" {
		return errors.New("SKU is required")
	}

	trimmedSKU := strings.TrimSpace(sku)
	if len(trimmedSKU) < 3 {
		return errors.New("SKU must be at least 3 characters")
	}

	if len(trimmedSKU) > 50 {
		return errors.New("SKU must not exceed 50 characters")
	}

	// Check if SKU contains only alphanumeric characters, hyphens, and underscores
	skuRegex := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
	if !skuRegex.MatchString(trimmedSKU) {
		return errors.New("SKU can only contain letters, numbers, hyphens, and underscores")
	}

	return nil
}

// validateDescription validates product description
func (v *ProductValidator) validateDescription(description string) error {
	if strings.TrimSpace(description) == "" {
		return nil // Description is optional
	}

	if len(description) > 1000 {
		return errors.New("description must not exceed 1000 characters")
	}

	return nil
}

// validateImageBase64 validates base64 image data
func (v *ProductValidator) validateImageBase64(imageBase64 string) error {
	if strings.TrimSpace(imageBase64) == "" {
		return nil // Image is optional
	}

	// Check if it's a valid base64 string
	base64Regex := regexp.MustCompile(`^[A-Za-z0-9+/]*={0,2}$`)
	if !base64Regex.MatchString(imageBase64) {
		return errors.New("invalid base64 image format")
	}

	// Check size limit (approximately 2MB for base64)
	if len(imageBase64) > 3000000 {
		return errors.New("image size must not exceed 2MB")
	}

	return nil
}

// validatePrice validates price values
func (v *ProductValidator) validatePrice(price float64, fieldName string) error {
	if price < 0 {
		return fmt.Errorf("%s must be non-negative", fieldName)
	}

	if price > 999999.99 {
		return fmt.Errorf("%s must not exceed 999,999.99", fieldName)
	}

	return nil
}

// validateQuantity validates quantity values
func (v *ProductValidator) validateQuantity(quantity int, fieldName string) error {
	if quantity < 0 {
		return fmt.Errorf("%s must be non-negative", fieldName)
	}

	if quantity > 999999 {
		return fmt.Errorf("%s must not exceed 999,999", fieldName)
	}

	return nil
}

// validatePagination validates pagination parameters
func (v *ProductValidator) validatePagination(page, limit int) error {
	if page < 1 {
		return errors.New("page must be at least 1")
	}

	if limit < 1 {
		return errors.New("limit must be at least 1")
	}

	if limit > 100 {
		return errors.New("limit must not exceed 100")
	}

	return nil
}

// validateSearchQuery validates search query
func (v *ProductValidator) validateSearchQuery(query string) error {
	if strings.TrimSpace(query) == "" {
		return nil // Empty query is allowed
	}

	if len(query) > 200 {
		return errors.New("search query must not exceed 200 characters")
	}

	return nil
}

// validateStockStatus validates stock status
func (v *ProductValidator) validateStockStatus(status models.StockStatus) error {
	switch status {
	case models.StockStatusAvailable, models.StockStatusLowStock, models.StockStatusOutOfStock:
		return nil
	default:
		return errors.New("stock status must be one of: Available, LowStock, OutOfStock")
	}
}

// validateSpecificationType validates specification type
func (v *ProductValidator) validateSpecificationType(specType models.SpecificationType) error {
	switch specType {
	case models.SpecificationTypeTire, models.SpecificationTypeWheel:
		return nil
	default:
		return errors.New("specification type must be one of: Tire, Wheel")
	}
}

// validateSpecificationData validates specification data
func (v *ProductValidator) validateSpecificationData(specData interface{}) error {
	if specData == nil {
		return errors.New("specification data is required")
	}

	// Convert to string for validation
	specDataStr := fmt.Sprintf("%v", specData)
	if strings.TrimSpace(specDataStr) == "" {
		return errors.New("specification data cannot be empty")
	}

	if len(specDataStr) > 500 {
		return errors.New("specification data must not exceed 500 characters")
	}

	return nil
}

// Validation constants
const (
	MinBrandLength             = 1
	MaxBrandLength             = 100
	MinModelLength             = 1
	MaxModelLength             = 100
	MinSKULength               = 3
	MaxSKULength               = 50
	MaxDescriptionLength       = 1000
	MaxImageBase64Length       = 3000000 // ~2MB
	MaxPrice                   = 999999.99
	MaxQuantity                = 999999
	MinPage                    = 1
	MaxLimit                   = 100
	MaxSearchQueryLength       = 200
	MaxSpecificationDataLength = 500
)
