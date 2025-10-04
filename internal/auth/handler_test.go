package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"tt-stock-api/internal/user"
	"tt-stock-api/pkg/response"
)

// MockAuthService is a mock implementation of auth.Service
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) ValidatePhoneNumber(phoneNumber string) error {
	args := m.Called(phoneNumber)
	return args.Error(0)
}

func (m *MockAuthService) ValidatePin(pin string) error {
	args := m.Called(pin)
	return args.Error(0)
}

func (m *MockAuthService) AuthenticateUser(phoneNumber, pin string) (*user.User, error) {
	args := m.Called(phoneNumber, pin)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockAuthService) GenerateAccessToken(userID uuid.UUID, phoneNumber string) (string, error) {
	args := m.Called(userID, phoneNumber)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) GenerateRefreshToken(userID uuid.UUID, phoneNumber string) (string, error) {
	args := m.Called(userID, phoneNumber)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) GenerateTokens(userID uuid.UUID, phoneNumber string) (*TokenPair, error) {
	args := m.Called(userID, phoneNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TokenPair), args.Error(1)
}

func (m *MockAuthService) ValidateToken(tokenString string) (*Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Claims), args.Error(1)
}

func (m *MockAuthService) ParseToken(tokenString string) (*Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Claims), args.Error(1)
}

func (m *MockAuthService) BlacklistToken(tokenString string) error {
	args := m.Called(tokenString)
	return args.Error(0)
}

func (m *MockAuthService) IsTokenBlacklisted(tokenString string) (bool, error) {
	args := m.Called(tokenString)
	return args.Bool(0), args.Error(1)
}

// Test setup helper
func setupTestHandler() (*handler, *MockAuthService, *fiber.App) {
	mockAuthService := &MockAuthService{}
	h := &handler{
		authService: mockAuthService,
	}
	
	app := fiber.New()
	
	return h, mockAuthService, app
}

