// Package unit contains unit tests for the TT Stock Backend API models.
// It tests model validation, business logic, and data transformation methods.
package unit

import (
	"testing"
	"time"

	"tt-stock-api/src/models"

	"github.com/stretchr/testify/assert"
)

// TestProduct_Validate tests the Product validation functionality
func TestProduct_Validate(t *testing.T) {
	tests := []struct {
		name    string
		product models.Product
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid tire product",
			product: models.Product{
				Type:              models.ProductTypeTire,
				Brand:             "Michelin",
				Model:             "Pilot Sport 4",
				SKU:               "MIC-PS4-225-45-17",
				CostPrice:         150.00,
				SellingPrice:      200.00,
				QuantityOnHand:    10,
				LowStockThreshold: 5,
			},
			wantErr: false,
		},
		{
			name: "valid wheel product",
			product: models.Product{
				Type:              models.ProductTypeWheel,
				Brand:             "BBS",
				Model:             "CH-R",
				SKU:               "BBS-CHR-18X8",
				CostPrice:         300.00,
				SellingPrice:      450.00,
				QuantityOnHand:    5,
				LowStockThreshold: 2,
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			product: models.Product{
				Type:              "InvalidType",
				Brand:             "Michelin",
				Model:             "Pilot Sport 4",
				SKU:               "MIC-PS4-225-45-17",
				CostPrice:         150.00,
				SellingPrice:      200.00,
				QuantityOnHand:    10,
				LowStockThreshold: 5,
			},
			wantErr: true,
			errMsg:  "type must be one of: Tire, Wheel",
		},
		{
			name: "empty brand",
			product: models.Product{
				Type:              models.ProductTypeTire,
				Brand:             "",
				Model:             "Pilot Sport 4",
				SKU:               "MIC-PS4-225-45-17",
				CostPrice:         150.00,
				SellingPrice:      200.00,
				QuantityOnHand:    10,
				LowStockThreshold: 5,
			},
			wantErr: true,
			errMsg:  "brand is required",
		},
		{
			name: "brand too long",
			product: models.Product{
				Type:              models.ProductTypeTire,
				Brand:             "This is a very long brand name that exceeds the maximum allowed length of 100 characters and should fail validation because it is too long",
				Model:             "Pilot Sport 4",
				SKU:               "MIC-PS4-225-45-17",
				CostPrice:         150.00,
				SellingPrice:      200.00,
				QuantityOnHand:    10,
				LowStockThreshold: 5,
			},
			wantErr: true,
			errMsg:  "brand must be 1-100 characters",
		},
		{
			name: "empty model",
			product: models.Product{
				Type:              models.ProductTypeTire,
				Brand:             "Michelin",
				Model:             "",
				SKU:               "MIC-PS4-225-45-17",
				CostPrice:         150.00,
				SellingPrice:      200.00,
				QuantityOnHand:    10,
				LowStockThreshold: 5,
			},
			wantErr: true,
			errMsg:  "model is required",
		},
		{
			name: "model too long",
			product: models.Product{
				Type:              models.ProductTypeTire,
				Brand:             "Michelin",
				Model:             "This is a very long model name that exceeds the maximum allowed length of 100 characters and should fail validation because it is too long",
				SKU:               "MIC-PS4-225-45-17",
				CostPrice:         150.00,
				SellingPrice:      200.00,
				QuantityOnHand:    10,
				LowStockThreshold: 5,
			},
			wantErr: true,
			errMsg:  "model must be 1-100 characters",
		},
		{
			name: "empty SKU",
			product: models.Product{
				Type:              models.ProductTypeTire,
				Brand:             "Michelin",
				Model:             "Pilot Sport 4",
				SKU:               "",
				CostPrice:         150.00,
				SellingPrice:      200.00,
				QuantityOnHand:    10,
				LowStockThreshold: 5,
			},
			wantErr: true,
			errMsg:  "SKU is required",
		},
		{
			name: "SKU too short",
			product: models.Product{
				Type:              models.ProductTypeTire,
				Brand:             "Michelin",
				Model:             "Pilot Sport 4",
				SKU:               "AB",
				CostPrice:         150.00,
				SellingPrice:      200.00,
				QuantityOnHand:    10,
				LowStockThreshold: 5,
			},
			wantErr: true,
			errMsg:  "SKU must be 3-50 characters",
		},
		{
			name: "SKU too long",
			product: models.Product{
				Type:              models.ProductTypeTire,
				Brand:             "Michelin",
				Model:             "Pilot Sport 4",
				SKU:               "This is a very long SKU that exceeds the maximum allowed length of 50 characters and should fail validation",
				CostPrice:         150.00,
				SellingPrice:      200.00,
				QuantityOnHand:    10,
				LowStockThreshold: 5,
			},
			wantErr: true,
			errMsg:  "SKU must be 3-50 characters",
		},
		{
			name: "negative cost price",
			product: models.Product{
				Type:              models.ProductTypeTire,
				Brand:             "Michelin",
				Model:             "Pilot Sport 4",
				SKU:               "MIC-PS4-225-45-17",
				CostPrice:         -150.00,
				SellingPrice:      200.00,
				QuantityOnHand:    10,
				LowStockThreshold: 5,
			},
			wantErr: true,
			errMsg:  "cost price must be positive",
		},
		{
			name: "zero cost price",
			product: models.Product{
				Type:              models.ProductTypeTire,
				Brand:             "Michelin",
				Model:             "Pilot Sport 4",
				SKU:               "MIC-PS4-225-45-17",
				CostPrice:         0.00,
				SellingPrice:      200.00,
				QuantityOnHand:    10,
				LowStockThreshold: 5,
			},
			wantErr: true,
			errMsg:  "cost price must be positive",
		},
		{
			name: "negative selling price",
			product: models.Product{
				Type:              models.ProductTypeTire,
				Brand:             "Michelin",
				Model:             "Pilot Sport 4",
				SKU:               "MIC-PS4-225-45-17",
				CostPrice:         150.00,
				SellingPrice:      -200.00,
				QuantityOnHand:    10,
				LowStockThreshold: 5,
			},
			wantErr: true,
			errMsg:  "selling price must be positive",
		},
		{
			name: "zero selling price",
			product: models.Product{
				Type:              models.ProductTypeTire,
				Brand:             "Michelin",
				Model:             "Pilot Sport 4",
				SKU:               "MIC-PS4-225-45-17",
				CostPrice:         150.00,
				SellingPrice:      0.00,
				QuantityOnHand:    10,
				LowStockThreshold: 5,
			},
			wantErr: true,
			errMsg:  "selling price must be positive",
		},
		{
			name: "negative quantity",
			product: models.Product{
				Type:              models.ProductTypeTire,
				Brand:             "Michelin",
				Model:             "Pilot Sport 4",
				SKU:               "MIC-PS4-225-45-17",
				CostPrice:         150.00,
				SellingPrice:      200.00,
				QuantityOnHand:    -10,
				LowStockThreshold: 5,
			},
			wantErr: true,
			errMsg:  "quantity on hand cannot be negative",
		},
		{
			name: "negative low stock threshold",
			product: models.Product{
				Type:              models.ProductTypeTire,
				Brand:             "Michelin",
				Model:             "Pilot Sport 4",
				SKU:               "MIC-PS4-225-45-17",
				CostPrice:         150.00,
				SellingPrice:      200.00,
				QuantityOnHand:    10,
				LowStockThreshold: -5,
			},
			wantErr: true,
			errMsg:  "low stock threshold cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.product.Validate()
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

// TestProduct_ValidateType tests product type validation
func TestProduct_ValidateType(t *testing.T) {
	tests := []struct {
		name    string
		product models.Product
		wantErr bool
	}{
		{"valid tire", models.Product{Type: models.ProductTypeTire}, false},
		{"valid wheel", models.Product{Type: models.ProductTypeWheel}, false},
		{"invalid type", models.Product{Type: "InvalidType"}, true},
		{"empty type", models.Product{Type: ""}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.product.ValidateType()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "type must be one of: Tire, Wheel")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestProduct_ValidateBrand tests brand validation
func TestProduct_ValidateBrand(t *testing.T) {
	tests := []struct {
		name    string
		product models.Product
		wantErr bool
		errMsg  string
	}{
		{"valid brand", models.Product{Brand: "Michelin"}, false, ""},
		{"valid short brand", models.Product{Brand: "A"}, false, ""},
		{"valid long brand", models.Product{Brand: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}, false, ""},
		{"empty brand", models.Product{Brand: ""}, true, "brand is required"},
		{"brand too long", models.Product{Brand: "This is a very long brand name that exceeds the maximum allowed length of 100 characters and should fail validation because it is too long"}, true, "brand must be 1-100 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.product.ValidateBrand()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestProduct_ValidateModel tests model validation
func TestProduct_ValidateModel(t *testing.T) {
	tests := []struct {
		name    string
		product models.Product
		wantErr bool
		errMsg  string
	}{
		{"valid model", models.Product{Model: "Pilot Sport 4"}, false, ""},
		{"valid short model", models.Product{Model: "A"}, false, ""},
		{"valid long model", models.Product{Model: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}, false, ""},
		{"empty model", models.Product{Model: ""}, true, "model is required"},
		{"model too long", models.Product{Model: "This is a very long model name that exceeds the maximum allowed length of 100 characters and should fail validation because it is too long"}, true, "model must be 1-100 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.product.ValidateModel()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestProduct_ValidateSKU tests SKU validation
func TestProduct_ValidateSKU(t *testing.T) {
	tests := []struct {
		name    string
		product models.Product
		wantErr bool
		errMsg  string
	}{
		{"valid SKU", models.Product{SKU: "MIC-PS4-225-45-17"}, false, ""},
		{"valid short SKU", models.Product{SKU: "ABC"}, false, ""},
		{"valid long SKU", models.Product{SKU: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}, false, ""},
		{"empty SKU", models.Product{SKU: ""}, true, "SKU is required"},
		{"SKU too short", models.Product{SKU: "AB"}, true, "SKU must be 3-50 characters"},
		{"SKU too long", models.Product{SKU: "This is a very long SKU that exceeds the maximum allowed length of 50 characters and should fail validation"}, true, "SKU must be 3-50 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.product.ValidateSKU()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestProduct_ValidatePrices tests price validation
func TestProduct_ValidatePrices(t *testing.T) {
	tests := []struct {
		name    string
		product models.Product
		wantErr bool
		errMsg  string
	}{
		{"valid prices", models.Product{CostPrice: 150.00, SellingPrice: 200.00}, false, ""},
		{"zero cost price", models.Product{CostPrice: 0.00, SellingPrice: 200.00}, true, "cost price must be positive"},
		{"negative cost price", models.Product{CostPrice: -150.00, SellingPrice: 200.00}, true, "cost price must be positive"},
		{"zero selling price", models.Product{CostPrice: 150.00, SellingPrice: 0.00}, true, "selling price must be positive"},
		{"negative selling price", models.Product{CostPrice: 150.00, SellingPrice: -200.00}, true, "selling price must be positive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.product.ValidatePrices()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestProduct_ValidateQuantity tests quantity validation
func TestProduct_ValidateQuantity(t *testing.T) {
	tests := []struct {
		name    string
		product models.Product
		wantErr bool
		errMsg  string
	}{
		{"valid quantities", models.Product{QuantityOnHand: 10, LowStockThreshold: 5}, false, ""},
		{"zero quantities", models.Product{QuantityOnHand: 0, LowStockThreshold: 0}, false, ""},
		{"negative quantity", models.Product{QuantityOnHand: -10, LowStockThreshold: 5}, true, "quantity on hand cannot be negative"},
		{"negative threshold", models.Product{QuantityOnHand: 10, LowStockThreshold: -5}, true, "low stock threshold cannot be negative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.product.ValidateQuantity()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestProduct_GetStockStatus tests stock status calculation
func TestProduct_GetStockStatus(t *testing.T) {
	tests := []struct {
		name     string
		product  models.Product
		expected models.StockStatus
	}{
		{"out of stock", models.Product{QuantityOnHand: 0, LowStockThreshold: 5}, models.StockStatusOutOfStock},
		{"low stock", models.Product{QuantityOnHand: 3, LowStockThreshold: 5}, models.StockStatusLowStock},
		{"low stock at threshold", models.Product{QuantityOnHand: 5, LowStockThreshold: 5}, models.StockStatusLowStock},
		{"available", models.Product{QuantityOnHand: 10, LowStockThreshold: 5}, models.StockStatusAvailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.product.GetStockStatus()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestProduct_IsLowStock tests low stock check
func TestProduct_IsLowStock(t *testing.T) {
	tests := []struct {
		name     string
		product  models.Product
		expected bool
	}{
		{"out of stock", models.Product{QuantityOnHand: 0, LowStockThreshold: 5}, false},
		{"low stock", models.Product{QuantityOnHand: 3, LowStockThreshold: 5}, true},
		{"low stock at threshold", models.Product{QuantityOnHand: 5, LowStockThreshold: 5}, true},
		{"available", models.Product{QuantityOnHand: 10, LowStockThreshold: 5}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.product.IsLowStock()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestProduct_IsOutOfStock tests out of stock check
func TestProduct_IsOutOfStock(t *testing.T) {
	tests := []struct {
		name     string
		product  models.Product
		expected bool
	}{
		{"out of stock", models.Product{QuantityOnHand: 0}, true},
		{"low stock", models.Product{QuantityOnHand: 3}, false},
		{"available", models.Product{QuantityOnHand: 10}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.product.IsOutOfStock()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestProduct_IsAvailable tests availability check
func TestProduct_IsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		product  models.Product
		expected bool
	}{
		{"out of stock", models.Product{QuantityOnHand: 0, LowStockThreshold: 5}, false},
		{"low stock", models.Product{QuantityOnHand: 3, LowStockThreshold: 5}, false},
		{"low stock at threshold", models.Product{QuantityOnHand: 5, LowStockThreshold: 5}, false},
		{"available", models.Product{QuantityOnHand: 10, LowStockThreshold: 5}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.product.IsAvailable()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestProduct_CanSell tests sellability check
func TestProduct_CanSell(t *testing.T) {
	tests := []struct {
		name     string
		product  models.Product
		quantity int
		expected bool
	}{
		{"can sell available product", models.Product{IsActive: true, QuantityOnHand: 10}, 5, true},
		{"cannot sell inactive product", models.Product{IsActive: false, QuantityOnHand: 10}, 5, false},
		{"cannot sell insufficient stock", models.Product{IsActive: true, QuantityOnHand: 3}, 5, false},
		{"can sell exact stock", models.Product{IsActive: true, QuantityOnHand: 5}, 5, true},
		{"cannot sell zero quantity", models.Product{IsActive: true, QuantityOnHand: 10}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.product.CanSell(tt.quantity)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestProduct_UpdateQuantity tests quantity update
func TestProduct_UpdateQuantity(t *testing.T) {
	tests := []struct {
		name        string
		product     models.Product
		quantity    int
		expectedErr bool
		expectedQty int
	}{
		{"valid quantity", models.Product{QuantityOnHand: 10}, 15, false, 15},
		{"zero quantity", models.Product{QuantityOnHand: 10}, 0, false, 0},
		{"negative quantity", models.Product{QuantityOnHand: 10}, -5, true, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.product.UpdateQuantity(tt.quantity)
			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "quantity cannot be negative")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedQty, tt.product.QuantityOnHand)
			}
		})
	}
}

// TestProduct_AdjustQuantity tests quantity adjustment
func TestProduct_AdjustQuantity(t *testing.T) {
	tests := []struct {
		name        string
		product     models.Product
		adjustment  int
		expectedErr bool
		expectedQty int
	}{
		{"positive adjustment", models.Product{QuantityOnHand: 10}, 5, false, 15},
		{"negative adjustment", models.Product{QuantityOnHand: 10}, -3, false, 7},
		{"zero adjustment", models.Product{QuantityOnHand: 10}, 0, false, 10},
		{"insufficient stock", models.Product{QuantityOnHand: 3}, -5, true, 3},
		{"exact stock", models.Product{QuantityOnHand: 5}, -5, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.product.AdjustQuantity(tt.adjustment)
			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "insufficient stock for this operation")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedQty, tt.product.QuantityOnHand)
			}
		})
	}
}

// TestProduct_GetTotalValue tests total value calculation
func TestProduct_GetTotalValue(t *testing.T) {
	tests := []struct {
		name     string
		product  models.Product
		expected float64
	}{
		{"normal values", models.Product{QuantityOnHand: 10, CostPrice: 150.00}, 1500.00},
		{"zero quantity", models.Product{QuantityOnHand: 0, CostPrice: 150.00}, 0.00},
		{"zero cost", models.Product{QuantityOnHand: 10, CostPrice: 0.00}, 0.00},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.product.GetTotalValue()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestProduct_GetProfitMargin tests profit margin calculation
func TestProduct_GetProfitMargin(t *testing.T) {
	tests := []struct {
		name     string
		product  models.Product
		expected float64
	}{
		{"normal profit", models.Product{CostPrice: 100.00, SellingPrice: 150.00}, 50.00},
		{"no profit", models.Product{CostPrice: 100.00, SellingPrice: 100.00}, 0.00},
		{"loss", models.Product{CostPrice: 100.00, SellingPrice: 80.00}, -20.00},
		{"zero cost", models.Product{CostPrice: 0.00, SellingPrice: 100.00}, 0.00},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.product.GetProfitMargin()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestProduct_ToResponse tests conversion to ProductResponse
func TestProduct_ToResponse(t *testing.T) {
	now := time.Now()
	description := "High-performance summer tire"
	imageBase64 := "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQABAAD..."

	product := models.Product{
		ID:                1,
		Type:              models.ProductTypeTire,
		Brand:             "Michelin",
		Model:             "Pilot Sport 4",
		SKU:               "MIC-PS4-225-45-17",
		Description:       &description,
		ImageBase64:       &imageBase64,
		CostPrice:         150.00,
		SellingPrice:      200.00,
		QuantityOnHand:    10,
		LowStockThreshold: 5,
		IsActive:          true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	response := product.ToResponse()

	assert.Equal(t, product.ID, response.ID)
	assert.Equal(t, product.Type, response.Type)
	assert.Equal(t, product.Brand, response.Brand)
	assert.Equal(t, product.Model, response.Model)
	assert.Equal(t, product.SKU, response.SKU)
	assert.Equal(t, product.Description, response.Description)
	assert.Equal(t, product.ImageBase64, response.ImageBase64)
	assert.Equal(t, product.CostPrice, response.CostPrice)
	assert.Equal(t, product.SellingPrice, response.SellingPrice)
	assert.Equal(t, product.QuantityOnHand, response.QuantityOnHand)
	assert.Equal(t, product.LowStockThreshold, response.LowStockThreshold)
	assert.Equal(t, product.IsActive, response.IsActive)
	assert.Equal(t, product.CreatedAt, response.CreatedAt)
	assert.Equal(t, product.UpdatedAt, response.UpdatedAt)
	assert.Equal(t, models.StockStatusAvailable, response.StockStatus)
}

// TestProduct_BeforeCreate tests the GORM BeforeCreate hook
func TestProduct_BeforeCreate(t *testing.T) {
	t.Run("valid product creation", func(t *testing.T) {
		product := models.Product{
			Type:              models.ProductTypeTire,
			Brand:             "Michelin",
			Model:             "Pilot Sport 4",
			SKU:               "MIC-PS4-225-45-17",
			CostPrice:         150.00,
			SellingPrice:      200.00,
			QuantityOnHand:    10,
			LowStockThreshold: 5,
		}

		err := product.BeforeCreate(nil)
		assert.NoError(t, err)
		assert.Equal(t, 10, product.QuantityOnHand)
		assert.Equal(t, 5, product.LowStockThreshold)
	})

	t.Run("invalid product creation", func(t *testing.T) {
		product := models.Product{
			Type:              "InvalidType",
			Brand:             "Michelin",
			Model:             "Pilot Sport 4",
			SKU:               "MIC-PS4-225-45-17",
			CostPrice:         150.00,
			SellingPrice:      200.00,
			QuantityOnHand:    10,
			LowStockThreshold: 5,
		}

		err := product.BeforeCreate(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "type must be one of: Tire, Wheel")
	})
}

// TestProduct_BeforeUpdate tests the GORM BeforeUpdate hook
func TestProduct_BeforeUpdate(t *testing.T) {
	t.Run("valid product update", func(t *testing.T) {
		product := models.Product{
			Type:              models.ProductTypeTire,
			Brand:             "Michelin",
			Model:             "Pilot Sport 4",
			SKU:               "MIC-PS4-225-45-17",
			CostPrice:         150.00,
			SellingPrice:      200.00,
			QuantityOnHand:    10,
			LowStockThreshold: 5,
		}

		err := product.BeforeUpdate(nil)
		assert.NoError(t, err)
	})

	t.Run("invalid product update", func(t *testing.T) {
		product := models.Product{
			Type:              "InvalidType",
			Brand:             "Michelin",
			Model:             "Pilot Sport 4",
			SKU:               "MIC-PS4-225-45-17",
			CostPrice:         150.00,
			SellingPrice:      200.00,
			QuantityOnHand:    10,
			LowStockThreshold: 5,
		}

		err := product.BeforeUpdate(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "type must be one of: Tire, Wheel")
	})
}
