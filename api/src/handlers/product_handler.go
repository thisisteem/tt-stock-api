// Package handlers contains the HTTP delivery layer implementations for the TT Stock Backend API.
// It provides HTTP handlers that process requests, validate input, call service layer methods,
// and return appropriate HTTP responses following RESTful conventions.
package handlers

import (
	"net/http"
	"strconv"

	"tt-stock-api/src/models"
	"tt-stock-api/src/services"

	"github.com/gin-gonic/gin"
)

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
	productService services.ProductService
}

// NewProductHandler creates a new ProductHandler instance
func NewProductHandler(productService services.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// CreateProduct handles product creation requests
// @Summary Create a new product
// @Description Create a new product with specifications
// @Tags Products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.ProductCreateRequest true "Product data"
// @Success 201 {object} models.ProductResponse "Product created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req models.ProductCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	response, err := h.productService.CreateProduct(c.Request.Context(), &req, user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "insufficient permissions to create products" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to create product",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetProduct handles product retrieval requests
// @Summary Get product by ID
// @Description Get a product by its ID
// @Tags Products
// @Security BearerAuth
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} models.ProductResponse "Product found"
// @Failure 400 {object} ErrorResponse "Invalid product ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Product not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/products/{id} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid product ID",
			Message: "Product ID must be a valid number",
		})
		return
	}

	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	response, err := h.productService.GetProduct(c.Request.Context(), uint(id), user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "product not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "insufficient permissions to view products" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to get product",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetProductBySKU handles product retrieval by SKU requests
// @Summary Get product by SKU
// @Description Get a product by its SKU
// @Tags Products
// @Security BearerAuth
// @Produce json
// @Param sku path string true "Product SKU"
// @Success 200 {object} models.ProductResponse "Product found"
// @Failure 400 {object} ErrorResponse "Invalid SKU"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Product not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/products/sku/{sku} [get]
func (h *ProductHandler) GetProductBySKU(c *gin.Context) {
	sku := c.Param("sku")
	if sku == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid SKU",
			Message: "SKU is required",
		})
		return
	}

	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	response, err := h.productService.GetProductBySKU(c.Request.Context(), sku, user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "product not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "insufficient permissions to view products" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to get product",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateProduct handles product update requests
// @Summary Update product
// @Description Update an existing product
// @Tags Products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param request body models.ProductUpdateRequest true "Product update data"
// @Success 200 {object} models.ProductResponse "Product updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 404 {object} ErrorResponse "Product not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid product ID",
			Message: "Product ID must be a valid number",
		})
		return
	}

	var req models.ProductUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	response, err := h.productService.UpdateProduct(c.Request.Context(), uint(id), &req, user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "product not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "insufficient permissions to update products" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to update product",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteProduct handles product deletion requests
// @Summary Delete product
// @Description Delete a product (soft delete)
// @Tags Products
// @Security BearerAuth
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} SuccessResponse "Product deleted successfully"
// @Failure 400 {object} ErrorResponse "Invalid product ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 404 {object} ErrorResponse "Product not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid product ID",
			Message: "Product ID must be a valid number",
		})
		return
	}

	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	err = h.productService.DeleteProduct(c.Request.Context(), uint(id), user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "product not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "insufficient permissions to delete products" {
			statusCode = http.StatusForbidden
		} else if err.Error() == "cannot delete product with remaining stock" {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to delete product",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Product deleted successfully",
	})
}

// ListProducts handles product listing requests
// @Summary List products
// @Description Get a list of products with pagination and filtering
// @Tags Products
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param query query string false "Search query"
// @Param type query string false "Product type"
// @Param brand query string false "Brand filter"
// @Param stockStatus query string false "Stock status filter"
// @Param minPrice query number false "Minimum price"
// @Param maxPrice query number false "Maximum price"
// @Success 200 {object} models.ProductSearchResponse "Products found"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/products [get]
func (h *ProductHandler) ListProducts(c *gin.Context) {
	// Parse query parameters
	req := h.parseProductSearchRequest(c)

	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	response, err := h.productService.ListProducts(c.Request.Context(), req, user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "insufficient permissions to list products" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to list products",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SearchProducts handles product search requests
// @Summary Search products
// @Description Search products with advanced filters
// @Tags Products
// @Security BearerAuth
// @Produce json
// @Param request body models.ProductSearchRequest true "Search criteria"
// @Success 200 {object} models.ProductSearchResponse "Search results"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/products/search [post]
func (h *ProductHandler) SearchProducts(c *gin.Context) {
	var req models.ProductSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	response, err := h.productService.SearchProducts(c.Request.Context(), &req, user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "insufficient permissions to search products" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to search products",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetLowStockProducts handles low stock products requests
// @Summary Get low stock products
// @Description Get products with low stock levels
// @Tags Products
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.ProductResponse "Low stock products"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/products/low-stock [get]
func (h *ProductHandler) GetLowStockProducts(c *gin.Context) {
	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	response, err := h.productService.GetLowStockProducts(c.Request.Context(), user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "insufficient permissions to view low stock products" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to get low stock products",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetOutOfStockProducts handles out of stock products requests
// @Summary Get out of stock products
// @Description Get products that are out of stock
// @Tags Products
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.ProductResponse "Out of stock products"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/products/out-of-stock [get]
func (h *ProductHandler) GetOutOfStockProducts(c *gin.Context) {
	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	response, err := h.productService.GetOutOfStockProducts(c.Request.Context(), user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "insufficient permissions to view out of stock products" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to get out of stock products",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetProductStatistics handles product statistics requests
// @Summary Get product statistics
// @Description Get product statistics and analytics
// @Tags Products
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.ProductStatistics "Product statistics"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/products/statistics [get]
func (h *ProductHandler) GetProductStatistics(c *gin.Context) {
	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	response, err := h.productService.GetProductStatistics(c.Request.Context(), user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "insufficient permissions to view product statistics" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to get product statistics",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Helper methods

// getUserFromContext extracts user from Gin context
func (h *ProductHandler) getUserFromContext(c *gin.Context) *models.User {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not found in context",
		})
		return nil
	}

	userModel, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Invalid user type in context",
		})
		return nil
	}

	return userModel
}

// parseProductSearchRequest parses query parameters into ProductSearchRequest
func (h *ProductHandler) parseProductSearchRequest(c *gin.Context) *models.ProductSearchRequest {
	req := &models.ProductSearchRequest{
		Page:  1,
		Limit: 10,
	}

	// Parse pagination parameters
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	// Parse search parameters
	if query := c.Query("query"); query != "" {
		req.Query = &query
	}

	if productType := c.Query("type"); productType != "" {
		if pt := models.ProductType(productType); pt == models.ProductTypeTire || pt == models.ProductTypeWheel {
			req.Type = &pt
		}
	}

	if brand := c.Query("brand"); brand != "" {
		req.Brand = &brand
	}

	if stockStatus := c.Query("stockStatus"); stockStatus != "" {
		if ss := models.StockStatus(stockStatus); ss == models.StockStatusAvailable || ss == models.StockStatusLowStock || ss == models.StockStatusOutOfStock {
			req.StockStatus = &ss
		}
	}

	if minPriceStr := c.Query("minPrice"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			req.MinPrice = &minPrice
		}
	}

	if maxPriceStr := c.Query("maxPrice"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			req.MaxPrice = &maxPrice
		}
	}

	return req
}
