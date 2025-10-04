package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"tt-stock-api/pkg/response"
)

// LoginRequest represents the request body for login endpoint
type LoginRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
	Pin         string `json:"pin" validate:"required"`
}

// RefreshRequest represents the request body for refresh token endpoint
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Handler defines the interface for authentication HTTP handlers
type Handler interface {
	Login(c *fiber.Ctx) error
	Refresh(c *fiber.Ctx) error
	Logout(c *fiber.Ctx) error
}

// handler implements the Handler interface
type handler struct {
	authService Service
}

// NewHandler creates a new authentication handler instance
func NewHandler(authService Service) Handler {
	return &handler{
		authService: authService,
	}
}

// Login handles POST /auth/login endpoint
// Authenticates user with phone number and PIN, returns access and refresh tokens
func (h *handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	
	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return response.SendValidationError(c, "Invalid request body")
	}

	// Validate required fields
	if req.PhoneNumber == "" {
		return response.SendValidationError(c, "Phone number is required")
	}
	if req.Pin == "" {
		return response.SendValidationError(c, "PIN is required")
	}

	// Authenticate user
	user, err := h.authService.AuthenticateUser(req.PhoneNumber, req.Pin)
	if err != nil {
		return response.SendAuthenticationError(c, err.Error())
	}

	// Generate tokens
	tokens, err := h.authService.GenerateTokens(user.ID, user.PhoneNumber)
	if err != nil {
		return response.SendInternalServerError(c, "Failed to generate authentication tokens")
	}

	// Return successful login response
	return response.SendLoginSuccess(
		c,
		tokens.AccessToken,
		tokens.RefreshToken,
		tokens.ExpiresIn,
		user.ID.String(),
		user.PhoneNumber,
	)
}

// Refresh handles POST /auth/refresh endpoint
// Validates refresh token and issues new access and refresh tokens
func (h *handler) Refresh(c *fiber.Ctx) error {
	var req RefreshRequest
	
	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return response.SendValidationError(c, "Invalid request body")
	}

	// Validate required fields
	if req.RefreshToken == "" {
		return response.SendValidationError(c, "Refresh token is required")
	}

	// Validate the refresh token
	claims, err := h.authService.ValidateToken(req.RefreshToken)
	if err != nil {
		return response.SendAuthenticationError(c, "Invalid or expired refresh token")
	}

	// Ensure this is actually a refresh token
	if claims.TokenType != "refresh" {
		return response.SendAuthenticationError(c, "Invalid token type")
	}

	// Blacklist the old refresh token
	if err := h.authService.BlacklistToken(req.RefreshToken); err != nil {
		return response.SendInternalServerError(c, "Failed to invalidate old refresh token")
	}

	// Generate new tokens
	tokens, err := h.authService.GenerateTokens(claims.UserID, claims.PhoneNumber)
	if err != nil {
		return response.SendInternalServerError(c, "Failed to generate new authentication tokens")
	}

	// Return new tokens
	return response.SendLoginSuccess(
		c,
		tokens.AccessToken,
		tokens.RefreshToken,
		tokens.ExpiresIn,
		claims.UserID.String(),
		claims.PhoneNumber,
	)
}

// Logout handles POST /auth/logout endpoint
// Invalidates both access and refresh tokens by adding them to blacklist
func (h *handler) Logout(c *fiber.Ctx) error {
	// Extract access token from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return response.SendValidationError(c, "Authorization header is required")
	}

	// Check if header starts with "Bearer "
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return response.SendValidationError(c, "Invalid authorization header format")
	}

	// Extract the token
	accessToken := strings.TrimPrefix(authHeader, "Bearer ")
	if accessToken == "" {
		return response.SendValidationError(c, "Access token is required")
	}

	// Validate the access token
	claims, err := h.authService.ValidateToken(accessToken)
	if err != nil {
		return response.SendAuthenticationError(c, "Invalid or expired access token")
	}

	// Ensure this is an access token
	if claims.TokenType != "access" {
		return response.SendAuthenticationError(c, "Invalid token type")
	}

	// Blacklist the access token
	if err := h.authService.BlacklistToken(accessToken); err != nil {
		return response.SendInternalServerError(c, "Failed to invalidate access token")
	}

	// Parse refresh token from request body (optional)
	var req RefreshRequest
	if err := c.BodyParser(&req); err == nil && req.RefreshToken != "" {
		// If refresh token is provided, blacklist it as well
		if err := h.authService.BlacklistToken(req.RefreshToken); err != nil {
			// Log error but don't fail the logout process
			// In a real application, you'd use a proper logger here
		}
	}

	// Return success response
	return response.SendSuccess(c, nil, "Logout successful")
}