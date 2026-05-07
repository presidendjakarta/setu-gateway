package circuitbreaker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/presidendjakarta/setu-gateway/internal/logger"
)

var errTest = errors.New("test error")

func TestCircuitBreaker_ClosedState(t *testing.T) {
	cb := New(Config{
		Name:             "test",
		FailureThreshold: 3,
		SuccessThreshold: 1,
		Timeout:          1 * time.Second,
	})

	// Should start closed
	if cb.State() != StateClosed {
		t.Errorf("Expected closed state, got %s", cb.State())
	}

	// Should allow requests
	if !cb.allowRequest() {
		t.Error("Should allow requests in closed state")
	}
}

func TestCircuitBreaker_TransitionToOpen(t *testing.T) {
	cb := New(Config{
		Name:             "test",
		FailureThreshold: 3,
		SuccessThreshold: 1,
		Timeout:          1 * time.Second,
	})

	// Simulate 3 failures
	for i := 0; i < 3; i++ {
		err := cb.Execute(func() error {
			return errTest
		})
		if err != errTest {
			t.Errorf("Expected test error, got %v", err)
		}
	}

	// Should be open now
	if cb.State() != StateOpen {
		t.Errorf("Expected open state after 3 failures, got %s", cb.State())
	}

	// Should reject requests
	if cb.allowRequest() {
		t.Error("Should reject requests in open state")
	}
}

