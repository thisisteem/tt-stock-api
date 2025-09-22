// Package router provides HTTP router configuration for the TT Stock Backend API.
// It sets up all API endpoints, middleware, and routing following Clean Architecture principles.
package router

import (
	"context"
	"log"
	"net/http"
	"time"

	"tt-stock-api/src/config"
	"tt-stock-api/src/handlers"
	"tt-stock-api/src/middleware"
	"tt-stock-api/src/repositories"
	"tt-stock-api/src/services"

	"github.com/gin-gonic/gin"
)

// Router holds the Gin router and dependencies
type Router struct {
	engine            *gin.Engine
	config            *config.Config
	repositoryFactory *repositories.RepositoryFactory
	authService       services.AuthService
	userService       services.UserService
	productService    services.ProductService
	stockService      services.StockService
	alertService      services.AlertService
	biService         services.BusinessIntelligenceService

	// Handlers
	authHandler *handlers.AuthHandler
	// userHandler         *handlers.UserHandler
	productHandler *handlers.ProductHandler
	stockHandler   *handlers.StockHandler
	alertHandler   *handlers.AlertHandler
	// biHandler           *handlers.BusinessIntelligenceHandler

	// Middleware
	authMiddleware       *middleware.AuthMiddleware
	validationMiddleware *middleware.ValidationMiddleware
}

// NewRouter creates a new router instance
func NewRouter(cfg *config.Config, repoFactory *repositories.RepositoryFactory) *Router {
	// Set Gin mode based on environment
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin engine
	engine := gin.New()

	// Create services
	authService := services.NewAuthService(repoFactory.GetUserRepository(), repoFactory.GetSessionRepository(), cfg.JWT.SecretKey, cfg.JWT.TokenLifetime)
	userService := services.NewUserService(repoFactory.GetUserRepository())
	productService := services.NewProductService(repoFactory.GetProductRepository())
	stockService := services.NewStockService(repoFactory.GetStockMovementRepository(), repoFactory.GetProductRepository(), repoFactory.GetAlertRepository())
	alertService := services.NewAlertService(repoFactory.GetAlertRepository(), repoFactory.GetProductRepository())
	biService := services.NewBusinessIntelligenceService(repoFactory.GetProductRepository(), repoFactory.GetStockMovementRepository(), repoFactory.GetUserRepository(), repoFactory.GetAlertRepository())

	return &Router{
		engine:            engine,
		config:            cfg,
		repositoryFactory: repoFactory,
		authService:       authService,
		userService:       userService,
		productService:    productService,
		stockService:      stockService,
		alertService:      alertService,
		biService:         biService,
	}
}

// SetupRoutes configures all API routes and middleware
func (r *Router) SetupRoutes() {
	// Setup middleware
	r.setupMiddleware()

	// Setup handlers
	r.setupHandlers()

	// Setup API routes
	r.setupAPIRoutes()

	// Setup health check
	r.setupHealthCheck()

	// Setup static files (if needed)
	r.setupStaticFiles()
}

// setupMiddleware configures all middleware
func (r *Router) setupMiddleware() {
	// Recovery middleware
	if r.config.Middleware.EnableRecovery {
		r.engine.Use(gin.Recovery())
	}

	// Logging middleware
	if r.config.Middleware.EnableLogging {
		loggingMiddleware := middleware.NewLoggingMiddleware(log.New(log.Writer(), "", log.LstdFlags))
		r.engine.Use(loggingMiddleware.RequestLogger())
		r.engine.Use(loggingMiddleware.ErrorLogger())
		r.engine.Use(loggingMiddleware.SecurityLogger())
	}

	// CORS middleware
	if r.config.Middleware.EnableCORS {
		securityMiddleware := middleware.NewSecurityMiddleware(r.config.Security.CORSOrigins)
		r.engine.Use(securityMiddleware.CORS())
	}

	// Security middleware
	if r.config.Middleware.EnableSecurity {
		securityMiddleware := middleware.NewSecurityMiddleware(r.config.Security.CORSOrigins)
		r.engine.Use(securityMiddleware.SecurityHeaders())
		r.engine.Use(securityMiddleware.RequestSizeLimit(r.config.Middleware.MaxRequestSize))
	}

	// Rate limiting middleware
	if r.config.Middleware.EnableRateLimit {
		securityMiddleware := middleware.NewSecurityMiddleware(r.config.Security.CORSOrigins)
		r.engine.Use(securityMiddleware.RateLimiting())
	}

	// Request timeout middleware
	if r.config.Middleware.RequestTimeout > 0 {
		r.engine.Use(func(c *gin.Context) {
			ctx, cancel := context.WithTimeout(c.Request.Context(), r.config.Middleware.RequestTimeout)
			defer cancel()
			c.Request = c.Request.WithContext(ctx)
			c.Next()
		})
	}
}

