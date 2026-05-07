package loadbalancer

import (
	"context"
	"testing"

	"github.com/presidendjakarta/setu-gateway/pkg/types"
)

func TestRoundRobin_SingleTarget(t *testing.T) {
	lb := NewRoundRobin(&types.Upstream{
		ID:   "test-upstream",
		Name: "Test",
		Targets: []types.Target{
			{ID: "target-1", Address: "localhost:8001", Healthy: true},
		},
	})

	ctx := context.Background()

	// Should always return the same target
	for i := 0; i < 10; i++ {
		target, err := lb.Next(ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if target.ID != "target-1" {
			t.Errorf("Expected target-1, got %s", target.ID)
		}
	}
}

func TestRoundRobin_MultipleTargets(t *testing.T) {
	lb := NewRoundRobin(&types.Upstream{
		ID:   "test-upstream",
		Name: "Test",
		Targets: []types.Target{
			{ID: "target-1", Address: "localhost:8001", Healthy: true},
			{ID: "target-2", Address: "localhost:8002", Healthy: true},
			{ID: "target-3", Address: "localhost:8003", Healthy: true},
		},
	})

	ctx := context.Background()

	// Test round-robin distribution
	expected := []string{"target-1", "target-2", "target-3", "target-1", "target-2", "target-3"}

	for i, expectedID := range expected {
		target, err := lb.Next(ctx)
		if err != nil {
			t.Fatalf("Iteration %d: Unexpected error: %v", i, err)
		}

		if target.ID != expectedID {
			t.Errorf("Iteration %d: Expected %s, got %s", i, expectedID, target.ID)
		}
	}
}

func TestRoundRobin_SkipUnhealthy(t *testing.T) {
	lb := NewRoundRobin(&types.Upstream{
		ID:   "test-upstream",
		Name: "Test",
		Targets: []types.Target{
			{ID: "target-1", Address: "localhost:8001", Healthy: false},
			{ID: "target-2", Address: "localhost:8002", Healthy: true},
			{ID: "target-3", Address: "localhost:8003", Healthy: true},
		},
	})

	ctx := context.Background()

	// Should skip unhealthy target
	for i := 0; i < 10; i++ {
		target, err := lb.Next(ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if target.Healthy != true {
			t.Errorf("Expected healthy target, got unhealthy: %s", target.ID)
		}
	}
}

func TestRoundRobin_NoHealthyTargets(t *testing.T) {
	lb := NewRoundRobin(&types.Upstream{
		ID:   "test-upstream",
		Name: "Test",
		Targets: []types.Target{
			{ID: "target-1", Address: "localhost:8001", Healthy: false},
			{ID: "target-2", Address: "localhost:8002", Healthy: false},
		},
	})

	ctx := context.Background()

	// Should return error when no healthy targets
	_, err := lb.Next(ctx)
	if err == nil {
		t.Error("Expected error when no healthy targets, got nil")
	}
}

func TestRoundRobin_UpdateTargets(t *testing.T) {
	lb := NewRoundRobin(&types.Upstream{
		ID:   "test-upstream",
		Name: "Test",
		Targets: []types.Target{
			{ID: "target-1", Address: "localhost:8001", Healthy: true},
		},
	})

	ctx := context.Background()

	// Test initial target
	target, err := lb.Next(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if target.ID != "target-1" {
		t.Errorf("Expected target-1, got %s", target.ID)
	}

	// Update targets
	lb.UpdateTargets(&types.Upstream{
		ID:   "test-upstream",
		Name: "Test",
		Targets: []types.Target{
			{ID: "target-2", Address: "localhost:8002", Healthy: true},
			{ID: "target-3", Address: "localhost:8003", Healthy: true},
		},
	})

	// Should now use new targets
	target, err = lb.Next(ctx)
	if err != nil {
		t.Fatalf("Unexpected error after update: %v", err)
	}

	if target.ID != "target-2" && target.ID != "target-3" {
		t.Errorf("Expected target-2 or target-3, got %s", target.ID)
	}
}

func TestRoundRobin_LoadDistribution(t *testing.T) {
	lb := NewRoundRobin(&types.Upstream{
		ID:   "test-upstream",
		Name: "Test",
		Targets: []types.Target{
			{ID: "target-1", Address: "localhost:8001", Healthy: true},
			{ID: "target-2", Address: "localhost:8002", Healthy: true},
		},
	})

	ctx := context.Background()

	// Test even distribution over many requests
	counts := make(map[string]int)
	totalRequests := 1000

	for i := 0; i < totalRequests; i++ {
		target, _ := lb.Next(ctx)
		counts[target.ID]++
	}

	// Each target should get ~50% of requests
	for id, count := range counts {
		percentage := float64(count) / float64(totalRequests) * 100
		if percentage < 45 || percentage > 55 {
			t.Errorf("Target %s got %.1f%% of requests, expected ~50%%", id, percentage)
		}
	}
}

func TestRoundRobin_ConcurrentAccess(t *testing.T) {
	lb := NewRoundRobin(&types.Upstream{
		ID:   "test-upstream",
		Name: "Test",
		Targets: []types.Target{
			{ID: "target-1", Address: "localhost:8001", Healthy: true},
			{ID: "target-2", Address: "localhost:8002", Healthy: true},
			{ID: "target-3", Address: "localhost:8003", Healthy: true},
		},
	})

	ctx := context.Background()
	done := make(chan bool)

	// Multiple concurrent requests
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				target, err := lb.Next(ctx)
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if target == nil {
					t.Error("Expected target, got nil")
				}
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	t.Log("Concurrent access test passed")
}
