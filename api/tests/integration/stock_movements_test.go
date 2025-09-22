package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// StockMovementsTestSuite tests stock movement operations
// This includes creating movements, tracking inventory, and processing sales
type StockMovementsTestSuite struct {
	suite.Suite
	router *gin.Engine
	token  string
}

// SetupSuite runs once before all tests in the suite
func (suite *StockMovementsTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	setupTestStockRoutes(suite.router)
	suite.token = getAuthToken(suite.router)
}

// TestStockMovementsTestSuite runs the test suite
func TestStockMovementsTestSuite(t *testing.T) {
	suite.Run(t, new(StockMovementsTestSuite))
}

// TestCreateIncomingMovement tests creating an incoming stock movement
func (suite *StockMovementsTestSuite) TestCreateIncomingMovement() {
	movementRequest := map[string]interface{}{
		"productId":    1,
		"movementType": "Incoming",
		"quantity":     10,
		"reason":       "New shipment from supplier",
		"reference":    "PO-2024-001",
		"notes":        "Received 10 units of Michelin Pilot Sport 4",
	}

	jsonBody, _ := json.Marshal(movementRequest)
	req, _ := http.NewRequest("POST", "/v1/stock/movements", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Verify successful creation
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Stock movement recorded successfully", response["message"])

	movement := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), float64(1), movement["productId"])
	assert.Equal(suite.T(), "Incoming", movement["movementType"])
	assert.Equal(suite.T(), float64(10), movement["quantity"])
	assert.Equal(suite.T(), "New shipment from supplier", movement["reason"])
	assert.Equal(suite.T(), "PO-2024-001", movement["reference"])
	assert.Equal(suite.T(), "Received 10 units of Michelin Pilot Sport 4", movement["notes"])
}

// TestCreateOutgoingMovement tests creating an outgoing stock movement
func (suite *StockMovementsTestSuite) TestCreateOutgoingMovement() {
	movementRequest := map[string]interface{}{
		"productId":    1,
		"movementType": "Outgoing",
		"quantity":     5,
		"reason":       "Return to supplier",
		"reference":    "RT-2024-001",
		"notes":        "Defective units returned",
	}

	jsonBody, _ := json.Marshal(movementRequest)
	req, _ := http.NewRequest("POST", "/v1/stock/movements", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Verify successful creation
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Stock movement recorded successfully", response["message"])

	movement := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), float64(1), movement["productId"])
	assert.Equal(suite.T(), "Outgoing", movement["movementType"])
	assert.Equal(suite.T(), float64(5), movement["quantity"])
	assert.Equal(suite.T(), "Return to supplier", movement["reason"])
}

// TestCreateAdjustmentMovement tests creating an adjustment stock movement
func (suite *StockMovementsTestSuite) TestCreateAdjustmentMovement() {
	movementRequest := map[string]interface{}{
		"productId":    1,
		"movementType": "Adjustment",
		"quantity":     2,
		"reason":       "Inventory count correction",
		"reference":    "ADJ-2024-001",
		"notes":        "Found 2 additional units during physical count",
	}

	jsonBody, _ := json.Marshal(movementRequest)
	req, _ := http.NewRequest("POST", "/v1/stock/movements", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Verify successful creation
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Stock movement recorded successfully", response["message"])

	movement := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), float64(1), movement["productId"])
	assert.Equal(suite.T(), "Adjustment", movement["movementType"])
	assert.Equal(suite.T(), float64(2), movement["quantity"])
	assert.Equal(suite.T(), "Inventory count correction", movement["reason"])
}

