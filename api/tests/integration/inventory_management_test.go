package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// InventoryManagementTestSuite tests inventory management functionality
// This includes stock tracking, low stock alerts, and inventory reports
type InventoryManagementTestSuite struct {
	suite.Suite
	router *gin.Engine
	token  string
}

// SetupSuite runs once before all tests in the suite
func (suite *InventoryManagementTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	setupTestInventoryRoutes(suite.router)
	suite.token = getAuthToken(suite.router)
}

// TestInventoryManagementTestSuite runs the test suite
func TestInventoryManagementTestSuite(t *testing.T) {
	suite.Run(t, new(InventoryManagementTestSuite))
}

// TestGetInventoryLevels tests retrieving current inventory levels
func (suite *InventoryManagementTestSuite) TestGetInventoryLevels() {
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
			name:     "Filter by stock status - available",
			query:    "?stockStatus=available",
			expected: http.StatusOK,
		},
		{
			name:     "Filter by stock status - lowStock",
			query:    "?stockStatus=lowStock",
			expected: http.StatusOK,
		},
		{
			name:     "Filter by stock status - outOfStock",
			query:    "?stockStatus=outOfStock",
			expected: http.StatusOK,
		},
		{
			name:     "Filter by product type - Tire",
			query:    "?type=Tire",
			expected: http.StatusOK,
		},
		{
			name:     "Filter by product type - Wheel",
			query:    "?type=Wheel",
			expected: http.StatusOK,
		},
		{
			name:     "Filter by brand",
			query:    "?brand=Michelin",
			expected: http.StatusOK,
		},
		{
			name:     "Filter with pagination",
			query:    "?page=1&limit=10",
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
				assert.Contains(suite.T(), data, "pagination")

				// Verify summary structure
				summary := data["summary"].(map[string]interface{})
				assert.Contains(suite.T(), summary, "totalProducts")
				assert.Contains(suite.T(), summary, "totalValue")
				assert.Contains(suite.T(), summary, "lowStockCount")
				assert.Contains(suite.T(), summary, "outOfStockCount")
				assert.Contains(suite.T(), summary, "tireCount")
				assert.Contains(suite.T(), summary, "wheelCount")

				// Verify inventory items structure
				inventory := data["inventory"].([]interface{})
				if len(inventory) > 0 {
					item := inventory[0].(map[string]interface{})
					assert.Contains(suite.T(), item, "productId")
					assert.Contains(suite.T(), item, "quantityOnHand")
					assert.Contains(suite.T(), item, "lowStockThreshold")
					assert.Contains(suite.T(), item, "stockStatus")
					assert.Contains(suite.T(), item, "lastMovementDate")
					assert.Contains(suite.T(), item, "totalValue")
				}

				// Verify pagination structure
				pagination := data["pagination"].(map[string]interface{})
				assert.Contains(suite.T(), pagination, "page")
				assert.Contains(suite.T(), pagination, "limit")
				assert.Contains(suite.T(), pagination, "total")
				assert.Contains(suite.T(), pagination, "totalPages")
			}
		})
	}
}

// TestGetStockAlerts tests retrieving stock alerts
func (suite *InventoryManagementTestSuite) TestGetStockAlerts() {
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
			name:     "Filter read alerts",
			query:    "?isRead=true",
			expected: http.StatusOK,
		},
		{
			name:     "Filter by alert type - LowStock",
			query:    "?alertType=LowStock",
			expected: http.StatusOK,
		},
		{
			name:     "Filter by alert type - OutOfStock",
			query:    "?alertType=OutOfStock",
			expected: http.StatusOK,
		},
		{
			name:     "Filter by product ID",
			query:    "?productId=1",
			expected: http.StatusOK,
		},
		{
			name:     "Filter with pagination",
			query:    "?page=1&limit=10",
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
				assert.Contains(suite.T(), data, "pagination")

				// Verify unread count
				unreadCount := data["unreadCount"].(float64)
				assert.GreaterOrEqual(suite.T(), unreadCount, float64(0))

				// Verify alerts structure
				alerts := data["alerts"].([]interface{})
				if len(alerts) > 0 {
					alert := alerts[0].(map[string]interface{})
					assert.Contains(suite.T(), alert, "id")
					assert.Contains(suite.T(), alert, "productId")
					assert.Contains(suite.T(), alert, "alertType")
					assert.Contains(suite.T(), alert, "message")
					assert.Contains(suite.T(), alert, "isRead")
					assert.Contains(suite.T(), alert, "isActive")
					assert.Contains(suite.T(), alert, "createdAt")
				}
			}
		})
	}
}

