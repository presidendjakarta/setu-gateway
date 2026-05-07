package gateway

import (
	"context"
	"net/http"

	"github.com/presidendjakarta/setu-gateway/pkg/types"
)

// Router defines the interface for route matching and resolution
type Router interface {
	// Match finds the best matching route for the request
	Match(r *http.Request) (*types.Route, error)
	
	// AddRoute adds a new route to the router
	AddRoute(ctx context.Context, route *types.Route) error
	
	// RemoveRoute removes a route from the router
	RemoveRoute(ctx context.Context, routeID string) error
	
	// UpdateRoute updates an existing route
	UpdateRoute(ctx context.Context, route *types.Route) error
	
	// GetRoute returns a route by ID
	GetRoute(routeID string) (*types.Route, error)
	
	// ListRoutes returns all routes
	ListRoutes() []*types.Route
	
	// Reload reloads all routes
	Reload(ctx context.Context, routes []*types.Route) error
}

// Proxy defines the interface for reverse proxy
type Proxy interface {
	// ServeHTTP proxies the request to upstream
	ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request, route *types.Route, target *types.Target) error
	
	// Close closes the proxy and releases resources
	Close() error
}

// Middleware defines the interface for request middleware
type Middleware interface {
	// Name returns the middleware name
	Name() string
	
	// Process processes the request
	Process(ctx context.Context, w http.ResponseWriter, r *http.Request, next http.Handler) http.Handler
}

// MiddlewareChain defines the interface for middleware chain
type MiddlewareChain interface {
	// Add adds a middleware to the chain
	Add(mw Middleware)
	
	// AddBefore adds a middleware before another middleware
	AddBefore(name string, mw Middleware) error
	
	// AddAfter adds a middleware after another middleware
	AddAfter(name string, mw Middleware) error
	
	// Remove removes a middleware from the chain
	Remove(name string) error
	
	// Execute executes the middleware chain
	Execute(w http.ResponseWriter, r *http.Request, handler http.Handler)
}

// AuthProvider defines the interface for authentication providers
type AuthProvider interface {
	// Name returns the provider name
	Name() string
	
	// Type returns the provider type
	Type() types.AuthProviderType
	
	// Authenticate authenticates the request
	Authenticate(ctx context.Context, r *http.Request, config types.AuthConfig) (*types.AuthResult, error)
	
	// Validate validates the provider configuration
	Validate(config types.AuthConfig) error
}

// AuthManager defines the interface for authentication management
type AuthManager interface {
	// Authenticate authenticates a request using the auth chain
	Authenticate(ctx context.Context, r *http.Request, route *types.Route) (*types.AuthResult, error)
	
	// RegisterProvider registers an authentication provider
	RegisterProvider(provider AuthProvider) error
	
	// GetProvider returns a provider by type
	GetProvider(authType types.AuthProviderType) (AuthProvider, error)
}

// RateLimiter defines the interface for rate limiting
type RateLimiter interface {
	// Allow checks if a request is allowed
	Allow(ctx context.Context, key string, limit int, burst int) bool
	
	// GetRemaining returns remaining requests
	GetRemaining(ctx context.Context, key string) int
	
	// Reset resets the rate limiter for a key
	Reset(ctx context.Context, key string)
	
	// Close closes the rate limiter
	Close() error
}

// LoadBalancer defines the interface for load balancing
type LoadBalancer interface {
	// Next selects the next target
	Next(ctx context.Context) (*types.Target, error)
	
	// RecordSuccess records a successful request
	RecordSuccess(targetID string)
	
	// RecordFailure records a failed request
	RecordFailure(targetID string)
	
	// UpdateTargets updates the target list
	UpdateTargets(targets []*types.Target) error
	
	// GetTargets returns current targets
	GetTargets() []*types.Target
}

// CircuitBreaker defines the interface for circuit breaker
type CircuitBreaker interface {
	// AllowRequest checks if a request is allowed
	AllowRequest() bool
	
	// RecordSuccess records a successful request
	RecordSuccess()
	
	// RecordFailure records a failed request
	RecordFailure()
	
	// GetState returns the current state
	GetState() CircuitState
	
	// Reset resets the circuit breaker
	Reset()
}

// CircuitState represents circuit breaker state
type CircuitState string

const (
	StateClosed   CircuitState = "closed"
	StateOpen     CircuitState = "open"
	StateHalfOpen CircuitState = "half_open"
)

// ServiceDiscovery defines the interface for service discovery
type ServiceDiscovery interface {
	// Initialize initializes the discovery provider
	Initialize(ctx context.Context, config map[string]interface{}) error
	
	// GetTargets returns current targets
	GetTargets(ctx context.Context) ([]*types.Target, error)
	
	// Watch watches for target changes
	Watch(ctx context.Context) (<-chan []*types.Target, error)
	
	// Close closes the discovery provider
	Close() error
}

// Plugin defines the interface for gateway plugins
type Plugin interface {
	// Name returns the plugin name
	Name() string
	
	// Priority returns the plugin priority
	Priority() int
	
	// OnRequest is called before proxying the request
	OnRequest(ctx context.Context, r *http.Request) error
	
	// OnResponse is called after receiving the response
	OnResponse(ctx context.Context, w http.ResponseWriter, resp *http.Response) error
	
	// OnError is called when an error occurs
	OnError(ctx context.Context, err error)
}

// PluginManager defines the interface for plugin management
type PluginManager interface {
	// Register registers a plugin
	Register(plugin Plugin) error
	
	// Unregister unregisters a plugin
	Unregister(name string) error
	
	// ExecuteOnRequest executes all OnRequest hooks
	ExecuteOnRequest(ctx context.Context, r *http.Request) error
	
	// ExecuteOnResponse executes all OnResponse hooks
	ExecuteOnResponse(ctx context.Context, w http.ResponseWriter, resp *http.Response) error
	
	// ExecuteOnError executes all OnError hooks
	ExecuteOnError(ctx context.Context, err error)
	
	// GetPlugins returns all registered plugins
	GetPlugins() []Plugin
}

// HealthChecker defines the interface for health checking
type HealthChecker interface {
	// CheckHealth checks the health of the gateway
	CheckHealth() map[string]HealthStatus
	
	// CheckReadiness checks if the gateway is ready
	CheckReadiness() bool
	
	// CheckLiveness checks if the gateway is alive
	CheckLiveness() bool
	
	// AddCheck adds a health check
	AddCheck(name string, check func() bool)
}

// HealthStatus represents health check status
type HealthStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// MetricsCollector defines the interface for metrics collection
type MetricsCollector interface {
	// RecordRequest records a request metric
	RecordRequest(route string, method string, statusCode int, duration float64)
	
	// RecordAuth records an authentication metric
	RecordAuth(provider string, success bool, duration float64)
	
	// RecordRateLimit records a rate limit event
	RecordRateLimit(route string, blocked bool)
	
	// RecordCircuitBreaker records a circuit breaker event
	RecordCircuitBreaker(route string, state string)
	
	// RecordActiveConnections records active connections
	RecordActiveConnections(count int)
}
