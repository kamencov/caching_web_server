package docs

import (
	"caching_web_server/internal/models"
	"caching_web_server/internal/storage/pq"
	"context"
	"errors"
	"log/slog"
	"os"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

var errStorage = errors.New("storage error")

func TestNewService(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockstorage(ctrl)
	mockS3 := NewMocks3(ctrl)

	type args struct {
		storage storage
		s3      s3
		log     *slog.Logger
	}
	tests := []struct {
		name string
		args args
		want *Service
	}{
		{
			name: "success",
			args: args{
				storage: mockStorage,
				s3:      mockS3,
				log:     log,
			},
			want: &Service{
				storage: mockStorage,
				s3:      mockS3,
				log:     log,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewService(tt.args.storage, tt.args.s3, tt.args.log); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_SaveDocument(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockstorage(ctrl)
	mockS3 := NewMocks3(ctrl)

	type fields struct {
		storage storage
		s3      s3
		log     *slog.Logger
	}
	type args struct {
		ctx      context.Context
		login    string
		meta     models.Meta
		jsonData []byte
		file     []byte
	}
	tests := []struct {
		name    string
		mock    func()
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success_save_document",
			mock: func() {
				mockS3.EXPECT().SaveFile(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("url", nil)
				mockStorage.EXPECT().GetUserID(gomock.Any(), gomock.Any()).Return(1, nil)
				mockStorage.EXPECT().SaveDocument(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			fields: fields{
				storage: mockStorage,
				s3:      mockS3,
				log:     log,
			},
			args: args{
				ctx:      context.Background(),
				login:    "test",
				meta:     models.Meta{Name: "test"},
				jsonData: []byte{},
				file:     []byte{},
			},
			wantErr: false,
		},
		{
			name: "error_s3",
			mock: func() {
				mockS3.EXPECT().SaveFile(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", errors.New("s3 error"))
			},
			fields: fields{
				storage: mockStorage,
				s3:      mockS3,
				log:     log,
			},
			args: args{
				ctx:      context.Background(),
				login:    "test",
				meta:     models.Meta{Name: "test"},
				jsonData: []byte{},
				file:     []byte{},
			},
			wantErr: true,
		},
		{
			name: "error_get_user_id",
			mock: func() {
				mockS3.EXPECT().SaveFile(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("url", nil)
				mockStorage.EXPECT().GetUserID(gomock.Any(), gomock.Any()).Return(0, errors.New("storage error"))
			},
			fields: fields{
				storage: mockStorage,
				s3:      mockS3,
				log:     log,
			},
			args: args{
				ctx:      context.Background(),
				login:    "test",
				meta:     models.Meta{Name: "test"},
				jsonData: []byte{},
				file:     []byte{},
			},
			wantErr: true,
		},
		{
			name: "error_save_document",
			mock: func() {
				mockS3.EXPECT().SaveFile(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("url", nil)
				mockStorage.EXPECT().GetUserID(gomock.Any(), gomock.Any()).Return(1, nil)
				mockStorage.EXPECT().SaveDocument(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("storage error"))
				mockS3.EXPECT().DeleteFile(gomock.Any()).Return(nil)
			},
			fields: fields{
				storage: mockStorage,
				s3:      mockS3,
				log:     log,
			},
			args: args{
				ctx:      context.Background(),
				login:    "test",
				meta:     models.Meta{Name: "test"},
				jsonData: []byte{},
				file:     []byte{},
			},
			wantErr: true,
		},
		{
			name: "error_delete_file",
			mock: func() {
				mockS3.EXPECT().SaveFile(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("url", nil)
				mockStorage.EXPECT().GetUserID(gomock.Any(), gomock.Any()).Return(1, nil)
				mockStorage.EXPECT().SaveDocument(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("storage error"))
				mockS3.EXPECT().DeleteFile(gomock.Any()).Return(errors.New("s3 error"))
			},
			fields: fields{
				storage: mockStorage,
				s3:      mockS3,
				log:     log,
			},
			args: args{
				ctx:      context.Background(),
				login:    "test",
				meta:     models.Meta{Name: "test"},
				jsonData: []byte{},
				file:     []byte{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			s := &Service{
				storage: tt.fields.storage,
				s3:      tt.fields.s3,
				log:     tt.fields.log,
			}
			if err := s.SaveDocument(tt.args.ctx, tt.args.login, tt.args.meta, tt.args.jsonData, tt.args.file); (err != nil) != tt.wantErr {
				t.Errorf("SaveDocument() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_GetDocuments(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockstorage(ctrl)
	mockS3 := NewMocks3(ctrl)

	type args struct {
		ctx   context.Context
		login string
		key   string
		value string
		limit int
	}
	tests := []struct {
		name    string
		mock    func()
		args    args
		wantErr error
	}{
		{
			name: "success_get_documents",
			mock: func() {
				mockStorage.EXPECT().GetDocuments(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]models.DocsData{}, nil)
			},
			args: args{
				ctx:   context.Background(),
				login: "test",
				key:   "",
				value: "",
				limit: 10,
			},
			wantErr: nil,
		},
		{
			name: "error_get_documents",
			mock: func() {
				mockStorage.EXPECT().GetDocuments(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errStorage)
			},
			args: args{
				ctx:   context.Background(),
				login: "test",
				key:   "",
				value: "",
				limit: 10,
			},
			wantErr: errStorage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			s := &Service{
				storage: mockStorage,
				s3:      mockS3,
				log:     log,
			}
			if _, err := s.GetDocuments(tt.args.ctx, tt.args.login, tt.args.key, tt.args.value, tt.args.limit); err != tt.wantErr {
				t.Errorf("GetDocuments() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_GetDocument(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockstorage(ctrl)
	mockS3 := NewMocks3(ctrl)

	docID := "4ebcbb61-8d0f-4c5c-a366-a65464bb9e5d"

	tests := []struct {
		name string
		mock func()
		want error
	}{
		{
			name: "success_get_document",
			mock: func() {
				mockStorage.EXPECT().
					GetDocumentByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&models.Document{
						Name:        "test.pdf",
						Mime:        "application/pdf",
						StoragePath: "url",
					}, nil)
				mockS3.EXPECT().
					GetFile(gomock.Any()).
					Return([]byte("file content"), nil)
			},
			want: nil,
		},
		{
			name: "error_get_document",
			mock: func() {
				mockStorage.EXPECT().
					GetDocumentByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errStorage)
			},
			want: errStorage,
		},
		{
			name: "error_get_file",
			mock: func() {
				mockStorage.EXPECT().
					GetDocumentByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&models.Document{
						Name:        "test.pdf",
						Mime:        "application/pdf",
						StoragePath: "url",
					}, nil)
				mockS3.EXPECT().
					GetFile(gomock.Any()).
					Return(nil, errStorage)
			},
			want: errStorage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			s := &Service{
				storage: mockStorage,
				s3:      mockS3,
				log:     log,
			}
			_, _, _, err := s.GetDocument(context.Background(), "test", docID)
			if !errors.Is(err, tt.want) {
				t.Errorf("GetDocument() error = %v, wantErr %v", err, tt.want)
			}
		})
	}
}
func TestService_DeleteDocument(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockstorage(ctrl)
	mockS3 := NewMocks3(ctrl)

	_, err := uuid.Parse("1")

	type args struct {
		ctx   context.Context
		login string
		id    string
	}
	tests := []struct {
		name    string
		mock    func()
		args    args
		wantErr error
	}{
		{
			name: "success_delete_document",
			mock: func() {
				mockStorage.EXPECT().DeleteDocument(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			args: args{
				ctx:   context.Background(),
				login: "test",
				id:    "4ebcbb61-8d0f-4c5c-a366-a65464bb9e5d",
			},
			wantErr: nil,
		},
		{
			name: "error_invalid_document_id",
			mock: func() {},
			args: args{
				ctx:   context.Background(),
				login: "test",
				id:    "1",
			},
			wantErr: err,
		},
		{
			name: "error_delete_document",
			mock: func() {
				mockStorage.EXPECT().DeleteDocument(gomock.Any(), gomock.Any(), gomock.Any()).Return(pq.ErrDocumentNotFound)
			},
			args: args{
				ctx:   context.Background(),
				login: "test",
				id:    "4ebcbb61-8d0f-4c5c-a366-a65464bb9e5d",
			},
			wantErr: pq.ErrDocumentNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			s := &Service{
				storage: mockStorage,
				s3:      mockS3,
				log:     log,
			}
			if err := s.DeleteDocument(tt.args.ctx, tt.args.login, tt.args.id); !errors.Is(err, tt.wantErr) {
				t.Errorf("DeleteDocument() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
