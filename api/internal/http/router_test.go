package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"xlab-blog/api/internal/assets"
	"xlab-blog/api/internal/auth"
	"xlab-blog/api/internal/comments"
	"xlab-blog/api/internal/likes"
	"xlab-blog/api/internal/search"
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

func TestRouterExposesCommentAndLikeRoutes(t *testing.T) {
	reader := users.User{ID: uuid.New(), Email: "reader@example.com", Role: users.RoleReader}
	userRepo := &routerFakeUserRepository{user: reader}
	tokens := auth.NewTokenService("router-comment-like-secret", time.Hour)
	authService := auth.NewService(userRepo, tokens)
	token, err := tokens.Issue(reader)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	fileID := uuid.New()
	commentID := uuid.New()
	router := NewRouter(Dependencies{
		AuthService:    authService,
		Tokens:         tokens,
		CommentService: comments.NewService(&routerFakeCommentRepository{fileID: fileID, commentID: commentID, user: reader}),
		LikeService:    likes.NewService(&routerFakeLikeRepository{fileID: fileID, comments: map[uuid.UUID]bool{commentID: false}}),
	})

	tests := []struct {
		method string
		path   string
		body   string
		token  string
		status int
		want   string
	}{
		{method: http.MethodGet, path: "/api/files/" + fileID.String() + "/comments", status: http.StatusOK, want: `"comments"`},
		{method: http.MethodPost, path: "/api/files/" + fileID.String() + "/comments", body: `{"body":"hello"}`, token: token, status: http.StatusCreated, want: `"body":"hello"`},
		{method: http.MethodPost, path: "/api/files/" + fileID.String() + "/comments", body: `{"body":"hello"}`, status: http.StatusUnauthorized, want: `"error":"authentication required"`},
		{method: http.MethodPut, path: "/api/files/" + fileID.String() + "/like", token: token, status: http.StatusOK, want: `"liked":true`},
		{method: http.MethodDelete, path: "/api/files/" + fileID.String() + "/like", token: token, status: http.StatusOK, want: `"liked":false`},
		{method: http.MethodPut, path: "/api/comments/" + commentID.String() + "/like", token: token, status: http.StatusOK, want: `"like_count"`},
		{method: http.MethodPut, path: "/api/comments/" + commentID.String() + "/like", status: http.StatusUnauthorized, want: `"error":"authentication required"`},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.token != "" {
				request.Header.Set("Authorization", "Bearer "+tt.token)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d; body=%s", response.Code, tt.status, response.Body.String())
			}
			if tt.want != "" && !strings.Contains(response.Body.String(), tt.want) {
				t.Fatalf("body = %s, want substring %s", response.Body.String(), tt.want)
			}
		})
	}
}

type routerFakeCommentRepository struct {
	fileID    uuid.UUID
	commentID uuid.UUID
	user      users.User
}

func (r *routerFakeCommentRepository) PublishedFileExists(_ context.Context, fileID uuid.UUID) (bool, error) {
	return fileID == r.fileID, nil
}

func (r *routerFakeCommentRepository) ListThread(_ context.Context, fileID uuid.UUID, _ *uuid.UUID) (comments.Thread, error) {
	return comments.Thread{FileID: fileID, Comments: []comments.Comment{}}, nil
}

func (r *routerFakeCommentRepository) FindComment(_ context.Context, commentID uuid.UUID) (comments.Comment, error) {
	if commentID != r.commentID {
		return comments.Comment{}, comments.ErrCommentNotFound
	}
	return comments.Comment{ID: commentID, FileNodeID: r.fileID, User: comments.PublicUser{ID: r.user.ID, DisplayName: "Reader"}, Replies: []comments.Comment{}}, nil
}

func (r *routerFakeCommentRepository) InsertComment(_ context.Context, fileID uuid.UUID, userID uuid.UUID, input comments.CreateInput) (comments.Comment, error) {
	return comments.Comment{ID: r.commentID, FileNodeID: fileID, User: comments.PublicUser{ID: userID, DisplayName: "Reader"}, Body: input.Body, Replies: []comments.Comment{}}, nil
}

