package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hlaclau/rate-it-api/internal/domain"
	"github.com/hlaclau/rate-it-api/internal/port"
	"github.com/hlaclau/rate-it-api/internal/usecase"
)

type ListHandler struct {
	uc port.ListUseCase
}

func NewListHandler(uc port.ListUseCase) *ListHandler {
	return &ListHandler{uc: uc}
}

func (h *ListHandler) Routes(r chi.Router) {
	r.With(RequireAuth).Get("/list", h.GetList)
	r.With(RequireAuth).Post("/list", h.AddOrUpdate)
	r.With(RequireAuth).Put("/list/{mediaID}", h.Update)
	r.With(RequireAuth).Delete("/list/{mediaID}", h.Remove)
	r.With(RequireAuth).Get("/list/status/{externalID}", h.GetStatus)
}

type addOrUpdateRequest struct {
	ExternalID string                 `json:"external_id"`
	Source     domain.MediaSource     `json:"source"`
	Type       domain.MediaType       `json:"type"`
	Status     domain.UserMediaStatus `json:"status"`
	Rating     *int16                 `json:"rating"`
	Review     *string                `json:"review"`
}

type updateRequest struct {
	Status domain.UserMediaStatus `json:"status"`
	Rating *int16                 `json:"rating"`
	Review *string                `json:"review"`
}

func (h *ListHandler) GetList(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)

	entries, err := h.uc.GetList(r.Context(), userID)
	if err != nil {
		h.writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func (h *ListHandler) AddOrUpdate(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)

	var req addOrUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid request"})
		return
	}

	if req.ExternalID == "" || req.Source == "" || req.Type == "" || req.Status == "" {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"message": "external_id, source, type and status are required"})
		return
	}

	params := port.AddOrUpdateParams{
		ExternalID: req.ExternalID,
		Source:     req.Source,
		Type:       req.Type,
		Status:     req.Status,
		Rating:     req.Rating,
		Review:     req.Review,
	}

	if err := h.uc.AddOrUpdate(r.Context(), userID, params); err != nil {
		h.writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ListHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)
	mediaID := chi.URLParam(r, "mediaID")

	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid request"})
		return
	}

	if req.Status == "" {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"message": "status is required"})
		return
	}

	if err := h.uc.Update(r.Context(), userID, mediaID, req.Status, req.Rating, req.Review); err != nil {
		h.writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ListHandler) Remove(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)
	mediaID := chi.URLParam(r, "mediaID")

	if err := h.uc.Remove(r.Context(), userID, mediaID); err != nil {
		h.writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ListHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)
	externalID := chi.URLParam(r, "externalID")
	source := r.URL.Query().Get("source")
	if source == "" {
		source = "tmdb"
	}

	entry, err := h.uc.GetStatus(r.Context(), userID, externalID, source)
	if err != nil {
		h.writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

func (h *ListHandler) writeError(w http.ResponseWriter, err error) {
	if errors.Is(err, port.ErrUserMediaNotFound) {
		h.writeJSON(w, http.StatusNotFound, map[string]string{"message": "not found"})
		return
	}
	if errors.Is(err, usecase.ErrInvalidRating) {
		h.writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
		return
	}
	if errors.Is(err, port.ErrNotFound) {
		h.writeJSON(w, http.StatusNotFound, map[string]string{"message": "media not found"})
		return
	}
	slog.Error("internal server error", "error", err)
	h.writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "internal server error"})
}

func (h *ListHandler) writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
