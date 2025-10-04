package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	JWTSecret string
	DBUrl     string
	Port      string
	Env       string
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		JWTSecret: getEnv("JWT_SECRET", "default-jwt-secret"),
		DBUrl:     getEnv("DB_URL", "postgres://localhost:5432/tt_stock_db?sslmode=disable"),
		Port:      getEnv("PORT", "8080"),
		Env:       getEnv("ENV", "development"),
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsInt gets an environment variable as integer with a fallback value
func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

// getEnvAsBool gets an environment variable as boolean with a fallback value
func getEnvAsBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return fallback
}