func (r *routerFakeCommentRepository) SoftDeleteComment(_ context.Context, commentID uuid.UUID, deletedBy uuid.UUID) (comments.Comment, error) {
	return comments.Comment{ID: commentID, FileNodeID: r.fileID, User: comments.PublicUser{ID: deletedBy, DisplayName: "Reader"}, Deleted: true, Replies: []comments.Comment{}}, nil
}

type routerFakeLikeRepository struct {
	fileID   uuid.UUID
	comments map[uuid.UUID]bool
	liked    bool
}

func (r *routerFakeLikeRepository) FileTargetExists(_ context.Context, fileID uuid.UUID) (bool, error) {
	return fileID == r.fileID, nil
}

func (r *routerFakeLikeRepository) CommentTargetState(_ context.Context, commentID uuid.UUID) (bool, bool, error) {
	deleted, ok := r.comments[commentID]
	return ok, deleted, nil
}

func (r *routerFakeLikeRepository) UpsertLike(context.Context, uuid.UUID, likes.Target) error {
	r.liked = true
	return nil
}

func (r *routerFakeLikeRepository) DeleteLike(context.Context, uuid.UUID, likes.Target) error {
	r.liked = false
	return nil
}

func (r *routerFakeLikeRepository) LikeState(context.Context, uuid.UUID, likes.Target) (likes.State, error) {
	count := 0
	if r.liked {
		count = 1
	}
	return likes.State{Liked: r.liked, LikeCount: count}, nil
}

func TestRouterExposesAssetRoutes(t *testing.T) {
	adminUser := users.User{ID: uuid.New(), Email: "admin@example.com", Role: users.RoleAdmin}
	userRepo := &routerFakeUserRepository{user: adminUser}
	tokens := auth.NewTokenService("router-asset-secret", time.Hour)
	authService := auth.NewService(userRepo, tokens)
	token, err := tokens.Issue(adminUser)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	fileID := uuid.New()
	assetService := assets.NewService(&routerFakeAssetRepository{files: map[uuid.UUID]bool{fileID: true}}, &routerFakeAssetStorage{objects: map[string][]byte{}})
	router := NewRouter(Dependencies{AuthService: authService, Tokens: tokens, AssetService: assetService})

	body, contentType := routerMultipartBody(t, "file", "demo.txt", "text/plain", "demo")
	request := httptest.NewRequest(http.MethodPost, "/api/admin/files/"+fileID.String()+"/assets", body)
	request.Header.Set("Content-Type", contentType)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("upload status = %d, want %d; body=%s", response.Code, http.StatusCreated, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"storage_key"`) || strings.Contains(response.Body.String(), "/tmp/") {
		t.Fatalf("upload body leaked or missed storage key: %s", response.Body.String())
	}

	assetID := assetServiceTestLastID(response.Body.String())
	request = httptest.NewRequest(http.MethodGet, "/api/assets/"+assetID+"/demo.txt", nil)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("serve status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	if got := response.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("Cache-Control = %q", got)
	}

	request = httptest.NewRequest(http.MethodDelete, "/api/admin/assets/"+assetID, nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d; body=%s", response.Code, http.StatusNoContent, response.Body.String())
	}
}

func routerMultipartBody(t *testing.T, field, filename, contentType, value string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="`+field+`"; filename="`+filename+`"`)
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte(value)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return body, writer.FormDataContentType()
}

func assetServiceTestLastID(body string) string {
	var payload struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal([]byte(body), &payload)
	return payload.ID
}

type routerFakeAssetRepository struct {
	files  map[uuid.UUID]bool
	assets map[uuid.UUID]assets.FileAsset
}

func (r *routerFakeAssetRepository) FileTargetExists(_ context.Context, fileID uuid.UUID) (bool, error) {
	return r.files[fileID], nil
}

func (r *routerFakeAssetRepository) FileAssetTotalBytes(context.Context, uuid.UUID) (int64, error) {
	return 0, nil
}

func (r *routerFakeAssetRepository) CreateAsset(_ context.Context, asset assets.FileAsset) (assets.FileAsset, error) {
	asset.PublicURL = "/api/assets/" + asset.ID.String() + "/" + asset.Filename
	if r.assets == nil {
		r.assets = map[uuid.UUID]assets.FileAsset{}
	}
	r.assets[asset.ID] = asset
	return asset, nil
}

