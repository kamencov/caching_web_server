package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var (
	reLogin = regexp.MustCompile(`^[A-Za-z0-9]{8,}$`)

	ErrorLogin    = errors.New("invalid login")
	ErrorPassword = errors.New("invalid password")
)

const NameCookie = "token"

//go:generate mockgen -source=service.go -destination=service_mock.go -package=auth
type storage interface {
	SaveUser(ctx context.Context, login, password string) error
	GetHashPass(ctx context.Context, login string) (string, error)
}

type Service struct {
	storage   storage
	log       *slog.Logger
	tokenSalt string
}

// NewService - создает новый сервис
func NewService(storage storage, log *slog.Logger, tokenSalt string) *Service {
	return &Service{
		storage:   storage,
		log:       log,
		tokenSalt: tokenSalt,
	}
}

// RegisterUser - регистрирует пользователя
func (s *Service) RegisterUser(ctx context.Context, login, password string) error {
	err := validateLogin(login)
	if err != nil {
		s.log.Error("RegisterUser", "invalid login", login)
		return err
	}

	err = validatePassword(password)
	if err != nil {
		s.log.Error("RegisterUser", "invalid password", password)
		return err
	}

	hash, err := HashPassword(password)
	if err != nil {
		s.log.Error("RegisterUser", "failed to hash password", err)
		return err
	}

	err = s.storage.SaveUser(ctx, login, hash)
	if err != nil {
		s.log.Error("RegisterUser", "failed to create user", err)
		return err
	}

	return nil
}

// login - условия
// Минимальная длина 8, латиница и цифры
// validateLogin - валидация логина
func validateLogin(login string) error {
	if !reLogin.MatchString(login) {
		return ErrorLogin
	}

	return nil
}

// pswd - условия
// минимальная длина 8,
// минимум 2 буквы в разных регистрах
// минимум 1 цифра
// минимум 1 символ (не буква и не цифра)
// validatePassword - валидация пароля
func validatePassword(password string) error {
	if len(password) < 8 {
		return ErrorPassword
	}
	var uppercase, lowercase, digit, special bool
	for _, char := range password {
		if char >= 'A' && char <= 'Z' {
			uppercase = true
		} else if char >= 'a' && char <= 'z' {
			lowercase = true
		} else if char >= '0' && char <= '9' {
			digit = true
		} else {
			special = true
		}
	}
	if !uppercase || !lowercase || !digit || !special {
		return ErrorPassword
	}
	return nil
}

type passHash struct {
	Salt string `json:"salt"`
	Sum  string `json:"sum"`
}

// HashPassword - хеширует пароль
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	h := sha256.Sum256(append(salt, []byte(password)...))
	ph := passHash{
		Salt: base64.StdEncoding.EncodeToString(salt),
		Sum:  base64.StdEncoding.EncodeToString(h[:]),
	}

	result, err := json.Marshal(ph)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

// AuthUser - авторизует пользователя
func (s *Service) AuthUser(ctx context.Context, login, password string) (string, error) {
	hashPass, err := s.storage.GetHashPass(ctx, login)
	if err != nil {
		return "", err
	}
	if !checkPassword(hashPass, password) {
		return "", ErrorPassword
	}

	token, err := s.generateToken(login)
	if err != nil {
		return "", err
	}
	return token, nil

}

// CheckPassword - проверяет пароль
func checkPassword(hashJSON, pw string) bool {
	var ph passHash
	if err := json.Unmarshal([]byte(hashJSON), &ph); err != nil {
		return false
	}
	salt, err := base64.StdEncoding.DecodeString(ph.Salt)
	if err != nil {
		return false
	}
	h := sha256.Sum256(append(salt, []byte(pw)...))
	return ph.Sum == base64.StdEncoding.EncodeToString(h[:])
}

// generateToken - генерирует токен
func (s *Service) generateToken(login string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": login,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString([]byte(s.tokenSalt))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// VerifyToken - проверяет токен
func (s *Service) VerifyToken(token string) (string, error) {
	login, err := s.checkToken(token)
	if err != nil {
		return "", err
	}
	return login, nil
}

// checkToken - проверяет токен
func (s *Service) checkToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.tokenSalt), nil
	})

	if err != nil {
		return "", fmt.Errorf("invalid token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return "", fmt.Errorf("token expired")
		}
	} else {
		return "", fmt.Errorf("exp claim missing")
	}

	login, ok := claims["login"].(string)
	if !ok {
		return "", fmt.Errorf("login not found in token claims")
	}

	return login, nil
}
