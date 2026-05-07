package observability

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// HealthStatus represents the health status of a component
type HealthStatus struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// HealthChecker manages health checks for all components
type HealthChecker struct {
	mu       sync.RWMutex
	checks   map[string]HealthCheck
	startedAt time.Time
}

// HealthCheck is a function that checks component health
type HealthCheck func(ctx context.Context) HealthStatus

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks:    make(map[string]HealthCheck),
		startedAt: time.Now(),
	}
}

// RegisterCheck registers a health check
func (hc *HealthChecker) RegisterCheck(name string, check HealthCheck) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.checks[name] = check
}

// Check runs all health checks and returns overall status
func (hc *HealthChecker) Check(ctx context.Context) HealthStatus {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	overall := HealthStatus{
		Status:  "healthy",
		Details: make(map[string]interface{}),
	}

	// Uptime
	uptime := time.Since(hc.startedAt)
	overall.Details["uptime"] = uptime.String()
	overall.Details["started_at"] = hc.startedAt.Format(time.RFC3339)

	// Go runtime info
	overall.Details["goroutines"] = runtime.NumGoroutine()
	overall.Details["go_version"] = runtime.Version()
	overall.Details["os"] = runtime.GOOS
	overall.Details["arch"] = runtime.GOARCH

	// Run all checks
	for name, check := range hc.checks {
		status := check(ctx)
		overall.Details[name] = status

		// If any component is unhealthy, mark overall as degraded
		if status.Status != "healthy" && overall.Status == "healthy" {
			overall.Status = "degraded"
		}
	}

	return overall
}

// WriteHealthCheck writes health check response
func (hc *HealthChecker) WriteHealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	status := hc.Check(ctx)

	w.Header().Set("Content-Type", "application/json")

	// Return 503 if unhealthy
	if status.Status == "unhealthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(status)
}

// WriteReady writes readiness probe response
func (hc *HealthChecker) WriteReady(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ready",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// WriteLive writes liveness probe response
func (hc *HealthChecker) WriteLive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now().Unix(),
	})
}
