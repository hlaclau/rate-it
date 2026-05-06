package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hlaclau/rate-it-api/internal/domain"
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

var validSortBy = map[string]bool{
	domain.SortByPopularityDesc:  true,
	domain.SortByPopularityAsc:   true,
	domain.SortByVoteAverageDesc: true,
	domain.SortByVoteAverageAsc:  true,
	domain.SortByReleaseDateDesc: true,
	domain.SortByReleaseDateAsc:  true,
}

func (h *MediaHandler) SearchMedia(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page := 1
	if raw := q.Get("page"); raw != "" {
		var err error
		page, err = strconv.Atoi(raw)
		if err != nil || page < 1 || page > 1000 {
			http.Error(w, `{"error":"page must be an integer between 1 and 1000"}`, http.StatusBadRequest)
			return
		}
	}

	var yearFrom, yearTo int
	if raw := q.Get("year_from"); raw != "" {
		var err error
		yearFrom, err = strconv.Atoi(raw)
		if err != nil {
			http.Error(w, `{"error":"year_from must be a valid year"}`, http.StatusBadRequest)
			return
		}
	}
	if raw := q.Get("year_to"); raw != "" {
		var err error
		yearTo, err = strconv.Atoi(raw)
		if err != nil {
			http.Error(w, `{"error":"year_to must be a valid year"}`, http.StatusBadRequest)
			return
		}
	}

	var voteMin, voteMax float64
	if raw := q.Get("vote_average_min"); raw != "" {
		var err error
		voteMin, err = strconv.ParseFloat(raw, 32)
		if err != nil || voteMin < 0 || voteMin > 10 {
			http.Error(w, `{"error":"vote_average_min must be a number between 0 and 10"}`, http.StatusBadRequest)
			return
		}
	}
	if raw := q.Get("vote_average_max"); raw != "" {
		var err error
		voteMax, err = strconv.ParseFloat(raw, 32)
		if err != nil || voteMax < 0 || voteMax > 10 {
			http.Error(w, `{"error":"vote_average_max must be a number between 0 and 10"}`, http.StatusBadRequest)
			return
		}
	}

	sortBy := q.Get("sort_by")
	if sortBy != "" && !validSortBy[sortBy] {
		http.Error(w, `{"error":"invalid sort_by value"}`, http.StatusBadRequest)
		return
	}

	var voteCountMin int
	if raw := q.Get("vote_count_min"); raw != "" {
		var err error
		voteCountMin, err = strconv.Atoi(raw)
		if err != nil || voteCountMin < 0 {
			http.Error(w, `{"error":"vote_count_min must be a non-negative integer"}`, http.StatusBadRequest)
			return
		}
	}

	params := domain.MediaParams{
		Query:          q.Get("q"),
		IncludeAdult:   q.Get("include_adult") == "true",
		Type:           q.Get("type"),
		Language:       q.Get("language"),
		Page:           page,
		SortBy:         sortBy,
		YearFrom:       yearFrom,
		YearTo:         yearTo,
		VoteAverageMin: float32(voteMin),
		VoteAverageMax: float32(voteMax),
		VoteCountMin:   voteCountMin,
		WithGenres:     q.Get("with_genres"),
		WatchProviders: q.Get("watch_providers"),
		WatchRegion:    q.Get("watch_region"),
	}

	raw, err := h.uc.SearchMedia(r.Context(), params)
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
	slog.Error("internal server error", "error", err)
	http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
}
