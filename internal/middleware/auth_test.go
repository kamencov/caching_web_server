package middleware

import (
	"caching_web_server/internal/service/auth"
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

func TestMiddleware_Authorize(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	ctlr := gomock.NewController(t)
	defer ctlr.Finish()

	mockService := NewMockservice(ctlr)

	type fields struct {
		service service
		log     *slog.Logger
	}

	tests := []struct {
		name       string
		mockUp     func()
		coolieBool bool
		fields     fields
		code       int
	}{
		{
			name: "success_authorize",
			mockUp: func() {
				mockService.EXPECT().VerifyToken(gomock.Any()).Return("test", nil)
			},
			coolieBool: true,
			fields: fields{
				service: mockService,
				log:     log,
			},
			code: http.StatusOK,
		},
		{
			name:       "error_cookie",
			mockUp:     func() {},
			coolieBool: false,
			fields: fields{
				service: mockService,
				log:     log,
			},
			code: http.StatusUnauthorized,
		},
		{
			name: "error_verify_token",
			mockUp: func() {
				mockService.EXPECT().VerifyToken(gomock.Any()).Return("", errors.New("error"))
			},
			coolieBool: true,
			fields: fields{
				service: mockService,
				log:     log,
			},
			code: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockUp()
			m := &Middleware{
				service: tt.fields.service,
				log:     tt.fields.log,
			}
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			if tt.coolieBool {
				r.AddCookie(&http.Cookie{
					Name:  auth.NameCookie,
					Value: "test",
				})

				ctx := context.WithValue(r.Context(), NameLogin, "test")
				r = r.WithContext(ctx)
			}

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("test_auth"))
			})

			m.Authorize(handler).ServeHTTP(w, r)
			if w.Code != tt.code {
				t.Errorf("Authorize() = %v, want %v", w.Code, http.StatusOK)
			}
		})
	}
}

func TestNewMiddleware(t *testing.T) {
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
		want *Middleware
	}{
		{
			name: "success_new_middleware",
			args: args{
				service: mockService,
				log:     log,
			},
			want: &Middleware{
				service: mockService,
				log:     log,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMiddleware(tt.args.service, tt.args.log); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMiddleware() = %v, want %v", got, tt.want)
			}
		})
	}
}
