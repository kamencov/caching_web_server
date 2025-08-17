package docs

import (
	"caching_web_server/internal/models"
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/google/uuid"
)

//go:generate mockgen -source=service.go -destination=service_mock.go -package=docs
type storage interface {
	GetUserID(ctx context.Context, login string) (int, error)
	SaveDocument(ctx context.Context, doc *models.Document, grants []string) error
	GetDocuments(ctx context.Context, login, filterKey, filterValue string, limit int) ([]models.DocsData, error)
	DeleteDocument(ctx context.Context, login string, id uuid.UUID) error
	GetDocumentByID(ctx context.Context, docID uuid.UUID, login string) (*models.Document, error)
}

type s3 interface {
	SaveFile(ctx context.Context,key string, data []byte, contentType string) (string, error)
	DeleteFile(key string) error
	GetFile(key string) ([]byte, error)
}

type Service struct {
	storage storage
	s3      s3
	log     *slog.Logger
}

// NewService - создает новый сервис
func NewService(storage storage, s3 s3, log *slog.Logger) *Service {
	return &Service{
		storage: storage,
		s3:      s3,
		log:     log,
	}
}

// SaveDocument сохраняет документ
func (s *Service) SaveDocument(ctx context.Context, login string, meta models.Meta, jsonData, file []byte) error {
	// положи в MINIO
	ext := filepath.Ext(meta.Name)
	key := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	_, err := s.s3.SaveFile(ctx,key, file, meta.Mime)
	if err != nil {
		s.log.Error("SaveDocument", "failed to save file", err)
		return err
	}

	// получаем userID
	userID, err := s.storage.GetUserID(ctx, login)
	if err != nil {
		s.log.Error("SaveDocument", "failed to get user id", err)
		return err
	}

	// создаем запрос
	doc := s.createDocument(meta, key, jsonData, userID)

	// сохрани в БД
	err = s.storage.SaveDocument(ctx, doc, meta.Grants)
	if err != nil {
		s.log.Error("SaveDocument", "failed to save document", err)
		errs3 := s.s3.DeleteFile(key)
		if errs3 != nil {
			return errs3
		}
		return err
	}

	return nil
}

// Создаем единый документ
func (s *Service) createDocument(meta models.Meta, path string, jsonData []byte, userID int) *models.Document {
	var doc models.Document

	doc.Name = meta.Name
	doc.OwnerID = int64(userID)
	doc.Mime = meta.Mime
	doc.HashFile = meta.File
	doc.Public = meta.Public
	doc.JsonDate = jsonData
	doc.StoragePath = path

	return &doc
}

// GetDocuments - возвращает список документов
func (s *Service) GetDocuments(ctx context.Context, login, filterKey, filterValue string, limit int) ([]models.DocsData, error) {
	allowedKeys := map[string]bool{
		"name": true,
		"mime": true,
	}

	key := ""
	if allowedKeys[filterKey] {
		key = filterKey
	}

	docs, err := s.storage.GetDocuments(ctx, login, key, filterValue, limit)
	if err != nil {
		s.log.Error("GetDocuments", "failed to get documents", err)
		return nil, err
	}

	return docs, nil

}

// GetDocument - возвращает документ
func (s *Service) GetDocument(ctx context.Context, login, docID string) ([]byte, []byte, string, error) {
	// превращаем в UUID
	id, err := uuid.Parse(docID)
	if err != nil {
		s.log.Error("GetDocument", "failed to parse document id", err)
		return nil, nil, "", err
	}

	doc, err := s.storage.GetDocumentByID(ctx, id, login)
	if err != nil {
		s.log.Error("GetDocument", "failed to get document", err)
		return nil, nil, "", err
	}
	s.log.Info(doc.StoragePath)
	file, err := s.s3.GetFile(doc.StoragePath)
	if err != nil {
		s.log.Error("GetDocument", "failed to get file", err)
		return nil, nil, "", err
	}

	json := doc.JsonDate
	mime := doc.Mime

	return file, json, mime, nil
}

// DeleteDocument - удаляет документ
func (s *Service) DeleteDocument(ctx context.Context, login, docID string) error {
	// переводим в UUID
	id, err := uuid.Parse(docID)
	if err != nil {
		s.log.Error("DeleteDocument", "failed to parse document id", err)
		return err
	}

	err = s.storage.DeleteDocument(ctx, login, id)
	if err != nil {
		s.log.Error("DeleteDocument", "failed to delete document", err)
		return err
	}

	return nil
}
