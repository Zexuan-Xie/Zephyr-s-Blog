package config

import "testing"

func TestLoadRequiresDatabaseURLAndJWTSecret(t *testing.T) {
	clearEnv(t)
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_SECRET", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected required environment error")
	}
}

func TestLoadDefaultsAndExactEmbeddingSettings(t *testing.T) {
	clearEnv(t)
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/blog")
	t.Setenv("JWT_SECRET", "test-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTPAddr != DefaultHTTPAddr {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, DefaultHTTPAddr)
	}
	if cfg.EmbeddingModel != DefaultEmbeddingModel {
		t.Fatalf("EmbeddingModel = %q, want %q", cfg.EmbeddingModel, DefaultEmbeddingModel)
	}
	if cfg.EmbeddingDimensions != DefaultEmbeddingDimensions {
		t.Fatalf("EmbeddingDimensions = %d, want %d", cfg.EmbeddingDimensions, DefaultEmbeddingDimensions)
	}
}

func TestLoadRejectsPartialAdminSeed(t *testing.T) {
	clearEnv(t)
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/blog")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("ADMIN_EMAIL", "admin@example.com")

	_, err := Load()
	if err == nil {
		t.Fatal("expected partial admin seed error")
	}
}


func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"DATABASE_URL",
		"JWT_SECRET",
		"ADMIN_EMAIL",
		"ADMIN_PASSWORD",
		"HTTP_ADDR",
		"ASSETS_DIR",
		"EMBEDDING_PROVIDER",
		"DASHSCOPE_API_KEY",
		"EMBEDDING_BASE_URL",
		"EMBEDDING_MODEL",
		"EMBEDDING_DIMENSIONS",
	} {
		t.Setenv(key, "")
	}
}
