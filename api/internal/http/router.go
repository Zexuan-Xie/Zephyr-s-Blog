package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"xlab-blog/api/internal/auth"
	"xlab-blog/api/internal/http/handlers"
	"xlab-blog/api/internal/http/middleware"
)

type Dependencies struct {
	Pool        *pgxpool.Pool
	AuthService *auth.Service
	Tokens      *auth.TokenService
}

func NewRouter(deps Dependencies) http.Handler {
	router := chi.NewRouter()
	router.Route("/api", func(api chi.Router) {
		healthHandler := handlers.NewHealthHandler(deps.Pool)
		api.Get("/health", healthHandler.Check)

		if deps.AuthService != nil && deps.Tokens != nil {
			authHandler := handlers.NewAuthHandler(deps.AuthService)
			authMiddleware := middleware.NewAuthenticator(deps.Tokens, deps.AuthService)
			api.Post("/auth/register", authHandler.Register)
			api.Post("/auth/login", authHandler.Login)
			api.With(authMiddleware.RequireAuth).Get("/auth/me", authHandler.Me)
		}
	})
	return router
}
