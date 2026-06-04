package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"xlab-blog/api/internal/auth"
	"xlab-blog/api/internal/http/handlers"
	"xlab-blog/api/internal/http/middleware"
	"xlab-blog/api/internal/tree"
)

type Dependencies struct {
	Pool             *pgxpool.Pool
	AuthService      *auth.Service
	Tokens           *auth.TokenService
	TreeService      *tree.Service
	LifecycleService *tree.LifecycleService
}

func NewRouter(deps Dependencies) http.Handler {
	router := chi.NewRouter()
	router.Route("/api", func(api chi.Router) {
		healthHandler := handlers.NewHealthHandler(deps.Pool)
		api.Get("/health", healthHandler.Check)

		treeService := deps.TreeService
		if treeService == nil && deps.Pool != nil {
			treeService = tree.NewService(tree.NewSQLRepository(deps.Pool))
		}
		if treeService != nil {
			treeHandler := handlers.NewTreeHandler(treeService)
			api.Get("/tree", treeHandler.Root)
			api.Get("/tree/resolve", treeHandler.Resolve)
			api.Get("/tree/{node_id}/children", treeHandler.Children)
		}

		if deps.AuthService != nil && deps.Tokens != nil {
			authHandler := handlers.NewAuthHandler(deps.AuthService)
			authMiddleware := middleware.NewAuthenticator(deps.Tokens, deps.AuthService)
			api.Post("/auth/register", authHandler.Register)
			api.Post("/auth/login", authHandler.Login)
			api.With(authMiddleware.RequireAuth).Get("/auth/me", authHandler.Me)

			lifecycleService := deps.LifecycleService
			if lifecycleService == nil && deps.Pool != nil {
				lifecycleService = tree.NewLifecycleService(tree.NewSQLRepository(deps.Pool))
			}
			if lifecycleService != nil {
				lifecycleHandler := handlers.NewTreeLifecycleHandler(lifecycleService)
				api.Route("/admin", func(admin chi.Router) {
					admin.Use(authMiddleware.RequireAdmin)
					admin.Delete("/nodes/{node_id}", lifecycleHandler.DeleteNode)
					admin.Put("/files/{file_id}/content", lifecycleHandler.UpsertFileContent)
					admin.Post("/files/{file_id}/publish", lifecycleHandler.PublishFile)
					admin.Post("/files/{file_id}/unpublish", lifecycleHandler.UnpublishFile)
				})
			}
		}
	})
	return router
}
