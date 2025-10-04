package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"tt-stock-api/internal/config"
	"tt-stock-api/internal/user"
	"tt-stock-api/pkg/utils"
)

// MockUserRepository is a mock implementation of user.Repository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByPhoneNumber(phoneNumber string) (*user.User, error) {
	args := m.Called(phoneNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) UpdateLastLogin(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

// MockBlacklistRepository is a mock implementation of BlacklistRepository
type MockBlacklistRepository struct {
	mock.Mock
}

func (m *MockBlacklistRepository) BlacklistToken(token, userID, tokenType string, expiresAt time.Time) error {
	args := m.Called(token, userID, tokenType, expiresAt)
	return args.Error(0)
}

func (m *MockBlacklistRepository) IsTokenBlacklisted(token string) (bool, error) {
	args := m.Called(token)
	return args.Bool(0), args.Error(1)
}

// Test setup helper
func setupTestService() (*service, *MockUserRepository, *MockBlacklistRepository) {
	mockUserRepo := &MockUserRepository{}
	mockBlacklistRepo := &MockBlacklistRepository{}
	cfg := &config.Config{
		JWTSecret: "test-secret-key",
	}
	
	svc := &service{
		userRepo:      mockUserRepo,
		blacklistRepo: mockBlacklistRepo,
		jwtSecret:     cfg.JWTSecret,
	}
	
	return svc, mockUserRepo, mockBlacklistRepo
}

func TestValidatePhoneNumber(t *testing.T) {
	svc, _, _ := setupTestService()

	tests := []struct {
		name        string
		phoneNumber string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid Thai phone number",
			phoneNumber: "0812345678",
			expectError: false,
		},
		{
			name:        "Valid Thai phone number with different prefix",
			phoneNumber: "0987654321",
			expectError: false,
		},
		{
			name:        "Empty phone number",
			phoneNumber: "",
			expectError: true,
			errorMsg:    "phone number is required",
		},
		{
			name:        "Phone number too short",
			phoneNumber: "081234567",
			expectError: true,
			errorMsg:    "invalid phone number format: must be 10 digits starting with 0",
		},
		{
			name:        "Phone number too long",
			phoneNumber: "08123456789",
			expectError: true,
			errorMsg:    "invalid phone number format: must be 10 digits starting with 0",
		},
		{
			name:        "Phone number not starting with 0",
			phoneNumber: "1812345678",
			expectError: true,
			errorMsg:    "invalid phone number format: must be 10 digits starting with 0",
		},
		{
			name:        "Phone number with non-digits",
			phoneNumber: "081234567a",
			expectError: true,
			errorMsg:    "invalid phone number format: must be 10 digits starting with 0",
		},
		{
			name:        "Phone number with spaces",
			phoneNumber: "081 234 5678",
			expectError: true,
			errorMsg:    "invalid phone number format: must be 10 digits starting with 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidatePhoneNumber(tt.phoneNumber)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePin(t *testing.T) {
	svc, _, _ := setupTestService()

	tests := []struct {
		name        string
		pin         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid 6-digit PIN",
			pin:         "123456",
			expectError: false,
		},
		{
			name:        "Valid PIN with zeros",
			pin:         "000000",
			expectError: false,
		},
		{
			name:        "Empty PIN",
			pin:         "",
			expectError: true,
			errorMsg:    "PIN is required",
		},
		{
			name:        "PIN too short",
			pin:         "12345",
			expectError: true,
			errorMsg:    "invalid PIN format: must be exactly 6 digits",
		},
		{
			name:        "PIN too long",
			pin:         "1234567",
			expectError: true,
			errorMsg:    "invalid PIN format: must be exactly 6 digits",
		},
		{
			name:        "PIN with letters",
			pin:         "12345a",
			expectError: true,
			errorMsg:    "invalid PIN format: must be exactly 6 digits",
		},
		{
			name:        "PIN with special characters",
			pin:         "12345!",
			expectError: true,
			errorMsg:    "invalid PIN format: must be exactly 6 digits",
		},
		{
			name:        "PIN with spaces",
			pin:         "123 456",
			expectError: true,
			errorMsg:    "invalid PIN format: must be exactly 6 digits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidatePin(tt.pin)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthenticateUser(t *testing.T) {
	svc, mockUserRepo, _ := setupTestService()

	// Create a test user with hashed PIN
	testUserID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	hashedPin, _ := utils.HashPin("123456")
	testUser := &user.User{
		ID:          testUserID,
		PhoneNumber: "0812345678",
		PinHash:     hashedPin,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name         string
		phoneNumber  string
		pin          string
		setupMocks   func()
		expectError  bool
		errorMsg     string
		expectedUser *user.User
	}{
		{
			name:        "Successful authentication",
			phoneNumber: "0812345678",
			pin:         "123456",
			setupMocks: func() {
				mockUserRepo.On("FindByPhoneNumber", "0812345678").Return(testUser, nil).Once()
				mockUserRepo.On("UpdateLastLogin", testUserID).Return(nil).Once()
			},
			expectError:  false,
			expectedUser: testUser,
		},
		{
			name:        "Invalid phone number format",
			phoneNumber: "invalid",
			pin:         "123456",
			setupMocks:  func() {},
			expectError: true,
			errorMsg:    "invalid phone number format: must be 10 digits starting with 0",
		},
		{
			name:        "Invalid PIN format",
			phoneNumber: "0812345678",
			pin:         "123",
			setupMocks:  func() {},
			expectError: true,
			errorMsg:    "invalid PIN format: must be exactly 6 digits",
		},
		{
			name:        "User not found",
			phoneNumber: "0812345678",
			pin:         "123456",
			setupMocks: func() {
				mockUserRepo.On("FindByPhoneNumber", "0812345678").Return(nil, errors.New("user not found")).Once()
			},
			expectError: true,
			errorMsg:    "invalid credentials",
		},
		{
			name:        "Wrong PIN",
			phoneNumber: "0812345678",
			pin:         "654321",
			setupMocks: func() {
				mockUserRepo.On("FindByPhoneNumber", "0812345678").Return(testUser, nil).Once()
			},
			expectError: true,
			errorMsg:    "invalid credentials",
		},
		{
			name:        "UpdateLastLogin fails but authentication succeeds",
			phoneNumber: "0812345678",
			pin:         "123456",
			setupMocks: func() {
				mockUserRepo.On("FindByPhoneNumber", "0812345678").Return(testUser, nil).Once()
				mockUserRepo.On("UpdateLastLogin", testUserID).Return(errors.New("db error")).Once()
			},
			expectError:  false,
			expectedUser: testUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockUserRepo.ExpectedCalls = nil
			
			// Setup mocks for this test
			tt.setupMocks()
			
			// Execute test
			result, err := svc.AuthenticateUser(tt.phoneNumber, tt.pin)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedUser.ID, result.ID)
				assert.Equal(t, tt.expectedUser.PhoneNumber, result.PhoneNumber)
			}
			
			// Verify all expectations were met
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestGenerateAccessToken(t *testing.T) {
	svc, _, _ := setupTestService()

	testUserID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name        string
		userID      uuid.UUID
		phoneNumber string
		expectError bool
	}{
		{
			name:        "Valid token generation",
			userID:      testUserID,
			phoneNumber: "0812345678",
			expectError: false,
		},
		{
			name:        "Empty userID",
			userID:      uuid.Nil,
			phoneNumber: "0812345678",
			expectError: false, // JWT library handles empty values
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := svc.GenerateAccessToken(tt.userID, tt.phoneNumber)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				
				// Verify token can be parsed and has correct claims
				claims, parseErr := svc.ParseToken(token)
				assert.NoError(t, parseErr)
				assert.Equal(t, tt.userID, claims.UserID)
				assert.Equal(t, tt.phoneNumber, claims.PhoneNumber)
				assert.Equal(t, "access", claims.TokenType)
				assert.Equal(t, "tt-stock-api", claims.Issuer)
				
				// Verify expiration is approximately 15 minutes from now
				expectedExpiry := time.Now().Add(15 * time.Minute)
				assert.WithinDuration(t, expectedExpiry, claims.ExpiresAt.Time, time.Minute)
			}
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	svc, _, _ := setupTestService()

	testUserID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name        string
		userID      uuid.UUID
		phoneNumber string
		expectError bool
	}{
		{
			name:        "Valid refresh token generation",
			userID:      testUserID,
			phoneNumber: "0812345678",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := svc.GenerateRefreshToken(tt.userID, tt.phoneNumber)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				
				// Verify token can be parsed and has correct claims
				claims, parseErr := svc.ParseToken(token)
				assert.NoError(t, parseErr)
				assert.Equal(t, tt.userID, claims.UserID)
				assert.Equal(t, tt.phoneNumber, claims.PhoneNumber)
				assert.Equal(t, "refresh", claims.TokenType)
				assert.Equal(t, "tt-stock-api", claims.Issuer)
				
				// Verify expiration is approximately 24 hours from now
				expectedExpiry := time.Now().Add(24 * time.Hour)
				assert.WithinDuration(t, expectedExpiry, claims.ExpiresAt.Time, time.Minute)
			}
		})
	}
}

func TestGenerateTokens(t *testing.T) {
	svc, _, _ := setupTestService()

	t.Run("Generate both tokens successfully", func(t *testing.T) {
		userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		phoneNumber := "0812345678"
		
		tokenPair, err := svc.GenerateTokens(userID, phoneNumber)
		
		assert.NoError(t, err)
		assert.NotNil(t, tokenPair)
		assert.NotEmpty(t, tokenPair.AccessToken)
		assert.NotEmpty(t, tokenPair.RefreshToken)
		assert.Equal(t, int64(15*60), tokenPair.ExpiresIn) // 15 minutes in seconds
		
		// Verify both tokens are valid and have correct types
		accessClaims, err := svc.ParseToken(tokenPair.AccessToken)
		assert.NoError(t, err)
		assert.Equal(t, "access", accessClaims.TokenType)
		
		refreshClaims, err := svc.ParseToken(tokenPair.RefreshToken)
		assert.NoError(t, err)
		assert.Equal(t, "refresh", refreshClaims.TokenType)
	})
}

func TestParseToken(t *testing.T) {
	svc, _, _ := setupTestService()

	testUserID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	// Generate a valid token for testing
	validToken, _ := svc.GenerateAccessToken(testUserID, "0812345678")
	
	// Create an expired token
	expiredClaims := &Claims{
		UserID:      testUserID,
		PhoneNumber: "0812345678",
		TokenType:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "tt-stock-api",
			Subject:   testUserID.String(),
		},
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenString, _ := expiredToken.SignedString([]byte("test-secret-key"))

	tests := []struct {
		name        string
		token       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid token",
			token:       validToken,
			expectError: false,
		},
		{
			name:        "Empty token",
			token:       "",
			expectError: true,
			errorMsg:    "token is required",
		},
		{
			name:        "Invalid token format",
			token:       "invalid.token.format",
			expectError: true,
			errorMsg:    "invalid token",
		},
		{
			name:        "Token with wrong signature",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidXNlci0xMjMiLCJwaG9uZV9udW1iZXIiOiIwODEyMzQ1Njc4IiwidG9rZW5fdHlwZSI6ImFjY2VzcyIsImV4cCI6MTcwMDAwMDAwMCwiaWF0IjoxNjk5OTk5MDAwLCJuYmYiOjE2OTk5OTkwMDAsImlzcyI6InR0LXN0b2NrLWFwaSIsInN1YiI6InVzZXItMTIzIn0.wrong_signature",
			expectError: true,
			errorMsg:    "invalid token",
		},
		{
			name:        "Expired token",
			token:       expiredTokenString,
			expectError: true,
			errorMsg:    "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := svc.ParseToken(tt.token)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, testUserID, claims.UserID)
				assert.Equal(t, "0812345678", claims.PhoneNumber)
				assert.Equal(t, "access", claims.TokenType)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	svc, _, mockBlacklistRepo := setupTestService()

	testUserID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	// Generate a valid token for testing
	validToken, _ := svc.GenerateAccessToken(testUserID, "0812345678")

	tests := []struct {
		name        string
		token       string
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name:  "Valid non-blacklisted token",
			token: validToken,
			setupMocks: func() {
				mockBlacklistRepo.On("IsTokenBlacklisted", validToken).Return(false, nil).Once()
			},
			expectError: false,
		},
		{
			name:  "Blacklisted token",
			token: validToken,
			setupMocks: func() {
				mockBlacklistRepo.On("IsTokenBlacklisted", validToken).Return(true, nil).Once()
			},
			expectError: true,
			errorMsg:    "token has been invalidated",
		},
		{
			name:  "Blacklist check fails",
			token: validToken,
			setupMocks: func() {
				mockBlacklistRepo.On("IsTokenBlacklisted", validToken).Return(false, errors.New("db error")).Once()
			},
			expectError: true,
			errorMsg:    "failed to check token blacklist status",
		},
		{
			name:        "Invalid token format",
			token:       "invalid.token",
			setupMocks: func() {
				mockBlacklistRepo.On("IsTokenBlacklisted", "invalid.token").Return(false, nil).Once()
			},
			expectError: true,
			errorMsg:    "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockBlacklistRepo.ExpectedCalls = nil
			
			// Setup mocks for this test
			tt.setupMocks()
			
			// Execute test
			claims, err := svc.ValidateToken(tt.token)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, testUserID, claims.UserID)
			}
			
			// Verify all expectations were met
			mockBlacklistRepo.AssertExpectations(t)
		})
	}
}

