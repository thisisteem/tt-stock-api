// Package middleware contains HTTP middleware implementations for the TT Stock Backend API.
// It provides middleware functions that handle cross-cutting concerns like authentication,
// logging, security, and request processing in the HTTP delivery layer.
package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware handles request/response logging
type LoggingMiddleware struct {
	logger *log.Logger
}

// NewLoggingMiddleware creates a new LoggingMiddleware instance
func NewLoggingMiddleware(logger *log.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// RequestLogger logs HTTP requests and responses
func (m *LoggingMiddleware) RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Create log message
		logMessage := fmt.Sprintf("[%s] %s %s %d %v %s %s",
			param.TimeStamp.Format(time.RFC3339),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Request.UserAgent(),
		)

		// Add user information if available
		if userID, exists := param.Keys["userID"]; exists {
			logMessage += fmt.Sprintf(" user_id=%v", userID)
		}
		if userRole, exists := param.Keys["userRole"]; exists {
			logMessage += fmt.Sprintf(" user_role=%v", userRole)
		}

		// Add error message if present
		if param.ErrorMessage != "" {
			logMessage += fmt.Sprintf(" error=%s", param.ErrorMessage)
		}

		// Log based on status code
		switch {
		case param.StatusCode >= 500:
			m.logger.Printf("ERROR: %s", logMessage)
		case param.StatusCode >= 400:
			m.logger.Printf("WARN: %s", logMessage)
		default:
			m.logger.Printf("INFO: %s", logMessage)
		}

		return "" // Return empty string since we're using structured logging
	})
}

// DetailedLogger logs detailed request/response information
func (m *LoggingMiddleware) DetailedLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Log request details
		m.logRequest(c)

		// Capture response
		responseWriter := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseWriter

		// Process request
		c.Next()

		// Log response details
		m.logResponse(c, responseWriter, start)
	}
}

// ErrorLogger logs errors with stack traces
func (m *LoggingMiddleware) ErrorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check for errors
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				// Get stack trace
				stack := make([]byte, 4096)
				length := runtime.Stack(stack, false)
				stackTrace := string(stack[:length])

				// Log error with context
				m.logger.Printf("ERROR: %s | Type: %v | Stack: %s | Method: %s | Path: %s | ClientIP: %s | UserAgent: %s",
					err.Error(),
					err.Type,
					stackTrace,
					c.Request.Method,
					c.Request.URL.Path,
					c.ClientIP(),
					c.Request.UserAgent(),
				)
			}
		}
	}
}

// PerformanceLogger logs slow requests
func (m *LoggingMiddleware) PerformanceLogger(threshold time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		if latency > threshold {
			m.logger.Printf("WARN: Slow request detected | Method: %s | Path: %s | Status: %d | Latency: %v | Threshold: %v | ClientIP: %s | UserAgent: %s",
				c.Request.Method,
				c.Request.URL.Path,
				c.Writer.Status(),
				latency,
				threshold,
				c.ClientIP(),
				c.Request.UserAgent(),
			)
		}
	}
}

// SecurityLogger logs security-related events
func (m *LoggingMiddleware) SecurityLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log authentication attempts
		if c.Request.URL.Path == "/v1/auth/login" {
			m.logger.Printf("INFO: Login attempt | Method: %s | ClientIP: %s | UserAgent: %s",
				c.Request.Method,
				c.ClientIP(),
				c.Request.UserAgent(),
			)
		}

		// Log authorization failures
		if c.Writer.Status() == http.StatusForbidden {
			m.logger.Printf("WARN: Authorization failure | Method: %s | Path: %s | ClientIP: %s | UserAgent: %s",
				c.Request.Method,
				c.Request.URL.Path,
				c.ClientIP(),
				c.Request.UserAgent(),
			)
		}

		// Log authentication failures
		if c.Writer.Status() == http.StatusUnauthorized {
			m.logger.Printf("WARN: Authentication failure | Method: %s | Path: %s | ClientIP: %s | UserAgent: %s",
				c.Request.Method,
				c.Request.URL.Path,
				c.ClientIP(),
				c.Request.UserAgent(),
			)
		}

		c.Next()
	}
}

// AuditLogger logs audit events for compliance
func (m *LoggingMiddleware) AuditLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only log for sensitive operations
		if m.isSensitiveOperation(c) {
			start := time.Now()

			c.Next()

			// Log audit event
			m.logger.Printf("AUDIT: %s | Method: %s | Path: %s | Status: %d | Latency: %v | ClientIP: %s | UserAgent: %s | Timestamp: %s",
				"audit_event",
				c.Request.Method,
				c.Request.URL.Path,
				c.Writer.Status(),
				time.Since(start),
				c.ClientIP(),
				c.Request.UserAgent(),
				time.Now().UTC().Format(time.RFC3339),
			)
		} else {
			c.Next()
		}
	}
}

