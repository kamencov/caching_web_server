package delete

import (
	"caching_web_server/internal/helper"
	"caching_web_server/internal/middleware"
	"context"
	"log/slog"
	"net/http"
	"strings"
)

//go:generate mockgen -source=handler.go -destination=handler_mock.go -package=delete
type service interface {
	DeleteDocument(ctx context.Context, login, docID string) error
}

type Handler struct {
	service service
	log     *slog.Logger
}

func NewHandler(service service, log *slog.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

func (h *Handler) DeleteData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.log.Error("DeleteData", "error", "invalid method")
		helper.FailResponse(w, http.StatusMethodNotAllowed, "invalid method")
		return
	}

	login, ok := r.Context().Value(middleware.NameLogin).(string)
	if !ok {
		h.log.Error("DeleteData", "error", "failed to get login from context")
		helper.FailResponse(w, http.StatusInternalServerError, "failed to get login from context")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[3] == "" {
		h.log.Error("GetDocument", "error", "failed to get document id from path")
		helper.FailResponse(w, http.StatusBadRequest, "failed to get document id")
		return
	}
	docID := parts[3]

	if err := h.service.DeleteDocument(r.Context(), login, docID); err != nil {
		h.log.Error("DeleteData", "failed to delete document", err)
		helper.FailResponse(w, http.StatusInternalServerError, "failed to delete document")
		return
	}

	helper.OkResponse(w, map[string]bool{docID: true})
}
