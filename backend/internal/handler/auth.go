package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hlaclau/rate-it-api/internal/port"
)

type AuthHandler struct {
	uc port.AuthUseCase
}

func NewAuthHandler(uc port.AuthUseCase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

func (h *AuthHandler) Routes(r chi.Router) {
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)
	r.Post("/logout", h.Logout)
	r.Get("/me", h.Me)
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if msg := h.validateRegister(req); msg != "" {
		h.writeError(w, msg, http.StatusBadRequest)
		return
	}

	user, err := h.uc.Register(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, port.ErrUserAlreadyExists) {
			h.writeError(w, err.Error(), http.StatusConflict)
			return
		}
		h.writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	user.PasswordHash = ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if msg := h.validateLogin(req); msg != "" {
		h.writeError(w, msg, http.StatusBadRequest)
		return
	}

	access, refresh, user, err := h.uc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, port.ErrInvalidCredentials) {
			h.writeError(w, err.Error(), http.StatusUnauthorized)
			return
		}
		h.writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.setAuthCookies(w, access, refresh)

	user.PasswordHash = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		h.writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	access, userID, err := h.uc.Refresh(r.Context(), cookie.Value)
	if err != nil {
		h.writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	h.setAuthCookies(w, access, cookie.Value)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"user_id": userID})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(string)
	if ok {
		h.uc.Logout(r.Context(), userID)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		h.writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.uc.GetUserByID(r.Context(), userID)
	if err != nil {
		h.writeError(w, "user not found", http.StatusNotFound)
		return
	}

	user.PasswordHash = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) setAuthCookies(w http.ResponseWriter, access, refresh string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    access,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(15 * time.Minute),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refresh,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})
}

func (h *AuthHandler) validateRegister(req registerRequest) string {
	if strings.TrimSpace(req.Username) == "" {
		return "username is required"
	}
	if strings.TrimSpace(req.Email) == "" {
		return "email is required"
	}
	if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		return "invalid email format"
	}
	if len(req.Password) < 8 {
		return "password must be at least 8 characters"
	}
	return ""
}

func (h *AuthHandler) validateLogin(req loginRequest) string {
	if strings.TrimSpace(req.Email) == "" {
		return "email is required"
	}
	if strings.TrimSpace(req.Password) == "" {
		return "password is required"
	}
	return ""
}

func (h *AuthHandler) writeError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"message": msg})
}