func TestBlacklistToken(t *testing.T) {
	svc, _, mockBlacklistRepo := setupTestService()

	testUserID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	// Generate a valid token for testing
	validToken, _ := svc.GenerateAccessToken(testUserID, "0812345678")

	tests := []struct {
		name        string
		token       string
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name:  "Successfully blacklist token",
			token: validToken,
			setupMocks: func() {
				mockBlacklistRepo.On("BlacklistToken", validToken, testUserID.String(), "access", mock.AnythingOfType("time.Time")).Return(nil).Once()
			},
			expectError: false,
		},
		{
			name:        "Empty token",
			token:       "",
			setupMocks:  func() {},
			expectError: true,
			errorMsg:    "token is required",
		},
		{
			name:        "Invalid token format",
			token:       "invalid.token",
			setupMocks:  func() {},
			expectError: true,
			errorMsg:    "invalid token",
		},
		{
			name:  "Blacklist repository fails",
			token: validToken,
			setupMocks: func() {
				mockBlacklistRepo.On("BlacklistToken", validToken, testUserID.String(), "access", mock.AnythingOfType("time.Time")).Return(errors.New("db error")).Once()
			},
			expectError: true,
			errorMsg:    "failed to blacklist token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockBlacklistRepo.ExpectedCalls = nil
			
			// Setup mocks for this test
			tt.setupMocks()
			
			// Execute test
			err := svc.BlacklistToken(tt.token)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
			
			// Verify all expectations were met
			mockBlacklistRepo.AssertExpectations(t)
		})
	}
}

