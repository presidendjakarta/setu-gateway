package admin

import (
	"encoding/json"
	"net/http"
	"strings"
)

// handleUpstreams handles GET /api/upstreams and POST /api/upstreams
func (h *Handler) handleUpstreams(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listUpstreams(w, r)
	case http.MethodPost:
		h.createUpstream(w, r)
	default:
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleUpstreamByID handles GET/PUT/DELETE /api/upstreams/{id}
func (h *Handler) handleUpstreamByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/upstreams/")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "Upstream ID is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getUpstream(w, r, id)
	case http.MethodPut:
		h.updateUpstream(w, r, id)
	case http.MethodDelete:
		h.deleteUpstream(w, r, id)
	default:
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// List upstreams
func (h *Handler) listUpstreams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	upstreams, err := h.upstreamRepo.ListAll(ctx)
	if err != nil {
		h.logger.Errorw("Failed to list upstreams", "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to list upstreams")
		return
	}

	h.writeSuccess(w, upstreams, "Upstreams retrieved successfully")
}

// Get upstream by ID
func (h *Handler) getUpstream(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	upstream, err := h.upstreamRepo.GetByID(ctx, id)
	if err != nil {
		if err.Error() == "upstream not found" {
			h.writeError(w, http.StatusNotFound, "Upstream not found")
			return
		}
		h.logger.Errorw("Failed to get upstream", "error", err, "id", id)
		h.writeError(w, http.StatusInternalServerError, "Failed to get upstream")
		return
	}

	h.writeSuccess(w, upstream, "Upstream retrieved successfully")
}

// Create upstream
func (h *Handler) createUpstream(w http.ResponseWriter, r *http.Request) {
	var upstream struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Algorithm   string `json:"algorithm"`
		Enabled     bool   `json:"enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&upstream); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if upstream.Name == "" {
		h.writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	h.logger.Infow("Upstream creation requested", "name", upstream.Name)
	h.writeError(w, http.StatusNotImplemented, "Upstream creation not yet implemented")
}

// Update upstream
func (h *Handler) updateUpstream(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	existing, err := h.upstreamRepo.GetByID(ctx, id)
	if err != nil {
		if err.Error() == "upstream not found" {
			h.writeError(w, http.StatusNotFound, "Upstream not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "Failed to get upstream")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	h.logger.Infow("Upstream update requested", "id", id, "updates", updates)
	h.writeSuccess(w, existing, "Upstream update not yet implemented")
}

// Delete upstream
func (h *Handler) deleteUpstream(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	err := h.upstreamRepo.Delete(ctx, id)
	if err != nil {
		if err.Error() == "upstream not found" {
			h.writeError(w, http.StatusNotFound, "Upstream not found")
			return
		}
		h.logger.Errorw("Failed to delete upstream", "error", err, "id", id)
		h.writeError(w, http.StatusInternalServerError, "Failed to delete upstream")
		return
	}

	h.writeSuccess(w, nil, "Upstream deleted successfully")
	h.logger.Infow("Upstream deleted", "id", id)
}

// Target management
func (h *Handler) handleTargets(w http.ResponseWriter, r *http.Request) {
	upstreamID := r.PathValue("id")
	if upstreamID == "" {
		h.writeError(w, http.StatusBadRequest, "Upstream ID is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.listTargets(w, r, upstreamID)
	case http.MethodPost:
		h.createTarget(w, r, upstreamID)
	default:
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) handleTargetByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/targets/")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "Target ID is required")
		return
	}

	switch r.Method {
	case http.MethodPut:
		h.updateTarget(w, r, id)
	case http.MethodDelete:
		h.deleteTarget(w, r, id)
	default:
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// List targets
func (h *Handler) listTargets(w http.ResponseWriter, r *http.Request, upstreamID string) {
	ctx := r.Context()

	targets, err := h.targetRepo.GetByUpstreamID(ctx, upstreamID)
	if err != nil {
		h.logger.Errorw("Failed to list targets", "error", err, "upstream_id", upstreamID)
		h.writeError(w, http.StatusInternalServerError, "Failed to list targets")
		return
	}

	h.writeSuccess(w, targets, "Targets retrieved successfully")
}

// Create target
func (h *Handler) createTarget(w http.ResponseWriter, r *http.Request, upstreamID string) {
	var target struct {
		Host   string `json:"host"`
		Port   int    `json:"port"`
		Weight int    `json:"weight"`
	}

	if err := json.NewDecoder(r.Body).Decode(&target); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if target.Host == "" || target.Port == 0 {
		h.writeError(w, http.StatusBadRequest, "host and port are required")
		return
	}

	h.logger.Infow("Target creation requested", "upstream_id", upstreamID, "host", target.Host, "port", target.Port)
	h.writeError(w, http.StatusNotImplemented, "Target creation not yet implemented")
}

// Update target
func (h *Handler) updateTarget(w http.ResponseWriter, r *http.Request, id string) {
	h.logger.Infow("Target update requested", "id", id)
	h.writeError(w, http.StatusNotImplemented, "Target update not yet implemented")
}

// Delete target
func (h *Handler) deleteTarget(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	err := h.targetRepo.Delete(ctx, id)
	if err != nil {
		if err.Error() == "target not found" {
			h.writeError(w, http.StatusNotFound, "Target not found")
			return
		}
		h.logger.Errorw("Failed to delete target", "error", err, "id", id)
		h.writeError(w, http.StatusInternalServerError, "Failed to delete target")
		return
	}

	h.writeSuccess(w, nil, "Target deleted successfully")
	h.logger.Infow("Target deleted", "id", id)
}
