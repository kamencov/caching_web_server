package post

import (
	"caching_web_server/internal/helper"
	"caching_web_server/internal/middleware"
	"caching_web_server/internal/models"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

//go:generate mockgen -source=handler.go -destination=handler_mock.go -package=post
type service interface {
	SaveDocument(ctx context.Context, login string, meta models.Meta, jsonData, file []byte) error
}

type Handler struct {
	service service
	log     *slog.Logger
	maxSize int64
}

func NewHandler(service service, log *slog.Logger, maxSize int64) *Handler {
	return &Handler{
		service: service,
		log:     log,
		maxSize: maxSize,
	}
}

func (h *Handler) SaveDocument(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxSize)
	err := r.ParseMultipartForm(h.maxSize)
	if err != nil {
		h.log.Error("SaveDocument", "failed to parse multipart form", err)
		helper.FailResponse(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	// meta
	var meta models.Meta
	metaStr := r.FormValue("meta")
	if metaStr == "" {
		h.log.Error("SaveDocument", "failed to get meta", err)
		helper.FailResponse(w, http.StatusBadRequest, "failed to get meta")
		return
	}
	if err := json.Unmarshal([]byte(metaStr), &meta); err != nil {
		h.log.Error("SaveDocument", "failed to unmarshal meta", err)
		helper.FailResponse(w, http.StatusBadRequest, "failed to unmarshal meta")
		return
	}

	// json
	var jsonData []byte
	if jsonStr := r.FormValue("json"); jsonStr != "" {
		jsonData = json.RawMessage(jsonStr)
	}

	// file
	file, _, err := r.FormFile("file")
	defer func() {
		if file != nil {
			err := file.Close()
			if err != nil {
				return
			}
		}
	}()
	if err != nil && !meta.File {
		h.log.Error("SaveDocument", "failed to get file", err)
		helper.FailResponse(w, http.StatusBadRequest, "failed to get file")
		return
	}

	var fileData []byte
	if file != nil {
		fileData, err = io.ReadAll(file)
		if err != nil {
			h.log.Error("SaveDocument", "failed to read file", err)
			helper.FailResponse(w, http.StatusBadRequest, "failed to read file")
			return
		}
	}

	login, ok := r.Context().Value(middleware.NameLogin).(string)
	if !ok {
		h.log.Error("SaveDocument", "failed to get login", err)
		helper.FailResponse(w, http.StatusInternalServerError, "failed to get login")
		return
	}

	err = h.service.SaveDocument(r.Context(), login, meta, jsonData, fileData)
	if err != nil {
		h.log.Error("SaveDocument", "failed to save document", err)
		helper.FailResponse(w, http.StatusInternalServerError, "failed to save document")
		return
	}

	respData := models.UploadResponse{}
	if jsonData != nil {
		respData.Data.JSON = jsonData
	}
	if meta.File {
		respData.Data.File = meta.Name
	}

	helper.OkDataResponse(w, respData)
}
