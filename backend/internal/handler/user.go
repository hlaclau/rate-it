package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hlaclau/rate-it-api/internal/port"
)

type UserHandler struct {
	uc     port.UserUseCase
	listUC port.ListUseCase
}

func NewUserHandler(uc port.UserUseCase, listUC port.ListUseCase) *UserHandler {
	return &UserHandler{uc: uc, listUC: listUC}
}

func (h *UserHandler) Routes(r chi.Router) {
	r.Get("/users/search", h.Search)
	r.Get("/users/{username}/list", h.GetList)
}

type userSummary struct {
	ID        string  `json:"id"`
	Username  string  `json:"username"`
	AvatarURL *string `json:"avatar_url"`
}

func (h *UserHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if len(q) < 2 {
		h.writeJSON(w, http.StatusOK, []userSummary{})
		return
	}

	users, err := h.uc.SearchUsers(r.Context(), q)
	if err != nil {
		slog.Error("search users", "error", err)
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "internal server error"})
		return
	}

	summaries := make([]userSummary, len(users))
	for i, u := range users {
		summaries[i] = userSummary{ID: u.ID, Username: u.Username, AvatarURL: u.AvatarURL}
	}
	h.writeJSON(w, http.StatusOK, summaries)
}

func (h *UserHandler) GetList(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	user, err := h.uc.GetByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, port.ErrUserNotFound) {
			h.writeJSON(w, http.StatusNotFound, map[string]string{"message": "user not found"})
			return
		}
		slog.Error("get user by username", "error", err)
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "internal server error"})
		return
	}

	entries, err := h.listUC.GetList(r.Context(), user.ID)
	if err != nil {
		slog.Error("get user list", "error", err)
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "internal server error"})
		return
	}

	if entries == nil {
		entries = []port.ListEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func (h *UserHandler) writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
