package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// RequestID middleware adds unique request ID
type RequestID struct {
	header string
}

// NewRequestID creates a new request ID middleware
func NewRequestID() *RequestID {
	return &RequestID{
		header: "X-Request-ID",
	}
}

// Name returns the middleware name
func (r *RequestID) Name() string {
	return "request_id"
}

// Priority returns the middleware priority
func (r *RequestID) Priority() int {
	return 2
}

// Handler implements Middleware interface
func (r *RequestID) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Get request ID from header or generate new one
		requestID := req.Header.Get(r.header)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add to response header
		w.Header().Set(r.header, requestID)

		// Add to request context
		ctx := context.WithValue(req.Context(), RequestIDKey{}, requestID)
		req = req.WithContext(ctx)

		next.ServeHTTP(w, req)
	})
}
