// Package handlers contains the HTTP delivery layer implementations for the TT Stock Backend API.
// It provides HTTP handlers that process requests, validate input, call service layer methods,
// and return appropriate HTTP responses following RESTful conventions.
package handlers

import (
	"net/http"

	"tt-stock-api/src/models"
	"tt-stock-api/src/services"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService services.AuthService
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login handles user login requests
// @Summary User login with phone number and PIN
// @Description Authenticate user with phone number and PIN, returns JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.UserLoginRequest true "Login credentials"
// @Success 200 {object} models.UserLoginResponse "Login successful"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Invalid credentials"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Call service layer
	response, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		statusCode := http.StatusUnauthorized
		if err.Error() == "account is deactivated" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Logout handles user logout requests
// @Summary User logout
// @Description Logout user and invalidate JWT token
// @Tags Authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} SuccessResponse "Logout successful"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get token from Authorization header
	token := h.extractToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Token is required",
		})
		return
	}

	// Call service layer
	err := h.authService.Logout(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Logout successful",
	})
}

// RefreshToken handles token refresh requests
// @Summary Refresh JWT token
// @Description Generate new JWT token using refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} models.UserLoginResponse "Token refreshed successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Invalid refresh token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Call service layer
	response, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Token refresh failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ValidateToken handles token validation requests
// @Summary Validate JWT token
// @Description Validate JWT token and return user information
// @Tags Authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.UserResponse "Token is valid"
// @Failure 401 {object} ErrorResponse "Invalid token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/auth/validate [get]
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	// Get token from Authorization header
	token := h.extractToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Token is required",
		})
		return
	}

	// Call service layer
	user, err := h.authService.ValidateToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Token validation failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// GetProfile handles user profile requests
// @Summary Get user profile
// @Description Get current user's profile information
// @Tags Authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.UserResponse "User profile"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Get user from context (set by auth middleware)
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not found in context",
		})
		return
	}

	userModel, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Invalid user type in context",
		})
		return
	}

	c.JSON(http.StatusOK, userModel.ToResponse())
}

// RevokeAllSessions handles revoking all user sessions
// @Summary Revoke all user sessions
// @Description Revoke all active sessions for the current user
// @Tags Authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} SuccessResponse "All sessions revoked"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/auth/revoke-all [post]
func (h *AuthHandler) RevokeAllSessions(c *gin.Context) {
	// Get user from context (set by auth middleware)
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not found in context",
		})
		return
	}

	userModel, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Invalid user type in context",
		})
		return
	}

	// Call service layer
	err := h.authService.RevokeAllUserSessions(c.Request.Context(), userModel.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to revoke sessions",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "All sessions revoked successfully",
	})
}

// Helper methods

// extractToken extracts JWT token from Authorization header
func (h *AuthHandler) extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check if header starts with "Bearer "
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return ""
	}

	return authHeader[len(bearerPrefix):]
}

// Response types

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}
