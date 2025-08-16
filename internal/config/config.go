package config

import (
	"os"
	"strings"
)

type Config struct {
	Port           string
	DatabaseURL    string
	RedisURL       string
	AllowedOrigins []string
	MaxMessages    int
	Environment    string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "1401"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://user:password@localhost/pollz_chat?sslmode=disable"),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379/0"),
		AllowedOrigins: strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000"), ","),
		MaxMessages:    1000,
		Environment:    getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}