package user

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"tt-stock-api/internal/db"
)

// Repository defines the interface for user data operations
type Repository interface {
	FindByPhoneNumber(phoneNumber string) (*User, error)
	UpdateLastLogin(userID uuid.UUID) error
}

// repository implements the Repository interface
type repository struct {
	db *db.DB
}

// NewRepository creates a new user repository instance
func NewRepository(database *db.DB) Repository {
	return &repository{
		db: database,
	}
}

// FindByPhoneNumber retrieves a user by their phone number
func (r *repository) FindByPhoneNumber(phoneNumber string) (*User, error) {
	if phoneNumber == "" {
		return nil, errors.New("phone number cannot be empty")
	}

	query := `
		SELECT id, phone_number, pin_hash, created_at, updated_at, last_login_at
		FROM users 
		WHERE phone_number = $1
	`

	var user User
	var lastLoginAt sql.NullTime

	err := r.db.QueryRow(query, phoneNumber).Scan(
		&user.ID,
		&user.PhoneNumber,
		&user.PinHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with phone number %s not found", phoneNumber)
		}
		return nil, fmt.Errorf("failed to query user by phone number: %w", err)
	}

	// Handle nullable last_login_at field
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

// UpdateLastLogin updates the last login timestamp for a user
func (r *repository) UpdateLastLogin(userID uuid.UUID) error {
	if userID == uuid.Nil {
		return errors.New("user ID cannot be empty")
	}

	query := `
		UPDATE users 
		SET last_login_at = $1, updated_at = $1 
		WHERE id = $2
	`

	now := time.Now()
	result, err := r.db.Exec(query, now, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login for user %s: %w", userID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with ID %s not found", userID)
	}

	return nil
}