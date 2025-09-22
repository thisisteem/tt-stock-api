// Package main provides the main entry point for the TT Stock Backend API.
// It handles application startup, configuration loading, and graceful shutdown.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tt-stock-api/src/config"
	"tt-stock-api/src/database"
	"tt-stock-api/src/repositories"
	"tt-stock-api/src/router"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set log level based on configuration
	setLogLevel(cfg.App.LogLevel)

	log.Printf("Starting %s v%s in %s mode", cfg.App.Name, cfg.App.Version, cfg.App.Environment)

	// Initialize database
	log.Println("Initializing database connection...")
	connectionManager, err := database.InitializeDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := connectionManager.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// Create repository factory
	repoFactory := repositories.NewRepositoryFactory(connectionManager)

	// Create router
	appRouter := router.NewRouter(cfg, repoFactory)
	appRouter.SetupRoutes()

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:      appRouter.GetEngine(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// setLogLevel sets the log level based on configuration
func setLogLevel(level string) {
	switch level {
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
}

// setupGracefulShutdown sets up graceful shutdown handling
func setupGracefulShutdown(server *http.Server) {
	// Create a channel to receive OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// validateConfiguration validates the loaded configuration
func validateConfiguration(cfg *config.Config) error {
	// Validate required fields
	if cfg.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if cfg.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	if cfg.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}

	if cfg.JWT.SecretKey == "" {
		return fmt.Errorf("JWT secret key is required")
	}

	// Validate JWT secret key strength in production
	if cfg.App.Environment == "production" && len(cfg.JWT.SecretKey) < 32 {
		return fmt.Errorf("JWT secret key must be at least 32 characters in production")
	}

	return nil
}

// setupHealthChecks sets up health check endpoints
func setupHealthChecks(router *gin.Engine, connectionManager *database.ConnectionManager) {
	// Basic health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
		})
	})

	// Detailed health check
	router.GET("/health/detailed", func(c *gin.Context) {
		// Check database health
		dbHealth := "healthy"
		if err := connectionManager.HealthCheck(c.Request.Context()); err != nil {
			dbHealth = "unhealthy"
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"checks": gin.H{
				"database": dbHealth,
			},
		})
	})
}

// setupErrorHandling sets up global error handling
func setupErrorHandling(router *gin.Engine) {
	// Global error handler
	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.String(http.StatusInternalServerError, fmt.Sprintf("error: %s", err))
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	}))
}

// setupCORS sets up CORS middleware
func setupCORS(router *gin.Engine, allowedOrigins []string) {
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}

// setupLogging sets up logging middleware
func setupLogging(router *gin.Engine) {
	// Custom logging middleware
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
}

// setupSecurity sets up security middleware
func setupSecurity(router *gin.Engine) {
	// Security headers
	router.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	})
}

// setupRateLimiting sets up rate limiting middleware
func setupRateLimiting(router *gin.Engine) {
	// Simple rate limiting (in production, use a proper rate limiter)
	router.Use(func(c *gin.Context) {
		// Add rate limiting logic here
		c.Next()
	})
}

// setupRequestSizeLimit sets up request size limiting
func setupRequestSizeLimit(router *gin.Engine, maxSize int64) {
	router.Use(func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "Request entity too large",
			})
			c.Abort()
			return
		}
		c.Next()
	})
}

// setupTimeout sets up request timeout middleware
func setupTimeout(router *gin.Engine, timeout time.Duration) {
	router.Use(func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
}

// setupStaticFiles sets up static file serving
func setupStaticFiles(router *gin.Engine) {
	// Serve static files if needed
	// router.Static("/static", "./static")
	// router.StaticFile("/favicon.ico", "./static/favicon.ico")
}

// setupAPIDocumentation sets up API documentation endpoints
func setupAPIDocumentation(router *gin.Engine) {
	// API documentation endpoints
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":        "TT Stock Backend API",
			"version":     "1.0.0",
			"description": "Backend API for TT Stock mobile application",
			"endpoints": gin.H{
				"health":   "/health",
				"api":      "/v1",
				"auth":     "/v1/auth",
				"products": "/v1/products",
				"stock":    "/v1/stock",
				"alerts":   "/v1/alerts",
				"profile":  "/v1/profile",
			},
		})
	})
}
