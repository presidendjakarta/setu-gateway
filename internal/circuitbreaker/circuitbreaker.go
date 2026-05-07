package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents circuit breaker state
type State int

const (
	StateClosed   State = iota // Normal operation
	StateOpen                  // Failing, reject requests
	StateHalfOpen              // Testing if service recovered
)

// String returns state name
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu sync.RWMutex

	// Configuration
	name            string
	failureThreshold int           // Failures before opening
	successThreshold int           // Successes in half-open before closing
	timeout         time.Duration // How long to stay open before half-open

	// State
	state          State
	failureCount   int
	successCount   int
	lastFailure    time.Time
	lastStateChange time.Time

	// Statistics
	totalRequests   int64
	totalFailures   int64
	totalSuccesses  int64
	totalRejected   int64
}

// Config holds circuit breaker configuration
type Config struct {
	Name             string
	FailureThreshold int
	SuccessThreshold int
	Timeout          time.Duration
}

// New creates a new circuit breaker
func New(cfg Config) *CircuitBreaker {
	return &CircuitBreaker{
		name:             cfg.Name,
		failureThreshold: cfg.FailureThreshold,
		successThreshold: cfg.SuccessThreshold,
		timeout:          cfg.Timeout,
		state:            StateClosed,
		lastStateChange:  time.Now(),
	}
}

// Execute runs a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	// Check if request should be allowed
	if !cb.allowRequest() {
		cb.mu.Lock()
		cb.totalRejected++
		cb.mu.Unlock()
		return ErrCircuitOpen
	}

	// Execute the function
	err := fn()

	// Record result
	cb.mu.Lock()
	cb.totalRequests++

	if err != nil {
		cb.totalFailures++
		cb.onFailure()
	} else {
		cb.totalSuccesses++
		cb.onSuccess()
	}
	cb.mu.Unlock()

	return err
}

// allowRequest checks if a request should be allowed
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true

	case StateOpen:
		// Check if timeout has elapsed
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.transitionTo(StateHalfOpen)
			return true
		}
		return false

	case StateHalfOpen:
		return true

	default:
		return false
	}
}

// onSuccess handles successful request
func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.transitionTo(StateClosed)
		}

	case StateClosed:
		// Reset failure count on success
		cb.failureCount = 0
	}
}

// onFailure handles failed request
func (cb *CircuitBreaker) onFailure() {
	cb.lastFailure = time.Now()

	switch cb.state {
	case StateClosed:
		cb.failureCount++
		if cb.failureCount >= cb.failureThreshold {
			cb.transitionTo(StateOpen)
		}

	case StateHalfOpen:
		// Any failure in half-open goes back to open
		cb.transitionTo(StateOpen)
	}
}

// transitionTo changes circuit breaker state
func (cb *CircuitBreaker) transitionTo(newState State) {
	if cb.state == newState {
		return
	}

	cb.state = newState
	cb.lastStateChange = time.Now()

	// Reset counters on state change
	cb.failureCount = 0
	cb.successCount = 0
}

// State returns current state
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	// Check if should transition from open to half-open
	if cb.state == StateOpen && time.Since(cb.lastFailure) > cb.timeout {
		cb.mu.RUnlock()
		cb.mu.Lock()
		if cb.state == StateOpen && time.Since(cb.lastFailure) > cb.timeout {
			cb.transitionTo(StateHalfOpen)
		}
		cb.mu.Unlock()
		cb.mu.RLock()
	}

	return cb.state
}

// Stats returns circuit breaker statistics
func (cb *CircuitBreaker) Stats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"name":              cb.name,
		"state":             cb.state.String(),
		"failure_count":     cb.failureCount,
		"success_count":     cb.successCount,
		"total_requests":    cb.totalRequests,
		"total_failures":    cb.totalFailures,
		"total_successes":   cb.totalSuccesses,
		"total_rejected":    cb.totalRejected,
		"failure_threshold": cb.failureThreshold,
		"success_threshold": cb.successThreshold,
		"timeout":           cb.timeout.String(),
		"last_failure":      cb.lastFailure,
		"last_state_change": cb.lastStateChange,
	}
}

// Reset resets circuit breaker to initial state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.transitionTo(StateClosed)
	cb.totalRequests = 0
	cb.totalFailures = 0
	cb.totalSuccesses = 0
	cb.totalRejected = 0
	cb.lastFailure = time.Time{}
}

// IsClosed returns true if circuit is closed
func (cb *CircuitBreaker) IsClosed() bool {
	return cb.State() == StateClosed
}

// IsOpen returns true if circuit is open
func (cb *CircuitBreaker) IsOpen() bool {
	return cb.State() == StateOpen
}

// Errors
var (
	ErrCircuitOpen = errors.New("circuit breaker is open")
)
