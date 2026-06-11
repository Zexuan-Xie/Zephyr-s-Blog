package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"xlab-blog/api/internal/tree"
)

func TestAdminNodeHandlerRoutes(t *testing.T) {
	nodeID := uuid.New()
	service := &fakeAdminNodeService{detail: tree.AdminNodeDetail{Node: tree.Node{ID: nodeID, Kind: tree.NodeKindDirectory, Name: "Notes", Slug: "notes", Path: "/notes"}}}
	handler := NewAdminNodeHandler(service)
	router := chi.NewRouter()
	router.Post("/nodes", handler.CreateNode)
	router.Get("/nodes/{node_id}", handler.GetNode)
	router.Patch("/nodes/{node_id}", handler.UpdateNode)

	tests := []struct {
		method string
		path   string
		body   string
		status int
	}{
		{method: http.MethodPost, path: "/nodes", body: `{"kind":"directory","name":"Notes","slug":"notes"}`, status: http.StatusCreated},
		{method: http.MethodGet, path: "/nodes/" + nodeID.String(), status: http.StatusOK},
		{method: http.MethodPatch, path: "/nodes/" + nodeID.String(), body: `{"name":"Renamed"}`, status: http.StatusOK},
		{method: http.MethodGet, path: "/nodes/not-a-uuid", status: http.StatusBadRequest},
	}

	for _, tt := range tests {
		request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != tt.status {
			t.Fatalf("%s %s status = %d, want %d; body=%s", tt.method, tt.path, response.Code, tt.status, response.Body.String())
		}
	}
}

func TestAdminNodeHandlerMapsConflicts(t *testing.T) {
	handler := NewAdminNodeHandler(&fakeAdminNodeService{err: tree.ErrDuplicateSlug})
	request := httptest.NewRequest(http.MethodPost, "/nodes", strings.NewReader(`{"kind":"directory","name":"Notes","slug":"notes"}`))
	response := httptest.NewRecorder()

	handler.CreateNode(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusConflict, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"error":"a node with this slug already exists under the selected parent"`) {
		t.Fatalf("body = %s, want actionable duplicate-slug error", response.Body.String())
	}
}

func TestAdminNodeHandlerMapsValidationErrorsPrecisely(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "missing name", err: tree.ErrNodeNameRequired, want: "node name is required"},
		{name: "missing slug", err: tree.ErrNodeSlugRequired, want: "node slug is required"},
		{name: "invalid kind", err: tree.ErrInvalidNodeKind, want: "node kind must be directory or file"},
		{name: "invalid slug", err: tree.ErrInvalidNodeSlug, want: "node slug must not be '.', '..', or contain '/'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewAdminNodeHandler(&fakeAdminNodeService{err: tt.err})
			request := httptest.NewRequest(http.MethodPost, "/nodes", strings.NewReader(`{"kind":"directory","name":"Notes","slug":"notes"}`))
			response := httptest.NewRecorder()

			handler.CreateNode(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusBadRequest, response.Body.String())
			}
			if !strings.Contains(response.Body.String(), `"error":"`+tt.want+`"`) {
				t.Fatalf("body = %s, want error %q", response.Body.String(), tt.want)
			}
		})
	}
}

type fakeAdminNodeService struct {
	detail tree.AdminNodeDetail
	err    error
}

func (f *fakeAdminNodeService) GetNode(context.Context, uuid.UUID) (tree.AdminNodeDetail, error) {
	return f.detail, f.err
}

func (f *fakeAdminNodeService) CreateNode(context.Context, tree.CreateNodeInput) (tree.AdminNodeDetail, error) {
	return f.detail, f.err
}

func (f *fakeAdminNodeService) UpdateNode(context.Context, uuid.UUID, tree.UpdateNodeInput) (tree.AdminNodeDetail, error) {
	return f.detail, f.err
}
