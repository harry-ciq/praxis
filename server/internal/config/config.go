package config

import (
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// Load .env file if present (silently ignore if missing)
	godotenv.Load()
}

type Config struct {
	ServerPort         string
	DatabaseURL        string
	RedisURL           string
	JWTSecret          string
	GitHubClientID     string
	GitHubClientSecret string
	FrontendURL        string
	Environment        string
}

func Load() *Config {
	return &Config{
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://praxis:praxis_dev@localhost:5432/praxis?sslmode=disable"),
		RedisURL:           getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:3000"),
		Environment:        getEnv("ENVIRONMENT", "development"),
	}
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
