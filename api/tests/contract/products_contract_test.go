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

// TestProductsContract tests the products API contract
// These tests verify the API endpoints match the OpenAPI specification
func TestProductsContract(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create test router (will be replaced with actual router in implementation)
	router := gin.New()
	setupTestProductRoutes(router)

	// Get auth token for protected endpoints
	token := getAuthToken(router)

	t.Run("GET /v1/products - List products with pagination", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/products?page=1&limit=20&type=Tire", nil)
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
		assert.Equal(t, "Products retrieved successfully", response["message"])

		data := response["data"].(map[string]interface{})
		products := data["products"].([]interface{})
		pagination := data["pagination"].(map[string]interface{})

		// Verify pagination structure
		assert.Equal(t, float64(1), pagination["page"])
		assert.Equal(t, float64(20), pagination["limit"])
		assert.GreaterOrEqual(t, pagination["total"], float64(0))
		assert.GreaterOrEqual(t, pagination["totalPages"], float64(0))
		assert.IsType(t, true, pagination["hasNext"])
		assert.IsType(t, true, pagination["hasPrev"])

		// Verify product structure if products exist
		if len(products) > 0 {
			product := products[0].(map[string]interface{})
			assert.Contains(t, product, "id")
			assert.Contains(t, product, "type")
			assert.Contains(t, product, "brand")
			assert.Contains(t, product, "model")
			assert.Contains(t, product, "sku")
			assert.Contains(t, product, "costPrice")
			assert.Contains(t, product, "sellingPrice")
			assert.Contains(t, product, "quantityOnHand")
		}
	})

	t.Run("POST /v1/products - Create new tire product", func(t *testing.T) {
		productRequest := map[string]interface{}{
			"type":              "Tire",
			"brand":             "Michelin",
			"model":             "Pilot Sport 4",
			"sku":               "MIC-PS4-225-45-17",
			"description":       "High-performance summer tire",
			"costPrice":         150.00,
			"sellingPrice":      200.00,
			"quantityOnHand":    25,
			"lowStockThreshold": 5,
			"specifications": map[string]interface{}{
				"specType": "Tire",
				"specData": map[string]interface{}{
					"width":       "225",
					"aspectRatio": "45",
					"diameter":    "17",
					"loadIndex":   "91",
					"speedRating": "W",
					"dotYear":     "2023",
					"season":      "All-Season",
					"runFlat":     false,
				},
			},
		}

		jsonBody, _ := json.Marshal(productRequest)
		req, _ := http.NewRequest("POST", "/v1/products", bytes.NewBuffer(jsonBody))
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
		assert.Equal(t, "Product created successfully", response["message"])

		product := response["data"].(map[string]interface{})
		assert.Equal(t, "Tire", product["type"])
		assert.Equal(t, "Michelin", product["brand"])
		assert.Equal(t, "Pilot Sport 4", product["model"])
		assert.Equal(t, "MIC-PS4-225-45-17", product["sku"])
		assert.Equal(t, 150.00, product["costPrice"])
		assert.Equal(t, 200.00, product["sellingPrice"])
		assert.Equal(t, float64(25), product["quantityOnHand"])
	})

	t.Run("POST /v1/products - Create new wheel product", func(t *testing.T) {
		productRequest := map[string]interface{}{
			"type":              "Wheel",
			"brand":             "Enkei",
			"model":             "Racing RPF1",
			"sku":               "ENK-RPF1-17-8.5-35",
			"description":       "Lightweight racing wheel",
			"costPrice":         300.00,
			"sellingPrice":      400.00,
			"quantityOnHand":    10,
			"lowStockThreshold": 3,
			"specifications": map[string]interface{}{
				"specType": "Wheel",
				"specData": map[string]interface{}{
					"diameter":    "17",
					"width":       "8.5",
					"offset":      "35",
					"boltPattern": "5x114.3",
					"centerBore":  "67.1",
					"color":       "Black",
					"finish":      "Matte",
					"weight":      "22.5",
				},
			},
		}

		jsonBody, _ := json.Marshal(productRequest)
		req, _ := http.NewRequest("POST", "/v1/products", bytes.NewBuffer(jsonBody))
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
		assert.Equal(t, "Product created successfully", response["message"])

		product := response["data"].(map[string]interface{})
		assert.Equal(t, "Wheel", product["type"])
		assert.Equal(t, "Enkei", product["brand"])
		assert.Equal(t, "Racing RPF1", product["model"])
		assert.Equal(t, "ENK-RPF1-17-8.5-35", product["sku"])
	})

	t.Run("GET /v1/products/{id} - Get product by ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/products/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Contract validation - should return 200
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Product retrieved successfully", response["message"])

		product := response["data"].(map[string]interface{})
		assert.Equal(t, float64(1), product["id"])
		assert.Contains(t, product, "type")
		assert.Contains(t, product, "brand")
		assert.Contains(t, product, "model")
		assert.Contains(t, product, "sku")
	})

	t.Run("GET /v1/products/{id} - Product not found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/products/999", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Contract validation - should return 404
		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.False(t, response["success"].(bool))
		assert.Contains(t, response["message"], "not found")
	})

	t.Run("PUT /v1/products/{id} - Update product", func(t *testing.T) {
		updateRequest := map[string]interface{}{
			"brand":        "Michelin Updated",
			"model":        "Pilot Sport 4 Updated",
			"costPrice":    160.00,
			"sellingPrice": 210.00,
		}

		jsonBody, _ := json.Marshal(updateRequest)
		req, _ := http.NewRequest("PUT", "/v1/products/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Contract validation - should return 200
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Product updated successfully", response["message"])

		product := response["data"].(map[string]interface{})
		assert.Equal(t, "Michelin Updated", product["brand"])
		assert.Equal(t, "Pilot Sport 4 Updated", product["model"])
		assert.Equal(t, 160.00, product["costPrice"])
		assert.Equal(t, 210.00, product["sellingPrice"])
	})

	t.Run("DELETE /v1/products/{id} - Delete product", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/v1/products/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Contract validation - should return 200
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Product deleted successfully", response["message"])
	})

	t.Run("POST /v1/products/search - Advanced product search", func(t *testing.T) {
		searchRequest := map[string]interface{}{
			"type": "Tire",
			"tireSpecs": map[string]interface{}{
				"width":       "225",
				"aspectRatio": "45",
				"diameter":    "17",
			},
			"stockStatus": "available",
			"page":        1,
			"limit":       20,
		}

		jsonBody, _ := json.Marshal(searchRequest)
		req, _ := http.NewRequest("POST", "/v1/products/search", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Contract validation - should return 200
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Search completed successfully", response["message"])

		data := response["data"].(map[string]interface{})
		_ = data["products"].([]interface{})
		assert.GreaterOrEqual(t, data["totalCount"], float64(0))
		assert.Contains(t, data, "searchCriteria")
	})

	t.Run("POST /v1/products - Validation error", func(t *testing.T) {
		invalidRequest := map[string]interface{}{
			"type":         "Tire",
			"brand":        "", // Invalid: empty brand
			"model":        "Test",
			"sku":          "TEST-001",
			"costPrice":    -10.00, // Invalid: negative price
			"sellingPrice": 100.00,
		}

		jsonBody, _ := json.Marshal(invalidRequest)
		req, _ := http.NewRequest("POST", "/v1/products", bytes.NewBuffer(jsonBody))
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

// getAuthToken helper function to get authentication token for tests
func getAuthToken(router *gin.Engine) string {
	loginRequest := map[string]string{
		"phoneNumber": "1234567890",
		"pin":         "1234",
	}

	jsonBody, _ := json.Marshal(loginRequest)
	req, _ := http.NewRequest("POST", "/v1/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})
	return data["token"].(string)
}

// setupTestProductRoutes sets up test routes for contract testing
// This is a placeholder that will be replaced with actual implementation
func setupTestProductRoutes(router *gin.Engine) {
	// Setup auth routes first
	setupTestAuthRoutes(router)

	// Placeholder product routes that return mock responses
	// These will be replaced with actual implementation in Phase 3.3
	products := router.Group("/v1/products")
	products.Use(func(c *gin.Context) {
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
		products.GET("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Products retrieved successfully",
				"data": gin.H{
					"products": []gin.H{
						{
							"id":                1,
							"type":              "Tire",
							"brand":             "Michelin",
							"model":             "Pilot Sport 4",
							"sku":               "MIC-PS4-225-45-17",
							"description":       "High-performance summer tire",
							"costPrice":         150.00,
							"sellingPrice":      200.00,
							"quantityOnHand":    25,
							"lowStockThreshold": 5,
							"isActive":          true,
						},
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

		products.POST("", func(c *gin.Context) {
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
			if req["brand"] == "" || req["costPrice"].(float64) < 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Validation failed",
					"errors": []gin.H{
						{"field": "brand", "message": "Brand is required"},
						{"field": "costPrice", "message": "Cost price must be positive"},
					},
				})
				return
			}

			c.JSON(http.StatusCreated, gin.H{
				"success": true,
				"message": "Product created successfully",
				"data": gin.H{
					"id":                1,
					"type":              req["type"],
					"brand":             req["brand"],
					"model":             req["model"],
					"sku":               req["sku"],
					"description":       req["description"],
					"costPrice":         req["costPrice"],
					"sellingPrice":      req["sellingPrice"],
					"quantityOnHand":    req["quantityOnHand"],
					"lowStockThreshold": req["lowStockThreshold"],
					"isActive":          true,
				},
			})
		})

		products.GET("/:id", func(c *gin.Context) {
			id := c.Param("id")
			if id == "999" {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"message": "Product not found",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Product retrieved successfully",
				"data": gin.H{
					"id":                1,
					"type":              "Tire",
					"brand":             "Michelin",
					"model":             "Pilot Sport 4",
					"sku":               "MIC-PS4-225-45-17",
					"description":       "High-performance summer tire",
					"costPrice":         150.00,
					"sellingPrice":      200.00,
					"quantityOnHand":    25,
					"lowStockThreshold": 5,
					"isActive":          true,
				},
			})
		})

		products.PUT("/:id", func(c *gin.Context) {
			var req map[string]interface{}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Validation failed",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Product updated successfully",
				"data": gin.H{
					"id":                1,
					"type":              "Tire",
					"brand":             req["brand"],
					"model":             req["model"],
					"sku":               "MIC-PS4-225-45-17",
					"description":       "High-performance summer tire",
					"costPrice":         req["costPrice"],
					"sellingPrice":      req["sellingPrice"],
					"quantityOnHand":    25,
					"lowStockThreshold": 5,
					"isActive":          true,
				},
			})
		})

		products.DELETE("/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Product deleted successfully",
			})
		})

		products.POST("/search", func(c *gin.Context) {
			var req map[string]interface{}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Validation failed",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Search completed successfully",
				"data": gin.H{
					"products": []gin.H{
						{
							"id":             1,
							"type":           "Tire",
							"brand":          "Michelin",
							"model":          "Pilot Sport 4",
							"sku":            "MIC-PS4-225-45-17",
							"costPrice":      150.00,
							"sellingPrice":   200.00,
							"quantityOnHand": 25,
						},
					},
					"totalCount":     1,
					"searchCriteria": req,
				},
			})
		})
	}
}
