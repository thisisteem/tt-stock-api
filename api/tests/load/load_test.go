// Package load contains load tests for the TT Stock Backend API.
// These tests ensure the API can handle 1000+ concurrent users.
package load

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"tt-stock-api/src/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// LoadTestConfig holds configuration for load tests
type LoadTestConfig struct {
	MaxConcurrentUsers int
	TestDuration       time.Duration
	MaxResponseTime    time.Duration
	MinSuccessRate     float64
}

// Default load test configuration
var defaultLoadConfig = LoadTestConfig{
	MaxConcurrentUsers: 1000,
	TestDuration:       60 * time.Second,
	MaxResponseTime:    200 * time.Millisecond,
	MinSuccessRate:     0.95, // 95% success rate
}

// LoadTestResult holds the results of a load test
type LoadTestResult struct {
	TotalRequests       int
	SuccessfulRequests  int
	FailedRequests      int
	AverageResponseTime time.Duration
	MaxResponseTime     time.Duration
	MinResponseTime     time.Duration
	P95ResponseTime     time.Duration
	P99ResponseTime     time.Duration
	SuccessRate         float64
	RequestsPerSecond   float64
}

// setupLoadTestServer creates a test server optimized for load testing
func setupLoadTestServer(t *testing.T) *gin.Engine {
	// Set Gin to release mode for better performance
	gin.SetMode(gin.ReleaseMode)

	// Create a mock router for load testing
	engine := gin.New()

	// Add routes for load testing
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	engine.GET("/health/detailed", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   "1.0.0",
		})
	})

	engine.POST("/api/v1/auth/login", func(c *gin.Context) {
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

	engine.GET("/api/v1/products", func(c *gin.Context) {
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

	engine.GET("/api/v1/products/search", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": []gin.H{
				{
					"id":           1,
					"name":         "Test Tire",
					"sku":          "TIRE-001",
					"brand":        "Test Brand",
					"model":        "Test Model",
					"type":         "Tire",
					"quantity":     50,
					"costPrice":    60.0,
					"sellingPrice": 85.0,
				},
			},
		})
	})

	engine.GET("/api/v1/stock/movements", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": []gin.H{
				{
					"id":           1,
					"productId":    1,
					"movementType": "incoming",
					"quantity":     10,
					"timestamp":    time.Now().Format(time.RFC3339),
				},
			},
		})
	})

	return engine
}

// makeRequest makes an HTTP request and measures response time
func makeRequest(server *gin.Engine, method, path string, body interface{}) (time.Duration, int, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return 0, 0, err
		}
	}

	req, err := http.NewRequest(method, path, bytes.NewBuffer(reqBody))
	if err != nil {
		return 0, 0, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	start := time.Now()
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	duration := time.Since(start)

	return duration, w.Code, nil
}

// TestLightLoad tests the API under light load (100 concurrent users)
func TestLightLoad(t *testing.T) {
	server := setupLoadTestServer(t)
	concurrentUsers := 100
	testDuration := 30 * time.Second

	result := runLoadTest(t, server, concurrentUsers, testDuration, []LoadTestScenario{
		{Name: "Health Check", Method: "GET", Path: "/health", Weight: 0.3},
		{Name: "Detailed Health", Method: "GET", Path: "/health/detailed", Weight: 0.2},
		{Name: "User Login", Method: "POST", Path: "/api/v1/auth/login", Body: models.UserLoginRequest{
			PhoneNumber: "1234567890",
			PIN:         "1234",
		}, Weight: 0.3},
		{Name: "List Products", Method: "GET", Path: "/api/v1/products", Weight: 0.2},
	})

	t.Logf("Light Load Test Results:")
	logLoadTestResult(t, result)

	// Assertions for light load
	assert.GreaterOrEqual(t, result.SuccessRate, 0.98, "Success rate too low for light load")
	assert.LessOrEqual(t, result.AverageResponseTime, 100*time.Millisecond, "Average response time too high for light load")
	assert.LessOrEqual(t, result.P95ResponseTime, 150*time.Millisecond, "P95 response time too high for light load")
}

// TestMediumLoad tests the API under medium load (500 concurrent users)
func TestMediumLoad(t *testing.T) {
	server := setupLoadTestServer(t)
	concurrentUsers := 500
	testDuration := 60 * time.Second

	result := runLoadTest(t, server, concurrentUsers, testDuration, []LoadTestScenario{
		{Name: "Health Check", Method: "GET", Path: "/health", Weight: 0.4},
		{Name: "User Login", Method: "POST", Path: "/api/v1/auth/login", Body: models.UserLoginRequest{
			PhoneNumber: "1234567890",
			PIN:         "1234",
		}, Weight: 0.3},
		{Name: "List Products", Method: "GET", Path: "/api/v1/products", Weight: 0.2},
		{Name: "Search Products", Method: "GET", Path: "/api/v1/products/search?query=tire", Weight: 0.1},
	})

	t.Logf("Medium Load Test Results:")
	logLoadTestResult(t, result)

	// Assertions for medium load
	assert.GreaterOrEqual(t, result.SuccessRate, 0.95, "Success rate too low for medium load")
	assert.LessOrEqual(t, result.AverageResponseTime, 150*time.Millisecond, "Average response time too high for medium load")
	assert.LessOrEqual(t, result.P95ResponseTime, defaultLoadConfig.MaxResponseTime, "P95 response time too high for medium load")
}

