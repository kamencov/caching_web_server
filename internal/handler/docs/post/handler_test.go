package post

import (
	"bytes"
	"caching_web_server/internal/middleware"
	"caching_web_server/internal/models"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
)

func createMultipart(t *testing.T, body models.Meta, meta, JSON, file bool) (*multipart.Writer, io.Reader, error) {
	bodyJson, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	defer func(writer *multipart.Writer) {
		err := writer.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(writer)
	if meta {
		metaPart, err := writer.CreateFormField("meta")
		if err != nil {
			t.Fatal(err)
		}
		_, err = metaPart.Write(bodyJson)
		if err != nil {
			t.Fatal(err)
		}

		if JSON {
			// json
			jsonPart, err := writer.CreateFormField("json")
			if err != nil {
				t.Fatal(err)
			}
			_, err = jsonPart.Write([]byte(`{"document": "test"}`))
			if err != nil {
				t.Fatal(err)
			}
		}

		if file {
			// file
			filePart, err := writer.CreateFormFile("file", "test.jpg")
			if err != nil {
				t.Fatal(err)
			}
			_, err = filePart.Write([]byte("test"))
			if err != nil {
				t.Fatal(err)
			}
		}

	}

	return writer, &buf, nil
}
func TestHandler_SaveDocument(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	ctlr := gomock.NewController(t)
	defer ctlr.Finish()
	mockService := NewMockservice(ctlr)

	type fields struct {
		service service
		log     *slog.Logger
		maxSize int64
	}
	tests := []struct {
		name   string
		fields fields
		body   models.Meta
		meta   bool
		json   bool
		file   bool
		mockUp func()
		code   int
	}{
		{
			name: "success",
			fields: fields{
				service: mockService,
				log:     log,
				maxSize: 10 << 20,
			},
			body: models.Meta{
				Name:   "test",
				File:   true,
				Public: true,
				Token:  "test",
				Mime:   "image/jpg",
				Grants: []string{"test1", "test2"},
			},
			meta: true,
			json: true,
			file: true,
			mockUp: func() {
				mockService.EXPECT().SaveDocument(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			code: http.StatusOK,
		},
		{
			name: "error_max_size",
			fields: fields{
				service: mockService,
				log:     log,
				maxSize: 10,
			},
			body: models.Meta{
				Name:   "test",
				File:   true,
				Public: true,
				Token:  "test",
				Mime:   "image/jpg",
				Grants: []string{"test1", "test2"},
			},
			mockUp: func() {},
			code:   http.StatusBadRequest,
		},
		{
			name: "error_meta",
			fields: fields{
				service: mockService,
				log:     log,
				maxSize: 10 << 20,
			},
			meta:   false,
			body:   models.Meta{},
			mockUp: func() {},
			code:   http.StatusBadRequest,
		},
		{
			name: "error_file",
			fields: fields{
				service: mockService,
				log:     log,
				maxSize: 10 << 20,
			},
			meta:   true,
			body:   models.Meta{},
			mockUp: func() {},
			code:   http.StatusBadRequest,
		},
		{
			name: "error_save_document",
			fields: fields{
				service: mockService,
				log:     log,
				maxSize: 10 << 20,
			},
			body: models.Meta{
				Name:   "test",
				File:   true,
				Public: true,
				Token:  "test",
				Mime:   "image/jpg",
				Grants: []string{"test1", "test2"},
			},
			meta: true,
			file: true,
			mockUp: func() {
				mockService.EXPECT().SaveDocument(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("test"))
			},
			code: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockUp()
			h := &Handler{
				service: tt.fields.service,
				log:     tt.fields.log,
				maxSize: tt.fields.maxSize,
			}
			w := httptest.NewRecorder()

			writer, buf, err := createMultipart(t, tt.body, tt.meta, tt.json, tt.file)
			if err != nil {
				t.Fatal(err)
			}

			r := httptest.NewRequest(http.MethodPost, "/api/docs", buf)
			ctx := context.WithValue(r.Context(), middleware.NameLogin, "test")
			r = r.WithContext(ctx)
			r.Header.Set("Content-Type", writer.FormDataContentType())
			h.SaveDocument(w, r)

			if w.Code != tt.code {
				t.Errorf("SaveDocument() = %v, want %v", w.Code, tt.code)
			}

		})
	}
}

func TestNewHandler(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	ctlr := gomock.NewController(t)
	defer ctlr.Finish()

	mockService := NewMockservice(ctlr)

	type args struct {
		service service
		log     *slog.Logger
		maxSize int64
	}
	tests := []struct {
		name string
		args args
		want *Handler
	}{
		{
			name: "success",
			args: args{
				service: mockService,
				log:     log,
				maxSize: 0,
			},
			want: &Handler{
				service: mockService,
				log:     log,
				maxSize: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHandler(tt.args.service, tt.args.log, tt.args.maxSize); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
