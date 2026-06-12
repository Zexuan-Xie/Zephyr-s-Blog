package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"xlab-blog/api/internal/auth"
	"xlab-blog/api/internal/users"
)

type authHandlerUserRepository struct {
	createErr error
	findErr   error
}

func (r *authHandlerUserRepository) CreateReader(context.Context, string, string, *string) (users.User, error) {
	if r.createErr != nil {
		return users.User{}, r.createErr
	}
	return users.User{
		ID:           uuid.New(),
		Email:        "reader@example.com",
		PasswordHash: mustPasswordHash(),
		Role:         users.RoleReader,
		Provider:     "local",
		CreatedAt:    time.Now(),
	}, nil
}

func (r *authHandlerUserRepository) FindByEmail(context.Context, string) (users.User, error) {
	if r.findErr != nil {
		return users.User{}, r.findErr
	}
	return users.User{
		ID:           uuid.New(),
		Email:        "reader@example.com",
		PasswordHash: mustPasswordHash(),
		Role:         users.RoleReader,
		Provider:     "local",
		CreatedAt:    time.Now(),
	}, nil
}

func (*authHandlerUserRepository) FindByID(context.Context, uuid.UUID) (users.User, error) {
	return users.User{}, users.ErrUserNotFound
}

func (*authHandlerUserRepository) UpsertAdmin(context.Context, string, string) (users.User, error) {
	return users.User{}, errors.New("not implemented")
}

func TestAuthHandlerPreservesKnownErrorMappings(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		repo       *authHandlerUserRepository
		handler    func(*AuthHandler, http.ResponseWriter, *http.Request)
		wantStatus int
		wantError  string
	}{
		{
			name:       "register invalid JSON",
			method:     http.MethodPost,
			path:       "/auth/register",
			body:       "{",
			repo:       &authHandlerUserRepository{},
			handler:    (*AuthHandler).Register,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid JSON body",
		},
		{
			name:       "register invalid email",
			method:     http.MethodPost,
			path:       "/auth/register",
			body:       `{"email":"invalid","password":"long-password"}`,
			repo:       &authHandlerUserRepository{},
			handler:    (*AuthHandler).Register,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid email",
		},
		{
			name:       "register short password",
			method:     http.MethodPost,
			path:       "/auth/register",
			body:       `{"email":"reader@example.com","password":"short"}`,
			repo:       &authHandlerUserRepository{},
			handler:    (*AuthHandler).Register,
			wantStatus: http.StatusBadRequest,
			wantError:  "password must be at least 8 characters",
		},
		{
			name:       "register conflict",
			method:     http.MethodPost,
			path:       "/auth/register",
			body:       `{"email":"reader@example.com","password":"long-password"}`,
			repo:       &authHandlerUserRepository{createErr: users.ErrEmailExists},
			handler:    (*AuthHandler).Register,
			wantStatus: http.StatusConflict,
			wantError:  "email already exists",
		},
		{
			name:       "login invalid credentials",
			method:     http.MethodPost,
			path:       "/auth/login",
			body:       `{"email":"reader@example.com","password":"wrong-password"}`,
			repo:       &authHandlerUserRepository{findErr: users.ErrUserNotFound},
			handler:    (*AuthHandler).Login,
			wantStatus: http.StatusUnauthorized,
			wantError:  "invalid credentials",
		},
		{
			name:       "login invalid JSON",
			method:     http.MethodPost,
			path:       "/auth/login",
			body:       "{",
			repo:       &authHandlerUserRepository{},
			handler:    (*AuthHandler).Login,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid JSON body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewAuthHandler(auth.NewService(tt.repo, auth.NewTokenService("secret", time.Hour)))
			request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			response := httptest.NewRecorder()

			tt.handler(handler, response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", response.Code, tt.wantStatus, response.Body.String())
			}
			if !strings.Contains(response.Body.String(), `"error":"`+tt.wantError) {
				t.Fatalf("body = %s, want error %q", response.Body.String(), tt.wantError)
			}
		})
	}
}

func TestAuthHandlerSanitizesUnexpectedErrors(t *testing.T) {
	internalErr := errors.New("internal database topology sentinel")
	tests := []struct {
		name    string
		body    string
		repo    *authHandlerUserRepository
		handler func(*AuthHandler, http.ResponseWriter, *http.Request)
	}{
		{
			name:    "register",
			body:    `{"email":"reader@example.com","password":"long-password"}`,
			repo:    &authHandlerUserRepository{createErr: internalErr},
			handler: (*AuthHandler).Register,
		},
		{
			name:    "login",
			body:    `{"email":"reader@example.com","password":"long-password"}`,
			repo:    &authHandlerUserRepository{findErr: internalErr},
			handler: (*AuthHandler).Login,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewAuthHandler(auth.NewService(tt.repo, auth.NewTokenService("secret", time.Hour)))
			request := httptest.NewRequest(http.MethodPost, "/auth/"+tt.name, strings.NewReader(tt.body))
			response := httptest.NewRecorder()

			tt.handler(handler, response, request)

			if response.Code != http.StatusInternalServerError {
				t.Fatalf("status = %d, want 500; body=%s", response.Code, response.Body.String())
			}
			if !strings.Contains(response.Body.String(), `"error":"internal server error"`) {
				t.Fatalf("body = %s, want generic internal server error", response.Body.String())
			}
			if strings.Contains(response.Body.String(), internalErr.Error()) {
				t.Fatalf("body exposes internal error: %s", response.Body.String())
			}
		})
	}
}

func mustPasswordHash() string {
	hash, err := auth.HashPassword("long-password")
	if err != nil {
		panic(err)
	}
	return hash
}
