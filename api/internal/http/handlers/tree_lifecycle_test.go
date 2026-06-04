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

func TestTreeLifecycleHandlerRoutes(t *testing.T) {
	fileID := uuid.New()
	service := &fakeTreeLifecycleService{
		content: tree.FileContent{NodeID: fileID, ContentFormat: tree.ContentFormatMarkdown, Status: tree.PublishStatusDraft},
	}
	handler := NewTreeLifecycleHandler(service)
	router := chi.NewRouter()
	router.Put("/files/{file_id}/content", handler.UpsertFileContent)
	router.Post("/files/{file_id}/publish", handler.PublishFile)
	router.Delete("/nodes/{node_id}", handler.DeleteNode)

	tests := []struct {
		method string
		path   string
		body   string
		status int
	}{
		{method: http.MethodPut, path: "/files/" + fileID.String() + "/content", body: `{"content_format":"markdown","body_raw":"hello","keywords":[]}`, status: http.StatusOK},
		{method: http.MethodPost, path: "/files/" + fileID.String() + "/publish", status: http.StatusOK},
		{method: http.MethodDelete, path: "/nodes/" + fileID.String(), status: http.StatusNoContent},
		{method: http.MethodPost, path: "/files/not-a-uuid/publish", status: http.StatusBadRequest},
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

func TestTreeLifecycleHandlerMapsConflict(t *testing.T) {
	fileID := uuid.New()
	handler := NewTreeLifecycleHandler(&fakeTreeLifecycleService{err: tree.ErrPublishedFileDelete})
	router := chi.NewRouter()
	router.Delete("/nodes/{node_id}", handler.DeleteNode)

	request := httptest.NewRequest(http.MethodDelete, "/nodes/"+fileID.String(), nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusConflict)
	}
}

type fakeTreeLifecycleService struct {
	content tree.FileContent
	err     error
}

func (f *fakeTreeLifecycleService) UpsertFileContent(context.Context, uuid.UUID, tree.UpsertFileContentInput) (tree.FileContent, error) {
	return f.content, f.err
}

func (f *fakeTreeLifecycleService) PublishFile(context.Context, uuid.UUID) (tree.FileContent, error) {
	return f.content, f.err
}

func (f *fakeTreeLifecycleService) UnpublishFile(context.Context, uuid.UUID) (tree.FileContent, error) {
	return f.content, f.err
}

func (f *fakeTreeLifecycleService) DeleteNode(context.Context, uuid.UUID) error {
	return f.err
}