func TestCircuitBreaker_TransitionToHalfOpen(t *testing.T) {
	cb := New(Config{
		Name:             "test",
		FailureThreshold: 2,
		SuccessThreshold: 1,
		Timeout:          100 * time.Millisecond,
	})

	// Trip the circuit breaker
	cb.Execute(func() error { return errTest })
	cb.Execute(func() error { return errTest })

	if cb.State() != StateOpen {
		t.Fatal("Should be open")
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Should transition to half-open
	if cb.State() != StateHalfOpen {
		t.Errorf("Expected half-open after timeout, got %s", cb.State())
	}

	// Should allow test request
	if !cb.allowRequest() {
		t.Error("Should allow test request in half-open state")
	}
}

func TestCircuitBreaker_HalfOpenToClosed(t *testing.T) {
	cb := New(Config{
		Name:             "test",
		FailureThreshold: 2,
		SuccessThreshold: 2,
		Timeout:          100 * time.Millisecond,
	})

	// Trip the circuit breaker
	cb.Execute(func() error { return errTest })
	cb.Execute(func() error { return errTest })

	// Wait for half-open
	time.Sleep(150 * time.Millisecond)

	// Success in half-open
	cb.Execute(func() error { return nil })
	cb.Execute(func() error { return nil })

	// Should be closed now
	if cb.State() != StateClosed {
		t.Errorf("Expected closed after successes, got %s", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenToOpen(t *testing.T) {
	cb := New(Config{
		Name:             "test",
		FailureThreshold: 2,
		SuccessThreshold: 2,
		Timeout:          100 * time.Millisecond,
	})

	// Trip the circuit breaker
	cb.Execute(func() error { return errTest })
	cb.Execute(func() error { return errTest })

	// Wait for half-open
	time.Sleep(150 * time.Millisecond)

	// Failure in half-open should go back to open
	cb.Execute(func() error { return errTest })

	if cb.State() != StateOpen {
		t.Errorf("Expected open after failure in half-open, got %s", cb.State())
	}
}

func TestCircuitBreaker_Statistics(t *testing.T) {
	cb := New(Config{
		Name:             "test",
		FailureThreshold: 5,
		SuccessThreshold: 1,
		Timeout:          1 * time.Second,
	})

	// Mix of successes and failures
	cb.Execute(func() error { return nil })
	cb.Execute(func() error { return nil })
	cb.Execute(func() error { return errTest })
	cb.Execute(func() error { return nil })

	stats := cb.Stats()

	if stats["total_requests"] != int64(4) {
		t.Errorf("Expected 4 total requests, got %v", stats["total_requests"])
	}

	if stats["total_successes"] != int64(3) {
		t.Errorf("Expected 3 successes, got %v", stats["total_successes"])
	}

	if stats["total_failures"] != int64(1) {
		t.Errorf("Expected 1 failure, got %v", stats["total_failures"])
	}
}

func TestCircuitBreaker_RejectedCount(t *testing.T) {
	cb := New(Config{
		Name:             "test",
		FailureThreshold: 2,
		SuccessThreshold: 1,
		Timeout:          10 * time.Second, // Long timeout
	})

	// Trip the circuit breaker
	cb.Execute(func() error { return errTest })
	cb.Execute(func() error { return errTest })

	// Try rejected requests
	for i := 0; i < 5; i++ {
		cb.Execute(func() error { return nil })
	}

	stats := cb.Stats()
	if stats["total_rejected"] != int64(5) {
		t.Errorf("Expected 5 rejected, got %v", stats["total_rejected"])
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := New(Config{
		Name:             "test",
		FailureThreshold: 2,
		SuccessThreshold: 1,
		Timeout:          1 * time.Second,
	})

	// Trip the circuit breaker
	cb.Execute(func() error { return errTest })
	cb.Execute(func() error { return errTest })

	if cb.State() != StateOpen {
		t.Fatal("Should be open")
	}

	// Reset
	cb.Reset()

	// Should be closed with zero counters
	if cb.State() != StateClosed {
		t.Errorf("Expected closed after reset, got %s", cb.State())
	}

	stats := cb.Stats()
	if stats["total_requests"] != int64(0) {
		t.Error("Reset should zero out counters")
	}
}

func TestCircuitBreaker_IsClosedIsOpen(t *testing.T) {
	cb := New(Config{
		Name:             "test",
		FailureThreshold: 2,
		SuccessThreshold: 1,
		Timeout:          1 * time.Second,
	})

	if !cb.IsClosed() {
		t.Error("Should be closed initially")
	}

	if cb.IsOpen() {
		t.Error("Should not be open initially")
	}

	// Trip it
	cb.Execute(func() error { return errTest })
	cb.Execute(func() error { return errTest })

	if cb.IsClosed() {
		t.Error("Should not be closed after failures")
	}

	if !cb.IsOpen() {
		t.Error("Should be open after failures")
	}
}

func TestCircuitBreaker_Concurrent(t *testing.T) {
	cb := New(Config{
		Name:             "test",
		FailureThreshold: 100,
		SuccessThreshold: 1,
		Timeout:          1 * time.Second,
	})

	done := make(chan bool)

	// Concurrent requests
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				cb.Execute(func() error {
					if j%10 == 0 {
						return errTest
					}
					return nil
				})
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	stats := cb.Stats()
	if stats["total_requests"] != int64(1000) {
		t.Errorf("Expected 1000 requests, got %v", stats["total_requests"])
	}
}

func TestRetry_BasicSuccess(t *testing.T) {
	ctx := context.Background()
	cfg := DefaultRetryConfig()

	attempt := 0
	err := Retry(ctx, cfg, func() error {
		attempt++
		if attempt < 2 {
			return errTest
		}
		return nil
	})

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}

	if attempt != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempt)
	}
}

func TestRetry_AllAttemptsFail(t *testing.T) {
	ctx := context.Background()
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	err := Retry(ctx, cfg, func() error {
		return errTest
	})

	if err != errTest {
		t.Errorf("Expected test error, got %v", err)
	}
}

func TestRetry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	cfg := DefaultRetryConfig()

	err := Retry(ctx, cfg, func() error {
		return errTest
	})

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}

func TestRetry_ExponentialBackoff(t *testing.T) {
	ctx := context.Background()
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.0,
	}

	start := time.Now()
	attempts := 0

	Retry(ctx, cfg, func() error {
		attempts++
		return errTest
	})

	elapsed := time.Since(start)

	// Expected delays: 50ms + 100ms + 200ms = 350ms (approximately)
	if elapsed < 300*time.Millisecond {
		t.Errorf("Expected at least 300ms, got %v", elapsed)
	}
}

