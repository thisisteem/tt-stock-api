// Package models contains the domain models for the TT Stock Backend API.
// It includes User, Product, ProductSpecification, StockMovement, Session, and Alert models
// with their respective validation rules and business logic.
package models

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserRole represents the role of a user in the system
type UserRole string

const (
	// UserRoleAdmin represents an admin user with full system access
	UserRoleAdmin UserRole = "Admin"
	// UserRoleOwner represents an owner user with business management access
	UserRoleOwner UserRole = "Owner"
	// UserRoleStaff represents a staff user with limited access
	UserRoleStaff UserRole = "Staff"
)

// User represents a system user with authentication and role-based access
type User struct {
	ID          uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	PhoneNumber string     `json:"phoneNumber" gorm:"uniqueIndex;not null;size:15"`
	PIN         string     `json:"-" gorm:"not null;size:255"` // Hashed PIN, not exposed in JSON
	Role        UserRole   `json:"role" gorm:"not null;type:varchar(20)"`
	Name        string     `json:"name" gorm:"not null;size:100"`
	IsActive    bool       `json:"isActive" gorm:"default:true"`
	CreatedAt   time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`
	LastLoginAt *time.Time `json:"lastLoginAt,omitempty" gorm:"index"`

	// Relationships
	StockMovements []StockMovement `json:"-" gorm:"foreignKey:UserID;references:ID"`
	Sessions       []Session       `json:"-" gorm:"foreignKey:UserID;references:ID"`
}

// UserCreateRequest represents the request payload for creating a user
type UserCreateRequest struct {
	PhoneNumber string   `json:"phoneNumber" binding:"required"`
	PIN         string   `json:"pin" binding:"required"`
	Role        UserRole `json:"role" binding:"required"`
	Name        string   `json:"name" binding:"required"`
}

// UserUpdateRequest represents the request payload for updating a user
type UserUpdateRequest struct {
	Role     *UserRole `json:"role,omitempty"`
	Name     *string   `json:"name,omitempty"`
	IsActive *bool     `json:"isActive,omitempty"`
}

// UserResponse represents the response payload for user data
type UserResponse struct {
	ID          uint       `json:"id"`
	PhoneNumber string     `json:"phoneNumber"`
	Role        UserRole   `json:"role"`
	Name        string     `json:"name"`
	IsActive    bool       `json:"isActive"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	LastLoginAt *time.Time `json:"lastLoginAt,omitempty"`
}

// UserLoginRequest represents the request payload for user login
type UserLoginRequest struct {
	PhoneNumber string `json:"phoneNumber" binding:"required"`
	PIN         string `json:"pin" binding:"required"`
}

// UserLoginResponse represents the response payload for user login
type UserLoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expiresAt"`
	User      UserResponse `json:"user"`
}

// UserListRequest represents the request payload for listing users
type UserListRequest struct {
	Role     *UserRole `json:"role,omitempty"`
	IsActive *bool     `json:"isActive,omitempty"`
	Search   *string   `json:"search,omitempty"`
	Page     int       `json:"page" binding:"min=1"`
	Limit    int       `json:"limit" binding:"min=1,max=100"`
}

// UserListResponse represents the response payload for user list
type UserListResponse struct {
	Users      []UserResponse     `json:"users"`
	Pagination PaginationResponse `json:"pagination"`
}

// BeforeCreate is a GORM hook that runs before creating a user
func (u *User) BeforeCreate(_ *gorm.DB) error {
	// Validate user data
	if err := u.Validate(); err != nil {
		return err
	}

	// Hash the PIN
	hashedPIN, err := u.HashPIN(u.PIN)
	if err != nil {
		return err
	}
	u.PIN = hashedPIN

	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a user
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// Validate user data
	if err := u.Validate(); err != nil {
		return err
	}

	// If PIN is being updated, hash it
	if tx.Statement.Changed("PIN") {
		hashedPIN, err := u.HashPIN(u.PIN)
		if err != nil {
			return err
		}
		u.PIN = hashedPIN
	}

	return nil
}

// Validate validates the user data
func (u *User) Validate() error {
	var validationErrors []string

	// Validate phone number
	if err := u.ValidatePhoneNumber(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate PIN
	if err := u.ValidatePIN(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate role
	if err := u.ValidateRole(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate name
	if err := u.ValidateName(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// ValidatePhoneNumber validates the phone number format
func (u *User) ValidatePhoneNumber() error {
	if u.PhoneNumber == "" {
		return errors.New("phone number is required")
	}

	// Remove any non-digit characters for validation
	phoneDigits := regexp.MustCompile(`\D`).ReplaceAllString(u.PhoneNumber, "")

	// Check if phone number has 10-15 digits
	if len(phoneDigits) < 10 || len(phoneDigits) > 15 {
		return errors.New("phone number must be 10-15 digits")
	}

	return nil
}

// ValidatePIN validates the PIN format
func (u *User) ValidatePIN() error {
	if u.PIN == "" {
		return errors.New("PIN is required")
	}

	// Check if PIN is already hashed (starts with $2a$)
	if strings.HasPrefix(u.PIN, "$2a$") {
		return nil // Already hashed, skip validation
	}

	// Check if PIN is 4-6 digits
	if len(u.PIN) < 4 || len(u.PIN) > 6 {
		return errors.New("PIN must be 4-6 digits")
	}

	// Check if PIN contains only digits
	if !regexp.MustCompile(`^\d+$`).MatchString(u.PIN) {
		return errors.New("PIN must contain only digits")
	}

	return nil
}

// ValidateRole validates the user role
func (u *User) ValidateRole() error {
	validRoles := []UserRole{UserRoleAdmin, UserRoleOwner, UserRoleStaff}

	for _, role := range validRoles {
		if u.Role == role {
			return nil
		}
	}

	return errors.New("role must be one of: Admin, Owner, Staff")
}

// ValidateName validates the user name
func (u *User) ValidateName() error {
	if u.Name == "" {
		return errors.New("name is required")
	}

	if len(u.Name) < 2 || len(u.Name) > 100 {
		return errors.New("name must be 2-100 characters")
	}

	return nil
}

// HashPIN hashes the PIN using bcrypt
func (u *User) HashPIN(pin string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPIN verifies the provided PIN against the stored hash
func (u *User) VerifyPIN(pin string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PIN), []byte(pin))
	return err == nil
}

// ToResponse converts a User to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:          u.ID,
		PhoneNumber: u.PhoneNumber,
		Role:        u.Role,
		Name:        u.Name,
		IsActive:    u.IsActive,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		LastLoginAt: u.LastLoginAt,
	}
}

// UpdateLastLogin updates the last login timestamp
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
}

// IsAdmin checks if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

// IsOwner checks if the user has owner role
func (u *User) IsOwner() bool {
	return u.Role == UserRoleOwner
}

// IsStaff checks if the user has staff role
func (u *User) IsStaff() bool {
	return u.Role == UserRoleStaff
}

// CanManageUsers checks if the user can manage other users
func (u *User) CanManageUsers() bool {
	return u.Role == UserRoleAdmin || u.Role == UserRoleOwner
}

// CanViewReports checks if the user can view business reports
func (u *User) CanViewReports() bool {
	return u.Role == UserRoleAdmin || u.Role == UserRoleOwner
}

// CanManageInventory checks if the user can manage inventory
func (u *User) CanManageInventory() bool {
	return u.Role == UserRoleAdmin || u.Role == UserRoleOwner || u.Role == UserRoleStaff
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
}
