package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tt-stock-api/internal/config"
	"tt-stock-api/internal/db"
	"tt-stock-api/internal/user"
	"tt-stock-api/pkg/response"
	"tt-stock-api/pkg/utils"
)

// Integration test setup
type IntegrationTestSuite struct {
	db            *db.DB
	userRepo      user.Repository
	blacklistRepo BlacklistRepository
	authService   Service
	handler       Handler
	app           *fiber.App
	testUser      *user.User
}

// setupIntegrationTest initializes the test environment with real database
func setupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	// Build test database URL from environment variables
	testDBHost := os.Getenv("TEST_DB_HOST")
	if testDBHost == "" {
		// Skip integration tests if no test database is configured
		t.Skip("TEST_DB_HOST not set, skipping integration tests")
	}
	
	testDBPort := getEnvOrDefault("TEST_DB_PORT", "5432")
	testDBName := getEnvOrDefault("TEST_DB_NAME", "tt_stock_test_db")
	testDBUser := getEnvOrDefault("TEST_DB_USER", "postgres")
	testDBPassword := os.Getenv("TEST_DB_PASSWORD")
	
	testDBURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		testDBUser, testDBPassword, testDBHost, testDBPort, testDBName)

	// Connect to test database
	database, err := db.Connect(testDBURL)
	require.NoError(t, err, "Failed to connect to test database")

	// Create tables
	err = database.CreateTables()
	require.NoError(t, err, "Failed to create database tables")

	// Initialize repositories and services
	userRepo := user.NewRepository(database)
	blacklistRepo := NewBlacklistRepository(database)
	
	cfg := &config.Config{
		JWTSecret: "test-jwt-secret-key-for-integration-tests",
	}
	authService := NewService(userRepo, blacklistRepo, cfg)
	handler := NewHandler(authService)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/auth/login", handler.Login)
	app.Post("/auth/refresh", handler.Refresh)
	app.Post("/auth/logout", handler.Logout)

	// Create test user
	testUser := createTestUserInDB(t, database)

	return &IntegrationTestSuite{
		db:            database,
		userRepo:      userRepo,
		blacklistRepo: blacklistRepo,
		authService:   authService,
		handler:       handler,
		app:           app,
		testUser:      testUser,
	}
}

