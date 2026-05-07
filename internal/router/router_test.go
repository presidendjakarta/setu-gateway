package router

import (
	"context"
	"net/http"
	"testing"

	"github.com/presidendjakarta/setu-gateway/pkg/types"
)

func TestRouter_ExactMatch(t *testing.T) {
	r := New()
	ctx := context.Background()

	// Add exact route
	route := &types.Route{
		ID:       "route-1",
		Name:     "Test Route",
		Path:     "/api/users",
		PathType: types.PathTypeExact,
		Methods:  []string{"GET"},
		Enabled:  true,
	}

	err := r.AddRoute(ctx, route)
	if err != nil {
		t.Fatalf("Failed to add route: %v", err)
	}

	// Test exact match
	req, _ := http.NewRequest("GET", "/api/users", nil)
	matched, err := r.Match(req)
	if err != nil {
		t.Fatalf("Expected match, got error: %v", err)
	}

	if matched.ID != "route-1" {
		t.Errorf("Expected route-1, got %s", matched.ID)
	}

	// Test no match (different path)
	req2, _ := http.NewRequest("GET", "/api/users/123", nil)
	_, err = r.Match(req2)
	if err == nil {
		t.Error("Expected error for non-matching path, got nil")
	}
}

func TestRouter_PrefixMatch(t *testing.T) {
	r := New()

	// Add prefix route
	route := &types.Route{
		ID:       "route-2",
		Name:     "API Route",
		Path:     "/api/",
		PathType: types.PathTypePrefix,
		Methods:  []string{"GET", "POST"},
		Enabled:  true,
	}

	err := r.AddRoute(context.Background(), route)
	if err != nil {
		t.Fatalf("Failed to add route: %v", err)
	}

	// Test prefix match
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"exact prefix", "/api/", "route-2"},
		{"nested path", "/api/users", "route-2"},
		{"deep nested", "/api/users/123/posts", "route-2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.path, nil)
			matched, err := r.Match(req)
			if err != nil {
				t.Fatalf("Expected match for %s, got error: %v", tt.path, err)
			}

			if matched.ID != tt.expected {
				t.Errorf("Expected %s for %s, got %s", tt.expected, tt.path, matched.ID)
			}
		})
	}
}

func TestRouter_Priority(t *testing.T) {
	r := New()

	// Add routes with different priorities
	route1 := &types.Route{
		ID:       "low-priority",
		Name:     "Low Priority",
		Path:     "/api/",
		PathType: types.PathTypePrefix,
		Priority: 1,
		Methods:  []string{"GET"},
		Enabled:  true,
	}

	route2 := &types.Route{
		ID:       "high-priority",
		Name:     "High Priority",
		Path:     "/api/users",
		PathType: types.PathTypeExact,
		Priority: 10,
		Methods:  []string{"GET"},
		Enabled:  true,
	}

	r.AddRoute(context.Background(), route1)
	r.AddRoute(context.Background(), route2)

	// Higher priority should match first
	req, _ := http.NewRequest("GET", "/api/users", nil)
	matched, err := r.Match(req)
	if err != nil {
		t.Fatalf("Expected match, got error: %v", err)
	}

	if matched.ID != "high-priority" {
		t.Errorf("Expected high-priority route, got %s", matched.ID)
	}
}

func TestRouter_MethodMatching(t *testing.T) {
	r := New()

	// Add route with specific methods
	route := &types.Route{
		ID:       "route-get",
		Name:     "GET Only",
		Path:     "/api/data",
		PathType: types.PathTypeExact,
		Methods:  []string{"GET"},
		Enabled:  true,
	}

	r.AddRoute(context.Background(), route)

	// Test GET (should match)
	reqGet, _ := http.NewRequest("GET", "/api/data", nil)
	_, err := r.Match(reqGet)
	if err != nil {
		t.Errorf("Expected GET to match, got error: %v", err)
	}

	// Test POST (should not match)
	reqPost, _ := http.NewRequest("POST", "/api/data", nil)
	_, err = r.Match(reqPost)
	if err == nil {
		t.Error("Expected POST to fail, got match")
	}
}

