package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"xlab-blog/api/internal/auth"
	"xlab-blog/api/internal/users"
)

type fakeLoader struct {
	users map[uuid.UUID]users.User
}

func (l fakeLoader) GetByID(_ context.Context, id uuid.UUID) (users.User, error) {
	user, ok := l.users[id]
	if !ok {
		return users.User{}, users.ErrUserNotFound
	}
	return user, nil
}

func TestRequireAuthRejectsMissingBearerToken(t *testing.T) {
	authenticator := NewAuthenticator(NewTokenService("secret", time.Hour), fakeLoader{})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	authenticator.RequireAuth(http.HandlerFunc(okHandler)).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuthAcceptsValidBearerToken(t *testing.T) {
	tokens := NewTokenService("secret", time.Hour)
	user := users.User{ID: uuid.New(), Email: "reader@example.com", Role: users.RoleReader}
	token, err := tokens.Issue(user)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	authenticator := NewAuthenticator(tokens, fakeLoader{users: map[uuid.UUID]users.User{user.ID: user}})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	authenticator.RequireAuth(http.HandlerFunc(okHandler)).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestRequireAdminRejectsReaderAndAcceptsAdmin(t *testing.T) {
	tokens := NewTokenService("secret", time.Hour)
	reader := users.User{ID: uuid.New(), Email: "reader@example.com", Role: users.RoleReader}
	admin := users.User{ID: uuid.New(), Email: "admin@example.com", Role: users.RoleAdmin}
	authenticator := NewAuthenticator(tokens, fakeLoader{users: map[uuid.UUID]users.User{reader.ID: reader, admin.ID: admin}})

	readerToken, err := tokens.Issue(reader)
	if err != nil {
		t.Fatalf("Issue(reader) error = %v", err)
	}
	readerResponse := httptest.NewRecorder()
	readerRequest := httptest.NewRequest(http.MethodGet, "/admin", nil)
	readerRequest.Header.Set("Authorization", "Bearer "+readerToken)
	authenticator.RequireAdmin(http.HandlerFunc(okHandler)).ServeHTTP(readerResponse, readerRequest)
	if readerResponse.Code != http.StatusForbidden {
		t.Fatalf("reader status = %d, want %d", readerResponse.Code, http.StatusForbidden)
	}

	adminToken, err := tokens.Issue(admin)
	if err != nil {
		t.Fatalf("Issue(admin) error = %v", err)
	}
	adminResponse := httptest.NewRecorder()
	adminRequest := httptest.NewRequest(http.MethodGet, "/admin", nil)
	adminRequest.Header.Set("Authorization", "Bearer "+adminToken)
	authenticator.RequireAdmin(http.HandlerFunc(okHandler)).ServeHTTP(adminResponse, adminRequest)
	if adminResponse.Code != http.StatusOK {
		t.Fatalf("admin status = %d, want %d", adminResponse.Code, http.StatusOK)
	}
}

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
