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

func TestStage3Gateway1LifecycleMapsRevisionConflict(t *testing.T) {
	fileID := uuid.New()
	handler := NewTreeLifecycleHandler(&fakeTreeLifecycleService{err: tree.ErrLostUpdate})
	router := chi.NewRouter()
	router.Put("/files/{file_id}/content", handler.UpsertFileContent)

	request := httptest.NewRequest(http.MethodPut, "/files/"+fileID.String()+"/content", strings.NewReader(`{"expected_revision":1,"content_format":"markdown","body_raw":"stale","keywords":[]}`))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusConflict, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "revision_conflict") {
		t.Fatalf("body = %s, want machine-readable revision_conflict", response.Body.String())
	}
}
