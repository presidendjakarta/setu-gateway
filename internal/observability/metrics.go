package observability

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the gateway
type Metrics struct {
	// HTTP Request metrics
	HttpRequestsTotal   *prometheus.CounterVec
	HttpRequestDuration *prometheus.HistogramVec
	HttpRequestSize     *prometheus.HistogramVec
	HttpResponseSize    *prometheus.HistogramVec
	ActiveRequests      prometheus.Gauge

	// Router metrics
	RouteMatches *prometheus.CounterVec

	// Authentication metrics
	AuthAttempts  *prometheus.CounterVec
	AuthFailures  *prometheus.CounterVec
	AuthDuration  *prometheus.HistogramVec

	// Rate limiting metrics
	RateLimitAttempts *prometheus.CounterVec
	RateLimitRejections *prometheus.CounterVec

	// Load balancer metrics
	LBRequests     *prometheus.CounterVec
	LBTargetHealth *prometheus.GaugeVec

	// Circuit breaker metrics
	CircuitBreakerState *prometheus.GaugeVec
	CircuitBreakerTrips *prometheus.CounterVec

	// Proxy metrics
	ProxyErrors    *prometheus.CounterVec
	ProxyLatency   *prometheus.HistogramVec
	UpstreamLatency *prometheus.HistogramVec

	// System metrics
	ConfigReloads      prometheus.Counter
	ConfigReloadErrors prometheus.Counter
	Goroutines         prometheus.Gauge
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics() *Metrics {
	namespace := "setu_gateway"

	m := &Metrics{
		// HTTP metrics
		HttpRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status", "route"},
		),

		HttpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "path", "route"},
		),

		HttpRequestSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_size_bytes",
				Help:      "HTTP request size in bytes",
				Buckets:   []float64{100, 1000, 10000, 100000, 1000000},
			},
			[]string{"method", "path"},
		),

		HttpResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_response_size_bytes",
				Help:      "HTTP response size in bytes",
				Buckets:   []float64{100, 1000, 10000, 100000, 1000000},
			},
			[]string{"method", "path", "status"},
		),

		ActiveRequests: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "active_requests",
				Help:      "Number of active HTTP requests",
			},
		),

		// Router metrics
		RouteMatches: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "route_matches_total",
				Help:      "Total number of route matches",
			},
			[]string{"route_id", "route_name"},
		),

		// Authentication metrics
		AuthAttempts: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "auth_attempts_total",
				Help:      "Total number of authentication attempts",
			},
			[]string{"provider", "status"},
		),

		AuthFailures: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "auth_failures_total",
				Help:      "Total number of authentication failures",
			},
			[]string{"provider", "reason"},
		),

		AuthDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "auth_duration_seconds",
				Help:      "Authentication duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"provider"},
		),

		// Rate limiting metrics
		RateLimitAttempts: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "ratelimit_attempts_total",
				Help:      "Total number of rate limit checks",
			},
			[]string{"limiter_id"},
		),

		RateLimitRejections: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "ratelimit_rejections_total",
				Help:      "Total number of rate limit rejections",
			},
			[]string{"limiter_id"},
		),

		// Load balancer metrics
		LBRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "lb_requests_total",
				Help:      "Total number of load balancer requests per target",
			},
			[]string{"upstream_id", "target_id"},
		),

		LBTargetHealth: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "lb_target_health",
				Help:      "Load balancer target health status (1=healthy, 0=unhealthy)",
			},
			[]string{"upstream_id", "target_id", "address"},
		),

		// Circuit breaker metrics
		CircuitBreakerState: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "circuit_breaker_state",
				Help:      "Circuit breaker state (0=closed, 1=open, 2=half-open)",
			},
			[]string{"name"},
		),

		CircuitBreakerTrips: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "circuit_breaker_trips_total",
				Help:      "Total number of circuit breaker trips (closed to open)",
			},
			[]string{"name"},
		),

		// Proxy metrics
		ProxyErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "proxy_errors_total",
				Help:      "Total number of proxy errors",
			},
			[]string{"upstream_id", "error_type"},
		),

		ProxyLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "proxy_latency_seconds",
				Help:      "Proxy latency in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"upstream_id"},
		),

		UpstreamLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "upstream_latency_seconds",
				Help:      "Upstream service latency in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"upstream_id", "target"},
		),

		// System metrics
		ConfigReloads: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "config_reloads_total",
				Help:      "Total number of configuration reloads",
			},
		),

		ConfigReloadErrors: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "config_reload_errors_total",
				Help:      "Total number of configuration reload errors",
			},
		),

		Goroutines: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "goroutines",
				Help:      "Number of running goroutines",
			},
		),
	}

	return m
}

