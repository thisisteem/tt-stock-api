// Package handlers contains the HTTP delivery layer implementations for the TT Stock Backend API.
// It provides HTTP handlers that process requests, validate input, call service layer methods,
// and return appropriate HTTP responses following RESTful conventions.
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"tt-stock-api/src/models"
	"tt-stock-api/src/services"

	"github.com/gin-gonic/gin"
)

// StockHandler handles stock-related HTTP requests
type StockHandler struct {
	stockService services.StockService
}

// NewStockHandler creates a new StockHandler instance
func NewStockHandler(stockService services.StockService) *StockHandler {
	return &StockHandler{
		stockService: stockService,
	}
}

// CreateStockMovement handles stock movement creation requests
// @Summary Create stock movement
// @Description Create a new stock movement (incoming, outgoing, adjustment, etc.)
// @Tags Stock
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.StockMovementCreateRequest true "Stock movement data"
// @Success 201 {object} models.StockMovementResponse "Stock movement created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/stock/movements [post]
func (h *StockHandler) CreateStockMovement(c *gin.Context) {
	var req models.StockMovementCreateRequest
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
	response, err := h.stockService.CreateStockMovement(c.Request.Context(), &req, user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "insufficient permissions to create stock movements" {
			statusCode = http.StatusForbidden
		} else if err.Error() == "product not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "insufficient stock for this movement" {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to create stock movement",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetStockMovement handles stock movement retrieval requests
// @Summary Get stock movement by ID
// @Description Get a stock movement by its ID
// @Tags Stock
// @Security BearerAuth
// @Produce json
// @Param id path int true "Stock movement ID"
// @Success 200 {object} models.StockMovementResponse "Stock movement found"
// @Failure 400 {object} ErrorResponse "Invalid movement ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Stock movement not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/stock/movements/{id} [get]
func (h *StockHandler) GetStockMovement(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid movement ID",
			Message: "Movement ID must be a valid number",
		})
		return
	}

	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	response, err := h.stockService.GetStockMovement(c.Request.Context(), uint(id), user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "stock movement not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "insufficient permissions to view stock movements" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to get stock movement",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateStockMovement handles stock movement update requests
// @Summary Update stock movement
// @Description Update an existing stock movement
// @Tags Stock
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Stock movement ID"
// @Param request body models.StockMovementUpdateRequest true "Stock movement update data"
// @Success 200 {object} models.StockMovementResponse "Stock movement updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 404 {object} ErrorResponse "Stock movement not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/stock/movements/{id} [put]
func (h *StockHandler) UpdateStockMovement(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid movement ID",
			Message: "Movement ID must be a valid number",
		})
		return
	}

	var req models.StockMovementUpdateRequest
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
	response, err := h.stockService.UpdateStockMovement(c.Request.Context(), uint(id), &req, user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "stock movement not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "insufficient permissions to update stock movements" {
			statusCode = http.StatusForbidden
		} else if err.Error() == "insufficient stock for this movement" {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to update stock movement",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteStockMovement handles stock movement deletion requests
// @Summary Delete stock movement
// @Description Delete a stock movement
// @Tags Stock
// @Security BearerAuth
// @Produce json
// @Param id path int true "Stock movement ID"
// @Success 200 {object} SuccessResponse "Stock movement deleted successfully"
// @Failure 400 {object} ErrorResponse "Invalid movement ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 404 {object} ErrorResponse "Stock movement not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/stock/movements/{id} [delete]
func (h *StockHandler) DeleteStockMovement(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid movement ID",
			Message: "Movement ID must be a valid number",
		})
		return
	}

	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	err = h.stockService.DeleteStockMovement(c.Request.Context(), uint(id), user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "stock movement not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "insufficient permissions to delete stock movements" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to delete stock movement",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Stock movement deleted successfully",
	})
}

