package pq

import (
	"caching_web_server/internal/models"
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

var errStorage = errors.New("error")

func TestStorage_Close(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании мок-базы данных: %v", err)
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	s := &Storage{
		db:  db,
		log: log,
	}

	mock.ExpectClose()
	err = s.Close()
	if err != nil {
		t.Fatalf("ошибка при закрытии базы данных: %v", err)
	}
}

func TestStorage_SaveUser(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("ошибка при создании мок-базы данных: %v", err)
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	tests := []struct {
		name         string
		login        string
		passwordHash string
		mockUp       func()
		wantErr      error
	}{
		{
			name:         "success_save_user",
			login:        "test",
			passwordHash: "test",
			mockUp: func() {
				mock.ExpectExec("INSERT INTO users").
					WithArgs("test", "test").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: nil,
		},
		{
			name: "error_save_user",
			mockUp: func() {
				mock.ExpectExec("INSERT INTO users").
					WillReturnError(errStorage)
			},
			wantErr: errStorage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := &Storage{
				db:  db,
				log: log,
			}
			tt.mockUp()
			err := s.SaveUser(ctx, tt.login, tt.passwordHash)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("SaveUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func TestStorage_GetHashPass(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании мок-базы данных: %v", err)
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	tests := []struct {
		name    string
		login   string
		mockUp  func()
		wantErr error
	}{
		{
			name:  "success_get_hash_pass",
			login: "test",
			mockUp: func() {
				mock.ExpectQuery("SELECT password_hash").
					WithArgs("test").
					WillReturnRows(sqlmock.NewRows([]string{"passwordHash"}).AddRow("test"))
			},
			wantErr: nil,
		},
		{
			name:  "error_get_hash_pass",
			login: "test",
			mockUp: func() {
				mock.ExpectQuery("SELECT password_hash").
					WithArgs("test").
					WillReturnError(errStorage)
			},
			wantErr: errStorage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tt.mockUp()
			s := &Storage{
				db:  db,
				log: log,
			}
			_, err := s.GetHashPass(ctx, tt.login)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("GetHashPass() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}

func TestStorage_SaveDocument(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании мок-базы данных: %v", err)
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	docID := uuid.New()

	tests := []struct {
		name    string
		mockUp  func()
		doc     models.Document
		grants  []string
		wantErr error
	}{
		{
			name: "success_save_document",
			mockUp: func() {
				mock.ExpectBegin()
				mock.ExpectQuery("INSERT INTO documents").
					WithArgs(
						sqlmock.AnyArg(), // id, может быть UUID
						int64(1),         // OwnerID
						"name",           // Name
						"mime",           // Mime
						true,             // HashFile
						true,             // Public
						[]byte{},         // JsonData
						"path").          // StoragePath).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(docID))
				mock.ExpectExec("INSERT INTO grants").
					WithArgs(docID, "login2").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			doc: models.Document{
				OwnerID:     1,
				Name:        "name",
				Mime:        "mime",
				Public:      true,
				HashFile:    true,
				JsonDate:    []byte{},
				StoragePath: "path",
			},
			grants:  []string{"login2"},
			wantErr: nil,
		},
		{
			name: "error_save_document",
			mockUp: func() {
				mock.ExpectBegin()
				mock.ExpectQuery("INSERT INTO documents").
					WillReturnError(errStorage)
				mock.ExpectRollback()
			},
			doc: models.Document{
				OwnerID:     1,
				Name:        "name",
				Mime:        "mime",
				Public:      true,
				HashFile:    true,
				JsonDate:    []byte{},
				StoragePath: "path",
			},
			wantErr: errStorage,
		},
		{
			name: "error_save_grants",
			mockUp: func() {
				mock.ExpectBegin()
				mock.ExpectQuery("INSERT INTO documents").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(docID))
				mock.ExpectExec("INSERT INTO grants").
					WillReturnError(errStorage)
				mock.ExpectRollback()
			},
			doc: models.Document{
				OwnerID:     1,
				Name:        "name",
				Mime:        "mime",
				Public:      true,
				HashFile:    true,
				JsonDate:    []byte{},
				StoragePath: "path",
			},
			grants:  []string{"login2"},
			wantErr: errStorage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tt.mockUp()
			s := &Storage{
				db:  db,
				log: log,
			}
			err := s.SaveDocument(ctx, &tt.doc, tt.grants)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("SaveDocument() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStorage_GetUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании мок-базы данных: %v", err)
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	tests := []struct {
		name    string
		mockUp  func()
		login   string
		wantErr error
	}{
		{
			name:  "success_get_user_id",
			login: "test",
			mockUp: func() {
				mock.ExpectQuery("SELECT id").
					WithArgs("test").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			wantErr: nil,
		},
		{
			name:  "error_get_user_id",
			login: "test",
			mockUp: func() {
				mock.ExpectQuery("SELECT id").
					WithArgs("test").
					WillReturnError(errStorage)
			},
			wantErr: errStorage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tt.mockUp()
			s := &Storage{
				db:  db,
				log: log,
			}
			_, err := s.GetUserID(ctx, tt.login)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("GetUserID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStorage_GetDocuments(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании мок-базы данных: %v", err)
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	tests := []struct {
		name        string
		mock        func()
		login       string
		filterKey   string
		filterValue string
		limit       int
		wantErr     error
		wantDocs    int
	}{
		{
			name: "success_get_documents_without_filter",
			mock: func() {
				mockRows := sqlmock.NewRows([]string{
					"id", "name", "mime", "hash_file", "public",
					"create_at", "grants",
				}).AddRow(
					"uuid1", "doc1", "mime", true, true,
					time.Now(), pq.Array([]string{"login2", "login3"}),
				).AddRow(
					"uuid2", "doc2", "mime2", false, true,
					time.Now(), pq.Array([]string{}),
				)

				mock.ExpectQuery("WITH owner_id AS").
					WithArgs("login1", 10).
					WillReturnRows(mockRows)
			},
			login:       "login1",
			filterKey:   "",
			filterValue: "",
			limit:       10,
			wantErr:     nil,
			wantDocs:    2,
		},
		{
			name: "success_get_documents_with_filter",
			mock: func() {
				mockRows := sqlmock.NewRows([]string{
					"id", "name", "mime", "hash_file", "public",
					"create_at", "grants",
				}).AddRow(
					"uuid3", "doc3", "mime3", true, false,
					time.Now(), pq.Array([]string{"login4"}),
				)

				mock.ExpectQuery("WITH owner_id AS").
					WithArgs("login1", "doc3", 5).
					WillReturnRows(mockRows)
			},
			login:       "login1",
			filterKey:   "name",
			filterValue: "doc3",
			limit:       5,
			wantErr:     nil,
			wantDocs:    1,
		},
		{
			name: "error_get_documents",
			mock: func() {
				mock.ExpectQuery("WITH owner_id AS").
					WithArgs("login1", "doc3", 5).
					WillReturnError(errStorage)
			},
			login:       "login1",
			filterKey:   "name",
			filterValue: "doc3",
			limit:       5,
			wantErr:     errStorage,
			wantDocs:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tt.mock()
			s := &Storage{
				db:  db,
				log: log,
			}
			docs, err := s.GetDocuments(ctx, tt.login, tt.filterKey, tt.filterValue, tt.limit)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("GetDocuments() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(docs) != tt.wantDocs {
				t.Errorf("expected %d docs, got %d", tt.wantDocs, len(docs))
			}

			for _, doc := range docs {
				for _, g := range doc.Grants {
					if g == "" {
						t.Errorf("found empty grant in doc %s", doc.Id)
					}
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestStorage_GetDocumentByID(t *testing.T) {
	docID := uuid.New()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании мок-базы данных: %v", err)
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	tests := []struct {
		name    string
		mock    func()
		docID   uuid.UUID
		login   string
		wantNil bool
		wantErr error
	}{
		{
			name: "success_get_document_by_id",
			mock: func() {
				mockRows := sqlmock.NewRows([]string{
					"id", "owner_id", "name", "mime", "hash_file", "public",
					"json_data", "storage_path", "create_at", "is_deleted",
				}).AddRow(
					"uuid1", int64(1), "doc1", "mime", true, true,
					[]byte{}, "path1", time.Now(), false)
				mock.ExpectQuery("WITH owner_id AS").
					WithArgs(docID, "login1").
					WillReturnRows(mockRows)
			},
			docID:   docID,
			login:   "login1",
			wantNil: false,
			wantErr: nil,
		},
		{
			name: "error_get_document_by_id",
			mock: func() {
				mock.ExpectQuery("WITH owner_id AS").
					WithArgs(docID, "login1").
					WillReturnError(errStorage)
			},
			docID:   docID,
			login:   "login1",
			wantNil: true,
			wantErr: errStorage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tt.mock()
			s := &Storage{
				db:  db,
				log: log,
			}
			doc, err := s.GetDocumentByID(ctx, tt.docID, tt.login)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("GetDocuments() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (doc == nil) != tt.wantNil {
				t.Errorf("GetDocumentByID() doc = %+v, expected nil=%v", doc, tt.wantNil)
			}
		})
	}
}

func TestStorage_DeleteDocument(t *testing.T) {
	docID := uuid.New()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании мок-базы данных: %v", err)
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	tests := []struct {
		name    string
		login   string
		mock    func()
		docID   uuid.UUID
		wantErr error
	}{
		{
			name:  "success_delete_document",
			login: "login1",
			mock: func() {
				mock.ExpectExec("UPDATE documents").
					WithArgs(docID, "login1").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			docID:   docID,
			wantErr: nil,
		},
		{
			name:  "error_document_not_found",
			login: "login1",
			mock: func() {
				mock.ExpectExec("UPDATE documents").
					WithArgs(docID, "login1").
					WillReturnResult(sqlmock.NewResult(1, 0))
			},
			docID:   docID,
			wantErr: ErrDocumentNotFound,
		},
		{
			name:  "error_delete_document",
			login: "login1",
			mock: func() {
				mock.ExpectExec("UPDATE documents").
					WithArgs(docID, "login1").
					WillReturnError(errStorage)
			},
			docID:   docID,
			wantErr: errStorage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tt.mock()
			s := &Storage{
				db:  db,
				log: log,
			}
			err := s.DeleteDocument(ctx, tt.login, tt.docID)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("DeleteDocument() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