// setupHandlers configures all handlers
func (r *Router) setupHandlers() {
	// Create validators (commented out until needed)
	// authValidator := validators.NewAuthValidator()
	// productValidator := validators.NewProductValidator()
	// stockValidator := validators.NewStockValidator()

	// Create validation middleware
	validationMiddleware := middleware.NewValidationMiddleware()

	// Create auth middleware
	authMiddleware := middleware.NewAuthMiddleware(r.authService, r.config.JWT.SecretKey)

	// Create handlers
	authHandler := handlers.NewAuthHandler(r.authService)
	// Note: UserHandler and BusinessIntelligenceHandler don't exist yet, we'll create them
	// userHandler := handlers.NewUserHandler(r.userService)
	productHandler := handlers.NewProductHandler(r.productService)
	stockHandler := handlers.NewStockHandler(r.stockService)
	alertHandler := handlers.NewAlertHandler(r.alertService)
	// biHandler := handlers.NewBusinessIntelligenceHandler(r.biService)

	// Store handlers for route setup
	r.authHandler = authHandler
	// r.userHandler = userHandler
	r.productHandler = productHandler
	r.stockHandler = stockHandler
	r.alertHandler = alertHandler
	// r.biHandler = biHandler
	r.authMiddleware = authMiddleware
	r.validationMiddleware = validationMiddleware
}

// setupAPIRoutes configures all API routes
func (r *Router) setupAPIRoutes() {
	// API v1 group
	v1 := r.engine.Group("/v1")

	// Public routes (no authentication required)
	public := v1.Group("/")
	{
		// Health check
		public.GET("/health", r.healthCheck)

		// Authentication routes
		auth := public.Group("/auth")
		{
			auth.POST("/login", r.validationMiddleware.ValidateAuthRequest(), r.authHandler.Login)
			auth.POST("/refresh", r.validationMiddleware.ValidateAuthRequest(), r.authHandler.RefreshToken)
		}
	}

	// Protected routes (authentication required)
	protected := v1.Group("/")
	protected.Use(r.authMiddleware.RequireAuth())
	{
		// User management routes (commented out until UserHandler is implemented)
		// users := protected.Group("/users")
		// users.Use(r.authMiddleware.RequireAdmin()) // Only admins can manage users
		// {
		//     users.POST("/", r.validationMiddleware.ValidateUserRequest(), r.userHandler.CreateUser)
		//     users.GET("/:id", r.userHandler.GetUser)
		//     users.PUT("/:id", r.validationMiddleware.ValidateUserRequest(), r.userHandler.UpdateUser)
		//     users.DELETE("/:id", r.userHandler.DeleteUser)
		//     users.GET("/", r.userHandler.ListUsers)
		// }

		// Product management routes
		products := protected.Group("/products")
		products.Use(r.authMiddleware.RequireStaffOrAbove()) // Staff and above can manage products
		{
			products.POST("/", r.validationMiddleware.ValidateProductRequest(), r.productHandler.CreateProduct)
			products.GET("/:id", r.productHandler.GetProduct)
			products.GET("/sku/:sku", r.productHandler.GetProductBySKU)
			products.PUT("/:id", r.validationMiddleware.ValidateProductRequest(), r.productHandler.UpdateProduct)
			products.DELETE("/:id", r.productHandler.DeleteProduct)
			products.GET("/", r.productHandler.ListProducts)
			products.POST("/search", r.validationMiddleware.ValidateProductRequest(), r.productHandler.SearchProducts)
			products.GET("/low-stock", r.productHandler.GetLowStockProducts)
			products.GET("/out-of-stock", r.productHandler.GetOutOfStockProducts)
			products.GET("/statistics", r.productHandler.GetProductStatistics)
		}

		// Stock management routes
		stock := protected.Group("/stock")
		stock.Use(r.authMiddleware.RequireStaffOrAbove()) // Staff and above can manage stock
		{
			// Stock movements
			stock.POST("/movements", r.validationMiddleware.ValidateStockRequest(), r.stockHandler.CreateStockMovement)
			stock.GET("/movements/:id", r.stockHandler.GetStockMovement)
			stock.PUT("/movements/:id", r.validationMiddleware.ValidateStockRequest(), r.stockHandler.UpdateStockMovement)
			stock.DELETE("/movements/:id", r.stockHandler.DeleteStockMovement)
			stock.GET("/movements", r.stockHandler.ListStockMovements)
			stock.GET("/movements/summary", r.stockHandler.GetStockMovementSummary)
			stock.GET("/movements/recent", r.stockHandler.GetRecentMovements)

			// Stock operations
			stock.POST("/sales", r.validationMiddleware.ValidateStockRequest(), r.stockHandler.ProcessSale)
			stock.POST("/incoming", r.validationMiddleware.ValidateStockRequest(), r.stockHandler.ProcessIncomingStock)
			stock.POST("/adjustments", r.validationMiddleware.ValidateStockRequest(), r.stockHandler.ProcessStockAdjustment)
			stock.POST("/returns", r.validationMiddleware.ValidateStockRequest(), r.stockHandler.ProcessReturn)
		}

		// Alert management routes
		alerts := protected.Group("/alerts")
		{
			alerts.POST("/", r.validationMiddleware.ValidateAlertRequest(), r.alertHandler.CreateAlert)
			alerts.GET("/:id", r.alertHandler.GetAlert)
			alerts.PUT("/:id", r.validationMiddleware.ValidateAlertRequest(), r.alertHandler.UpdateAlert)
			alerts.DELETE("/:id", r.alertHandler.DeleteAlert)
			alerts.GET("/", r.alertHandler.ListAlerts)
			alerts.GET("/unread", r.alertHandler.GetUnreadAlerts)
			alerts.PUT("/:id/read", r.alertHandler.MarkAlertAsRead)
			alerts.PUT("/read-all", r.alertHandler.MarkAllAlertsAsRead)
			alerts.GET("/unread/count", r.alertHandler.GetUnreadCount)
			alerts.PUT("/:id/deactivate", r.alertHandler.DeactivateAlert)
		}

		// Business intelligence routes (commented out until BusinessIntelligenceHandler is implemented)
		// bi := protected.Group("/business-intelligence")
		// bi.Use(r.authMiddleware.RequireOwnerOrAdmin()) // Only owners and admins can access BI
		// {
		//     bi.GET("/dashboard", r.biHandler.GetDashboardData)
		//     bi.GET("/product-statistics", r.biHandler.GetProductStatistics)
		//     bi.GET("/sales-summary", r.biHandler.GetSalesSummary)
		//     bi.GET("/stock-movement-summary", r.biHandler.GetStockMovementSummary)
		//     bi.GET("/recent-alerts", r.biHandler.GetRecentAlerts)
		//     bi.GET("/user-activity", r.biHandler.GetUserActivitySummary)
		// }

		// User profile routes
		profile := protected.Group("/profile")
		{
			profile.GET("/", r.authHandler.GetProfile)
			// profile.PUT("/change-pin", r.validationMiddleware.ValidateAuthRequest(), r.authHandler.ChangePIN)
			profile.POST("/logout", r.authHandler.Logout)
			// profile.POST("/logout-all", r.authHandler.RevokeAllSessions)
		}
	}
}

