// Package unit contains unit tests for the TT Stock Backend API models.
// It tests model validation, business logic, and data transformation methods.
package unit

import (
	"testing"
	"time"

	"tt-stock-api/src/models"

	"github.com/stretchr/testify/assert"
)

// TestStockMovement_Validate tests the StockMovement validation functionality
func TestStockMovement_Validate(t *testing.T) {
	tests := []struct {
		name     string
		movement models.StockMovement
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid incoming movement",
			movement: models.StockMovement{
				ProductID:    1,
				UserID:       1,
				MovementType: models.MovementTypeIncoming,
				Quantity:     10,
			},
			wantErr: false,
		},
		{
			name: "valid outgoing movement",
			movement: models.StockMovement{
				ProductID:    1,
				UserID:       1,
				MovementType: models.MovementTypeOutgoing,
				Quantity:     -5,
			},
			wantErr: false,
		},
		{
			name: "valid sale movement",
			movement: models.StockMovement{
				ProductID:    1,
				UserID:       1,
				MovementType: models.MovementTypeSale,
				Quantity:     -2,
			},
			wantErr: false,
		},
		{
			name: "valid adjustment movement",
			movement: models.StockMovement{
				ProductID:    1,
				UserID:       1,
				MovementType: models.MovementTypeAdjustment,
				Quantity:     3,
			},
			wantErr: false,
		},
		{
			name: "valid return movement",
			movement: models.StockMovement{
				ProductID:    1,
				UserID:       1,
				MovementType: models.MovementTypeReturn,
				Quantity:     1,
			},
			wantErr: false,
		},
		{
			name: "zero product ID",
			movement: models.StockMovement{
				ProductID:    0,
				UserID:       1,
				MovementType: models.MovementTypeIncoming,
				Quantity:     10,
			},
			wantErr: true,
			errMsg:  "productId is required",
		},
		{
			name: "zero user ID",
			movement: models.StockMovement{
				ProductID:    1,
				UserID:       0,
				MovementType: models.MovementTypeIncoming,
				Quantity:     10,
			},
			wantErr: true,
			errMsg:  "userId is required",
		},
		{
			name: "invalid movement type",
			movement: models.StockMovement{
				ProductID:    1,
				UserID:       1,
				MovementType: "InvalidType",
				Quantity:     10,
			},
			wantErr: true,
			errMsg:  "movementType must be one of: Incoming, Outgoing, Sale, Adjustment, Return",
		},
		{
			name: "zero quantity",
			movement: models.StockMovement{
				ProductID:    1,
				UserID:       1,
				MovementType: models.MovementTypeIncoming,
				Quantity:     0,
			},
			wantErr: true,
			errMsg:  "quantity cannot be zero",
		},
		{
			name: "negative quantity for incoming",
			movement: models.StockMovement{
				ProductID:    1,
				UserID:       1,
				MovementType: models.MovementTypeIncoming,
				Quantity:     -5,
			},
			wantErr: true,
			errMsg:  "quantity must be positive for incoming/adjustment movements",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.movement.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestStockMovement_ValidateMovementType tests movement type validation
func TestStockMovement_ValidateMovementType(t *testing.T) {
	tests := []struct {
		name     string
		movement models.StockMovement
		wantErr  bool
	}{
		{"valid incoming", models.StockMovement{MovementType: models.MovementTypeIncoming}, false},
		{"valid outgoing", models.StockMovement{MovementType: models.MovementTypeOutgoing}, false},
		{"valid sale", models.StockMovement{MovementType: models.MovementTypeSale}, false},
		{"valid adjustment", models.StockMovement{MovementType: models.MovementTypeAdjustment}, false},
		{"valid return", models.StockMovement{MovementType: models.MovementTypeReturn}, false},
		{"invalid type", models.StockMovement{MovementType: "InvalidType"}, true},
		{"empty type", models.StockMovement{MovementType: ""}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.movement.ValidateMovementType()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "movementType must be one of: Incoming, Outgoing, Sale, Adjustment, Return")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestStockMovement_ValidateQuantity tests quantity validation
func TestStockMovement_ValidateQuantity(t *testing.T) {
	tests := []struct {
		name     string
		movement models.StockMovement
		wantErr  bool
		errMsg   string
	}{
		{"positive quantity for incoming", models.StockMovement{MovementType: models.MovementTypeIncoming, Quantity: 10}, false, ""},
		{"zero quantity", models.StockMovement{MovementType: models.MovementTypeIncoming, Quantity: 0}, true, "quantity cannot be zero"},
		{"negative quantity for incoming", models.StockMovement{MovementType: models.MovementTypeIncoming, Quantity: -5}, true, "quantity must be positive for incoming/adjustment movements"},
		{"negative quantity for outgoing", models.StockMovement{MovementType: models.MovementTypeOutgoing, Quantity: -5}, false, ""},
		{"positive quantity for outgoing", models.StockMovement{MovementType: models.MovementTypeOutgoing, Quantity: 5}, true, "quantity must be negative for outgoing/sale movements"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.movement.ValidateQuantity()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestStockMovement_IsIncoming tests incoming movement check
func TestStockMovement_IsIncoming(t *testing.T) {
	tests := []struct {
		name     string
		movement models.StockMovement
		expected bool
	}{
		{"incoming movement", models.StockMovement{MovementType: models.MovementTypeIncoming}, true},
		{"outgoing movement", models.StockMovement{MovementType: models.MovementTypeOutgoing}, false},
		{"sale movement", models.StockMovement{MovementType: models.MovementTypeSale}, false},
		{"adjustment movement", models.StockMovement{MovementType: models.MovementTypeAdjustment}, false},
		{"return movement", models.StockMovement{MovementType: models.MovementTypeReturn}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.movement.IsIncoming()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestStockMovement_IsOutgoing tests outgoing movement check
func TestStockMovement_IsOutgoing(t *testing.T) {
	tests := []struct {
		name     string
		movement models.StockMovement
		expected bool
	}{
		{"incoming movement", models.StockMovement{MovementType: models.MovementTypeIncoming}, false},
		{"outgoing movement", models.StockMovement{MovementType: models.MovementTypeOutgoing}, true},
		{"sale movement", models.StockMovement{MovementType: models.MovementTypeSale}, true},
		{"adjustment movement", models.StockMovement{MovementType: models.MovementTypeAdjustment}, false},
		{"return movement", models.StockMovement{MovementType: models.MovementTypeReturn}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.movement.IsOutgoing()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestStockMovement_IsAdjustment tests adjustment movement check
func TestStockMovement_IsAdjustment(t *testing.T) {
	tests := []struct {
		name     string
		movement models.StockMovement
		expected bool
	}{
		{"incoming movement", models.StockMovement{MovementType: models.MovementTypeIncoming}, false},
		{"outgoing movement", models.StockMovement{MovementType: models.MovementTypeOutgoing}, false},
		{"sale movement", models.StockMovement{MovementType: models.MovementTypeSale}, false},
		{"adjustment movement", models.StockMovement{MovementType: models.MovementTypeAdjustment}, true},
		{"return movement", models.StockMovement{MovementType: models.MovementTypeReturn}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.movement.IsAdjustment()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestStockMovement_GetAbsoluteQuantity tests absolute quantity calculation
func TestStockMovement_GetAbsoluteQuantity(t *testing.T) {
	tests := []struct {
		name     string
		movement models.StockMovement
		expected int
	}{
		{"positive quantity", models.StockMovement{Quantity: 10}, 10},
		{"negative quantity", models.StockMovement{Quantity: -5}, 5},
		{"zero quantity", models.StockMovement{Quantity: 0}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.movement.GetAbsoluteQuantity()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestStockMovement_IsSale tests sale movement check
func TestStockMovement_IsSale(t *testing.T) {
	tests := []struct {
		name     string
		movement models.StockMovement
		expected bool
	}{
		{"incoming movement", models.StockMovement{MovementType: models.MovementTypeIncoming}, false},
		{"outgoing movement", models.StockMovement{MovementType: models.MovementTypeOutgoing}, false},
		{"sale movement", models.StockMovement{MovementType: models.MovementTypeSale}, true},
		{"adjustment movement", models.StockMovement{MovementType: models.MovementTypeAdjustment}, false},
		{"return movement", models.StockMovement{MovementType: models.MovementTypeReturn}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.movement.IsSale()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestStockMovement_GetQuantityChange tests quantity change calculation
func TestStockMovement_GetQuantityChange(t *testing.T) {
	tests := []struct {
		name     string
		movement models.StockMovement
		expected int
	}{
		{"positive quantity", models.StockMovement{Quantity: 10}, 10},
		{"negative quantity", models.StockMovement{Quantity: -5}, -5},
		{"zero quantity", models.StockMovement{Quantity: 0}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.movement.GetQuantityChange()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestStockMovement_ToResponse tests conversion to StockMovementResponse
func TestStockMovement_ToResponse(t *testing.T) {
	now := time.Now()
	reason := "Restocking"
	reference := "PO-12345"
	notes := "Received from supplier"

	movement := models.StockMovement{
		ID:           1,
		ProductID:    1,
		UserID:       1,
		MovementType: models.MovementTypeIncoming,
		Quantity:     10,
		Reason:       &reason,
		Reference:    &reference,
		Notes:        &notes,
		CreatedAt:    now,
	}

	response := movement.ToResponse()

	assert.Equal(t, movement.ID, response.ID)
	assert.Equal(t, movement.ProductID, response.ProductID)
	assert.Equal(t, movement.UserID, response.UserID)
	assert.Equal(t, movement.MovementType, response.MovementType)
	assert.Equal(t, movement.Quantity, response.Quantity)
	assert.Equal(t, movement.Reason, response.Reason)
	assert.Equal(t, movement.Reference, response.Reference)
	assert.Equal(t, movement.Notes, response.Notes)
	assert.Equal(t, movement.CreatedAt, response.CreatedAt)
}

// TestStockMovement_BeforeCreate tests the GORM BeforeCreate hook
func TestStockMovement_BeforeCreate(t *testing.T) {
	t.Run("valid movement creation", func(t *testing.T) {
		movement := models.StockMovement{
			ProductID:    1,
			UserID:       1,
			MovementType: models.MovementTypeIncoming,
			Quantity:     10,
		}

		err := movement.BeforeCreate(nil)
		assert.NoError(t, err)
	})

	t.Run("invalid movement creation", func(t *testing.T) {
		movement := models.StockMovement{
			ProductID:    0, // Invalid product ID
			UserID:       1,
			MovementType: models.MovementTypeIncoming,
			Quantity:     10,
		}

		err := movement.BeforeCreate(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "productId is required")
	})
}

// TestStockMovement_BeforeUpdate tests the GORM BeforeUpdate hook
func TestStockMovement_BeforeUpdate(t *testing.T) {
	t.Run("valid movement update", func(t *testing.T) {
		movement := models.StockMovement{
			ProductID:    1,
			UserID:       1,
			MovementType: models.MovementTypeIncoming,
			Quantity:     10,
		}

		err := movement.BeforeUpdate(nil)
		assert.NoError(t, err)
	})

	t.Run("invalid movement update", func(t *testing.T) {
		movement := models.StockMovement{
			ProductID:    0, // Invalid product ID
			UserID:       1,
			MovementType: models.MovementTypeIncoming,
			Quantity:     10,
		}

		err := movement.BeforeUpdate(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "productId is required")
	})
}