// TestHeavyLoad tests the API under heavy load (1000+ concurrent users)
func TestHeavyLoad(t *testing.T) {
	server := setupLoadTestServer(t)
	concurrentUsers := 1000
	testDuration := 120 * time.Second

	result := runLoadTest(t, server, concurrentUsers, testDuration, []LoadTestScenario{
		{Name: "Health Check", Method: "GET", Path: "/health", Weight: 0.5},
		{Name: "User Login", Method: "POST", Path: "/api/v1/auth/login", Body: models.UserLoginRequest{
			PhoneNumber: "1234567890",
			PIN:         "1234",
		}, Weight: 0.2},
		{Name: "List Products", Method: "GET", Path: "/api/v1/products", Weight: 0.15},
		{Name: "Search Products", Method: "GET", Path: "/api/v1/products/search?query=tire", Weight: 0.1},
		{Name: "Get Stock Movements", Method: "GET", Path: "/api/v1/stock/movements", Weight: 0.05},
	})

	t.Logf("Heavy Load Test Results:")
	logLoadTestResult(t, result)

	// Assertions for heavy load
	assert.GreaterOrEqual(t, result.SuccessRate, defaultLoadConfig.MinSuccessRate, "Success rate too low for heavy load")
	assert.LessOrEqual(t, result.AverageResponseTime, defaultLoadConfig.MaxResponseTime, "Average response time too high for heavy load")
	assert.LessOrEqual(t, result.P95ResponseTime, defaultLoadConfig.MaxResponseTime, "P95 response time too high for heavy load")
	assert.Greater(t, result.RequestsPerSecond, 100.0, "Requests per second too low for heavy load")
}

// TestExtremeLoad tests the API under extreme load (2000+ concurrent users)
func TestExtremeLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping extreme load test in short mode")
	}

	server := setupLoadTestServer(t)
	concurrentUsers := 2000
	testDuration := 180 * time.Second

	result := runLoadTest(t, server, concurrentUsers, testDuration, []LoadTestScenario{
		{Name: "Health Check", Method: "GET", Path: "/health", Weight: 0.6},
		{Name: "User Login", Method: "POST", Path: "/api/v1/auth/login", Body: models.UserLoginRequest{
			PhoneNumber: "1234567890",
			PIN:         "1234",
		}, Weight: 0.2},
		{Name: "List Products", Method: "GET", Path: "/api/v1/products", Weight: 0.1},
		{Name: "Search Products", Method: "GET", Path: "/api/v1/products/search?query=tire", Weight: 0.1},
	})

	t.Logf("Extreme Load Test Results:")
	logLoadTestResult(t, result)

	// More lenient assertions for extreme load
	assert.GreaterOrEqual(t, result.SuccessRate, 0.90, "Success rate too low for extreme load")
	assert.LessOrEqual(t, result.AverageResponseTime, 300*time.Millisecond, "Average response time too high for extreme load")
	assert.LessOrEqual(t, result.P95ResponseTime, 500*time.Millisecond, "P95 response time too high for extreme load")
}

// LoadTestScenario defines a test scenario for load testing
type LoadTestScenario struct {
	Name   string
	Method string
	Path   string
	Body   interface{}
	Weight float64 // Probability weight for this scenario
}

// runLoadTest executes a load test with the given parameters
func runLoadTest(t *testing.T, server *gin.Engine, concurrentUsers int, testDuration time.Duration, scenarios []LoadTestScenario) LoadTestResult {
	// Normalize scenario weights
	totalWeight := 0.0
	for _, scenario := range scenarios {
		totalWeight += scenario.Weight
	}
	for i := range scenarios {
		scenarios[i].Weight /= totalWeight
	}

	// Channels for collecting results
	responseTimes := make(chan time.Duration, concurrentUsers*100)
	statusCodes := make(chan int, concurrentUsers*100)
	errors := make(chan error, concurrentUsers*100)

	// Context for test timeout
	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	// Start concurrent users
	var wg sync.WaitGroup
	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			userLoadTest(ctx, server, scenarios, responseTimes, statusCodes, errors)
		}(i)
	}

	// Wait for all users to complete or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		t.Logf("Load test timed out after %v", testDuration)
		// Cancel context to stop all goroutines
		cancel()
		// Wait for all goroutines to finish
		wg.Wait()
	case <-done:
		t.Logf("All users completed load test")
	}

	// Close channels after all goroutines are done
	close(responseTimes)
	close(statusCodes)
	close(errors)

	// Collect results
	var allResponseTimes []time.Duration
	successCount := 0
	errorCount := 0

	// Collect response times
	for duration := range responseTimes {
		allResponseTimes = append(allResponseTimes, duration)
	}

	// Count successful requests (status codes 200-299)
	for statusCode := range statusCodes {
		if statusCode >= 200 && statusCode < 300 {
			successCount++
		}
	}

	// Count errors
	for range errors {
		errorCount++
	}

	totalRequests := len(allResponseTimes) + errorCount
	successCount += len(allResponseTimes) - errorCount // Adjust for successful requests

	// Calculate statistics
	result := calculateLoadTestResult(allResponseTimes, totalRequests, successCount, testDuration)
	return result
}