func TestRouter_DisabledRoute(t *testing.T) {
	r := New()

	// Add disabled route
	route := &types.Route{
		ID:       "disabled",
		Name:     "Disabled Route",
		Path:     "/api/test",
		PathType: types.PathTypeExact,
		Methods:  []string{"GET"},
		Enabled:  false,
	}

	r.AddRoute(context.Background(), route)

	// Should not match disabled route
	req, _ := http.NewRequest("GET", "/api/test", nil)
	_, err := r.Match(req)
	if err == nil {
		t.Error("Expected disabled route to not match")
	}
}

func TestRouter_UpdateRoute(t *testing.T) {
	r := New()

	route := &types.Route{
		ID:       "route-1",
		Name:     "Original",
		Path:     "/api/v1",
		PathType: types.PathTypePrefix,
		Methods:  []string{"GET"},
		Enabled:  true,
	}

	r.AddRoute(context.Background(), route)

	// Update route
	route.Name = "Updated"
	route.Path = "/api/v2"
	r.UpdateRoute(context.Background(), route)

	// Test updated route
	req, _ := http.NewRequest("GET", "/api/v2", nil)
	matched, err := r.Match(req)
	if err != nil {
		t.Fatalf("Expected match after update, got error: %v", err)
	}

	if matched.Name != "Updated" {
		t.Errorf("Expected updated name, got %s", matched.Name)
	}

	// Old path should not match
	reqOld, _ := http.NewRequest("GET", "/api/v1", nil)
	_, err = r.Match(reqOld)
	if err == nil {
		t.Error("Expected old path to not match after update")
	}
}

func TestRouter_RemoveRoute(t *testing.T) {
	r := New()

	route := &types.Route{
		ID:       "to-remove",
		Name:     "Remove Me",
		Path:     "/api/remove",
		PathType: types.PathTypeExact,
		Methods:  []string{"GET"},
		Enabled:  true,
	}

	r.AddRoute(context.Background(), route)

	// Verify route exists
	req, _ := http.NewRequest("GET", "/api/remove", nil)
	_, err := r.Match(req)
	if err != nil {
		t.Fatalf("Expected route to exist: %v", err)
	}

	// Remove route
	r.RemoveRoute(context.Background(), "to-remove")

	// Verify route removed
	_, err = r.Match(req)
	if err == nil {
		t.Error("Expected route to be removed")
	}
}

func TestRouter_WildcardMatch(t *testing.T) {
	r := New()

	// Add wildcard route
	route := &types.Route{
		ID:       "wildcard",
		Name:     "Wildcard Route",
		Path:     "*",
		PathType: types.PathTypeWildcard,
		Methods:  []string{"GET"},
		Enabled:  true,
	}

	r.AddRoute(context.Background(), route)

	// Test various paths
	paths := []string{"/", "/api", "/api/users", "/anything"}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			req, _ := http.NewRequest("GET", path, nil)
			matched, err := r.Match(req)
			if err != nil {
				t.Fatalf("Expected wildcard match for %s, got error: %v", path, err)
			}

			if matched.ID != "wildcard" {
				t.Errorf("Expected wildcard route for %s, got %s", path, matched.ID)
			}
		})
	}
}

func TestRouter_ConcurrentAccess(t *testing.T) {
	r := New()

	// Add initial route
	route := &types.Route{
		ID:       "concurrent-1",
		Name:     "Concurrent Test",
		Path:     "/api/test",
		PathType: types.PathTypeExact,
		Methods:  []string{"GET"},
		Enabled:  true,
	}

	r.AddRoute(context.Background(), route)

	// Concurrent reads and writes
	done := make(chan bool)

	// Goroutine 1: Read
	go func() {
		for i := 0; i < 100; i++ {
			req, _ := http.NewRequest("GET", "/api/test", nil)
			r.Match(req)
		}
		done <- true
	}()

	// Goroutine 2: Write
	go func() {
		for i := 0; i < 100; i++ {
			newRoute := &types.Route{
				ID:       "concurrent-new",
				Name:     "New Route",
				Path:     "/api/new",
				PathType: types.PathTypeExact,
				Methods:  []string{"GET"},
				Enabled:  true,
			}
			r.AddRoute(context.Background(), newRoute)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// If we reach here without panic, test passes
	t.Log("Concurrent access test passed")
}
