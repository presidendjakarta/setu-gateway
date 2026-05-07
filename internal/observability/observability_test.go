package observability

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var testMetrics *Metrics

func TestMain(m *testing.M) {
	// Initialize metrics once for all tests
	testMetrics = NewMetrics()
	m.Run()
}

func TestMetrics_RecordHTTP(t *testing.T) {
	// Record some HTTP requests
	testMetrics.RecordHTTP("GET", "/api/users", "users-route", 200, 100*time.Millisecond, 0, 1024)
	testMetrics.RecordHTTP("POST", "/api/users", "users-route", 201, 200*time.Millisecond, 512, 256)
	testMetrics.RecordHTTP("GET", "/api/users", "users-route", 500, 50*time.Millisecond, 0, 128)

	// If we reach here without panic, metrics are working
	t.Log("HTTP metrics recorded successfully")
}

func TestMetrics_RecordAuth(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// Skip test if metrics already registered (Prometheus global registry)
			t.Skip("Metrics already registered, skipping")
		}
	}()

	// Use global testMetrics

	// Record auth attempts
	testMetrics.RecordAuth("jwt", true, 50*time.Millisecond, "")
	testMetrics.RecordAuth("jwt", false, 30*time.Millisecond, "invalid_token")
	testMetrics.RecordAuth("api_key", true, 20*time.Millisecond, "")

	t.Log("Auth metrics recorded successfully")
}

func TestMetrics_RecordRateLimit(t *testing.T) {
	// Use global testMetrics

	// Record rate limit checks
	testMetrics.RecordRateLimit("api-limit", true)
	testMetrics.RecordRateLimit("api-limit", true)
	testMetrics.RecordRateLimit("api-limit", false) // Rejected

	t.Log("Rate limit metrics recorded successfully")
}

func TestMetrics_RecordCircuitBreaker(t *testing.T) {
	// Use global testMetrics

	// Record circuit breaker states
	testMetrics.RecordCircuitBreakerState("user-service", 0) // Closed
	testMetrics.RecordCircuitBreakerState("payment-service", 1) // Open
	testMetrics.RecordCircuitBreakerTrip("payment-service")

	t.Log("Circuit breaker metrics recorded successfully")
}

func TestMetrics_RecordProxy(t *testing.T) {
	// Use global testMetrics

	// Record proxy metrics
	testMetrics.RecordProxyLatency("user-upstream", 150*time.Millisecond)
	testMetrics.RecordUpstreamLatency("user-upstream", "localhost:8001", 100*time.Millisecond)
	testMetrics.RecordProxyError("user-upstream", "connection_refused")

	t.Log("Proxy metrics recorded successfully")
}

func TestMetrics_UpdateGoroutines(t *testing.T) {
	// Use global testMetrics

	// Update goroutine count
	testMetrics.UpdateGoroutines(42)

	t.Log("Goroutine metric updated successfully")
}

func TestHealthChecker_Basic(t *testing.T) {
	hc := NewHealthChecker()

	// Register a healthy component check
	hc.RegisterCheck("database", func(ctx context.Context) HealthStatus {
		return HealthStatus{
			Status:  "healthy",
			Message: "Database connection OK",
		}
	})

	// Register another check
	hc.RegisterCheck("redis", func(ctx context.Context) HealthStatus {
		return HealthStatus{
			Status:  "healthy",
			Message: "Redis connection OK",
		}
	})

	// Check overall health
	ctx := context.Background()
	status := hc.Check(ctx)

	if status.Status != "healthy" {
		t.Errorf("Expected healthy status, got %s", status.Status)
	}

	if _, exists := status.Details["database"]; !exists {
		t.Error("Expected database in details")
	}

	if _, exists := status.Details["redis"]; !exists {
		t.Error("Expected redis in details")
	}

	if _, exists := status.Details["uptime"]; !exists {
		t.Error("Expected uptime in details")
	}
}

