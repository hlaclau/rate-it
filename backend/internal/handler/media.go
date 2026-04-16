package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hlaclau/rate-it-api/internal/port"
	"github.com/hlaclau/rate-it-api/internal/usecase"
)

type MediaHandler struct {
	uc *usecase.MediaUseCase
}

func NewMediaHandler(uc *usecase.MediaUseCase) *MediaHandler {
	return &MediaHandler{uc: uc}
}

func (h *MediaHandler) Routes(r chi.Router) {
	r.Get("/media/movie/{id}", h.GetMovie)
	r.Get("/media/series/{id}", h.GetSeries)
	r.Get("/media/search", h.SearchMedia)
}

func (h *MediaHandler) GetMovie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	raw, err := h.uc.GetMovie(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func (h *MediaHandler) GetSeries(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	raw, err := h.uc.GetSeries(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func (h *MediaHandler) SearchMedia(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	raw, err := h.uc.SearchMedia(r.Context(), q)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func writeError(w http.ResponseWriter, err error) {
	if errors.Is(err, port.ErrNotFound) {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
}
