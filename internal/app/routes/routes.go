package routes

import (
	"github.com/gofiber/fiber/v2"
	"tt-stock-api/internal/auth"
	"tt-stock-api/internal/config"
	"tt-stock-api/internal/db"
	"tt-stock-api/internal/health"
	"tt-stock-api/internal/user"
)

// Dependencies holds all the dependencies needed for route handlers
type Dependencies struct {
	DB     *db.DB
	Config *config.Config
}

// RegisterRoutes sets up all application routes with dependency injection
func RegisterRoutes(app *fiber.App, deps *Dependencies) {
	// Initialize repositories
	userRepo := user.NewRepository(deps.DB)
	blacklistRepo := auth.NewBlacklistRepository(deps.DB)

	// Initialize services
	authService := auth.NewService(userRepo, blacklistRepo, deps.Config)

	// Initialize handlers
	authHandler := auth.NewHandler(authService)
	healthHandler := health.NewHandler(deps.DB.DB, deps.Config)

	// Health check routes (no authentication required)
	app.Get("/health", healthHandler.Health)
	app.Get("/ready", healthHandler.Readiness)
	app.Get("/live", healthHandler.Liveness)

	// Create API v1 group
	api := app.Group("/api/v1")

	// Authentication routes
	authGroup := api.Group("/auth")
	{
		// POST /api/v1/auth/login - User login
		authGroup.Post("/login", authHandler.Login)

		// POST /api/v1/auth/refresh - Refresh access token
		authGroup.Post("/refresh", authHandler.Refresh)

		// POST /api/v1/auth/logout - User logout (requires authentication)
		authGroup.Post("/logout", auth.JWTProtected(authService), authHandler.Logout)
	}

	// Protected routes group (for future endpoints)
	protected := api.Group("/protected", auth.JWTProtected(authService))
	{
		// Example protected endpoint for testing
		protected.Get("/profile", func(c *fiber.Ctx) error {
			userID, phoneNumber, ok := auth.ExtractUserFromContext(c)
			if !ok {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"success": false,
					"error": fiber.Map{
						"code":    "UNAUTHORIZED",
						"message": "Failed to extract user information",
					},
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"message": "Profile retrieved successfully",
				"data": fiber.Map{
					"user_id":      userID,
					"phone_number": phoneNumber,
				},
			})
		})
	}

	// API documentation endpoint
	api.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "TT Stock API v1",
			"data": fiber.Map{
				"version": "1.0.0",
				"endpoints": fiber.Map{
					"auth": fiber.Map{
						"login":   "POST /api/v1/auth/login",
						"refresh": "POST /api/v1/auth/refresh",
						"logout":  "POST /api/v1/auth/logout",
					},
					"protected": fiber.Map{
						"profile": "GET /api/v1/protected/profile",
					},
				},
			},
		})
	})
}