package auth

import (
	"caching_web_server/internal/helper"
	"caching_web_server/internal/service/auth"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

//go:generate mockgen -source=handler.go -destination=handler_mock.go -package=auth
type service interface {
	RegisterUser(ctx context.Context, login, password string) error
	AuthUser(ctx context.Context, login, password string) (string, error)
}

type Handler struct {
	service    service
	log        *slog.Logger
	adminToken string
}

// NewHandler - конструктор
func NewHandler(service service, log *slog.Logger, adminToken string) *Handler {
	return &Handler{
		service:    service,
		log:        log,
		adminToken: adminToken,
	}
}

// Register - ручка регистрации пользователя
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.log.Error("Register", "error", "invalid method")
		helper.FailResponse(w, http.StatusMethodNotAllowed, "invalid method")
		return
	}

	var req struct {
		Token    string `json:"token"`
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("Register", "failed to decode request", err)
		helper.FailResponse(w, http.StatusBadRequest, "failed to decode request")
		return
	}

	if req.Token != h.adminToken {
		h.log.Error("Register", "error", "failed to invalid token")
		helper.FailResponse(w, http.StatusUnauthorized, "invalid token")
		return
	}

	err := h.service.RegisterUser(r.Context(), req.Login, req.Password)
	if err != nil {
		h.log.Error("Register", "failed to register user", err)
		helper.FailResponse(w, http.StatusInternalServerError, "failed to register user")
		return
	}

	helper.OkResponse(w, map[string]any{"login": req.Login})

}

// Auth - ручка авторизации
func (h *Handler) Auth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.log.Error("Auth", "error", "invalid method")
		helper.FailResponse(w, http.StatusMethodNotAllowed, "invalid method")
		return
	}
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("Auth", "failed to decode request", err)
		helper.FailResponse(w, http.StatusBadRequest, "failed to decode request")
		return
	}

	if req.Login == "" || req.Password == "" {
		h.log.Error("Auth", "error", "invalid login or password")
		helper.FailResponse(w, http.StatusBadRequest, "invalid login or password")
		return
	}

	token, err := h.service.AuthUser(r.Context(), req.Login, req.Password)
	if err != nil {
		h.log.Error("Auth", "error", err)
		helper.FailResponse(w, http.StatusInternalServerError, "failed to auth user")
		return
	}

	setTokenCookie(w, token)

	helper.OkResponse(w, map[string]any{auth.NameCookie: token})

}

func setTokenCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     auth.NameCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour),
	}
	http.SetCookie(w, cookie)
}

// Logout - ручка выхода
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.log.Error("Logout", "error", "invalid method")
		helper.FailResponse(w, http.StatusMethodNotAllowed, "invalid method")
		return
	}
	cookie, err := r.Cookie(auth.NameCookie)
	if err != nil {
		h.log.Error("Logout", "error", "missing or invalid Authorization header")
		helper.FailResponse(w, http.StatusUnauthorized, "missing or invalid Authorization header")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookie.Name,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	h.log.Info("Logout", "status", "cookie deleted")
	helper.OkResponse(w, map[string]string{"message": "logged out"})
}
