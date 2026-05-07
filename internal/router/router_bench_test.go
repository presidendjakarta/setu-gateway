package router

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/presidendjakarta/setu-gateway/pkg/types"
)

// Benchmark exact path matching
func BenchmarkRouter_ExactMatch(b *testing.B) {
	r := New()

	// Add 1000 routes
	for i := 0; i < 1000; i++ {
		route := &types.Route{
			ID:       fmt.Sprintf("route-%d", i),
			Name:     fmt.Sprintf("Route %d", i),
			Path:     fmt.Sprintf("/api/v1/resource-%d", i),
			PathType: types.PathTypeExact,
			Methods:  []string{"GET"},
			Enabled:  true,
		}
		r.AddRoute(context.Background(), route)
	}

	req, _ := http.NewRequest("GET", "/api/v1/resource-500", nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.Match(req)
	}
}

// Benchmark prefix matching
func BenchmarkRouter_PrefixMatch(b *testing.B) {
	r := New()

	// Add 500 prefix routes
	for i := 0; i < 500; i++ {
		route := &types.Route{
			ID:       fmt.Sprintf("prefix-%d", i),
			Name:     fmt.Sprintf("Prefix %d", i),
			Path:     fmt.Sprintf("/api/v%d/", i),
			PathType: types.PathTypePrefix,
			Methods:  []string{"GET", "POST"},
			Enabled:  true,
		}
		r.AddRoute(context.Background(), route)
	}

	req, _ := http.NewRequest("GET", "/api/v250/users/123", nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.Match(req)
	}
}

// Benchmark with many routes (10K)
func BenchmarkRouter_LargeRouteTable(b *testing.B) {
	r := New()

	// Add 10,000 routes
	for i := 0; i < 10000; i++ {
		route := &types.Route{
			ID:       fmt.Sprintf("route-%d", i),
			Name:     fmt.Sprintf("Route %d", i),
			Path:     fmt.Sprintf("/api/resources/%d/action", i),
			PathType: types.PathTypeExact,
			Methods:  []string{"GET"},
			Enabled:  true,
		}
		r.AddRoute(context.Background(), route)
	}

	req, _ := http.NewRequest("GET", "/api/resources/9999/action", nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.Match(req)
	}
}

// Benchmark concurrent reads
func BenchmarkRouter_ConcurrentReads(b *testing.B) {
	r := New()

	// Add routes
	for i := 0; i < 1000; i++ {
		route := &types.Route{
			ID:       fmt.Sprintf("route-%d", i),
			Path:     fmt.Sprintf("/api/v1/resource-%d", i),
			PathType: types.PathTypeExact,
			Methods:  []string{"GET"},
			Enabled:  true,
		}
		r.AddRoute(context.Background(), route)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/resource-%d", i%1000), nil)
			r.Match(req)
			i++
		}
	})
}

// Benchmark route addition
func BenchmarkRouter_AddRoute(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := New()
		route := &types.Route{
			ID:       "route",
			Path:     "/api/test",
			PathType: types.PathTypeExact,
			Methods:  []string{"GET"},
			Enabled:  true,
		}
		r.AddRoute(context.Background(), route)
	}
}

// Benchmark route update
func BenchmarkRouter_UpdateRoute(b *testing.B) {
	r := New()
	route := &types.Route{
		ID:       "route-1",
		Path:     "/api/v1",
		PathType: types.PathTypePrefix,
		Methods:  []string{"GET"},
		Enabled:  true,
	}
	r.AddRoute(context.Background(), route)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		route.Path = fmt.Sprintf("/api/v%d", i%100)
		r.UpdateRoute(context.Background(), route)
	}
}
