package auth

import (
	"errors"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"tt-stock-api/internal/config"
	"tt-stock-api/internal/user"
	"tt-stock-api/pkg/utils"
)

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Access token expiration in seconds
}

// Claims represents JWT token claims
type Claims struct {
	UserID      uuid.UUID `json:"user_id"`
	PhoneNumber string    `json:"phone_number"`
	TokenType   string    `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// Service defines the interface for authentication operations
type Service interface {
	ValidatePhoneNumber(phoneNumber string) error
	ValidatePin(pin string) error
	AuthenticateUser(phoneNumber, pin string) (*user.User, error)
	GenerateAccessToken(userID uuid.UUID, phoneNumber string) (string, error)
	GenerateRefreshToken(userID uuid.UUID, phoneNumber string) (string, error)
	GenerateTokens(userID uuid.UUID, phoneNumber string) (*TokenPair, error)
	ValidateToken(tokenString string) (*Claims, error)
	ParseToken(tokenString string) (*Claims, error)
	BlacklistToken(tokenString string) error
	IsTokenBlacklisted(tokenString string) (bool, error)
}

// service implements the Service interface
type service struct {
	userRepo       user.Repository
	blacklistRepo  BlacklistRepository
	jwtSecret      string
}

// NewService creates a new authentication service instance
func NewService(userRepo user.Repository, blacklistRepo BlacklistRepository, cfg *config.Config) Service {
	return &service{
		userRepo:      userRepo,
		blacklistRepo: blacklistRepo,
		jwtSecret:     cfg.JWTSecret,
	}
}

// ValidatePhoneNumber validates Thai phone number format (^0[0-9]{9}$)
func (s *service) ValidatePhoneNumber(phoneNumber string) error {
	if phoneNumber == "" {
		return errors.New("phone number is required")
	}

	// Thai phone number format: starts with 0, followed by 9 digits (total 10 digits)
	thaiPhoneRegex := regexp.MustCompile(`^0[0-9]{9}$`)
	if !thaiPhoneRegex.MatchString(phoneNumber) {
		return errors.New("invalid phone number format: must be 10 digits starting with 0")
	}

	return nil
}

// ValidatePin validates 6-digit PIN format (^[0-9]{6}$)
func (s *service) ValidatePin(pin string) error {
	if pin == "" {
		return errors.New("PIN is required")
	}

	// PIN format: exactly 6 digits
	pinRegex := regexp.MustCompile(`^[0-9]{6}$`)
	if !pinRegex.MatchString(pin) {
		return errors.New("invalid PIN format: must be exactly 6 digits")
	}

	return nil
}

// AuthenticateUser validates user credentials and returns the user if authentication succeeds
func (s *service) AuthenticateUser(phoneNumber, pin string) (*user.User, error) {
	// Validate input format
	if err := s.ValidatePhoneNumber(phoneNumber); err != nil {
		return nil, err
	}

	if err := s.ValidatePin(pin); err != nil {
		return nil, err
	}

	// Find user by phone number
	foundUser, err := s.userRepo.FindByPhoneNumber(phoneNumber)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Verify PIN against stored hash
	if err := utils.CheckPin(foundUser.PinHash, pin); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Update last login timestamp
	if err := s.userRepo.UpdateLastLogin(foundUser.ID); err != nil {
		// Log error but don't fail authentication
		// In a real application, you'd use a proper logger here
	}

	return foundUser, nil
}

// GenerateAccessToken creates a new access token with 15-minute expiration
func (s *service) GenerateAccessToken(userID uuid.UUID, phoneNumber string) (string, error) {
	expirationTime := time.Now().Add(15 * time.Minute)
	
	claims := &Claims{
		UserID:      userID,
		PhoneNumber: phoneNumber,
		TokenType:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "tt-stock-api",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", errors.New("failed to generate access token")
	}

	return tokenString, nil
}

// GenerateRefreshToken creates a new refresh token with 1-day expiration
func (s *service) GenerateRefreshToken(userID uuid.UUID, phoneNumber string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	
	claims := &Claims{
		UserID:      userID,
		PhoneNumber: phoneNumber,
		TokenType:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "tt-stock-api",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", errors.New("failed to generate refresh token")
	}

	return tokenString, nil
}

// GenerateTokens creates both access and refresh tokens for a user
func (s *service) GenerateTokens(userID uuid.UUID, phoneNumber string) (*TokenPair, error) {
	accessToken, err := s.GenerateAccessToken(userID, phoneNumber)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.GenerateRefreshToken(userID, phoneNumber)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    15 * 60, // 15 minutes in seconds
	}, nil
}

// ValidateToken validates a JWT token and returns its claims
func (s *service) ValidateToken(tokenString string) (*Claims, error) {
	// First check if token is blacklisted
	isBlacklisted, err := s.IsTokenBlacklisted(tokenString)
	if err != nil {
		return nil, errors.New("failed to check token blacklist status")
	}
	if isBlacklisted {
		return nil, errors.New("token has been invalidated")
	}

	// Then parse and validate the token
	return s.ParseToken(tokenString)
}

// ParseToken parses and validates a JWT token, returning its claims
func (s *service) ParseToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, errors.New("token is required")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid token signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}

// BlacklistToken adds a token to the blacklist to invalidate it
func (s *service) BlacklistToken(tokenString string) error {
	if tokenString == "" {
		return errors.New("token is required")
	}

	// Parse the token to get its claims
	claims, err := s.ParseToken(tokenString)
	if err != nil {
		return errors.New("invalid token")
	}

	// Add token to blacklist
	expiresAt := claims.ExpiresAt.Time
	err = s.blacklistRepo.BlacklistToken(tokenString, claims.UserID.String(), claims.TokenType, expiresAt)
	if err != nil {
		return errors.New("failed to blacklist token")
	}

	return nil
}

// IsTokenBlacklisted checks if a token is in the blacklist
func (s *service) IsTokenBlacklisted(tokenString string) (bool, error) {
	if tokenString == "" {
		return false, errors.New("token is required")
	}

	return s.blacklistRepo.IsTokenBlacklisted(tokenString)
}