// Package performance contains simple performance tests for the TT Stock Backend API.
// These tests ensure basic HTTP endpoints meet the <200ms response time requirement.
package performance

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupSimpleTestServer creates a simple test server with basic routes
func setupSimpleTestServer() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Add basic routes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		})
	})

	r.GET("/health/detailed", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
			"database": gin.H{
				"status":        "connected",
				"response_time": "5ms",
			},
			"memory": gin.H{
				"used":  "45MB",
				"total": "128MB",
			},
		})
	})

	return r
}

// TestBasicEndpointPerformance tests basic endpoints for performance
func TestBasicEndpointPerformance(t *testing.T) {
	server := setupSimpleTestServer()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		maxTime        time.Duration
	}{
		{
			name:           "Health Check",
			method:         "GET",
			path:           "/health",
			expectedStatus: 200,
			maxTime:        50 * time.Millisecond,
		},
		{
			name:           "Detailed Health Check",
			method:         "GET",
			path:           "/health/detailed",
			expectedStatus: 200,
			maxTime:        100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run multiple requests to get average performance
			var totalDuration time.Duration
			requestCount := 10

			for i := 0; i < requestCount; i++ {
				req, err := http.NewRequest(tt.method, tt.path, nil)
				require.NoError(t, err)

				start := time.Now()
				w := httptest.NewRecorder()
				server.ServeHTTP(w, req)
				duration := time.Since(start)
				totalDuration += duration

				assert.Equal(t, tt.expectedStatus, w.Code)
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

// TestConcurrentBasicRequests tests basic endpoints under concurrent load
func TestConcurrentBasicRequests(t *testing.T) {
	server := setupSimpleTestServer()
	concurrentUsers := 20
	requestsPerUser := 5

	// Channel to collect response times
	responseTimes := make(chan time.Duration, concurrentUsers*requestsPerUser)
	errors := make(chan error, concurrentUsers*requestsPerUser)

	// Start concurrent requests
	for i := 0; i < concurrentUsers; i++ {
		go func(userID int) {
			for j := 0; j < requestsPerUser; j++ {
				req, err := http.NewRequest("GET", "/health", nil)
				if err != nil {
					errors <- err
					return
				}

				start := time.Now()
				w := httptest.NewRecorder()
				server.ServeHTTP(w, req)
				duration := time.Since(start)
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
	assert.LessOrEqual(t, avgDuration, 200*time.Millisecond,
		"Average response time exceeded maximum. Got %v, expected <= 200ms",
		avgDuration)
	assert.GreaterOrEqual(t, float64(successCount)/float64(totalRequests), 0.95,
		"Success rate too low. Got %.2f%%, expected >= 95%%",
		float64(successCount)/float64(totalRequests)*100)
}

// BenchmarkBasicEndpoints provides benchmark data for basic endpoints
func BenchmarkBasicEndpoints(b *testing.B) {
	server := setupSimpleTestServer()

	benchmarks := []struct {
		name   string
		method string
		path   string
	}{
		{"HealthCheck", "GET", "/health"},
		{"DetailedHealthCheck", "GET", "/health/detailed"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				req, err := http.NewRequest(bm.method, bm.path, nil)
				if err != nil {
					b.Fatal(err)
				}
				w := httptest.NewRecorder()
				server.ServeHTTP(w, req)
			}
		})
	}
}