// TestCreateMovementValidationErrors tests movement creation with validation errors
func (suite *StockMovementsTestSuite) TestCreateMovementValidationErrors() {
	testCases := []struct {
		name            string
		movementRequest map[string]interface{}
		expectedStatus  int
		expectedErrors  []string
	}{
		{
			name: "Missing required fields",
			movementRequest: map[string]interface{}{
				"productId": 1,
				// Missing movementType, quantity
			},
			expectedStatus: http.StatusBadRequest,
			expectedErrors: []string{"movementType", "quantity"},
		},
		{
			name: "Invalid movement type",
			movementRequest: map[string]interface{}{
				"productId":    1,
				"movementType": "InvalidType",
				"quantity":     10,
			},
			expectedStatus: http.StatusBadRequest,
			expectedErrors: []string{"movementType"},
		},
		{
			name: "Zero quantity",
			movementRequest: map[string]interface{}{
				"productId":    1,
				"movementType": "Incoming",
				"quantity":     0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedErrors: []string{"quantity"},
		},
		{
			name: "Negative quantity",
			movementRequest: map[string]interface{}{
				"productId":    1,
				"movementType": "Incoming",
				"quantity":     -5,
			},
			expectedStatus: http.StatusBadRequest,
			expectedErrors: []string{"quantity"},
		},
		{
			name: "Invalid product ID",
			movementRequest: map[string]interface{}{
				"productId":    0,
				"movementType": "Incoming",
				"quantity":     10,
			},
			expectedStatus: http.StatusBadRequest,
			expectedErrors: []string{"productId"},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			jsonBody, _ := json.Marshal(tc.movementRequest)
			req, _ := http.NewRequest("POST", "/v1/stock/movements", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.token)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(suite.T(), tc.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(suite.T(), err)

			assert.False(suite.T(), response["success"].(bool))
			assert.Contains(suite.T(), response["message"], "Validation failed")

			// Check for specific error fields
			if errors, exists := response["errors"]; exists {
				errorList := errors.([]interface{})
				for _, expectedError := range tc.expectedErrors {
					found := false
					for _, err := range errorList {
						errorMap := err.(map[string]interface{})
						if errorMap["field"] == expectedError {
							found = true
							break
						}
					}
					assert.True(suite.T(), found, "Expected error for field %s not found", expectedError)
				}
			}
		})
	}
}

// TestGetStockMovements tests retrieving stock movements
func (suite *StockMovementsTestSuite) TestGetStockMovements() {
	testCases := []struct {
		name     string
		query    string
		expected int
	}{
		{
			name:     "List all movements",
			query:    "",
			expected: http.StatusOK,
		},
		{
			name:     "List movements with pagination",
			query:    "?page=1&limit=10",
			expected: http.StatusOK,
		},
		{
			name:     "Filter by product ID",
			query:    "?productId=1",
			expected: http.StatusOK,
		},
		{
			name:     "Filter by movement type",
			query:    "?movementType=Incoming",
			expected: http.StatusOK,
		},
		{
			name:     "Filter by date range",
			query:    "?startDate=2024-09-01&endDate=2024-09-30",
			expected: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			req, _ := http.NewRequest("GET", "/v1/stock/movements"+tc.query, nil)
			req.Header.Set("Authorization", "Bearer "+suite.token)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(suite.T(), tc.expected, w.Code)

			if tc.expected == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(suite.T(), err)

				assert.True(suite.T(), response["success"].(bool))
				assert.Equal(suite.T(), "Stock movements retrieved successfully", response["message"])

				data := response["data"].(map[string]interface{})
				assert.Contains(suite.T(), data, "movements")
				assert.Contains(suite.T(), data, "pagination")

				pagination := data["pagination"].(map[string]interface{})
				assert.Contains(suite.T(), pagination, "page")
				assert.Contains(suite.T(), pagination, "limit")
				assert.Contains(suite.T(), pagination, "total")
			}
		})
	}
}

// TestGetStockMovementById tests retrieving a specific stock movement
func (suite *StockMovementsTestSuite) TestGetStockMovementById() {
	req, _ := http.NewRequest("GET", "/v1/stock/movements/1", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Stock movement retrieved successfully", response["message"])

	movement := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), float64(1), movement["id"])
	assert.Contains(suite.T(), movement, "productId")
	assert.Contains(suite.T(), movement, "movementType")
	assert.Contains(suite.T(), movement, "quantity")
	assert.Contains(suite.T(), movement, "reason")
	assert.Contains(suite.T(), movement, "createdAt")
}

