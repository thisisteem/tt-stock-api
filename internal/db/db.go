package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// DB holds the database connection
type DB struct {
	*sql.DB
}

// Connect establishes a connection to PostgreSQL database
func Connect(databaseURL string) (*DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database")
	return &DB{db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// CreateTables creates the necessary tables for the application
func (db *DB) CreateTables() error {
	// Create users table
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		phone_number VARCHAR(10) UNIQUE NOT NULL,
		pin_hash VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		last_login_at TIMESTAMP WITH TIME ZONE
	);`

	if _, err := db.Exec(usersTable); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create token_blacklist table
	tokenBlacklistTable := `
	CREATE TABLE IF NOT EXISTS token_blacklist (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		token TEXT NOT NULL,
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token_type VARCHAR(10) NOT NULL CHECK (token_type IN ('access', 'refresh')),
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		blacklisted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);`

	if _, err := db.Exec(tokenBlacklistTable); err != nil {
		return fmt.Errorf("failed to create token_blacklist table: %w", err)
	}

	// Create index on phone_number for faster lookups
	phoneIndex := `CREATE INDEX IF NOT EXISTS idx_users_phone_number ON users(phone_number);`
	if _, err := db.Exec(phoneIndex); err != nil {
		return fmt.Errorf("failed to create phone number index: %w", err)
	}

	// Create index on token for faster blacklist lookups
	tokenIndex := `CREATE INDEX IF NOT EXISTS idx_token_blacklist_token ON token_blacklist(token);`
	if _, err := db.Exec(tokenIndex); err != nil {
		return fmt.Errorf("failed to create token index: %w", err)
	}

	// Create index on user_id for faster user token lookups
	userTokenIndex := `CREATE INDEX IF NOT EXISTS idx_token_blacklist_user_id ON token_blacklist(user_id);`
	if _, err := db.Exec(userTokenIndex); err != nil {
		return fmt.Errorf("failed to create user token index: %w", err)
	}

	log.Println("Database tables created successfully")
	return nil
}