func TestRetryWithCircuitBreaker(t *testing.T) {
	ctx := context.Background()
	retryCfg := RetryConfig{
		MaxAttempts:  2,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	cb := New(Config{
		Name:             "test",
		FailureThreshold: 5,
		SuccessThreshold: 1,
		Timeout:          1 * time.Second,
	})

	attempt := 0
	err := RetryWithCircuitBreaker(ctx, retryCfg, cb, func() error {
		attempt++
		if attempt < 2 {
			return errTest
		}
		return nil
	})

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
}

func TestManager_CreateAndGet(t *testing.T) {
	log := setupTestLogger(t)
	mgr := NewManager(log)

	cb := mgr.CreateBreaker("test-breaker", 5, 1*time.Second)
	if cb == nil {
		t.Fatal("Expected circuit breaker")
	}

	retrieved, exists := mgr.GetBreaker("test-breaker")
	if !exists {
		t.Fatal("Expected to retrieve breaker")
	}

	if retrieved != cb {
		t.Error("Expected same breaker instance")
	}
}

func TestManager_GetOrCreateBreaker(t *testing.T) {
	log := setupTestLogger(t)
	mgr := NewManager(log)

	// Create
	cb1 := mgr.GetOrCreateBreaker("auto-breaker", 3, 1*time.Second)

	// Get existing
	cb2 := mgr.GetOrCreateBreaker("auto-breaker", 10, 5*time.Second)

	if cb1 != cb2 {
		t.Error("Should return existing breaker")
	}
}

func TestManager_ListBreakers(t *testing.T) {
	log := setupTestLogger(t)
	mgr := NewManager(log)

	mgr.CreateBreaker("breaker-1", 3, 1*time.Second)
	mgr.CreateBreaker("breaker-2", 5, 2*time.Second)
	mgr.CreateBreaker("breaker-3", 7, 3*time.Second)

	names := mgr.ListBreakers()
	if len(names) != 3 {
		t.Errorf("Expected 3 breakers, got %d", len(names))
	}
}

func TestManager_Stats(t *testing.T) {
	log := setupTestLogger(t)
	mgr := NewManager(log)

	cb := mgr.CreateBreaker("stats-test", 3, 1*time.Second)
	cb.Execute(func() error { return nil })

	stats := mgr.Stats()
	if _, exists := stats["stats-test"]; !exists {
		t.Error("Expected stats for stats-test")
	}
}

func TestManager_ResetAll(t *testing.T) {
	log := setupTestLogger(t)
	mgr := NewManager(log)

	mgr.CreateBreaker("reset-1", 2, 1*time.Second)
	mgr.CreateBreaker("reset-2", 2, 1*time.Second)

	// Trip both breakers
	cb1, _ := mgr.GetBreaker("reset-1")
	cb2, _ := mgr.GetBreaker("reset-2")

	cb1.Execute(func() error { return errTest })
	cb1.Execute(func() error { return errTest })
	cb2.Execute(func() error { return errTest })
	cb2.Execute(func() error { return errTest })

	// Reset all
	mgr.ResetAll()

	// Should be closed
	if !cb1.IsClosed() || !cb2.IsClosed() {
		t.Error("All breakers should be closed after reset")
	}
}

func TestManager_HealthCheck(t *testing.T) {
	log := setupTestLogger(t)
	mgr := NewManager(log)

	mgr.CreateBreaker("healthy", 5, 1*time.Second)
	mgr.CreateBreaker("unhealthy", 2, 1*time.Second)

	// Trip unhealthy breaker
	cb, _ := mgr.GetBreaker("unhealthy")
	cb.Execute(func() error { return errTest })
	cb.Execute(func() error { return errTest })

	health := mgr.HealthCheck()

	if !health["healthy"] {
		t.Error("Expected healthy breaker to be healthy")
	}

	if health["unhealthy"] {
		t.Error("Expected unhealthy breaker to be unhealthy")
	}
}

func setupTestLogger(t *testing.T) *logger.Logger {
	log, err := logger.New("error", "console", "stdout")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	return log
}
