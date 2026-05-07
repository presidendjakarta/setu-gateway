package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/presidendjakarta/setu-gateway/internal/logger"
)

// Recovery middleware recovers from panics
type Recovery struct {
	logger *logger.Logger
}

// NewRecovery creates a new recovery middleware
func NewRecovery(log *logger.Logger) *Recovery {
	return &Recovery{
		logger: log,
	}
}

// Name returns the middleware name
func (r *Recovery) Name() string {
	return "recovery"
}

// Priority returns the middleware priority (should be first)
func (r *Recovery) Priority() int {
	return 1
}

// Handler implements Middleware interface
func (r *Recovery) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				r.logger.Errorw("Panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
					"method", req.Method,
					"path", req.URL.Path,
					"remote_addr", req.RemoteAddr,
				)

				// Return 500 Internal Server Error
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"INTERNAL_ERROR","message":"Internal server error"}`))
			}
		}()

		next.ServeHTTP(w, req)
	})
}
