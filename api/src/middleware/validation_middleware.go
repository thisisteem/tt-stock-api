// Package middleware contains HTTP middleware implementations for the TT Stock Backend API.
// It provides middleware functions that handle cross-cutting concerns like authentication,
// logging, security, and request processing in the HTTP delivery layer.
package middleware

import (
	"net/http"
	"strings"

	"tt-stock-api/src/models"
	"tt-stock-api/src/validators"

	"github.com/gin-gonic/gin"
)

// ValidationMiddleware handles input validation for HTTP requests
type ValidationMiddleware struct {
	authValidator    *validators.AuthValidator
	productValidator *validators.ProductValidator
	stockValidator   *validators.StockValidator
}

// NewValidationMiddleware creates a new ValidationMiddleware instance
func NewValidationMiddleware() *ValidationMiddleware {
	return &ValidationMiddleware{
		authValidator:    validators.NewAuthValidator(),
		productValidator: validators.NewProductValidator(),
		stockValidator:   validators.NewStockValidator(),
	}
}

// ValidateAuthRequest validates authentication-related requests
func (m *ValidationMiddleware) ValidateAuthRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		var validationErrors []string

		// Determine request type based on path
		switch c.Request.URL.Path {
		case "/v1/auth/login":
			validationErrors = m.validateLoginRequest(c)
		case "/v1/auth/refresh":
			validationErrors = m.validateRefreshTokenRequest(c)
		case "/v1/auth/change-pin":
			validationErrors = m.validatePINChangeRequest(c)
		}

		if len(validationErrors) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Validation failed",
				"errors":  validationErrors,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateProductRequest validates product-related requests
func (m *ValidationMiddleware) ValidateProductRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		var validationErrors []string

		// Determine request type based on method and path
		switch {
		case c.Request.Method == "POST" && c.Request.URL.Path == "/v1/products":
			validationErrors = m.validateProductCreateRequest(c)
		case c.Request.Method == "PUT" && m.isProductUpdatePath(c.Request.URL.Path):
			validationErrors = m.validateProductUpdateRequest(c)
		case c.Request.Method == "POST" && c.Request.URL.Path == "/v1/products/search":
			validationErrors = m.validateProductSearchRequest(c)
		case c.Request.Method == "POST" && m.isProductSpecificationPath(c.Request.URL.Path):
			validationErrors = m.validateProductSpecificationRequest(c)
		}

		if len(validationErrors) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Validation failed",
				"errors":  validationErrors,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateStockRequest validates stock-related requests
func (m *ValidationMiddleware) ValidateStockRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		var validationErrors []string

		// Determine request type based on method and path
		switch {
		case c.Request.Method == "POST" && c.Request.URL.Path == "/v1/stock/movements":
			validationErrors = m.validateStockMovementCreateRequest(c)
		case c.Request.Method == "PUT" && m.isStockMovementUpdatePath(c.Request.URL.Path):
			validationErrors = m.validateStockMovementUpdateRequest(c)
		case c.Request.Method == "GET" && c.Request.URL.Path == "/v1/stock/movements":
			validationErrors = m.validateStockMovementListRequest(c)
		case c.Request.Method == "POST" && c.Request.URL.Path == "/v1/stock/sales":
			validationErrors = m.validateSaleRequest(c)
		case c.Request.Method == "POST" && c.Request.URL.Path == "/v1/stock/incoming":
			validationErrors = m.validateIncomingStockRequest(c)
		case c.Request.Method == "POST" && c.Request.URL.Path == "/v1/stock/adjustments":
			validationErrors = m.validateStockAdjustmentRequest(c)
		case c.Request.Method == "POST" && c.Request.URL.Path == "/v1/stock/returns":
			validationErrors = m.validateReturnRequest(c)
		}

		if len(validationErrors) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Validation failed",
				"errors":  validationErrors,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateAlertRequest validates alert-related requests
func (m *ValidationMiddleware) ValidateAlertRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		var validationErrors []string

		// Determine request type based on method and path
		switch {
		case c.Request.Method == "POST" && c.Request.URL.Path == "/v1/alerts":
			validationErrors = m.validateAlertCreateRequest(c)
		case c.Request.Method == "PUT" && m.isAlertUpdatePath(c.Request.URL.Path):
			validationErrors = m.validateAlertUpdateRequest(c)
		case c.Request.Method == "GET" && c.Request.URL.Path == "/v1/alerts":
			validationErrors = m.validateAlertListRequest(c)
		}

		if len(validationErrors) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Validation failed",
				"errors":  validationErrors,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateUserRequest validates user-related requests