// TestMarkAlertAsRead tests marking an alert as read
func (suite *InventoryManagementTestSuite) TestMarkAlertAsRead() {
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

// TestMarkAllAlertsAsRead tests marking all alerts as read
func (suite *InventoryManagementTestSuite) TestMarkAllAlertsAsRead() {
	req, _ := http.NewRequest("PUT", "/v1/stock/alerts/read-all", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "All alerts marked as read", response["message"])

	// Verify the count of marked alerts
	data := response["data"].(map[string]interface{})
	markedCount := data["markedCount"].(float64)
	assert.GreaterOrEqual(suite.T(), markedCount, float64(0))
}

// TestGetInventoryReport tests generating inventory reports
func (suite *InventoryManagementTestSuite) TestGetInventoryReport() {
	testCases := []struct {
		name     string
		query    string
		expected int
	}{
		{
			name:     "Get full inventory report",
			query:    "",
			expected: http.StatusOK,
		},
		{
			name:     "Get report by date range",
			query:    "?startDate=2024-09-01&endDate=2024-09-30",
			expected: http.StatusOK,
		},
		{
			name:     "Get report by product type",
			query:    "?type=Tire",
			expected: http.StatusOK,
		},
		{
			name:     "Get report by brand",
			query:    "?brand=Michelin",
			expected: http.StatusOK,
		},
		{
			name:     "Get report with stock status filter",
			query:    "?stockStatus=lowStock",
			expected: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			req, _ := http.NewRequest("GET", "/v1/stock/reports/inventory"+tc.query, nil)
			req.Header.Set("Authorization", "Bearer "+suite.token)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(suite.T(), tc.expected, w.Code)

			if tc.expected == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(suite.T(), err)

				assert.True(suite.T(), response["success"].(bool))
				assert.Equal(suite.T(), "Inventory report generated successfully", response["message"])

				data := response["data"].(map[string]interface{})
				assert.Contains(suite.T(), data, "report")
				assert.Contains(suite.T(), data, "summary")
				assert.Contains(suite.T(), data, "generatedAt")

				// Verify report structure
				report := data["report"].(map[string]interface{})
				assert.Contains(suite.T(), report, "products")
				assert.Contains(suite.T(), report, "totalValue")
				assert.Contains(suite.T(), report, "lowStockItems")
				assert.Contains(suite.T(), report, "outOfStockItems")

				// Verify summary structure
				summary := data["summary"].(map[string]interface{})
				assert.Contains(suite.T(), summary, "totalProducts")
				assert.Contains(suite.T(), summary, "totalValue")
				assert.Contains(suite.T(), summary, "lowStockCount")
				assert.Contains(suite.T(), summary, "outOfStockCount")
			}
		})
	}
}

// TestGetStockMovementReport tests generating stock movement reports
func (suite *InventoryManagementTestSuite) TestGetStockMovementReport() {
	testCases := []struct {
		name     string
		query    string
		expected int
	}{
		{
			name:     "Get stock movement report",
			query:    "?startDate=2024-09-01&endDate=2024-09-30",
			expected: http.StatusOK,
		},
		{
			name:     "Get report by product ID",
			query:    "?productId=1&startDate=2024-09-01&endDate=2024-09-30",
			expected: http.StatusOK,
		},
		{
			name:     "Get report by movement type",
			query:    "?movementType=Incoming&startDate=2024-09-01&endDate=2024-09-30",
			expected: http.StatusOK,
		},
		{
			name:     "Get report by user ID",
			query:    "?userId=1&startDate=2024-09-01&endDate=2024-09-30",
			expected: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			req, _ := http.NewRequest("GET", "/v1/stock/reports/movements"+tc.query, nil)
			req.Header.Set("Authorization", "Bearer "+suite.token)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(suite.T(), tc.expected, w.Code)

			if tc.expected == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(suite.T(), err)

				assert.True(suite.T(), response["success"].(bool))
				assert.Equal(suite.T(), "Stock movement report generated successfully", response["message"])

				data := response["data"].(map[string]interface{})
				assert.Contains(suite.T(), data, "report")
				assert.Contains(suite.T(), data, "summary")
				assert.Contains(suite.T(), data, "generatedAt")

				// Verify report structure
				report := data["report"].(map[string]interface{})
				assert.Contains(suite.T(), report, "movements")
				assert.Contains(suite.T(), report, "totalMovements")
				assert.Contains(suite.T(), report, "incomingTotal")
				assert.Contains(suite.T(), report, "outgoingTotal")

				// Verify summary structure
				summary := data["summary"].(map[string]interface{})
				assert.Contains(suite.T(), summary, "totalMovements")
				assert.Contains(suite.T(), summary, "incomingCount")
				assert.Contains(suite.T(), summary, "outgoingCount")
				assert.Contains(suite.T(), summary, "adjustmentCount")
			}
		})
	}
}

