package s3

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/require"
)

func TestMinioStorage_SaveFile(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockMinioClient(ctrl)

	data := []byte("hello world")

	tests := []struct {
		name    string
		mock    func()
		wantErr bool
	}{
		{
			name: "success_save_file",
			mock: func() {
				mockClient.EXPECT().
					PutObject(gomock.Any(), "documents", "file.txt", gomock.Any(), int64(len(data)), gomock.Any()).
					Return(minio.UploadInfo{}, nil)
			},
			wantErr: false,
		},
		{
			name: "error_save_file",
			mock: func() {
				mockClient.EXPECT().
					PutObject(gomock.Any(), "documents", "file.txt", gomock.Any(), int64(len(data)), gomock.Any()).
					Return(minio.UploadInfo{}, fmt.Errorf("error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &MinioStorage{
				client:     mockClient,
				bucketName: "documents",
				endpoint:   "http://minio:9000",
				useSSL:     false,
				log:        log,
			}
			tt.mock()
			_, err := storage.SaveFile(context.Background(), "file.txt", data, "text/plain")
			if (err != nil) != tt.wantErr {
				t.Errorf("MinioStorage.SaveFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMinioStorage_DeleteFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockMinioClient(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	key := "file.txt"

	mockClient.EXPECT().
		RemoveObject(gomock.Any(), "documents", key, minio.RemoveObjectOptions{}).
		Return(nil)

	ms := &MinioStorage{
		client:     mockClient,
		bucketName: "documents",
		log:        logger,
	}

	err := ms.DeleteFile(key)
	require.NoError(t, err)
}

func TestMinioStorage_GetFileURL(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	ms := &MinioStorage{
		bucketName: "documents",
		endpoint:   "localhost:9000",
		useSSL:     false,
		log:        logger,
	}

	url := ms.GetFileURL("file.txt")
	require.Equal(t, "http://localhost:9000/documents/file.txt", url)

	ms.useSSL = true
	url = ms.GetFileURL("file.txt")
	require.Equal(t, "https://localhost:9000/documents/file.txt", url)
}