func (m *ValidationMiddleware) ValidateUserRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		var validationErrors []string

		// Determine request type based on method and path
		switch {
		case c.Request.Method == "POST" && c.Request.URL.Path == "/v1/users":
			validationErrors = m.validateUserCreateRequest(c)
		case c.Request.Method == "PUT" && m.isUserUpdatePath(c.Request.URL.Path):
			validationErrors = m.validateUserUpdateRequest(c)
		}

		if len(validationErrors) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Validation failed",
				"errors":  validationErrors,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Generic validation helper methods

// validateLoginRequest validates login request
func (m *ValidationMiddleware) validateLoginRequest(c *gin.Context) []string {
	var req models.UserLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		return []string{err.Error()}
	}

	if err := m.authValidator.ValidateLoginRequest(&req); err != nil {
		return []string{err.Error()}
	}

	return nil
}

// validateRefreshTokenRequest validates refresh token request
func (m *ValidationMiddleware) validateRefreshTokenRequest(c *gin.Context) []string {
	var req validators.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		return []string{err.Error()}
	}

	if err := m.authValidator.ValidateRefreshTokenRequest(&req); err != nil {
		return []string{err.Error()}
	}

	return nil
}

// validatePINChangeRequest validates PIN change request
func (m *ValidationMiddleware) validatePINChangeRequest(c *gin.Context) []string {
	var req validators.PINChangeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		return []string{err.Error()}
	}

	if err := m.authValidator.ValidatePINChangeRequest(&req); err != nil {
		return []string{err.Error()}
	}

	return nil
}

// validateProductCreateRequest validates product creation request
func (m *ValidationMiddleware) validateProductCreateRequest(c *gin.Context) []string {
	var req models.ProductCreateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		return []string{err.Error()}
	}

	if err := m.productValidator.ValidateProductCreateRequest(&req); err != nil {
		return []string{err.Error()}
	}

	return nil
}

// validateProductUpdateRequest validates product update request
func (m *ValidationMiddleware) validateProductUpdateRequest(c *gin.Context) []string {
	var req models.ProductUpdateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		return []string{err.Error()}
	}

	if err := m.productValidator.ValidateProductUpdateRequest(&req); err != nil {
		return []string{err.Error()}
	}

	return nil
}

// validateProductSearchRequest validates product search request
func (m *ValidationMiddleware) validateProductSearchRequest(c *gin.Context) []string {
	var req models.ProductSearchRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		return []string{err.Error()}
	}

	if err := m.productValidator.ValidateProductSearchRequest(&req); err != nil {
		return []string{err.Error()}
	}

	return nil
}

// validateProductSpecificationRequest validates product specification request
func (m *ValidationMiddleware) validateProductSpecificationRequest(c *gin.Context) []string {
	var req models.ProductSpecificationCreateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		return []string{err.Error()}
	}

	if err := m.productValidator.ValidateProductSpecificationRequest(&req); err != nil {
		return []string{err.Error()}
	}

	return nil
}

// validateStockMovementCreateRequest validates stock movement creation request
func (m *ValidationMiddleware) validateStockMovementCreateRequest(c *gin.Context) []string {
	var req models.StockMovementCreateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		return []string{err.Error()}
	}

	if err := m.stockValidator.ValidateStockMovementCreateRequest(&req); err != nil {
		return []string{err.Error()}
	}

	return nil
}

// validateStockMovementUpdateRequest validates stock movement update request
func (m *ValidationMiddleware) validateStockMovementUpdateRequest(c *gin.Context) []string {
	var req models.StockMovementUpdateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		return []string{err.Error()}
	}

	if err := m.stockValidator.ValidateStockMovementUpdateRequest(&req); err != nil {
		return []string{err.Error()}
	}

	return nil
}

// validateStockMovementListRequest validates stock movement list request
func (m *ValidationMiddleware) validateStockMovementListRequest(c *gin.Context) []string {
	var req models.StockMovementListRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		return []string{err.Error()}
	}

	if err := m.stockValidator.ValidateStockMovementListRequest(&req); err != nil {
		return []string{err.Error()}
	}

	return nil
}

