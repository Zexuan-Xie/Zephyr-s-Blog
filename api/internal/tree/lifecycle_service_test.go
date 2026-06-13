package tree

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestLifecycleUpsertRejectsPublishedFormatChange(t *testing.T) {
	fileID := uuid.New()
	repo := newFakeLifecycleRepository()
	repo.nodes[fileID] = Node{ID: fileID, Kind: NodeKindFile, Name: "Post", Path: "/post"}
	repo.contents[fileID] = FileContent{NodeID: fileID, ContentFormat: ContentFormatMarkdown, Status: PublishStatusPublished}

	_, err := NewLifecycleService(repo).UpsertFileContent(context.Background(), fileID, UpsertFileContentInput{
		ContentFormat: ContentFormatHTMLDocument,
		BodyRaw:       "<html></html>",
	})
	if !errors.Is(err, ErrPublishedContentFormatChange) {
		t.Fatalf("error = %v, want ErrPublishedContentFormatChange", err)
	}
}

func TestLifecycleUpsertNormalizesKeywordsAndBuildsSearchText(t *testing.T) {
	fileID := uuid.New()
	repo := newFakeLifecycleRepository()
	repo.nodes[fileID] = Node{ID: fileID, Kind: NodeKindFile, Name: "Go Notes", Path: "/notes/go"}

	content, err := NewLifecycleService(repo).UpsertFileContent(context.Background(), fileID, UpsertFileContentInput{
		ContentFormat: ContentFormatMarkdown,
		BodyRaw:       "Concurrency",
		Keywords:      []string{" Go ", "go", "", "Systems"},
	})
	if err != nil {
		t.Fatalf("UpsertFileContent() error = %v", err)
	}
	if len(content.Keywords) != 2 || content.Keywords[0] != "Go" || content.Keywords[1] != "Systems" {
		t.Fatalf("keywords = %#v, want normalized unique values", content.Keywords)
	}
	if content.SearchText == "" {
		t.Fatal("search text was not generated")
	}
}

func TestLifecycleDeleteProtectsPublishedContent(t *testing.T) {
	fileID := uuid.New()
	directoryID := uuid.New()
	repo := newFakeLifecycleRepository()
	repo.nodes[fileID] = Node{ID: fileID, Kind: NodeKindFile}
	repo.nodes[directoryID] = Node{ID: directoryID, Kind: NodeKindDirectory}
	repo.contents[fileID] = FileContent{NodeID: fileID, Status: PublishStatusPublished}
	repo.hasPublished[directoryID] = true
	service := NewLifecycleService(repo)

	if err := service.DeleteNode(context.Background(), fileID); !errors.Is(err, ErrPublishedFileDelete) {
		t.Fatalf("file delete error = %v, want ErrPublishedFileDelete", err)
	}
	if err := service.DeleteNode(context.Background(), directoryID); !errors.Is(err, ErrDirectoryHasPublishedDescendants) {
		t.Fatalf("directory delete error = %v, want ErrDirectoryHasPublishedDescendants", err)
	}
}

func TestLifecycleRecordDirectoryPathChangeCreatesRedirects(t *testing.T) {
	directoryID := uuid.New()
	fileID := uuid.New()
	repo := newFakeLifecycleRepository()
	repo.nodes[directoryID] = Node{ID: directoryID, Kind: NodeKindDirectory}
	repo.descendantPaths[directoryID] = []PublishedFilePath{{NodeID: fileID, Path: "/new/post"}}

	err := NewLifecycleService(repo).RecordPathChange(context.Background(), directoryID, "/old", "/new")
	if err != nil {
		t.Fatalf("RecordPathChange() error = %v", err)
	}
	if got := repo.redirects["/old/post"]; got != "/new/post" {
		t.Fatalf("redirect target = %q, want /new/post", got)
	}
}

func TestLifecycleRecordPublishedFilePathChangeCreatesRedirect(t *testing.T) {
	fileID := uuid.New()
	repo := newFakeLifecycleRepository()
	repo.nodes[fileID] = Node{ID: fileID, Kind: NodeKindFile}
	repo.contents[fileID] = FileContent{NodeID: fileID, Status: PublishStatusPublished}
	repo.redirects["/first/post"] = "/old/post"

	err := NewLifecycleService(repo).RecordPathChange(context.Background(), fileID, "/old/post", "/new/post")
	if err != nil {
		t.Fatalf("RecordPathChange() error = %v", err)
	}
	if got := repo.redirects["/old/post"]; got != "/new/post" {
		t.Fatalf("new redirect target = %q, want /new/post", got)
	}
	if got := repo.redirects["/first/post"]; got != "/new/post" {
		t.Fatalf("existing redirect target = %q, want final path /new/post", got)
	}
}

