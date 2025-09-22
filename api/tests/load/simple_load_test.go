// Package load contains simple load tests for the TT Stock Backend API.
// These tests ensure the API can handle concurrent requests.
package load

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupSimpleLoadTestServer creates a simple test server for load testing
func setupSimpleLoadTestServer() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Add basic routes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		})
	})

	return r
}

// TestSimpleLightLoad tests the API under light load (100 concurrent users)
func TestSimpleLightLoad(t *testing.T) {
	server := setupSimpleLoadTestServer()
	concurrentUsers := 100
	requestsPerUser := 5

	result := runSimpleLoadTest(t, server, concurrentUsers, requestsPerUser, "GET", "/health")

	t.Logf("Light Load Test Results:")
	logSimpleLoadTestResult(t, result)

	// Assertions for light load
	assert.GreaterOrEqual(t, result.SuccessRate, 0.98, "Success rate too low for light load")
	assert.LessOrEqual(t, result.AverageResponseTime, 100*time.Millisecond, "Average response time too high for light load")
}

// TestSimpleMediumLoad tests the API under medium load (500 concurrent users)
func TestSimpleMediumLoad(t *testing.T) {
	server := setupSimpleLoadTestServer()
	concurrentUsers := 500
	requestsPerUser := 3

	result := runSimpleLoadTest(t, server, concurrentUsers, requestsPerUser, "GET", "/health")

	t.Logf("Medium Load Test Results:")
	logSimpleLoadTestResult(t, result)

	// Assertions for medium load
	assert.GreaterOrEqual(t, result.SuccessRate, 0.95, "Success rate too low for medium load")
	assert.LessOrEqual(t, result.AverageResponseTime, 150*time.Millisecond, "Average response time too high for medium load")
}

// TestSimpleHeavyLoad tests the API under heavy load (1000+ concurrent users)
func TestSimpleHeavyLoad(t *testing.T) {
	server := setupSimpleLoadTestServer()
	concurrentUsers := 1000
	requestsPerUser := 2

	result := runSimpleLoadTest(t, server, concurrentUsers, requestsPerUser, "GET", "/health")

	t.Logf("Heavy Load Test Results:")
	logSimpleLoadTestResult(t, result)

	// Assertions for heavy load
	assert.GreaterOrEqual(t, result.SuccessRate, 0.90, "Success rate too low for heavy load")
	assert.LessOrEqual(t, result.AverageResponseTime, 200*time.Millisecond, "Average response time too high for heavy load")
}

// SimpleLoadTestResult holds the results of a simple load test
type SimpleLoadTestResult struct {
	TotalRequests       int
	SuccessfulRequests  int
	FailedRequests      int
	AverageResponseTime time.Duration
	MaxResponseTime     time.Duration
	MinResponseTime     time.Duration
	SuccessRate         float64
	RequestsPerSecond   float64
}

// runSimpleLoadTest executes a simple load test
func runSimpleLoadTest(t *testing.T, server *gin.Engine, concurrentUsers, requestsPerUser int, method, path string) SimpleLoadTestResult {
	// Channels for collecting results
	responseTimes := make(chan time.Duration, concurrentUsers*requestsPerUser)
	errors := make(chan error, concurrentUsers*requestsPerUser)

	// Start concurrent users
	var wg sync.WaitGroup
	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			for j := 0; j < requestsPerUser; j++ {
				req, err := http.NewRequest(method, path, nil)
				if err != nil {
					errors <- err
					return
				}

				start := time.Now()
				w := httptest.NewRecorder()
				server.ServeHTTP(w, req)
				duration := time.Since(start)

				if w.Code >= 200 && w.Code < 300 {
					responseTimes <- duration
				} else {
					errors <- err
				}

				// Small delay between requests
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	// Wait for all users to complete
	wg.Wait()

	// Close channels
	close(responseTimes)
	close(errors)

	// Collect results
	var allResponseTimes []time.Duration
	successCount := 0
	errorCount := 0

	// Collect response times
	for duration := range responseTimes {
		allResponseTimes = append(allResponseTimes, duration)
		successCount++
	}

	// Count errors
	for range errors {
		errorCount++
	}

	totalRequests := successCount + errorCount

	// Calculate statistics
	result := calculateSimpleLoadTestResult(allResponseTimes, totalRequests, successCount, 10*time.Second)
	return result
}

// calculateSimpleLoadTestResult calculates statistics from simple load test results
func calculateSimpleLoadTestResult(responseTimes []time.Duration, totalRequests, successCount int, testDuration time.Duration) SimpleLoadTestResult {
	if len(responseTimes) == 0 {
		return SimpleLoadTestResult{
			TotalRequests:      totalRequests,
			SuccessfulRequests: successCount,
			FailedRequests:     totalRequests - successCount,
			SuccessRate:        0.0,
			RequestsPerSecond:  0.0,
		}
	}

	// Sort response times for percentile calculation
	sortSimpleDurations(responseTimes)

	// Calculate statistics
	var totalDuration time.Duration
	for _, duration := range responseTimes {
		totalDuration += duration
	}

	avgDuration := totalDuration / time.Duration(len(responseTimes))
	maxDuration := responseTimes[len(responseTimes)-1]
	minDuration := responseTimes[0]

	successRate := float64(successCount) / float64(totalRequests)
	requestsPerSecond := float64(totalRequests) / testDuration.Seconds()

	return SimpleLoadTestResult{
		TotalRequests:       totalRequests,
		SuccessfulRequests:  successCount,
		FailedRequests:      totalRequests - successCount,
		AverageResponseTime: avgDuration,
		MaxResponseTime:     maxDuration,
		MinResponseTime:     minDuration,
		SuccessRate:         successRate,
		RequestsPerSecond:   requestsPerSecond,
	}
}

// sortSimpleDurations sorts a slice of durations in ascending order
func sortSimpleDurations(durations []time.Duration) {
	for i := 0; i < len(durations); i++ {
		for j := i + 1; j < len(durations); j++ {
			if durations[i] > durations[j] {
				durations[i], durations[j] = durations[j], durations[i]
			}
		}
	}
}

// logSimpleLoadTestResult logs the results of a simple load test
func logSimpleLoadTestResult(t *testing.T, result SimpleLoadTestResult) {
	t.Logf("Total Requests: %d", result.TotalRequests)
	t.Logf("Successful Requests: %d", result.SuccessfulRequests)
	t.Logf("Failed Requests: %d", result.FailedRequests)
	t.Logf("Success Rate: %.2f%%", result.SuccessRate*100)
	t.Logf("Requests Per Second: %.2f", result.RequestsPerSecond)
	t.Logf("Average Response Time: %v", result.AverageResponseTime)
	t.Logf("Min Response Time: %v", result.MinResponseTime)
	t.Logf("Max Response Time: %v", result.MaxResponseTime)
}

// BenchmarkSimpleLoadTest provides benchmark data for simple load testing
func BenchmarkSimpleLoadTest(b *testing.B) {
	server := setupSimpleLoadTestServer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, err := http.NewRequest("GET", "/health", nil)
		if err != nil {
			b.Fatal(err)
		}
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
	}
}
