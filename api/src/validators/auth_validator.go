// Package validators contains input validation logic for the TT Stock Backend API.
// It provides validation structs, rules, and functions to ensure data integrity
// and security at the API boundary layer.
package validators

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"tt-stock-api/src/models"
)

// AuthValidator handles authentication-related input validation
type AuthValidator struct{}

// NewAuthValidator creates a new AuthValidator instance
func NewAuthValidator() *AuthValidator {
	return &AuthValidator{}
}

// ValidateLoginRequest validates user login request data
func (v *AuthValidator) ValidateLoginRequest(req *models.UserLoginRequest) error {
	if req == nil {
		return errors.New("login request cannot be nil")
	}

	var validationErrors []string

	// Validate phone number
	if err := v.validatePhoneNumber(req.PhoneNumber); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("phone number: %s", err.Error()))
	}

	// Validate PIN
	if err := v.validatePIN(req.PIN); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("PIN: %s", err.Error()))
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// ValidateRefreshTokenRequest validates refresh token request data
func (v *AuthValidator) ValidateRefreshTokenRequest(req *RefreshTokenRequest) error {
	if req == nil {
		return errors.New("refresh token request cannot be nil")
	}

	if strings.TrimSpace(req.RefreshToken) == "" {
		return errors.New("refresh token is required")
	}

	if len(req.RefreshToken) < 32 {
		return errors.New("refresh token must be at least 32 characters")
	}

	if len(req.RefreshToken) > 512 {
		return errors.New("refresh token must not exceed 512 characters")
	}

	return nil
}

// ValidateUserCreateRequest validates user creation request data
func (v *AuthValidator) ValidateUserCreateRequest(req *models.UserCreateRequest) error {
	if req == nil {
		return errors.New("user creation request cannot be nil")
	}

	var validationErrors []string

	// Validate phone number
	if err := v.validatePhoneNumber(req.PhoneNumber); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("phone number: %s", err.Error()))
	}

	// Validate PIN
	if err := v.validatePIN(req.PIN); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("PIN: %s", err.Error()))
	}

	// Validate role
	if err := v.validateRole(req.Role); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("role: %s", err.Error()))
	}

	// Validate name
	if err := v.validateName(req.Name); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("name: %s", err.Error()))
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// ValidateUserUpdateRequest validates user update request data
func (v *AuthValidator) ValidateUserUpdateRequest(req *models.UserUpdateRequest) error {
	if req == nil {
		return errors.New("user update request cannot be nil")
	}

	var validationErrors []string

	// Note: UserUpdateRequest only has Role, Name, and IsActive fields
	// Phone number and PIN updates are handled separately

	// Validate role if provided
	if req.Role != nil {
		if err := v.validateRole(*req.Role); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("role: %s", err.Error()))
		}
	}

	// Validate name if provided
	if req.Name != nil {
		if err := v.validateName(*req.Name); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("name: %s", err.Error()))
		}
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// ValidatePINChangeRequest validates PIN change request data
func (v *AuthValidator) ValidatePINChangeRequest(req *PINChangeRequest) error {
	if req == nil {
		return errors.New("PIN change request cannot be nil")
	}

	var validationErrors []string

	// Validate current PIN
	if err := v.validatePIN(req.CurrentPIN); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("current PIN: %s", err.Error()))
	}

	// Validate new PIN
	if err := v.validatePIN(req.NewPIN); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("new PIN: %s", err.Error()))
	}

	// Ensure new PIN is different from current PIN
	if req.CurrentPIN == req.NewPIN {
		validationErrors = append(validationErrors, "new PIN must be different from current PIN")
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// Helper validation methods

// validatePhoneNumber validates phone number format
func (v *AuthValidator) validatePhoneNumber(phoneNumber string) error {
	if strings.TrimSpace(phoneNumber) == "" {
		return errors.New("phone number is required")
	}

	// Remove all non-digit characters for validation
	digitsOnly := regexp.MustCompile(`\D`).ReplaceAllString(phoneNumber, "")

	if len(digitsOnly) < 10 {
		return errors.New("phone number must have at least 10 digits")
	}

	if len(digitsOnly) > 15 {
		return errors.New("phone number must not exceed 15 digits")
	}

	// Check if it contains only digits and common separators
	phoneRegex := regexp.MustCompile(`^[\+]?[\d\s\-\(\)\.]+$`)
	if !phoneRegex.MatchString(phoneNumber) {
		return errors.New("phone number contains invalid characters")
	}

	return nil
}

// validatePIN validates PIN format
func (v *AuthValidator) validatePIN(pin string) error {
	if strings.TrimSpace(pin) == "" {
		return errors.New("PIN is required")
	}

	if len(pin) < 4 {
		return errors.New("PIN must be at least 4 digits")
	}

	if len(pin) > 6 {
		return errors.New("PIN must not exceed 6 digits")
	}

	// Check if PIN contains only digits
	for _, char := range pin {
		if !unicode.IsDigit(char) {
			return errors.New("PIN must contain only digits")
		}
	}

	return nil
}

// validateRole validates user role
func (v *AuthValidator) validateRole(role models.UserRole) error {
	switch role {
	case models.UserRoleAdmin, models.UserRoleOwner, models.UserRoleStaff:
		return nil
	default:
		return errors.New("role must be one of: Admin, Owner, Staff")
	}
}

// validateName validates user name
func (v *AuthValidator) validateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("name is required")
	}

	trimmedName := strings.TrimSpace(name)
	if len(trimmedName) < 2 {
		return errors.New("name must be at least 2 characters")
	}

	if len(trimmedName) > 100 {
		return errors.New("name must not exceed 100 characters")
	}

	// Check if name contains only letters, spaces, hyphens, and apostrophes
	nameRegex := regexp.MustCompile(`^[a-zA-Z\s\-']+$`)
	if !nameRegex.MatchString(trimmedName) {
		return errors.New("name can only contain letters, spaces, hyphens, and apostrophes")
	}

	return nil
}