func TestLifecycleRecordDraftFilePathChangeDoesNotCreateRedirect(t *testing.T) {
	fileID := uuid.New()
	repo := newFakeLifecycleRepository()
	repo.nodes[fileID] = Node{ID: fileID, Kind: NodeKindFile}
	repo.contents[fileID] = FileContent{NodeID: fileID, Status: PublishStatusDraft}

	err := NewLifecycleService(repo).RecordPathChange(context.Background(), fileID, "/old/post", "/new/post")
	if err != nil {
		t.Fatalf("RecordPathChange() error = %v", err)
	}
	if len(repo.redirects) != 0 {
		t.Fatalf("redirects = %#v, want none", repo.redirects)
	}
}

type fakeLifecycleRepository struct {
	nodes           map[uuid.UUID]Node
	contents        map[uuid.UUID]FileContent
	hasPublished    map[uuid.UUID]bool
	descendantPaths map[uuid.UUID][]PublishedFilePath
	redirects       map[string]string
	deleted         map[uuid.UUID]bool
}

func newFakeLifecycleRepository() *fakeLifecycleRepository {
	return &fakeLifecycleRepository{
		nodes:           map[uuid.UUID]Node{},
		contents:        map[uuid.UUID]FileContent{},
		hasPublished:    map[uuid.UUID]bool{},
		descendantPaths: map[uuid.UUID][]PublishedFilePath{},
		redirects:       map[string]string{},
		deleted:         map[uuid.UUID]bool{},
	}
}

func (f *fakeLifecycleRepository) GetNode(_ context.Context, nodeID uuid.UUID) (Node, error) {
	node, ok := f.nodes[nodeID]
	if !ok {
		return Node{}, ErrNodeNotFound
	}
	return node, nil
}

func (f *fakeLifecycleRepository) GetFileContent(_ context.Context, nodeID uuid.UUID) (FileContent, error) {
	content, ok := f.contents[nodeID]
	if !ok {
		return FileContent{}, ErrFileContentNotFound
	}
	return content, nil
}

func (f *fakeLifecycleRepository) UpsertFileContent(_ context.Context, nodeID uuid.UUID, input UpsertFileContentInput) (FileContent, error) {
	content := FileContent{
		NodeID:        nodeID,
		ContentFormat: input.ContentFormat,
		Keywords:      input.Keywords,
		BodyRaw:       input.BodyRaw,
		BodyHTML:      input.BodyHTML,
		SearchText:    input.SearchText,
		Status:        PublishStatusDraft,
	}
	f.contents[nodeID] = content
	return content, nil
}

func (f *fakeLifecycleRepository) PublishFile(_ context.Context, nodeID uuid.UUID) (FileContent, error) {
	content, err := f.GetFileContent(context.Background(), nodeID)
	content.Status = PublishStatusPublished
	f.contents[nodeID] = content
	return content, err
}

func (f *fakeLifecycleRepository) UnpublishFile(_ context.Context, nodeID uuid.UUID) (FileContent, error) {
	content, err := f.GetFileContent(context.Background(), nodeID)
	content.Status = PublishStatusDraft
	f.contents[nodeID] = content
	return content, err
}

func (f *fakeLifecycleRepository) DeleteNode(_ context.Context, nodeID uuid.UUID) error {
	f.deleted[nodeID] = true
	return nil
}

func (f *fakeLifecycleRepository) HasChildNodes(_ context.Context, directoryID uuid.UUID) (bool, error) {
	for _, node := range f.nodes {
		if node.ParentID != nil && *node.ParentID == directoryID {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeLifecycleRepository) HasPublishedDescendantFiles(_ context.Context, directoryID uuid.UUID) (bool, error) {
	return f.hasPublished[directoryID], nil
}

func (f *fakeLifecycleRepository) PublishedDescendantFilePaths(_ context.Context, directoryID uuid.UUID) ([]PublishedFilePath, error) {
	return f.descendantPaths[directoryID], nil
}

func (f *fakeLifecycleRepository) UpdateRedirectTargets(_ context.Context, nodeID uuid.UUID, finalPath string) error {
	for oldPath := range f.redirects {
		f.redirects[oldPath] = finalPath
	}
	return nil
}

func (f *fakeLifecycleRepository) UpsertPathRedirect(_ context.Context, oldPath, newPath string, _ uuid.UUID) error {
	f.redirects[oldPath] = newPath
	return nil
}
