package auth

import (
	"errors"
	"fmt"
	"time"

	"tt-stock-api/internal/db"
)

// BlacklistRepository defines the interface for token blacklist operations
type BlacklistRepository interface {
	BlacklistToken(token, userID, tokenType string, expiresAt time.Time) error
	IsTokenBlacklisted(token string) (bool, error)
}

// blacklistRepository implements the BlacklistRepository interface
type blacklistRepository struct {
	db *db.DB
}

// NewBlacklistRepository creates a new blacklist repository instance
func NewBlacklistRepository(database *db.DB) BlacklistRepository {
	return &blacklistRepository{
		db: database,
	}
}

// BlacklistToken adds a token to the blacklist
func (r *blacklistRepository) BlacklistToken(token, userID, tokenType string, expiresAt time.Time) error {
	if token == "" {
		return errors.New("token cannot be empty")
	}
	if userID == "" {
		return errors.New("user ID cannot be empty")
	}
	if tokenType == "" {
		return errors.New("token type cannot be empty")
	}

	query := `
		INSERT INTO token_blacklist (token, user_id, token_type, expires_at, blacklisted_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	now := time.Now()
	_, err := r.db.Exec(query, token, userID, tokenType, expiresAt, now)
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

// IsTokenBlacklisted checks if a token is in the blacklist
func (r *blacklistRepository) IsTokenBlacklisted(token string) (bool, error) {
	if token == "" {
		return false, errors.New("token cannot be empty")
	}

	query := `
		SELECT EXISTS(
			SELECT 1 FROM token_blacklist 
			WHERE token = $1 AND expires_at > NOW()
		)
	`

	var exists bool
	err := r.db.QueryRow(query, token).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist status: %w", err)
	}

	return exists, nil
}