package httpapi

import (
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

func TestStage3Gateway1AdminRoutesExposeVersionPreviewAndAssetContracts(t *testing.T) {
	adminUser := users.User{ID: uuid.New(), Email: "admin@example.com", Role: users.RoleAdmin}
	userRepo := &routerFakeUserRepository{user: adminUser}
	tokens := auth.NewTokenService("stage3-gateway1-routes", time.Hour)
	authService := auth.NewService(userRepo, tokens)
	token, err := tokens.Issue(adminUser)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	fileID := uuid.New()
	router := NewRouter(Dependencies{
		AuthService:      authService,
		Tokens:           tokens,
		LifecycleService: tree.NewLifecycleService(newFakeLifecycleRepository()),
	})

	tests := []struct {
		method string
		path   string
		body   string
		want   int
	}{
		{method: http.MethodGet, path: "/api/admin/files/" + fileID.String() + "/content", want: http.StatusOK},
		{method: http.MethodPost, path: "/api/admin/files/" + fileID.String() + "/previous/restore", body: `{"expected_revision":1}`, want: http.StatusOK},
		{method: http.MethodGet, path: "/api/admin/files/" + fileID.String() + "/publish-summary", want: http.StatusOK},
		{method: http.MethodGet, path: "/api/admin/preview/" + fileID.String(), want: http.StatusOK},
		{method: http.MethodGet, path: "/api/admin/files/" + fileID.String() + "/assets", want: http.StatusOK},
	}
	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
		req.Header.Set("Authorization", "Bearer "+token)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != tt.want {
			t.Fatalf("%s %s status = %d, want %d; body=%s", tt.method, tt.path, res.Code, tt.want, res.Body.String())
		}
	}
}

func TestStage3Gateway1DraftPreviewDeniedToReader(t *testing.T) {
	reader := users.User{ID: uuid.New(), Email: "reader@example.com", Role: users.RoleReader}
	userRepo := &routerFakeUserRepository{user: reader}
	tokens := auth.NewTokenService("stage3-gateway1-preview-denial", time.Hour)
	authService := auth.NewService(userRepo, tokens)
	token, err := tokens.Issue(reader)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	router := NewRouter(Dependencies{AuthService: authService, Tokens: tokens})

	req := httptest.NewRequest(http.MethodGet, "/api/admin/preview/"+uuid.NewString(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusForbidden {
		t.Fatalf("reader draft preview status = %d, want %d; body=%s", res.Code, http.StatusForbidden, res.Body.String())
	}
}
