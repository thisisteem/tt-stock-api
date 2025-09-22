// Package services contains the business logic layer implementations for the TT Stock Backend API.
// It provides service interfaces and implementations that orchestrate repository operations
// and implement business rules and validation logic.
package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"tt-stock-api/src/models"
	"tt-stock-api/src/repositories"

	"github.com/golang-jwt/jwt/v5"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	// Login authenticates a user and returns a JWT token
	Login(ctx context.Context, req *models.UserLoginRequest) (*models.UserLoginResponse, error)

	// Logout invalidates a user's session
	Logout(ctx context.Context, token string) error

	// RefreshToken generates a new token using a refresh token
	RefreshToken(ctx context.Context, refreshToken string) (*models.UserLoginResponse, error)

	// ValidateToken validates a JWT token and returns user information
	ValidateToken(ctx context.Context, token string) (*models.User, error)

	// GetUserFromToken extracts user information from a JWT token
	GetUserFromToken(ctx context.Context, token string) (*models.User, error)

	// RevokeAllUserSessions revokes all sessions for a user
	RevokeAllUserSessions(ctx context.Context, userID uint) error
}

// authService implements the AuthService interface
type authService struct {
	userRepo    repositories.UserRepository
	sessionRepo repositories.SessionRepository
	jwtSecret   string
	tokenExpiry time.Duration
}

// NewAuthService creates a new AuthService instance
func NewAuthService(
	userRepo repositories.UserRepository,
	sessionRepo repositories.SessionRepository,
	jwtSecret string,
	tokenExpiry time.Duration,
) AuthService {
	return &authService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtSecret:   jwtSecret,
		tokenExpiry: tokenExpiry,
	}
}

// Login authenticates a user and returns a JWT token
func (s *authService) Login(ctx context.Context, req *models.UserLoginRequest) (*models.UserLoginResponse, error) {
	if req == nil {
		return nil, errors.New("login request cannot be nil")
	}

	// Validate input
	if req.PhoneNumber == "" {
		return nil, errors.New("phone number is required")
	}
	if req.PIN == "" {
		return nil, errors.New("PIN is required")
	}

	// Get user by phone number
	user, err := s.userRepo.GetByPhoneNumber(ctx, req.PhoneNumber)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Verify PIN
	if !user.VerifyPIN(req.PIN) {
		return nil, errors.New("invalid credentials")
	}

	// Generate JWT token
	token, refreshToken, expiresAt, err := s.generateTokens(user.ID, user.PhoneNumber, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create session
	session := &models.Session{
		UserID:       user.ID,
		Token:        token,
		RefreshToken: refreshToken,
		Status:       models.SessionStatusActive,
		ExpiresAt:    expiresAt,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Log error but don't fail the login
		// TODO: Add proper logging
	}

	return &models.UserLoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user.ToResponse(),
	}, nil
}

// Logout invalidates a user's session
func (s *authService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("token is required")
	}

	// Get session by token
	session, err := s.sessionRepo.GetByToken(ctx, token)
	if err != nil {
		return errors.New("invalid session")
	}

	// Revoke session
	return s.sessionRepo.RevokeSession(ctx, session.ID)
}

// RefreshToken generates a new token using a refresh token
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*models.UserLoginResponse, error) {
	if refreshToken == "" {
		return nil, errors.New("refresh token is required")
	}

	// Get session by refresh token
	session, err := s.sessionRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Check if session is active
	if !session.CanBeUsed() {
		return nil, errors.New("session is expired or revoked")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Check if user is still active
	if !user.IsActive {
		return nil, errors.New("user account is deactivated")
	}

	// Generate new tokens
	newToken, newRefreshToken, newExpiresAt, err := s.generateTokens(user.ID, user.PhoneNumber, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	// Update session with new tokens
	if err := session.Refresh(newToken, newRefreshToken, newExpiresAt); err != nil {
		return nil, fmt.Errorf("failed to refresh session: %w", err)
	}

	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return &models.UserLoginResponse{
		Token:     newToken,
		ExpiresAt: newExpiresAt,
		User:      user.ToResponse(),
	}, nil
}

// ValidateToken validates a JWT token and returns user information
func (s *authService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}

	// Parse and validate JWT token
	claims, err := s.parseToken(token)
	if err != nil {
		return nil, errors.New("invalid token")
	}

	// Get user from token claims
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Get user from database
	user, err := s.userRepo.GetByID(ctx, uint(userID))
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("user account is deactivated")
	}

	// Check if session is still valid
	session, err := s.sessionRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, errors.New("session not found")
	}

	if !session.CanBeUsed() {
		return nil, errors.New("session is expired or revoked")
	}

	// Update last used timestamp
	session.UpdateLastUsed()
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		// Log error but don't fail validation
		// TODO: Add proper logging
	}

	return user, nil
}

// GetUserFromToken extracts user information from a JWT token
func (s *authService) GetUserFromToken(ctx context.Context, token string) (*models.User, error) {
	return s.ValidateToken(ctx, token)
}

// RevokeAllUserSessions revokes all sessions for a user
func (s *authService) RevokeAllUserSessions(ctx context.Context, userID uint) error {
	if userID == 0 {
		return errors.New("user ID cannot be zero")
	}

	return s.sessionRepo.RevokeAllUserSessions(ctx, userID)
}

// generateTokens generates JWT token and refresh token
func (s *authService) generateTokens(userID uint, phoneNumber string, role models.UserRole) (string, string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(s.tokenExpiry)

	// Create JWT claims
	claims := jwt.MapClaims{
		"user_id":      userID,
		"phone_number": phoneNumber,
		"role":         string(role),
		"exp":          expiresAt.Unix(),
		"iat":          now.Unix(),
		"type":         "access",
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", time.Time{}, err
	}

	// Generate refresh token (simple random string for now)
	// In production, use a more secure method
	refreshToken := fmt.Sprintf("refresh_%d_%d", userID, now.Unix())

	return tokenString, refreshToken, expiresAt, nil
}

// parseToken parses and validates a JWT token
func (s *authService) parseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
