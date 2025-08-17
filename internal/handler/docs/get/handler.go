package get

import (
	"caching_web_server/internal/helper"
	"caching_web_server/internal/models"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

//go:generate mockgen -source=handler.go -destination=handler_mock.go -package=get
type Service interface {
	GetDocuments(ctx context.Context, login, filterKey, filterValue string, limit int) ([]models.DocsData, error)
	GetDocument(ctx context.Context, login, docID string) ([]byte, []byte, string, error)
}

type Handler struct {
	service Service
	log     *slog.Logger
}

func NewHandler(service Service, log *slog.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// GetDocuments - ручка получения документов
func (h *Handler) GetDocuments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.log.Error("GetDocuments", "error", "invalid method")
		helper.FailResponse(w, http.StatusMethodNotAllowed, "invalid method")
		return
	}

	var req struct {
		Token       string `json:"token"`
		Login       string `json:"login"`
		FilterKey   string `json:"key" `
		FilterValue string `json:"value"`
		Limit       int    `json:"limit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("GetDocuments", "failed to decode request", err)
		helper.FailResponse(w, http.StatusBadRequest, "failed to decode request")
		return
	}

	if req.Login == "" {
		login, ok := r.Context().Value("login").(string)
		if !ok {
			h.log.Error("GetDocuments", "error", "failed to get login from context")
			helper.FailResponse(w, http.StatusInternalServerError, "failed to get login from context")
			return
		}
		req.Login = login
	}

	docs, err := h.service.GetDocuments(r.Context(), req.Login, req.FilterKey, req.FilterValue, req.Limit)
	if err != nil {
		h.log.Error("GetDocuments", "failed to get documents", err)
		helper.FailResponse(w, http.StatusInternalServerError, "failed to get documents")
		return
	}

	helper.OkDataResponse(w, docs)
}

// GetDocument - ручка получения документа
func (h *Handler) GetDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.log.Error("GetDocument", "error", "invalid method")
		helper.FailResponse(w, http.StatusMethodNotAllowed, "invalid method")
		return
	}

	// Извлекаем id из пути: /api/docs/{id}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[3] == "" {
		h.log.Error("GetDocument", "error", "failed to get document id from path")
		helper.FailResponse(w, http.StatusBadRequest, "failed to get document id")
		return
	}
	docID := parts[3]

	login, ok := r.Context().Value("login").(string)
	if !ok {
		h.log.Error("GetDocument", "error", "failed to get login from context")
		helper.FailResponse(w, http.StatusInternalServerError, "failed to get login from context")
		return
	}

	file, JSON, mime, err := h.service.GetDocument(r.Context(), login, docID)
	if err != nil {
		h.log.Error("GetDocument", "failed to get document", err)
		helper.FailResponse(w, http.StatusInternalServerError, "failed to get document")
		return
	}

	if file != nil {
		w.Header().Set("Content-Type", mime)
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(file)
		if err != nil {
			h.log.Error("GetDocument", "failed to write file", err)
			helper.FailResponse(w, http.StatusInternalServerError, "failed to write file")
			return
		}
		return
	}

	helper.OkDataResponse(w, JSON)
}
