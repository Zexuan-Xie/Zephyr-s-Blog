package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"xlab-blog/api/internal/auth"
	"xlab-blog/api/internal/tree"
	"xlab-blog/api/internal/users"
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

func TestRouterExposesAdminNodeRoutes(t *testing.T) {
	adminUser := users.User{ID: uuid.New(), Email: "admin@example.com", Role: users.RoleAdmin}
	userRepo := &routerFakeUserRepository{user: adminUser}
	tokens := auth.NewTokenService("router-test-secret", time.Hour)
	authService := auth.NewService(userRepo, tokens)
	token, err := tokens.Issue(adminUser)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	nodeID := uuid.New()
	adminRepo := &routerFakeAdminRepository{detail: tree.AdminNodeDetail{
		Node:             tree.Node{ID: nodeID, Kind: tree.NodeKindDirectory, Name: "Notes", Slug: "notes", Path: "/notes"},
		Assets:           []tree.FileAsset{},
		RedirectsCreated: []tree.PathRedirect{},
	}}
	router := NewRouter(Dependencies{
		AuthService:  authService,
		Tokens:       tokens,
		AdminService: tree.NewAdminService(adminRepo, nil),
	})

	tests := []struct {
		method string
		path   string
		body   string
		status int
	}{
		{method: http.MethodPost, path: "/api/admin/nodes", body: `{"kind":"directory","name":"Notes","slug":"notes"}`, status: http.StatusCreated},
		{method: http.MethodGet, path: "/api/admin/nodes/" + nodeID.String(), status: http.StatusOK},
		{method: http.MethodPatch, path: "/api/admin/nodes/" + nodeID.String(), body: `{"name":"Renamed"}`, status: http.StatusOK},
	}

	for _, tt := range tests {
		request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
		request.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != tt.status {
			t.Fatalf("%s %s status = %d, want %d; body=%s", tt.method, tt.path, response.Code, tt.status, response.Body.String())
		}
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

type routerFakeAdminRepository struct {
	detail tree.AdminNodeDetail
}

func (f *routerFakeAdminRepository) GetAdminNode(context.Context, uuid.UUID) (tree.AdminNodeDetail, error) {
	return f.detail, nil
}

func (f *routerFakeAdminRepository) CreateNode(context.Context, tree.CreateNodeInput) (tree.AdminNodeDetail, error) {
	return f.detail, nil
}

func (f *routerFakeAdminRepository) UpdateNode(context.Context, uuid.UUID, tree.UpdateNodeInput) (tree.AdminNodeDetail, error) {
	return f.detail, nil
}

type routerFakeUserRepository struct {
	user users.User
}

func (f *routerFakeUserRepository) CreateReader(context.Context, string, string, *string) (users.User, error) {
	return users.User{}, errors.New("not implemented")
}

func (f *routerFakeUserRepository) FindByEmail(context.Context, string) (users.User, error) {
	return f.user, nil
}

func (f *routerFakeUserRepository) FindByID(context.Context, uuid.UUID) (users.User, error) {
	return f.user, nil
}

func (f *routerFakeUserRepository) UpsertAdmin(context.Context, string, string) (users.User, error) {
	return f.user, nil
}
