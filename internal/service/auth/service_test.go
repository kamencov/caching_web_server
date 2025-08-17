package auth

import (
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
)

var errStorage = errors.New("storage error")

func TestNewService(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockstorage(ctrl)

	newService := NewService(mockStorage, log, "")

	if newService == nil {
		t.Errorf("NewService() = %v, want %v", newService, "not nil")
	}
}

func TestService_RegisterUser(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	tests := []struct {
		name       string
		login      string
		password   string
		errStorage error
		wantErr    error
	}{
		{
			name:     "success_register",
			login:    "Document",
			password: "DocumenT1@",
			wantErr:  nil,
		},
		{
			name:     "err_invalid_login",
			login:    "Docu",
			password: "DocumenT1",
			wantErr:  ErrorLogin,
		},
		{
			name:     "err_invalid_password_v1",
			login:    "Document",
			password: "Docum",
			wantErr:  ErrorPassword,
		},
		{
			name:     "err_invalid_password_v2",
			login:    "Document",
			password: "Document1",
			wantErr:  ErrorPassword,
		},
		{
			name:       "err_storage",
			login:      "Document",
			password:   "DocumenT1@",
			errStorage: errStorage,
			wantErr:    errStorage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := NewMockstorage(ctrl)
			mockStorage.EXPECT().SaveUser(gomock.Any(), tt.login, gomock.Any()).Return(tt.errStorage).AnyTimes()
			s := &Service{
				storage: mockStorage,
				log:     log,
			}
			if err := s.RegisterUser(nil, tt.login, tt.password); !errors.Is(err, tt.wantErr) {
				t.Errorf("Service.RegisterUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_AuthUser(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	hpw, err := HashPassword("DocumenT1@")
	if err != nil {
		log.Error("TestService_AuthUser", "failed to hash password", err)
	}

	tests := []struct {
		name       string
		login      string
		password   string
		hpw        string
		errStorage error
		wantErr    error
	}{
		{
			name:     "success_auth",
			login:    "Document",
			password: "DocumenT1@",
			hpw:      hpw,
			wantErr:  nil,
		},
		{
			name:       "error_get_hash_pass",
			errStorage: errStorage,
			wantErr:    errStorage,
		},
		{
			name:    "bad_hash_pass",
			hpw:     "bad_hash_pass",
			wantErr: ErrorPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			newMockStorage := NewMockstorage(ctrl)
			newMockStorage.EXPECT().GetHashPass(gomock.Any(), gomock.Any()).Return(tt.hpw, tt.errStorage).AnyTimes()

			s := NewService(newMockStorage, log, "")

			_, err := s.AuthUser(nil, tt.login, tt.password)
			if !errors.Is(err, tt.wantErr) {
				log.Error("TestService_AuthUser", "failed to save user", err)
			}
		})
	}
}

func TestService_VerifyToken(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockstorage(ctrl)

	service := &Service{
		storage: mockStorage,
		log:     log,
	}

	token, err := service.generateToken("Document")
	if err != nil {
		log.Error("TestService_VerifyToken", "failed to generate token", err)
		return
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "success_verify_token",
			token:   token,
			wantErr: false,
		},
		{
			name:    "error_verify_token",
			token:   "bad_token",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.VerifyToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.VerifyToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
