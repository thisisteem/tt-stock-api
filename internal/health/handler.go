package health

import (
	"database/sql"
	"time"

	"github.com/gofiber/fiber/v2"
	"tt-stock-api/internal/config"
	"tt-stock-api/pkg/response"
)

// Handler handles health check requests
type Handler struct {
	db     *sql.DB
	config *config.Config
}

// NewHandler creates a new health check handler
func NewHandler(db *sql.DB, config *config.Config) *Handler {
	return &Handler{
		db:     db,
		config: config,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Uptime    string            `json:"uptime"`
	Database  DatabaseHealth    `json:"database"`
	System    SystemInfo        `json:"system"`
}

// DatabaseHealth represents database health information
type DatabaseHealth struct {
	Status      string `json:"status"`
	Connected   bool   `json:"connected"`
	ResponseTime string `json:"response_time"`
}

// SystemInfo represents system information
type SystemInfo struct {
	Environment string `json:"environment"`
	Port        string `json:"port"`
}

var startTime = time.Now()

// Health handles GET /health requests
func (h *Handler) Health(c *fiber.Ctx) error {
	// Check database connectivity
	dbHealth := h.checkDatabase()
	
	// Determine overall status
	status := "healthy"
	if !dbHealth.Connected {
		status = "unhealthy"
		// Return 503 Service Unavailable if database is not connected
		c.Status(fiber.StatusServiceUnavailable)
	}

	healthResponse := HealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0", // TODO: Get from build info
		Uptime:    time.Since(startTime).String(),
		Database:  dbHealth,
		System: SystemInfo{
			Environment: h.config.Env,
			Port:        h.config.Port,
		},
	}

	return response.SendSuccess(c, healthResponse, "Health check completed")
}

// Readiness handles GET /ready requests (for Kubernetes readiness probes)
func (h *Handler) Readiness(c *fiber.Ctx) error {
	// Check if the application is ready to serve requests
	dbHealth := h.checkDatabase()
	
	if !dbHealth.Connected {
		return response.SendError(c, fiber.StatusServiceUnavailable, "SERVICE_NOT_READY", "Database connection failed")
	}

	return response.SendSuccess(c, map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"database":  dbHealth,
	}, "Service is ready")
}

// Liveness handles GET /live requests (for Kubernetes liveness probes)
func (h *Handler) Liveness(c *fiber.Ctx) error {
	// Simple liveness check - if we can respond, we're alive
	return response.SendSuccess(c, map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    time.Since(startTime).String(),
	}, "Service is alive")
}

// checkDatabase checks database connectivity and measures response time
func (h *Handler) checkDatabase() DatabaseHealth {
	if h.db == nil {
		return DatabaseHealth{
			Status:       "error",
			Connected:    false,
			ResponseTime: "0ms",
		}
	}

	start := time.Now()
	
	// Simple ping to check database connectivity
	err := h.db.Ping()
	responseTime := time.Since(start)

	if err != nil {
		return DatabaseHealth{
			Status:       "error",
			Connected:    false,
			ResponseTime: responseTime.String(),
		}
	}

	// Additional check: try a simple query
	var result int
	err = h.db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		return DatabaseHealth{
			Status:       "error",
			Connected:    false,
			ResponseTime: responseTime.String(),
		}
	}

	return DatabaseHealth{
		Status:       "healthy",
		Connected:    true,
		ResponseTime: responseTime.String(),
	}
}