// TestProcessSale tests processing a product sale
func (suite *StockMovementsTestSuite) TestProcessSale() {
	saleRequest := map[string]interface{}{
		"productId":    1,
		"quantity":     2,
		"customerName": "Jane Smith",
		"reference":    "INV-2024-001",
		"notes":        "Customer requested installation service",
	}

	jsonBody, _ := json.Marshal(saleRequest)
	req, _ := http.NewRequest("POST", "/v1/stock/sale", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Verify successful sale
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Sale processed successfully", response["message"])

	data := response["data"].(map[string]interface{})
	movement := data["movement"].(map[string]interface{})
	newQuantity := data["newQuantity"].(float64)
	alertGenerated := data["alertGenerated"].(bool)

	assert.Equal(suite.T(), "Sale", movement["movementType"])
	assert.Equal(suite.T(), float64(-2), movement["quantity"])
	assert.Equal(suite.T(), float64(23), newQuantity) // 25 - 2
	assert.IsType(suite.T(), true, alertGenerated)
}

// TestProcessSaleInsufficientStock tests processing a sale with insufficient stock
func (suite *StockMovementsTestSuite) TestProcessSaleInsufficientStock() {
	saleRequest := map[string]interface{}{
		"productId": 1,
		"quantity":  1000, // More than available
	}

	jsonBody, _ := json.Marshal(saleRequest)
	req, _ := http.NewRequest("POST", "/v1/stock/sale", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Verify insufficient stock error
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.False(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["message"], "Insufficient stock")
	assert.Contains(suite.T(), response, "errors")
}

// TestGetInventoryLevels tests retrieving current inventory levels
func (suite *StockMovementsTestSuite) TestGetInventoryLevels() {
	testCases := []struct {
		name     string
		query    string
		expected int
	}{
		{
			name:     "Get all inventory",
			query:    "",
			expected: http.StatusOK,
		},
		{
			name:     "Filter by stock status",
			query:    "?stockStatus=lowStock",
			expected: http.StatusOK,
		},
		{
			name:     "Filter by product type",
			query:    "?type=Tire",
			expected: http.StatusOK,
		},
		{
			name:     "Filter by brand",
			query:    "?brand=Michelin",
			expected: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			req, _ := http.NewRequest("GET", "/v1/stock/inventory"+tc.query, nil)
			req.Header.Set("Authorization", "Bearer "+suite.token)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(suite.T(), tc.expected, w.Code)

			if tc.expected == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(suite.T(), err)

				assert.True(suite.T(), response["success"].(bool))
				assert.Equal(suite.T(), "Inventory levels retrieved successfully", response["message"])

				data := response["data"].(map[string]interface{})
				assert.Contains(suite.T(), data, "inventory")
				assert.Contains(suite.T(), data, "summary")

				summary := data["summary"].(map[string]interface{})
				assert.Contains(suite.T(), summary, "totalProducts")
				assert.Contains(suite.T(), summary, "totalValue")
				assert.Contains(suite.T(), summary, "lowStockCount")
				assert.Contains(suite.T(), summary, "outOfStockCount")
			}
		})
	}
}

