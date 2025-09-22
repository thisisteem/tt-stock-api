// Package config provides database connection management for the TT Stock API.
// It handles PostgreSQL connections with connection pooling and health checks.
package config

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database holds the database connection and configuration
type Database struct {
	DB     *gorm.DB
	Config *DatabaseConfig
}

// NewDatabase creates a new database connection
func NewDatabase(config *DatabaseConfig) (*Database, error) {
	// Create GORM logger based on environment
	var gormLogger logger.Interface
	if config.Host == "localhost" {
		// Development: verbose logging
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		// Production: only errors
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(config.GetDSN()), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool with optimal settings
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Set connection pool monitoring
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	log.Printf("Database connection pool configured: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v, MaxIdleTime=%v",
		config.MaxOpenConns, config.MaxIdleConns, config.ConnMaxLifetime, config.ConnMaxIdleTime)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Database connected successfully to %s:%s/%s",
		config.Host, config.Port, config.DBName)

	return &Database{
		DB:     db,
		Config: config,
	}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Close()
}

// Health checks the database connection
func (d *Database) Health() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Ping()
}

// GetConnectionStats returns database connection statistics
func (d *Database) GetConnectionStats() (map[string]interface{}, error) {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}, nil
}

// MonitorConnectionPool logs connection pool statistics periodically
func (d *Database) MonitorConnectionPool(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		stats, err := d.GetConnectionStats()
		if err != nil {
			log.Printf("Failed to get connection stats: %v", err)
			continue
		}

		log.Printf("Database Connection Pool Stats: Open=%v, InUse=%v, Idle=%v, WaitCount=%v",
			stats["open_connections"], stats["in_use"], stats["idle"], stats["wait_count"])
	}
}

// IsConnectionPoolHealthy checks if the connection pool is healthy
func (d *Database) IsConnectionPoolHealthy() bool {
	stats, err := d.GetConnectionStats()
	if err != nil {
		return false
	}

	openConnections := stats["open_connections"].(int)
	maxOpenConnections := stats["max_open_connections"].(int)
	waitCount := stats["wait_count"].(int64)

	// Consider unhealthy if:
	// 1. All connections are in use
	// 2. There are many wait requests
	// 3. Connection pool is at capacity
	if openConnections >= maxOpenConnections || waitCount > 10 {
		return false
	}

	return true
}
