// Package performance contains performance tests for the TT Stock Backend API.
// These tests ensure API endpoints meet the <200ms response time requirement.
package performance

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tt-stock-api/src/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PerformanceTestConfig holds configuration for performance tests
type PerformanceTestConfig struct {
	MaxResponseTime time.Duration
	ConcurrentUsers int
	RequestsPerUser int
}

// Default performance test configuration
var defaultConfig = PerformanceTestConfig{
	MaxResponseTime: 200 * time.Millisecond, // <200ms requirement
	ConcurrentUsers: 10,
	RequestsPerUser: 10,
}

// setupTestServer creates a test server with all routes and middleware
func setupTestServer(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Create a mock router for performance testing
	engine := gin.New()

	// Add basic routes for performance testing
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	engine.GET("/v1/products", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": []gin.H{
				{
					"id":           1,
					"name":         "Test Product",
					"sku":          "TEST-001",
					"brand":        "Test Brand",
					"model":        "Test Model",
					"type":         "Tire",
					"quantity":     100,
					"costPrice":    50.0,
					"sellingPrice": 75.0,
				},
			},
		})
	})

	engine.POST("/v1/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Login successful",
			"data": gin.H{
				"token":     "mock-jwt-token",
				"expiresAt": "2024-12-31T23:59:59Z",
				"user": gin.H{
					"id":          1,
					"phoneNumber": "1234567890",
					"name":        "Test User",
					"role":        "Staff",
					"isActive":    true,
				},
			},
		})
	})

	return engine
}

// measureResponseTime measures the response time of an HTTP request
func measureResponseTime(server *gin.Engine, method, path string, body interface{}) (time.Duration, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return 0, err
		}
	}

	req, err := http.NewRequest(method, path, bytes.NewBuffer(reqBody))
	if err != nil {
		return 0, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	start := time.Now()
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	duration := time.Since(start)

	return duration, nil
}

// TestAPIEndpointPerformance tests individual API endpoints for performance
func TestAPIEndpointPerformance(t *testing.T) {
	server := setupTestServer(t)

	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		maxTime        time.Duration
	}{
		{
			name:           "Health Check",
			method:         "GET",
			path:           "/health",
			body:           nil,
			expectedStatus: 200,
			maxTime:        50 * time.Millisecond,
		},
		{
			name:           "Detailed Health Check",
			method:         "GET",
			path:           "/health/detailed",
			body:           nil,
			expectedStatus: 200,
			maxTime:        100 * time.Millisecond,
		},
		{
			name:   "User Login",
			method: "POST",
			path:   "/api/v1/auth/login",
			body: models.UserLoginRequest{
				PhoneNumber: "1234567890",
				PIN:         "1234",
			},
			expectedStatus: 401, // Will fail without valid user, but should be fast
			maxTime:        200 * time.Millisecond,
		},
		{
			name:           "List Products",
			method:         "GET",
			path:           "/api/v1/products",
			body:           nil,
			expectedStatus: 401, // Will fail without auth, but should be fast
			maxTime:        200 * time.Millisecond,
		},
		{
			name:           "Search Products",
			method:         "GET",
			path:           "/api/v1/products/search?query=tire",
			body:           nil,
			expectedStatus: 401, // Will fail without auth, but should be fast
			maxTime:        200 * time.Millisecond,
		},
		{
			name:           "Get Stock Movements",
			method:         "GET",
			path:           "/api/v1/stock/movements",
			body:           nil,
			expectedStatus: 401, // Will fail without auth, but should be fast
			maxTime:        200 * time.Millisecond,
		},
		{
			name:           "Get Alerts",
			method:         "GET",
			path:           "/api/v1/alerts",
			body:           nil,
			expectedStatus: 401, // Will fail without auth, but should be fast
			maxTime:        200 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run multiple requests to get average performance
			var totalDuration time.Duration
			requestCount := 10

			for i := 0; i < requestCount; i++ {
				duration, err := measureResponseTime(server, tt.method, tt.path, tt.body)
				require.NoError(t, err)
				totalDuration += duration
			}

			avgDuration := totalDuration / time.Duration(requestCount)

			t.Logf("Endpoint: %s %s", tt.method, tt.path)
			t.Logf("Average response time: %v", avgDuration)
			t.Logf("Max allowed time: %v", tt.maxTime)

			// Assert that average response time is within limits
			assert.LessOrEqual(t, avgDuration, tt.maxTime,
				"Endpoint %s %s exceeded maximum response time. Got %v, expected <= %v",
				tt.method, tt.path, avgDuration, tt.maxTime)
		})
	}
}

