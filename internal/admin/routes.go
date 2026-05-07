package admin

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// handleRoutes handles GET /api/routes and POST /api/routes
func (h *Handler) handleRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listRoutes(w, r)
	case http.MethodPost:
		h.createRoute(w, r)
	default:
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleRouteByID handles GET/PUT/DELETE /api/routes/{id}
func (h *Handler) handleRouteByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/routes/")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "Route ID is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getRoute(w, r, id)
	case http.MethodPut:
		h.updateRoute(w, r, id)
	case http.MethodDelete:
		h.deleteRoute(w, r, id)
	default:
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// List routes
func (h *Handler) listRoutes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	routes, err := h.routeRepo.ListAll(ctx)
	if err != nil {
		h.logger.Errorw("Failed to list routes", "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to list routes")
		return
	}

	h.writeSuccess(w, routes, "Routes retrieved successfully")
}

// Get route by ID
func (h *Handler) getRoute(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	route, err := h.routeRepo.GetByID(ctx, id)
	if err != nil {
		if err.Error() == "route not found" {
			h.writeError(w, http.StatusNotFound, "Route not found")
			return
		}
		h.logger.Errorw("Failed to get route", "error", err, "id", id)
		h.writeError(w, http.StatusInternalServerError, "Failed to get route")
		return
	}

	h.writeSuccess(w, route, "Route retrieved successfully")
}

// Create route
func (h *Handler) createRoute(w http.ResponseWriter, r *http.Request) {
	var route struct {
		Name           string   `json:"name"`
		Description    string   `json:"description"`
		Path           string   `json:"path"`
		PathType       string   `json:"path_type"`
		Methods        []string `json:"methods"`
		StripPath      bool     `json:"strip_path"`
		PreserveHost   bool     `json:"preserve_host"`
		Enabled        bool     `json:"enabled"`
		Priority       int      `json:"priority"`
		UpstreamID     string   `json:"upstream_id"`
		AuthChain      []string `json:"auth_chain,omitempty"`
		Plugins        []string `json:"plugins,omitempty"`
		RateLimitID    *string  `json:"rate_limit_id,omitempty"`
		Timeout        int      `json:"timeout"`
		RetryEnabled   bool     `json:"retry_enabled"`
		CircuitBreaker string   `json:"circuit_breaker,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if route.Name == "" || route.Path == "" || route.UpstreamID == "" {
		h.writeError(w, http.StatusBadRequest, "name, path, and upstream_id are required")
		return
	}

	h.logger.Infow("Route creation requested", "name", route.Name, "path", route.Path)
	h.writeError(w, http.StatusNotImplemented, "Route creation not yet implemented - requires repository update")
}

// Update route
func (h *Handler) updateRoute(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	// Get existing route
	existing, err := h.routeRepo.GetByID(ctx, id)
	if err != nil {
		if err.Error() == "route not found" {
			h.writeError(w, http.StatusNotFound, "Route not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "Failed to get route")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	h.logger.Infow("Route update requested", "id", id, "updates", updates)
	h.writeSuccess(w, existing, "Route update not yet implemented - requires repository update")
}

// Delete route
func (h *Handler) deleteRoute(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	err := h.routeRepo.Delete(ctx, id)
	if err != nil {
		if err.Error() == "route not found" {
			h.writeError(w, http.StatusNotFound, "Route not found")
			return
		}
		h.logger.Errorw("Failed to delete route", "error", err, "id", id)
		h.writeError(w, http.StatusInternalServerError, "Failed to delete route")
		return
	}

	h.writeSuccess(w, nil, "Route deleted successfully")
	h.logger.Infow("Route deleted", "id", id)
}

func generateID() string {
	return uuid.New().String()
}
