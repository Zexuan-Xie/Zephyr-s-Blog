package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"xlab-blog/api/internal/tree"
)

func TestRouterExposesPublicTreeRoutes(t *testing.T) {
	repo := newRouterFakeTreeRepository()
	dirID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	repo.directoryPages[uuid.Nil] = tree.DirectoryPage{Path: "/", Entries: []any{}}
	repo.directoryPages[dirID] = tree.DirectoryPage{Node: &tree.Node{ID: dirID, Kind: tree.NodeKindDirectory, Path: "/notes"}, Path: "/notes", Entries: []any{}}
	repo.nodes[routerParentSlugKey{slug: "notes"}] = tree.Node{ID: dirID, Kind: tree.NodeKindDirectory, Slug: "notes", Path: "/notes"}

	router := NewRouter(Dependencies{TreeService: tree.NewService(repo)})

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{name: "root", path: "/api/tree", wantStatus: http.StatusOK, wantBody: `"path":"/"`},
		{name: "resolve directory", path: "/api/tree/resolve?path=/notes", wantStatus: http.StatusOK, wantBody: `"type":"directory"`},
		{name: "children", path: "/api/tree/" + dirID.String() + "/children", wantStatus: http.StatusOK, wantBody: `"path":"/notes"`},
		{name: "invalid child id", path: "/api/tree/not-a-uuid/children", wantStatus: http.StatusBadRequest, wantBody: `"error":"invalid node_id"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", response.Code, tt.wantStatus, response.Body.String())
			}
			if !strings.Contains(response.Body.String(), tt.wantBody) {
				t.Fatalf("body = %s, want substring %s", response.Body.String(), tt.wantBody)
			}
		})
	}
}

type routerParentSlugKey struct {
	parent uuid.UUID
	slug   string
}

type routerFakeTreeRepository struct {
	directoryPages map[uuid.UUID]tree.DirectoryPage
	nodes          map[routerParentSlugKey]tree.Node
}

func newRouterFakeTreeRepository() *routerFakeTreeRepository {
	return &routerFakeTreeRepository{
		directoryPages: map[uuid.UUID]tree.DirectoryPage{},
		nodes:          map[routerParentSlugKey]tree.Node{},
	}
}

func (f *routerFakeTreeRepository) DirectoryPage(_ context.Context, parentID *uuid.UUID) (tree.DirectoryPage, error) {
	key := uuid.Nil
	if parentID != nil {
		key = *parentID
	}
	page, ok := f.directoryPages[key]
	if !ok {
		return tree.DirectoryPage{}, tree.ErrNotFound
	}
	return page, nil
}

func (f *routerFakeTreeRepository) FilePage(_ context.Context, node tree.Node) (tree.FilePage, error) {
	return tree.FilePage{}, tree.ErrNotFound
}

func (f *routerFakeTreeRepository) FindNodeByParentAndSlug(_ context.Context, parentID *uuid.UUID, slug string) (tree.Node, error) {
	key := routerParentSlugKey{slug: slug}
	if parentID != nil {
		key.parent = *parentID
	}
	node, ok := f.nodes[key]
	if !ok {
		return tree.Node{}, tree.ErrNotFound
	}
	return node, nil
}

func (f *routerFakeTreeRepository) RedirectPath(_ context.Context, oldPath string) (string, error) {
	return "", tree.ErrNotFound
}
