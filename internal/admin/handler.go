package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/presidendjakarta/setu-gateway/internal/logger"
	"github.com/presidendjakarta/setu-gateway/internal/repository"
)

// Handler manages admin API endpoints
type Handler struct {
	routeRepo     repository.RouteRepository
	upstreamRepo  repository.UpstreamRepository
	targetRepo    repository.TargetRepository
	rateLimitRepo repository.RateLimitRepository
	logger        *logger.Logger
	startedAt     time.Time
}

// NewHandler creates a new admin handler
func NewHandler(
	routeRepo repository.RouteRepository,
	upstreamRepo repository.UpstreamRepository,
	targetRepo repository.TargetRepository,
	rateLimitRepo repository.RateLimitRepository,
	log *logger.Logger,
) *Handler {
	return &Handler{
		routeRepo:     routeRepo,
		upstreamRepo:  upstreamRepo,
		targetRepo:    targetRepo,
		rateLimitRepo: rateLimitRepo,
		logger:        log,
		startedAt:     time.Now(),
	}
}

// RegisterRoutes registers admin API routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// API Info
	mux.HandleFunc("/api/", h.handleAPIInfo)

	// Route management
	mux.HandleFunc("/api/routes", h.handleRoutes)
	mux.HandleFunc("/api/routes/", h.handleRouteByID)

	// Upstream management
	mux.HandleFunc("/api/upstreams", h.handleUpstreams)
	mux.HandleFunc("/api/upstreams/", h.handleUpstreamByID)

	// Target management
	mux.HandleFunc("/api/upstreams/{id}/targets", h.handleTargets)
	mux.HandleFunc("/api/targets/", h.handleTargetByID)

	// Rate limit management
	mux.HandleFunc("/api/rate-limits", h.handleRateLimits)
	mux.HandleFunc("/api/rate-limits/", h.handleRateLimitByID)

	h.logger.Info("Admin API routes registered")
}

// Response helpers
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, APIResponse{
		Success: false,
		Error:   message,
	})
}

func (h *Handler) writeSuccess(w http.ResponseWriter, data interface{}, message string) {
	h.writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// API Info
func (h *Handler) handleAPIInfo(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/" {
		h.writeError(w, http.StatusNotFound, "Not found")
		return
	}

	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	info := map[string]interface{}{
		"name":        "Setu API Gateway Admin",
		"version":     "1.0.0",
		"uptime":      time.Since(h.startedAt).String(),
		"endpoints": map[string]string{
			"GET    /api/routes":           "List all routes",
			"POST   /api/routes":           "Create a new route",
			"GET    /api/routes/{id}":      "Get route by ID",
			"PUT    /api/routes/{id}":      "Update route",
			"DELETE /api/routes/{id}":      "Delete route",
			"GET    /api/upstreams":        "List all upstreams",
			"POST   /api/upstreams":        "Create a new upstream",
			"GET    /api/upstreams/{id}":   "Get upstream by ID",
			"PUT    /api/upstreams/{id}":   "Update upstream",
			"DELETE /api/upstreams/{id}":   "Delete upstream",
			"GET    /api/upstreams/{id}/targets": "List targets for upstream",
			"POST   /api/upstreams/{id}/targets": "Add target to upstream",
			"PUT    /api/targets/{id}":     "Update target",
			"DELETE /api/targets/{id}":     "Delete target",
			"GET    /api/rate-limits":      "List all rate limits",
			"POST   /api/rate-limits":      "Create rate limit",
			"PUT    /api/rate-limits/{id}": "Update rate limit",
			"DELETE /api/rate-limits/{id}": "Delete rate limit",
		},
	}

	h.writeSuccess(w, info, "Admin API is running")
}
