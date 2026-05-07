package loadbalancer

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/presidendjakarta/setu-gateway/pkg/types"
)

// RoundRobin implements round-robin load balancing
type RoundRobin struct {
	mu       sync.Mutex
	targets  []*types.Target
	current  int32
	counter  int64
}

// NewRoundRobin creates a new round-robin load balancer
func NewRoundRobin() *RoundRobin {
	return &RoundRobin{
		targets: make([]*types.Target, 0),
	}
}

// Next selects the next target using round-robin
func (rb *RoundRobin) Next(ctx context.Context) (*types.Target, error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if len(rb.targets) == 0 {
		return nil, types.ErrServiceUnavailable
	}

	// Filter healthy and enabled targets
	var healthyTargets []*types.Target
	for _, t := range rb.targets {
		if t.Enabled && t.Healthy {
			healthyTargets = append(healthyTargets, t)
		}
	}

	if len(healthyTargets) == 0 {
		return nil, types.ErrServiceUnavailable
	}

	// Round-robin selection
	idx := atomic.AddInt32(&rb.current, 1) % int32(len(healthyTargets))
	if idx < 0 {
		idx = 0
	}

	atomic.AddInt64(&rb.counter, 1)
	return healthyTargets[idx], nil
}

// RecordSuccess records a successful request
func (rb *RoundRobin) RecordSuccess(targetID string) {
	// Could track success metrics here
}

// RecordFailure records a failed request
func (rb *RoundRobin) RecordFailure(targetID string) {
	// Could track failure metrics here
}

// UpdateTargets updates the target list
func (rb *RoundRobin) UpdateTargets(targets []*types.Target) error {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.targets = make([]*types.Target, len(targets))
	copy(rb.targets, targets)

	return nil
}

// GetTargets returns current targets
func (rb *RoundRobin) GetTargets() []*types.Target {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	result := make([]*types.Target, len(rb.targets))
	copy(result, rb.targets)
	return result
}
