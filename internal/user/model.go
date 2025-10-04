package user

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	PhoneNumber string     `json:"phone_number" db:"phone_number"`
	PinHash     string     `json:"-" db:"pin_hash"` // Hidden from JSON responses
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
}