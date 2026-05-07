package circuitbreaker

import (
	"sync"
	"time"

	"github.com/presidendjakarta/setu-gateway/internal/logger"
)

// Manager manages multiple circuit breakers
type Manager struct {
	mu           sync.RWMutex
	breakers     map[string]*CircuitBreaker
	logger       *logger.Logger
}

// NewManager creates a new circuit breaker manager
func NewManager(log *logger.Logger) *Manager {
	return &Manager{
		breakers: make(map[string]*CircuitBreaker),
		logger:   log,
	}
}

// CreateBreaker creates a new circuit breaker
func (m *Manager) CreateBreaker(name string, failureThreshold int, timeout time.Duration) *CircuitBreaker {
	m.mu.Lock()
	defer m.mu.Unlock()

	cb := New(Config{
		Name:             name,
		FailureThreshold: failureThreshold,
		SuccessThreshold: 1,
		Timeout:          timeout,
	})

	m.breakers[name] = cb
	m.logger.Infow("Circuit breaker created",
		"name", name,
		"failure_threshold", failureThreshold,
		"timeout", timeout,
	)

	return cb
}

// GetBreaker returns a circuit breaker by name
func (m *Manager) GetBreaker(name string) (*CircuitBreaker, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cb, exists := m.breakers[name]
	return cb, exists
}

// GetOrCreateBreaker gets existing or creates new circuit breaker
func (m *Manager) GetOrCreateBreaker(name string, failureThreshold int, timeout time.Duration) *CircuitBreaker {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cb, exists := m.breakers[name]; exists {
		return cb
	}

	cb := New(Config{
		Name:             name,
		FailureThreshold: failureThreshold,
		SuccessThreshold: 1,
		Timeout:          timeout,
	})

	m.breakers[name] = cb
	m.logger.Infow("Circuit breaker created (auto)",
		"name", name,
		"failure_threshold", failureThreshold,
		"timeout", timeout,
	)

	return cb
}

// RemoveBreaker removes a circuit breaker
func (m *Manager) RemoveBreaker(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.breakers, name)
	m.logger.Infow("Circuit breaker removed", "name", name)
}

// ListBreakers returns all circuit breaker names
func (m *Manager) ListBreakers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.breakers))
	for name := range m.breakers {
		names = append(names, name)
	}

	return names
}

// Stats returns all circuit breaker statistics
func (m *Manager) Stats() map[string]map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]map[string]interface{})
	for name, cb := range m.breakers {
		stats[name] = cb.Stats()
	}

	return stats
}

// ResetAll resets all circuit breakers
func (m *Manager) ResetAll() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, cb := range m.breakers {
		cb.Reset()
		m.logger.Infow("Circuit breaker reset", "name", name)
	}
}

// HealthCheck returns health status of all circuit breakers
func (m *Manager) HealthCheck() map[string]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	health := make(map[string]bool)
	for name, cb := range m.breakers {
		health[name] = cb.IsClosed()
	}

	return health
}