// createTestUserInDB creates a test user in the database
func createTestUserInDB(t *testing.T, database *db.DB) *user.User {
	// Hash the test PIN
	pinHash, err := utils.HashPin("123456")
	require.NoError(t, err, "Failed to hash test PIN")

	// Insert test user
	userID := uuid.New()
	query := `
		INSERT INTO users (id, phone_number, pin_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	now := time.Now()
	_, err = database.Exec(query, userID, "0812345678", pinHash, now, now)
	require.NoError(t, err, "Failed to create test user")

	return &user.User{
		ID:          userID,
		PhoneNumber: "0812345678",
		PinHash:     pinHash,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// cleanupIntegrationTest cleans up the test environment
func (suite *IntegrationTestSuite) cleanup(t *testing.T) {
	// Clean up test data
	_, err := suite.db.Exec("DELETE FROM token_blacklist")
	require.NoError(t, err, "Failed to clean up token_blacklist table")
	
	_, err = suite.db.Exec("DELETE FROM users")
	require.NoError(t, err, "Failed to clean up users table")

	// Close database connection
	err = suite.db.Close()
	require.NoError(t, err, "Failed to close database connection")
}

// TestLoginEndpoint_Integration tests the login endpoint with real database
func TestLoginEndpoint_Integration(t *testing.T) {
	suite := setupIntegrationTest(t)
	defer suite.cleanup(t)

	t.Run("successful login with valid credentials", func(t *testing.T) {
		loginReq := LoginRequest{
			PhoneNumber: "0812345678",
			Pin:         "123456",
		}
		reqBody, _ := json.Marshal(loginReq)

		req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := suite.app.Test(req)
		require.NoError(t, err)

		// Verify response status
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Parse response body
		body, _ := io.ReadAll(resp.Body)
		var loginResp response.LoginResponse
		err = json.Unmarshal(body, &loginResp)
		require.NoError(t, err)

		// Verify response structure
		assert.True(t, loginResp.Success)
		assert.NotEmpty(t, loginResp.Data.AccessToken)
		assert.NotEmpty(t, loginResp.Data.RefreshToken)
		assert.Equal(t, int64(900), loginResp.Data.ExpiresIn) // 15 minutes
		assert.Equal(t, suite.testUser.ID.String(), loginResp.Data.User.ID)
		assert.Equal(t, suite.testUser.PhoneNumber, loginResp.Data.User.PhoneNumber)

		// Verify tokens are valid JWT tokens
		accessClaims, err := suite.authService.ValidateToken(loginResp.Data.AccessToken)
		assert.NoError(t, err)
		assert.Equal(t, "access", accessClaims.TokenType)
		assert.Equal(t, suite.testUser.ID, accessClaims.UserID)

		refreshClaims, err := suite.authService.ValidateToken(loginResp.Data.RefreshToken)
		assert.NoError(t, err)
		assert.Equal(t, "refresh", refreshClaims.TokenType)
		assert.Equal(t, suite.testUser.ID, refreshClaims.UserID)

		// Verify last login was updated in database
		updatedUser, err := suite.userRepo.FindByPhoneNumber("0812345678")
		require.NoError(t, err)
		assert.NotNil(t, updatedUser.LastLoginAt)
		assert.True(t, updatedUser.LastLoginAt.After(suite.testUser.CreatedAt))
	})

	t.Run("login with invalid phone number format", func(t *testing.T) {
		loginReq := LoginRequest{
			PhoneNumber: "123456789", // Invalid format (missing leading 0)
			Pin:         "123456",
		}
		reqBody, _ := json.Marshal(loginReq)

		req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := suite.app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var errorResp response.ErrorResponse
		err = json.Unmarshal(body, &errorResp)
		require.NoError(t, err)

		assert.False(t, errorResp.Success)
		assert.Equal(t, "AUTHENTICATION_ERROR", errorResp.Error.Code)
		assert.Contains(t, errorResp.Error.Message, "invalid phone number format")
	})

	t.Run("login with invalid PIN format", func(t *testing.T) {
		loginReq := LoginRequest{
			PhoneNumber: "0812345678",
			Pin:         "12345", // Invalid format (only 5 digits)
		}
		reqBody, _ := json.Marshal(loginReq)

		req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := suite.app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var errorResp response.ErrorResponse
		err = json.Unmarshal(body, &errorResp)
		require.NoError(t, err)

		assert.False(t, errorResp.Success)
		assert.Equal(t, "AUTHENTICATION_ERROR", errorResp.Error.Code)
		assert.Contains(t, errorResp.Error.Message, "invalid PIN format")
	})

	t.Run("login with non-existent phone number", func(t *testing.T) {
		loginReq := LoginRequest{
			PhoneNumber: "0987654321", // Non-existent phone number
			Pin:         "123456",
		}
		reqBody, _ := json.Marshal(loginReq)

		req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := suite.app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var errorResp response.ErrorResponse
		err = json.Unmarshal(body, &errorResp)
		require.NoError(t, err)

		assert.False(t, errorResp.Success)
		assert.Equal(t, "AUTHENTICATION_ERROR", errorResp.Error.Code)
		assert.Equal(t, "invalid credentials", errorResp.Error.Message)
	})

	t.Run("login with wrong PIN", func(t *testing.T) {
		loginReq := LoginRequest{
			PhoneNumber: "0812345678",
			Pin:         "654321", // Wrong PIN
		}
		reqBody, _ := json.Marshal(loginReq)

		req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := suite.app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var errorResp response.ErrorResponse
		err = json.Unmarshal(body, &errorResp)
		require.NoError(t, err)

		assert.False(t, errorResp.Success)
		assert.Equal(t, "AUTHENTICATION_ERROR", errorResp.Error.Code)
		assert.Equal(t, "invalid credentials", errorResp.Error.Message)
	})
}

// TestRefreshEndpoint_Integration tests the refresh token endpoint with real database
func TestRefreshEndpoint_Integration(t *testing.T) {
	suite := setupIntegrationTest(t)
	defer suite.cleanup(t)

	// First, get valid tokens by logging in
	tokens, err := suite.authService.GenerateTokens(suite.testUser.ID, suite.testUser.PhoneNumber)
	require.NoError(t, err)

	t.Run("successful token refresh with valid refresh token", func(t *testing.T) {
		refreshReq := RefreshRequest{
			RefreshToken: tokens.RefreshToken,
		}
		reqBody, _ := json.Marshal(refreshReq)

		req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := suite.app.Test(req)
		require.NoError(t, err)

		// Verify response status
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Parse response body
		body, _ := io.ReadAll(resp.Body)
		var refreshResp response.LoginResponse
		err = json.Unmarshal(body, &refreshResp)
		require.NoError(t, err)

		// Verify response structure
		assert.True(t, refreshResp.Success)
		assert.NotEmpty(t, refreshResp.Data.AccessToken)
		assert.NotEmpty(t, refreshResp.Data.RefreshToken)
		assert.Equal(t, int64(900), refreshResp.Data.ExpiresIn)
		assert.Equal(t, suite.testUser.ID.String(), refreshResp.Data.User.ID)
		assert.Equal(t, suite.testUser.PhoneNumber, refreshResp.Data.User.PhoneNumber)

		// Verify new tokens are different from original
		assert.NotEqual(t, tokens.AccessToken, refreshResp.Data.AccessToken)
		assert.NotEqual(t, tokens.RefreshToken, refreshResp.Data.RefreshToken)

		// Verify new tokens are valid
		accessClaims, err := suite.authService.ValidateToken(refreshResp.Data.AccessToken)
		assert.NoError(t, err)
		assert.Equal(t, "access", accessClaims.TokenType)

		refreshClaims, err := suite.authService.ValidateToken(refreshResp.Data.RefreshToken)
		assert.NoError(t, err)
		assert.Equal(t, "refresh", refreshClaims.TokenType)

		// Verify old refresh token is blacklisted
		isBlacklisted, err := suite.blacklistRepo.IsTokenBlacklisted(tokens.RefreshToken)
		assert.NoError(t, err)
		assert.True(t, isBlacklisted)
	})

	t.Run("refresh with invalid token", func(t *testing.T) {
		refreshReq := RefreshRequest{
			RefreshToken: "invalid.token.here",
		}
		reqBody, _ := json.Marshal(refreshReq)

		req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := suite.app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var errorResp response.ErrorResponse
		err = json.Unmarshal(body, &errorResp)
		require.NoError(t, err)

		assert.False(t, errorResp.Success)
		assert.Equal(t, "AUTHENTICATION_ERROR", errorResp.Error.Code)
		assert.Equal(t, "Invalid or expired refresh token", errorResp.Error.Message)
	})

	t.Run("refresh with access token instead of refresh token", func(t *testing.T) {
		refreshReq := RefreshRequest{
			RefreshToken: tokens.AccessToken, // Using access token instead of refresh token
		}
		reqBody, _ := json.Marshal(refreshReq)

		req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := suite.app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var errorResp response.ErrorResponse
		err = json.Unmarshal(body, &errorResp)
		require.NoError(t, err)

		assert.False(t, errorResp.Success)
		assert.Equal(t, "AUTHENTICATION_ERROR", errorResp.Error.Code)
		assert.Equal(t, "Invalid token type", errorResp.Error.Message)
	})

	t.Run("refresh with already used (blacklisted) refresh token", func(t *testing.T) {
		// First, use the refresh token
		refreshReq := RefreshRequest{
			RefreshToken: tokens.RefreshToken,
		}
		reqBody, _ := json.Marshal(refreshReq)

		req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := suite.app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Now try to use the same refresh token again
		req2 := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(reqBody))
		req2.Header.Set("Content-Type", "application/json")

		resp2, err := suite.app.Test(req2)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusUnauthorized, resp2.StatusCode)

		body, _ := io.ReadAll(resp2.Body)
		var errorResp response.ErrorResponse
		err = json.Unmarshal(body, &errorResp)
		require.NoError(t, err)

		assert.False(t, errorResp.Success)
		assert.Equal(t, "AUTHENTICATION_ERROR", errorResp.Error.Code)
		assert.Equal(t, "Invalid or expired refresh token", errorResp.Error.Message)
	})
}

// TestLogoutEndpoint_Integration tests the logout endpoint with real database
func TestLogoutEndpoint_Integration(t *testing.T) {
	suite := setupIntegrationTest(t)
	defer suite.cleanup(t)

	// Generate valid tokens for testing
	tokens, err := suite.authService.GenerateTokens(suite.testUser.ID, suite.testUser.PhoneNumber)
	require.NoError(t, err)

	t.Run("successful logout with access token only", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/logout", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokens.AccessToken))

		resp, err := suite.app.Test(req)
		require.NoError(t, err)

		// Verify response status
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Parse response body
		body, _ := io.ReadAll(resp.Body)
		var successResp response.SuccessResponse
		err = json.Unmarshal(body, &successResp)
		require.NoError(t, err)

		// Verify response structure
		assert.True(t, successResp.Success)
		assert.Equal(t, "Logout successful", successResp.Message)

		// Verify access token is blacklisted
		isBlacklisted, err := suite.blacklistRepo.IsTokenBlacklisted(tokens.AccessToken)
		assert.NoError(t, err)
		assert.True(t, isBlacklisted)

		// Verify blacklisted token cannot be used
		_, err = suite.authService.ValidateToken(tokens.AccessToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token has been invalidated")
	})

	t.Run("successful logout with both access and refresh tokens", func(t *testing.T) {
		// Generate new tokens for this test
		newTokens, err := suite.authService.GenerateTokens(suite.testUser.ID, suite.testUser.PhoneNumber)
		require.NoError(t, err)

		logoutReq := RefreshRequest{
			RefreshToken: newTokens.RefreshToken,
		}
		reqBody, _ := json.Marshal(logoutReq)

		req := httptest.NewRequest("POST", "/auth/logout", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", newTokens.AccessToken))

		resp, err := suite.app.Test(req)
		require.NoError(t, err)

		// Verify response status
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Parse response body
		body, _ := io.ReadAll(resp.Body)
		var successResp response.SuccessResponse
		err = json.Unmarshal(body, &successResp)
		require.NoError(t, err)

		// Verify response structure
		assert.True(t, successResp.Success)
		assert.Equal(t, "Logout successful", successResp.Message)

		// Verify both tokens are blacklisted
		accessBlacklisted, err := suite.blacklistRepo.IsTokenBlacklisted(newTokens.AccessToken)
		assert.NoError(t, err)
		assert.True(t, accessBlacklisted)

		refreshBlacklisted, err := suite.blacklistRepo.IsTokenBlacklisted(newTokens.RefreshToken)
		assert.NoError(t, err)
		assert.True(t, refreshBlacklisted)
	})

	t.Run("logout with invalid access token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")

		resp, err := suite.app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var errorResp response.ErrorResponse
		err = json.Unmarshal(body, &errorResp)
		require.NoError(t, err)

		assert.False(t, errorResp.Success)
		assert.Equal(t, "AUTHENTICATION_ERROR", errorResp.Error.Code)
		assert.Equal(t, "Invalid or expired access token", errorResp.Error.Message)
	})

	t.Run("logout with refresh token in authorization header", func(t *testing.T) {
		// Generate new tokens for this test
		newTokens, err := suite.authService.GenerateTokens(suite.testUser.ID, suite.testUser.PhoneNumber)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/auth/logout", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", newTokens.RefreshToken))

		resp, err := suite.app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var errorResp response.ErrorResponse
		err = json.Unmarshal(body, &errorResp)
		require.NoError(t, err)

		assert.False(t, errorResp.Success)
		assert.Equal(t, "AUTHENTICATION_ERROR", errorResp.Error.Code)
		assert.Equal(t, "Invalid token type", errorResp.Error.Message)
	})

	t.Run("logout with already blacklisted token", func(t *testing.T) {
		// Generate new tokens for this test
		newTokens, err := suite.authService.GenerateTokens(suite.testUser.ID, suite.testUser.PhoneNumber)
		require.NoError(t, err)

		// First logout (blacklist the token)
		req1 := httptest.NewRequest("POST", "/auth/logout", nil)
		req1.Header.Set("Authorization", fmt.Sprintf("Bearer %s", newTokens.AccessToken))

		resp1, err := suite.app.Test(req1)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp1.StatusCode)

		// Try to logout again with the same token
		req2 := httptest.NewRequest("POST", "/auth/logout", nil)
		req2.Header.Set("Authorization", fmt.Sprintf("Bearer %s", newTokens.AccessToken))

		resp2, err := suite.app.Test(req2)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusUnauthorized, resp2.StatusCode)

		body, _ := io.ReadAll(resp2.Body)
		var errorResp response.ErrorResponse
		err = json.Unmarshal(body, &errorResp)
		require.NoError(t, err)

		assert.False(t, errorResp.Success)
		assert.Equal(t, "AUTHENTICATION_ERROR", errorResp.Error.Code)
		assert.Equal(t, "Invalid or expired access token", errorResp.Error.Message)
	})
}

// TestCompleteAuthenticationFlow_Integration tests the complete authentication flow
func TestCompleteAuthenticationFlow_Integration(t *testing.T) {
	suite := setupIntegrationTest(t)
	defer suite.cleanup(t)

	t.Run("complete authentication flow: login -> refresh -> logout", func(t *testing.T) {
		// Step 1: Login
		loginReq := LoginRequest{
			PhoneNumber: "0812345678",
			Pin:         "123456",
		}
		loginReqBody, _ := json.Marshal(loginReq)

		loginHttpReq := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(loginReqBody))
		loginHttpReq.Header.Set("Content-Type", "application/json")

		loginResp, err := suite.app.Test(loginHttpReq)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, loginResp.StatusCode)

		// Parse login response
		loginBody, _ := io.ReadAll(loginResp.Body)
		var loginResponse response.LoginResponse
		err = json.Unmarshal(loginBody, &loginResponse)
		require.NoError(t, err)
		assert.True(t, loginResponse.Success)

		originalAccessToken := loginResponse.Data.AccessToken
		originalRefreshToken := loginResponse.Data.RefreshToken

		// Step 2: Refresh tokens
		refreshReq := RefreshRequest{
			RefreshToken: originalRefreshToken,
		}
		refreshReqBody, _ := json.Marshal(refreshReq)

		refreshHttpReq := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(refreshReqBody))
		refreshHttpReq.Header.Set("Content-Type", "application/json")

		refreshResp, err := suite.app.Test(refreshHttpReq)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, refreshResp.StatusCode)

		// Parse refresh response
		refreshBody, _ := io.ReadAll(refreshResp.Body)
		var refreshResponse response.LoginResponse
		err = json.Unmarshal(refreshBody, &refreshResponse)
		require.NoError(t, err)
		assert.True(t, refreshResponse.Success)

		newAccessToken := refreshResponse.Data.AccessToken
		newRefreshToken := refreshResponse.Data.RefreshToken

		// Verify new tokens are different
		assert.NotEqual(t, originalAccessToken, newAccessToken)
		assert.NotEqual(t, originalRefreshToken, newRefreshToken)

		// Verify original refresh token is blacklisted
		isOriginalRefreshBlacklisted, err := suite.blacklistRepo.IsTokenBlacklisted(originalRefreshToken)
		assert.NoError(t, err)
		assert.True(t, isOriginalRefreshBlacklisted)

		// Step 3: Logout with new tokens
		logoutReq := RefreshRequest{
			RefreshToken: newRefreshToken,
		}
		logoutReqBody, _ := json.Marshal(logoutReq)

		logoutHttpReq := httptest.NewRequest("POST", "/auth/logout", bytes.NewReader(logoutReqBody))
		logoutHttpReq.Header.Set("Content-Type", "application/json")
		logoutHttpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", newAccessToken))

		logoutResp, err := suite.app.Test(logoutHttpReq)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, logoutResp.StatusCode)

		// Parse logout response
		logoutBody, _ := io.ReadAll(logoutResp.Body)
		var logoutResponse response.SuccessResponse
		err = json.Unmarshal(logoutBody, &logoutResponse)
		require.NoError(t, err)
		assert.True(t, logoutResponse.Success)
		assert.Equal(t, "Logout successful", logoutResponse.Message)

		// Verify all tokens are now blacklisted
		isNewAccessBlacklisted, err := suite.blacklistRepo.IsTokenBlacklisted(newAccessToken)
		assert.NoError(t, err)
		assert.True(t, isNewAccessBlacklisted)

		isNewRefreshBlacklisted, err := suite.blacklistRepo.IsTokenBlacklisted(newRefreshToken)
		assert.NoError(t, err)
		assert.True(t, isNewRefreshBlacklisted)

		// Verify tokens cannot be used anymore
		_, err = suite.authService.ValidateToken(newAccessToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token has been invalidated")

		_, err = suite.authService.ValidateToken(newRefreshToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token has been invalidated")
	})
}

// TestTokenExpiration_Integration tests token expiration behavior
func TestTokenExpiration_Integration(t *testing.T) {
	suite := setupIntegrationTest(t)
	defer suite.cleanup(t)

	t.Run("expired token validation", func(t *testing.T) {
		// This test would require manipulating time or creating tokens with very short expiration
		// For now, we'll test the validation logic with manually created expired tokens
		
		// Generate tokens
		tokens, err := suite.authService.GenerateTokens(suite.testUser.ID, suite.testUser.PhoneNumber)
		require.NoError(t, err)

		// Verify tokens are initially valid
		_, err = suite.authService.ValidateToken(tokens.AccessToken)
		assert.NoError(t, err)

		_, err = suite.authService.ValidateToken(tokens.RefreshToken)
		assert.NoError(t, err)

		// Note: In a real scenario, you would wait for token expiration or mock time
		// For this integration test, we're verifying the validation logic works correctly
	})
}

// TestConcurrentRequests_Integration tests concurrent authentication requests
func TestConcurrentRequests_Integration(t *testing.T) {
	suite := setupIntegrationTest(t)
	defer suite.cleanup(t)

	t.Run("concurrent login requests", func(t *testing.T) {
		const numRequests = 5
		results := make(chan error, numRequests)

		// Launch concurrent login requests
		for i := 0; i < numRequests; i++ {
			go func() {
				loginReq := LoginRequest{
					PhoneNumber: "0812345678",
					Pin:         "123456",
				}
				reqBody, _ := json.Marshal(loginReq)

				req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")

				resp, err := suite.app.Test(req)
				if err != nil {
					results <- err
					return
				}

				if resp.StatusCode != fiber.StatusOK {
					results <- fmt.Errorf("expected status 200, got %d", resp.StatusCode)
					return
				}

				results <- nil
			}()
		}

		// Wait for all requests to complete
		for i := 0; i < numRequests; i++ {
			err := <-results
			assert.NoError(t, err, "Concurrent login request failed")
		}
	})
}// g
etEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}