package contract

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStockContract tests the stock management API contract
// These tests verify the API endpoints match the OpenAPI specification
func TestStockContract(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create test router (will be replaced with actual router in implementation)
	router := gin.New()
	setupTestStockRoutes(router)

	// Get auth token for protected endpoints
	token := getAuthToken(router)

	t.Run("GET /v1/stock/movements - List stock movements", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/stock/movements?productId=1&limit=10", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Contract validation - should return 200
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify response structure matches OpenAPI spec
		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Stock movements retrieved successfully", response["message"])

		data := response["data"].(map[string]interface{})
		movements := data["movements"].([]interface{})
		pagination := data["pagination"].(map[string]interface{})

		// Verify pagination structure
		assert.Equal(t, float64(1), pagination["page"])
		assert.Equal(t, float64(10), pagination["limit"])
		assert.GreaterOrEqual(t, pagination["total"], float64(0))

		// Verify movement structure if movements exist
		if len(movements) > 0 {
			movement := movements[0].(map[string]interface{})
			assert.Contains(t, movement, "id")
			assert.Contains(t, movement, "productId")
			assert.Contains(t, movement, "userId")
			assert.Contains(t, movement, "movementType")
			assert.Contains(t, movement, "quantity")
			assert.Contains(t, movement, "createdAt")
		}
	})

	t.Run("POST /v1/stock/movements - Create stock movement", func(t *testing.T) {
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
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Contract validation - should return 201
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Stock movement recorded successfully", response["message"])

		movement := response["data"].(map[string]interface{})
		assert.Equal(t, float64(1), movement["productId"])
		assert.Equal(t, "Incoming", movement["movementType"])
		assert.Equal(t, float64(10), movement["quantity"])
		assert.Equal(t, "New shipment from supplier", movement["reason"])
		assert.Equal(t, "PO-2024-001", movement["reference"])
	})

	t.Run("GET /v1/stock/movements/{id} - Get stock movement by ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/stock/movements/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Contract validation - should return 200
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Stock movement retrieved successfully", response["message"])

		movement := response["data"].(map[string]interface{})
		assert.Equal(t, float64(1), movement["id"])
		assert.Contains(t, movement, "productId")
		assert.Contains(t, movement, "movementType")
		assert.Contains(t, movement, "quantity")
	})

	t.Run("GET /v1/stock/inventory - Get current inventory", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/stock/inventory?stockStatus=lowStock", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Contract validation - should return 200
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Inventory levels retrieved successfully", response["message"])

		data := response["data"].(map[string]interface{})
		inventory := data["inventory"].([]interface{})
		summary := data["summary"].(map[string]interface{})

		// Verify summary structure
		assert.Contains(t, summary, "totalProducts")
		assert.Contains(t, summary, "totalValue")
		assert.Contains(t, summary, "lowStockCount")
		assert.Contains(t, summary, "outOfStockCount")

		// Verify inventory item structure if items exist
		if len(inventory) > 0 {
			item := inventory[0].(map[string]interface{})
			assert.Contains(t, item, "productId")
			assert.Contains(t, item, "quantityOnHand")
			assert.Contains(t, item, "stockStatus")
			assert.Contains(t, item, "totalValue")
		}
	})

	t.Run("GET /v1/stock/alerts - Get stock alerts", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/stock/alerts?isRead=false", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Contract validation - should return 200
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Stock alerts retrieved successfully", response["message"])

		data := response["data"].(map[string]interface{})
		alerts := data["alerts"].([]interface{})
		unreadCount := data["unreadCount"].(float64)

		assert.GreaterOrEqual(t, unreadCount, float64(0))

		// Verify alert structure if alerts exist
		if len(alerts) > 0 {
			alert := alerts[0].(map[string]interface{})
			assert.Contains(t, alert, "id")
			assert.Contains(t, alert, "productId")
			assert.Contains(t, alert, "alertType")
			assert.Contains(t, alert, "message")
			assert.Contains(t, alert, "isRead")
			assert.Contains(t, alert, "createdAt")
		}
	})

	t.Run("PUT /v1/stock/alerts/{id}/read - Mark alert as read", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/v1/stock/alerts/1/read", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Contract validation - should return 200
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Alert marked as read", response["message"])
	})

	t.Run("POST /v1/stock/sale - Process product sale", func(t *testing.T) {
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
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Contract validation - should return 201
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Sale processed successfully", response["message"])

		data := response["data"].(map[string]interface{})
		movement := data["movement"].(map[string]interface{})
		newQuantity := data["newQuantity"].(float64)
		alertGenerated := data["alertGenerated"].(bool)

		assert.Equal(t, "Sale", movement["movementType"])
		assert.Equal(t, float64(-2), movement["quantity"])
		assert.Equal(t, float64(23), newQuantity) // 25 - 2
		assert.IsType(t, true, alertGenerated)
	})

	t.Run("POST /v1/stock/sale - Insufficient stock", func(t *testing.T) {
		saleRequest := map[string]interface{}{
			"productId": 1,
			"quantity":  1000, // More than available
		}

		jsonBody, _ := json.Marshal(saleRequest)
		req, _ := http.NewRequest("POST", "/v1/stock/sale", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Contract validation - should return 400
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.False(t, response["success"].(bool))
		assert.Contains(t, response["message"], "Insufficient stock")
		assert.Contains(t, response, "errors")
	})

	t.Run("POST /v1/stock/movements - Validation error", func(t *testing.T) {
		invalidRequest := map[string]interface{}{
			"productId":    1,
			"movementType": "InvalidType", // Invalid movement type
			"quantity":     0,             // Invalid: zero quantity
		}

		jsonBody, _ := json.Marshal(invalidRequest)
		req, _ := http.NewRequest("POST", "/v1/stock/movements", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Contract validation - should return 400
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.False(t, response["success"].(bool))
		assert.Contains(t, response["message"], "Validation failed")
		assert.Contains(t, response, "errors")
	})
}

// setupTestStockRoutes sets up test routes for contract testing
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
			if req["movementType"] == "InvalidType" || req["quantity"].(float64) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Validation failed",
					"errors": []gin.H{
						{"field": "movementType", "message": "Invalid movement type"},
						{"field": "quantity", "message": "Quantity cannot be zero"},
					},
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
