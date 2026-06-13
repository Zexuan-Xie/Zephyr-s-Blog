package tree

import (
	"context"
	"encoding/json"
	"errors"
	"path"
	"testing"

	"github.com/google/uuid"
)

func TestAdminCreateNodeRejectsReservedRootSlug(t *testing.T) {
	repo := newFakeAdminRepository()
	service := NewAdminService(repo, nil)

	_, err := service.CreateNode(context.Background(), CreateNodeInput{
		Kind: NodeKindDirectory,
		Name: "Admin",
		Slug: " ADMIN ",
	})
	if !errors.Is(err, ErrReservedRootSlug) {
		t.Fatalf("CreateNode() error = %v, want ErrReservedRootSlug", err)
	}
	if len(repo.nodes) != 0 {
		t.Fatalf("created nodes = %d, want 0", len(repo.nodes))
	}
}

func TestAdminCreateFileRequiresContentFormat(t *testing.T) {
	repo := newFakeAdminRepository()
	service := NewAdminService(repo, nil)

	_, err := service.CreateNode(context.Background(), CreateNodeInput{
		Kind: NodeKindFile,
		Name: "Post",
		Slug: "post",
	})
	if !errors.Is(err, ErrInvalidContentFormat) {
		t.Fatalf("CreateNode() error = %v, want ErrInvalidContentFormat", err)
	}
}

func TestAdminUpdateNodeRecordsPersistedPathChange(t *testing.T) {
	nodeID := uuid.New()
	repo := newFakeAdminRepository()
	repo.nodes[nodeID] = AdminNodeDetail{Node: Node{ID: nodeID, Kind: NodeKindFile, Name: "Post", Slug: "post", Path: "/old/post"}}
	redirects := &fakePathChangeRecorder{}
	service := NewAdminService(repo, redirects)
	newSlug := "renamed"

	detail, err := service.UpdateNode(context.Background(), nodeID, UpdateNodeInput{Slug: &newSlug})
	if err != nil {
		t.Fatalf("UpdateNode() error = %v", err)
	}
	if detail.Node.Path != "/old/renamed" {
		t.Fatalf("updated path = %q, want /old/renamed", detail.Node.Path)
	}
	if redirects.oldPath != "/old/post" || redirects.newPath != "/old/renamed" {
		t.Fatalf("recorded redirect = %q -> %q, want /old/post -> /old/renamed", redirects.oldPath, redirects.newPath)
	}
}

func TestAdminUpdateNodeDoesNotRecordPathChangeTwiceForAtomicRepository(t *testing.T) {
	nodeID := uuid.New()
	repo := &atomicFakeAdminRepository{fakeAdminRepository: newFakeAdminRepository()}
	repo.nodes[nodeID] = AdminNodeDetail{Node: Node{ID: nodeID, Kind: NodeKindFile, Name: "Post", Slug: "post", Path: "/old/post"}}
	redirects := &fakePathChangeRecorder{err: errors.New("must not be called")}
	service := NewAdminService(repo, redirects)
	newSlug := "renamed"

	detail, err := service.UpdateNode(context.Background(), nodeID, UpdateNodeInput{Slug: &newSlug})
	if err != nil {
		t.Fatalf("UpdateNode() error = %v", err)
	}
	if detail.Node.Path != "/old/renamed" {
		t.Fatalf("updated path = %q, want /old/renamed", detail.Node.Path)
	}
	if redirects.oldPath != "" || redirects.newPath != "" {
		t.Fatalf("separate redirect recorder called with %q -> %q", redirects.oldPath, redirects.newPath)
	}
}

func TestAdminUpdateNodeRejectsReservedSlugWhenMovedToRoot(t *testing.T) {
	parentID := uuid.New()
	nodeID := uuid.New()
	repo := newFakeAdminRepository()
	repo.nodes[nodeID] = AdminNodeDetail{Node: Node{
		ID:       nodeID,
		ParentID: &parentID,
		Kind:     NodeKindDirectory,
		Name:     "Notes",
		Slug:     "notes",
		Path:     "/parent/notes",
	}}
	var input UpdateNodeInput
	if err := json.Unmarshal([]byte(`{"parent_id":null,"slug":"search"}`), &input); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	_, err := NewAdminService(repo, nil).UpdateNode(context.Background(), nodeID, input)
	if !errors.Is(err, ErrReservedRootSlug) {
		t.Fatalf("UpdateNode() error = %v, want ErrReservedRootSlug", err)
	}
	if repo.updateCalls != 0 {
		t.Fatalf("UpdateNode repository calls = %d, want 0", repo.updateCalls)
	}
}

func TestUpdateNodeInputTracksExplicitNullParent(t *testing.T) {
	var absent UpdateNodeInput
	if err := json.Unmarshal([]byte(`{"name":"Renamed"}`), &absent); err != nil {
		t.Fatalf("Unmarshal(absent) error = %v", err)
	}
	if absent.ParentIDSet {
		t.Fatal("absent parent_id marked as set")
	}

	var explicitNull UpdateNodeInput
	if err := json.Unmarshal([]byte(`{"parent_id":null}`), &explicitNull); err != nil {
		t.Fatalf("Unmarshal(null) error = %v", err)
	}
	if !explicitNull.ParentIDSet || explicitNull.ParentID != nil {
		t.Fatalf("explicit null parent = (%v, %v), want (true, nil)", explicitNull.ParentIDSet, explicitNull.ParentID)
	}
}

