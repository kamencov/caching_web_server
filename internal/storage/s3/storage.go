package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

//go:generate mockgen -source=storage.go -destination=storage_mock.go -package=s3
type MinioClient interface {
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error)
	RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error
}

type MinioStorage struct {
	client     MinioClient
	bucketName string
	endpoint   string
	useSSL     bool
	log        *slog.Logger
}

// NewMinioStorage - конструктор
func NewMinioStorage(log *slog.Logger) (*MinioStorage, error) {
	endpoint := "minio:9000"
	accessKey := "admin"
	secretKey := "password"
	bucketName := "documents"
	useSSL := false

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})

	if err != nil {
		log.Error("NewMimioStorage", "failed to open connection to mimio", err)
		return nil, err
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}
	if !exists {
		if err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &MinioStorage{
		client:     client,
		bucketName: bucketName,
		endpoint:   endpoint,
		useSSL:     useSSL,
	}, nil
}

// SaveFile - сохранение файла
func (s *MinioStorage) SaveFile(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	_, err := s.client.PutObject(
		ctx,
		s.bucketName,
		key,
		bytes.NewReader(data),
		int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return "", err
	}

	return s.GetFileURL(key), nil
}

// GetFile - получение файла
func (s *MinioStorage) GetFile(key string) ([]byte, error) {
	obj, err := s.client.GetObject(context.Background(), s.bucketName, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer func(obj *minio.Object) {
		err := obj.Close()
		if err != nil {
			s.log.Error("GetFile", "failed to close object", err)
		}
	}(obj)

	return io.ReadAll(obj)
}

// DeleteFile - удаление файла
func (s *MinioStorage) DeleteFile(key string) error {
	return s.client.RemoveObject(context.Background(), s.bucketName, key, minio.RemoveObjectOptions{})
}

// GetFileURL - получение ссылки на файл
func (s *MinioStorage) GetFileURL(key string) string {
	protocol := "http"
	if s.useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, s.endpoint, s.bucketName, key)
}
