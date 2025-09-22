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

// AlertHandler handles alert-related HTTP requests
type AlertHandler struct {
	alertService services.AlertService
}

// NewAlertHandler creates a new AlertHandler instance
func NewAlertHandler(alertService services.AlertService) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
	}
}

// CreateAlert handles alert creation requests
// @Summary Create an alert
// @Description Create a new alert
// @Tags Alerts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.AlertCreateRequest true "Alert data"
// @Success 201 {object} models.AlertResponse "Alert created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/alerts [post]
func (h *AlertHandler) CreateAlert(c *gin.Context) {
	var req models.AlertCreateRequest
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
	response, err := h.alertService.CreateAlert(c.Request.Context(), &req, user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "insufficient permissions to create alerts" {
			statusCode = http.StatusForbidden
		} else if err.Error() == "product not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to create alert",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetAlert handles alert retrieval requests
// @Summary Get alert by ID
// @Description Get an alert by its ID
// @Tags Alerts
// @Security BearerAuth
// @Produce json
// @Param id path int true "Alert ID"
// @Success 200 {object} models.AlertResponse "Alert found"
// @Failure 400 {object} ErrorResponse "Invalid alert ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Alert not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/alerts/{id} [get]
func (h *AlertHandler) GetAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid alert ID",
			Message: "Alert ID must be a valid number",
		})
		return
	}

	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	response, err := h.alertService.GetAlert(c.Request.Context(), uint(id), user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "alert not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "insufficient permissions to view alerts" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to get alert",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateAlert handles alert update requests
// @Summary Update alert
// @Description Update an existing alert
// @Tags Alerts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Alert ID"
// @Param request body models.AlertUpdateRequest true "Alert update data"
// @Success 200 {object} models.AlertResponse "Alert updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 404 {object} ErrorResponse "Alert not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/alerts/{id} [put]
func (h *AlertHandler) UpdateAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid alert ID",
			Message: "Alert ID must be a valid number",
		})
		return
	}

	var req models.AlertUpdateRequest
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
	response, err := h.alertService.UpdateAlert(c.Request.Context(), uint(id), &req, user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "alert not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "insufficient permissions to update alerts" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to update alert",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteAlert handles alert deletion requests
// @Summary Delete alert
// @Description Delete an alert
// @Tags Alerts
// @Security BearerAuth
// @Produce json
// @Param id path int true "Alert ID"
// @Success 200 {object} SuccessResponse "Alert deleted successfully"
// @Failure 400 {object} ErrorResponse "Invalid alert ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 404 {object} ErrorResponse "Alert not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/alerts/{id} [delete]
func (h *AlertHandler) DeleteAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid alert ID",
			Message: "Alert ID must be a valid number",
		})
		return
	}

	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	err = h.alertService.DeleteAlert(c.Request.Context(), uint(id), user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "alert not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "insufficient permissions to delete alerts" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to delete alert",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Alert deleted successfully",
	})
}

// ListAlerts handles alert listing requests
// @Summary List alerts
// @Description Get a list of alerts with pagination and filtering
// @Tags Alerts
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param productId query int false "Filter by product ID"
// @Param userId query int false "Filter by user ID"
// @Param alertType query string false "Filter by alert type"
// @Param priority query string false "Filter by priority"
// @Param isRead query bool false "Filter by read status"
// @Param isActive query bool false "Filter by active status"
// @Success 200 {object} models.AlertListResponse "Alerts found"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/alerts [get]
func (h *AlertHandler) ListAlerts(c *gin.Context) {
	// Parse query parameters
	req := h.parseAlertListRequest(c)

	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	response, err := h.alertService.ListAlerts(c.Request.Context(), req, user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "insufficient permissions to list alerts" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to list alerts",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetUnreadAlerts handles unread alerts requests
// @Summary Get unread alerts
// @Description Get unread alerts for the current user
// @Tags Alerts
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.AlertResponse "Unread alerts"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/alerts/unread [get]
func (h *AlertHandler) GetUnreadAlerts(c *gin.Context) {
	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	response, err := h.alertService.GetUnreadAlerts(c.Request.Context(), user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "insufficient permissions to view unread alerts" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to get unread alerts",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// MarkAlertAsRead handles marking alert as read requests
// @Summary Mark alert as read
// @Description Mark an alert as read
// @Tags Alerts
// @Security BearerAuth
// @Produce json
// @Param id path int true "Alert ID"
// @Success 200 {object} SuccessResponse "Alert marked as read"
// @Failure 400 {object} ErrorResponse "Invalid alert ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Alert not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/alerts/{id}/read [post]
func (h *AlertHandler) MarkAlertAsRead(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid alert ID",
			Message: "Alert ID must be a valid number",
		})
		return
	}

	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	err = h.alertService.MarkAlertAsRead(c.Request.Context(), uint(id), user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "alert not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "insufficient permissions to mark alerts as read" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to mark alert as read",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Alert marked as read",
	})
}

