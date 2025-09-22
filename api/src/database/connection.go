// Package database contains database connection and configuration for the TT Stock Backend API.
// It provides database connection management, migration handling, and connection pooling
// following Clean Architecture principles.
package database

import (
	"context"
	"fmt"

	"tt-stock-api/src/config"

	"gorm.io/gorm"
)

// ConnectionManager manages database connections and provides repository access
type ConnectionManager struct {
	db     *gorm.DB
	config *config.DatabaseConfig
}

// NewConnectionManager creates a new database connection manager
func NewConnectionManager() (*ConnectionManager, error) {
	// Load configuration from environment
	appConfig, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create database connection using existing config package
	configDB, err := config.NewDatabase(&appConfig.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &ConnectionManager{
		db:     configDB.DB,
		config: &appConfig.Database,
	}, nil
}

// GetDB returns the GORM database instance
func (cm *ConnectionManager) GetDB() *gorm.DB {
	return cm.db
}

// GetConfig returns the database configuration
func (cm *ConnectionManager) GetConfig() *config.DatabaseConfig {
	return cm.config
}

// Close closes the database connection
func (cm *ConnectionManager) Close() error {
	sqlDB, err := cm.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Close()
}

// Ping tests the database connection
func (cm *ConnectionManager) Ping(ctx context.Context) error {
	sqlDB, err := cm.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.PingContext(ctx)
}

// HealthCheck performs a database health check
func (cm *ConnectionManager) HealthCheck(ctx context.Context) error {
	// Test connection
	if err := cm.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Test a simple query
	var count int64
	if err := cm.db.Raw("SELECT 1").Count(&count).Error; err != nil {
		return fmt.Errorf("database query test failed: %w", err)
	}

	return nil
}

// Transaction executes a function within a database transaction
func (cm *ConnectionManager) Transaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return cm.db.WithContext(ctx).Transaction(fn)
}