// Helper methods

// logRequest logs request details
func (m *LoggingMiddleware) logRequest(c *gin.Context) {
	// Read request body
	var requestBody string
	if c.Request.Body != nil {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err == nil {
			requestBody = string(bodyBytes)
			// Restore request body for further processing
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}

	// Log request
	m.logger.Printf("REQUEST: %s | Method: %s | Path: %s | Query: %s | Headers: %s | Body: %s | ClientIP: %s | UserAgent: %s | ContentType: %s | Timestamp: %s",
		"incoming_request",
		c.Request.Method,
		c.Request.URL.Path,
		c.Request.URL.RawQuery,
		m.sanitizeHeaders(c.Request.Header),
		m.sanitizeBody(requestBody),
		c.ClientIP(),
		c.Request.UserAgent(),
		c.Request.Header.Get("Content-Type"),
		time.Now().UTC().Format(time.RFC3339),
	)
}

// logResponse logs response details
func (m *LoggingMiddleware) logResponse(c *gin.Context, responseWriter *responseBodyWriter, start time.Time) {
	// Log response
	m.logger.Printf("RESPONSE: %s | Method: %s | Path: %s | Status: %d | Latency: %v | BodySize: %d | Body: %s | ClientIP: %s | Timestamp: %s",
		"outgoing_response",
		c.Request.Method,
		c.Request.URL.Path,
		c.Writer.Status(),
		time.Since(start),
		responseWriter.body.Len(),
		m.sanitizeBody(responseWriter.body.String()),
		c.ClientIP(),
		time.Now().UTC().Format(time.RFC3339),
	)
}

// sanitizeHeaders removes sensitive headers from logging
func (m *LoggingMiddleware) sanitizeHeaders(headers http.Header) string {
	sensitiveHeaders := []string{"authorization", "cookie", "x-api-key"}
	sanitized := make(map[string]string)

	for key, values := range headers {
		keyLower := strings.ToLower(key)
		isSensitive := false

		for _, sensitive := range sensitiveHeaders {
			if keyLower == sensitive {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			sanitized[key] = "[REDACTED]"
		} else {
			sanitized[key] = strings.Join(values, ", ")
		}
	}

	// Convert to JSON string
	jsonBytes, _ := json.Marshal(sanitized)
	return string(jsonBytes)
}

// sanitizeBody removes sensitive data from request/response body
func (m *LoggingMiddleware) sanitizeBody(body string) string {
	if body == "" {
		return body
	}

	// Try to parse as JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(body), &jsonData); err != nil {
		// Not JSON, return as is
		return body
	}

	// Sanitize sensitive fields
	sensitiveFields := []string{"password", "pin", "token", "secret", "key"}
	m.sanitizeJSON(jsonData, sensitiveFields)

	// Convert back to JSON
	sanitizedBytes, err := json.Marshal(jsonData)
	if err != nil {
		return "[ERROR_SERIALIZING]"
	}

	return string(sanitizedBytes)
}

// sanitizeJSON recursively sanitizes sensitive fields in JSON
func (m *LoggingMiddleware) sanitizeJSON(data map[string]interface{}, sensitiveFields []string) {
	for key, value := range data {
		keyLower := strings.ToLower(key)
		for _, sensitive := range sensitiveFields {
			if strings.Contains(keyLower, sensitive) {
				data[key] = "[REDACTED]"
				break
			}
		}

		// Recursively sanitize nested objects
		if nestedMap, ok := value.(map[string]interface{}); ok {
			m.sanitizeJSON(nestedMap, sensitiveFields)
		}
	}
}

// isSensitiveOperation checks if the operation is sensitive and should be audited
func (m *LoggingMiddleware) isSensitiveOperation(c *gin.Context) bool {
	sensitivePaths := []string{
		"/v1/auth/login",
		"/v1/auth/logout",
		"/v1/users",
		"/v1/products",
		"/v1/stock/movements",
		"/v1/alerts",
	}

	path := c.Request.URL.Path
	for _, sensitivePath := range sensitivePaths {
		if strings.HasPrefix(path, sensitivePath) {
			return true
		}
	}

	return false
}

// responseBodyWriter captures response body for logging
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Logging configuration constants
const (
	DefaultPerformanceThreshold = 1 * time.Second
	MaxBodySizeForLogging       = 1024 * 1024 // 1MB
)
