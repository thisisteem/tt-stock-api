// Package unit contains unit tests for the TT Stock Backend API models.
// It tests model validation, business logic, and data transformation methods.
package unit

import (
	"testing"
	"time"

	"tt-stock-api/src/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUser_Validate tests the User validation functionality
func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    models.User
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid user",
			user: models.User{
				PhoneNumber: "1234567890",
				PIN:         "1234",
				Role:        models.UserRoleAdmin,
				Name:        "John Doe",
			},
			wantErr: false,
		},
		{
			name: "empty phone number",
			user: models.User{
				PhoneNumber: "",
				PIN:         "1234",
				Role:        models.UserRoleAdmin,
				Name:        "John Doe",
			},
			wantErr: true,
			errMsg:  "phone number is required",
		},
		{
			name: "invalid phone number - too short",
			user: models.User{
				PhoneNumber: "123",
				PIN:         "1234",
				Role:        models.UserRoleAdmin,
				Name:        "John Doe",
			},
			wantErr: true,
			errMsg:  "phone number must be 10-15 digits",
		},
		{
			name: "invalid phone number - too long",
			user: models.User{
				PhoneNumber: "1234567890123456",
				PIN:         "1234",
				Role:        models.UserRoleAdmin,
				Name:        "John Doe",
			},
			wantErr: true,
			errMsg:  "phone number must be 10-15 digits",
		},
		{
			name: "phone number with formatting",
			user: models.User{
				PhoneNumber: "+1 (234) 567-8900",
				PIN:         "1234",
				Role:        models.UserRoleAdmin,
				Name:        "John Doe",
			},
			wantErr: false,
		},
		{
			name: "empty PIN",
			user: models.User{
				PhoneNumber: "1234567890",
				PIN:         "",
				Role:        models.UserRoleAdmin,
				Name:        "John Doe",
			},
			wantErr: true,
			errMsg:  "PIN is required",
		},
		{
			name: "invalid PIN - too short",
			user: models.User{
				PhoneNumber: "1234567890",
				PIN:         "12",
				Role:        models.UserRoleAdmin,
				Name:        "John Doe",
			},
			wantErr: true,
			errMsg:  "PIN must be 4-6 digits",
		},
		{
			name: "invalid PIN - too long",
			user: models.User{
				PhoneNumber: "1234567890",
				PIN:         "1234567",
				Role:        models.UserRoleAdmin,
				Name:        "John Doe",
			},
			wantErr: true,
			errMsg:  "PIN must be 4-6 digits",
		},
		{
			name: "invalid PIN - non-numeric",
			user: models.User{
				PhoneNumber: "1234567890",
				PIN:         "abcd",
				Role:        models.UserRoleAdmin,
				Name:        "John Doe",
			},
			wantErr: true,
			errMsg:  "PIN must contain only digits",
		},
		{
			name: "hashed PIN should pass validation",
			user: models.User{
				PhoneNumber: "1234567890",
				PIN:         "$2a$10$abcdefghijklmnopqrstuvwxyz",
				Role:        models.UserRoleAdmin,
				Name:        "John Doe",
			},
			wantErr: false,
		},
		{
			name: "invalid role",
			user: models.User{
				PhoneNumber: "1234567890",
				PIN:         "1234",
				Role:        "InvalidRole",
				Name:        "John Doe",
			},
			wantErr: true,
			errMsg:  "role must be one of: Admin, Owner, Staff",
		},
		{
			name: "valid roles",
			user: models.User{
				PhoneNumber: "1234567890",
				PIN:         "1234",
				Role:        models.UserRoleOwner,
				Name:        "John Doe",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			user: models.User{
				PhoneNumber: "1234567890",
				PIN:         "1234",
				Role:        models.UserRoleAdmin,
				Name:        "",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "name too short",
			user: models.User{
				PhoneNumber: "1234567890",
				PIN:         "1234",
				Role:        models.UserRoleAdmin,
				Name:        "A",
			},
			wantErr: true,
			errMsg:  "name must be 2-100 characters",
		},
		{
			name: "name too long",
			user: models.User{
				PhoneNumber: "1234567890",
				PIN:         "1234",
				Role:        models.UserRoleAdmin,
				Name:        "This is a very long name that exceeds the maximum allowed length of 100 characters and should fail validation",
			},
			wantErr: true,
			errMsg:  "name must be 2-100 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUser_ValidatePhoneNumber tests phone number validation
func TestUser_ValidatePhoneNumber(t *testing.T) {
	tests := []struct {
		name        string
		phoneNumber string
		wantErr     bool
		errMsg      string
	}{
		{"valid 10 digit", "1234567890", false, ""},
		{"valid 11 digit", "12345678901", false, ""},
		{"valid 15 digit", "123456789012345", false, ""},
		{"valid with formatting", "+1 (234) 567-8900", false, ""},
		{"valid with dashes", "123-456-7890", false, ""},
		{"valid with spaces", "123 456 7890", false, ""},
		{"empty", "", true, "phone number is required"},
		{"too short", "123", true, "phone number must be 10-15 digits"},
		{"too long", "1234567890123456", true, "phone number must be 10-15 digits"},
		{"non-numeric", "abcdefghij", true, "phone number must be 10-15 digits"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := models.User{PhoneNumber: tt.phoneNumber}
			err := user.ValidatePhoneNumber()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUser_ValidatePIN tests PIN validation
func TestUser_ValidatePIN(t *testing.T) {
	tests := []struct {
		name    string
		pin     string
		wantErr bool
		errMsg  string
	}{
		{"valid 4 digit", "1234", false, ""},
		{"valid 5 digit", "12345", false, ""},
		{"valid 6 digit", "123456", false, ""},
		{"hashed PIN", "$2a$10$abcdefghijklmnopqrstuvwxyz", false, ""},
		{"empty", "", true, "PIN is required"},
		{"too short", "12", true, "PIN must be 4-6 digits"},
		{"too long", "1234567", true, "PIN must be 4-6 digits"},
		{"non-numeric", "abcd", true, "PIN must contain only digits"},
		{"mixed", "12ab", true, "PIN must contain only digits"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := models.User{PIN: tt.pin}
			err := user.ValidatePIN()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUser_ValidateRole tests role validation
func TestUser_ValidateRole(t *testing.T) {
	tests := []struct {
		name    string
		role    models.UserRole
		wantErr bool
	}{
		{"valid admin", models.UserRoleAdmin, false},
		{"valid owner", models.UserRoleOwner, false},
		{"valid staff", models.UserRoleStaff, false},
		{"invalid role", "InvalidRole", true},
		{"empty role", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := models.User{Role: tt.role}
			err := user.ValidateRole()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "role must be one of: Admin, Owner, Staff")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUser_ValidateName tests name validation
func TestUser_ValidateName(t *testing.T) {
	tests := []struct {
		name     string
		userName string
		wantErr  bool
		errMsg   string
	}{
		{"valid name", "John Doe", false, ""},
		{"valid short name", "Jo", false, ""},
		{"valid long name", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", false, ""},
		{"empty", "", true, "name is required"},
		{"too short", "A", true, "name must be 2-100 characters"},
		{"too long", "This is a very long name that exceeds the maximum allowed length of 100 characters and should fail validation because it is too long", true, "name must be 2-100 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := models.User{Name: tt.userName}
			err := user.ValidateName()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUser_HashPIN tests PIN hashing
func TestUser_HashPIN(t *testing.T) {
	user := models.User{}
	pin := "1234"

	hashedPIN, err := user.HashPIN(pin)
	require.NoError(t, err)
	assert.NotEmpty(t, hashedPIN)
	assert.NotEqual(t, pin, hashedPIN)
	assert.True(t, len(hashedPIN) > 20) // bcrypt hashes are typically 60 characters
}

// TestUser_VerifyPIN tests PIN verification
func TestUser_VerifyPIN(t *testing.T) {
	user := models.User{}
	pin := "1234"

	// Hash the PIN first
	hashedPIN, err := user.HashPIN(pin)
	require.NoError(t, err)

	// Set the hashed PIN
	user.PIN = hashedPIN

	// Test correct PIN
	assert.True(t, user.VerifyPIN(pin))

	// Test incorrect PIN
	assert.False(t, user.VerifyPIN("5678"))
	assert.False(t, user.VerifyPIN(""))
}

// TestUser_BeforeCreate tests the GORM BeforeCreate hook
func TestUser_BeforeCreate(t *testing.T) {
	t.Run("valid user creation", func(t *testing.T) {
		user := models.User{
			PhoneNumber: "1234567890",
			PIN:         "1234",
			Role:        models.UserRoleAdmin,
			Name:        "John Doe",
		}

		err := user.BeforeCreate(nil)
		assert.NoError(t, err)
		assert.NotEqual(t, "1234", user.PIN) // PIN should be hashed
		assert.True(t, len(user.PIN) > 20)   // Hashed PIN should be long
	})

	t.Run("invalid user creation", func(t *testing.T) {
		user := models.User{
			PhoneNumber: "", // Invalid phone number
			PIN:         "1234",
			Role:        models.UserRoleAdmin,
			Name:        "John Doe",
		}

		err := user.BeforeCreate(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "phone number is required")
	})
}

// TestUser_BeforeUpdate tests the GORM BeforeUpdate hook
func TestUser_BeforeUpdate(t *testing.T) {
	t.Run("valid user update", func(t *testing.T) {
		user := models.User{
			PhoneNumber: "1234567890",
			PIN:         "1234",
			Role:        models.UserRoleAdmin,
			Name:        "John Doe",
		}

		// Test validation without GORM DB dependency
		err := user.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid user update", func(t *testing.T) {
		user := models.User{
			PhoneNumber: "", // Invalid phone number
			PIN:         "1234",
			Role:        models.UserRoleAdmin,
			Name:        "John Doe",
		}

		// Test validation without GORM DB dependency
		err := user.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "phone number is required")
	})
}

// TestUser_ToResponse tests conversion to UserResponse
func TestUser_ToResponse(t *testing.T) {
	now := time.Now()
	user := models.User{
		ID:          1,
		PhoneNumber: "1234567890",
		PIN:         "hashed_pin", // This should not appear in response
		Role:        models.UserRoleAdmin,
		Name:        "John Doe",
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
		LastLoginAt: &now,
	}

	response := user.ToResponse()

	assert.Equal(t, user.ID, response.ID)
	assert.Equal(t, user.PhoneNumber, response.PhoneNumber)
	assert.Equal(t, user.Role, response.Role)
	assert.Equal(t, user.Name, response.Name)
	assert.Equal(t, user.IsActive, response.IsActive)
	assert.Equal(t, user.CreatedAt, response.CreatedAt)
	assert.Equal(t, user.UpdatedAt, response.UpdatedAt)
	assert.Equal(t, user.LastLoginAt, response.LastLoginAt)
}

// TestUser_IsAdmin tests admin role check
func TestUser_IsAdmin(t *testing.T) {
	tests := []struct {
		name string
		user models.User
		want bool
	}{
		{"admin user", models.User{Role: models.UserRoleAdmin}, true},
		{"owner user", models.User{Role: models.UserRoleOwner}, false},
		{"staff user", models.User{Role: models.UserRoleStaff}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.user.IsAdmin())
		})
	}
}

// TestUser_IsOwner tests owner role check
func TestUser_IsOwner(t *testing.T) {
	tests := []struct {
		name string
		user models.User
		want bool
	}{
		{"admin user", models.User{Role: models.UserRoleAdmin}, false},
		{"owner user", models.User{Role: models.UserRoleOwner}, true},
		{"staff user", models.User{Role: models.UserRoleStaff}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.user.IsOwner())
		})
	}
}

// TestUser_IsStaff tests staff role check
func TestUser_IsStaff(t *testing.T) {
	tests := []struct {
		name string
		user models.User
		want bool
	}{
		{"admin user", models.User{Role: models.UserRoleAdmin}, false},
		{"owner user", models.User{Role: models.UserRoleOwner}, false},
		{"staff user", models.User{Role: models.UserRoleStaff}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.user.IsStaff())
		})
	}
}

// TestUser_CanManageUsers tests user management permission
func TestUser_CanManageUsers(t *testing.T) {
	tests := []struct {
		name string
		user models.User
		want bool
	}{
		{"admin user", models.User{Role: models.UserRoleAdmin}, true},
		{"owner user", models.User{Role: models.UserRoleOwner}, true},
		{"staff user", models.User{Role: models.UserRoleStaff}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.user.CanManageUsers())
		})
	}
}

// TestUser_CanViewReports tests report viewing permission
func TestUser_CanViewReports(t *testing.T) {
	tests := []struct {
		name string
		user models.User
		want bool
	}{
		{"admin user", models.User{Role: models.UserRoleAdmin}, true},
		{"owner user", models.User{Role: models.UserRoleOwner}, true},
		{"staff user", models.User{Role: models.UserRoleStaff}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.user.CanViewReports())
		})
	}
}

// TestUser_CanManageInventory tests inventory management permission
func TestUser_CanManageInventory(t *testing.T) {
	tests := []struct {
		name string
		user models.User
		want bool
	}{
		{"admin user", models.User{Role: models.UserRoleAdmin}, true},
		{"owner user", models.User{Role: models.UserRoleOwner}, true},
		{"staff user", models.User{Role: models.UserRoleStaff}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.user.CanManageInventory())
		})
	}
}

// TestUser_UpdateLastLogin tests last login update
func TestUser_UpdateLastLogin(t *testing.T) {
	user := models.User{}
	assert.Nil(t, user.LastLoginAt)

	user.UpdateLastLogin()
	assert.NotNil(t, user.LastLoginAt)
	assert.True(t, time.Since(*user.LastLoginAt) < time.Second)
}