// ValidateSessionRequest validates session-related requests
func (v *AuthValidator) ValidateSessionRequest(req *SessionRequest) error {
	if req == nil {
		return errors.New("session request cannot be nil")
	}

	var validationErrors []string

	// Validate device info if provided
	if req.DeviceInfo != nil {
		if err := v.validateDeviceInfo(*req.DeviceInfo); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("device info: %s", err.Error()))
		}
	}

	// Validate IP address if provided
	if req.IPAddress != nil {
		if err := v.validateIPAddress(*req.IPAddress); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("IP address: %s", err.Error()))
		}
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// validateDeviceInfo validates device information
func (v *AuthValidator) validateDeviceInfo(deviceInfo string) error {
	if strings.TrimSpace(deviceInfo) == "" {
		return nil // Device info is optional
	}

	if len(deviceInfo) > 500 {
		return errors.New("device info must not exceed 500 characters")
	}

	return nil
}

// validateIPAddress validates IP address format
func (v *AuthValidator) validateIPAddress(ipAddress string) error {
	if strings.TrimSpace(ipAddress) == "" {
		return nil // IP address is optional
	}

	// Basic IP address validation (IPv4 and IPv6)
	ipv4Regex := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	ipv6Regex := regexp.MustCompile(`^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$`)

	if !ipv4Regex.MatchString(ipAddress) && !ipv6Regex.MatchString(ipAddress) {
		return errors.New("invalid IP address format")
	}

	return nil
}

// Validation constants
const (
	MinPhoneNumberLength  = 10
	MaxPhoneNumberLength  = 15
	MinPINLength          = 4
	MaxPINLength          = 6
	MinNameLength         = 2
	MaxNameLength         = 100
	MaxDeviceInfoLength   = 500
	MinRefreshTokenLength = 32
	MaxRefreshTokenLength = 512
)

// RefreshTokenRequest represents a refresh token request for validation
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// PINChangeRequest represents a PIN change request for validation
type PINChangeRequest struct {
	CurrentPIN string `json:"currentPin" binding:"required"`
	NewPIN     string `json:"newPin" binding:"required"`
}

// SessionRequest represents a session request for validation
type SessionRequest struct {
	DeviceInfo *string `json:"deviceInfo,omitempty"`
	IPAddress  *string `json:"ipAddress,omitempty"`
}