// TestInventoryManagementWithoutAuth tests that all inventory operations require authentication
func (suite *InventoryManagementTestSuite) TestInventoryManagementWithoutAuth() {
	testCases := []struct {
		name string
		req  func() *http.Request
	}{
		{
			name: "Get inventory without auth",
			req: func() *http.Request {
				req, _ := http.NewRequest("GET", "/v1/stock/inventory", nil)
				return req
			},
		},
		{
			name: "Get alerts without auth",
			req: func() *http.Request {
				req, _ := http.NewRequest("GET", "/v1/stock/alerts", nil)
				return req
			},
		},
		{
			name: "Mark alert as read without auth",
			req: func() *http.Request {
				req, _ := http.NewRequest("PUT", "/v1/stock/alerts/1/read", nil)
				return req
			},
		},
		{
			name: "Get inventory report without auth",
			req: func() *http.Request {
				req, _ := http.NewRequest("GET", "/v1/stock/reports/inventory", nil)
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

// TestInventoryValidationErrors tests inventory operations with validation errors
func (suite *InventoryManagementTestSuite) TestInventoryValidationErrors() {
	testCases := []struct {
		name           string
		url            string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Invalid stock status filter",
			url:            "/v1/stock/inventory?stockStatus=InvalidStatus",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid stock status",
		},
		{
			name:           "Invalid product type filter",
			url:            "/v1/stock/inventory?type=InvalidType",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid product type",
		},
		{
			name:           "Invalid alert type filter",
			url:            "/v1/stock/alerts?alertType=InvalidType",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid alert type",
		},
		{
			name:           "Invalid date range in report",
			url:            "/v1/stock/reports/inventory?startDate=2024-09-30&endDate=2024-09-01",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid date range",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			req, _ := http.NewRequest("GET", tc.url, nil)
			req.Header.Set("Authorization", "Bearer "+suite.token)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(suite.T(), tc.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(suite.T(), err)

			assert.False(suite.T(), response["success"].(bool))
			assert.Contains(suite.T(), response["message"], tc.expectedError)
		})
	}
}

// setupTestInventoryRoutes sets up test routes for integration testing
// This is a placeholder that will be replaced with actual implementation
func setupTestInventoryRoutes(router *gin.Engine) {
	// Setup auth routes first
	setupTestAuthRoutes(router)

	// Placeholder inventory routes that return mock responses
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
		stock.GET("/inventory", func(c *gin.Context) {
			// Mock validation
			stockStatus := c.Query("stockStatus")
			if stockStatus != "" && stockStatus != "available" && stockStatus != "lowStock" && stockStatus != "outOfStock" {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid stock status",
				})
				return
			}

			productType := c.Query("type")
			if productType != "" && productType != "Tire" && productType != "Wheel" {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid product type",
				})
				return
			}

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
			// Mock validation
			alertType := c.Query("alertType")
			if alertType != "" && alertType != "LowStock" && alertType != "OutOfStock" {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid alert type",
				})
				return
			}

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

		stock.PUT("/alerts/read-all", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "All alerts marked as read",
				"data": gin.H{
					"markedCount": 5,
				},
			})
		})

		stock.GET("/reports/inventory", func(c *gin.Context) {
			// Mock validation
			startDate := c.Query("startDate")
			endDate := c.Query("endDate")
			if startDate != "" && endDate != "" && startDate > endDate {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid date range",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Inventory report generated successfully",
				"data": gin.H{
					"report": gin.H{
						"products": []gin.H{
							{
								"id":                1,
								"type":              "Tire",
								"brand":             "Michelin",
								"model":             "Pilot Sport 4",
								"sku":               "MIC-PS4-225-45-17",
								"quantityOnHand":    25,
								"lowStockThreshold": 5,
								"stockStatus":       "available",
								"totalValue":        5000.00,
							},
						},
						"totalValue":      75000.00,
						"lowStockItems":   []gin.H{},
						"outOfStockItems": []gin.H{},
					},
					"summary": gin.H{
						"totalProducts":   150,
						"totalValue":      75000.00,
						"lowStockCount":   5,
						"outOfStockCount": 2,
					},
					"generatedAt": "2024-09-21T10:00:00Z",
				},
			})
		})

		stock.GET("/reports/movements", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Stock movement report generated successfully",
				"data": gin.H{
					"report": gin.H{
						"movements": []gin.H{
							{
								"id":           1,
								"productId":    1,
								"movementType": "Incoming",
								"quantity":     10,
								"reason":       "New shipment",
								"createdAt":    "2024-09-21T10:00:00Z",
							},
						},
						"totalMovements": 1,
						"incomingTotal":  10,
						"outgoingTotal":  0,
					},
					"summary": gin.H{
						"totalMovements":  1,
						"incomingCount":   1,
						"outgoingCount":   0,
						"adjustmentCount": 0,
					},
					"generatedAt": "2024-09-21T10:00:00Z",
				},
			})
		})
	}
}
