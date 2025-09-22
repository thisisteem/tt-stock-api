package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"tt-stock-api/src/models"
	"tt-stock-api/src/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStockMovementRepository is a mock implementation of StockMovementRepository
type MockStockMovementRepository struct {
	mock.Mock
}

func (m *MockStockMovementRepository) Create(ctx context.Context, movement *models.StockMovement) error {
	args := m.Called(ctx, movement)
	return args.Error(0)
}

func (m *MockStockMovementRepository) GetByID(ctx context.Context, id uint) (*models.StockMovement, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.StockMovement), args.Error(1)
}

func (m *MockStockMovementRepository) Update(ctx context.Context, movement *models.StockMovement) error {
	args := m.Called(ctx, movement)
	return args.Error(0)
}

func (m *MockStockMovementRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStockMovementRepository) List(ctx context.Context, req *models.StockMovementListRequest) (*models.StockMovementListResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.StockMovementListResponse), args.Error(1)
}

func (m *MockStockMovementRepository) GetMovementsByProduct(ctx context.Context, productID uint) ([]models.StockMovement, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.StockMovement), args.Error(1)
}

func (m *MockStockMovementRepository) GetMovementsByUser(ctx context.Context, userID uint) ([]models.StockMovement, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.StockMovement), args.Error(1)
}

func (m *MockStockMovementRepository) GetMovementSummary(ctx context.Context, productID *uint, startDate, endDate *time.Time) (*models.StockMovementSummary, error) {
	args := m.Called(ctx, productID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.StockMovementSummary), args.Error(1)
}

func (m *MockStockMovementRepository) GetRecentMovements(ctx context.Context, limit int) ([]models.StockMovement, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.StockMovement), args.Error(1)
}

func (m *MockStockMovementRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStockMovementRepository) GetTotalQuantityByProduct(ctx context.Context, productID uint) (int, error) {
	args := m.Called(ctx, productID)
	return args.Int(0), args.Error(1)
}

func (m *MockStockMovementRepository) GetMovementsByType(ctx context.Context, movementType models.MovementType) ([]models.StockMovement, error) {
	args := m.Called(ctx, movementType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.StockMovement), args.Error(1)
}

func (m *MockStockMovementRepository) GetMovementsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]models.StockMovement, error) {
	args := m.Called(ctx, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.StockMovement), args.Error(1)
}

// MockAlertRepository is a mock implementation of AlertRepository
type MockAlertRepository struct {
	mock.Mock
}

func (m *MockAlertRepository) Create(ctx context.Context, alert *models.Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockAlertRepository) GetByID(ctx context.Context, id uint) (*models.Alert, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Alert), args.Error(1)
}

func (m *MockAlertRepository) Update(ctx context.Context, alert *models.Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockAlertRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAlertRepository) List(ctx context.Context, req *models.AlertListRequest) (*models.AlertListResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AlertListResponse), args.Error(1)
}

func (m *MockAlertRepository) GetActiveAlerts(ctx context.Context) ([]models.Alert, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Alert), args.Error(1)
}

func (m *MockAlertRepository) GetAlertsByUser(ctx context.Context, userID uint) ([]models.Alert, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Alert), args.Error(1)
}

func (m *MockAlertRepository) GetAlertsByProduct(ctx context.Context, productID uint) ([]models.Alert, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Alert), args.Error(1)
}

func (m *MockAlertRepository) MarkAsRead(ctx context.Context, alertID uint) error {
	args := m.Called(ctx, alertID)
	return args.Error(0)
}

func (m *MockAlertRepository) MarkAllAsRead(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAlertRepository) DeactivateAlert(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAlertRepository) GetUnreadCount(ctx context.Context, userID *uint) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAlertRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAlertRepository) CountActiveAlerts(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAlertRepository) GetUnreadAlerts(ctx context.Context) ([]models.Alert, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Alert), args.Error(1)
}

func (m *MockAlertRepository) GetAlertsByType(ctx context.Context, alertType models.AlertType) ([]models.Alert, error) {
	args := m.Called(ctx, alertType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Alert), args.Error(1)
}