// validateSaleRequest validates sale request
func (m *ValidationMiddleware) validateSaleRequest(c *gin.Context) []string {
	var req models.SaleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		return []string{err.Error()}
	}

	if err := m.stockValidator.ValidateSaleRequest(&req); err != nil {
		return []string{err.Error()}
	}

	return nil
}

// validateIncomingStockRequest validates incoming stock request
func (m *ValidationMiddleware) validateIncomingStockRequest(c *gin.Context) []string {
	// Same as stock movement create request
	return m.validateStockMovementCreateRequest(c)
}

// validateStockAdjustmentRequest validates stock adjustment request
func (m *ValidationMiddleware) validateStockAdjustmentRequest(c *gin.Context) []string {
	// Same as stock movement create request
	return m.validateStockMovementCreateRequest(c)
}

// validateReturnRequest validates return request
func (m *ValidationMiddleware) validateReturnRequest(c *gin.Context) []string {
	// Same as stock movement create request
	return m.validateStockMovementCreateRequest(c)
}

// validateAlertCreateRequest validates alert creation request
func (m *ValidationMiddleware) validateAlertCreateRequest(c *gin.Context) []string {
	var req models.AlertCreateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		return []string{err.Error()}
	}

	// Basic validation - more complex validation would be in the alert validator
	if req.Title == "" {
		return []string{"title is required"}
	}
	if req.Message == "" {
		return []string{"message is required"}
	}
	if req.AlertType == "" {
		return []string{"alertType is required"}
	}

	return nil
}

// validateAlertUpdateRequest validates alert update request
func (m *ValidationMiddleware) validateAlertUpdateRequest(c *gin.Context) []string {
	var req models.AlertUpdateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		return []string{err.Error()}
	}

	// Basic validation
	if req.Title != nil && *req.Title == "" {
		return []string{"title cannot be empty"}
	}
	if req.Message != nil && *req.Message == "" {
		return []string{"message cannot be empty"}
	}

	return nil
}

// validateAlertListRequest validates alert list request
func (m *ValidationMiddleware) validateAlertListRequest(c *gin.Context) []string {
	var req models.AlertListRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		return []string{err.Error()}
	}

	return nil
}

// validateUserCreateRequest validates user creation request
func (m *ValidationMiddleware) validateUserCreateRequest(c *gin.Context) []string {
	var req models.UserCreateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		return []string{err.Error()}
	}

	if err := m.authValidator.ValidateUserCreateRequest(&req); err != nil {
		return []string{err.Error()}
	}

	return nil
}

// validateUserUpdateRequest validates user update request
func (m *ValidationMiddleware) validateUserUpdateRequest(c *gin.Context) []string {
	var req models.UserUpdateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		return []string{err.Error()}
	}

	if err := m.authValidator.ValidateUserUpdateRequest(&req); err != nil {
		return []string{err.Error()}
	}

	return nil
}

// Helper methods for path matching

// isProductUpdatePath checks if the path is a product update path
func (m *ValidationMiddleware) isProductUpdatePath(path string) bool {
	return len(path) > 10 && path[:10] == "/v1/products/" && path[len(path)-1] != '/'
}

// isProductSpecificationPath checks if the path is a product specification path
func (m *ValidationMiddleware) isProductSpecificationPath(path string) bool {
	return len(path) > 20 && path[:20] == "/v1/products/" && strings.Contains(path, "/specifications")
}

// isStockMovementUpdatePath checks if the path is a stock movement update path
func (m *ValidationMiddleware) isStockMovementUpdatePath(path string) bool {
	return len(path) > 20 && path[:20] == "/v1/stock/movements/" && path[len(path)-1] != '/'
}

// isAlertUpdatePath checks if the path is an alert update path
func (m *ValidationMiddleware) isAlertUpdatePath(path string) bool {
	return len(path) > 8 && path[:8] == "/v1/alerts/" && path[len(path)-1] != '/'
}

// isUserUpdatePath checks if the path is a user update path
func (m *ValidationMiddleware) isUserUpdatePath(path string) bool {
	return len(path) > 7 && path[:7] == "/v1/users/" && path[len(path)-1] != '/'
}

// Validation configuration constants
const (
	DefaultPageSize = 10
	MaxPageSize     = 100
	MinPageSize     = 1
	DefaultPage     = 1
)
