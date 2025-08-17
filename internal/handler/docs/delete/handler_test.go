package delete

import (
	"caching_web_server/internal/middleware"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestHandler_DeleteData(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	ctlr := gomock.NewController(t)
	defer ctlr.Finish()

	mockService := NewMockservice(ctlr)

	type fields struct {
		service service
		log     *slog.Logger
	}

	tests := []struct {
		name   string
		mockUp func()
		method string
		bodyID string
		fields fields
		code   int
	}{
		{
			name: "success_delete_data",
			mockUp: func() {
				mockService.EXPECT().DeleteDocument(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			method: http.MethodDelete,
			bodyID: "test",
			fields: fields{
				service: mockService,
				log:     log,
			},
			code: http.StatusOK,
		},
		{
			name:   "error_method",
			mockUp: func() {},
			method: http.MethodGet,
			bodyID: "test",
			fields: fields{
				service: mockService,
				log:     log,
			},
			code: http.StatusMethodNotAllowed,
		},
		{
			name:   "error_no_id",
			mockUp: func() {},
			method: http.MethodDelete,
			bodyID: "",
			fields: fields{
				service: mockService,
				log:     log,
			},
			code: http.StatusBadRequest,
		},
		{
			name: "error_delete_data",
			mockUp: func() {
				mockService.EXPECT().DeleteDocument(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New(
					"error"))
			},
			method: http.MethodDelete,
			bodyID: "test",
			fields: fields{
				service: mockService,
				log:     log,
			},
			code: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				service: tt.fields.service,
				log:     tt.fields.log,
			}
			tt.mockUp()
			w := httptest.NewRecorder()
			r, err := http.NewRequest(tt.method, "/api/docs/"+tt.bodyID, nil)
			if err != nil {
				t.Errorf("Failed to create request: %v", err)
			}

			ctx := context.WithValue(r.Context(), middleware.NameLogin, "test")
			r = r.WithContext(ctx)
			h.DeleteData(w, r)
			if w.Code != tt.code {
				t.Errorf("Status code is not correct. Got %d, want %d.", w.Code, tt.code)
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
