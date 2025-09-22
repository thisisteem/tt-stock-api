// Package validators contains input validation logic for the TT Stock Backend API.
// It provides validation structs, rules, and functions to ensure data integrity
// and security at the API boundary layer.
package validators

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"tt-stock-api/src/models"
)

// StockValidator handles stock-related input validation
type StockValidator struct{}

// NewStockValidator creates a new StockValidator instance
func NewStockValidator() *StockValidator {
	return &StockValidator{}
}

// ValidateStockMovementCreateRequest validates stock movement creation request data
func (v *StockValidator) ValidateStockMovementCreateRequest(req *models.StockMovementCreateRequest) error {
	if req == nil {
		return errors.New("stock movement creation request cannot be nil")
	}

	var validationErrors []string

	// Validate product ID
	if err := v.validateProductID(req.ProductID); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("product ID: %s", err.Error()))
	}

	// Validate movement type
	if err := v.validateMovementType(req.MovementType); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("movement type: %s", err.Error()))
	}

	// Validate quantity
	if err := v.validateMovementQuantity(req.Quantity); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("quantity: %s", err.Error()))
	}

	// Validate reason if provided
	if req.Reason != nil {
		if err := v.validateReason(*req.Reason); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("reason: %s", err.Error()))
		}
	}

	// Validate reference if provided
	if req.Reference != nil {
		if err := v.validateReference(*req.Reference); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("reference: %s", err.Error()))
		}
	}

	// Validate notes if provided
	if req.Notes != nil {
		if err := v.validateNotes(*req.Notes); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("notes: %s", err.Error()))
		}
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// ValidateStockMovementUpdateRequest validates stock movement update request data
func (v *StockValidator) ValidateStockMovementUpdateRequest(req *models.StockMovementUpdateRequest) error {
	if req == nil {
		return errors.New("stock movement update request cannot be nil")
	}

	var validationErrors []string

	// Validate movement type if provided
	if req.MovementType != nil {
		if err := v.validateMovementType(*req.MovementType); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("movement type: %s", err.Error()))
		}
	}

	// Validate quantity if provided
	if req.Quantity != nil {
		if err := v.validateMovementQuantity(*req.Quantity); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("quantity: %s", err.Error()))
		}
	}

	// Validate reason if provided
	if req.Reason != nil {
		if err := v.validateReason(*req.Reason); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("reason: %s", err.Error()))
		}
	}

	// Validate reference if provided
	if req.Reference != nil {
		if err := v.validateReference(*req.Reference); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("reference: %s", err.Error()))
		}
	}

	// Validate notes if provided
	if req.Notes != nil {
		if err := v.validateNotes(*req.Notes); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("notes: %s", err.Error()))
		}
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// ValidateStockMovementListRequest validates stock movement list request data
func (v *StockValidator) ValidateStockMovementListRequest(req *models.StockMovementListRequest) error {
	if req == nil {
		return errors.New("stock movement list request cannot be nil")
	}

	var validationErrors []string

	// Validate pagination
	if err := v.validatePagination(req.Page, req.Limit); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("pagination: %s", err.Error()))
	}

	// Validate product ID if provided
	if req.ProductID != nil {
		if err := v.validateProductID(*req.ProductID); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("product ID: %s", err.Error()))
		}
	}

	// Validate user ID if provided
	if req.UserID != nil {
		if err := v.validateUserID(*req.UserID); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("user ID: %s", err.Error()))
		}
	}

	// Validate movement type if provided
	if req.MovementType != nil {
		if err := v.validateMovementType(*req.MovementType); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("movement type: %s", err.Error()))
		}
	}

	// Validate date range if provided
	if req.StartDate != nil && req.EndDate != nil {
		if err := v.validateDateRange(*req.StartDate, *req.EndDate); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("date range: %s", err.Error()))
		}
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// ValidateSaleRequest validates sale request data
func (v *StockValidator) ValidateSaleRequest(req *models.SaleRequest) error {
	if req == nil {
		return errors.New("sale request cannot be nil")
	}

	var validationErrors []string

	// Validate product ID
	if err := v.validateProductID(req.ProductID); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("product ID: %s", err.Error()))
	}

	// Validate quantity
	if err := v.validateSaleQuantity(req.Quantity); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("quantity: %s", err.Error()))
	}

	// Validate customer name if provided
	if req.CustomerName != nil {
		if err := v.validateCustomerName(*req.CustomerName); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("customer name: %s", err.Error()))
		}
	}

	// Validate reference if provided
	if req.Reference != nil {
		if err := v.validateReference(*req.Reference); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("reference: %s", err.Error()))
		}
	}

	// Validate notes if provided
	if req.Notes != nil {
		if err := v.validateNotes(*req.Notes); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("notes: %s", err.Error()))
		}
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// ValidateStockAdjustmentRequest validates stock adjustment request data
func (v *StockValidator) ValidateStockAdjustmentRequest(req *models.StockMovementCreateRequest) error {
	if req == nil {
		return errors.New("stock adjustment request cannot be nil")
	}

	var validationErrors []string

	// Validate product ID
	if err := v.validateProductID(req.ProductID); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("product ID: %s", err.Error()))
	}

	// Validate movement type (must be adjustment)
	if req.MovementType != models.MovementTypeAdjustment {
		validationErrors = append(validationErrors, "movement type must be Adjustment for stock adjustments")
	}

	// Validate quantity (can be positive or negative for adjustments)
	if err := v.validateAdjustmentQuantity(req.Quantity); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("quantity: %s", err.Error()))
	}

	// Validate reason (required for adjustments)
	if req.Reason == nil || strings.TrimSpace(*req.Reason) == "" {
		validationErrors = append(validationErrors, "reason is required for stock adjustments")
	} else if err := v.validateReason(*req.Reason); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("reason: %s", err.Error()))
	}

	// Validate reference if provided
	if req.Reference != nil {
		if err := v.validateReference(*req.Reference); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("reference: %s", err.Error()))
		}
	}

	// Validate notes if provided
	if req.Notes != nil {
		if err := v.validateNotes(*req.Notes); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("notes: %s", err.Error()))
		}
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// Helper validation methods

