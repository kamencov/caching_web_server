package auth

import (
	"bytes"
	"caching_web_server/internal/middleware"
	"caching_web_server/internal/service/auth"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestNewHandler(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockservice(ctrl)
	type args struct {
		service    service
		log        *slog.Logger
		adminToken string
	}
	tests := []struct {
		name string
		args args
		want *Handler
	}{
		{
			name: "success",
			args: args{
				service:    mockService,
				log:        log,
				adminToken: "",
			},
			want: &Handler{
				service:    mockService,
				log:        log,
				adminToken: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHandler(tt.args.service, tt.args.log, tt.args.adminToken); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler_Register(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	type body struct {
		Token    string `json:"token"`
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	tests := []struct {
		name              string
		method            string
		token             string
		body              body
		flagBody          bool
		errorRegisterUser error
		wantCode          int
	}{
		{
			name:   "success_register",
			method: http.MethodPost,
			token:  "test",
			body: body{
				Token:    "test",
				Login:    "test",
				Password: "test",
			},
			errorRegisterUser: nil,
			wantCode:          http.StatusOK,
		},
		{
			name:   "err_method",
			method: http.MethodGet,
			token:  "test",
			body: body{
				Token:    "test",
				Login:    "test",
				Password: "test",
			},
			errorRegisterUser: nil,
			wantCode:          http.StatusMethodNotAllowed,
		},
		{
			name:              "err_no_body",
			method:            http.MethodPost,
			flagBody:          true,
			errorRegisterUser: nil,
			wantCode:          http.StatusBadRequest,
		},
		{
			name:   "err_register_user",
			method: http.MethodPost,
			token:  "test",
			body: body{
				Token:    "test",
				Login:    "test",
				Password: "test",
			},
			errorRegisterUser: fmt.Errorf("test"),
			wantCode:          http.StatusInternalServerError,
		},
		{
			name:   "err_invalid_token",
			method: http.MethodPost,
			token:  "test",
			body: body{
				Token:    "tes",
				Login:    "test",
				Password: "test",
			},
			errorRegisterUser: nil,
			wantCode:          http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := NewMockservice(ctrl)

			h := &Handler{
				service:    mockService,
				log:        log,
				adminToken: tt.token,
			}

			mockService.EXPECT().RegisterUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(tt.errorRegisterUser).AnyTimes()

			w := httptest.NewRecorder()

			var reqBody io.Reader
			if !tt.flagBody {
				buf := bytes.NewBuffer(nil)
				err := json.NewEncoder(buf).Encode(tt.body)
				if err != nil {
					return
				}
				reqBody = buf
			} else {
				reqBody = nil
			}
			r := httptest.NewRequest(tt.method, "/register", reqBody)
			h.Register(w, r)
			if w.Code != tt.wantCode {
				t.Errorf("Handler.Register() = %v, want %v", w.Code, tt.wantCode)
			}

		})
	}
}

func TestHandler_Auth(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	type body struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	tests := []struct {
		name          string
		method        string
		token         string
		body          body
		flagBody      bool
		errorAuthUser error
		wantCode      int
	}{
		{
			name:   "success_register",
			method: http.MethodPost,
			token:  "test",
			body: body{
				Login:    "test",
				Password: "test",
			},
			errorAuthUser: nil,
			wantCode:      http.StatusOK,
		},
		{
			name:   "err_method",
			method: http.MethodGet,
			token:  "test",
			body: body{
				Login:    "test",
				Password: "test",
			},
			errorAuthUser: nil,
			wantCode:      http.StatusMethodNotAllowed,
		},
		{
			name:          "err_no_body",
			method:        http.MethodPost,
			flagBody:      true,
			errorAuthUser: nil,
			wantCode:      http.StatusBadRequest,
		},
		{
			name:   "login_is_empty",
			method: http.MethodPost,
			body: body{
				Login:    "",
				Password: "test",
			},
			errorAuthUser: nil,
			wantCode:      http.StatusBadRequest,
		},
		{
			name:   "err_auth_user",
			method: http.MethodPost,
			token:  "test",
			body: body{
				Login:    "test",
				Password: "test",
			},
			errorAuthUser: fmt.Errorf("test"),
			wantCode:      http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := NewMockservice(ctrl)

			h := &Handler{
				service: mockService,
				log:     log,
			}

			mockService.EXPECT().AuthUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(tt.token, tt.errorAuthUser).AnyTimes()

			w := httptest.NewRecorder()

			var reqBody io.Reader
			if !tt.flagBody {
				buf := bytes.NewBuffer(nil)
				err := json.NewEncoder(buf).Encode(tt.body)
				if err != nil {
					return
				}
				reqBody = buf
			} else {
				reqBody = nil
			}
			r := httptest.NewRequest(tt.method, "/auth", reqBody)
			h.Auth(w, r)
			if w.Code != tt.wantCode {
				t.Errorf("Handler.Register() = %v, want %v", w.Code, tt.wantCode)
			}

		})
	}
}

func TestHandler_Logout(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockservice(ctrl)

	type fields struct {
		service    service
		log        *slog.Logger
		adminToken string
	}

	tests := []struct {
		name    string
		login   string
		method  string
		cookies []*http.Cookie
		fields  fields
		want    int
	}{

		{
			name:   "success_logout",
			login:  "test",
			method: http.MethodDelete,
			fields: fields{
				service:    mockService,
				log:        log,
				adminToken: "test",
			},
			want: http.StatusOK,
		},
		{
			name:   "err_method",
			method: http.MethodGet,
			fields: fields{
				service:    mockService,
				log:        log,
				adminToken: "test",
			},
			want: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				service:    tt.fields.service,
				log:        tt.fields.log,
				adminToken: tt.fields.adminToken,
			}
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, "/logout", nil)
			r = r.WithContext(context.WithValue(r.Context(), middleware.NameLogin, tt.login))
			r.AddCookie(&http.Cookie{
				Name:  auth.NameCookie,
				Value: "test-token",
			})
			h.Logout(w, r)
			if w.Code != tt.want {
				t.Errorf("Handler.Logout() = %v, want %v", w.Code, tt.want)
			}
		})
	}
}
