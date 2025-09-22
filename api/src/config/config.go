// Package config provides configuration management for the TT Stock API.
// It handles loading configuration from environment variables and provides
// structured configuration objects for different parts of the application.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Default configuration constants
const (
	defaultServerReadTimeout  = 30 * time.Second
	defaultServerWriteTimeout = 30 * time.Second
	defaultServerIdleTimeout  = 60 * time.Second
	defaultDBMaxOpenConns     = 20
	defaultDBMaxIdleConns     = 5
	defaultDBConnMaxIdleTime  = 30 * time.Minute
	defaultJWTTokenLifetime   = 24 * time.Hour
)

// Config holds all configuration for the application
type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	JWT        JWTConfig
	App        AppConfig
	Logging    LoggingConfig
	Security   SecurityConfig
	Middleware MiddlewareConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey     string
	TokenLifetime time.Duration
}

// AppConfig holds application configuration
type AppConfig struct {
	Name        string
	Version     string
	Environment string
	Debug       bool
	LogLevel    string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string
	Format     string
	Output     string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	CORSOrigins    []string
	RateLimitRPS   int
	RateLimitBurst int
	TrustedProxies []string
	EnableHTTPS    bool
	CertFile       string
	KeyFile        string
}

// MiddlewareConfig holds middleware configuration
type MiddlewareConfig struct {
	EnableCORS      bool
	EnableRateLimit bool
	EnableLogging   bool
	EnableRecovery  bool
	EnableSecurity  bool
	RequestTimeout  time.Duration
	MaxRequestSize  int64
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", defaultServerReadTimeout),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", defaultServerWriteTimeout),
			IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", defaultServerIdleTimeout),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "password"),
			DBName:          getEnv("DB_NAME", "tt_stock"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", defaultDBMaxOpenConns),
			MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", defaultDBMaxIdleConns),
			ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 1*time.Hour),
			ConnMaxIdleTime: getDurationEnv("DB_CONN_MAX_IDLE_TIME", defaultDBConnMaxIdleTime),
		},
		JWT: JWTConfig{
			SecretKey:     getEnv("JWT_SECRET_KEY", "your-secret-key-change-in-production"),
			TokenLifetime: getDurationEnv("JWT_TOKEN_LIFETIME", defaultJWTTokenLifetime),
		},
		App: AppConfig{
			Name:        getEnv("APP_NAME", "TT Stock API"),
			Version:     getEnv("APP_VERSION", "1.0.0"),
			Environment: getEnv("APP_ENV", "development"),
			Debug:       getBoolEnv("APP_DEBUG", false),
			LogLevel:    getEnv("LOG_LEVEL", "info"),
		},
		Logging: LoggingConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			Output:     getEnv("LOG_OUTPUT", "stdout"),
			MaxSize:    getIntEnv("LOG_MAX_SIZE", 100),
			MaxBackups: getIntEnv("LOG_MAX_BACKUPS", 3),
			MaxAge:     getIntEnv("LOG_MAX_AGE", 28),
			Compress:   getBoolEnv("LOG_COMPRESS", true),
		},
		Security: SecurityConfig{
			CORSOrigins:    getStringSliceEnv("CORS_ORIGINS", []string{"*"}),
			RateLimitRPS:   getIntEnv("RATE_LIMIT_RPS", 100),
			RateLimitBurst: getIntEnv("RATE_LIMIT_BURST", 200),
			TrustedProxies: getStringSliceEnv("TRUSTED_PROXIES", []string{}),
			EnableHTTPS:    getBoolEnv("ENABLE_HTTPS", false),
			CertFile:       getEnv("TLS_CERT_FILE", ""),
			KeyFile:        getEnv("TLS_KEY_FILE", ""),
		},
		Middleware: MiddlewareConfig{
			EnableCORS:      getBoolEnv("ENABLE_CORS", true),
			EnableRateLimit: getBoolEnv("ENABLE_RATE_LIMIT", true),
			EnableLogging:   getBoolEnv("ENABLE_LOGGING", true),
			EnableRecovery:  getBoolEnv("ENABLE_RECOVERY", true),
			EnableSecurity:  getBoolEnv("ENABLE_SECURITY", true),
			RequestTimeout:  getDurationEnv("REQUEST_TIMEOUT", 30*time.Second),
			MaxRequestSize:  getInt64Env("MAX_REQUEST_SIZE", 10*1024*1024), // 10MB
		},
	}

	// Validate required configuration
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// validate validates the configuration
func (c *Config) validate() error {
	if c.JWT.SecretKey == "your-secret-key-change-in-production" && c.App.Environment == "production" {
		return fmt.Errorf("JWT secret key must be changed in production")
	}

	if c.Database.MaxOpenConns <= 0 {
		return fmt.Errorf("database max open connections must be greater than 0")
	}

	if c.Database.MaxIdleConns <= 0 {
		return fmt.Errorf("database max idle connections must be greater than 0")
	}

	return nil
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// GetServerAddress returns the server address
func (c *ServerConfig) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getBoolEnv gets a boolean environment variable with a default value
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getInt64Env gets an int64 environment variable with a default value
func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getStringSliceEnv gets a string slice environment variable with a default value
func getStringSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Split by comma and trim spaces
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return defaultValue
}
