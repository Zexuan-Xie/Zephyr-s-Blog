package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const (
	DefaultHTTPAddr            = ":8080"
	DefaultAssetsDir           = "/app/uploads"
	DefaultEmbeddingProvider   = "qwen"
	DefaultEmbeddingBaseURL    = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	DefaultEmbeddingModel      = "text-embedding-v4"
	DefaultEmbeddingDimensions = 1024
)

type Config struct {
	HTTPAddr            string
	DatabaseURL         string
	JWTSecret           string
	AdminEmail          string
	AdminPassword       string
	AssetsDir           string
	EmbeddingProvider   string
	DashScopeAPIKey     string
	EmbeddingBaseURL    string
	EmbeddingModel      string
	EmbeddingDimensions int
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		HTTPAddr:            getEnv("HTTP_ADDR", DefaultHTTPAddr),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		AdminEmail:          os.Getenv("ADMIN_EMAIL"),
		AdminPassword:       os.Getenv("ADMIN_PASSWORD"),
		AssetsDir:           getEnv("ASSETS_DIR", DefaultAssetsDir),
		EmbeddingProvider:   getEnv("EMBEDDING_PROVIDER", DefaultEmbeddingProvider),
		DashScopeAPIKey:     os.Getenv("DASHSCOPE_API_KEY"),
		EmbeddingBaseURL:    getEnv("EMBEDDING_BASE_URL", DefaultEmbeddingBaseURL),
		EmbeddingModel:      getEnv("EMBEDDING_MODEL", DefaultEmbeddingModel),
		EmbeddingDimensions: DefaultEmbeddingDimensions,
	}

	if raw := os.Getenv("EMBEDDING_DIMENSIONS"); raw != "" {
		dimensions, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, fmt.Errorf("parse EMBEDDING_DIMENSIONS: %w", err)
		}
		cfg.EmbeddingDimensions = dimensions
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return Config{}, errors.New("JWT_SECRET is required")
	}
	if (cfg.AdminEmail == "") != (cfg.AdminPassword == "") {
		return Config{}, errors.New("ADMIN_EMAIL and ADMIN_PASSWORD must be set together")
	}
	if cfg.EmbeddingProvider == DefaultEmbeddingProvider && cfg.EmbeddingModel != DefaultEmbeddingModel {
		return Config{}, fmt.Errorf("EMBEDDING_MODEL must be %q for qwen provider", DefaultEmbeddingModel)
	}
	if cfg.EmbeddingProvider == DefaultEmbeddingProvider && cfg.EmbeddingDimensions != DefaultEmbeddingDimensions {
		return Config{}, fmt.Errorf("EMBEDDING_DIMENSIONS must be %d for qwen provider", DefaultEmbeddingDimensions)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
