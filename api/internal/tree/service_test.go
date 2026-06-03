package tree

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "empty defaults root", input: "", want: "/"},
		{name: "root", input: "/", want: "/"},
		{name: "collapses slashes and trims trailing", input: "//notes///go/", want: "/notes/go"},
		{name: "trims whitespace", input: " /notes/go ", want: "/notes/go"},
		{name: "requires absolute path", input: "notes/go", wantErr: true},
		{name: "rejects dot segment", input: "/notes/./go", wantErr: true},
		{name: "rejects parent segment", input: "/notes/../go", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizePath(tt.input)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidPath) {
					t.Fatalf("error = %v, want ErrInvalidPath", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizePath() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("NormalizePath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveRootReturnsDirectoryPage(t *testing.T) {
	repo := newFakeRepository()
	root := DirectoryPage{Path: "/", Entries: []any{}}
	repo.directoryPages[uuid.Nil] = root

	got, err := NewService(repo).Resolve(context.Background(), "/")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.Type != ResolveTypeDirectory || got.Directory == nil || got.Directory.Path != "/" {
		t.Fatalf("Resolve() = %#v, want root directory", got)
	}
}

func TestResolveWalksCurrentTreeBeforeRedirect(t *testing.T) {
	repo := newFakeRepository()
	dirID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	repo.nodes[parentSlugKey{slug: "notes"}] = Node{ID: dirID, Kind: NodeKindDirectory, Slug: "notes", Path: "/notes"}
	repo.directoryPages[dirID] = DirectoryPage{Node: &Node{ID: dirID, Kind: NodeKindDirectory, Path: "/notes"}, Path: "/notes", Entries: []any{}}
	repo.redirects["/notes"] = "/new-notes"

	got, err := NewService(repo).Resolve(context.Background(), "/notes")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.Type != ResolveTypeDirectory {
		t.Fatalf("Resolve() type = %q, want directory", got.Type)
	}
	if repo.redirectLookups != 0 {
		t.Fatalf("redirect lookups = %d, want 0 for current-tree match", repo.redirectLookups)
	}
}

func TestResolveDraftFileUsesRedirectFallback(t *testing.T) {
	repo := newFakeRepository()
	fileID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	repo.nodes[parentSlugKey{slug: "draft"}] = Node{ID: fileID, Kind: NodeKindFile, Slug: "draft", Path: "/draft"}
	repo.fileErr[fileID] = ErrNotFound
	repo.redirects["/draft"] = "/published"

	got, err := NewService(repo).Resolve(context.Background(), "/draft")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.Type != ResolveTypeRedirect || got.NewPath != "/published" {
		t.Fatalf("Resolve() = %#v, want redirect", got)
	}
}

func TestChildrenReturnsDirectoryPage(t *testing.T) {
	repo := newFakeRepository()
	dirID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	repo.directoryPages[dirID] = DirectoryPage{Node: &Node{ID: dirID, Kind: NodeKindDirectory, Path: "/notes"}, Path: "/notes", Entries: []any{}}

	got, err := NewService(repo).Children(context.Background(), dirID)
	if err != nil {
		t.Fatalf("Children() error = %v", err)
	}
	if got.Path != "/notes" {
		t.Fatalf("Children().Path = %q, want /notes", got.Path)
	}
}

type parentSlugKey struct {
	parent uuid.UUID
	slug   string
}

type fakeRepository struct {
	directoryPages  map[uuid.UUID]DirectoryPage
	nodes           map[parentSlugKey]Node
	files           map[uuid.UUID]FilePage
	fileErr         map[uuid.UUID]error
	redirects       map[string]string
	redirectLookups int
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		directoryPages: map[uuid.UUID]DirectoryPage{},
		nodes:          map[parentSlugKey]Node{},
		files:          map[uuid.UUID]FilePage{},
		fileErr:        map[uuid.UUID]error{},
		redirects:      map[string]string{},
	}
}

func (f *fakeRepository) DirectoryPage(_ context.Context, parentID *uuid.UUID) (DirectoryPage, error) {
	key := uuid.Nil
	if parentID != nil {
		key = *parentID
	}
	page, ok := f.directoryPages[key]
	if !ok {
		return DirectoryPage{}, ErrNotFound
	}
	return page, nil
}

func (f *fakeRepository) FilePage(_ context.Context, node Node) (FilePage, error) {
	if err := f.fileErr[node.ID]; err != nil {
		return FilePage{}, err
	}
	page, ok := f.files[node.ID]
	if !ok {
		return FilePage{}, ErrNotFound
	}
	return page, nil
}

func (f *fakeRepository) FindNodeByParentAndSlug(_ context.Context, parentID *uuid.UUID, slug string) (Node, error) {
	key := parentSlugKey{slug: slug}
	if parentID != nil {
		key.parent = *parentID
	}
	node, ok := f.nodes[key]
	if !ok {
		return Node{}, ErrNotFound
	}
	return node, nil
}

func (f *fakeRepository) RedirectPath(_ context.Context, oldPath string) (string, error) {
	f.redirectLookups++
	newPath, ok := f.redirects[oldPath]
	if !ok {
		return "", ErrNotFound
	}
	return newPath, nil
}