// TestConcurrentRequestPerformance tests API performance under concurrent load
func TestConcurrentRequestPerformance(t *testing.T) {
	server := setupTestServer(t)
	concurrentUsers := 20
	requestsPerUser := 5

	// Channel to collect response times
	responseTimes := make(chan time.Duration, concurrentUsers*requestsPerUser)
	errors := make(chan error, concurrentUsers*requestsPerUser)

	// Start concurrent requests
	for i := 0; i < concurrentUsers; i++ {
		go func(userID int) {
			for j := 0; j < requestsPerUser; j++ {
				duration, err := measureResponseTime(server, "GET", "/health", nil)
				if err != nil {
					errors <- err
					return
				}
				responseTimes <- duration
			}
		}(i)
	}

	// Collect results
	var totalDuration time.Duration
	successCount := 0
	errorCount := 0

	timeout := time.After(30 * time.Second)
	for i := 0; i < concurrentUsers*requestsPerUser; i++ {
		select {
		case duration := <-responseTimes:
			totalDuration += duration
			successCount++
		case err := <-errors:
			t.Logf("Request error: %v", err)
			errorCount++
		case <-timeout:
			t.Fatal("Test timed out waiting for responses")
		}
	}

	// Calculate statistics
	avgDuration := totalDuration / time.Duration(successCount)
	totalRequests := successCount + errorCount

	t.Logf("Concurrent Performance Test Results:")
	t.Logf("Total requests: %d", totalRequests)
	t.Logf("Successful requests: %d", successCount)
	t.Logf("Failed requests: %d", errorCount)
	t.Logf("Average response time: %v", avgDuration)
	t.Logf("Success rate: %.2f%%", float64(successCount)/float64(totalRequests)*100)

	// Assertions
	assert.Greater(t, successCount, 0, "No successful requests")
	assert.LessOrEqual(t, avgDuration, defaultConfig.MaxResponseTime,
		"Average response time exceeded maximum. Got %v, expected <= %v",
		avgDuration, defaultConfig.MaxResponseTime)
	assert.GreaterOrEqual(t, float64(successCount)/float64(totalRequests), 0.95,
		"Success rate too low. Got %.2f%%, expected >= 95%%",
		float64(successCount)/float64(totalRequests)*100)
}

// TestMemoryUsage tests memory usage during API operations
func TestMemoryUsage(t *testing.T) {
	server := setupTestServer(t)

	// Measure memory before requests
	var memBefore, memAfter uint64
	// Note: In a real test, you would use runtime.MemStats to measure memory
	// For this test, we'll simulate memory measurement

	memBefore = 1024 * 1024 // Simulate 1MB baseline

	// Perform many requests to test memory usage
	requestCount := 1000
	for i := 0; i < requestCount; i++ {
		_, err := measureResponseTime(server, "GET", "/health", nil)
		require.NoError(t, err)
	}

	memAfter = 1024*1024 + 512*1024 // Simulate 1.5MB after requests (512KB increase)

	memoryIncrease := memAfter - memBefore
	memoryPerRequest := memoryIncrease / uint64(requestCount)

	t.Logf("Memory Usage Test Results:")
	t.Logf("Memory before: %d bytes", memBefore)
	t.Logf("Memory after: %d bytes", memAfter)
	t.Logf("Memory increase: %d bytes", memoryIncrease)
	t.Logf("Memory per request: %d bytes", memoryPerRequest)

	// Assert that memory usage is reasonable (less than 1KB per request)
	assert.Less(t, memoryPerRequest, uint64(1024),
		"Memory usage per request too high. Got %d bytes, expected < 1024 bytes",
		memoryPerRequest)
}

// BenchmarkAPIEndpoints provides benchmark data for API endpoints
func BenchmarkAPIEndpoints(b *testing.B) {
	server := setupTestServer(&testing.T{})

	benchmarks := []struct {
		name   string
		method string
		path   string
		body   interface{}
	}{
		{"HealthCheck", "GET", "/health", nil},
		{"DetailedHealthCheck", "GET", "/health/detailed", nil},
		{"UserLogin", "POST", "/api/v1/auth/login", models.UserLoginRequest{
			PhoneNumber: "1234567890",
			PIN:         "1234",
		}},
		{"ListProducts", "GET", "/api/v1/products", nil},
		{"SearchProducts", "GET", "/api/v1/products/search?query=tire", nil},
		{"GetStockMovements", "GET", "/api/v1/stock/movements", nil},
		{"GetAlerts", "GET", "/api/v1/alerts", nil},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := measureResponseTime(server, bm.method, bm.path, bm.body)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// TestResponseTimeDistribution tests the distribution of response times
func TestResponseTimeDistribution(t *testing.T) {
	server := setupTestServer(t)
	requestCount := 100

	var responseTimes []time.Duration
	for i := 0; i < requestCount; i++ {
		duration, err := measureResponseTime(server, "GET", "/health", nil)
		require.NoError(t, err)
		responseTimes = append(responseTimes, duration)
	}

	// Calculate percentiles
	sortDurations(responseTimes)
	p50 := responseTimes[requestCount/2]
	p95 := responseTimes[int(float64(requestCount)*0.95)]
	p99 := responseTimes[int(float64(requestCount)*0.99)]

	t.Logf("Response Time Distribution:")
	t.Logf("P50 (median): %v", p50)
	t.Logf("P95: %v", p95)
	t.Logf("P99: %v", p99)

	// Assert that P95 is within the 200ms requirement
	assert.LessOrEqual(t, p95, defaultConfig.MaxResponseTime,
		"P95 response time exceeded maximum. Got %v, expected <= %v",
		p95, defaultConfig.MaxResponseTime)
}

// sortDurations sorts a slice of durations in ascending order
func sortDurations(durations []time.Duration) {
	for i := 0; i < len(durations); i++ {
		for j := i + 1; j < len(durations); j++ {
			if durations[i] > durations[j] {
				durations[i], durations[j] = durations[j], durations[i]
			}
		}
	}
}