// MarkAllAlertsAsRead handles marking all alerts as read requests
// @Summary Mark all alerts as read
// @Description Mark all alerts as read for the current user
// @Tags Alerts
// @Security BearerAuth
// @Produce json
// @Success 200 {object} SuccessResponse "All alerts marked as read"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/alerts/read-all [post]
func (h *AlertHandler) MarkAllAlertsAsRead(c *gin.Context) {
	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	err := h.alertService.MarkAllAlertsAsRead(c.Request.Context(), user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "insufficient permissions to mark alerts as read" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to mark all alerts as read",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "All alerts marked as read",
	})
}

// GetUnreadCount handles unread alert count requests
// @Summary Get unread alert count
// @Description Get the count of unread alerts for the current user
// @Tags Alerts
// @Security BearerAuth
// @Produce json
// @Success 200 {object} UnreadCountResponse "Unread alert count"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/alerts/unread-count [get]
func (h *AlertHandler) GetUnreadCount(c *gin.Context) {
	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	count, err := h.alertService.GetUnreadCount(c.Request.Context(), user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "insufficient permissions to view unread count" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to get unread count",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UnreadCountResponse{
		Count: count,
	})
}

// DeactivateAlert handles alert deactivation requests
// @Summary Deactivate alert
// @Description Deactivate an alert
// @Tags Alerts
// @Security BearerAuth
// @Produce json
// @Param id path int true "Alert ID"
// @Success 200 {object} SuccessResponse "Alert deactivated"
// @Failure 400 {object} ErrorResponse "Invalid alert ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 404 {object} ErrorResponse "Alert not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/alerts/{id}/deactivate [post]
func (h *AlertHandler) DeactivateAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid alert ID",
			Message: "Alert ID must be a valid number",
		})
		return
	}

	// Get user from context
	user := h.getUserFromContext(c)
	if user == nil {
		return
	}

	// Call service layer
	err = h.alertService.DeactivateAlert(c.Request.Context(), uint(id), user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "alert not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "insufficient permissions to deactivate alerts" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "Failed to deactivate alert",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Alert deactivated successfully",
	})
}

// Helper methods

// getUserFromContext extracts user from Gin context
func (h *AlertHandler) getUserFromContext(c *gin.Context) *models.User {
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

// parseAlertListRequest parses query parameters into AlertListRequest
func (h *AlertHandler) parseAlertListRequest(c *gin.Context) *models.AlertListRequest {
	req := &models.AlertListRequest{
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

	if alertType := c.Query("alertType"); alertType != "" {
		if at := models.AlertType(alertType); at == models.AlertTypeLowStock || at == models.AlertTypeOutOfStock || at == models.AlertTypeExpired || at == models.AlertTypeMaintenance || at == models.AlertTypeSystem {
			req.AlertType = &at
		}
	}

	if priority := c.Query("priority"); priority != "" {
		if p := models.AlertPriority(priority); p == models.AlertPriorityLow || p == models.AlertPriorityMedium || p == models.AlertPriorityHigh || p == models.AlertPriorityCritical {
			req.Priority = &p
		}
	}

	if isReadStr := c.Query("isRead"); isReadStr != "" {
		if isRead, err := strconv.ParseBool(isReadStr); err == nil {
			req.IsRead = &isRead
		}
	}

	if isActiveStr := c.Query("isActive"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			req.IsActive = &isActive
		}
	}

	return req
}

// parseOptionalUintParam parses an optional uint parameter from query string
func (h *AlertHandler) parseOptionalUintParam(c *gin.Context, param string) *uint {
	if paramStr := c.Query(param); paramStr != "" {
		if parsed, err := strconv.ParseUint(paramStr, 10, 32); err == nil {
			value := uint(parsed)
			return &value
		}
	}
	return nil
}

// Response types

// UnreadCountResponse represents an unread count response
type UnreadCountResponse struct {
	Count int64 `json:"count"`
}