type fakeAdminRepository struct {
	nodes           map[uuid.UUID]AdminNodeDetail
	adminTree       []AdminTreeNode
	updateCalls     int
	reorderParent   uuid.UUID
	reorderChildren []uuid.UUID
	reorderVersion  int
	deleteErr       error
}

type atomicFakeAdminRepository struct {
	*fakeAdminRepository
}

func (*atomicFakeAdminRepository) recordsPathChangesAtomically() {}

func newFakeAdminRepository() *fakeAdminRepository {
	return &fakeAdminRepository{nodes: map[uuid.UUID]AdminNodeDetail{}}
}

func (f *fakeAdminRepository) GetAdminNode(_ context.Context, nodeID uuid.UUID) (AdminNodeDetail, error) {
	detail, ok := f.nodes[nodeID]
	if !ok {
		return AdminNodeDetail{}, ErrNodeNotFound
	}
	return detail, nil
}

func (f *fakeAdminRepository) CreateNode(_ context.Context, input CreateNodeInput) (AdminNodeDetail, error) {
	slug := input.Slug
	if slug == "" {
		slug = generateURLSegment(input.Name)
	}
	base := slug
	for suffix := 2; f.hasSiblingSlug(input.ParentID, slug); suffix++ {
		slug = base + "-" + string(rune('0'+suffix))
	}
	node := Node{ID: uuid.New(), ParentID: input.ParentID, Kind: input.Kind, Name: input.Name, Slug: slug, Path: "/" + slug, SortOrder: input.SortOrder}
	detail := AdminNodeDetail{Node: node, Assets: []FileAsset{}, RedirectsCreated: []PathRedirect{}}
	f.nodes[node.ID] = detail
	return detail, nil
}

func (f *fakeAdminRepository) hasSiblingSlug(parentID *uuid.UUID, slug string) bool {
	for _, detail := range f.nodes {
		if detail.Node.Slug != slug {
			continue
		}
		if parentID == nil && detail.Node.ParentID == nil {
			return true
		}
		if parentID != nil && detail.Node.ParentID != nil && *parentID == *detail.Node.ParentID {
			return true
		}
	}
	return false
}

func (f *fakeAdminRepository) UpdateNode(_ context.Context, nodeID uuid.UUID, input UpdateNodeInput) (AdminNodeDetail, error) {
	f.updateCalls++
	detail, ok := f.nodes[nodeID]
	if !ok {
		return AdminNodeDetail{}, ErrNodeNotFound
	}
	if input.URLPath != nil && *input.URLPath == "existing" {
		return AdminNodeDetail{}, ErrDuplicatePath
	}
	node := detail.Node
	if input.ParentIDSet {
		node.ParentID = input.ParentID
	}
	if input.Name != nil {
		node.Name = *input.Name
	}
	if input.Slug != nil {
		node.Slug = *input.Slug
		node.Path = path.Join(path.Dir(node.Path), node.Slug)
	}
	if input.URLPath != nil {
		node.Slug = *input.URLPath
		node.Path = path.Join(path.Dir(node.Path), node.Slug)
	}
	if input.SortOrder != nil {
		node.SortOrder = *input.SortOrder
	}
	detail.Node = node
	f.nodes[nodeID] = detail
	return detail, nil
}

type fakePathChangeRecorder struct {
	oldPath string
	newPath string
	err     error
}

func (f *fakePathChangeRecorder) RecordPathChange(_ context.Context, _ uuid.UUID, oldPath, newPath string) error {
	f.oldPath = oldPath
	f.newPath = newPath
	return f.err
}

func (f *fakeAdminRepository) AdminTree(context.Context) (AdminTreeResponse, error) {
	return AdminTreeResponse{Nodes: f.adminTree}, nil
}

func (f *fakeAdminRepository) ReorderChildren(_ context.Context, parentID uuid.UUID, input ReorderChildrenInput) (ReorderChildrenResult, error) {
	f.reorderParent = parentID
	f.reorderChildren = append([]uuid.UUID(nil), input.ChildIDs...)
	f.reorderVersion = input.ExpectedVersion
	return ReorderChildrenResult{ParentID: parentID, ChildIDs: f.reorderChildren, Version: input.ExpectedVersion + 1}, nil
}

func (f *fakeAdminRepository) PreviewMove(_ context.Context, nodeID uuid.UUID, input MoveNodeInput) (MovePreview, error) {
	return MovePreview{
		NodeID:          nodeID,
		DestinationPath: "/moved",
		AffectedPaths:   []string{"/old/post"},
		Redirects:       []PathRedirectPreview{{OldPath: "/old/post", NewPath: "/moved/post", NodeID: nodeID}},
		BlockedReasons:  []string{},
	}, nil
}

func (f *fakeAdminRepository) MoveNode(_ context.Context, nodeID uuid.UUID, input MoveNodeInput) (AdminNodeDetail, error) {
	return f.GetAdminNode(context.Background(), nodeID)
}

func (f *fakeAdminRepository) DeleteNode(context.Context, uuid.UUID) error {
	return f.deleteErr
}