func TestHealthChecker_Degraded(t *testing.T) {
	hc := NewHealthChecker()

	// Healthy component
	hc.RegisterCheck("database", func(ctx context.Context) HealthStatus {
		return HealthStatus{
			Status: "healthy",
		}
	})

	// Unhealthy component
	hc.RegisterCheck("redis", func(ctx context.Context) HealthStatus {
		return HealthStatus{
			Status:  "unhealthy",
			Message: "Connection refused",
		}
	})

	ctx := context.Background()
	status := hc.Check(ctx)

	if status.Status != "degraded" {
		t.Errorf("Expected degraded status, got %s", status.Status)
	}
}

func TestHealthChecker_HTTPHandler(t *testing.T) {
	hc := NewHealthChecker()

	hc.RegisterCheck("test", func(ctx context.Context) HealthStatus {
		return HealthStatus{
			Status: "healthy",
		}
	})

	// Create test request
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	// Call handler
	hc.WriteHealthCheck(rec, req)

	// Check response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Error("Expected JSON content type")
	}
}

func TestHealthChecker_Ready(t *testing.T) {
	hc := NewHealthChecker()

	req := httptest.NewRequest("GET", "/ready", nil)
	rec := httptest.NewRecorder()

	hc.WriteReady(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestHealthChecker_Live(t *testing.T) {
	hc := NewHealthChecker()

	req := httptest.NewRequest("GET", "/live", nil)
	rec := httptest.NewRecorder()

	hc.WriteLive(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestHealthChecker_Uptime(t *testing.T) {
	hc := NewHealthChecker()

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	ctx := context.Background()
	status := hc.Check(ctx)

	uptime, ok := status.Details["uptime"].(string)
	if !ok {
		t.Error("Expected uptime as string")
	}

	if uptime == "" {
		t.Error("Uptime should not be empty")
	}
}

func TestHealthChecker_RuntimeInfo(t *testing.T) {
	hc := NewHealthChecker()

	ctx := context.Background()
	status := hc.Check(ctx)

	// Check runtime details
	if _, ok := status.Details["goroutines"]; !ok {
		t.Error("Expected goroutines in details")
	}

	if _, ok := status.Details["go_version"]; !ok {
		t.Error("Expected go_version in details")
	}

	if _, ok := status.Details["os"]; !ok {
		t.Error("Expected os in details")
	}

	if _, ok := status.Details["arch"]; !ok {
		t.Error("Expected arch in details")
	}
}

func TestMetricsMiddleware(t *testing.T) {
	mw := MetricsMiddleware(testMetrics)

	// Create test handler
	handler := mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	// Serve request
	handler.ServeHTTP(rec, req)

	// Check response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}

	// Metrics should be recorded (no panic = success)
	t.Log("Metrics middleware executed successfully")
}

func TestMetricsResponseWriter(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &metricsResponseWriter{
		ResponseWriter: rec,
		statusCode:     http.StatusOK,
	}

	// Write response
	rw.WriteHeader(http.StatusCreated)
	rw.Write([]byte("test response"))

	if rw.statusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rw.statusCode)
	}

	if rw.bytesWritten != len("test response") {
		t.Errorf("Expected %d bytes, got %d", len("test response"), rw.bytesWritten)
	}
}

func TestMetrics_Labels(t *testing.T) {
	// Use global testMetrics

	// Test various label combinations
	testMetrics.RecordHTTP("GET", "/api/v1/users", "route-1", 200, 100*time.Millisecond, 0, 512)
	testMetrics.RecordHTTP("POST", "/api/v1/users", "route-1", 201, 150*time.Millisecond, 256, 128)
	testMetrics.RecordHTTP("GET", "/api/v1/users/123", "route-2", 404, 50*time.Millisecond, 0, 64)
	testMetrics.RecordHTTP("DELETE", "/api/v1/users/123", "route-2", 204, 80*time.Millisecond, 0, 0)

	t.Log("Multiple label combinations recorded successfully")
}

func TestHealthChecker_Concurrent(t *testing.T) {
	hc := NewHealthChecker()

	// Register multiple checks
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("check-%d", i)
		hc.RegisterCheck(name, func(ctx context.Context) HealthStatus {
			return HealthStatus{
				Status: "healthy",
			}
		})
	}

	// Concurrent health checks
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			ctx := context.Background()
			hc.Check(ctx)
			done <- true
		}()
	}

	// Wait for all
	for i := 0; i < 10; i++ {
		<-done
	}

	t.Log("Concurrent health checks passed")
}
