package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	authpkg "github.com/presidendjakarta/setu-gateway/internal/auth"
	authproviders "github.com/presidendjakarta/setu-gateway/internal/auth/providers"
	"github.com/presidendjakarta/setu-gateway/internal/circuitbreaker"
	"github.com/presidendjakarta/setu-gateway/internal/config"
	"github.com/presidendjakarta/setu-gateway/internal/loadbalancer"
	"github.com/presidendjakarta/setu-gateway/internal/logger"
	"github.com/presidendjakarta/setu-gateway/internal/middleware"
	"github.com/presidendjakarta/setu-gateway/internal/proxy"
	"github.com/presidendjakarta/setu-gateway/internal/ratelimiter"
	"github.com/presidendjakarta/setu-gateway/internal/router"
	"github.com/presidendjakarta/setu-gateway/pkg/types"
)

// Gateway is the main gateway handler
type Gateway struct {
	router        *router.Router
	proxy         *proxy.Proxy
	config        *config.RawConfig
	logger        *logger.Logger
	lbs           map[string]*loadbalancer.RoundRobin
	lbsMu         sync.RWMutex
	mwChain       *middleware.Chain
	authMgr       *authpkg.Manager
	rateLimiter   *ratelimiter.Manager
	circuitBreaker *circuitbreaker.Manager
}

// New creates a new gateway instance
func New(cfg *config.RawConfig, log *logger.Logger) (*Gateway, error) {
	// Create router
	r := router.New()

	// Create proxy
	p := proxy.New(&cfg.Proxy)

	// Create middleware chain
	mwChain := middleware.New()
	mwChain.Use(middleware.NewRecovery(log))
	mwChain.Use(middleware.NewRequestID())
	mwChain.Use(middleware.NewLogger(log))
	mwChain.Use(middleware.NewCORS())

	// Create auth manager
	authMgr := authpkg.NewManager(log)
	
	// Register default auth providers
	authMgr.RegisterProvider(authproviders.NewJWTProvider("jwt", types.AuthConfig{}, log))
	authMgr.RegisterProvider(authproviders.NewAPIKeyProvider("api_key", types.AuthConfig{}, log))
	
	// Create auth middleware
	authMiddleware := middleware.NewAuth(authMgr, log)
	mwChain.Use(authMiddleware)

	// Create rate limiter manager
	rateLimiterMgr := ratelimiter.NewManager(log)

	// Create circuit breaker manager
	circuitBreakerMgr := circuitbreaker.NewManager(log)

	return &Gateway{
		router:        r,
		proxy:         p,
		config:        cfg,
		logger:        log,
		lbs:           make(map[string]*loadbalancer.RoundRobin),
		mwChain:       mwChain,
		authMgr:       authMgr,
		rateLimiter:   rateLimiterMgr,
		circuitBreaker: circuitBreakerMgr,
	}, nil
}

// ServeHTTP handles incoming HTTP requests
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Build handler with middleware
	handler := g.mwChain.Build(http.HandlerFunc(g.handleRequest))
	handler.ServeHTTP(w, r)
}

