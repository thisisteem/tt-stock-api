package models

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// SessionStatus represents the status of a user session
type SessionStatus string

const (
	// SessionStatusActive represents an active session
	SessionStatusActive SessionStatus = "active"
	// SessionStatusExpired represents an expired session
	SessionStatusExpired SessionStatus = "expired"
	// SessionStatusRevoked represents a revoked session
	SessionStatusRevoked SessionStatus = "revoked"
)

// Session represents user authentication sessions
type Session struct {
	ID           uint          `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID       uint          `json:"userId" gorm:"not null;index"`
	Token        string        `json:"-" gorm:"uniqueIndex;not null;size:500"` // JWT token, not exposed in JSON
	RefreshToken string        `json:"-" gorm:"uniqueIndex;not null;size:500"` // Refresh token, not exposed in JSON
	Status       SessionStatus `json:"status" gorm:"not null;type:varchar(20);default:'active';index"`
	ExpiresAt    time.Time     `json:"expiresAt" gorm:"not null;index"`
	CreatedAt    time.Time     `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    time.Time     `json:"updatedAt" gorm:"autoUpdateTime"`
	LastUsedAt   *time.Time    `json:"lastUsedAt,omitempty" gorm:"index"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
}

// SessionCreateRequest represents the request payload for creating a session
type SessionCreateRequest struct {
	UserID       uint      `json:"userId" binding:"required"`
	Token        string    `json:"token" binding:"required"`
	RefreshToken string    `json:"refreshToken" binding:"required"`
	ExpiresAt    time.Time `json:"expiresAt" binding:"required"`
}

// SessionUpdateRequest represents the request payload for updating a session
type SessionUpdateRequest struct {
	Status     *SessionStatus `json:"status,omitempty"`
	ExpiresAt  *time.Time     `json:"expiresAt,omitempty"`
	LastUsedAt *time.Time     `json:"lastUsedAt,omitempty"`
}

// SessionResponse represents the response payload for session data
type SessionResponse struct {
	ID         uint          `json:"id"`
	UserID     uint          `json:"userId"`
	Status     SessionStatus `json:"status"`
	ExpiresAt  time.Time     `json:"expiresAt"`
	CreatedAt  time.Time     `json:"createdAt"`
	UpdatedAt  time.Time     `json:"updatedAt"`
	LastUsedAt *time.Time    `json:"lastUsedAt,omitempty"`
	User       *UserResponse `json:"user,omitempty"`
}

// SessionListRequest represents the request payload for listing sessions
type SessionListRequest struct {
	UserID  *uint          `json:"userId,omitempty"`
	Status  *SessionStatus `json:"status,omitempty"`
	Expired *bool          `json:"expired,omitempty"`
	Page    int            `json:"page" binding:"min=1"`
	Limit   int            `json:"limit" binding:"min=1,max=100"`
}

// SessionListResponse represents the response payload for session list
type SessionListResponse struct {
	Sessions   []SessionResponse  `json:"sessions"`
	Pagination PaginationResponse `json:"pagination"`
}

// BeforeCreate is a GORM hook that runs before creating a session
func (s *Session) BeforeCreate(_ *gorm.DB) error {
	// Validate session data
	if err := s.Validate(); err != nil {
		return err
	}

	// Set default status if not provided
	if s.Status == "" {
		s.Status = SessionStatusActive
	}

	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a session
func (s *Session) BeforeUpdate(_ *gorm.DB) error {
	// Validate session data
	if err := s.Validate(); err != nil {
		return err
	}

	return nil
}

// Validate validates the session data
func (s *Session) Validate() error {
	var validationErrors []string

	// Validate user ID
	if s.UserID == 0 {
		validationErrors = append(validationErrors, "userId is required")
	}

	// Validate token
	if err := s.ValidateToken(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate refresh token
	if err := s.ValidateRefreshToken(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate status
	if err := s.ValidateStatus(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate expires at
	if err := s.ValidateExpiresAt(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// ValidateToken validates the JWT token
func (s *Session) ValidateToken() error {
	if s.Token == "" {
		return errors.New("token is required")
	}

	const minTokenLength = 10
	if len(s.Token) < minTokenLength {
		return errors.New("token must be at least 10 characters")
	}

	return nil
}

// ValidateRefreshToken validates the refresh token
func (s *Session) ValidateRefreshToken() error {
	if s.RefreshToken == "" {
		return errors.New("refreshToken is required")
	}

	const minRefreshTokenLength = 10
	if len(s.RefreshToken) < minRefreshTokenLength {
		return errors.New("refreshToken must be at least 10 characters")
	}

	return nil
}

// ValidateStatus validates the session status
func (s *Session) ValidateStatus() error {
	if s.Status == "" {
		return nil // Will be set to default in BeforeCreate
	}

	validStatuses := []SessionStatus{
		SessionStatusActive,
		SessionStatusExpired,
		SessionStatusRevoked,
	}

	for _, status := range validStatuses {
		if s.Status == status {
			return nil
		}
	}

	return errors.New("status must be one of: active, expired, revoked")
}

// ValidateExpiresAt validates the expiration time
func (s *Session) ValidateExpiresAt() error {
	if s.ExpiresAt.IsZero() {
		return errors.New("expiresAt is required")
	}

	// Expiration time should be in the future
	if s.ExpiresAt.Before(time.Now()) {
		return errors.New("expiresAt must be in the future")
	}

	return nil
}

// IsActive checks if the session is active
func (s *Session) IsActive() bool {
	return s.Status == SessionStatusActive && s.ExpiresAt.After(time.Now())
}

// IsExpired checks if the session is expired
func (s *Session) IsExpired() bool {
	return s.ExpiresAt.Before(time.Now())
}

// IsRevoked checks if the session is revoked
func (s *Session) IsRevoked() bool {
	return s.Status == SessionStatusRevoked
}

// CanBeUsed checks if the session can be used
func (s *Session) CanBeUsed() bool {
	return s.IsActive() && !s.IsRevoked()
}

// UpdateLastUsed updates the last used timestamp
func (s *Session) UpdateLastUsed() {
	now := time.Now()
	s.LastUsedAt = &now
}

// Revoke revokes the session
func (s *Session) Revoke() {
	s.Status = SessionStatusRevoked
}

// Expire expires the session
func (s *Session) Expire() {
	s.Status = SessionStatusExpired
}

// Refresh updates the session with new tokens and expiration
func (s *Session) Refresh(token, refreshToken string, expiresAt time.Time) error {
	if !s.CanBeUsed() {
		return errors.New("cannot refresh an inactive or revoked session")
	}

	s.Token = token
	s.RefreshToken = refreshToken
	s.ExpiresAt = expiresAt
	s.UpdateLastUsed()

	return s.Validate()
}

// GetTimeUntilExpiry returns the time until the session expires
func (s *Session) GetTimeUntilExpiry() time.Duration {
	return time.Until(s.ExpiresAt)
}

// IsExpiringSoon checks if the session is expiring within the given duration
func (s *Session) IsExpiringSoon(duration time.Duration) bool {
	return s.GetTimeUntilExpiry() <= duration
}

// ToResponse converts a Session to SessionResponse
func (s *Session) ToResponse() SessionResponse {
	response := SessionResponse{
		ID:         s.ID,
		UserID:     s.UserID,
		Status:     s.Status,
		ExpiresAt:  s.ExpiresAt,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
		LastUsedAt: s.LastUsedAt,
	}

	// Include user data if it is loaded
	if s.User.ID != 0 {
		userResponse := s.User.ToResponse()
		response.User = &userResponse
	}

	return response
}

// TableName returns the table name for the Session model
func (Session) TableName() string {
	return "sessions"
}
