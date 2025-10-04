package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"tt-stock-api/pkg/response"
)

// JWTProtected creates a middleware function that validates JWT tokens for protected routes
func JWTProtected(authService Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.SendAuthenticationError(c, "Authorization header is required")
		}

		// Check if header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return response.SendAuthenticationError(c, "Invalid authorization header format")
		}

		// Extract the token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			return response.SendAuthenticationError(c, "Access token is required")
		}

		// Validate the token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			if strings.Contains(err.Error(), "expired") {
				return response.SendTokenExpiredError(c, "Access token has expired")
			}
			if strings.Contains(err.Error(), "invalidated") {
				return response.SendAuthenticationError(c, "Token has been invalidated")
			}
			return response.SendAuthenticationError(c, "Invalid access token")
		}

		// Ensure this is an access token (not a refresh token)
		if claims.TokenType != "access" {
			return response.SendAuthenticationError(c, "Invalid token type: access token required")
		}

		// Add user information to context for use in handlers
		c.Locals("user_id", claims.UserID.String())
		c.Locals("phone_number", claims.PhoneNumber)
		c.Locals("token_claims", claims)

		// Continue to the next handler
		return c.Next()
	}
}

// ExtractUserFromContext extracts user information from the Fiber context
// This is a helper function for handlers to get user info from protected routes
func ExtractUserFromContext(c *fiber.Ctx) (userID string, phoneNumber string, ok bool) {
	userIDInterface := c.Locals("user_id")
	phoneNumberInterface := c.Locals("phone_number")

	if userIDInterface == nil || phoneNumberInterface == nil {
		return "", "", false
	}

	userID, userIDOk := userIDInterface.(string)
	phoneNumber, phoneNumberOk := phoneNumberInterface.(string)

	if !userIDOk || !phoneNumberOk {
		return "", "", false
	}

	return userID, phoneNumber, true
}

// ExtractClaimsFromContext extracts the full JWT claims from the Fiber context
// This is a helper function for handlers that need access to all token claims
func ExtractClaimsFromContext(c *fiber.Ctx) (*Claims, bool) {
	claimsInterface := c.Locals("token_claims")
	if claimsInterface == nil {
		return nil, false
	}

	claims, ok := claimsInterface.(*Claims)
	return claims, ok
}