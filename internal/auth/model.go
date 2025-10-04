package auth

import (
	"time"

	"github.com/google/uuid"
)

// TokenBlacklist represents a blacklisted token in the system
type TokenBlacklist struct {
	ID            uuid.UUID `json:"id" db:"id"`
	Token         string    `json:"token" db:"token"`
	UserID        uuid.UUID `json:"user_id" db:"user_id"`
	TokenType     string    `json:"token_type" db:"token_type"` // "access" or "refresh"
	ExpiresAt     time.Time `json:"expires_at" db:"expires_at"`
	BlacklistedAt time.Time `json:"blacklisted_at" db:"blacklisted_at"`
}