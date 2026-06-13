package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"xlab-blog/api/internal/tree"
)

func TestStage2AdminNodeHandlerRoutesProtectedTreeReorderAndMove(t *testing.T) {
	handler := NewAdminNodeHandler(&fakeAdminNodeService{})
	router := chi.NewRouter()
	router.Get("/tree", handler.AdminTree)
	router.Put("/nodes/{parent_id}/children/order", handler.ReorderChildren)
	router.Post("/nodes/{node_id}/move-preview", handler.PreviewMove)
	router.Post("/nodes/{node_id}/move", handler.MoveNode)

	tests := []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodGet, path: "/tree"},
		{method: http.MethodPut, path: "/nodes/11111111-1111-1111-1111-111111111111/children/order", body: `{"child_ids":["22222222-2222-2222-2222-222222222222"],"expected_version":1}`},
		{method: http.MethodPost, path: "/nodes/11111111-1111-1111-1111-111111111111/move-preview", body: `{"new_parent_id":null,"expected_version":1}`},
		{method: http.MethodPost, path: "/nodes/11111111-1111-1111-1111-111111111111/move", body: `{"new_parent_id":null,"expected_version":1}`},
	}
	for _, tt := range tests {
		request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s %s status = %d, want 200; body=%s", tt.method, tt.path, response.Code, response.Body.String())
		}
	}
}

func TestStage2AdminNodeHandlerMapsMachineReadableDeleteReasons(t *testing.T) {
	handler := NewTreeLifecycleHandler(&fakeTreeLifecycleService{err: tree.ErrNonEmptyDirectoryDelete})
	router := chi.NewRouter()
	router.Delete("/nodes/{node_id}", handler.DeleteNode)

	request := httptest.NewRequest(http.MethodDelete, "/nodes/11111111-1111-1111-1111-111111111111", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body=%s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"reason":"non_empty_directory"`) {
		t.Fatalf("body = %s, want machine-readable non_empty_directory reason", response.Body.String())
	}
}