func TestIsTokenBlacklisted(t *testing.T) {
	svc, _, mockBlacklistRepo := setupTestService()

	tests := []struct {
		name           string
		token          string
		setupMocks     func()
		expectError    bool
		errorMsg       string
		expectedResult bool
	}{
		{
			name:  "Token is blacklisted",
			token: "some-token",
			setupMocks: func() {
				mockBlacklistRepo.On("IsTokenBlacklisted", "some-token").Return(true, nil).Once()
			},
			expectError:    false,
			expectedResult: true,
		},
		{
			name:  "Token is not blacklisted",
			token: "some-token",
			setupMocks: func() {
				mockBlacklistRepo.On("IsTokenBlacklisted", "some-token").Return(false, nil).Once()
			},
			expectError:    false,
			expectedResult: false,
		},
		{
			name:        "Empty token",
			token:       "",
			setupMocks:  func() {},
			expectError: true,
			errorMsg:    "token is required",
		},
		{
			name:  "Repository error",
			token: "some-token",
			setupMocks: func() {
				mockBlacklistRepo.On("IsTokenBlacklisted", "some-token").Return(false, errors.New("db error")).Once()
			},
			expectError: true,
			errorMsg:    "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockBlacklistRepo.ExpectedCalls = nil
			
			// Setup mocks for this test
			tt.setupMocks()
			
			// Execute test
			result, err := svc.IsTokenBlacklisted(tt.token)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
			
			// Verify all expectations were met
			mockBlacklistRepo.AssertExpectations(t)
		})
	}
}

