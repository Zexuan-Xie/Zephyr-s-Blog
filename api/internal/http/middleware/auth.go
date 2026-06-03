package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"xlab-blog/api/internal/auth"
	httpapi "xlab-blog/api/internal/http"
	"xlab-blog/api/internal/users"
)

type UserLoader interface {
	GetByID(ctx context.Context, id uuid.UUID) (users.User, error)
}

type Authenticator struct {
	tokens *auth.TokenService
	loader UserLoader
}

type contextKey string

const userContextKey contextKey = "current_user"

func NewAuthenticator(tokens *auth.TokenService, loader UserLoader) *Authenticator {
	return &Authenticator{tokens: tokens, loader: loader}
}

func (a *Authenticator) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, ok := bearerToken(r)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		user, err := a.authenticate(r.Context(), tokenString)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r.WithContext(withCurrentUser(r.Context(), user)))
	})
}

func (a *Authenticator) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, ok := bearerToken(r)
		if !ok {
			httpapi.WriteError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		user, err := a.authenticate(r.Context(), tokenString)
		if err != nil {
			httpapi.WriteError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		next.ServeHTTP(w, r.WithContext(withCurrentUser(r.Context(), user)))
	})
}

func (a *Authenticator) RequireAdmin(next http.Handler) http.Handler {
	return a.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := CurrentUser(r.Context())
		if !ok || user.Role != users.RoleAdmin {
			httpapi.WriteError(w, http.StatusForbidden, "admin role required")
			return
		}
		next.ServeHTTP(w, r)
	}))
}

func CurrentUser(ctx context.Context) (users.User, bool) {
	user, ok := ctx.Value(userContextKey).(users.User)
	return user, ok
}

func withCurrentUser(ctx context.Context, user users.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func (a *Authenticator) authenticate(ctx context.Context, tokenString string) (users.User, error) {
	claims, err := a.tokens.Parse(tokenString)
	if err != nil {
		return users.User{}, err
	}
	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return users.User{}, err
	}
	return a.loader.GetByID(ctx, id)
}

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", false
	}
	prefix := "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return token, token != ""
}
