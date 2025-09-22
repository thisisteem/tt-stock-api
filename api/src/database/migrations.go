// Package database contains database migration and indexing functionality for the TT Stock Backend API.
// It provides database schema management, migration handling, and performance optimization
// following Clean Architecture principles.
package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"tt-stock-api/src/config"
	"tt-stock-api/src/models"

	"gorm.io/gorm"
)

// MigrationManager represents the database migration and indexing manager
type MigrationManager struct {
	DB     *gorm.DB
	config *config.DatabaseConfig
}

// NewMigrationManager creates a new migration manager using the existing config package
func NewMigrationManager(dbConfig *config.DatabaseConfig) (*MigrationManager, error) {
	// Use the existing config package to create database connection
	configDB, err := config.NewDatabase(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	return &MigrationManager{
		DB:     configDB.DB,
		config: dbConfig,
	}, nil
}

// Close closes the database connection
func (m *MigrationManager) Close() error {
	sqlDB, err := m.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Close()
}

// Ping tests the database connection
func (m *MigrationManager) Ping(ctx context.Context) error {
	sqlDB, err := m.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.PingContext(ctx)
}

// Migrate runs database migrations for all models
func (m *MigrationManager) Migrate() error {
	log.Println("Running database migrations...")

	// Auto-migrate all models
	err := m.DB.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.ProductSpecification{},
		&models.StockMovement{},
		&models.Session{},
		&models.Alert{},
	)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// CreateIndexes creates database indexes for performance optimization
func (m *MigrationManager) CreateIndexes() error {
	log.Println("Creating database indexes...")

	// User indexes
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_users_phone_number ON users(phone_number)").Error; err != nil {
		return fmt.Errorf("failed to create users phone_number index: %w", err)
	}
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_users_role ON users(role)").Error; err != nil {
		return fmt.Errorf("failed to create users role index: %w", err)
	}
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active)").Error; err != nil {
		return fmt.Errorf("failed to create users is_active index: %w", err)
	}

	// Product indexes
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku)").Error; err != nil {
		return fmt.Errorf("failed to create products sku index: %w", err)
	}
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_products_type ON products(type)").Error; err != nil {
		return fmt.Errorf("failed to create products type index: %w", err)
	}
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_products_brand ON products(brand)").Error; err != nil {
		return fmt.Errorf("failed to create products brand index: %w", err)
	}
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_products_quantity_on_hand ON products(quantity_on_hand)").Error; err != nil {
		return fmt.Errorf("failed to create products quantity_on_hand index: %w", err)
	}
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_products_is_active ON products(is_active)").Error; err != nil {
		return fmt.Errorf("failed to create products is_active index: %w", err)
	}

	// Product specification indexes
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_product_specifications_product_id ON product_specifications(product_id)").Error; err != nil {
		return fmt.Errorf("failed to create product_specifications product_id index: %w", err)
	}
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_product_specifications_spec_type ON product_specifications(spec_type)").Error; err != nil {
		return fmt.Errorf("failed to create product_specifications spec_type index: %w", err)
	}

	// Stock movement indexes
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_stock_movements_product_id ON stock_movements(product_id)").Error; err != nil {
		return fmt.Errorf("failed to create stock_movements product_id index: %w", err)
	}
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_stock_movements_user_id ON stock_movements(user_id)").Error; err != nil {
		return fmt.Errorf("failed to create stock_movements user_id index: %w", err)
	}
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_stock_movements_movement_type ON stock_movements(movement_type)").Error; err != nil {
		return fmt.Errorf("failed to create stock_movements movement_type index: %w", err)
	}
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_stock_movements_movement_date ON stock_movements(movement_date)").Error; err != nil {
		return fmt.Errorf("failed to create stock_movements movement_date index: %w", err)
	}

	// Session indexes
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)").Error; err != nil {
		return fmt.Errorf("failed to create sessions user_id index: %w", err)
	}
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token)").Error; err != nil {
		return fmt.Errorf("failed to create sessions token index: %w", err)
	}
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at)").Error; err != nil {
		return fmt.Errorf("failed to create sessions expires_at index: %w", err)
	}

	// Alert indexes
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_alerts_product_id ON alerts(product_id)").Error; err != nil {
		return fmt.Errorf("failed to create alerts product_id index: %w", err)
	}
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_alerts_user_id ON alerts(user_id)").Error; err != nil {
		return fmt.Errorf("failed to create alerts user_id index: %w", err)
	}
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_alerts_alert_type ON alerts(alert_type)").Error; err != nil {
		return fmt.Errorf("failed to create alerts alert_type index: %w", err)
	}
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_alerts_priority ON alerts(priority)").Error; err != nil {
		return fmt.Errorf("failed to create alerts priority index: %w", err)
	}
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_alerts_is_read ON alerts(is_read)").Error; err != nil {
		return fmt.Errorf("failed to create alerts is_read index: %w", err)
	}
	if err := m.DB.Exec("CREATE INDEX IF NOT EXISTS idx_alerts_is_active ON alerts(is_active)").Error; err != nil {
		return fmt.Errorf("failed to create alerts is_active index: %w", err)
	}

	log.Println("Database indexes created successfully")
	return nil
}

// GetDB returns the GORM database instance
func (m *MigrationManager) GetDB() *gorm.DB {
	return m.DB
}

// GetConfig returns the database configuration
func (m *MigrationManager) GetConfig() *config.DatabaseConfig {
	return m.config
}

// HealthCheck performs a database health check
func (m *MigrationManager) HealthCheck(ctx context.Context) error {
	// Test connection
	if err := m.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Test a simple query
	var count int64
	if err := m.DB.Model(&models.User{}).Count(&count).Error; err != nil {
		return fmt.Errorf("database query test failed: %w", err)
	}

	return nil
}

// Transaction executes a function within a database transaction
func (m *MigrationManager) Transaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return m.DB.WithContext(ctx).Transaction(fn)
}

// Database connection constants
const (
	DefaultMaxOpenConns    = 25
	DefaultMaxIdleConns    = 5
	DefaultConnMaxLifetime = 5 * time.Minute
	DefaultConnMaxIdleTime = 1 * time.Minute
	DefaultPingTimeout     = 5 * time.Second
)
