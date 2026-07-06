// Package config loads application configuration from environment variables.
package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config is the root application configuration.
type Config struct {
	Env      string `envconfig:"APP_ENV" default:"development"`
	HTTP     HTTP
	Postgres Postgres
	Session  Session
	Log      Log
}

// HTTP holds the HTTP server configuration.
type HTTP struct {
	Host            string        `envconfig:"HTTP_HOST" default:"0.0.0.0"`
	Port            int           `envconfig:"HTTP_PORT" default:"8080"`
	ReadTimeout     time.Duration `envconfig:"HTTP_READ_TIMEOUT" default:"15s"`
	WriteTimeout    time.Duration `envconfig:"HTTP_WRITE_TIMEOUT" default:"15s"`
	ShutdownTimeout time.Duration `envconfig:"HTTP_SHUTDOWN_TIMEOUT" default:"10s"`
}

// Addr returns the host:port the server should listen on.
func (h HTTP) Addr() string {
	return fmt.Sprintf("%s:%d", h.Host, h.Port)
}

// Postgres holds the PostgreSQL connection configuration.
type Postgres struct {
	Host     string `envconfig:"POSTGRES_HOST" default:"localhost"`
	Port     int    `envconfig:"POSTGRES_PORT" default:"5432"`
	User     string `envconfig:"POSTGRES_USER" default:"todoapp"`
	Password string `envconfig:"POSTGRES_PASSWORD" default:"todoapp"`
	DB       string `envconfig:"POSTGRES_DB" default:"messenger"`
	SSLMode  string `envconfig:"POSTGRES_SSLMODE" default:"disable"`
	MaxConns int32  `envconfig:"POSTGRES_MAX_CONNS" default:"10"`
	MinConns int32  `envconfig:"POSTGRES_MIN_CONNS" default:"2"`
}

// DSN builds a libpq/pgx connection string from the configuration.
func (p Postgres) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.DB, p.SSLMode,
	)
}

// Session holds authentication session configuration.
type Session struct {
	TTL time.Duration `envconfig:"SESSION_TTL" default:"720h"`
}

// Log holds logger configuration.
type Log struct {
	Level string `envconfig:"LOG_LEVEL" default:"info"`
	File  string `envconfig:"LOG_FILE" default:"out/logs/app.log"`
}

// Load reads the configuration from environment variables.
func Load() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, fmt.Errorf("process env config: %w", err)
	}
	return cfg, nil
}
