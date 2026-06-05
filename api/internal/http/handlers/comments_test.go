package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"xlab-blog/api/internal/auth"
	"xlab-blog/api/internal/comments"
	"xlab-blog/api/internal/http/middleware"
	"xlab-blog/api/internal/users"
)

func TestCommentHandlerRoutes(t *testing.T) {
	fileID := uuid.New()
	commentID := uuid.New()
	reader := users.User{ID: uuid.New(), Email: "reader@example.com", Role: users.RoleReader}
	service := &fakeCommentService{
		thread:  comments.Thread{FileID: fileID, Comments: []comments.Comment{}},
		comment: comments.Comment{ID: commentID, FileNodeID: fileID, User: comments.PublicUser{ID: reader.ID, DisplayName: "Reader"}, Body: "hello"},
	}
	handler := NewCommentHandler(service)
	router := chi.NewRouter()
	router.Get("/files/{file_id}/comments", handler.Thread)
	router.Post("/files/{file_id}/comments", withCurrentUser(reader, handler.Create))
	router.Delete("/comments/{comment_id}", withCurrentUser(reader, handler.Delete))

	tests := []struct {
		method string
		path   string
		body   string
		status int
		want   string
	}{
		{method: http.MethodGet, path: "/files/" + fileID.String() + "/comments", status: http.StatusOK, want: `"comments":[]`},
		{method: http.MethodPost, path: "/files/" + fileID.String() + "/comments", body: `{"body":"hello"}`, status: http.StatusCreated, want: `"body":"hello"`},
		{method: http.MethodDelete, path: "/comments/" + commentID.String(), status: http.StatusNoContent},
		{method: http.MethodGet, path: "/files/not-a-uuid/comments", status: http.StatusBadRequest, want: `"error":"invalid file_id"`},
	}

	for _, tt := range tests {
		request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != tt.status {
			t.Fatalf("%s %s status = %d, want %d; body=%s", tt.method, tt.path, response.Code, tt.status, response.Body.String())
		}
		if tt.want != "" && !strings.Contains(response.Body.String(), tt.want) {
			t.Fatalf("body = %s, want substring %s", response.Body.String(), tt.want)
		}
	}
}

func TestCommentHandlerRequiresAuthForWrites(t *testing.T) {
	fileID := uuid.New()
	commentID := uuid.New()
	handler := NewCommentHandler(&fakeCommentService{})
	router := chi.NewRouter()
	router.Post("/files/{file_id}/comments", handler.Create)
	router.Delete("/comments/{comment_id}", handler.Delete)

	for _, tt := range []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodPost, path: "/files/" + fileID.String() + "/comments", body: `{"body":"hello"}`},
		{method: http.MethodDelete, path: "/comments/" + commentID.String()},
	} {
		request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s status = %d, want %d", tt.method, tt.path, response.Code, http.StatusUnauthorized)
		}
	}
}

func TestCommentHandlerMapsServiceErrors(t *testing.T) {
	fileID := uuid.New()
	handler := NewCommentHandler(&fakeCommentService{err: comments.ErrInvalidCommentBody})
	router := chi.NewRouter()
	router.Post("/files/{file_id}/comments", withCurrentUser(users.User{ID: uuid.New(), Role: users.RoleReader}, handler.Create))

	request := httptest.NewRequest(http.MethodPost, "/files/"+fileID.String()+"/comments", strings.NewReader(`{"body":""}`))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusBadRequest, response.Body.String())
	}
}

type fakeCommentService struct {
	thread  comments.Thread
	comment comments.Comment
	err     error
}

func (f *fakeCommentService) Thread(context.Context, uuid.UUID, *uuid.UUID) (comments.Thread, error) {
	return f.thread, f.err
}

func (f *fakeCommentService) Create(context.Context, uuid.UUID, uuid.UUID, comments.CreateInput) (comments.Comment, error) {
	return f.comment, f.err
}

func (f *fakeCommentService) Delete(context.Context, uuid.UUID, users.User) (comments.Comment, error) {
	return f.comment, f.err
}

func withCurrentUser(user users.User, handler http.HandlerFunc) http.HandlerFunc {
	tokens := auth.NewTokenService("comment-handler-test-secret", time.Hour)
	token, err := tokens.Issue(user)
	if err != nil {
		panic(err)
	}
	authenticator := middleware.NewAuthenticator(tokens, &commentHandlerUserLoader{user: user})
	return func(w http.ResponseWriter, r *http.Request) {
		request := r.Clone(r.Context())
		request.Body = r.Body
		request.Header.Set("Authorization", "Bearer "+token)
		authenticator.RequireAuth(http.HandlerFunc(handler)).ServeHTTP(w, request)
	}
}

type commentHandlerUserLoader struct {
	user users.User
}

func (l *commentHandlerUserLoader) GetByID(context.Context, uuid.UUID) (users.User, error) {
	return l.user, nil
}
