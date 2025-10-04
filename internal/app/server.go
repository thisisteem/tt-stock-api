package app

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"tt-stock-api/internal/config"
)

// Server represents the Fiber application server
type Server struct {
	app    *fiber.App
	config *config.Config
}

// NewServer creates a new Fiber server instance with proper middleware configuration
func NewServer(cfg *config.Config) *Server {
	// Create Fiber app with custom configuration
	app := fiber.New(fiber.Config{
		
		// Error handling
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Default error code
			code := fiber.StatusInternalServerError
			
			// Check if it's a Fiber error
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			
			// Return error response
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "SERVER_ERROR",
					"message": err.Error(),
				},
			})
		},
		
		// Server settings
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
		
		// Disable startup message in production
		DisableStartupMessage: cfg.Env == "production",
	})

	// Add recovery middleware (should be first)
	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.Env == "development",
	}))

	// Add logger middleware
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} - ${latency}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
	}))

	// Add CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*", // In production, specify exact origins
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}))

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Server is healthy",
			"data": fiber.Map{
				"status":    "ok",
				"timestamp": time.Now().UTC(),
				"version":   "1.0.0",
			},
		})
	})

	return &Server{
		app:    app,
		config: cfg,
	}
}

// GetApp returns the Fiber app instance for route registration
func (s *Server) GetApp() *fiber.App {
	return s.app
}

// Start starts the Fiber server on the configured port
func (s *Server) Start() error {
	log.Printf("Starting server on port %s", s.config.Port)
	return s.app.Listen(":" + s.config.Port)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	log.Println("Shutting down server...")
	return s.app.Shutdown()
}