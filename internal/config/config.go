package config

import (
	"flag"
	"log/slog"
	"os"
	"strings"
)

// Config contém configurações principais da aplicação.
type Config struct {
	Port              string
	ExpectedAuthToken string
	Logger            *slog.Logger
}

// New carrega a configuração a partir de flags e variáveis de ambiente.
func New() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	expectedToken := strings.TrimSpace(os.Getenv("EXPECTED_AUTH_TOKEN"))
	if expectedToken == "" {
		return nil, ErrMissingAuthToken
	}

	// Flags opcionais (mantidas para extensão futura)
	_ = flag.CommandLine.Parse([]string{})

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return &Config{Port: port, ExpectedAuthToken: expectedToken, Logger: logger}, nil
}

// ErrMissingAuthToken indica ausência de token.
var ErrMissingAuthToken = &ConfigError{"EXPECTED_AUTH_TOKEN não definido"}

// ConfigError erro simples de config.
type ConfigError struct{ Msg string }

func (e *ConfigError) Error() string { return e.Msg }