// Helper function to create test user
func createTestUser() *user.User {
	return &user.User{
		ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		PhoneNumber: "0812345678",
		PinHash:     "hashed_pin",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Helper function to create test token pair
func createTestTokenPair() *TokenPair {
	return &TokenPair{
		AccessToken:  "test.access.token",
		RefreshToken: "test.refresh.token",
		ExpiresIn:    900, // 15 minutes
	}
}

// Helper function to create test claims
func createTestClaims(tokenType string) *Claims {
	return &Claims{
		UserID:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		PhoneNumber: "0812345678",
		TokenType:   tokenType,
	}
}

func TestLogin_Success(t *testing.T) {
	h, mockAuthService, app := setupTestHandler()
	
	testUser := createTestUser()
	testTokens := createTestTokenPair()
	
	// Setup route
	app.Post("/auth/login", h.Login)
	
	// Setup mocks
	mockAuthService.On("AuthenticateUser", "0812345678", "123456").Return(testUser, nil).Once()
	mockAuthService.On("GenerateTokens", testUser.ID, testUser.PhoneNumber).Return(testTokens, nil).Once()
	
	// Create request body
	loginReq := LoginRequest{
		PhoneNumber: "0812345678",
		Pin:         "123456",
	}
	reqBody, _ := json.Marshal(loginReq)
	
	// Create request
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Verify response
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	
	// Parse response body
	body, _ := io.ReadAll(resp.Body)
	var loginResp response.LoginResponse
	err = json.Unmarshal(body, &loginResp)
	assert.NoError(t, err)
	
	// Verify response structure
	assert.True(t, loginResp.Success)
	assert.Equal(t, testTokens.AccessToken, loginResp.Data.AccessToken)
	assert.Equal(t, testTokens.RefreshToken, loginResp.Data.RefreshToken)
	assert.Equal(t, testTokens.ExpiresIn, loginResp.Data.ExpiresIn)
	assert.Equal(t, testUser.ID.String(), loginResp.Data.User.ID)
	assert.Equal(t, testUser.PhoneNumber, loginResp.Data.User.PhoneNumber)
	
	// Verify all expectations were met
	mockAuthService.AssertExpectations(t)
}

func TestLogin_InvalidRequestBody(t *testing.T) {
	h, mockAuthService, app := setupTestHandler()
	
	// Setup route
	app.Post("/auth/login", h.Login)
	
	// Create invalid request body
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Verify response
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	
	// Parse response body
	body, _ := io.ReadAll(resp.Body)
	var errorResp response.ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	assert.NoError(t, err)
	
	// Verify error response
	assert.False(t, errorResp.Success)
	assert.Equal(t, "VALIDATION_ERROR", errorResp.Error.Code)
	assert.Equal(t, "Invalid request body", errorResp.Error.Message)
	
	// Verify no service calls were made
	mockAuthService.AssertExpectations(t)
}

func TestLogin_MissingPhoneNumber(t *testing.T) {
	h, mockAuthService, app := setupTestHandler()
	
	// Setup route
	app.Post("/auth/login", h.Login)
	
	// Create request with missing phone number
	loginReq := LoginRequest{
		Pin: "123456",
	}
	reqBody, _ := json.Marshal(loginReq)
	
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Verify response
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	
	// Parse response body
	body, _ := io.ReadAll(resp.Body)
	var errorResp response.ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	assert.NoError(t, err)
	
	// Verify error response
	assert.False(t, errorResp.Success)
	assert.Equal(t, "VALIDATION_ERROR", errorResp.Error.Code)
	assert.Equal(t, "Phone number is required", errorResp.Error.Message)
	
	// Verify no service calls were made
	mockAuthService.AssertExpectations(t)
}

func TestLogin_MissingPin(t *testing.T) {
	h, mockAuthService, app := setupTestHandler()
	
	// Setup route
	app.Post("/auth/login", h.Login)
	
	// Create request with missing PIN
	loginReq := LoginRequest{
		PhoneNumber: "0812345678",
	}
	reqBody, _ := json.Marshal(loginReq)
	
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Verify response
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	
	// Parse response body
	body, _ := io.ReadAll(resp.Body)
	var errorResp response.ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	assert.NoError(t, err)
	
	// Verify error response
	assert.False(t, errorResp.Success)
	assert.Equal(t, "VALIDATION_ERROR", errorResp.Error.Code)
	assert.Equal(t, "PIN is required", errorResp.Error.Message)
	
	// Verify no service calls were made
	mockAuthService.AssertExpectations(t)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	h, mockAuthService, app := setupTestHandler()
	
	// Setup route
	app.Post("/auth/login", h.Login)
	
	// Setup mocks - authentication fails
	mockAuthService.On("AuthenticateUser", "0812345678", "123456").Return(nil, errors.New("invalid credentials")).Once()
	
	// Create request body
	loginReq := LoginRequest{
		PhoneNumber: "0812345678",
		Pin:         "123456",
	}
	reqBody, _ := json.Marshal(loginReq)
	
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Verify response
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	
	// Parse response body
	body, _ := io.ReadAll(resp.Body)
	var errorResp response.ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	assert.NoError(t, err)
	
	// Verify error response
	assert.False(t, errorResp.Success)
	assert.Equal(t, "AUTHENTICATION_ERROR", errorResp.Error.Code)
	assert.Equal(t, "invalid credentials", errorResp.Error.Message)
	
	// Verify all expectations were met
	mockAuthService.AssertExpectations(t)
}

func TestLogin_TokenGenerationFails(t *testing.T) {
	h, mockAuthService, app := setupTestHandler()
	
	testUser := createTestUser()
	
	// Setup route
	app.Post("/auth/login", h.Login)
	
	// Setup mocks - authentication succeeds but token generation fails
	mockAuthService.On("AuthenticateUser", "0812345678", "123456").Return(testUser, nil).Once()
	mockAuthService.On("GenerateTokens", testUser.ID, testUser.PhoneNumber).Return(nil, errors.New("token generation failed")).Once()
	
	// Create request body
	loginReq := LoginRequest{
		PhoneNumber: "0812345678",
		Pin:         "123456",
	}
	reqBody, _ := json.Marshal(loginReq)
	
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Verify response
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	
	// Parse response body
	body, _ := io.ReadAll(resp.Body)
	var errorResp response.ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	assert.NoError(t, err)
	
	// Verify error response
	assert.False(t, errorResp.Success)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", errorResp.Error.Code)
	assert.Equal(t, "Failed to generate authentication tokens", errorResp.Error.Message)
	
	// Verify all expectations were met
	mockAuthService.AssertExpectations(t)
}

func TestRefresh_Success(t *testing.T) {
	h, mockAuthService, app := setupTestHandler()
	
	testClaims := createTestClaims("refresh")
	testTokens := createTestTokenPair()
	
	// Setup route
	app.Post("/auth/refresh", h.Refresh)
	
	// Setup mocks
	mockAuthService.On("ValidateToken", "test.refresh.token").Return(testClaims, nil).Once()
	mockAuthService.On("BlacklistToken", "test.refresh.token").Return(nil).Once()
	mockAuthService.On("GenerateTokens", testClaims.UserID, testClaims.PhoneNumber).Return(testTokens, nil).Once()
	
	// Create request body
	refreshReq := RefreshRequest{
		RefreshToken: "test.refresh.token",
	}
	reqBody, _ := json.Marshal(refreshReq)
	
	req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Verify response
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	
	// Parse response body
	body, _ := io.ReadAll(resp.Body)
	var loginResp response.LoginResponse
	err = json.Unmarshal(body, &loginResp)
	assert.NoError(t, err)
	
	// Verify response structure
	assert.True(t, loginResp.Success)
	assert.Equal(t, testTokens.AccessToken, loginResp.Data.AccessToken)
	assert.Equal(t, testTokens.RefreshToken, loginResp.Data.RefreshToken)
	assert.Equal(t, testTokens.ExpiresIn, loginResp.Data.ExpiresIn)
	assert.Equal(t, testClaims.UserID.String(), loginResp.Data.User.ID)
	assert.Equal(t, testClaims.PhoneNumber, loginResp.Data.User.PhoneNumber)
	
	// Verify all expectations were met
	mockAuthService.AssertExpectations(t)
}

func TestRefresh_InvalidRequestBody(t *testing.T) {
	h, mockAuthService, app := setupTestHandler()
	
	// Setup route
	app.Post("/auth/refresh", h.Refresh)
	
	// Create invalid request body
	req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Verify response
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	
	// Parse response body
	body, _ := io.ReadAll(resp.Body)
	var errorResp response.ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	assert.NoError(t, err)
	
	// Verify error response
	assert.False(t, errorResp.Success)
	assert.Equal(t, "VALIDATION_ERROR", errorResp.Error.Code)
	assert.Equal(t, "Invalid request body", errorResp.Error.Message)
	
	// Verify no service calls were made
	mockAuthService.AssertExpectations(t)
}

func TestRefresh_MissingRefreshToken(t *testing.T) {
	h, mockAuthService, app := setupTestHandler()
	
	// Setup route
	app.Post("/auth/refresh", h.Refresh)
	
	// Create request with missing refresh token
	refreshReq := RefreshRequest{}
	reqBody, _ := json.Marshal(refreshReq)
	
	req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Verify response
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	
	// Parse response body
	body, _ := io.ReadAll(resp.Body)
	var errorResp response.ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	assert.NoError(t, err)
	
	// Verify error response
	assert.False(t, errorResp.Success)
	assert.Equal(t, "VALIDATION_ERROR", errorResp.Error.Code)
	assert.Equal(t, "Refresh token is required", errorResp.Error.Message)
	
	// Verify no service calls were made
	mockAuthService.AssertExpectations(t)
}

func TestRefresh_InvalidRefreshToken(t *testing.T) {
	h, mockAuthService, app := setupTestHandler()
	
	// Setup route
	app.Post("/auth/refresh", h.Refresh)
	
	// Setup mocks - token validation fails
	mockAuthService.On("ValidateToken", "invalid.refresh.token").Return(nil, errors.New("invalid token")).Once()
	
	// Create request body
	refreshReq := RefreshRequest{
		RefreshToken: "invalid.refresh.token",
	}
	reqBody, _ := json.Marshal(refreshReq)
	
	req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Verify response
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	
	// Parse response body
	body, _ := io.ReadAll(resp.Body)
	var errorResp response.ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	assert.NoError(t, err)
	
	// Verify error response
	assert.False(t, errorResp.Success)
	assert.Equal(t, "AUTHENTICATION_ERROR", errorResp.Error.Code)
	assert.Equal(t, "Invalid or expired refresh token", errorResp.Error.Message)
	
	// Verify all expectations were met
	mockAuthService.AssertExpectations(t)
}

func TestRefresh_AccessTokenInsteadOfRefreshToken(t *testing.T) {
	h, mockAuthService, app := setupTestHandler()
	
	testClaims := createTestClaims("access") // Wrong token type
	
	// Setup route
	app.Post("/auth/refresh", h.Refresh)
	
	// Setup mocks - token validation succeeds but wrong type
	mockAuthService.On("ValidateToken", "test.access.token").Return(testClaims, nil).Once()
	
	// Create request body
	refreshReq := RefreshRequest{
		RefreshToken: "test.access.token",
	}
	reqBody, _ := json.Marshal(refreshReq)
	
	req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Verify response
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	
	// Parse response body
	body, _ := io.ReadAll(resp.Body)
	var errorResp response.ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	assert.NoError(t, err)
	
	// Verify error response
	assert.False(t, errorResp.Success)
	assert.Equal(t, "AUTHENTICATION_ERROR", errorResp.Error.Code)
	assert.Equal(t, "Invalid token type", errorResp.Error.Message)
	
	// Verify all expectations were met
	mockAuthService.AssertExpectations(t)
}

func TestRefresh_BlacklistFails(t *testing.T) {
	h, mockAuthService, app := setupTestHandler()
	
	testClaims := createTestClaims("refresh")
	
	// Setup route
	app.Post("/auth/refresh", h.Refresh)
	
	// Setup mocks - blacklisting fails
	mockAuthService.On("ValidateToken", "test.refresh.token").Return(testClaims, nil).Once()
	mockAuthService.On("BlacklistToken", "test.refresh.token").Return(errors.New("blacklist failed")).Once()
	
	// Create request body
	refreshReq := RefreshRequest{
		RefreshToken: "test.refresh.token",
	}
	reqBody, _ := json.Marshal(refreshReq)
	
	req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Verify response
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	
	// Parse response body
	body, _ := io.ReadAll(resp.Body)
	var errorResp response.ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	assert.NoError(t, err)
	
	// Verify error response
	assert.False(t, errorResp.Success)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", errorResp.Error.Code)
	assert.Equal(t, "Failed to invalidate old refresh token", errorResp.Error.Message)
	
	// Verify all expectations were met
	mockAuthService.AssertExpectations(t)
}

func TestRefresh_TokenGenerationFails(t *testing.T) {
	h, mockAuthService, app := setupTestHandler()
	
	testClaims := createTestClaims("refresh")
	
	// Setup route
	app.Post("/auth/refresh", h.Refresh)
	
	// Setup mocks - token generation fails
	mockAuthService.On("ValidateToken", "test.refresh.token").Return(testClaims, nil).Once()
	mockAuthService.On("BlacklistToken", "test.refresh.token").Return(nil).Once()
	mockAuthService.On("GenerateTokens", testClaims.UserID, testClaims.PhoneNumber).Return(nil, errors.New("token generation failed")).Once()
	
	// Create request body
	refreshReq := RefreshRequest{
		RefreshToken: "test.refresh.token",
	}
	reqBody, _ := json.Marshal(refreshReq)
	
	req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Verify response
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	
	// Parse response body
	body, _ := io.ReadAll(resp.Body)
	var errorResp response.ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	assert.NoError(t, err)
	
	// Verify error response
	assert.False(t, errorResp.Success)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", errorResp.Error.Code)
	assert.Equal(t, "Failed to generate new authentication tokens", errorResp.Error.Message)
	
	// Verify all expectations were met
	mockAuthService.AssertExpectations(t)
}

func TestLogout_Success(t *testing.T) {
	h, mockAuthService, app := setupTestHandler()
	
	testClaims := createTestClaims("access")
	
	// Setup route
	app.Post("/auth/logout", h.Logout)
	
	// Setup mocks
	mockAuthService.On("ValidateToken", "test.access.token").Return(testClaims, nil).Once()
	mockAuthService.On("BlacklistToken", "test.access.token").Return(nil).Once()
	mockAuthService.On("BlacklistToken", "test.refresh.token").Return(nil).Once()
	
	// Create request body with refresh token
	refreshReq := RefreshRequest{
		RefreshToken: "test.refresh.token",
	}
	reqBody, _ := json.Marshal(refreshReq)
	
	req := httptest.NewRequest("POST", "/auth/logout", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test.access.token")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Verify response
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	
	// Parse response body
	body, _ := io.ReadAll(resp.Body)
	var successResp response.SuccessResponse
	err = json.Unmarshal(body, &successResp)
	assert.NoError(t, err)
	
	// Verify response structure
	assert.True(t, successResp.Success)
	assert.Equal(t, "Logout successful", successResp.Message)
	
	// Verify all expectations were met
	mockAuthService.AssertExpectations(t)
}

func TestLogout_MissingAuthorizationHeader(t *testing.T) {
	h, mockAuthService, app := setupTestHandler()
	
	// Setup route
	app.Post("/auth/logout", h.Logout)
	
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Verify response
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	
	// Parse response body
	body, _ := io.ReadAll(resp.Body)
	var errorResp response.ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	assert.NoError(t, err)
	
	// Verify error response
	assert.False(t, errorResp.Success)
	assert.Equal(t, "VALIDATION_ERROR", errorResp.Error.Code)
	assert.Equal(t, "Authorization header is required", errorResp.Error.Message)
	
	// Verify no service calls were made
	mockAuthService.AssertExpectations(t)
}

func TestLogout_InvalidAuthorizationHeaderFormat(t *testing.T) {
	h, mockAuthService, app := setupTestHandler()
	
	// Setup route
	app.Post("/auth/logout", h.Logout)
	
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	req.Header.Set("Authorization", "Invalid format")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Verify response
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	
	// Parse response body
	body, _ := io.ReadAll(resp.Body)
	var errorResp response.ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	assert.NoError(t, err)
	
	// Verify error response
	assert.False(t, errorResp.Success)
	assert.Equal(t, "VALIDATION_ERROR", errorResp.Error.Code)
	assert.Equal(t, "Invalid authorization header format", errorResp.Error.Message)
	
	// Verify no service calls were made
	mockAuthService.AssertExpectations(t)
}

func TestLogout_EmptyAccessToken(t *testing.T) {
	h, mockAuthService, app := setupTestHandler()
	
	// Setup route
	app.Post("/auth/logout", h.Logout)
	
	// Test with "Bearer " - Fiber trims trailing spaces, so this becomes "Bearer" 
	// which fails the HasPrefix check for "Bearer " (with space)
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer ")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Verify response
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	
	// Parse response body
	body, _ := io.ReadAll(resp.Body)
	var errorResp response.ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	assert.NoError(t, err)
	
	// Verify error response - Fiber trims "Bearer " to "Bearer", so HasPrefix fails
	assert.False(t, errorResp.Success)
	assert.Equal(t, "VALIDATION_ERROR", errorResp.Error.Code)
	assert.Equal(t, "Invalid authorization header format", errorResp.Error.Message)
	
	// Verify no service calls were made
	mockAuthService.AssertExpectations(t)
}


func TestLogout_InvalidAccessToken(t *testing.T) {
	h, mockAuthService, app := setupTestHandler()
	
	// Setup route
	app.Post("/auth/logout", h.Logout)
	
	// Setup mocks - token validation fails
	mockAuthService.On("ValidateToken", "invalid.access.token").Return(nil, errors.New("invalid token")).Once()
	
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer invalid.access.token")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Verify response
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	
	// Parse response body
	body, _ := io.ReadAll(resp.Body)
	var errorResp response.ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	assert.NoError(t, err)
	
	// Verify error response
	assert.False(t, errorResp.Success)
	assert.Equal(t, "AUTHENTICATION_ERROR", errorResp.Error.Code)
	assert.Equal(t, "Invalid or expired access token", errorResp.Error.Message)
	
	// Verify all expectations were met
	mockAuthService.AssertExpectations(t)
}

func TestLogout_RefreshTokenInsteadOfAccessToken(t *testing.T) {
	h, mockAuthService, app := setupTestHandler()
	
	testClaims := createTestClaims("refresh") // Wrong token type
	
	// Setup route
	app.Post("/auth/logout", h.Logout)
	
	// Setup mocks - token validation succeeds but wrong type
	mockAuthService.On("ValidateToken", "test.refresh.token").Return(testClaims, nil).Once()
	
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer test.refresh.token")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Verify response
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	
	// Parse response body
	body, _ := io.ReadAll(resp.Body)
	var errorResp response.ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	assert.NoError(t, err)
	
	// Verify error response
	assert.False(t, errorResp.Success)
	assert.Equal(t, "AUTHENTICATION_ERROR", errorResp.Error.Code)
	assert.Equal(t, "Invalid token type", errorResp.Error.Message)
	
	// Verify all expectations were met
	mockAuthService.AssertExpectations(t)
}

func TestLogout_AccessTokenBlacklistFails(t *testing.T) {
	h, mockAuthService, app := setupTestHandler()
	
	testClaims := createTestClaims("access")
	
	// Setup route
	app.Post("/auth/logout", h.Logout)
	
	// Setup mocks - access token blacklisting fails
	mockAuthService.On("ValidateToken", "test.access.token").Return(testClaims, nil).Once()
	mockAuthService.On("BlacklistToken", "test.access.token").Return(errors.New("blacklist failed")).Once()
	
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer test.access.token")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Verify response
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	
	// Parse response body
	body, _ := io.ReadAll(resp.Body)
	var errorResp response.ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	assert.NoError(t, err)
	
	// Verify error response
	assert.False(t, errorResp.Success)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", errorResp.Error.Code)
	assert.Equal(t, "Failed to invalidate access token", errorResp.Error.Message)
	
	// Verify all expectations were met
	mockAuthService.AssertExpectations(t)
}

func TestLogout_WithoutRefreshToken(t *testing.T) {
	h, mockAuthService, app := setupTestHandler()
	
	testClaims := createTestClaims("access")
	
	// Setup route
	app.Post("/auth/logout", h.Logout)
	
	// Setup mocks - only access token blacklisting
	mockAuthService.On("ValidateToken", "test.access.token").Return(testClaims, nil).Once()
	mockAuthService.On("BlacklistToken", "test.access.token").Return(nil).Once()
	
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer test.access.token")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Verify response
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	
	// Parse response body
	body, _ := io.ReadAll(resp.Body)
	var successResp response.SuccessResponse
	err = json.Unmarshal(body, &successResp)
	assert.NoError(t, err)
	
	// Verify response structure
	assert.True(t, successResp.Success)
	assert.Equal(t, "Logout successful", successResp.Message)
	
	// Verify all expectations were met
	mockAuthService.AssertExpectations(t)
}

// Integration test for complete authentication flow via HTTP handlers
func TestAuthenticationHandlers_Integration(t *testing.T) {
	h, mockAuthService, app := setupTestHandler()
	
	testUser := createTestUser()
	testTokens := createTestTokenPair()
	testRefreshClaims := createTestClaims("refresh")
	
	// Setup routes
	app.Post("/auth/login", h.Login)
	app.Post("/auth/refresh", h.Refresh)
	app.Post("/auth/logout", h.Logout)
	
	t.Run("Complete authentication flow", func(t *testing.T) {
		// 1. Login
		mockAuthService.On("AuthenticateUser", "0812345678", "123456").Return(testUser, nil).Once()
		mockAuthService.On("GenerateTokens", testUser.ID, testUser.PhoneNumber).Return(testTokens, nil).Once()
		
		loginReq := LoginRequest{
			PhoneNumber: "0812345678",
			Pin:         "123456",
		}
		loginReqBody, _ := json.Marshal(loginReq)
		
		loginHttpReq := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(loginReqBody))
		loginHttpReq.Header.Set("Content-Type", "application/json")
		
		loginResp, err := app.Test(loginHttpReq)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, loginResp.StatusCode)
		
		// Parse login response
		loginBody, _ := io.ReadAll(loginResp.Body)
		var loginResponse response.LoginResponse
		err = json.Unmarshal(loginBody, &loginResponse)
		assert.NoError(t, err)
		assert.True(t, loginResponse.Success)
		
		// 2. Refresh tokens
		mockAuthService.On("ValidateToken", testTokens.RefreshToken).Return(testRefreshClaims, nil).Once()
		mockAuthService.On("BlacklistToken", testTokens.RefreshToken).Return(nil).Once()
		
		newTokens := &TokenPair{
			AccessToken:  "new.access.token",
			RefreshToken: "new.refresh.token",
			ExpiresIn:    900,
		}
		mockAuthService.On("GenerateTokens", testRefreshClaims.UserID, testRefreshClaims.PhoneNumber).Return(newTokens, nil).Once()
		
		refreshReq := RefreshRequest{
			RefreshToken: testTokens.RefreshToken,
		}
		refreshReqBody, _ := json.Marshal(refreshReq)
		
		refreshHttpReq := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(refreshReqBody))
		refreshHttpReq.Header.Set("Content-Type", "application/json")
		
		refreshResp, err := app.Test(refreshHttpReq)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, refreshResp.StatusCode)
		
		// Parse refresh response
		refreshBody, _ := io.ReadAll(refreshResp.Body)
		var refreshResponse response.LoginResponse
		err = json.Unmarshal(refreshBody, &refreshResponse)
		assert.NoError(t, err)
		assert.True(t, refreshResponse.Success)
		assert.Equal(t, newTokens.AccessToken, refreshResponse.Data.AccessToken)
		
		// 3. Logout
		newAccessClaims := &Claims{
			UserID:      testUser.ID,
			PhoneNumber: testUser.PhoneNumber,
			TokenType:   "access",
		}
		mockAuthService.On("ValidateToken", newTokens.AccessToken).Return(newAccessClaims, nil).Once()
		mockAuthService.On("BlacklistToken", newTokens.AccessToken).Return(nil).Once()
		mockAuthService.On("BlacklistToken", newTokens.RefreshToken).Return(nil).Once()
		
		logoutReq := RefreshRequest{
			RefreshToken: newTokens.RefreshToken,
		}
		logoutReqBody, _ := json.Marshal(logoutReq)
		
		logoutHttpReq := httptest.NewRequest("POST", "/auth/logout", bytes.NewReader(logoutReqBody))
		logoutHttpReq.Header.Set("Content-Type", "application/json")
		logoutHttpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", newTokens.AccessToken))
		
		logoutResp, err := app.Test(logoutHttpReq)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, logoutResp.StatusCode)
		
		// Parse logout response
		logoutBody, _ := io.ReadAll(logoutResp.Body)
		var logoutResponse response.SuccessResponse
		err = json.Unmarshal(logoutBody, &logoutResponse)
		assert.NoError(t, err)
		assert.True(t, logoutResponse.Success)
		assert.Equal(t, "Logout successful", logoutResponse.Message)
		
		// Verify all expectations were met
		mockAuthService.AssertExpectations(t)
	})
}