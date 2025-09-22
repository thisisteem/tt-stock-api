// Package database contains database connection and configuration for the TT Stock Backend API.
// It provides database connection management, migration handling, and connection pooling
// following Clean Architecture principles.
package database

import (
	"context"
	"fmt"
	"log"

	"tt-stock-api/migrations"
	"tt-stock-api/src/models"
)

// InitializeDatabase initializes the database connection and runs migrations
func InitializeDatabase() (*ConnectionManager, error) {
	log.Println("Initializing database connection...")

	// Create connection manager
	connectionManager, err := NewConnectionManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create connection manager: %w", err)
	}

	// Run SQL migrations
	if err := runSQLMigrations(connectionManager); err != nil {
		connectionManager.Close()
		return nil, fmt.Errorf("failed to run SQL migrations: %w", err)
	}

	log.Println("Database initialization completed successfully")
	return connectionManager, nil
}

// runSQLMigrations runs the SQL migrations using the migration runner
func runSQLMigrations(connectionManager *ConnectionManager) error {
	// Get the underlying sql.DB
	sqlDB, err := connectionManager.GetDB().DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Create migration runner
	migrationRunner := migrations.NewMigrationRunner(sqlDB)

	// Run migrations
	if err := migrationRunner.RunMigrations("migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// HealthCheck performs a comprehensive database health check
func HealthCheck(connectionManager *ConnectionManager) error {
	ctx := context.Background()
	return connectionManager.HealthCheck(ctx)
}

// CreateTestData creates initial test data for development
func CreateTestData(connectionManager *ConnectionManager) error {
	log.Println("Creating test data...")

	db := connectionManager.GetDB()

	// Check if test data already exists
	var userCount int64
	if err := db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return fmt.Errorf("failed to check existing users: %w", err)
	}

	if userCount > 0 {
		log.Println("Test data already exists, skipping creation")
		return nil
	}

	// Create test users
	testUsers := []models.User{
		{
			PhoneNumber: "+1234567890",
			PIN:         "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password"
			Role:        models.UserRoleAdmin,
			Name:        "Admin User",
			IsActive:    true,
		},
		{
			PhoneNumber: "+1234567891",
			PIN:         "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password"
			Role:        models.UserRoleOwner,
			Name:        "Owner User",
			IsActive:    true,
		},
		{
			PhoneNumber: "+1234567892",
			PIN:         "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password"
			Role:        models.UserRoleStaff,
			Name:        "Staff User",
			IsActive:    true,
		},
	}

	for _, user := range testUsers {
		if err := db.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create test user: %w", err)
		}
	}

	// Create test products
	testProducts := []models.Product{
		{
			Type:              models.ProductTypeTire,
			Brand:             "Michelin",
			Model:             "Pilot Sport 4",
			SKU:               "MIC-PS4-225-45-17",
			Description:       stringPtr("High-performance summer tire"),
			CostPrice:         150.00,
			SellingPrice:      200.00,
			QuantityOnHand:    10,
			LowStockThreshold: 5,
			IsActive:          true,
		},
		{
			Type:              models.ProductTypeTire,
			Brand:             "Bridgestone",
			Model:             "Potenza RE-71R",
			SKU:               "BRI-RE71R-245-40-18",
			Description:       stringPtr("Ultra-high performance tire"),
			CostPrice:         180.00,
			SellingPrice:      250.00,
			QuantityOnHand:    8,
			LowStockThreshold: 3,
			IsActive:          true,
		},
		{
			Type:              models.ProductTypeWheel,
			Brand:             "Enkei",
			Model:             "Racing RPF1",
			SKU:               "ENK-RPF1-17X8-5X114",
			Description:       stringPtr("Lightweight racing wheel"),
			CostPrice:         300.00,
			SellingPrice:      400.00,
			QuantityOnHand:    5,
			LowStockThreshold: 2,
			IsActive:          true,
		},
	}

	for _, product := range testProducts {
		if err := db.Create(&product).Error; err != nil {
			return fmt.Errorf("failed to create test product: %w", err)
		}
	}

	log.Println("Test data created successfully")
	return nil
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}
