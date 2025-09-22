// Package middleware contains HTTP middleware implementations for the TT Stock Backend API.
// It provides middleware functions that handle cross-cutting concerns like authentication,
// logging, security, and request processing in the HTTP delivery layer.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityMiddleware handles CORS and security headers
type SecurityMiddleware struct {
	allowedOrigins []string
	allowedMethods []string
	allowedHeaders []string
}

// NewSecurityMiddleware creates a new SecurityMiddleware instance
func NewSecurityMiddleware(allowedOrigins []string) *SecurityMiddleware {
	return &SecurityMiddleware{
		allowedOrigins: allowedOrigins,
		allowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		allowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-API-Key",
			"X-Client-Version",
			"X-Device-ID",
		},
	}
}

// CORS handles Cross-Origin Resource Sharing
func (m *SecurityMiddleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		if m.isOriginAllowed(origin) {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			// For development, allow localhost origins
			if m.isDevelopmentOrigin(origin) {
				c.Header("Access-Control-Allow-Origin", origin)
			}
		}

		// Set CORS headers
		c.Header("Access-Control-Allow-Methods", strings.Join(m.allowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(m.allowedHeaders, ", "))
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SecurityHeaders adds security headers to responses
func (m *SecurityMiddleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Strict Transport Security (HTTPS only)
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' https:; connect-src 'self'")

		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Remove server information
		c.Header("Server", "")

		c.Next()
	}
}

// RateLimiting provides basic rate limiting (placeholder for future implementation)
func (m *SecurityMiddleware) RateLimiting() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement rate limiting logic
		// For now, just pass through
		c.Next()
	}
}

// RequestSizeLimit limits the size of incoming requests
func (m *SecurityMiddleware) RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check Content-Length header
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":   "Request Entity Too Large",
				"message": "Request size exceeds maximum allowed size",
			})
			c.Abort()
			return
		}

		// Limit request body size
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

		c.Next()
	}
}

// IPWhitelist restricts access to specific IP addresses
func (m *SecurityMiddleware) IPWhitelist(allowedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Check if IP is in whitelist
		if !m.isIPAllowed(clientIP, allowedIPs) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Access denied from this IP address",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// APIKeyAuth provides API key authentication
func (m *SecurityMiddleware) APIKeyAuth(validAPIKeys []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "API key is required",
			})
			c.Abort()
			return
		}

		// Check if API key is valid
		if !m.isValidAPIKey(apiKey, validAPIKeys) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid API key",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// HTTPSRedirect redirects HTTP requests to HTTPS
func (m *SecurityMiddleware) HTTPSRedirect() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request is HTTP
		if c.Request.TLS == nil && c.Request.Header.Get("X-Forwarded-Proto") != "https" {
			// Redirect to HTTPS
			httpsURL := "https://" + c.Request.Host + c.Request.RequestURI
			c.Redirect(http.StatusMovedPermanently, httpsURL)
			c.Abort()
			return
		}

		c.Next()
	}
}

// SecurityAudit logs security-related events
func (m *SecurityMiddleware) SecurityAudit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log suspicious activities
		if m.isSuspiciousRequest(c) {
			// TODO: Implement security audit logging
			// For now, just pass through
		}

		c.Next()
	}
}

// Helper methods

// isOriginAllowed checks if the origin is in the allowed list
func (m *SecurityMiddleware) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}

	for _, allowedOrigin := range m.allowedOrigins {
		if origin == allowedOrigin {
			return true
		}
	}

	return false
}

// isDevelopmentOrigin checks if the origin is a development origin
func (m *SecurityMiddleware) isDevelopmentOrigin(origin string) bool {
	developmentOrigins := []string{
		"http://localhost:3000",
		"http://localhost:3001",
		"http://127.0.0.1:3000",
		"http://127.0.0.1:3001",
		"http://localhost:8080",
		"http://127.0.0.1:8080",
	}

	for _, devOrigin := range developmentOrigins {
		if origin == devOrigin {
			return true
		}
	}

	return false
}

// isIPAllowed checks if the IP address is in the allowed list
func (m *SecurityMiddleware) isIPAllowed(clientIP string, allowedIPs []string) bool {
	for _, allowedIP := range allowedIPs {
		if clientIP == allowedIP {
			return true
		}
	}
	return false
}

// isValidAPIKey checks if the API key is valid
func (m *SecurityMiddleware) isValidAPIKey(apiKey string, validAPIKeys []string) bool {
	for _, validKey := range validAPIKeys {
		if apiKey == validKey {
			return true
		}
	}
	return false
}

// isSuspiciousRequest checks for suspicious request patterns
func (m *SecurityMiddleware) isSuspiciousRequest(c *gin.Context) bool {
	// Check for common attack patterns
	suspiciousPatterns := []string{
		"../",
		"..\\",
		"<script",
		"javascript:",
		"onload=",
		"onerror=",
		"union select",
		"drop table",
		"delete from",
		"insert into",
		"update set",
	}

	path := strings.ToLower(c.Request.URL.Path)
	query := strings.ToLower(c.Request.URL.RawQuery)

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(path, pattern) || strings.Contains(query, pattern) {
			return true
		}
	}

	// Check for excessive request size
	if c.Request.ContentLength > 10*1024*1024 { // 10MB
		return true
	}

	// Check for unusual user agents
	userAgent := strings.ToLower(c.Request.UserAgent())
	suspiciousUserAgents := []string{
		"sqlmap",
		"nikto",
		"nmap",
		"masscan",
		"zap",
		"burp",
	}

	for _, suspicious := range suspiciousUserAgents {
		if strings.Contains(userAgent, suspicious) {
			return true
		}
	}

	return false
}

// Security configuration constants
const (
	DefaultMaxRequestSize = 10 * 1024 * 1024 // 10MB
	DefaultRateLimit      = 1000             // requests per minute
)
