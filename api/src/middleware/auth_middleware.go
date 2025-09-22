// Package middleware contains HTTP middleware implementations for the TT Stock Backend API.
// It provides middleware functions that handle cross-cutting concerns like authentication,
// logging, security, and request processing in the HTTP delivery layer.
package middleware

import (
	"net/http"
	"time"

	"tt-stock-api/src/models"
	"tt-stock-api/src/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware handles JWT authentication for protected routes
type AuthMiddleware struct {
	authService services.AuthService
	jwtSecret   string
}

// NewAuthMiddleware creates a new AuthMiddleware instance
func NewAuthMiddleware(authService services.AuthService, jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		jwtSecret:   jwtSecret,
	}
}

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	UserID      uint      `json:"userId"`
	PhoneNumber string    `json:"phoneNumber"`
	Role        string    `json:"role"`
	ExpiresAt   time.Time `json:"expiresAt"`
	jwt.RegisteredClaims
}

// RequireAuth is the main authentication middleware
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		token := m.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Token is required",
			})
			c.Abort()
			return
		}

		// Validate and parse JWT token
		claims, err := m.validateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid token",
			})
			c.Abort()
			return
		}

		// Check if token is expired
		if time.Now().After(claims.ExpiresAt) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Token has expired",
			})
			c.Abort()
			return
		}

		// Get user from service to ensure user still exists and is active
		user, err := m.authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "User not found",
			})
			c.Abort()
			return
		}

		// Check if user is active
		if !user.IsActive {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Account is deactivated",
			})
			c.Abort()
			return
		}

		// Set user in context for use in handlers
		c.Set("user", user)
		c.Set("userID", user.ID)
		c.Set("userRole", user.Role)

		c.Next()
	}
}

// RequireRole creates middleware that requires specific user roles
func (m *AuthMiddleware) RequireRole(requiredRoles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if user is authenticated
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		userModel, ok := user.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Invalid user type in context",
			})
			c.Abort()
			return
		}

		// Check if user has required role
		hasRequiredRole := false
		for _, role := range requiredRoles {
			if userModel.Role == role {
				hasRequiredRole = true
				break
			}
		}

		if !hasRequiredRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin creates middleware that requires admin role
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRole(models.UserRoleAdmin)
}

// RequireOwnerOrAdmin creates middleware that requires owner or admin role
func (m *AuthMiddleware) RequireOwnerOrAdmin() gin.HandlerFunc {
	return m.RequireRole(models.UserRoleOwner, models.UserRoleAdmin)
}

// RequireStaffOrAbove creates middleware that requires staff, owner, or admin role
func (m *AuthMiddleware) RequireStaffOrAbove() gin.HandlerFunc {
	return m.RequireRole(models.UserRoleStaff, models.UserRoleOwner, models.UserRoleAdmin)
}

// OptionalAuth creates middleware that validates token if present but doesn't require it
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		// Validate token if provided
		claims, err := m.validateToken(token)
		if err != nil {
			// Invalid token, continue without authentication
			c.Next()
			return
		}

		// Check if token is expired
		if time.Now().After(claims.ExpiresAt) {
			// Expired token, continue without authentication
			c.Next()
			return
		}

		// Get user from service
		user, err := m.authService.ValidateToken(c.Request.Context(), token)
		if err != nil || !user.IsActive {
			// User not found or inactive, continue without authentication
			c.Next()
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Set("userID", user.ID)
		c.Set("userRole", user.Role)

		c.Next()
	}
}

// RefreshTokenMiddleware handles token refresh requests
func (m *AuthMiddleware) RefreshTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract refresh token from request
		var req struct {
			RefreshToken string `json:"refreshToken" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Bad Request",
				"message": "Refresh token is required",
			})
			c.Abort()
			return
		}

		// Validate refresh token
		user, err := m.authService.ValidateToken(c.Request.Context(), req.RefreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid refresh token",
			})
			c.Abort()
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Set("userID", user.ID)
		c.Set("userRole", user.Role)

		c.Next()
	}
}

// Helper methods

// extractToken extracts JWT token from Authorization header
func (m *AuthMiddleware) extractToken(c *gin.Context) string {
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

// validateToken validates and parses JWT token
func (m *AuthMiddleware) validateToken(tokenString string) (*JWTClaims, error) {
	// Check if token string is empty or too short to be valid
	if tokenString == "" || len(tokenString) < 10 {
		return nil, jwt.ErrTokenMalformed
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(m.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenMalformed
	}

	return claims, nil
}

// GenerateToken generates a new JWT token for a user
func (m *AuthMiddleware) GenerateToken(user *models.User) (string, error) {
	claims := &JWTClaims{
		UserID:      user.ID,
		PhoneNumber: user.PhoneNumber,
		Role:        string(user.Role),
		ExpiresAt:   time.Now().Add(24 * time.Hour), // 1 day expiration
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "tt-stock-api",
			Subject:   user.PhoneNumber,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.jwtSecret))
}

// GenerateRefreshToken generates a new refresh token for a user
func (m *AuthMiddleware) GenerateRefreshToken(user *models.User) (string, error) {
	claims := &JWTClaims{
		UserID:      user.ID,
		PhoneNumber: user.PhoneNumber,
		Role:        string(user.Role),
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour), // 7 days expiration
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "tt-stock-api",
			Subject:   user.PhoneNumber,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.jwtSecret))
}

// GetUserFromContext extracts user from Gin context
func GetUserFromContext(c *gin.Context) (*models.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	userModel, ok := user.(*models.User)
	if !ok {
		return nil, false
	}

	return userModel, true
}

// GetUserIDFromContext extracts user ID from Gin context
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, false
	}

	id, ok := userID.(uint)
	if !ok {
		return 0, false
	}

	return id, true
}

// GetUserRoleFromContext extracts user role from Gin context
func GetUserRoleFromContext(c *gin.Context) (models.UserRole, bool) {
	userRole, exists := c.Get("userRole")
	if !exists {
		return "", false
	}

	role, ok := userRole.(models.UserRole)
	if !ok {
		return "", false
	}

	return role, true
}
