package observability

import (
	"context"
	"net/http"
	"time"

	"github.com/presidendjakarta/setu-gateway/internal/middleware"
)

// Context keys for route information
type contextKey string

const (
	RouteIDKey   contextKey = "route_id"
	RouteNameKey contextKey = "route_name"
)

// SetRouteInfo sets route information in context
func SetRouteInfo(ctx context.Context, routeID, routeName string) context.Context {
	ctx = context.WithValue(ctx, RouteIDKey, routeID)
	ctx = context.WithValue(ctx, RouteNameKey, routeName)
	return ctx
}

// GetRouteInfo gets route information from context
func GetRouteInfo(ctx context.Context) (string, string) {
	routeID, _ := ctx.Value(RouteIDKey).(string)
	routeName, _ := ctx.Value(RouteNameKey).(string)
	return routeID, routeName
}

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

		// Get route info from context
		_, routeName := GetRouteInfo(r.Context())

		// Record metrics
		duration := time.Since(start)
		m.metrics.RecordHTTP(
			r.Method,
			r.URL.Path,
			routeName, // Use route name from context
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