func (m *MockAlertRepository) GetAlertsByPriority(ctx context.Context, priority models.AlertPriority) ([]models.Alert, error) {
	args := m.Called(ctx, priority)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Alert), args.Error(1)
}

func (m *MockAlertRepository) MarkAsUnread(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestStockService_CreateStockMovement(t *testing.T) {
	tests := []struct {
		name          string
		request       *models.StockMovementCreateRequest
		user          *models.User
		setupMocks    func(*MockStockMovementRepository, *MockProductRepository, *MockAlertRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name: "successful incoming movement",
			request: &models.StockMovementCreateRequest{
				ProductID:    1,
				MovementType: models.MovementTypeIncoming,
				Quantity:     10,
				Reason:       stringPtr("Supplier delivery"),
				Reference:    stringPtr("PO-001"),
			},
			user: createTestUser(models.UserRoleAdmin),
			setupMocks: func(stockRepo *MockStockMovementRepository, productRepo *MockProductRepository, alertRepo *MockAlertRepository) {
				product := &models.Product{
					ID:                1,
					SKU:               "TEST-001",
					QuantityOnHand:    5,
					LowStockThreshold: 3,
					IsActive:          true,
				}
				productRepo.On("GetByID", mock.Anything, uint(1)).Return(product, nil)
				stockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.StockMovement")).Return(nil)
				productRepo.On("AdjustQuantity", mock.Anything, uint(1), 10).Return(nil)
			},
			expectSuccess: true,
		},
		{
			name: "successful outgoing movement with sufficient stock",
			request: &models.StockMovementCreateRequest{
				ProductID:    1,
				MovementType: models.MovementTypeOutgoing,
				Quantity:     -3,
				Reason:       stringPtr("Sale"),
				Reference:    stringPtr("SALE-001"),
			},
			user: createTestUser(models.UserRoleAdmin),
			setupMocks: func(stockRepo *MockStockMovementRepository, productRepo *MockProductRepository, alertRepo *MockAlertRepository) {
				product := &models.Product{
					ID:                1,
					SKU:               "TEST-001",
					QuantityOnHand:    10,
					LowStockThreshold: 3,
					IsActive:          true,
				}
				productRepo.On("GetByID", mock.Anything, uint(1)).Return(product, nil)
				stockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.StockMovement")).Return(nil)
				productRepo.On("AdjustQuantity", mock.Anything, uint(1), -3).Return(nil)
			},
			expectSuccess: true,
		},
		{
			name:    "nil request",
			request: nil,
			user:    createTestUser(models.UserRoleAdmin),
			setupMocks: func(stockRepo *MockStockMovementRepository, productRepo *MockProductRepository, alertRepo *MockAlertRepository) {
				// No mocks needed
			},
			expectedError: "create request cannot be nil",
		},
		{
			name: "insufficient permissions",
			request: &models.StockMovementCreateRequest{
				ProductID:    1,
				MovementType: models.MovementTypeIncoming,
				Quantity:     10,
			},
			user: nil, // No user provided
			setupMocks: func(stockRepo *MockStockMovementRepository, productRepo *MockProductRepository, alertRepo *MockAlertRepository) {
				// No mocks needed
			},
			expectedError: "insufficient permissions to create stock movements",
		},
		{
			name: "product not found",
			request: &models.StockMovementCreateRequest{
				ProductID:    1,
				MovementType: models.MovementTypeIncoming,
				Quantity:     10,
			},
			user: createTestUser(models.UserRoleAdmin),
			setupMocks: func(stockRepo *MockStockMovementRepository, productRepo *MockProductRepository, alertRepo *MockAlertRepository) {
				productRepo.On("GetByID", mock.Anything, uint(1)).Return(nil, errors.New("product not found"))
			},
			expectedError: "product not found",
		},
		{
			name: "invalid movement - positive quantity for outgoing",
			request: &models.StockMovementCreateRequest{
				ProductID:    1,
				MovementType: models.MovementTypeOutgoing,
				Quantity:     5, // Should be negative
			},
			user: createTestUser(models.UserRoleAdmin),
			setupMocks: func(stockRepo *MockStockMovementRepository, productRepo *MockProductRepository, alertRepo *MockAlertRepository) {
				product := &models.Product{
					ID:                1,
					SKU:               "TEST-001",
					QuantityOnHand:    10,
					LowStockThreshold: 3,
					IsActive:          true,
				}
				productRepo.On("GetByID", mock.Anything, uint(1)).Return(product, nil)
			},
			expectedError: "outgoing and sale movements must have negative quantities",
		},
		{
			name: "insufficient stock for outgoing movement",
			request: &models.StockMovementCreateRequest{
				ProductID:    1,
				MovementType: models.MovementTypeOutgoing,
				Quantity:     -15, // More than available stock
			},
			user: createTestUser(models.UserRoleAdmin),
			setupMocks: func(stockRepo *MockStockMovementRepository, productRepo *MockProductRepository, alertRepo *MockAlertRepository) {
				product := &models.Product{
					ID:                1,
					SKU:               "TEST-001",
					QuantityOnHand:    10,
					LowStockThreshold: 3,
					IsActive:          true,
				}
				productRepo.On("GetByID", mock.Anything, uint(1)).Return(product, nil)
			},
			expectedError: "insufficient stock for this movement",
		},
		{
			name: "adjustment with zero quantity",
			request: &models.StockMovementCreateRequest{
				ProductID:    1,
				MovementType: models.MovementTypeAdjustment,
				Quantity:     0, // Zero adjustment not allowed
			},
			user: createTestUser(models.UserRoleAdmin),
			setupMocks: func(stockRepo *MockStockMovementRepository, productRepo *MockProductRepository, alertRepo *MockAlertRepository) {
				product := &models.Product{
					ID:                1,
					SKU:               "TEST-001",
					QuantityOnHand:    10,
					LowStockThreshold: 3,
					IsActive:          true,
				}
				productRepo.On("GetByID", mock.Anything, uint(1)).Return(product, nil)
			},
			expectedError: "adjustment quantity cannot be zero",
		},
		{
			name: "adjustment would result in negative stock",
			request: &models.StockMovementCreateRequest{
				ProductID:    1,
				MovementType: models.MovementTypeAdjustment,
				Quantity:     -15, // Would result in negative stock
			},
			user: createTestUser(models.UserRoleAdmin),
			setupMocks: func(stockRepo *MockStockMovementRepository, productRepo *MockProductRepository, alertRepo *MockAlertRepository) {
				product := &models.Product{
					ID:                1,
					SKU:               "TEST-001",
					QuantityOnHand:    10,
					LowStockThreshold: 3,
					IsActive:          true,
				}
				productRepo.On("GetByID", mock.Anything, uint(1)).Return(product, nil)
			},
			expectedError: "adjustment would result in negative stock",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			stockRepo := &MockStockMovementRepository{}
			productRepo := &MockProductRepository{}
			alertRepo := &MockAlertRepository{}

			// Setup mocks
			tt.setupMocks(stockRepo, productRepo, alertRepo)

			// Create service
			stockService := services.NewStockService(stockRepo, productRepo, alertRepo)

			// Execute
			result, err := stockService.CreateStockMovement(context.Background(), tt.request, tt.user)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.ProductID, result.ProductID)
				assert.Equal(t, tt.request.MovementType, result.MovementType)
				assert.Equal(t, tt.request.Quantity, result.Quantity)
			}

			// Verify all expectations
			stockRepo.AssertExpectations(t)
			productRepo.AssertExpectations(t)
			alertRepo.AssertExpectations(t)
		})
	}
}

