package ratelimiter

import (
	"context"
	"sync"
	"time"
)

// TokenBucket implements token bucket rate limiting algorithm
type TokenBucket struct {
	mu         sync.Mutex
	capacity   int           // Maximum tokens
	tokens     float64       // Current tokens
	refillRate float64       // Tokens per second
	lastRefill time.Time     // Last refill time
}

// NewTokenBucket creates a new token bucket rate limiter
func NewTokenBucket(capacity int, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     float64(capacity),
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed
func (tb *TokenBucket) Allow() bool {
	return tb.AllowN(1)
}

// AllowN checks if N requests are allowed
func (tb *TokenBucket) AllowN(n int) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens
	tb.refill()

	// Check if enough tokens
	if tb.tokens >= float64(n) {
		tb.tokens -= float64(n)
		return true
	}

	return false
}

// refill adds tokens based on elapsed time
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	// Calculate new tokens
	newTokens := elapsed * tb.refillRate
	tb.tokens = min(tb.tokens+newTokens, float64(tb.capacity))

	// Update last refill time
	tb.lastRefill = now
}

// Wait blocks until a token is available or context is cancelled
func (tb *TokenBucket) Wait(ctx context.Context) error {
	for {
		if tb.Allow() {
			return nil
		}

		// Wait a bit before retrying
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
			// Retry
		}
	}
}

// Tokens returns current token count
func (tb *TokenBucket) Tokens() float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	
	tb.refill()
	return tb.tokens
}

// Capacity returns the bucket capacity
func (tb *TokenBucket) Capacity() int {
	return tb.capacity
}

// Reset resets the bucket to full capacity
func (tb *TokenBucket) Reset() {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	
	tb.tokens = float64(tb.capacity)
	tb.lastRefill = time.Now()
}

// min returns the minimum of two values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
