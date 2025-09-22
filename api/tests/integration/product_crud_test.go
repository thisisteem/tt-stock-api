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

// ProductCRUDTestSuite tests the complete product CRUD operations
// This includes creating, reading, updating, and deleting products
type ProductCRUDTestSuite struct {
	suite.Suite
	router *gin.Engine
	token  string
}

// SetupSuite runs once before all tests in the suite
func (suite *ProductCRUDTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	setupTestProductRoutes(suite.router)
	suite.token = getAuthToken(suite.router)
}

// TestProductCRUDTestSuite runs the test suite
func TestProductCRUDTestSuite(t *testing.T) {
	suite.Run(t, new(ProductCRUDTestSuite))
}

// TestCreateTireProduct tests creating a new tire product
func (suite *ProductCRUDTestSuite) TestCreateTireProduct() {
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
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Verify successful creation
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Product created successfully", response["message"])

	product := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "Tire", product["type"])
	assert.Equal(suite.T(), "Michelin", product["brand"])
	assert.Equal(suite.T(), "Pilot Sport 4", product["model"])
	assert.Equal(suite.T(), "MIC-PS4-225-45-17", product["sku"])
	assert.Equal(suite.T(), 150.00, product["costPrice"])
	assert.Equal(suite.T(), 200.00, product["sellingPrice"])
	assert.Equal(suite.T(), float64(25), product["quantityOnHand"])
	assert.Equal(suite.T(), float64(5), product["lowStockThreshold"])
	assert.True(suite.T(), product["isActive"].(bool))
}

// TestCreateWheelProduct tests creating a new wheel product
func (suite *ProductCRUDTestSuite) TestCreateWheelProduct() {
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
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Verify successful creation
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Product created successfully", response["message"])

	product := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "Wheel", product["type"])
	assert.Equal(suite.T(), "Enkei", product["brand"])
	assert.Equal(suite.T(), "Racing RPF1", product["model"])
	assert.Equal(suite.T(), "ENK-RPF1-17-8.5-35", product["sku"])
	assert.Equal(suite.T(), 300.00, product["costPrice"])
	assert.Equal(suite.T(), 400.00, product["sellingPrice"])
	assert.Equal(suite.T(), float64(10), product["quantityOnHand"])
}

