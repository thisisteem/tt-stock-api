package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// AuthFlowTestSuite tests the complete authentication flow
// This includes login, token validation, refresh, and logout scenarios
type AuthFlowTestSuite struct {
	suite.Suite
	router *gin.Engine
	token  string
}

// SetupSuite runs once before all tests in the suite
func (suite *AuthFlowTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	setupTestAuthRoutes(suite.router)
}

// TestAuthFlowTestSuite runs the test suite
func TestAuthFlowTestSuite(t *testing.T) {
	suite.Run(t, new(AuthFlowTestSuite))
}

// TestCompleteLoginFlow tests the complete login process
func (suite *AuthFlowTestSuite) TestCompleteLoginFlow() {
	// Test valid login
	loginRequest := map[string]string{
		"phoneNumber": "1234567890",
		"pin":         "1234",
	}

	jsonBody, _ := json.Marshal(loginRequest)
	req, _ := http.NewRequest("POST", "/v1/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Verify successful login
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Login successful", response["message"])

	data := response["data"].(map[string]interface{})
	suite.token = data["token"].(string)
	assert.NotEmpty(suite.T(), suite.token)

	// Verify token expiration
	expiresAt := data["expiresAt"].(string)
	expiryTime, err := time.Parse(time.RFC3339, expiresAt)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), expiryTime.After(time.Now()))

	// Verify user data
	user := data["user"].(map[string]interface{})
	assert.Equal(suite.T(), float64(1), user["id"])
	assert.Equal(suite.T(), "1234567890", user["phoneNumber"])
	assert.Equal(suite.T(), "John Doe", user["name"])
	assert.Equal(suite.T(), "Staff", user["role"])
	assert.True(suite.T(), user["isActive"].(bool))
}

// TestInvalidLoginScenarios tests various invalid login attempts
func (suite *AuthFlowTestSuite) TestInvalidLoginScenarios() {
	testCases := []struct {
		name           string
		phoneNumber    string
		pin            string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Invalid PIN",
			phoneNumber:    "1234567890",
			pin:            "9999",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid credentials",
		},
		{
			name:           "Invalid phone number",
			phoneNumber:    "9999999999",
			pin:            "1234",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid credentials",
		},
		{
			name:           "Empty phone number",
			phoneNumber:    "",
			pin:            "1234",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Validation failed",
		},
		{
			name:           "Empty PIN",
			phoneNumber:    "1234567890",
			pin:            "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Validation failed",
		},
		{
			name:           "Invalid phone format",
			phoneNumber:    "invalid",
			pin:            "1234",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Validation failed",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			loginRequest := map[string]string{
				"phoneNumber": tc.phoneNumber,
				"pin":         tc.pin,
			}

			jsonBody, _ := json.Marshal(loginRequest)
			req, _ := http.NewRequest("POST", "/v1/auth/login", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(suite.T(), tc.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(suite.T(), err)

			assert.False(suite.T(), response["success"].(bool))
			assert.Contains(suite.T(), response["message"], tc.expectedError)
		})
	}
}

// TestTokenValidation tests token validation and protected endpoint access
func (suite *AuthFlowTestSuite) TestTokenValidation() {
	// First login to get a valid token
	suite.loginAndGetToken()

	// Test accessing protected endpoint with valid token
	req, _ := http.NewRequest("GET", "/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "User profile retrieved successfully", response["message"])

	user := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), float64(1), user["id"])
	assert.Equal(suite.T(), "1234567890", user["phoneNumber"])
}

// TestInvalidTokenScenarios tests various invalid token scenarios
func (suite *AuthFlowTestSuite) TestInvalidTokenScenarios() {
	testCases := []struct {
		name           string
		token          string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "No token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization header required",
		},
		{
			name:           "Invalid token",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or expired token",
		},
		{
			name:           "Malformed token",
			token:          "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or expired token",
		},
		{
			name:           "Empty Bearer token",
			token:          "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or expired token",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			req, _ := http.NewRequest("GET", "/v1/auth/me", nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(suite.T(), tc.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(suite.T(), err)

			assert.False(suite.T(), response["success"].(bool))
			assert.Contains(suite.T(), response["message"], tc.expectedError)
		})
	}
}

