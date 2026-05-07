package ratelimiter

import (
	"context"
	"testing"
	"time"

	"github.com/presidendjakarta/setu-gateway/internal/logger"
)

func TestTokenBucket_BasicAllow(t *testing.T) {
	// Create bucket with 10 tokens, refill 1 token/second
	tb := NewTokenBucket(10, 1.0)

	// Should allow first 10 requests
	for i := 0; i < 10; i++ {
		if !tb.Allow() {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 11th request should be denied
	if tb.Allow() {
		t.Error("Request 11 should be denied (bucket empty)")
	}
}

func TestTokenBucket_AllowN(t *testing.T) {
	tb := NewTokenBucket(10, 1.0)

	// Allow 5 tokens
	if !tb.AllowN(5) {
		t.Error("Should allow 5 tokens")
	}

	// Try to allow 6 tokens (only 5 left)
	if tb.AllowN(6) {
		t.Error("Should not allow 6 tokens (only 5 available)")
	}

	// Allow remaining 5
	if !tb.AllowN(5) {
		t.Error("Should allow remaining 5 tokens")
	}
}

func TestTokenBucket_Refill(t *testing.T) {
	// Create bucket with 5 tokens, refill 10 tokens/second
	tb := NewTokenBucket(5, 10.0)

	// Exhaust all tokens
	for i := 0; i < 5; i++ {
		tb.Allow()
	}

	// Should be empty
	if tb.Allow() {
		t.Error("Bucket should be empty")
	}

	// Wait for refill (100ms = 1 token at 10/sec)
	time.Sleep(150 * time.Millisecond)

	// Should have ~1-2 tokens now
	if !tb.Allow() {
		t.Error("Should have refilled at least 1 token")
	}
}

func TestTokenBucket_NoOverfill(t *testing.T) {
	tb := NewTokenBucket(10, 100.0) // Fast refill

	// Wait to accumulate tokens
	time.Sleep(200 * time.Millisecond)

	// Should not exceed capacity
	tokens := tb.Tokens()
	if tokens > 10 {
		t.Errorf("Tokens (%f) should not exceed capacity (10)", tokens)
	}
}

func TestTokenBucket_Wait(t *testing.T) {
	tb := NewTokenBucket(1, 10.0) // 1 token, refill 10/sec

	// Use the token
	tb.Allow()

	// Wait should block until token refills
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := tb.Wait(ctx)
	if err != nil {
		t.Errorf("Wait should succeed, got error: %v", err)
	}
}

func TestTokenBucket_WaitTimeout(t *testing.T) {
	tb := NewTokenBucket(1, 0.1) // Very slow refill

	// Use the token
	tb.Allow()

	// Wait should timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := tb.Wait(ctx)
	if err == nil {
		t.Error("Wait should timeout")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded, got: %v", err)
	}
}

func TestTokenBucket_Concurrent(t *testing.T) {
	tb := NewTokenBucket(100, 0) // No refill

	// Concurrent access
	done := make(chan bool)
	allowed := make(chan bool, 200)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 20; j++ {
				if tb.Allow() {
					allowed <- true
				}
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have allowed exactly 100 requests
	count := len(allowed)
	if count != 100 {
		t.Errorf("Expected 100 allowed requests, got %d", count)
	}
}

func TestTokenBucket_Reset(t *testing.T) {
	tb := NewTokenBucket(10, 1.0)

	// Exhaust tokens
	for i := 0; i < 10; i++ {
		tb.Allow()
	}

	if tb.Allow() {
		t.Error("Bucket should be empty")
	}

	// Reset
	tb.Reset()

	// Should have full capacity
	if tb.Tokens() < 10 {
		t.Error("Reset should restore full capacity")
	}
}

func TestTokenBucket_Tokens(t *testing.T) {
	tb := NewTokenBucket(10, 5.0) // 5 tokens/sec

	// Wait 1 second
	time.Sleep(1 * time.Second)

	// Should have ~5 tokens (started at 10, capped at 10)
	tokens := tb.Tokens()
	if tokens < 9 || tokens > 10 {
		t.Errorf("Expected ~10 tokens after 1s, got %f", tokens)
	}
}

func TestTokenBucket_Capacity(t *testing.T) {
	tb := NewTokenBucket(50, 1.0)

	if tb.Capacity() != 50 {
		t.Errorf("Expected capacity 50, got %d", tb.Capacity())
	}
}

func TestManager_CreateAndGet(t *testing.T) {
	log := setupTestLogger(t)
	mgr := NewManager(log)

	// Create limiter
	mgr.CreateLimiter("test-limiter", 100, 10.0)

	// Get limiter
	limiter, exists := mgr.GetLimiter("test-limiter")
	if !exists {
		t.Fatal("Limiter should exist")
	}

	if limiter.Capacity() != 100 {
		t.Errorf("Expected capacity 100, got %d", limiter.Capacity())
	}
}

func TestManager_CheckRateLimit(t *testing.T) {
	log := setupTestLogger(t)
	mgr := NewManager(log)

	// Create limiter: 5 requests
	mgr.CreateLimiter("api-limit", 5, 1.0)

	// Should allow 5 requests
	for i := 0; i < 5; i++ {
		if !mgr.CheckRateLimit("api-limit") {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 6th should be denied
	if mgr.CheckRateLimit("api-limit") {
		t.Error("Request 6 should be denied")
	}
}

func TestManager_NonExistentLimiter(t *testing.T) {
	log := setupTestLogger(t)
	mgr := NewManager(log)

	// Should allow if limiter doesn't exist
	if !mgr.CheckRateLimit("non-existent") {
		t.Error("Should allow request when limiter doesn't exist")
	}
}

func TestManager_RemoveLimiter(t *testing.T) {
	log := setupTestLogger(t)
	mgr := NewManager(log)

	mgr.CreateLimiter("to-remove", 100, 10.0)

	// Verify exists
	_, exists := mgr.GetLimiter("to-remove")
	if !exists {
		t.Fatal("Limiter should exist")
	}

	// Remove
	mgr.RemoveLimiter("to-remove")

	// Verify removed
	_, exists = mgr.GetLimiter("to-remove")
	if exists {
		t.Error("Limiter should be removed")
	}
}

func TestManager_UpdateLimiter(t *testing.T) {
	log := setupTestLogger(t)
	mgr := NewManager(log)

	// Create with 100 capacity
	mgr.CreateLimiter("update-test", 100, 10.0)

	// Update to 200 capacity
	mgr.UpdateLimiter("update-test", 200, 20.0)

	limiter, _ := mgr.GetLimiter("update-test")
	if limiter.Capacity() != 200 {
		t.Errorf("Expected capacity 200 after update, got %d", limiter.Capacity())
	}
}

func TestManager_ListLimiters(t *testing.T) {
	log := setupTestLogger(t)
	mgr := NewManager(log)

	mgr.CreateLimiter("limiter-1", 100, 10.0)
	mgr.CreateLimiter("limiter-2", 200, 20.0)
	mgr.CreateLimiter("limiter-3", 300, 30.0)

	ids := mgr.ListLimiters()
	if len(ids) != 3 {
		t.Errorf("Expected 3 limiters, got %d", len(ids))
	}
}

func TestManager_Clear(t *testing.T) {
	log := setupTestLogger(t)
	mgr := NewManager(log)

	mgr.CreateLimiter("clear-1", 100, 10.0)
	mgr.CreateLimiter("clear-2", 200, 20.0)

	mgr.Clear()

	ids := mgr.ListLimiters()
	if len(ids) != 0 {
		t.Errorf("Expected 0 limiters after clear, got %d", len(ids))
	}
}

func TestManager_Stats(t *testing.T) {
	log := setupTestLogger(t)
	mgr := NewManager(log)

	mgr.CreateLimiter("stats-test", 100, 10.0)

	stats := mgr.Stats()
	if _, exists := stats["stats-test"]; !exists {
		t.Error("Stats should include stats-test limiter")
	}

	limiterStats := stats["stats-test"]
	if limiterStats["capacity"] != 100 {
		t.Errorf("Expected capacity 100, got %v", limiterStats["capacity"])
	}
}

func setupTestLogger(t *testing.T) *logger.Logger {
	// Create minimal logger for tests
	log, err := logger.New("error", "console", "stdout")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	return log
}