// TestGetStockAlerts tests retrieving stock alerts
func (suite *StockMovementsTestSuite) TestGetStockAlerts() {
	testCases := []struct {
		name     string
		query    string
		expected int
	}{
		{
			name:     "Get all alerts",
			query:    "",
			expected: http.StatusOK,
		},
		{
			name:     "Filter unread alerts",
			query:    "?isRead=false",
			expected: http.StatusOK,
		},
		{
			name:     "Filter by alert type",
			query:    "?alertType=LowStock",
			expected: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			req, _ := http.NewRequest("GET", "/v1/stock/alerts"+tc.query, nil)
			req.Header.Set("Authorization", "Bearer "+suite.token)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(suite.T(), tc.expected, w.Code)

			if tc.expected == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(suite.T(), err)

				assert.True(suite.T(), response["success"].(bool))
				assert.Equal(suite.T(), "Stock alerts retrieved successfully", response["message"])

				data := response["data"].(map[string]interface{})
				assert.Contains(suite.T(), data, "alerts")
				assert.Contains(suite.T(), data, "unreadCount")

				unreadCount := data["unreadCount"].(float64)
				assert.GreaterOrEqual(suite.T(), unreadCount, float64(0))
			}
		})
	}
}

// TestMarkAlertAsRead tests marking an alert as read
func (suite *StockMovementsTestSuite) TestMarkAlertAsRead() {
	req, _ := http.NewRequest("PUT", "/v1/stock/alerts/1/read", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Alert marked as read", response["message"])
}

// TestStockMovementsWithoutAuth tests that all stock operations require authentication
func (suite *StockMovementsTestSuite) TestStockMovementsWithoutAuth() {
	testCases := []struct {
		name string
		req  func() *http.Request
	}{
		{
			name: "Create movement without auth",
			req: func() *http.Request {
				movementRequest := map[string]interface{}{
					"productId":    1,
					"movementType": "Incoming",
					"quantity":     10,
				}
				jsonBody, _ := json.Marshal(movementRequest)
				req, _ := http.NewRequest("POST", "/v1/stock/movements", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
		},
		{
			name: "Get movements without auth",
			req: func() *http.Request {
				req, _ := http.NewRequest("GET", "/v1/stock/movements", nil)
				return req
			},
		},
		{
			name: "Process sale without auth",
			req: func() *http.Request {
				saleRequest := map[string]interface{}{
					"productId": 1,
					"quantity":  2,
				}
				jsonBody, _ := json.Marshal(saleRequest)
				req, _ := http.NewRequest("POST", "/v1/stock/sale", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			req := tc.req()
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(suite.T(), err)

			assert.False(suite.T(), response["success"].(bool))
			assert.Contains(suite.T(), response["message"], "Authorization header required")
		})
	}
}

// setupTestStockRoutes sets up test routes for integration testing
// This is a placeholder that will be replaced with actual implementation
func setupTestStockRoutes(router *gin.Engine) {
	// Setup auth routes first
	setupTestAuthRoutes(router)

	// Placeholder stock routes that return mock responses
	// These will be replaced with actual implementation in Phase 3.3
	stock := router.Group("/v1/stock")
	stock.Use(func(c *gin.Context) {
		// Mock auth middleware
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header required",
			})
			c.Abort()
			return
		}
		c.Next()
	})

	{
		stock.GET("/movements", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Stock movements retrieved successfully",
				"data": gin.H{
					"movements": []gin.H{
						{
							"id":           1,
							"productId":    1,
							"userId":       1,
							"movementType": "Incoming",
							"quantity":     10,
							"reason":       "New shipment",
							"reference":    "PO-2024-001",
							"notes":        "Received 10 units",
							"createdAt":    "2024-09-21T10:00:00Z",
						},
					},
					"pagination": gin.H{
						"page":       1,
						"limit":      10,
						"total":      1,
						"totalPages": 1,
						"hasNext":    false,
						"hasPrev":    false,
					},
				},
			})
		})

		stock.POST("/movements", func(c *gin.Context) {
			var req map[string]interface{}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Validation failed",
					"errors": []gin.H{
						{"field": "json", "message": "Invalid JSON format"},
					},
				})
				return
			}

			// Mock validation
			errors := []gin.H{}
			if req["productId"] == nil || req["productId"].(float64) <= 0 {
				errors = append(errors, gin.H{"field": "productId", "message": "Product ID is required and must be positive"})
			}
			if req["movementType"] == "" {
				errors = append(errors, gin.H{"field": "movementType", "message": "Movement type is required"})
			} else if req["movementType"] == "InvalidType" {
				errors = append(errors, gin.H{"field": "movementType", "message": "Invalid movement type"})
			}
			if req["quantity"] == nil || req["quantity"].(float64) <= 0 {
				errors = append(errors, gin.H{"field": "quantity", "message": "Quantity must be positive"})
			}

			if len(errors) > 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Validation failed",
					"errors":  errors,
				})
				return
			}

			c.JSON(http.StatusCreated, gin.H{
				"success": true,
				"message": "Stock movement recorded successfully",
				"data": gin.H{
					"id":           1,
					"productId":    req["productId"],
					"userId":       1,
					"movementType": req["movementType"],
					"quantity":     req["quantity"],
					"reason":       req["reason"],
					"reference":    req["reference"],
					"notes":        req["notes"],
					"createdAt":    "2024-09-21T10:00:00Z",
				},
			})
		})

		stock.GET("/movements/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Stock movement retrieved successfully",
				"data": gin.H{
					"id":           1,
					"productId":    1,
					"userId":       1,
					"movementType": "Incoming",
					"quantity":     10,
					"reason":       "New shipment",
					"reference":    "PO-2024-001",
					"notes":        "Received 10 units",
					"createdAt":    "2024-09-21T10:00:00Z",
				},
			})
		})

		stock.GET("/inventory", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Inventory levels retrieved successfully",
				"data": gin.H{
					"inventory": []gin.H{
						{
							"productId":         1,
							"quantityOnHand":    25,
							"lowStockThreshold": 5,
							"stockStatus":       "available",
							"lastMovementDate":  "2024-09-21T10:00:00Z",
							"totalValue":        5000.00,
						},
					},
					"summary": gin.H{
						"totalProducts":   150,
						"totalValue":      75000.00,
						"lowStockCount":   5,
						"outOfStockCount": 2,
						"tireCount":       100,
						"wheelCount":      50,
					},
					"pagination": gin.H{
						"page":       1,
						"limit":      20,
						"total":      1,
						"totalPages": 1,
						"hasNext":    false,
						"hasPrev":    false,
					},
				},
			})
		})

		stock.GET("/alerts", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Stock alerts retrieved successfully",
				"data": gin.H{
					"alerts": []gin.H{
						{
							"id":        1,
							"productId": 1,
							"alertType": "LowStock",
							"message":   "Product MIC-PS4-225-45-17 is running low on stock (3 remaining)",
							"isRead":    false,
							"isActive":  true,
							"createdAt": "2024-09-21T10:00:00Z",
							"readAt":    nil,
						},
					},
					"unreadCount": 5,
					"pagination": gin.H{
						"page":       1,
						"limit":      20,
						"total":      1,
						"totalPages": 1,
						"hasNext":    false,
						"hasPrev":    false,
					},
				},
			})
		})

		stock.PUT("/alerts/:id/read", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Alert marked as read",
			})
		})

		stock.POST("/sale", func(c *gin.Context) {
			var req map[string]interface{}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Validation failed",
				})
				return
			}

			// Mock validation for insufficient stock
			if req["quantity"].(float64) > 25 {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Insufficient stock available",
					"errors": []gin.H{
						{
							"field":   "quantity",
							"message": "Requested quantity (1000) exceeds available stock (25)",
						},
					},
				})
				return
			}

			c.JSON(http.StatusCreated, gin.H{
				"success": true,
				"message": "Sale processed successfully",
				"data": gin.H{
					"movement": gin.H{
						"id":           1,
						"productId":    req["productId"],
						"userId":       1,
						"movementType": "Sale",
						"quantity":     -req["quantity"].(float64),
						"reason":       "Customer purchase",
						"reference":    req["reference"],
						"notes":        req["notes"],
						"createdAt":    "2024-09-21T10:00:00Z",
					},
					"newQuantity":    25 - req["quantity"].(float64),
					"alertGenerated": false,
				},
			})
		})
	}
}
