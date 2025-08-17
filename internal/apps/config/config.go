package config

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	LogLevel   slog.Level `env:"LOG_LEVEL"`
	AdminToken string     `env:"ADMIN_TOKEN"`
	Addr       string     `env:"ADDR"`
	TokenSalt  string     `env:"TOKEN_SALT"`
	MaxSizFile int64      `env:"MAX_SIZE_FILE"`
}

func New() *Config {
	return &Config{}
}

func (c *Config) Parse() error {
	err := godotenv.Load(".env")
	if err != nil {
		return err
	}

	level := os.Getenv("LOG_LEVEL")
	logLevel, err := strconv.Atoi(level)
	if err != nil {
		return err
	}
	c.LogLevel = slog.Level(logLevel)
	c.AdminToken = os.Getenv("ADMIN_TOKEN")
	c.Addr = os.Getenv("ADDR")
	c.TokenSalt = os.Getenv("TOKEN_SALT")
	maxSize := os.Getenv("MAX_SIZE_FILE")
	maxSizeInt, err := strconv.Atoi(maxSize)
	if err != nil {
		return err
	}
	c.MaxSizFile = int64(maxSizeInt) << 20
	return nil
}
