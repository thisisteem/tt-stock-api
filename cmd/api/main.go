package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tt-stock-api/internal/app"
	"tt-stock-api/internal/app/routes"
	"tt-stock-api/internal/config"
	"tt-stock-api/internal/db"
)

func main() {
	// Validate environment variables before any other initialization
	if err := config.ValidateEnvironment(); err != nil {
		config.PrintValidationError(err)
		os.Exit(1)
	}

	// Load configuration from environment variables
	cfg := config.Load()
	log.Printf("Starting TT Stock API with configuration: Port=%s, Env=%s", cfg.Port, cfg.Env)

	// Initialize database connection
	database, err := db.Connect(cfg.DBUrl)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// Create database tables if they don't exist
	if err := database.CreateTables(); err != nil {
		log.Fatalf("Failed to create database tables: %v", err)
	}

	// Create Fiber server with configuration
	server := app.NewServer(cfg)

	// Set up dependency injection for all layers
	deps := &routes.Dependencies{
		DB:     database,
		Config: cfg,
	}

	// Register all routes with dependency injection
	routes.RegisterRoutes(server.GetApp(), deps)

	// Channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Printf("Server startup error: %v", err)
		}
	}()

	log.Printf("Server started successfully on port %s", cfg.Port)
	log.Println("Press Ctrl+C to gracefully shutdown the server...")

	// Wait for interrupt signal
	<-quit
	log.Println("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Graceful shutdown with timeout
	shutdownComplete := make(chan error, 1)
	go func() {
		shutdownComplete <- server.Shutdown()
	}()

	select {
	case <-ctx.Done():
		log.Println("Shutdown timeout exceeded, forcing exit...")
	case err := <-shutdownComplete:
		if err != nil {
			log.Printf("Server shutdown error: %v", err)
		} else {
			log.Println("Server shutdown completed successfully")
		}
	}

	log.Println("TT Stock API stopped")
}