// Integration test for the complete authentication flow
func TestAuthenticationFlow_Integration(t *testing.T) {
	svc, mockUserRepo, mockBlacklistRepo := setupTestService()

	testUserID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	// Create a test user with hashed PIN
	hashedPin, _ := utils.HashPin("123456")
	testUser := &user.User{
		ID:          testUserID,
		PhoneNumber: "0812345678",
		PinHash:     hashedPin,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	t.Run("Complete authentication and token lifecycle", func(t *testing.T) {
		// Setup mocks for authentication
		mockUserRepo.On("FindByPhoneNumber", "0812345678").Return(testUser, nil).Once()
		mockUserRepo.On("UpdateLastLogin", testUserID).Return(nil).Once()

		// 1. Authenticate user
		authenticatedUser, err := svc.AuthenticateUser("0812345678", "123456")
		assert.NoError(t, err)
		assert.Equal(t, testUser.ID, authenticatedUser.ID)

		// 2. Generate tokens
		tokenPair, err := svc.GenerateTokens(authenticatedUser.ID, authenticatedUser.PhoneNumber)
		assert.NoError(t, err)
		assert.NotEmpty(t, tokenPair.AccessToken)
		assert.NotEmpty(t, tokenPair.RefreshToken)

		// 3. Validate access token (not blacklisted)
		mockBlacklistRepo.On("IsTokenBlacklisted", tokenPair.AccessToken).Return(false, nil).Once()
		claims, err := svc.ValidateToken(tokenPair.AccessToken)
		assert.NoError(t, err)
		assert.Equal(t, authenticatedUser.ID, claims.UserID)
		assert.Equal(t, "access", claims.TokenType)

		// 4. Blacklist the token (logout)
		mockBlacklistRepo.On("BlacklistToken", tokenPair.AccessToken, testUserID.String(), "access", mock.AnythingOfType("time.Time")).Return(nil).Once()
		err = svc.BlacklistToken(tokenPair.AccessToken)
		assert.NoError(t, err)

		// 5. Try to validate blacklisted token
		mockBlacklistRepo.On("IsTokenBlacklisted", tokenPair.AccessToken).Return(true, nil).Once()
		_, err = svc.ValidateToken(tokenPair.AccessToken)
		assert.Error(t, err)
		assert.Equal(t, "token has been invalidated", err.Error())

		// Verify all expectations were met
		mockUserRepo.AssertExpectations(t)
		mockBlacklistRepo.AssertExpectations(t)
	})
}