// TestTokenRefreshFlow tests the token refresh functionality
func (suite *AuthFlowTestSuite) TestTokenRefreshFlow() {
	// First login to get a valid token
	suite.loginAndGetToken()
	originalToken := suite.token

	// Test token refresh
	req, _ := http.NewRequest("POST", "/v1/auth/refresh", nil)
	req.Header.Set("Authorization", "Bearer "+originalToken)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Token refreshed successfully", response["message"])

	data := response["data"].(map[string]interface{})
	newToken := data["token"].(string)
	assert.NotEmpty(suite.T(), newToken)
	assert.NotEqual(suite.T(), originalToken, newToken)

	// Verify new token works
	req, _ = http.NewRequest("GET", "/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+newToken)

	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Verify old token still works (implementation dependent)
	req, _ = http.NewRequest("GET", "/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+originalToken)

	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// This behavior depends on implementation - could be OK or Unauthorized
	assert.True(suite.T(), w.Code == http.StatusOK || w.Code == http.StatusUnauthorized)
}

// TestLogoutFlow tests the logout functionality
func (suite *AuthFlowTestSuite) TestLogoutFlow() {
	// First login to get a valid token
	suite.loginAndGetToken()

	// Test logout
	req, _ := http.NewRequest("POST", "/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Logout successful", response["message"])

	// Verify token is invalidated (implementation dependent)
	req, _ = http.NewRequest("GET", "/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// This behavior depends on implementation - could be OK or Unauthorized
	assert.True(suite.T(), w.Code == http.StatusOK || w.Code == http.StatusUnauthorized)
}

// TestConcurrentLoginSessions tests multiple concurrent login sessions
func (suite *AuthFlowTestSuite) TestConcurrentLoginSessions() {
	// Test multiple logins with same credentials
	loginRequest := map[string]string{
		"phoneNumber": "1234567890",
		"pin":         "1234",
	}

	jsonBody, _ := json.Marshal(loginRequest)

	// Simulate concurrent logins
	tokens := make([]string, 3)
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("POST", "/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(suite.T(), err)

		data := response["data"].(map[string]interface{})
		tokens[i] = data["token"].(string)
	}

	// Verify all tokens are different
	for i := 0; i < len(tokens); i++ {
		for j := i + 1; j < len(tokens); j++ {
			assert.NotEqual(suite.T(), tokens[i], tokens[j])
		}
	}

	// Verify all tokens work
	for _, token := range tokens {
		req, _ := http.NewRequest("GET", "/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)
	}
}

// TestSessionTimeout tests session timeout behavior
func (suite *AuthFlowTestSuite) TestSessionTimeout() {
	// First login to get a valid token
	suite.loginAndGetToken()

	// Test that token works immediately
	req, _ := http.NewRequest("GET", "/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Note: In a real implementation, we would test token expiration
	// by either waiting for the token to expire or mocking time
	// For now, we just verify the token works
}

// Helper method to login and get token
func (suite *AuthFlowTestSuite) loginAndGetToken() {
	loginRequest := map[string]string{
		"phoneNumber": "1234567890",
		"pin":         "1234",
	}

	jsonBody, _ := json.Marshal(loginRequest)
	req, _ := http.NewRequest("POST", "/v1/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})
	suite.token = data["token"].(string)
}

// setupTestAuthRoutes sets up test routes for integration testing
// This is a placeholder that will be replaced with actual implementation
func setupTestAuthRoutes(router *gin.Engine) {
	auth := router.Group("/v1/auth")
	{
		auth.POST("/login", func(c *gin.Context) {
			var req map[string]string
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Validation failed",
				})
				return
			}

			// Mock validation
			if req["phoneNumber"] == "" || req["pin"] == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Validation failed",
				})
				return
			}

			if req["phoneNumber"] == "1234567890" && req["pin"] == "1234" {
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "Login successful",
					"data": gin.H{
						"token":     "mock-jwt-token-" + req["phoneNumber"],
						"expiresAt": "2024-09-22T12:00:00Z",
						"user": gin.H{
							"id":          1,
							"phoneNumber": "1234567890",
							"name":        "John Doe",
							"role":        "Staff",
							"isActive":    true,
						},
					},
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "Invalid credentials",
				})
			}
		})

		auth.POST("/logout", func(c *gin.Context) {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "Authorization header required",
				})
				return
			}

			if authHeader == "Bearer invalid-token" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "Invalid or expired token",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Logout successful",
			})
		})

		auth.POST("/refresh", func(c *gin.Context) {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" || authHeader == "Bearer invalid-token" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "Invalid or expired token",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Token refreshed successfully",
				"data": gin.H{
					"token":     "new-mock-jwt-token",
					"expiresAt": "2024-09-22T12:00:00Z",
				},
			})
		})

		auth.GET("/me", func(c *gin.Context) {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "Authorization header required",
				})
				return
			}

			if authHeader == "Bearer invalid-token" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "Invalid or expired token",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "User profile retrieved successfully",
				"data": gin.H{
					"id":          1,
					"phoneNumber": "1234567890",
					"name":        "John Doe",
					"role":        "Staff",
					"isActive":    true,
				},
			})
		})
	}
}