func TestStockService_ProcessSale(t *testing.T) {
	tests := []struct {
		name          string
		request       *models.SaleRequest
		user          *models.User
		setupMocks    func(*MockStockMovementRepository, *MockProductRepository, *MockAlertRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name: "successful sale",
			request: &models.SaleRequest{
				ProductID:    1,
				Quantity:     2,
				CustomerName: stringPtr("John Doe"),
				Reference:    stringPtr("SALE-001"),
			},
			user: createTestUser(models.UserRoleAdmin),
			setupMocks: func(stockRepo *MockStockMovementRepository, productRepo *MockProductRepository, alertRepo *MockAlertRepository) {
				product := &models.Product{
					ID:                1,
					SKU:               "TEST-001",
					QuantityOnHand:    10,
					LowStockThreshold: 3,
					IsActive:          true,
				}
				updatedProduct := &models.Product{
					ID:                1,
					SKU:               "TEST-001",
					QuantityOnHand:    8,
					LowStockThreshold: 3,
					IsActive:          true,
				}
				productRepo.On("GetByID", mock.Anything, uint(1)).Return(product, nil).Once()
				stockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.StockMovement")).Return(nil)
				productRepo.On("AdjustQuantity", mock.Anything, uint(1), -2).Return(nil)
				productRepo.On("GetByID", mock.Anything, uint(1)).Return(updatedProduct, nil).Once()
			},
			expectSuccess: true,
		},
		{
			name:    "nil sale request",
			request: nil,
			user:    createTestUser(models.UserRoleAdmin),
			setupMocks: func(stockRepo *MockStockMovementRepository, productRepo *MockProductRepository, alertRepo *MockAlertRepository) {
				// No mocks needed
			},
			expectedError: "sale request cannot be nil",
		},
		{
			name: "insufficient permissions",
			request: &models.SaleRequest{
				ProductID: 1,
				Quantity:  2,
			},
			user: nil, // No user provided
			setupMocks: func(stockRepo *MockStockMovementRepository, productRepo *MockProductRepository, alertRepo *MockAlertRepository) {
				// No mocks needed
			},
			expectedError: "insufficient permissions to process sales",
		},
		{
			name: "product not found",
			request: &models.SaleRequest{
				ProductID: 1,
				Quantity:  2,
			},
			user: createTestUser(models.UserRoleAdmin),
			setupMocks: func(stockRepo *MockStockMovementRepository, productRepo *MockProductRepository, alertRepo *MockAlertRepository) {
				productRepo.On("GetByID", mock.Anything, uint(1)).Return(nil, errors.New("product not found"))
			},
			expectedError: "product not found",
		},
		{
			name: "insufficient stock for sale",
			request: &models.SaleRequest{
				ProductID: 1,
				Quantity:  15, // More than available
			},
			user: createTestUser(models.UserRoleAdmin),
			setupMocks: func(stockRepo *MockStockMovementRepository, productRepo *MockProductRepository, alertRepo *MockAlertRepository) {
				product := &models.Product{
					ID:                1,
					SKU:               "TEST-001",
					QuantityOnHand:    10,
					LowStockThreshold: 3,
					IsActive:          true,
				}
				productRepo.On("GetByID", mock.Anything, uint(1)).Return(product, nil)
			},
			expectedError: "insufficient stock for sale",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			stockRepo := &MockStockMovementRepository{}
			productRepo := &MockProductRepository{}
			alertRepo := &MockAlertRepository{}

			// Setup mocks
			tt.setupMocks(stockRepo, productRepo, alertRepo)

			// Create service
			stockService := services.NewStockService(stockRepo, productRepo, alertRepo)

			// Execute
			result, err := stockService.ProcessSale(context.Background(), tt.request, tt.user)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotNil(t, result.Movement)
				assert.Equal(t, models.MovementTypeSale, result.Movement.MovementType)
				assert.Equal(t, -tt.request.Quantity, result.Movement.Quantity) // Should be negative
			}

			// Verify all expectations
			stockRepo.AssertExpectations(t)
			productRepo.AssertExpectations(t)
			alertRepo.AssertExpectations(t)
		})
	}
}

