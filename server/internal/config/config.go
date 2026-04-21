package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func init() {
	// Load .env if present. Use Overload so the file wins over any pre-existing
	// shell env vars — common gotcha when launching from IDEs or desktop apps
	// that export their own empty defaults (e.g. ANTHROPIC_API_KEY="").
	_ = godotenv.Overload()
}

type Config struct {
	ServerPort         string
	DatabaseURL        string
	RedisURL           string
	JWTSecret          string
	GitHubClientID     string
	GitHubClientSecret string
	GoogleClientID     string
	GoogleClientSecret string
	GitHubWebhookSecret string
	AnthropicAPIKey    string
	SmartFeedModel     string
	SmartFeedCacheTTL  int // seconds
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
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GitHubWebhookSecret: getEnv("GITHUB_WEBHOOK_SECRET", ""),
		AnthropicAPIKey:    getEnv("ANTHROPIC_API_KEY", ""),
		SmartFeedModel:     getEnv("SMART_FEED_MODEL", "claude-haiku-4-5"),
		SmartFeedCacheTTL:  getEnvInt("SMART_FEED_CACHE_TTL", 900),
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

func getEnvInt(key string, fallback int) int {
	if val, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return fallback
}