// validateProductID validates product ID
func (v *StockValidator) validateProductID(productID uint) error {
	if productID == 0 {
		return errors.New("product ID is required")
	}

	if productID > 999999999 {
		return errors.New("product ID must not exceed 999,999,999")
	}

	return nil
}

// validateUserID validates user ID
func (v *StockValidator) validateUserID(userID uint) error {
	if userID == 0 {
		return errors.New("user ID is required")
	}

	if userID > 999999999 {
		return errors.New("user ID must not exceed 999,999,999")
	}

	return nil
}

// validateMovementType validates movement type
func (v *StockValidator) validateMovementType(movementType models.MovementType) error {
	switch movementType {
	case models.MovementTypeIncoming, models.MovementTypeOutgoing, models.MovementTypeSale,
		models.MovementTypeAdjustment, models.MovementTypeReturn:
		return nil
	default:
		return errors.New("movement type must be one of: Incoming, Outgoing, Sale, Adjustment, Return")
	}
}

// validateMovementQuantity validates movement quantity
func (v *StockValidator) validateMovementQuantity(quantity int) error {
	if quantity == 0 {
		return errors.New("quantity cannot be zero")
	}

	if quantity < -999999 {
		return errors.New("quantity must not be less than -999,999")
	}

	if quantity > 999999 {
		return errors.New("quantity must not exceed 999,999")
	}

	return nil
}

// validateSaleQuantity validates sale quantity
func (v *StockValidator) validateSaleQuantity(quantity int) error {
	if quantity <= 0 {
		return errors.New("sale quantity must be positive")
	}

	if quantity > 999999 {
		return errors.New("sale quantity must not exceed 999,999")
	}

	return nil
}

// validateAdjustmentQuantity validates adjustment quantity
func (v *StockValidator) validateAdjustmentQuantity(quantity int) error {
	if quantity == 0 {
		return errors.New("adjustment quantity cannot be zero")
	}

	if quantity < -999999 {
		return errors.New("adjustment quantity must not be less than -999,999")
	}

	if quantity > 999999 {
		return errors.New("adjustment quantity must not exceed 999,999")
	}

	return nil
}

// validateReason validates reason
func (v *StockValidator) validateReason(reason string) error {
	if strings.TrimSpace(reason) == "" {
		return nil // Reason is optional for most movements
	}

	trimmedReason := strings.TrimSpace(reason)
	if len(trimmedReason) < 3 {
		return errors.New("reason must be at least 3 characters")
	}

	if len(trimmedReason) > 200 {
		return errors.New("reason must not exceed 200 characters")
	}

	return nil
}

// validateReference validates reference
func (v *StockValidator) validateReference(reference string) error {
	if strings.TrimSpace(reference) == "" {
		return nil // Reference is optional
	}

	trimmedRef := strings.TrimSpace(reference)
	if len(trimmedRef) < 3 {
		return errors.New("reference must be at least 3 characters")
	}

	if len(trimmedRef) > 100 {
		return errors.New("reference must not exceed 100 characters")
	}

	return nil
}

// validateNotes validates notes
func (v *StockValidator) validateNotes(notes string) error {
	if strings.TrimSpace(notes) == "" {
		return nil // Notes are optional
	}

	if len(notes) > 1000 {
		return errors.New("notes must not exceed 1000 characters")
	}

	return nil
}

// validateCustomerName validates customer name
func (v *StockValidator) validateCustomerName(name string) error {
	if strings.TrimSpace(name) == "" {
		return nil // Customer name is optional
	}

	trimmedName := strings.TrimSpace(name)
	if len(trimmedName) < 2 {
		return errors.New("customer name must be at least 2 characters")
	}

	if len(trimmedName) > 100 {
		return errors.New("customer name must not exceed 100 characters")
	}

	// Check if name contains only letters, spaces, hyphens, and apostrophes
	nameRegex := regexp.MustCompile(`^[a-zA-Z\s\-']+$`)
	if !nameRegex.MatchString(trimmedName) {
		return errors.New("customer name can only contain letters, spaces, hyphens, and apostrophes")
	}

	return nil
}

// validatePagination validates pagination parameters
func (v *StockValidator) validatePagination(page, limit int) error {
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

// validateDateRange validates date range
func (v *StockValidator) validateDateRange(startDate, endDate time.Time) error {
	if startDate.IsZero() || endDate.IsZero() {
		return errors.New("start date and end date are required")
	}

	if startDate.After(endDate) {
		return errors.New("start date cannot be after end date")
	}

	if endDate.After(time.Now()) {
		return errors.New("end date cannot be in the future")
	}

	// Check if date range is not too wide (more than 1 year)
	if endDate.Sub(startDate) > 365*24*time.Hour {
		return errors.New("date range cannot exceed 1 year")
	}

	return nil
}

// Validation constants
const (
	MaxProductID          = 999999999
	MaxUserID             = 999999999
	MaxMovementQuantity   = 999999
	MinMovementQuantity   = -999999
	MaxSaleQuantity       = 999999
	MinReasonLength       = 3
	MaxReasonLength       = 200
	MinReferenceLength    = 3
	MaxReferenceLength    = 100
	MaxNotesLength        = 1000
	MaxCustomerNameLength = 100
	MinCustomerNameLength = 2
	MaxDateRangeDays      = 365
)
