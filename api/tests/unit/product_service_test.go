package unit

import (
	"context"
	"errors"
	"testing"

	"tt-stock-api/src/models"
	"tt-stock-api/src/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProductRepository is a mock implementation of ProductRepository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) GetByID(ctx context.Context, id uint) (*models.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) GetBySKU(ctx context.Context, sku string) (*models.Product, error) {
	args := m.Called(ctx, sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) Update(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductRepository) List(ctx context.Context, req *models.ProductSearchRequest) (*models.ProductSearchResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductSearchResponse), args.Error(1)
}

func (m *MockProductRepository) GetActiveProducts(ctx context.Context) ([]models.Product, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductRepository) GetProductsByType(ctx context.Context, productType models.ProductType) ([]models.Product, error) {
	args := m.Called(ctx, productType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductRepository) GetProductsByBrand(ctx context.Context, brand string) ([]models.Product, error) {
	args := m.Called(ctx, brand)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductRepository) GetLowStockProducts(ctx context.Context) ([]models.Product, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductRepository) GetOutOfStockProducts(ctx context.Context) ([]models.Product, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductRepository) UpdateQuantity(ctx context.Context, productID uint, quantity int) error {
	args := m.Called(ctx, productID, quantity)
	return args.Error(0)
}

func (m *MockProductRepository) AdjustQuantity(ctx context.Context, productID uint, adjustment int) error {
	args := m.Called(ctx, productID, adjustment)
	return args.Error(0)
}

func (m *MockProductRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProductRepository) Exists(ctx context.Context, sku string) (bool, error) {
	args := m.Called(ctx, sku)
	return args.Bool(0), args.Error(1)
}

func (m *MockProductRepository) GetProductsWithSpecifications(ctx context.Context, productIDs []uint) ([]models.Product, error) {
	args := m.Called(ctx, productIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Product), args.Error(1)
}

// Helper function to create test users with different roles
func createTestUser(role models.UserRole) *models.User {
	return &models.User{
		ID:       1,
		Role:     role,
		IsActive: true,
	}
}

func TestProductService_SearchProducts(t *testing.T) {
	tests := []struct {
		name          string
		request       *models.ProductSearchRequest
		user          *models.User
		setupMocks    func(*MockProductRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name: "successful search with filters",
			request: &models.ProductSearchRequest{
				Query:    stringPtr("Michelin"),
				Type:     productTypePtr(models.ProductTypeTire),
				Brand:    stringPtr("Michelin"),
				MinPrice: floatPtr(100.0),
				MaxPrice: floatPtr(500.0),
				Page:     1,
				Limit:    10,
			},
			user: createTestUser(models.UserRoleAdmin),
			setupMocks: func(productRepo *MockProductRepository) {
				response := &models.ProductSearchResponse{
					Products: []models.ProductResponse{
						{
							ID:           1,
							Type:         models.ProductTypeTire,
							Brand:        "Michelin",
							Model:        "Pilot Sport 4",
							SKU:          "MICH-PS4-225-45-17",
							CostPrice:    150.0,
							SellingPrice: 200.0,
							IsActive:     true,
						},
					},
					Pagination: models.PaginationResponse{
						Page:       1,
						Limit:      10,
						Total:      1,
						TotalPages: 1,
						HasNext:    false,
						HasPrev:    false,
					},
				}
				productRepo.On("List", mock.Anything, mock.AnythingOfType("*models.ProductSearchRequest")).Return(response, nil)
			},
			expectSuccess: true,
		},
		{
			name:    "nil search request",
			request: nil,
			user:    createTestUser(models.UserRoleAdmin),
			setupMocks: func(productRepo *MockProductRepository) {
				// No mocks needed
			},
			expectedError: "search request cannot be nil",
		},
		{
			name: "insufficient permissions",
			request: &models.ProductSearchRequest{
				Page:  1,
				Limit: 10,
			},
			user: nil, // No user provided
			setupMocks: func(productRepo *MockProductRepository) {
				// No mocks needed
			},
			expectedError: "insufficient permissions to search products",
		},
		{
			name: "repository error",
			request: &models.ProductSearchRequest{
				Query: stringPtr("test"),
				Page:  1,
				Limit: 10,
			},
			user: createTestUser(models.UserRoleAdmin),
			setupMocks: func(productRepo *MockProductRepository) {
				productRepo.On("List", mock.Anything, mock.AnythingOfType("*models.ProductSearchRequest")).Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock
			productRepo := &MockProductRepository{}

			// Setup mocks
			tt.setupMocks(productRepo)

			// Create service
			productService := services.NewProductService(productRepo)

			// Execute
			result, err := productService.SearchProducts(context.Background(), tt.request, tt.user)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.Products)
			}

			// Verify all expectations
			productRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_ListProducts(t *testing.T) {
	tests := []struct {
		name          string
		request       *models.ProductSearchRequest
		user          *models.User
		setupMocks    func(*MockProductRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name: "successful list",
			request: &models.ProductSearchRequest{
				Page:  1,
				Limit: 10,
			},
			user: createTestUser(models.UserRoleAdmin),
			setupMocks: func(productRepo *MockProductRepository) {
				response := &models.ProductSearchResponse{
					Products: []models.ProductResponse{
						{
							ID:           1,
							Type:         models.ProductTypeTire,
							Brand:        "Michelin",
							Model:        "Pilot Sport 4",
							SKU:          "MICH-PS4-225-45-17",
							CostPrice:    150.0,
							SellingPrice: 200.0,
							IsActive:     true,
						},
					},
					Pagination: models.PaginationResponse{
						Page:       1,
						Limit:      10,
						Total:      1,
						TotalPages: 1,
						HasNext:    false,
						HasPrev:    false,
					},
				}
				productRepo.On("List", mock.Anything, mock.AnythingOfType("*models.ProductSearchRequest")).Return(response, nil)
			},
			expectSuccess: true,
		},
		{
			name:    "nil list request",
			request: nil,
			user:    createTestUser(models.UserRoleAdmin),
			setupMocks: func(productRepo *MockProductRepository) {
				// No mocks needed
			},
			expectedError: "list request cannot be nil",
		},
		{
			name: "insufficient permissions",
			request: &models.ProductSearchRequest{
				Page:  1,
				Limit: 10,
			},
			user: nil, // No user provided
			setupMocks: func(productRepo *MockProductRepository) {
				// No mocks needed
			},
			expectedError: "insufficient permissions to list products",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock
			productRepo := &MockProductRepository{}

			// Setup mocks
			tt.setupMocks(productRepo)

			// Create service
			productService := services.NewProductService(productRepo)

			// Execute
			result, err := productService.ListProducts(context.Background(), tt.request, tt.user)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotNil(t, result.Products)
			}

			// Verify all expectations
			productRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_GetLowStockProducts(t *testing.T) {
	tests := []struct {
		name          string
		user          *models.User
		setupMocks    func(*MockProductRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name: "successful get low stock products",
			user: createTestUser(models.UserRoleAdmin),
			setupMocks: func(productRepo *MockProductRepository) {
				products := []models.Product{
					{
						ID:                1,
						Type:              models.ProductTypeTire,
						Brand:             "Michelin",
						Model:             "Pilot Sport 4",
						SKU:               "MICH-PS4-225-45-17",
						CostPrice:         150.0,
						SellingPrice:      200.0,
						QuantityOnHand:    2,
						LowStockThreshold: 5,
						IsActive:          true,
					},
				}
				productRepo.On("GetLowStockProducts", mock.Anything).Return(products, nil)
			},
			expectSuccess: true,
		},
		{
			name: "insufficient permissions",
			user: nil, // No user provided
			setupMocks: func(productRepo *MockProductRepository) {
				// No mocks needed
			},
			expectedError: "insufficient permissions to view low stock products",
		},
		{
			name: "repository error",
			user: createTestUser(models.UserRoleAdmin),
			setupMocks: func(productRepo *MockProductRepository) {
				productRepo.On("GetLowStockProducts", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock
			productRepo := &MockProductRepository{}

			// Setup mocks
			tt.setupMocks(productRepo)

			// Create service
			productService := services.NewProductService(productRepo)

			// Execute
			result, err := productService.GetLowStockProducts(context.Background(), tt.user)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if len(result) > 0 {
					assert.Equal(t, "Michelin", result[0].Brand)
					assert.Equal(t, models.ProductTypeTire, result[0].Type)
				}
			}

			// Verify all expectations
			productRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_GetProductsByBrand(t *testing.T) {
	tests := []struct {
		name          string
		brand         string
		user          *models.User
		setupMocks    func(*MockProductRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name:  "successful get products by brand",
			brand: "Michelin",
			user:  createTestUser(models.UserRoleAdmin),
			setupMocks: func(productRepo *MockProductRepository) {
				products := []models.Product{
					{
						ID:           1,
						Type:         models.ProductTypeTire,
						Brand:        "Michelin",
						Model:        "Pilot Sport 4",
						SKU:          "MICH-PS4-225-45-17",
						CostPrice:    150.0,
						SellingPrice: 200.0,
						IsActive:     true,
					},
				}
				productRepo.On("GetProductsByBrand", mock.Anything, "Michelin").Return(products, nil)
			},
			expectSuccess: true,
		},
		{
			name:  "empty brand",
			brand: "",
			user:  createTestUser(models.UserRoleAdmin),
			setupMocks: func(productRepo *MockProductRepository) {
				// No mocks needed
			},
			expectedError: "brand cannot be empty",
		},
		{
			name:  "insufficient permissions",
			brand: "Michelin",
			user:  nil, // No user provided
			setupMocks: func(productRepo *MockProductRepository) {
				// No mocks needed
			},
			expectedError: "insufficient permissions to view products by brand",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock
			productRepo := &MockProductRepository{}

			// Setup mocks
			tt.setupMocks(productRepo)

			// Create service
			productService := services.NewProductService(productRepo)

			// Execute
			result, err := productService.GetProductsByBrand(context.Background(), tt.brand, tt.user)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if len(result) > 0 {
					assert.Equal(t, "Michelin", result[0].Brand)
				}
			}

			// Verify all expectations
			productRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_GetProductsByType(t *testing.T) {
	tests := []struct {
		name          string
		productType   models.ProductType
		user          *models.User
		setupMocks    func(*MockProductRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name:        "successful get products by type",
			productType: models.ProductTypeTire,
			user:        createTestUser(models.UserRoleAdmin),
			setupMocks: func(productRepo *MockProductRepository) {
				products := []models.Product{
					{
						ID:           1,
						Type:         models.ProductTypeTire,
						Brand:        "Michelin",
						Model:        "Pilot Sport 4",
						SKU:          "MICH-PS4-225-45-17",
						CostPrice:    150.0,
						SellingPrice: 200.0,
						IsActive:     true,
					},
				}
				productRepo.On("GetProductsByType", mock.Anything, models.ProductTypeTire).Return(products, nil)
			},
			expectSuccess: true,
		},
		{
			name:        "insufficient permissions",
			productType: models.ProductTypeTire,
			user:        nil, // No user provided
			setupMocks: func(productRepo *MockProductRepository) {
				// No mocks needed
			},
			expectedError: "insufficient permissions to view products by type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock
			productRepo := &MockProductRepository{}

			// Setup mocks
			tt.setupMocks(productRepo)

			// Create service
			productService := services.NewProductService(productRepo)

			// Execute
			result, err := productService.GetProductsByType(context.Background(), tt.productType, tt.user)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if len(result) > 0 {
					assert.Equal(t, models.ProductTypeTire, result[0].Type)
				}
			}

			// Verify all expectations
			productRepo.AssertExpectations(t)
		})
	}
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}

func boolPtr(b bool) *bool {
	return &b
}

func productTypePtr(pt models.ProductType) *models.ProductType {
	return &pt
}
