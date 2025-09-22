package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"tt-stock-api/src/models"
	"tt-stock-api/src/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error) {
	args := m.Called(ctx, phoneNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, req *models.UserListRequest) (*models.UserListResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserListResponse), args.Error(1)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) GetActiveUsers(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) GetUsersByRole(ctx context.Context, role models.UserRole) ([]models.User, error) {
	args := m.Called(ctx, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) Exists(ctx context.Context, phoneNumber string) (bool, error) {
	args := m.Called(ctx, phoneNumber)
	return args.Bool(0), args.Error(1)
}

// MockSessionRepository is a mock implementation of SessionRepository
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session *models.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByID(ctx context.Context, id uint) (*models.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionRepository) GetByToken(ctx context.Context, token string) (*models.Session, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionRepository) Update(ctx context.Context, session *models.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) List(ctx context.Context, req *models.SessionListRequest) (*models.SessionListResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SessionListResponse), args.Error(1)
}

func (m *MockSessionRepository) RevokeSession(ctx context.Context, sessionID uint) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSessionRepository) RevokeAllUserSessions(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockSessionRepository) CleanupExpiredSessions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSessionRepository) GetSessionsByUser(ctx context.Context, userID uint) ([]models.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Session), args.Error(1)
}

func (m *MockSessionRepository) GetActiveSessions(ctx context.Context) ([]models.Session, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Session), args.Error(1)
}

func (m *MockSessionRepository) GetExpiredSessions(ctx context.Context) ([]models.Session, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Session), args.Error(1)
}

func (m *MockSessionRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSessionRepository) CountActiveSessions(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		name          string
		request       *models.UserLoginRequest
		setupMocks    func(*MockUserRepository, *MockSessionRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name: "successful login",
			request: &models.UserLoginRequest{
				PhoneNumber: "1234567890",
				PIN:         "1234",
			},
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository) {
				// Hash the PIN "1234" for testing
				hashedPIN, _ := bcrypt.GenerateFromPassword([]byte("1234"), bcrypt.DefaultCost)
				user := &models.User{
					ID:          1,
					PhoneNumber: "1234567890",
					PIN:         string(hashedPIN),
					Role:        models.UserRoleStaff,
					Name:        "Test User",
					IsActive:    true,
				}
				userRepo.On("GetByPhoneNumber", mock.Anything, "1234567890").Return(user, nil)
				userRepo.On("UpdateLastLogin", mock.Anything, uint(1)).Return(nil)
				sessionRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Session")).Return(nil)
			},
			expectSuccess: true,
		},
		{
			name:    "nil request",
			request: nil,
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository) {
				// No mocks needed
			},
			expectedError: "login request cannot be nil",
		},
		{
			name: "empty phone number",
			request: &models.UserLoginRequest{
				PhoneNumber: "",
				PIN:         "1234",
			},
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository) {
				// No mocks needed
			},
			expectedError: "phone number is required",
		},
		{
			name: "empty PIN",
			request: &models.UserLoginRequest{
				PhoneNumber: "1234567890",
				PIN:         "",
			},
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository) {
				// No mocks needed
			},
			expectedError: "PIN is required",
		},
		{
			name: "user not found",
			request: &models.UserLoginRequest{
				PhoneNumber: "1234567890",
				PIN:         "1234",
			},
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository) {
				userRepo.On("GetByPhoneNumber", mock.Anything, "1234567890").Return(nil, errors.New("user not found"))
			},
			expectedError: "invalid credentials",
		},
		{
			name: "inactive user",
			request: &models.UserLoginRequest{
				PhoneNumber: "1234567890",
				PIN:         "1234",
			},
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository) {
				user := &models.User{
					ID:          1,
					PhoneNumber: "1234567890",
					PIN:         "hashedpin",
					Role:        models.UserRoleStaff,
					Name:        "Test User",
					IsActive:    false,
				}
				userRepo.On("GetByPhoneNumber", mock.Anything, "1234567890").Return(user, nil)
			},
			expectedError: "account is deactivated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			userRepo := &MockUserRepository{}
			sessionRepo := &MockSessionRepository{}

			// Setup mocks
			tt.setupMocks(userRepo, sessionRepo)

			// Create service
			authService := services.NewAuthService(userRepo, sessionRepo, "test-secret", time.Hour)

			// Execute
			result, err := authService.Login(context.Background(), tt.request)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.Token)
				assert.NotZero(t, result.ExpiresAt)
				assert.NotNil(t, result.User)
			}

			// Verify all expectations
			userRepo.AssertExpectations(t)
			sessionRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		setupMocks    func(*MockSessionRepository)
		expectedError string
	}{
		{
			name:  "successful logout",
			token: "valid-token",
			setupMocks: func(sessionRepo *MockSessionRepository) {
				session := &models.Session{
					ID:     1,
					UserID: 1,
					Token:  "valid-token",
					Status: models.SessionStatusActive,
				}
				sessionRepo.On("GetByToken", mock.Anything, "valid-token").Return(session, nil)
				sessionRepo.On("RevokeSession", mock.Anything, uint(1)).Return(nil)
			},
			expectedError: "",
		},
		{
			name:  "empty token",
			token: "",
			setupMocks: func(sessionRepo *MockSessionRepository) {
				// No mocks needed
			},
			expectedError: "token is required",
		},
		{
			name:  "session not found",
			token: "invalid-token",
			setupMocks: func(sessionRepo *MockSessionRepository) {
				sessionRepo.On("GetByToken", mock.Anything, "invalid-token").Return(nil, errors.New("session not found"))
			},
			expectedError: "invalid session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			userRepo := &MockUserRepository{}
			sessionRepo := &MockSessionRepository{}

			// Setup mocks
			tt.setupMocks(sessionRepo)

			// Create service
			authService := services.NewAuthService(userRepo, sessionRepo, "test-secret", time.Hour)

			// Execute
			err := authService.Logout(context.Background(), tt.token)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations
			sessionRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_RevokeAllUserSessions(t *testing.T) {
	tests := []struct {
		name          string
		userID        uint
		setupMocks    func(*MockSessionRepository)
		expectedError string
	}{
		{
			name:   "successful revocation",
			userID: 1,
			setupMocks: func(sessionRepo *MockSessionRepository) {
				sessionRepo.On("RevokeAllUserSessions", mock.Anything, uint(1)).Return(nil)
			},
			expectedError: "",
		},
		{
			name:   "zero user ID",
			userID: 0,
			setupMocks: func(sessionRepo *MockSessionRepository) {
				// No mocks needed
			},
			expectedError: "user ID cannot be zero",
		},
		{
			name:   "database error",
			userID: 1,
			setupMocks: func(sessionRepo *MockSessionRepository) {
				sessionRepo.On("RevokeAllUserSessions", mock.Anything, uint(1)).Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			userRepo := &MockUserRepository{}
			sessionRepo := &MockSessionRepository{}

			// Setup mocks
			tt.setupMocks(sessionRepo)

			// Create service
			authService := services.NewAuthService(userRepo, sessionRepo, "test-secret", time.Hour)

			// Execute
			err := authService.RevokeAllUserSessions(context.Background(), tt.userID)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations
			sessionRepo.AssertExpectations(t)
		})
	}
}
