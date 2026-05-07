package observability

import (
	"net/http"
	"time"

	"github.com/presidendjakarta/setu-gateway/internal/middleware"
)

// MetricsMiddleware creates a middleware that records HTTP metrics
func MetricsMiddleware(metrics *Metrics) middleware.Middleware {
	return &metricsMiddleware{metrics: metrics}
}

type metricsMiddleware struct {
	metrics *Metrics
}

func (m *metricsMiddleware) Name() string {
	return "metrics"
}

func (m *metricsMiddleware) Priority() int {
	return 4 // After logger, before CORS
}

func (m *metricsMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Track active requests
		m.metrics.ActiveRequests.Inc()
		defer m.metrics.ActiveRequests.Dec()

		// Wrap response writer to capture status code and size
		rw := &metricsResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Process request
		next.ServeHTTP(rw, r)

		// Record metrics
		duration := time.Since(start)
		m.metrics.RecordHTTP(
			r.Method,
			r.URL.Path,
			"", // Route will be set later
			rw.statusCode,
			duration,
			int(r.ContentLength),
			rw.bytesWritten,
		)
	})
}

// metricsResponseWriter wraps http.ResponseWriter to capture metrics
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode    int
	bytesWritten  int
}

func (rw *metricsResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *metricsResponseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}
