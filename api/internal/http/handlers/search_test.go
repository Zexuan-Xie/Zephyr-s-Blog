package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"xlab-blog/api/internal/search"
	"xlab-blog/api/internal/tree"
)

func TestSearchHandlerRoutesQueryAndAdminRefresh(t *testing.T) {
	fileID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	service := &fakeSearchService{fileID: fileID}
	handler := NewSearchHandler(service)

	request := httptest.NewRequest(http.MethodGet, "/api/search?q=go&limit=5&offset=1", nil)
	response := httptest.NewRecorder()
	handler.Search(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("search status = %d body=%s", response.Code, response.Body.String())
	}
	if service.query != "go" || service.options.Limit != 5 || service.options.Offset != 1 {
		t.Fatalf("query/options = %q %#v", service.query, service.options)
	}
	if !strings.Contains(response.Body.String(), `"match_sources":["text"]`) {
		t.Fatalf("search body = %s", response.Body.String())
	}

	request = httptest.NewRequest(http.MethodPost, "/api/admin/files/"+fileID.String()+"/refresh-embedding", nil)
	response = httptest.NewRecorder()
	withTreeParam(handler.RefreshEmbedding, "file_id", fileID.String()).ServeHTTP(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("refresh status = %d body=%s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"dimensions":1024`) {
		t.Fatalf("refresh body = %s", response.Body.String())
	}
}

func TestSearchHandlerRejectsInvalidQuery(t *testing.T) {
	handler := NewSearchHandler(&fakeSearchService{})
	request := httptest.NewRequest(http.MethodGet, "/api/search?q=", nil)
	response := httptest.NewRecorder()
	handler.Search(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", response.Code)
	}
}

type fakeSearchService struct {
	fileID  uuid.UUID
	query   string
	options search.Options
}

func (f *fakeSearchService) Search(_ context.Context, query string, options search.Options) (search.Response, error) {
	f.query = query
	f.options = options
	if strings.TrimSpace(query) == "" {
		return search.Response{}, search.ErrInvalidQuery
	}
	reading := 1
	return search.Response{Query: query, Items: []search.Result{{
		File: tree.FileEntry{Node: tree.Node{ID: f.fileID, Kind: tree.NodeKindFile, Path: "/go", Name: "Go"}, ContentFormat: tree.ContentFormatMarkdown, Status: tree.PublishStatusPublished, ReadingTimeMinutes: &reading},
		Path: "/go", Snippet: "go", Score: 1, MatchSources: []string{search.SourceText},
	}}}, nil
}

func (f *fakeSearchService) RefreshFileEmbedding(context.Context, uuid.UUID) (search.EmbeddingState, error) {
	return search.EmbeddingState{FileID: f.fileID, Provider: search.ProviderQwen, Model: "text-embedding-v4", Dimensions: 1024, Status: tree.EmbeddingStatusReady}, nil
}

func (f *fakeSearchService) Rebuild(context.Context) (search.RebuildState, error) {
	return search.RebuildState{Status: "accepted"}, nil
}

func withTreeParam(handler http.HandlerFunc, key string, value string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add(key, value)
		handler(w, r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx)))
	})
}
