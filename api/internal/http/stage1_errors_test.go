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

func TestStage1CreateAndAuthErrors(t *testing.T) {
	admin := users.User{ID: uuid.New(), Email: "admin@example.com", Role: users.RoleAdmin}
	reader := users.User{ID: uuid.New(), Email: "reader@example.com", Role: users.RoleReader}
	userRepo := &stage1UserRepository{users: map[uuid.UUID]users.User{
		admin.ID:  admin,
		reader.ID: reader,
	}}
	tokens := auth.NewTokenService("stage-1-errors-secret", time.Hour)
	adminToken, err := tokens.Issue(admin)
	if err != nil {
		t.Fatalf("Issue(admin) error = %v", err)
	}
	readerToken, err := tokens.Issue(reader)
	if err != nil {
		t.Fatalf("Issue(reader) error = %v", err)
	}

	tests := []struct {
		name       string
		body       string
		token      string
		repoErr    error
		wantStatus int
		wantError  string
	}{
		{
			name:       "missing authentication",
			body:       `{"kind":"directory","name":"Notes","slug":"notes"}`,
			wantStatus: http.StatusUnauthorized,
			wantError:  "authentication required",
		},
		{
			name:       "invalid token",
			body:       `{"kind":"directory","name":"Notes","slug":"notes"}`,
			token:      "not-a-token",
			wantStatus: http.StatusUnauthorized,
			wantError:  "invalid token",
		},
		{
			name:       "reader lacks admin role",
			body:       `{"kind":"directory","name":"Notes","slug":"notes"}`,
			token:      readerToken,
			wantStatus: http.StatusForbidden,
			wantError:  "admin role required",
		},
		{
			name:       "invalid JSON",
			body:       `{"kind":`,
			token:      adminToken,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid JSON body",
		},
		{
			name:       "missing name",
			body:       `{"kind":"directory","name":" ","slug":"notes"}`,
			token:      adminToken,
			wantStatus: http.StatusBadRequest,
			wantError:  "node name is required",
		},
		{
			name:       "missing slug",
			body:       `{"kind":"directory","name":"Notes","slug":" "}`,
			token:      adminToken,
			wantStatus: http.StatusBadRequest,
			wantError:  "node slug is required",
		},
		{
			name:       "invalid kind",
			body:       `{"kind":"folder","name":"Notes","slug":"notes"}`,
			token:      adminToken,
			wantStatus: http.StatusBadRequest,
			wantError:  "node kind must be directory or file",
		},
		{
			name:       "duplicate sibling slug",
			body:       `{"kind":"directory","name":"Notes","slug":"notes"}`,
			token:      adminToken,
			repoErr:    tree.ErrDuplicateSlug,
			wantStatus: http.StatusConflict,
			wantError:  "a node with this slug already exists under the selected parent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adminRepo := &stage1AdminRepository{
				detail: tree.AdminNodeDetail{
					Node:             tree.Node{ID: uuid.New(), Kind: tree.NodeKindDirectory, Name: "Notes", Slug: "notes", Path: "/notes"},
					Assets:           []tree.FileAsset{},
					RedirectsCreated: []tree.PathRedirect{},
				},
				err: tt.repoErr,
			}
			router := NewRouter(Dependencies{
				AuthService:  auth.NewService(userRepo, tokens),
				Tokens:       tokens,
				AdminService: tree.NewAdminService(adminRepo, nil),
			})
			request := httptest.NewRequest(http.MethodPost, "/api/admin/nodes", strings.NewReader(tt.body))
			if tt.token != "" {
				request.Header.Set("Authorization", "Bearer "+tt.token)
			}
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", response.Code, tt.wantStatus, response.Body.String())
			}
			wantBody := `"error":"` + tt.wantError + `"`
			if !strings.Contains(response.Body.String(), wantBody) {
				t.Fatalf("body = %s, want substring %s", response.Body.String(), wantBody)
			}
		})
	}
}

type stage1AdminRepository struct {
	detail tree.AdminNodeDetail
	err    error
}

func (r *stage1AdminRepository) GetAdminNode(context.Context, uuid.UUID) (tree.AdminNodeDetail, error) {
	return r.detail, r.err
}

func (r *stage1AdminRepository) CreateNode(context.Context, tree.CreateNodeInput) (tree.AdminNodeDetail, error) {
	return r.detail, r.err
}

func (r *stage1AdminRepository) UpdateNode(context.Context, uuid.UUID, tree.UpdateNodeInput) (tree.AdminNodeDetail, error) {
	return r.detail, r.err
}

type stage1UserRepository struct {
	users map[uuid.UUID]users.User
}

func (*stage1UserRepository) CreateReader(context.Context, string, string, *string) (users.User, error) {
	return users.User{}, errors.New("not implemented")
}

func (*stage1UserRepository) FindByEmail(context.Context, string) (users.User, error) {
	return users.User{}, errors.New("not implemented")
}

func (r *stage1UserRepository) FindByID(_ context.Context, id uuid.UUID) (users.User, error) {
	user, ok := r.users[id]
	if !ok {
		return users.User{}, users.ErrUserNotFound
	}
	return user, nil
}

func (*stage1UserRepository) UpsertAdmin(context.Context, string, string) (users.User, error) {
	return users.User{}, errors.New("not implemented")
}
