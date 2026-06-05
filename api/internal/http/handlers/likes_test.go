package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"xlab-blog/api/internal/likes"
	"xlab-blog/api/internal/users"
)

func TestLikeHandlerRoutes(t *testing.T) {
	fileID := uuid.New()
	commentID := uuid.New()
	reader := users.User{ID: uuid.New(), Email: "reader@example.com", Role: users.RoleReader}
	service := &fakeLikeService{state: likes.State{Liked: true, LikeCount: 3}}
	handler := NewLikeHandler(service)
	router := chi.NewRouter()
	router.Put("/files/{file_id}/like", withCurrentUser(reader, handler.LikeFile))
	router.Delete("/files/{file_id}/like", withCurrentUser(reader, handler.UnlikeFile))
	router.Put("/comments/{comment_id}/like", withCurrentUser(reader, handler.LikeComment))
	router.Delete("/comments/{comment_id}/like", withCurrentUser(reader, handler.UnlikeComment))

	tests := []struct {
		method string
		path   string
		status int
		want   string
	}{
		{method: http.MethodPut, path: "/files/" + fileID.String() + "/like", status: http.StatusOK, want: `"like_count":3`},
		{method: http.MethodDelete, path: "/files/" + fileID.String() + "/like", status: http.StatusOK, want: `"liked":true`},
		{method: http.MethodPut, path: "/comments/" + commentID.String() + "/like", status: http.StatusOK, want: `"like_count":3`},
		{method: http.MethodDelete, path: "/comments/" + commentID.String() + "/like", status: http.StatusOK, want: `"liked":true`},
		{method: http.MethodPut, path: "/files/not-a-uuid/like", status: http.StatusBadRequest, want: `"error":"invalid file_id"`},
	}

	for _, tt := range tests {
		request := httptest.NewRequest(tt.method, tt.path, nil)
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

func TestLikeHandlerRequiresAuth(t *testing.T) {
	fileID := uuid.New()
	handler := NewLikeHandler(&fakeLikeService{})
	router := chi.NewRouter()
	router.Put("/files/{file_id}/like", handler.LikeFile)

	request := httptest.NewRequest(http.MethodPut, "/files/"+fileID.String()+"/like", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestLikeHandlerMapsDeletedTargetConflict(t *testing.T) {
	commentID := uuid.New()
	reader := users.User{ID: uuid.New(), Email: "reader@example.com", Role: users.RoleReader}
	handler := NewLikeHandler(&fakeLikeService{err: likes.ErrTargetDeleted})
	router := chi.NewRouter()
	router.Put("/comments/{comment_id}/like", withCurrentUser(reader, handler.LikeComment))

	request := httptest.NewRequest(http.MethodPut, "/comments/"+commentID.String()+"/like", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusConflict, response.Body.String())
	}
}

type fakeLikeService struct {
	state likes.State
	err   error
}

func (f *fakeLikeService) LikeFile(context.Context, uuid.UUID, uuid.UUID) (likes.State, error) {
	return f.state, f.err
}

func (f *fakeLikeService) UnlikeFile(context.Context, uuid.UUID, uuid.UUID) (likes.State, error) {
	return f.state, f.err
}

func (f *fakeLikeService) LikeComment(context.Context, uuid.UUID, uuid.UUID) (likes.State, error) {
	return f.state, f.err
}

func (f *fakeLikeService) UnlikeComment(context.Context, uuid.UUID, uuid.UUID) (likes.State, error) {
	return f.state, f.err
}
