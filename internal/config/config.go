package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPPort      string
	PostgresDSN   string
	RabbitDSN     string
	JWTSecret     string
	TokenTTL      time.Duration
	AdminUser     string
	AdminPassword string
}

func Load() Config {
	return Config{
		HTTPPort:      valueOrDefault("HTTP_PORT", "8080"),
		PostgresDSN:   valueOrDefault("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/users?sslmode=disable"),
		RabbitDSN:     valueOrDefault("RABBITMQ_DSN", "amqp://guest:guest@localhost:5672/"),
		JWTSecret:     valueOrDefault("JWT_SECRET", "supersecret"),
		TokenTTL:      durationOrDefault("TOKEN_TTL_MINUTES", time.Hour),
		AdminUser:     valueOrDefault("ADMIN_USERNAME", "admin"),
		AdminPassword: valueOrDefault("ADMIN_PASSWORD", "changeme"),
	}
}

func (c Config) Addr() string {
	return ":" + c.HTTPPort
}

func valueOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func durationOrDefault(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if minutes, err := strconv.Atoi(v); err == nil && minutes > 0 {
			return time.Duration(minutes) * time.Minute
		}
	}
	return def
}