// handleRequest is the main request handler (wrapped by middleware)
func (g *Gateway) handleRequest(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	
	// Get request ID from middleware
	requestID := middleware.GetRequestID(r.Context())
	if requestID == "" {
		requestID = uuid.New().String()
	}

	// Add request ID to context
	ctx := context.WithValue(r.Context(), "request_id", requestID)
	r = r.WithContext(ctx)

	// Set gateway header
	w.Header().Set("X-Gateway", "setu-gateway")

	// Health check endpoints (don't require routing)
	if r.URL.Path == g.config.Health.Path || 
	   r.URL.Path == g.config.Health.ReadinessPath || 
	   r.URL.Path == g.config.Health.LivenessPath {
		g.handleHealth(w, r)
		return
	}

	// Metrics endpoint
	if g.config.Metrics.Enabled && r.URL.Path == g.config.Metrics.Path {
		g.handleMetrics(w, r)
		return
	}

	// Match route
	route, err := g.router.Match(r)
	if err != nil {
		g.logger.Warnw("No route matched", 
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
		)
		g.writeError(w, types.ErrRouteNotFound)
		return
	}

	g.logger.Debugw("Route matched",
		"request_id", requestID,
		"route_id", route.ID,
		"route_name", route.Name,
	)

	// Get load balancer for upstream
	lb, err := g.getLoadBalancer(route.UpstreamID)
	if err != nil {
		g.logger.Errorw("Failed to get load balancer",
			"request_id", requestID,
			"upstream_id", route.UpstreamID,
			"error", err,
		)
		g.writeError(w, types.ErrServiceUnavailable)
		return
	}

	// Select target
	target, err := lb.Next(ctx)
	if err != nil {
		g.logger.Errorw("No available target",
			"request_id", requestID,
			"upstream_id", route.UpstreamID,
			"error", err,
		)
		g.writeError(w, types.ErrServiceUnavailable)
		return
	}

	// Apply timeout if configured
	if route.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, route.Timeout)
		defer cancel()
		r = r.WithContext(ctx)
	}

	// Proxy request
	err = g.proxy.ServeHTTP(ctx, w, r, route, target)
	
	duration := time.Since(startTime)

	if err != nil {
		g.logger.Errorw("Proxy error",
			"request_id", requestID,
			"route_id", route.ID,
			"target", fmt.Sprintf("%s:%d", target.Host, target.Port),
			"error", err,
			"duration_ms", duration.Milliseconds(),
		)
		lb.RecordFailure(target.ID)
		return
	}

	// Record success
	lb.RecordSuccess(target.ID)

	// Log access
	if g.config.Logging.AccessLog {
		g.logger.Infow("Request completed",
			"request_id", requestID,
			"route_id", route.ID,
			"route_name", route.Name,
			"method", r.Method,
			"path", r.URL.Path,
			"target", fmt.Sprintf("%s:%d", target.Host, target.Port),
			"duration_ms", duration.Milliseconds(),
		)
	}
}

// getLoadBalancer gets or creates a load balancer for an upstream
func (g *Gateway) getLoadBalancer(upstreamID string) (*loadbalancer.RoundRobin, error) {
	g.lbsMu.RLock()
	lb, exists := g.lbs[upstreamID]
	g.lbsMu.RUnlock()

	if exists {
		return lb, nil
	}

	g.lbsMu.Lock()
	defer g.lbsMu.Unlock()

	// Double-check after acquiring write lock
	if lb, exists = g.lbs[upstreamID]; exists {
		return lb, nil
	}

	// Create new load balancer
	lb = loadbalancer.NewRoundRobin()
	g.lbs[upstreamID] = lb

	return lb, nil
}

// handleHealth handles health check requests
func (g *Gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status": "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version": g.config.Gateway.Version,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(health)
}

// handleMetrics handles metrics requests (placeholder)
func (g *Gateway) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("# Metrics endpoint - Prometheus integration required"))
}

// writeError writes a structured error response
func (g *Gateway) writeError(w http.ResponseWriter, err error) {
	gwErr, ok := types.GetGatewayError(err)
	if !ok {
		gwErr = types.NewGatewayError(
			types.ErrCodeInternal,
			"Internal server error",
			http.StatusInternalServerError,
		)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(gwErr.StatusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   gwErr.Code,
		"message": gwErr.Message,
		"details": gwErr.Details,
	})
}

// ReloadRoutes reloads all routes from database
func (g *Gateway) ReloadRoutes(ctx context.Context, routes []*types.Route) error {
	return g.router.Reload(ctx, routes)
}

// UpdateTargets updates targets for an upstream
func (g *Gateway) UpdateTargets(upstreamID string, targets []*types.Target) error {
	lb, err := g.getLoadBalancer(upstreamID)
	if err != nil {
		return err
	}

	return lb.UpdateTargets(targets)
}

// Close closes the gateway and releases resources
func (g *Gateway) Close() error {
	return g.proxy.Close()
}