func TestStockService_GetStockMovement(t *testing.T) {
	tests := []struct {
		name          string
		id            uint
		user          *models.User
		setupMocks    func(*MockStockMovementRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name: "successful get stock movement",
			id:   1,
			user: createTestUser(models.UserRoleAdmin),
			setupMocks: func(stockRepo *MockStockMovementRepository) {
				movement := &models.StockMovement{
					ID:           1,
					ProductID:    1,
					UserID:       1,
					MovementType: models.MovementTypeIncoming,
					Quantity:     10,
				}
				stockRepo.On("GetByID", mock.Anything, uint(1)).Return(movement, nil)
			},
			expectSuccess: true,
		},
		{
			name: "zero ID",
			id:   0,
			user: createTestUser(models.UserRoleAdmin),
			setupMocks: func(stockRepo *MockStockMovementRepository) {
				// No mocks needed
			},
			expectedError: "stock movement ID cannot be zero",
		},
		{
			name: "insufficient permissions",
			id:   1,
			user: nil, // No user provided
			setupMocks: func(stockRepo *MockStockMovementRepository) {
				// No mocks needed
			},
			expectedError: "insufficient permissions to view stock movements",
		},
		{
			name: "movement not found",
			id:   1,
			user: createTestUser(models.UserRoleAdmin),
			setupMocks: func(stockRepo *MockStockMovementRepository) {
				stockRepo.On("GetByID", mock.Anything, uint(1)).Return(nil, errors.New("movement not found"))
			},
			expectedError: "movement not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			stockRepo := &MockStockMovementRepository{}
			productRepo := &MockProductRepository{}
			alertRepo := &MockAlertRepository{}

			// Setup mocks
			tt.setupMocks(stockRepo)

			// Create service
			stockService := services.NewStockService(stockRepo, productRepo, alertRepo)

			// Execute
			result, err := stockService.GetStockMovement(context.Background(), tt.id, tt.user)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.id, result.ID)
			}

			// Verify all expectations
			stockRepo.AssertExpectations(t)
		})
	}
}