// RecordHTTP records HTTP request metrics
func (m *Metrics) RecordHTTP(method, path, route string, status int, duration time.Duration, requestSize, responseSize int) {
	statusStr := http.StatusText(status)
	if statusStr == "" {
		statusStr = "unknown"
	}

	m.HttpRequestsTotal.WithLabelValues(method, path, statusStr, route).Inc()
	m.HttpRequestDuration.WithLabelValues(method, path, route).Observe(duration.Seconds())
	m.HttpResponseSize.WithLabelValues(method, path, statusStr).Observe(float64(responseSize))

	if requestSize > 0 {
		m.HttpRequestSize.WithLabelValues(method, path).Observe(float64(requestSize))
	}
}

// RecordRouteMatch records a route match
func (m *Metrics) RecordRouteMatch(routeID, routeName string) {
	m.RouteMatches.WithLabelValues(routeID, routeName).Inc()
}

// RecordAuth records authentication metrics
func (m *Metrics) RecordAuth(provider string, success bool, duration time.Duration, reason string) {
	if success {
		m.AuthAttempts.WithLabelValues(provider, "success").Inc()
	} else {
		m.AuthAttempts.WithLabelValues(provider, "failure").Inc()
		m.AuthFailures.WithLabelValues(provider, reason).Inc()
	}

	m.AuthDuration.WithLabelValues(provider).Observe(duration.Seconds())
}

// RecordRateLimit records rate limit check
func (m *Metrics) RecordRateLimit(limiterID string, allowed bool) {
	m.RateLimitAttempts.WithLabelValues(limiterID).Inc()
	if !allowed {
		m.RateLimitRejections.WithLabelValues(limiterID).Inc()
	}
}

// RecordLBRequest records load balancer request
func (m *Metrics) RecordLBRequest(upstreamID, targetID string) {
	m.LBRequests.WithLabelValues(upstreamID, targetID).Inc()
}

// RecordTargetHealth records target health status
func (m *Metrics) RecordTargetHealth(upstreamID, targetID, address string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	m.LBTargetHealth.WithLabelValues(upstreamID, targetID, address).Set(value)
}

// RecordCircuitBreakerState records circuit breaker state
func (m *Metrics) RecordCircuitBreakerState(name string, state int) {
	m.CircuitBreakerState.WithLabelValues(name).Set(float64(state))
}

// RecordCircuitBreakerTrip records circuit breaker trip
func (m *Metrics) RecordCircuitBreakerTrip(name string) {
	m.CircuitBreakerTrips.WithLabelValues(name).Inc()
}

// RecordProxyError records proxy error
func (m *Metrics) RecordProxyError(upstreamID, errorType string) {
	m.ProxyErrors.WithLabelValues(upstreamID, errorType).Inc()
}

// RecordProxyLatency records proxy latency
func (m *Metrics) RecordProxyLatency(upstreamID string, duration time.Duration) {
	m.ProxyLatency.WithLabelValues(upstreamID).Observe(duration.Seconds())
}

// RecordUpstreamLatency records upstream service latency
func (m *Metrics) RecordUpstreamLatency(upstreamID, target string, duration time.Duration) {
	m.UpstreamLatency.WithLabelValues(upstreamID, target).Observe(duration.Seconds())
}

// UpdateGoroutines updates goroutine count
func (m *Metrics) UpdateGoroutines(count int) {
	m.Goroutines.Set(float64(count))
}
