package admin

import (
	"encoding/json"
	"net/http"
	"strings"
)

// handleRateLimits handles GET /api/rate-limits and POST /api/rate-limits
func (h *Handler) handleRateLimits(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listRateLimits(w, r)
	case http.MethodPost:
		h.createRateLimit(w, r)
	default:
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleRateLimitByID handles GET/PUT/DELETE /api/rate-limits/{id}
func (h *Handler) handleRateLimitByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/rate-limits/")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "Rate limit ID is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getRateLimit(w, r, id)
	case http.MethodPut:
		h.updateRateLimit(w, r, id)
	case http.MethodDelete:
		h.deleteRateLimit(w, r, id)
	default:
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// List rate limits
func (h *Handler) listRateLimits(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limits, err := h.rateLimitRepo.List(ctx)
	if err != nil {
		h.logger.Errorw("Failed to list rate limits", "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to list rate limits")
		return
	}

	h.writeSuccess(w, limits, "Rate limits retrieved successfully")
}

// Get rate limit by ID
func (h *Handler) getRateLimit(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	limit, err := h.rateLimitRepo.GetByID(ctx, id)
	if err != nil {
		if err.Error() == "rate limit not found" {
			h.writeError(w, http.StatusNotFound, "Rate limit not found")
			return
		}
		h.logger.Errorw("Failed to get rate limit", "error", err, "id", id)
		h.writeError(w, http.StatusInternalServerError, "Failed to get rate limit")
		return
	}

	h.writeSuccess(w, limit, "Rate limit retrieved successfully")
}

// Create rate limit
func (h *Handler) createRateLimit(w http.ResponseWriter, r *http.Request) {
	var limit struct {
		Name     string `json:"name"`
		Requests int    `json:"requests"`
		Window   string `json:"window"`
		KeyBy    string `json:"key_by"`
	}

	if err := json.NewDecoder(r.Body).Decode(&limit); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if limit.Name == "" || limit.Requests == 0 {
		h.writeError(w, http.StatusBadRequest, "name and requests are required")
		return
	}

	h.logger.Infow("Rate limit creation requested", "name", limit.Name)
	h.writeError(w, http.StatusNotImplemented, "Rate limit creation not yet implemented")
}

// Update rate limit
func (h *Handler) updateRateLimit(w http.ResponseWriter, r *http.Request, id string) {
	h.logger.Infow("Rate limit update requested", "id", id)
	h.writeError(w, http.StatusNotImplemented, "Rate limit update not yet implemented")
}

// Delete rate limit
func (h *Handler) deleteRateLimit(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	err := h.rateLimitRepo.Delete(ctx, id)
	if err != nil {
		if err.Error() == "rate limit not found" {
			h.writeError(w, http.StatusNotFound, "Rate limit not found")
			return
		}
		h.logger.Errorw("Failed to delete rate limit", "error", err, "id", id)
		h.writeError(w, http.StatusInternalServerError, "Failed to delete rate limit")
		return
	}

	h.writeSuccess(w, nil, "Rate limit deleted successfully")
	h.logger.Infow("Rate limit deleted", "id", id)
}
