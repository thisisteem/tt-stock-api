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

// ProductSearchTestSuite tests product search functionality
// This includes basic search, advanced search, and filtering capabilities
type ProductSearchTestSuite struct {
	suite.Suite
	router *gin.Engine
	token  string
}

// SetupSuite runs once before all tests in the suite
func (suite *ProductSearchTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	setupTestProductSearchRoutes(suite.router)
	suite.token = getAuthToken(suite.router)
}

// TestProductSearchTestSuite runs the test suite
func TestProductSearchTestSuite(t *testing.T) {
	suite.Run(t, new(ProductSearchTestSuite))
}

// TestBasicProductSearch tests basic product search functionality
func (suite *ProductSearchTestSuite) TestBasicProductSearch() {
	searchRequest := map[string]interface{}{
		"query": "Michelin",
		"page":  1,
		"limit": 20,
	}

	jsonBody, _ := json.Marshal(searchRequest)
	req, _ := http.NewRequest("POST", "/v1/products/search", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Verify successful search
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Search completed successfully", response["message"])

	data := response["data"].(map[string]interface{})
	_ = data["products"].([]interface{})
	totalCount := data["totalCount"].(float64)
	searchCriteria := data["searchCriteria"].(map[string]interface{})

	assert.GreaterOrEqual(suite.T(), totalCount, float64(0))
	assert.Equal(suite.T(), "Michelin", searchCriteria["query"])

	// Verify product structure if products exist
	products := data["products"].([]interface{})
	if len(products) > 0 {
		product := products[0].(map[string]interface{})
		assert.Contains(suite.T(), product, "id")
		assert.Contains(suite.T(), product, "type")
		assert.Contains(suite.T(), product, "brand")
		assert.Contains(suite.T(), product, "model")
		assert.Contains(suite.T(), product, "sku")
	}
}

// TestAdvancedTireSearch tests advanced tire search with specifications
func (suite *ProductSearchTestSuite) TestAdvancedTireSearch() {
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
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Verify successful search
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Search completed successfully", response["message"])

	data := response["data"].(map[string]interface{})
	_ = data["products"].([]interface{})
	searchCriteria := data["searchCriteria"].(map[string]interface{})

	assert.Equal(suite.T(), "Tire", searchCriteria["type"])
	assert.Equal(suite.T(), "available", searchCriteria["stockStatus"])

	// Verify tire specifications in search criteria
	tireSpecs := searchCriteria["tireSpecs"].(map[string]interface{})
	assert.Equal(suite.T(), "225", tireSpecs["width"])
	assert.Equal(suite.T(), "45", tireSpecs["aspectRatio"])
	assert.Equal(suite.T(), "17", tireSpecs["diameter"])
}

// TestAdvancedWheelSearch tests advanced wheel search with specifications
func (suite *ProductSearchTestSuite) TestAdvancedWheelSearch() {
	searchRequest := map[string]interface{}{
		"type": "Wheel",
		"wheelSpecs": map[string]interface{}{
			"diameter":    "17",
			"width":       "8.5",
			"offset":      "35",
			"boltPattern": "5x114.3",
		},
		"stockStatus": "available",
		"page":        1,
		"limit":       20,
	}

	jsonBody, _ := json.Marshal(searchRequest)
	req, _ := http.NewRequest("POST", "/v1/products/search", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Verify successful search
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Search completed successfully", response["message"])

	data := response["data"].(map[string]interface{})
	searchCriteria := data["searchCriteria"].(map[string]interface{})

	assert.Equal(suite.T(), "Wheel", searchCriteria["type"])

	// Verify wheel specifications in search criteria
	wheelSpecs := searchCriteria["wheelSpecs"].(map[string]interface{})
	assert.Equal(suite.T(), "17", wheelSpecs["diameter"])
	assert.Equal(suite.T(), "8.5", wheelSpecs["width"])
	assert.Equal(suite.T(), "35", wheelSpecs["offset"])
	assert.Equal(suite.T(), "5x114.3", wheelSpecs["boltPattern"])
}

// TestSearchWithFilters tests search with various filters
func (suite *ProductSearchTestSuite) TestSearchWithFilters() {
	testCases := []struct {
		name           string
		searchRequest  map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Search by brand",
			searchRequest: map[string]interface{}{
				"brand": "Michelin",
				"page":  1,
				"limit": 20,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Search by price range",
			searchRequest: map[string]interface{}{
				"minPrice": 100.00,
				"maxPrice": 300.00,
				"page":     1,
				"limit":    20,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Search by stock status",
			searchRequest: map[string]interface{}{
				"stockStatus": "lowStock",
				"page":        1,
				"limit":       20,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Search by multiple criteria",
			searchRequest: map[string]interface{}{
				"type":        "Tire",
				"brand":       "Michelin",
				"stockStatus": "available",
				"minPrice":    150.00,
				"maxPrice":    250.00,
				"page":        1,
				"limit":       20,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Search with pagination",
			searchRequest: map[string]interface{}{
				"query": "tire",
				"page":  2,
				"limit": 10,
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			jsonBody, _ := json.Marshal(tc.searchRequest)
			req, _ := http.NewRequest("POST", "/v1/products/search", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.token)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(suite.T(), tc.expectedStatus, w.Code)

			if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(suite.T(), err)

				assert.True(suite.T(), response["success"].(bool))
				assert.Equal(suite.T(), "Search completed successfully", response["message"])

				data := response["data"].(map[string]interface{})
				assert.Contains(suite.T(), data, "products")
				assert.Contains(suite.T(), data, "totalCount")
				assert.Contains(suite.T(), data, "searchCriteria")
			}
		})
	}
}

// TestSearchValidationErrors tests search with validation errors
func (suite *ProductSearchTestSuite) TestSearchValidationErrors() {
	testCases := []struct {
		name           string
		searchRequest  map[string]interface{}
		expectedStatus int
		expectedErrors []string
	}{
		{
			name: "Invalid page number",
			searchRequest: map[string]interface{}{
				"query": "test",
				"page":  0, // Invalid: page must be >= 1
				"limit": 20,
			},
			expectedStatus: http.StatusBadRequest,
			expectedErrors: []string{"page"},
		},
		{
			name: "Invalid limit",
			searchRequest: map[string]interface{}{
				"query": "test",
				"page":  1,
				"limit": 0, // Invalid: limit must be >= 1
			},
			expectedStatus: http.StatusBadRequest,
			expectedErrors: []string{"limit"},
		},
		{
			name: "Invalid price range",
			searchRequest: map[string]interface{}{
				"minPrice": 300.00,
				"maxPrice": 100.00, // Invalid: maxPrice < minPrice
				"page":     1,
				"limit":    20,
			},
			expectedStatus: http.StatusBadRequest,
			expectedErrors: []string{"priceRange"},
		},
		{
			name: "Invalid stock status",
			searchRequest: map[string]interface{}{
				"stockStatus": "InvalidStatus",
				"page":        1,
				"limit":       20,
			},
			expectedStatus: http.StatusBadRequest,
			expectedErrors: []string{"stockStatus"},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			jsonBody, _ := json.Marshal(tc.searchRequest)
			req, _ := http.NewRequest("POST", "/v1/products/search", bytes.NewBuffer(jsonBody))
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

// TestSearchPerformance tests search performance with large result sets
func (suite *ProductSearchTestSuite) TestSearchPerformance() {
	// Test search with large limit to verify performance
	searchRequest := map[string]interface{}{
		"query": "tire",
		"page":  1,
		"limit": 100, // Large limit to test performance
	}

	jsonBody, _ := json.Marshal(searchRequest)
	req, _ := http.NewRequest("POST", "/v1/products/search", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Verify successful search
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Search completed successfully", response["message"])

	data := response["data"].(map[string]interface{})
	products := data["products"].([]interface{})
	totalCount := data["totalCount"].(float64)

	// Verify pagination works correctly
	assert.LessOrEqual(suite.T(), len(products), 100)
	assert.GreaterOrEqual(suite.T(), totalCount, float64(len(products)))
}

// TestSearchWithoutAuth tests that search requires authentication
func (suite *ProductSearchTestSuite) TestSearchWithoutAuth() {
	searchRequest := map[string]interface{}{
		"query": "test",
		"page":  1,
		"limit": 20,
	}

	jsonBody, _ := json.Marshal(searchRequest)
	req, _ := http.NewRequest("POST", "/v1/products/search", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.False(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["message"], "Authorization header required")
}

// TestSearchEmptyResults tests search that returns no results
func (suite *ProductSearchTestSuite) TestSearchEmptyResults() {
	searchRequest := map[string]interface{}{
		"query": "nonexistentproduct",
		"page":  1,
		"limit": 20,
	}

	jsonBody, _ := json.Marshal(searchRequest)
	req, _ := http.NewRequest("POST", "/v1/products/search", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Verify successful search with no results
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Search completed successfully", response["message"])

	data := response["data"].(map[string]interface{})
	products := data["products"].([]interface{})
	totalCount := data["totalCount"].(float64)

	assert.Equal(suite.T(), 0, len(products))
	assert.Equal(suite.T(), float64(0), totalCount)
}

// setupTestProductSearchRoutes sets up test routes for integration testing
// This is a placeholder that will be replaced with actual implementation
func setupTestProductSearchRoutes(router *gin.Engine) {
	// Setup auth routes first
	setupTestAuthRoutes(router)

	// Placeholder product search routes that return mock responses
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
		products.POST("/search", func(c *gin.Context) {
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
			if req["page"] != nil && req["page"].(float64) < 1 {
				errors = append(errors, gin.H{"field": "page", "message": "Page must be >= 1"})
			}
			if req["limit"] != nil && req["limit"].(float64) < 1 {
				errors = append(errors, gin.H{"field": "limit", "message": "Limit must be >= 1"})
			}
			if req["minPrice"] != nil && req["maxPrice"] != nil && req["minPrice"].(float64) > req["maxPrice"].(float64) {
				errors = append(errors, gin.H{"field": "priceRange", "message": "Min price cannot be greater than max price"})
			}
			if req["stockStatus"] == "InvalidStatus" {
				errors = append(errors, gin.H{"field": "stockStatus", "message": "Invalid stock status"})
			}

			if len(errors) > 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Validation failed",
					"errors":  errors,
				})
				return
			}

			// Mock search results
			products := []gin.H{}
			if req["query"] != "nonexistentproduct" {
				products = append(products, gin.H{
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
					"stockStatus":       "available",
					"isActive":          true,
				})
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Search completed successfully",
				"data": gin.H{
					"products":       products,
					"totalCount":     len(products),
					"searchCriteria": req,
				},
			})
		})
	}
}
