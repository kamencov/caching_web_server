package get

import (
	"bytes"
	"caching_web_server/internal/middleware"
	"caching_web_server/internal/models"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestHandler_GetDocuments(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	ctlr := gomock.NewController(t)
	defer ctlr.Finish()

	mockService := NewMockService(ctlr)
	type fields struct {
		service Service
		log     *slog.Logger
	}
	type body struct {
		Token       string `json:"token"`
		Login       string `json:"login"`
		FilterKey   string `json:"key" `
		FilterValue string `json:"value"`
		Limit       int    `json:"limit"`
	}

	tests := []struct {
		name     string
		mockUp   func()
		login    string
		method   string
		bodyBool bool
		body     body
		fields   fields
		code     int
	}{
		{
			name: "success_get_documents",
			mockUp: func() {
				mockService.EXPECT().GetDocuments(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]models.DocsData{}, nil)
			},
			login:    "test",
			method:   http.MethodGet,
			bodyBool: true,
			body: body{
				Token:       "test",
				Login:       "test",
				FilterKey:   "",
				FilterValue: "",
				Limit:       1,
			},
			fields: fields{
				service: mockService,
				log:     log,
			},
			code: http.StatusOK,
		},
		{
			name:   "error_method",
			mockUp: func() {},
			login:  "test",
			method: http.MethodPost,
			body: body{
				Token:       "test",
				Login:       "test",
				FilterKey:   "",
				FilterValue: "",
				Limit:       1,
			},
			fields: fields{
				service: mockService,
				log:     log,
			},
			code: http.StatusMethodNotAllowed,
		},
		{
			name:   "error_body",
			mockUp: func() {},
			login:  "test",
			method: http.MethodGet,
			body:   body{},
			fields: fields{
				service: mockService,
				log:     log,
			},
			code: http.StatusBadRequest,
		},
		{
			name:     "error_no_login",
			mockUp:   func() {},
			login:    "",
			method:   http.MethodGet,
			bodyBool: true,
			body: body{
				Token:       "test",
				Login:       "",
				FilterKey:   "",
				FilterValue: "",
				Limit:       1,
			},
			fields: fields{
				service: mockService,
				log:     log,
			},
			code: http.StatusInternalServerError,
		},
		{
			name: "error_get_documents",
			mockUp: func() {
				mockService.EXPECT().GetDocuments(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("get documents error"))
			},
			login:    "test",
			bodyBool: true,
			method:   http.MethodGet,
			body: body{
				Token:       "test",
				Login:       "test",
				FilterKey:   "",
				FilterValue: "",
				Limit:       1,
			},
			fields: fields{
				service: mockService,
				log:     log,
			},
			code: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockUp()

			// создаем body
			var buf bytes.Buffer
			if tt.bodyBool {
				enc := json.NewEncoder(&buf)
				err := enc.Encode(tt.body)
				if err != nil {
					t.Fatal(err)
				}
			}

			w := httptest.NewRecorder()
			ctx := context.WithValue(context.Background(), middleware.NameLogin, tt.login)
			r, err := http.NewRequest(tt.method, "/docs", &buf)
			r.WithContext(ctx)

			if err != nil {
				log.Error("TestHandler_GetDocuments", "failed to create request", err)
				return
			}
			h := &Handler{
				service: tt.fields.service,
				log:     tt.fields.log,
			}
			h.GetDocuments(w, r)

			if w.Code != tt.code {
				t.Errorf("GetDocuments() = %v, want %v", w.Code, tt.code)
			}
		})
	}
}

func TestHandler_GetDocument(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	ctlr := gomock.NewController(t)
	defer ctlr.Finish()

	mockService := NewMockService(ctlr)

	type args struct {
		service Service
		log     *slog.Logger
	}
	tests := []struct {
		name       string
		mockUp     func()
		method     string
		args       args
		docID      string
		cookieBool bool
		code       int
	}{
		{
			name: "success_new_handler_file",
			mockUp: func() {
				mockService.EXPECT().GetDocument(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]byte{}, nil, "image/jpg", nil)
			},
			method: http.MethodGet,
			args: args{
				service: mockService,
				log:     log,
			},
			docID:      "4ebcbb61-8d0f-4c5c-a366-a65464bb9e5d",
			cookieBool: true,
			code:       http.StatusOK,
		},
		{
			name: "success_new_handler_json",
			mockUp: func() {
				mockService.EXPECT().GetDocument(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, []byte{}, "application/json", nil)
			},
			method: http.MethodGet,
			args: args{
				service: mockService,
				log:     log,
			},
			docID:      "4ebcbb61-8d0f-4c5c-a366-a65464bb9e5d",
			cookieBool: true,
			code:       http.StatusOK,
		},
		{
			name:   "error_method",
			mockUp: func() {},
			method: http.MethodPost,
			args: args{
				service: mockService,
				log:     log,
			},
			docID:      "4ebcbb61-8d0f-4c5c-a366-a65464bb9e5d",
			cookieBool: true,
			code:       http.StatusMethodNotAllowed,
		},
		{
			name:   "error_no_doc_id",
			mockUp: func() {},
			method: http.MethodGet,
			args: args{
				service: mockService,
				log:     log,
			},
			docID:      "",
			cookieBool: true,
			code:       http.StatusBadRequest,
		},
		{
			name:       "error_no_login",
			mockUp:     func() {},
			method:     http.MethodGet,
			args:       args{service: mockService, log: log},
			docID:      "4ebcbb61-8d0f-4c5c-a366-a65464bb9e5d",
			cookieBool: false,
			code:       http.StatusInternalServerError,
		},
		{
			name: "error_get_document",
			mockUp: func() {
				mockService.EXPECT().GetDocument(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil, "", errors.New("get document error"))
			},
			method:     http.MethodGet,
			args:       args{service: mockService, log: log},
			docID:      "4ebcbb61-8d0f-4c5c-a366-a65464bb9e5d",
			cookieBool: true,
			code:       http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockUp()
			w := httptest.NewRecorder()
			r, err := http.NewRequest(tt.method, "/api/docs/"+tt.docID, nil)
			if err != nil {
				log.Error("TestHandler_GetDocument", "failed to create request", err)
				return
			}
			h := &Handler{
				service: tt.args.service,
				log:     tt.args.log,
			}

			if tt.cookieBool {
				ctx := context.WithValue(r.Context(), middleware.NameLogin, "test")
				r = r.WithContext(ctx)
			}
			h.GetDocument(w, r)

			if w.Code != tt.code {
				t.Errorf("GetDocument() = %v, want %v", w.Code, tt.code)
			}
		})
	}
}

func TestNewHandler(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	ctlr := gomock.NewController(t)
	defer ctlr.Finish()

	mockService := NewMockService(ctlr)
	type args struct {
		service Service
		log     *slog.Logger
	}
	tests := []struct {
		name string
		args args
		want *Handler
	}{

		{
			name: "success_new_handler",
			args: args{
				service: mockService,
				log:     log,
			},
			want: &Handler{
				service: mockService,
				log:     log,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHandler(tt.args.service, tt.args.log); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
