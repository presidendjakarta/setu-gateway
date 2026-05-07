package middleware

import (
	"context"
	"net/http"
)

// Middleware defines the interface for HTTP middleware
type Middleware interface {
	// Name returns the middleware name
	Name() string
	
	// Priority returns the middleware priority (lower = higher priority)
	Priority() int
	
	// Handler returns the HTTP handler
	Handler(next http.Handler) http.Handler
}

// MiddlewareFunc is an adapter to use ordinary functions as Middleware
type MiddlewareFunc func(next http.Handler) http.Handler

// Handler implements Middleware interface for MiddlewareFunc
func (fn MiddlewareFunc) Handler(next http.Handler) http.Handler {
	return fn(next)
}

// Name returns empty name for MiddlewareFunc
func (fn MiddlewareFunc) Name() string {
	return "anonymous"
}

// Priority returns default priority for MiddlewareFunc
func (fn MiddlewareFunc) Priority() int {
	return 100
}

// Chain represents a chain of middlewares
type Chain struct {
	middlewares []Middleware
}

// New creates a new middleware chain
func New() *Chain {
	return &Chain{
		middlewares: make([]Middleware, 0),
	}
}

// Use adds a middleware to the chain
func (c *Chain) Use(mw Middleware) {
	c.middlewares = append(c.middlewares, mw)
}

// UseFunc adds a middleware function to the chain
func (c *Chain) UseFunc(fn MiddlewareFunc) {
	c.Use(mwAdapter{fn: fn, name: fn.Name(), priority: fn.Priority()})
}

// Build builds the middleware chain
func (c *Chain) Build(final http.Handler) http.Handler {
	// Sort by priority
	sortMiddlewares(c.middlewares)
	
	// Build chain from last to first
	handler := final
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		handler = c.middlewares[i].Handler(handler)
	}
	
	return handler
}

// Handler returns the final handler with all middlewares applied
func (c *Chain) Handler(final http.Handler) http.Handler {
	return c.Build(final)
}

// sortMiddlewares sorts middlewares by priority
func sortMiddlewares(mws []Middleware) {
	for i := 0; i < len(mws); i++ {
		for j := i + 1; j < len(mws); j++ {
			if mws[i].Priority() > mws[j].Priority() {
				mws[i], mws[j] = mws[j], mws[i]
			}
		}
	}
}

// mwAdapter adapts MiddlewareFunc to Middleware interface
type mwAdapter struct {
	fn       MiddlewareFunc
	name     string
	priority int
}

func (a mwAdapter) Name() string {
	return a.name
}

func (a mwAdapter) Priority() int {
	return a.priority
}

func (a mwAdapter) Handler(next http.Handler) http.Handler {
	return a.fn(next)
}

// RequestContextKey is the context key for request context
type RequestContextKey struct{}

// GetRequestID gets request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey{}).(string); ok {
		return id
	}
	return ""
}

// RequestIDKey is the context key for request ID
type RequestIDKey struct{}
