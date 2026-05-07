package middleware

import (
	"net/http"
	"time"

	"github.com/presidendjakarta/setu-gateway/internal/logger"
)

// Logger middleware logs HTTP requests
type Logger struct {
	logger *logger.Logger
}

// NewLogger creates a new logger middleware
func NewLogger(log *logger.Logger) *Logger {
	return &Logger{
		logger: log,
	}
}

// Name returns the middleware name
func (l *Logger) Name() string {
	return "logger"
}

// Priority returns the middleware priority
func (l *Logger) Priority() int {
	return 3
}

// Handler implements Middleware interface
func (l *Logger) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Process request
		next.ServeHTTP(rw, req)

		// Log request
		duration := time.Since(start)
		requestID := GetRequestID(req.Context())

		l.logger.Infow("HTTP Request",
			"request_id", requestID,
			"method", req.Method,
			"path", req.URL.Path,
			"status", rw.statusCode,
			"duration_ms", duration.Milliseconds(),
			"remote_addr", req.RemoteAddr,
			"user_agent", req.UserAgent(),
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
