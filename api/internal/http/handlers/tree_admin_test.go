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
