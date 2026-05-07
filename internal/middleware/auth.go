package middleware

import (
	"net/http"

	"github.com/presidendjakarta/setu-gateway/internal/auth"
	"github.com/presidendjakarta/setu-gateway/internal/logger"
	"github.com/presidendjakarta/setu-gateway/pkg/types"
)

// Auth middleware handles authentication
type Auth struct {
	authManager *auth.Manager
	logger      *logger.Logger
}

// NewAuth creates a new auth middleware
func NewAuth(mgr *auth.Manager, log *logger.Logger) *Auth {
	return &Auth{
		authManager: mgr,
		logger:      log,
	}
}

// Name returns the middleware name
func (a *Auth) Name() string {
	return "auth"
}

// Priority returns the middleware priority
func (a *Auth) Priority() int {
	return 20
}

// Handler implements Middleware interface
func (a *Auth) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For now, pass through all requests
		// Auth will be applied per-route in the gateway handler
		// This middleware is for global auth if needed

		// TODO: Implement global auth check if configured
		// if global auth required:
		//     result, err := a.authManager.Authenticate(r.Context(), r, route)
		//     if err != nil {
		//         writeAuthError(w, err)
		//         return
		//     }
		//     // Add auth result to context
		//     ctx := context.WithValue(r.Context(), types.AuthResultKey{}, result)
		//     r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// writeAuthError writes authentication error response
func writeAuthError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"AUTH_FAILED","message":"Authentication required"}`))
}

// GetAuthResult gets auth result from context
func GetAuthResult(r *http.Request) *types.AuthResult {
	if result, ok := r.Context().Value(types.AuthResultKey{}).(*types.AuthResult); ok {
		return result
	}
	return nil
}