// TestCreateProductValidationErrors tests product creation with validation errors
func (suite *ProductCRUDTestSuite) TestCreateProductValidationErrors() {
	testCases := []struct {
		name           string
		productRequest map[string]interface{}
		expectedStatus int
		expectedErrors []string
	}{
		{
			name: "Missing required fields",
			productRequest: map[string]interface{}{
				"type": "Tire",
				// Missing brand, model, sku, prices
			},
			expectedStatus: http.StatusBadRequest,
			expectedErrors: []string{"brand", "model", "sku", "costPrice", "sellingPrice"},
		},
		{
			name: "Invalid product type",
			productRequest: map[string]interface{}{
				"type":         "InvalidType",
				"brand":        "Test",
				"model":        "Test",
				"sku":          "TEST-001",
				"costPrice":    100.00,
				"sellingPrice": 150.00,
			},
			expectedStatus: http.StatusBadRequest,
			expectedErrors: []string{"type"},
		},
		{
			name: "Negative prices",
			productRequest: map[string]interface{}{
				"type":         "Tire",
				"brand":        "Test",
				"model":        "Test",
				"sku":          "TEST-001",
				"costPrice":    -10.00,
				"sellingPrice": -5.00,
			},
			expectedStatus: http.StatusBadRequest,
			expectedErrors: []string{"costPrice", "sellingPrice"},
		},
		{
			name: "Negative quantity",
			productRequest: map[string]interface{}{
				"type":           "Tire",
				"brand":          "Test",
				"model":          "Test",
				"sku":            "TEST-001",
				"costPrice":      100.00,
				"sellingPrice":   150.00,
				"quantityOnHand": -5,
			},
			expectedStatus: http.StatusBadRequest,
			expectedErrors: []string{"quantityOnHand"},
		},
		{
			name: "Empty brand and model",
			productRequest: map[string]interface{}{
				"type":         "Tire",
				"brand":        "",
				"model":        "",
				"sku":          "TEST-001",
				"costPrice":    100.00,
				"sellingPrice": 150.00,
			},
			expectedStatus: http.StatusBadRequest,
			expectedErrors: []string{"brand", "model"},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			jsonBody, _ := json.Marshal(tc.productRequest)
			req, _ := http.NewRequest("POST", "/v1/products", bytes.NewBuffer(jsonBody))
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

// TestGetProductById tests retrieving a product by ID
func (suite *ProductCRUDTestSuite) TestGetProductById() {
	// Test existing product
	req, _ := http.NewRequest("GET", "/v1/products/1", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Product retrieved successfully", response["message"])

	product := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), float64(1), product["id"])
	assert.Contains(suite.T(), product, "type")
	assert.Contains(suite.T(), product, "brand")
	assert.Contains(suite.T(), product, "model")
	assert.Contains(suite.T(), product, "sku")
	assert.Contains(suite.T(), product, "costPrice")
	assert.Contains(suite.T(), product, "sellingPrice")
	assert.Contains(suite.T(), product, "quantityOnHand")
}

// TestGetProductNotFound tests retrieving a non-existent product
func (suite *ProductCRUDTestSuite) TestGetProductNotFound() {
	req, _ := http.NewRequest("GET", "/v1/products/999", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.False(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["message"], "not found")
}

// TestUpdateProduct tests updating an existing product
func (suite *ProductCRUDTestSuite) TestUpdateProduct() {
	updateRequest := map[string]interface{}{
		"brand":        "Michelin Updated",
		"model":        "Pilot Sport 4 Updated",
		"costPrice":    160.00,
		"sellingPrice": 210.00,
		"description":  "Updated description",
	}

	jsonBody, _ := json.Marshal(updateRequest)
	req, _ := http.NewRequest("PUT", "/v1/products/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Product updated successfully", response["message"])

	product := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "Michelin Updated", product["brand"])
	assert.Equal(suite.T(), "Pilot Sport 4 Updated", product["model"])
	assert.Equal(suite.T(), 160.00, product["costPrice"])
	assert.Equal(suite.T(), 210.00, product["sellingPrice"])
	assert.Equal(suite.T(), "Updated description", product["description"])
}

// TestUpdateProductValidationErrors tests updating a product with validation errors
func (suite *ProductCRUDTestSuite) TestUpdateProductValidationErrors() {
	updateRequest := map[string]interface{}{
		"brand":        "",     // Invalid: empty brand
		"costPrice":    -10.00, // Invalid: negative price
		"sellingPrice": 0,      // Invalid: zero price
	}

	jsonBody, _ := json.Marshal(updateRequest)
	req, _ := http.NewRequest("PUT", "/v1/products/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.False(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["message"], "Validation failed")
}

// TestDeleteProduct tests deleting a product
func (suite *ProductCRUDTestSuite) TestDeleteProduct() {
	req, _ := http.NewRequest("DELETE", "/v1/products/1", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Product deleted successfully", response["message"])

	// Verify product is deleted by trying to get it
	req, _ = http.NewRequest("GET", "/v1/products/1", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// TestDeleteProductNotFound tests deleting a non-existent product
func (suite *ProductCRUDTestSuite) TestDeleteProductNotFound() {
	req, _ := http.NewRequest("DELETE", "/v1/products/999", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.False(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["message"], "not found")
}

// TestListProducts tests listing products with pagination
func (suite *ProductCRUDTestSuite) TestListProducts() {
	testCases := []struct {
		name     string
		query    string
		expected int
	}{
		{
			name:     "List all products",
			query:    "",
			expected: http.StatusOK,
		},
		{
			name:     "List products with pagination",
			query:    "?page=1&limit=10",
			expected: http.StatusOK,
		},
		{
			name:     "Filter by type",
			query:    "?type=Tire",
			expected: http.StatusOK,
		},
		{
			name:     "Filter by brand",
			query:    "?brand=Michelin",
			expected: http.StatusOK,
		},
		{
			name:     "Filter by stock status",
			query:    "?stockStatus=lowStock",
			expected: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			req, _ := http.NewRequest("GET", "/v1/products"+tc.query, nil)
			req.Header.Set("Authorization", "Bearer "+suite.token)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(suite.T(), tc.expected, w.Code)

			if tc.expected == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(suite.T(), err)

				assert.True(suite.T(), response["success"].(bool))
				assert.Equal(suite.T(), "Products retrieved successfully", response["message"])

				data := response["data"].(map[string]interface{})
				assert.Contains(suite.T(), data, "products")
				assert.Contains(suite.T(), data, "pagination")

				pagination := data["pagination"].(map[string]interface{})
				assert.Contains(suite.T(), pagination, "page")
				assert.Contains(suite.T(), pagination, "limit")
				assert.Contains(suite.T(), pagination, "total")
				assert.Contains(suite.T(), pagination, "totalPages")
			}
		})
	}
}

// TestProductCRUDWithoutAuth tests that all CRUD operations require authentication
func (suite *ProductCRUDTestSuite) TestProductCRUDWithoutAuth() {
	testCases := []struct {
		name string
		req  func() *http.Request
	}{
		{
			name: "Create product without auth",
			req: func() *http.Request {
				productRequest := map[string]interface{}{
					"type":         "Tire",
					"brand":        "Test",
					"model":        "Test",
					"sku":          "TEST-001",
					"costPrice":    100.00,
					"sellingPrice": 150.00,
				}
				jsonBody, _ := json.Marshal(productRequest)
				req, _ := http.NewRequest("POST", "/v1/products", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
		},
		{
			name: "Get product without auth",
			req: func() *http.Request {
				req, _ := http.NewRequest("GET", "/v1/products/1", nil)
				return req
			},
		},
		{
			name: "Update product without auth",
			req: func() *http.Request {
				updateRequest := map[string]interface{}{
					"brand": "Updated",
				}
				jsonBody, _ := json.Marshal(updateRequest)
				req, _ := http.NewRequest("PUT", "/v1/products/1", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
		},
		{
			name: "Delete product without auth",
			req: func() *http.Request {
				req, _ := http.NewRequest("DELETE", "/v1/products/1", nil)
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

// Helper function to get auth token
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

// setupTestProductRoutes sets up test routes for integration testing
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
			errors := []gin.H{}
			if req["brand"] == "" {
				errors = append(errors, gin.H{"field": "brand", "message": "Brand is required"})
			}
			if req["model"] == "" {
				errors = append(errors, gin.H{"field": "model", "message": "Model is required"})
			}
			if req["sku"] == "" {
				errors = append(errors, gin.H{"field": "sku", "message": "SKU is required"})
			}
			if req["costPrice"] == nil {
				errors = append(errors, gin.H{"field": "costPrice", "message": "Cost price is required"})
			} else if req["costPrice"].(float64) <= 0 {
				errors = append(errors, gin.H{"field": "costPrice", "message": "Cost price must be positive"})
			}
			if req["sellingPrice"] == nil {
				errors = append(errors, gin.H{"field": "sellingPrice", "message": "Selling price is required"})
			} else if req["sellingPrice"].(float64) <= 0 {
				errors = append(errors, gin.H{"field": "sellingPrice", "message": "Selling price must be positive"})
			}
			if req["type"] == "InvalidType" {
				errors = append(errors, gin.H{"field": "type", "message": "Invalid product type"})
			}
			if req["quantityOnHand"] != nil && req["quantityOnHand"].(float64) < 0 {
				errors = append(errors, gin.H{"field": "quantityOnHand", "message": "Quantity cannot be negative"})
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

			// Mock validation
			errors := []gin.H{}
			if req["brand"] == "" {
				errors = append(errors, gin.H{"field": "brand", "message": "Brand cannot be empty"})
			}
			if req["costPrice"] != nil && req["costPrice"].(float64) < 0 {
				errors = append(errors, gin.H{"field": "costPrice", "message": "Cost price cannot be negative"})
			}
			if req["sellingPrice"] != nil && req["sellingPrice"].(float64) <= 0 {
				errors = append(errors, gin.H{"field": "sellingPrice", "message": "Selling price must be positive"})
			}

			if len(errors) > 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Validation failed",
					"errors":  errors,
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
					"description":       req["description"],
					"costPrice":         req["costPrice"],
					"sellingPrice":      req["sellingPrice"],
					"quantityOnHand":    25,
					"lowStockThreshold": 5,
					"isActive":          true,
				},
			})
		})

		products.DELETE("/:id", func(c *gin.Context) {
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
				"message": "Product deleted successfully",
			})
		})
	}
}
