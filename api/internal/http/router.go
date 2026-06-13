package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"xlab-blog/api/internal/assets"
	"xlab-blog/api/internal/auth"
	"xlab-blog/api/internal/comments"
	"xlab-blog/api/internal/http/handlers"
	"xlab-blog/api/internal/http/middleware"
	"xlab-blog/api/internal/http/respond"
	"xlab-blog/api/internal/likes"
	"xlab-blog/api/internal/search"
	"xlab-blog/api/internal/tree"
)

type Dependencies struct {
	Pool             *pgxpool.Pool
	AuthService      *auth.Service
	Tokens           *auth.TokenService
	TreeService      *tree.Service
	RecentService    handlers.RecentTreeService
	LifecycleService *tree.LifecycleService
	AdminService     *tree.AdminService
	CommentService   *comments.Service
	LikeService      *likes.Service
	AssetService     *assets.Service
	SearchService    handlers.SearchService
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
		recentService := deps.RecentService
		if recentService == nil && deps.Pool != nil {
			recentService = tree.NewRecentService(tree.NewSQLRepository(deps.Pool))
		}
		if recentService != nil {
			api.Get("/recent", handlers.RecentFiles(recentService))
		}

		assetService := deps.AssetService
		if assetService != nil {
			assetHandler := handlers.NewAssetHandler(assetService)
			api.Get("/assets/{asset_id}/{filename}", assetHandler.ServePublished)
		}

		searchService := deps.SearchService
		if searchService == nil && deps.Pool != nil {
			searchService = search.NewService(search.NewSQLRepository(deps.Pool), nil, "text-embedding-v4", 1024)
		}
		if searchService != nil {
			searchHandler := handlers.NewSearchHandler(searchService)
			api.Get("/search", searchHandler.Search)
		}

		var authMiddleware *middleware.Authenticator
		if deps.AuthService != nil && deps.Tokens != nil {
			authHandler := handlers.NewAuthHandler(deps.AuthService)
			authMiddleware = middleware.NewAuthenticator(deps.Tokens, deps.AuthService)
			api.Post("/auth/register", authHandler.Register)
			api.Post("/auth/login", authHandler.Login)
			api.With(authMiddleware.RequireAuth).Get("/auth/me", authHandler.Me)
		}

		commentService := deps.CommentService
		likeService := deps.LikeService
		if deps.Pool != nil && (commentService == nil || likeService == nil) {
			if commentService == nil {
				commentService = comments.NewService(comments.NewSQLRepository(deps.Pool))
			}
			if likeService == nil {
				likeService = likes.NewService(likes.NewSQLRepository(deps.Pool))
			}
		}
		if commentService != nil {
			commentHandler := handlers.NewCommentHandler(commentService)
			if authMiddleware != nil {
				api.With(authMiddleware.OptionalAuth).Get("/files/{file_id}/comments", commentHandler.Thread)
				api.With(authMiddleware.RequireAuth).Post("/files/{file_id}/comments", commentHandler.Create)
				api.With(authMiddleware.RequireAuth).Delete("/comments/{comment_id}", commentHandler.Delete)
			} else {
				api.Get("/files/{file_id}/comments", commentHandler.Thread)
			}
		}
		if likeService != nil && authMiddleware != nil {
			likeHandler := handlers.NewLikeHandler(likeService)
			api.With(authMiddleware.RequireAuth).Put("/files/{file_id}/like", likeHandler.LikeFile)
			api.With(authMiddleware.RequireAuth).Delete("/files/{file_id}/like", likeHandler.UnlikeFile)
			api.With(authMiddleware.RequireAuth).Put("/comments/{comment_id}/like", likeHandler.LikeComment)
			api.With(authMiddleware.RequireAuth).Delete("/comments/{comment_id}/like", likeHandler.UnlikeComment)
		}

		if authMiddleware != nil {
			lifecycleService := deps.LifecycleService
			adminService := deps.AdminService
			if deps.Pool != nil && (lifecycleService == nil || adminService == nil) {
				repo := tree.NewSQLRepository(deps.Pool)
				if lifecycleService == nil {
					lifecycleService = tree.NewLifecycleService(repo)
				}
				if adminService == nil {
					adminService = tree.NewAdminService(repo, lifecycleService)
				}
			}
			if lifecycleService == nil && adminService == nil && assetService == nil && searchService == nil {
				api.Route("/admin", func(admin chi.Router) {
					admin.Use(authMiddleware.RequireAdmin)
					admin.Get("/preview/{file_id}", func(w http.ResponseWriter, r *http.Request) {
						respond.Error(w, http.StatusNotFound, "file content not found")
					})
				})
			}
			if lifecycleService != nil || adminService != nil || assetService != nil || searchService != nil {
				var lifecycleHandler *handlers.TreeLifecycleHandler
				if lifecycleService != nil {
					lifecycleHandler = handlers.NewTreeLifecycleHandler(lifecycleService)
				}
				var assetHandler *handlers.AssetHandler
				if assetService != nil {
					assetHandler = handlers.NewAssetHandler(assetService)
				}
				var searchHandler *handlers.SearchHandler
				if searchService != nil {
					searchHandler = handlers.NewSearchHandler(searchService)
				}
				api.Route("/admin", func(admin chi.Router) {
					admin.Use(authMiddleware.RequireAdmin)
					if adminService != nil {
						adminHandler := handlers.NewAdminNodeHandler(adminService)
						admin.Get("/tree", adminHandler.AdminTree)
						admin.Post("/nodes", adminHandler.CreateNode)
						admin.Get("/nodes/{node_id}", adminHandler.GetNode)
						admin.Patch("/nodes/{node_id}", adminHandler.UpdateNode)
						admin.Put("/nodes/{parent_id}/children/order", adminHandler.ReorderChildren)
						admin.Post("/nodes/{node_id}/move-preview", adminHandler.PreviewMove)
						admin.Post("/nodes/{node_id}/move", adminHandler.MoveNode)
					}
					if lifecycleService != nil {
						admin.Delete("/nodes/{node_id}", lifecycleHandler.DeleteNode)
						admin.Get("/files/{file_id}/content", lifecycleHandler.GetFileContent)
						admin.Put("/files/{file_id}/content", lifecycleHandler.UpsertFileContent)
						admin.Post("/files/{file_id}/previous/restore", lifecycleHandler.RestorePreviousContent)
						admin.Get("/files/{file_id}/publish-summary", lifecycleHandler.PublishSummary)
						admin.Post("/files/{file_id}/publish", lifecycleHandler.PublishFile)
						admin.Post("/files/{file_id}/unpublish", lifecycleHandler.UnpublishFile)
						admin.Get("/preview/{file_id}", lifecycleHandler.DraftPreview)
						admin.Get("/files/{file_id}/assets", lifecycleHandler.FileAssetState)
					} else {
						admin.Get("/preview/{file_id}", func(w http.ResponseWriter, r *http.Request) {
							respond.Error(w, http.StatusNotFound, "file content not found")
						})
					}
					if assetHandler != nil {
						admin.Post("/files/{file_id}/assets", assetHandler.Upload)
						admin.Get("/assets/{asset_id}/{filename}", assetHandler.ServeDraft)
						admin.Delete("/assets/{asset_id}", assetHandler.Delete)
					}
					if searchHandler != nil {
						admin.Post("/files/{file_id}/refresh-embedding", searchHandler.RefreshEmbedding)
						admin.Post("/search-index/rebuild", searchHandler.Rebuild)
					}
				})
			}
		}
	})
	return router
}
