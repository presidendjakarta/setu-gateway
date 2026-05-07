package middleware

import (
	"net/http"
	"strings"
)

// CORS middleware handles Cross-Origin Resource Sharing
type CORS struct {
	allowedOrigins []string
	allowedMethods []string
	allowedHeaders []string
	maxAge         int
}

// NewCORS creates a new CORS middleware
func NewCORS(opts ...CORSOption) *CORS {
	c := &CORS{
		allowedOrigins: []string{"*"},
		allowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		allowedHeaders: []string{"*"},
		maxAge:         86400, // 24 hours
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// CORSOption is a function that configures CORS
type CORSOption func(*CORS)

// WithOrigins sets allowed origins
func WithOrigins(origins ...string) CORSOption {
	return func(c *CORS) {
		c.allowedOrigins = origins
	}
}

// WithMethods sets allowed methods
func WithMethods(methods ...string) CORSOption {
	return func(c *CORS) {
		c.allowedMethods = methods
	}
}

// WithHeaders sets allowed headers
func WithHeaders(headers ...string) CORSOption {
	return func(c *CORS) {
		c.allowedHeaders = headers
	}
}

// Name returns the middleware name
func (c *CORS) Name() string {
	return "cors"
}

// Priority returns the middleware priority
func (c *CORS) Priority() int {
	return 10
}

// Handler implements Middleware interface
func (c *CORS) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		origin := req.Header.Get("Origin")

		// Check if origin is allowed
		if c.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(c.allowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(c.allowedHeaders, ", "))
			w.Header().Set("Access-Control-Max-Age", string(rune(c.maxAge)))
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight request
		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, req)
	})
}

// isOriginAllowed checks if origin is allowed
func (c *CORS) isOriginAllowed(origin string) bool {
	if len(c.allowedOrigins) == 0 {
		return false
	}

	// Wildcard allows all
	if c.allowedOrigins[0] == "*" {
		return true
	}

	// Check specific origins
	for _, allowed := range c.allowedOrigins {
		if origin == allowed {
			return true
		}
	}

	return false
}
