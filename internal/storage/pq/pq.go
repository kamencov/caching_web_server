package pq

import (
	"caching_web_server/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

var (
	ErrDocumentNotFound = errors.New("document not found")
)

type Storage struct {
	db  *sql.DB
	log *slog.Logger
}

func NewStorage(log *slog.Logger) (*Storage, error) {
	db := &Storage{
		log: log,
	}

	newDB, err := sql.Open("postgres", "postgres://postgres:password@postgres:5432/documents?sslmode=disable")
	if err != nil {
		log.Error("NewDB", "failed to open connection to db", err)
		return nil, err
	}

	db.db = newDB

	err = db.Migrate(newDB)

	if err != nil {
		log.Error("NewDB", "failed to migrate db", err)
		return nil, err
	}
	return db, nil

}

func (s *Storage) Migrate(sql *sql.DB) error {
	err := goose.Up(sql, "migrations")
	if err != nil {
		s.log.Error("Migrate", "failed to migrate s", err)
		return err
	}
	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) SaveUser(ctx context.Context, login, passwordHash string) error {
	query := `INSERT INTO users (login, password_hash) VALUES ($1, $2)`
	_, err := s.db.ExecContext(ctx, query, login, passwordHash)
	if err != nil {
		s.log.Error("SaveUser", "failed to save user", err)
		return err
	}
	return nil
}

// GetHashPass - получить пароль из базы по логину
func (s *Storage) GetHashPass(ctx context.Context, login string) (string, error) {
	query := `SELECT password_hash FROM users WHERE login = $1`
	var passwordHash string
	err := s.db.QueryRowContext(ctx, query, login).Scan(&passwordHash)
	if err != nil {
		s.log.Error("GetHashPass", "failed to get hash pass", err)
		return "", err
	}
	return passwordHash, nil
}

// SaveDocument сохраняет документ
func (s *Storage) SaveDocument(ctx context.Context, doc *models.Document, grants []string) error {
	var docID uuid.UUID
	id := uuid.New()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				s.log.Error("SaveDocument", "rollback failed", rbErr)
			}
		} else {
			if cmErr := tx.Commit(); cmErr != nil {
				s.log.Error("SaveDocument", "commit failed", cmErr)
				err = cmErr
			}
		}
	}()

	query := `INSERT INTO documents (
                       id,
                       owner_id, 
                       name, 
                       mime, 
                       hash_file, 
                       public, 
                       json_data, 
                       storage_path)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	err = tx.QueryRowContext(ctx, query,
		id,
		doc.OwnerID,
		doc.Name,
		doc.Mime,
		doc.HashFile,
		doc.Public,
		doc.JsonDate,
		doc.StoragePath).
		Scan(&docID)
	if err != nil {
		s.log.Error("SaveDocument", "failed to save document", err)
		return err
	}

	// сохраняем гранты
	for _, grant := range grants {
		query = `INSERT INTO grants (doc_id, user_id)
					SELECT d.id, u.id
					FROM users u
					JOIN documents d ON d.id = $1
					WHERE u.login = $2`
		_, err = tx.ExecContext(ctx, query, docID, grant)
		if err != nil {
			s.log.Error("SaveDocument", "failed to save grant", err)
			return err
		}
	}
	return nil
}

// GetUserID - возвращает id пользователя
func (s *Storage) GetUserID(ctx context.Context, login string) (int, error) {
	query := `SELECT id FROM users WHERE login = $1`
	var id int
	err := s.db.QueryRowContext(ctx, query, login).Scan(&id)
	if err != nil {
		s.log.Error("GetUserID", "failed to get user id", err)
		return 0, err
	}
	return id, nil
}

// GetDocuments - возвращает список документов
func (s *Storage) GetDocuments(ctx context.Context, login, filterKey, filterValue string, limit int) ([]models.DocsData, error) {
	query := `
WITH owner_id AS (
    SELECT id
    FROM users
    WHERE login = $1
)
SELECT DISTINCT d.id, d.name, d.mime, d.hash_file, d.public,
       d.create_at,
       ARRAY_REMOVE(ARRAY_AGG(g_user.login), NULL) AS grants
FROM documents d
LEFT JOIN grants g ON d.id = g.doc_id
LEFT JOIN users g_user ON g.user_id = g_user.id
JOIN owner_id o ON d.owner_id = o.id
WHERE d.is_deleted = false
`

	args := []any{login}

	if filterKey != "" && filterValue != "" {
		query += fmt.Sprintf(" AND d.%s = $%d", filterKey, len(args)+1)
		args = append(args, filterValue)
	}

	query += fmt.Sprintf(`
GROUP BY d.id
ORDER BY d.name ASC, d.create_at DESC
LIMIT $%d
`, len(args)+1)
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		s.log.Error("GetDocuments", "failed to get documents", err)
		return nil, err
	}
	defer func(rows *sql.Rows) {
		if cerr := rows.Close(); cerr != nil {
			s.log.Error("GetDocuments", "failed to close rows", cerr)
		}
	}(rows)

	var docs []models.DocsData
	for rows.Next() {
		var doc models.DocsData
		var grants []sql.NullString

		err := rows.Scan(
			&doc.Id,
			&doc.Name,
			&doc.Mime,
			&doc.File,
			&doc.Public,
			&doc.Created,
			pq.Array(&grants),
		)
		if err != nil {
			s.log.Error("GetDocuments", "failed to scan row", err)
			return nil, err
		}

		for _, g := range grants {
			if g.Valid {
				doc.Grants = append(doc.Grants, g.String)
			}
		}

		docs = append(docs, doc)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return docs, nil
}

// GetDocumentByID - возвращает документ
func (s *Storage) GetDocumentByID(ctx context.Context, docID uuid.UUID, login string) (*models.Document, error) {
	query := `
WITH owner_id AS (
    SELECT id
    FROM users
    WHERE login = $2
)
SELECT d.id, d.owner_id, d.name, d.mime, d.hash_file, d.public,
       d.json_data, d.storage_path, d.create_at, d.is_deleted
FROM documents d
LEFT JOIN grants g ON d.id = g.doc_id
JOIN owner_id o ON true
WHERE d.id = $1
  AND d.is_deleted = false
  AND (d.owner_id = o.id OR g.user_id = o.id)
LIMIT 1
`

	var doc models.Document
	err := s.db.QueryRowContext(ctx, query, docID, login).Scan(
		&doc.ID,
		&doc.OwnerID,
		&doc.Name,
		&doc.Mime,
		&doc.HashFile,
		&doc.Public,
		&doc.JsonDate,
		&doc.StoragePath,
		&doc.CreatedAt,
		&doc.IsDeleted,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Error("GetDocumentByID", "document not found", err)
			return nil, nil
		}
		s.log.Error("GetDocumentByID", "failed to get document", err)
		return nil, err
	}
	return &doc, nil
}

// DeleteDocument - удаляет документ
func (s *Storage) DeleteDocument(ctx context.Context, login string, docID uuid.UUID) error {
	query := `UPDATE documents SET is_deleted = true WHERE id = $1 AND owner_id = (SELECT id FROM users WHERE login = $2)`
	res, err := s.db.ExecContext(ctx, query, docID, login)
	if err != nil {
		s.log.Error("DeleteDocument", "failed to delete document", err)
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		s.log.Error("DeleteDocument", "document not found or not owned by user", err)
		return ErrDocumentNotFound
	}
	return nil
}