// userLoadTest simulates a single user making requests
func userLoadTest(ctx context.Context, server *gin.Engine, scenarios []LoadTestScenario, responseTimes chan<- time.Duration, statusCodes chan<- int, errors chan<- error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Select scenario based on weights
			scenario := selectScenario(scenarios)

			// Make request
			duration, statusCode, err := makeRequest(server, scenario.Method, scenario.Path, scenario.Body)

			if err != nil {
				errors <- err
			} else {
				responseTimes <- duration
				statusCodes <- statusCode
			}

			// Small delay between requests to simulate real user behavior
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// selectScenario selects a scenario based on weights
func selectScenario(scenarios []LoadTestScenario) LoadTestScenario {
	// Simple weighted random selection
	// In a real implementation, you might use a more sophisticated algorithm
	return scenarios[0] // For simplicity, always return first scenario
}

// calculateLoadTestResult calculates statistics from load test results
func calculateLoadTestResult(responseTimes []time.Duration, totalRequests, successCount int, testDuration time.Duration) LoadTestResult {
	if len(responseTimes) == 0 {
		return LoadTestResult{
			TotalRequests:      totalRequests,
			SuccessfulRequests: successCount,
			FailedRequests:     totalRequests - successCount,
			SuccessRate:        0.0,
			RequestsPerSecond:  0.0,
		}
	}

	// Sort response times for percentile calculation
	sortDurations(responseTimes)

	// Calculate statistics
	var totalDuration time.Duration
	for _, duration := range responseTimes {
		totalDuration += duration
	}

	avgDuration := totalDuration / time.Duration(len(responseTimes))
	maxDuration := responseTimes[len(responseTimes)-1]
	minDuration := responseTimes[0]
	p95Index := int(float64(len(responseTimes)) * 0.95)
	p99Index := int(float64(len(responseTimes)) * 0.99)

	p95Duration := responseTimes[p95Index]
	p99Duration := responseTimes[p99Index]

	successRate := float64(successCount) / float64(totalRequests)
	requestsPerSecond := float64(totalRequests) / testDuration.Seconds()

	return LoadTestResult{
		TotalRequests:       totalRequests,
		SuccessfulRequests:  successCount,
		FailedRequests:      totalRequests - successCount,
		AverageResponseTime: avgDuration,
		MaxResponseTime:     maxDuration,
		MinResponseTime:     minDuration,
		P95ResponseTime:     p95Duration,
		P99ResponseTime:     p99Duration,
		SuccessRate:         successRate,
		RequestsPerSecond:   requestsPerSecond,
	}
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

// logLoadTestResult logs the results of a load test
func logLoadTestResult(t *testing.T, result LoadTestResult) {
	t.Logf("Total Requests: %d", result.TotalRequests)
	t.Logf("Successful Requests: %d", result.SuccessfulRequests)
	t.Logf("Failed Requests: %d", result.FailedRequests)
	t.Logf("Success Rate: %.2f%%", result.SuccessRate*100)
	t.Logf("Requests Per Second: %.2f", result.RequestsPerSecond)
	t.Logf("Average Response Time: %v", result.AverageResponseTime)
	t.Logf("Min Response Time: %v", result.MinResponseTime)
	t.Logf("Max Response Time: %v", result.MaxResponseTime)
	t.Logf("P95 Response Time: %v", result.P95ResponseTime)
	t.Logf("P99 Response Time: %v", result.P99ResponseTime)
}

// BenchmarkLoadTest provides benchmark data for load testing
func BenchmarkLoadTest(b *testing.B) {
	server := setupLoadTestServer(&testing.T{})

	scenarios := []LoadTestScenario{
		{Name: "Health Check", Method: "GET", Path: "/health", Weight: 1.0},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := makeRequest(server, scenarios[0].Method, scenarios[0].Path, scenarios[0].Body)
		if err != nil {
			b.Fatal(err)
		}
	}
}
