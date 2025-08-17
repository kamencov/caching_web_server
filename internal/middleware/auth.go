package middleware

import (
	"caching_web_server/internal/helper"
	"caching_web_server/internal/service/auth"
	"context"
	"log/slog"
	"net/http"
)

const NameLogin = "login"

//go:generate mockgen -source=auth.go -destination=auth_mock.go -package=middleware
type service interface {
	VerifyToken(auth string) (string, error)
}

type Middleware struct {
	service service
	log     *slog.Logger
}

func NewMiddleware(service service, log *slog.Logger) *Middleware {
	return &Middleware{
		service: service,
		log:     log,
	}
}

func (m *Middleware) Authorize(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(auth.NameCookie)
		if err != nil {
			m.log.Error("Authorize", "error", "missing or invalid Authorization header")
			helper.FailResponse(w, http.StatusUnauthorized, "missing or invalid Authorization header")
			return
		}

		token := cookie.Value

		login, err := m.service.VerifyToken(token)
		if err != nil {
			m.log.Error("Authorize", "error", err.Error())
			helper.FailResponse(w, http.StatusUnauthorized, "Authorize")
			return
		}

		ctx := context.WithValue(r.Context(), NameLogin, login)

		next.ServeHTTP(w, r.WithContext(ctx))
	})

	return fn
}
