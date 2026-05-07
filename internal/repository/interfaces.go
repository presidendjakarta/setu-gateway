package repository

import (
	"context"

	"github.com/presidendjakarta/setu-gateway/pkg/types"
)

// RouteRepository defines the interface for route data access
type RouteRepository interface {
	// Create creates a new route
	Create(ctx context.Context, route *types.Route) error
	
	// GetByID retrieves a route by ID
	GetByID(ctx context.Context, id string) (*types.Route, error)
	
	// GetByPath retrieves routes by path pattern
	GetByPath(ctx context.Context, path string) ([]*types.Route, error)
	
	// Update updates an existing route
	Update(ctx context.Context, route *types.Route) error
	
	// Delete deletes a route by ID
	Delete(ctx context.Context, id string) error
	
	// List retrieves all enabled routes
	List(ctx context.Context) ([]*types.Route, error)
	
	// ListAll retrieves all routes (including disabled)
	ListAll(ctx context.Context) ([]*types.Route, error)
	
	// Count returns the total number of routes
	Count(ctx context.Context) (int64, error)
}

// UpstreamRepository defines the interface for upstream data access
type UpstreamRepository interface {
	// Create creates a new upstream
	Create(ctx context.Context, upstream *types.Upstream) error
	
	// GetByID retrieves an upstream by ID
	GetByID(ctx context.Context, id string) (*types.Upstream, error)
	
	// GetByName retrieves an upstream by name
	GetByName(ctx context.Context, name string) (*types.Upstream, error)
	
	// Update updates an existing upstream
	Update(ctx context.Context, upstream *types.Upstream) error
	
	// Delete deletes an upstream by ID
	Delete(ctx context.Context, id string) error
	
	// List retrieves all enabled upstreams
	List(ctx context.Context) ([]*types.Upstream, error)
	
	// ListAll retrieves all upstreams
	ListAll(ctx context.Context) ([]*types.Upstream, error)
}

// TargetRepository defines the interface for target data access
type TargetRepository interface {
	// Create creates a new target
	Create(ctx context.Context, target *types.Target) error
	
	// GetByID retrieves a target by ID
	GetByID(ctx context.Context, id string) (*types.Target, error)
	
	// GetByUpstreamID retrieves all targets for an upstream
	GetByUpstreamID(ctx context.Context, upstreamID string) ([]*types.Target, error)
	
	// Update updates an existing target
	Update(ctx context.Context, target *types.Target) error
	
	// Delete deletes a target by ID
	Delete(ctx context.Context, id string) error
	
	// UpdateHealth updates target health status
	UpdateHealth(ctx context.Context, id string, healthy bool) error
}

// AuthRepository defines the interface for auth provider data access
type AuthRepository interface {
	// Create creates a new auth provider
	Create(ctx context.Context, provider *types.AuthProvider) error
	
	// GetByID retrieves an auth provider by ID
	GetByID(ctx context.Context, id string) (*types.AuthProvider, error)
	
	// GetByName retrieves an auth provider by name
	GetByName(ctx context.Context, name string) (*types.AuthProvider, error)
	
	// Update updates an existing auth provider
	Update(ctx context.Context, provider *types.AuthProvider) error
	
	// Delete deletes an auth provider by ID
	Delete(ctx context.Context, id string) error
	
	// List retrieves all enabled auth providers
	List(ctx context.Context) ([]*types.AuthProvider, error)
	
	// ListAll retrieves all auth providers
	ListAll(ctx context.Context) ([]*types.AuthProvider, error)
	
	// GetRouteAuthProviders retrieves auth providers for a route
	GetRouteAuthProviders(ctx context.Context, routeID string) ([]*types.AuthProvider, error)
}

// PluginRepository defines the interface for plugin data access
type PluginRepository interface {
	// Create creates a new plugin
	Create(ctx context.Context, plugin *types.Plugin) error
	
	// GetByID retrieves a plugin by ID
	GetByID(ctx context.Context, id string) (*types.Plugin, error)
	
	// Update updates an existing plugin
	Update(ctx context.Context, plugin *types.Plugin) error
	
	// Delete deletes a plugin by ID
	Delete(ctx context.Context, id string) error
	
	// List retrieves all enabled plugins
	List(ctx context.Context) ([]*types.Plugin, error)
	
	// ListByRoute retrieves plugins for a specific route
	ListByRoute(ctx context.Context, routeID string) ([]*types.Plugin, error)
	
	// ListGlobal retrieves global plugins (not route-specific)
	ListGlobal(ctx context.Context) ([]*types.Plugin, error)
}

// RateLimitRepository defines the interface for rate limit data access
type RateLimitRepository interface {
	// Create creates a new rate limit config
	Create(ctx context.Context, limit *types.RateLimit) error
	
	// GetByID retrieves a rate limit config by ID
	GetByID(ctx context.Context, id string) (*types.RateLimit, error)
	
	// GetByRouteID retrieves rate limit config for a route
	GetByRouteID(ctx context.Context, routeID string) (*types.RateLimit, error)
	
	// Update updates an existing rate limit config
	Update(ctx context.Context, limit *types.RateLimit) error
	
	// Delete deletes a rate limit config by ID
	Delete(ctx context.Context, id string) error
	
	// List retrieves all rate limit configs
	List(ctx context.Context) ([]*types.RateLimit, error)
}
