package user

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tt-stock-api/internal/db"
)

func TestRepository_FindByPhoneNumber(t *testing.T) {
	tests := []struct {
		name        string
		phoneNumber string
		setupMock   func(mock sqlmock.Sqlmock)
		expected    *User
		expectError bool
		errorMsg    string
	}{
		{
			name:        "successful user retrieval",
			phoneNumber: "0812345678",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "phone_number", "pin_hash", "created_at", "updated_at", "last_login_at"}).
					AddRow("123e4567-e89b-12d3-a456-426614174000", "0812345678", "$2a$12$hashedpin", 
						time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC))
				
				mock.ExpectQuery(`SELECT id, phone_number, pin_hash, created_at, updated_at, last_login_at FROM users WHERE phone_number = \$1`).
					WithArgs("0812345678").
					WillReturnRows(rows)
			},
			expected: &User{
				ID:          uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				PhoneNumber: "0812345678",
				PinHash:     "$2a$12$hashedpin",
				CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				LastLoginAt: func() *time.Time { t := time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC); return &t }(),
			},
			expectError: false,
		},
		{
			name:        "successful user retrieval with null last_login_at",
			phoneNumber: "0812345679",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "phone_number", "pin_hash", "created_at", "updated_at", "last_login_at"}).
					AddRow("123e4567-e89b-12d3-a456-426614174001", "0812345679", "$2a$12$hashedpin2", 
						time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						nil)
				
				mock.ExpectQuery(`SELECT id, phone_number, pin_hash, created_at, updated_at, last_login_at FROM users WHERE phone_number = \$1`).
					WithArgs("0812345679").
					WillReturnRows(rows)
			},
			expected: &User{
				ID:          uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
				PhoneNumber: "0812345679",
				PinHash:     "$2a$12$hashedpin2",
				CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				LastLoginAt: nil,
			},
			expectError: false,
		},
		{
			name:        "user not found",
			phoneNumber: "0899999999",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, phone_number, pin_hash, created_at, updated_at, last_login_at FROM users WHERE phone_number = \$1`).
					WithArgs("0899999999").
					WillReturnError(sql.ErrNoRows)
			},
			expected:    nil,
			expectError: true,
			errorMsg:    "user with phone number 0899999999 not found",
		},
		{
			name:        "empty phone number",
			phoneNumber: "",
			setupMock:   func(mock sqlmock.Sqlmock) {
				// No mock setup needed as validation happens before query
			},
			expected:    nil,
			expectError: true,
			errorMsg:    "phone number cannot be empty",
		},
		{
			name:        "database error",
			phoneNumber: "0812345678",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, phone_number, pin_hash, created_at, updated_at, last_login_at FROM users WHERE phone_number = \$1`).
					WithArgs("0812345678").
					WillReturnError(errors.New("database connection error"))
			},
			expected:    nil,
			expectError: true,
			errorMsg:    "failed to query user by phone number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock database
			mockDB, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer mockDB.Close()

			// Setup mock expectations
			tt.setupMock(mock)

			// Create repository with mock database
			dbWrapper := &db.DB{DB: mockDB}
			repo := NewRepository(dbWrapper)

			// Execute the method
			result, err := repo.FindByPhoneNumber(tt.phoneNumber)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepository_UpdateLastLogin(t *testing.T) {
	testUserID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	testUserID2 := uuid.MustParse("123e4567-e89b-12d3-a456-426614174999")
	
	tests := []struct {
		name        string
		userID      uuid.UUID
		setupMock   func(mock sqlmock.Sqlmock)
		expectError bool
		errorMsg    string
	}{
		{
			name:   "successful update",
			userID: testUserID,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE users SET last_login_at = \$1, updated_at = \$1 WHERE id = \$2`).
					WithArgs(sqlmock.AnyArg(), testUserID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
		{
			name:   "user not found",
			userID: testUserID2,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE users SET last_login_at = \$1, updated_at = \$1 WHERE id = \$2`).
					WithArgs(sqlmock.AnyArg(), testUserID2).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectError: true,
			errorMsg:    "user with ID 123e4567-e89b-12d3-a456-426614174999 not found",
		},
		{
			name:   "empty user ID",
			userID: uuid.Nil,
			setupMock: func(mock sqlmock.Sqlmock) {
				// No mock setup needed as validation happens before query
			},
			expectError: true,
			errorMsg:    "user ID cannot be empty",
		},
		{
			name:   "database error on exec",
			userID: testUserID,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE users SET last_login_at = \$1, updated_at = \$1 WHERE id = \$2`).
					WithArgs(sqlmock.AnyArg(), testUserID).
					WillReturnError(errors.New("database connection error"))
			},
			expectError: true,
			errorMsg:    "failed to update last login for user",
		},
		{
			name:   "database error on rows affected",
			userID: testUserID,
			setupMock: func(mock sqlmock.Sqlmock) {
				result := sqlmock.NewErrorResult(errors.New("rows affected error"))
				mock.ExpectExec(`UPDATE users SET last_login_at = \$1, updated_at = \$1 WHERE id = \$2`).
					WithArgs(sqlmock.AnyArg(), testUserID).
					WillReturnResult(result)
			},
			expectError: true,
			errorMsg:    "failed to get rows affected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock database
			mockDB, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer mockDB.Close()

			// Setup mock expectations
			tt.setupMock(mock)

			// Create repository with mock database
			dbWrapper := &db.DB{DB: mockDB}
			repo := NewRepository(dbWrapper)

			// Execute the method
			err = repo.UpdateLastLogin(tt.userID)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestRepository_Interface verifies that repository implements the Repository interface
func TestRepository_Interface(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	dbWrapper := &db.DB{DB: mockDB}
	repo := NewRepository(dbWrapper)

	// Verify that repo implements Repository interface
	var _ Repository = repo
}

// Helper function to create a mock result that returns an error for RowsAffected
type errorResult struct {
	err error
}

func (er errorResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (er errorResult) RowsAffected() (int64, error) {
	return 0, er.err
}

// Custom matcher for sqlmock to handle any time argument
type anyTime struct{}

func (a anyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}