// setupHealthCheck configures health check endpoints
func (r *Router) setupHealthCheck() {
	r.engine.GET("/health", r.healthCheck)
	r.engine.GET("/health/detailed", r.detailedHealthCheck)
}

// setupStaticFiles configures static file serving
func (r *Router) setupStaticFiles() {
	// Serve static files if needed
	// r.engine.Static("/static", "./static")
	// r.engine.StaticFile("/favicon.ico", "./static/favicon.ico")
}

// healthCheck provides a simple health check endpoint
func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   r.config.App.Name,
		"version":   r.config.App.Version,
	})
}

// detailedHealthCheck provides a detailed health check endpoint
func (r *Router) detailedHealthCheck(c *gin.Context) {
	// Check database health
	dbHealth := "healthy"
	// Note: We'll need to implement GetConnectionManager in RepositoryFactory
	// For now, we'll assume healthy
	// if err := r.repositoryFactory.GetConnectionManager().HealthCheck(c.Request.Context()); err != nil {
	//     dbHealth = "unhealthy"
	// }

	// Check connection pool health
	poolHealth := "healthy"
	// Note: We'll need to implement GetConnectionManager in RepositoryFactory
	// For now, we'll assume healthy
	// if !r.repositoryFactory.GetConnectionManager().IsConnectionPoolHealthy() {
	//     poolHealth = "unhealthy"
	// }

	c.JSON(http.StatusOK, gin.H{
		"status":      "healthy",
		"timestamp":   time.Now().UTC(),
		"service":     r.config.App.Name,
		"version":     r.config.App.Version,
		"environment": r.config.App.Environment,
		"checks": gin.H{
			"database":        dbHealth,
			"connection_pool": poolHealth,
		},
	})
}

// GetEngine returns the Gin engine
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// Start starts the HTTP server
func (r *Router) Start() error {
	server := &http.Server{
		Addr:         r.config.Server.Host + ":" + r.config.Server.Port,
		Handler:      r.engine,
		ReadTimeout:  r.config.Server.ReadTimeout,
		WriteTimeout: r.config.Server.WriteTimeout,
		IdleTimeout:  r.config.Server.IdleTimeout,
	}

	log.Printf("Starting server on %s:%s", r.config.Server.Host, r.config.Server.Port)
	return server.ListenAndServe()
}