// ListStockMovements handles stock movement listing requests
// @Summary List stock movements
// @Description Get a list of stock movements with pagination and filtering
// @Tags Stock
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param productId query int false "Filter by product ID"
// @Param userId query int false "Filter by user ID"
// @Param movementType query string false "Filter by movement type"
// @Param startDate query string false "Start date (ISO 8601)"
// @Param endDate query string false "End date (ISO 8601)"
// @Success 200 {object} models.StockMovementListResponse "Stock movements found"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/stock/movements [get]
func (h *StockHandler) ListStockMovements(c *gin.Context) {
	// Parse query parameters
	req := h.parseStockMovementListRequest(c)

	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	response, err := h.stockService.ListStockMovements(c.Request.Context(), req, user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "insufficient permissions to list stock movements" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to list stock movements",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ProcessSale handles sale processing requests
// @Summary Process a sale
// @Description Process a product sale and update stock
// @Tags Stock
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.SaleRequest true "Sale data"
// @Success 201 {object} models.SaleResponse "Sale processed successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/stock/sales [post]
func (h *StockHandler) ProcessSale(c *gin.Context) {
	var req models.SaleRequest
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
	response, err := h.stockService.ProcessSale(c.Request.Context(), &req, user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "insufficient permissions to process sales" {
			statusCode = http.StatusForbidden
		} else if err.Error() == "product not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "insufficient stock for sale" {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to process sale",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// ProcessIncomingStock handles incoming stock processing requests
// @Summary Process incoming stock
// @Description Process incoming stock and update inventory
// @Tags Stock
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.StockMovementCreateRequest true "Incoming stock data"
// @Success 201 {object} models.StockMovementResponse "Incoming stock processed successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/stock/incoming [post]
func (h *StockHandler) ProcessIncomingStock(c *gin.Context) {
	var req models.StockMovementCreateRequest
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
	response, err := h.stockService.ProcessIncomingStock(c.Request.Context(), &req, user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "insufficient permissions to create stock movements" {
			statusCode = http.StatusForbidden
		} else if err.Error() == "product not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to process incoming stock",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// ProcessStockAdjustment handles stock adjustment requests
// @Summary Process stock adjustment
// @Description Process a stock adjustment (manual correction)
// @Tags Stock
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.StockMovementCreateRequest true "Stock adjustment data"
// @Success 201 {object} models.StockMovementResponse "Stock adjustment processed successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/stock/adjustments [post]
func (h *StockHandler) ProcessStockAdjustment(c *gin.Context) {
	var req models.StockMovementCreateRequest
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
	response, err := h.stockService.ProcessStockAdjustment(c.Request.Context(), &req, user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "insufficient permissions to create stock movements" {
			statusCode = http.StatusForbidden
		} else if err.Error() == "product not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "adjustment would result in negative stock" {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to process stock adjustment",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// ProcessReturn handles product return requests
// @Summary Process product return
// @Description Process a product return and update stock
// @Tags Stock
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.StockMovementCreateRequest true "Return data"
// @Success 201 {object} models.StockMovementResponse "Return processed successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/stock/returns [post]
func (h *StockHandler) ProcessReturn(c *gin.Context) {
	var req models.StockMovementCreateRequest
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
	response, err := h.stockService.ProcessReturn(c.Request.Context(), &req, user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "insufficient permissions to create stock movements" {
			statusCode = http.StatusForbidden
		} else if err.Error() == "product not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to process return",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetStockMovementSummary handles stock movement summary requests
// @Summary Get stock movement summary
// @Description Get summary statistics for stock movements
// @Tags Stock
// @Security BearerAuth
// @Produce json
// @Param productId query int false "Filter by product ID"
// @Param startDate query string false "Start date (ISO 8601)"
// @Param endDate query string false "End date (ISO 8601)"
// @Success 200 {object} models.StockMovementSummary "Stock movement summary"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/stock/movements/summary [get]
func (h *StockHandler) GetStockMovementSummary(c *gin.Context) {
	// Parse query parameters
	productID := h.parseOptionalUintParam(c, "productId")
	startDate := h.parseOptionalTimeParam(c, "startDate")
	endDate := h.parseOptionalTimeParam(c, "endDate")

	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	response, err := h.stockService.GetStockMovementSummary(c.Request.Context(), productID, startDate, endDate, user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "insufficient permissions to view movement summary" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to get stock movement summary",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetRecentMovements handles recent stock movements requests
// @Summary Get recent stock movements
// @Description Get recent stock movements
// @Tags Stock
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Number of recent movements" default(10)
// @Success 200 {array} models.StockMovementResponse "Recent stock movements"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/stock/movements/recent [get]
func (h *StockHandler) GetRecentMovements(c *gin.Context) {
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	response, err := h.stockService.GetRecentMovements(c.Request.Context(), limit, user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "insufficient permissions to view recent movements" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to get recent movements",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Helper methods

// getUserFromContext extracts user from Gin context
func (h *StockHandler) getUserFromContext(c *gin.Context) *models.User {
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

// parseStockMovementListRequest parses query parameters into StockMovementListRequest
func (h *StockHandler) parseStockMovementListRequest(c *gin.Context) *models.StockMovementListRequest {
	req := &models.StockMovementListRequest{
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

	// Parse filter parameters
	if productID := h.parseOptionalUintParam(c, "productId"); productID != nil {
		req.ProductID = productID
	}

	if userID := h.parseOptionalUintParam(c, "userId"); userID != nil {
		req.UserID = userID
	}

	if movementType := c.Query("movementType"); movementType != "" {
		if mt := models.MovementType(movementType); mt == models.MovementTypeIncoming || mt == models.MovementTypeOutgoing || mt == models.MovementTypeSale || mt == models.MovementTypeAdjustment || mt == models.MovementTypeReturn {
			req.MovementType = &mt
		}
	}

	if startDate := h.parseOptionalTimeParam(c, "startDate"); startDate != nil {
		req.StartDate = startDate
	}

	if endDate := h.parseOptionalTimeParam(c, "endDate"); endDate != nil {
		req.EndDate = endDate
	}

	return req
}

// parseOptionalUintParam parses an optional uint parameter from query string
func (h *StockHandler) parseOptionalUintParam(c *gin.Context, param string) *uint {
	if paramStr := c.Query(param); paramStr != "" {
		if parsed, err := strconv.ParseUint(paramStr, 10, 32); err == nil {
			value := uint(parsed)
			return &value
		}
	}
	return nil
}

// parseOptionalTimeParam parses an optional time parameter from query string
func (h *StockHandler) parseOptionalTimeParam(c *gin.Context, param string) *time.Time {
	if paramStr := c.Query(param); paramStr != "" {
		if parsed, err := time.Parse(time.RFC3339, paramStr); err == nil {
			return &parsed
		}
	}
	return nil
}
