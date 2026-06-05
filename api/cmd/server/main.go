package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"xlab-blog/api/internal/assets"
	"xlab-blog/api/internal/auth"
	"xlab-blog/api/internal/config"
	"xlab-blog/api/internal/db"
	httpapi "xlab-blog/api/internal/http"
	"xlab-blog/api/internal/search"
	"xlab-blog/api/internal/users"
)

const shutdownTimeout = 10 * time.Second

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.OpenPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := db.RunMigrations(ctx, pool, "migrations"); err != nil {
		logger.Error("run migrations", "error", err)
		os.Exit(1)
	}

	userRepo := users.NewSQLRepository(pool)
	tokens := auth.NewTokenService(cfg.JWTSecret, auth.DefaultTokenTTL)
	authService := auth.NewService(userRepo, tokens)
	if err := authService.SeedAdmin(ctx, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		logger.Error("seed admin", "error", err)
		os.Exit(1)
	}

	assetRepo := assets.NewSQLRepository(pool, cfg.AssetPublicBaseURL)
	assetStorage := assets.NewLocalStorage(cfg.AssetsDir)
	assetService := assets.NewService(assetRepo, assetStorage)

	var embeddingProvider search.EmbeddingProvider
	if cfg.DashScopeAPIKey != "" {
		embeddingProvider = search.NewQwenProvider(cfg.EmbeddingBaseURL, cfg.DashScopeAPIKey, cfg.EmbeddingModel, cfg.EmbeddingDimensions, nil)
	}
	searchService := search.NewService(search.NewSQLRepository(pool), embeddingProvider, cfg.EmbeddingModel, cfg.EmbeddingDimensions)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.NewRouter(httpapi.Dependencies{Pool: pool, AuthService: authService, Tokens: tokens, AssetService: assetService, SearchService: searchService}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("api server listening", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen and serve", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown server", "error", err)
	}
}
