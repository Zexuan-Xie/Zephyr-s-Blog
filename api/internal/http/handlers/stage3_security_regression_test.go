package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"xlab-blog/api/internal/tree"
)

func TestStage3PublishRequiresPositiveExpectedRevisionAndValidJSON(t *testing.T) {
	fileID := uuid.New()
	handler := NewTreeLifecycleHandler(&fakeTreeLifecycleService{content: tree.FileContent{NodeID: fileID, Revision: 7, ContentFormat: tree.ContentFormatMarkdown}})
	router := chi.NewRouter()
	router.Post("/files/{file_id}/publish", handler.PublishFile)

	for _, body := range []string{"", `{`, `{"expected_revision":0}`, `{"expected_revision":-1}`} {
		request := httptest.NewRequest(http.MethodPost, "/files/"+fileID.String()+"/publish", strings.NewReader(body))
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("publish body %q status = %d, want %d; response=%s", body, response.Code, http.StatusBadRequest, response.Body.String())
		}
	}
}

func TestStage3RevisionConflictIncludesCurrentRevision(t *testing.T) {
	fileID := uuid.New()
	handler := NewTreeLifecycleHandler(&fakeTreeLifecycleService{content: tree.FileContent{NodeID: fileID, Revision: 9, ContentFormat: tree.ContentFormatMarkdown}, err: tree.ErrLostUpdate})
	router := chi.NewRouter()
	router.Put("/files/{file_id}/content", handler.UpsertFileContent)

	request := httptest.NewRequest(http.MethodPut, "/files/"+fileID.String()+"/content", strings.NewReader(`{"expected_revision":1,"content_format":"markdown","body_raw":"stale","keywords":[]}`))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusConflict, response.Body.String())
	}
	body := response.Body.String()
	if !strings.Contains(body, "revision_conflict") || !strings.Contains(body, `"current_revision":9`) {
		t.Fatalf("body = %s, want revision_conflict with current_revision 9", body)
	}
}