func (r *routerFakeAssetRepository) FindPublishedAsset(_ context.Context, assetID uuid.UUID, filename string) (assets.FileAsset, error) {
	asset, ok := r.assets[assetID]
	if !ok || asset.Filename != filename {
		return assets.FileAsset{}, assets.ErrAssetNotFound
	}
	return asset, nil
}

func (r *routerFakeAssetRepository) DeleteAsset(_ context.Context, assetID uuid.UUID) (assets.FileAsset, error) {
	asset, ok := r.assets[assetID]
	if !ok {
		return assets.FileAsset{}, assets.ErrAssetNotFound
	}
	delete(r.assets, assetID)
	return asset, nil
}

type routerFakeAssetStorage struct {
	objects map[string][]byte
}

func (s *routerFakeAssetStorage) Put(key string, reader io.Reader) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	s.objects[key] = data
	return nil
}

func (s *routerFakeAssetStorage) Open(key string) (assets.StoredObject, error) {
	data, ok := s.objects[key]
	if !ok {
		return assets.StoredObject{}, assets.ErrAssetNotFound
	}
	return assets.StoredObject{Reader: io.NopCloser(bytes.NewReader(data)), Size: int64(len(data)), ContentType: "text/plain"}, nil
}

func (s *routerFakeAssetStorage) Delete(key string) error {
	delete(s.objects, key)
	return nil
}

func TestRouterExposesSearchRoutes(t *testing.T) {
	adminUser := users.User{ID: uuid.New(), Email: "admin@example.com", Role: users.RoleAdmin}
	userRepo := &routerFakeUserRepository{user: adminUser}
	tokens := auth.NewTokenService("router-search-secret", time.Hour)
	authService := auth.NewService(userRepo, tokens)
	token, err := tokens.Issue(adminUser)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	fileID := uuid.New()
	router := NewRouter(Dependencies{AuthService: authService, Tokens: tokens, SearchService: &routerFakeSearchService{fileID: fileID}})

	tests := []struct {
		method string
		path   string
		token  string
		status int
		want   string
	}{
		{method: http.MethodGet, path: "/api/search?q=go", status: http.StatusOK, want: `"query":"go"`},
		{method: http.MethodPost, path: "/api/admin/files/" + fileID.String() + "/refresh-embedding", token: token, status: http.StatusAccepted, want: `"dimensions":1024`},
		{method: http.MethodPost, path: "/api/admin/search-index/rebuild", token: token, status: http.StatusAccepted, want: `"status":"accepted"`},
		{method: http.MethodPost, path: "/api/admin/search-index/rebuild", status: http.StatusUnauthorized, want: `"error":"authentication required"`},
	}
	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		if tt.token != "" {
			req.Header.Set("Authorization", "Bearer "+tt.token)
		}
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != tt.status {
			t.Fatalf("%s %s status = %d, want %d; body=%s", tt.method, tt.path, res.Code, tt.status, res.Body.String())
		}
		if !strings.Contains(res.Body.String(), tt.want) {
			t.Fatalf("body = %s, want %s", res.Body.String(), tt.want)
		}
	}
}

type routerFakeSearchService struct {
	fileID uuid.UUID
}

func (s *routerFakeSearchService) Search(_ context.Context, query string, _ search.Options) (search.Response, error) {
	reading := 1
	return search.Response{Query: strings.TrimSpace(query), Items: []search.Result{{
		File: tree.FileEntry{Node: tree.Node{ID: s.fileID, Kind: tree.NodeKindFile, Name: "Search", Path: "/search-file"}, ContentFormat: tree.ContentFormatMarkdown, Status: tree.PublishStatusPublished, ReadingTimeMinutes: &reading},
		Path: "/search-file", Snippet: "Search", Score: 1, MatchSources: []string{search.SourceText},
	}}}, nil
}

func (s *routerFakeSearchService) RefreshFileEmbedding(context.Context, uuid.UUID) (search.EmbeddingState, error) {
	return search.EmbeddingState{FileID: s.fileID, Provider: search.ProviderQwen, Model: "text-embedding-v4", Dimensions: 1024, Status: tree.EmbeddingStatusReady}, nil
}

func (s *routerFakeSearchService) Rebuild(context.Context) (search.RebuildState, error) {
	return search.RebuildState{Status: "accepted"}, nil
}