func TestStockService_ProcessIncomingStock(t *testing.T) {
	// Create mocks
	stockRepo := &MockStockMovementRepository{}
	productRepo := &MockProductRepository{}
	alertRepo := &MockAlertRepository{}

	// Setup mocks for successful incoming stock processing
	product := &models.Product{
		ID:                1,
		SKU:               "TEST-001",
		QuantityOnHand:    5,
		LowStockThreshold: 3,
		IsActive:          true,
	}
	productRepo.On("GetByID", mock.Anything, uint(1)).Return(product, nil)
	stockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.StockMovement")).Return(nil)
	productRepo.On("AdjustQuantity", mock.Anything, uint(1), 10).Return(nil)

	// Create service
	stockService := services.NewStockService(stockRepo, productRepo, alertRepo)

	// Test data
	request := &models.StockMovementCreateRequest{
		ProductID: 1,
		Quantity:  10,
		Reason:    stringPtr("Supplier delivery"),
		Reference: stringPtr("PO-001"),
	}
	user := createTestUser(models.UserRoleAdmin)

	// Execute
	result, err := stockService.ProcessIncomingStock(context.Background(), request, user)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, models.MovementTypeIncoming, result.MovementType)
	assert.Equal(t, request.ProductID, result.ProductID)
	assert.Equal(t, request.Quantity, result.Quantity)

	// Verify all expectations
	stockRepo.AssertExpectations(t)
	productRepo.AssertExpectations(t)
	alertRepo.AssertExpectations(t)
}

func TestStockService_ProcessStockAdjustment(t *testing.T) {
	// Create mocks
	stockRepo := &MockStockMovementRepository{}
	productRepo := &MockProductRepository{}
	alertRepo := &MockAlertRepository{}

	// Setup mocks for successful adjustment processing
	product := &models.Product{
		ID:                1,
		SKU:               "TEST-001",
		QuantityOnHand:    10,
		LowStockThreshold: 3,
		IsActive:          true,
	}
	productRepo.On("GetByID", mock.Anything, uint(1)).Return(product, nil)
	stockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.StockMovement")).Return(nil)
	productRepo.On("AdjustQuantity", mock.Anything, uint(1), -2).Return(nil)

	// Create service
	stockService := services.NewStockService(stockRepo, productRepo, alertRepo)

	// Test data
	request := &models.StockMovementCreateRequest{
		ProductID: 1,
		Quantity:  -2,
		Reason:    stringPtr("Inventory adjustment"),
		Reference: stringPtr("ADJ-001"),
	}
	user := createTestUser(models.UserRoleAdmin)

	// Execute
	result, err := stockService.ProcessStockAdjustment(context.Background(), request, user)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, models.MovementTypeAdjustment, result.MovementType)
	assert.Equal(t, request.ProductID, result.ProductID)
	assert.Equal(t, request.Quantity, result.Quantity)

	// Verify all expectations
	stockRepo.AssertExpectations(t)
	productRepo.AssertExpectations(t)
	alertRepo.AssertExpectations(t)
}
