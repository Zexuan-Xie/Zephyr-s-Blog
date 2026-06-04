package tree

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestAdminCreateNodeRejectsReservedRootSlug(t *testing.T) {
	repo := newFakeAdminRepository()
	service := NewAdminService(repo, nil)

	_, err := service.CreateNode(context.Background(), CreateNodeInput{
		Kind: NodeKindDirectory,
		Name: "Admin",
		Slug: "admin",
	})
	if !errors.Is(err, ErrReservedRootSlug) {
		t.Fatalf("CreateNode() error = %v, want ErrReservedRootSlug", err)
	}
	if len(repo.nodes) != 0 {
		t.Fatalf("created nodes = %d, want 0", len(repo.nodes))
	}
}

type fakeAdminRepository struct {
	nodes map[uuid.UUID]Node
}

func newFakeAdminRepository() *fakeAdminRepository {
	return &fakeAdminRepository{nodes: map[uuid.UUID]Node{}}
}

func (f *fakeAdminRepository) CreateNode(_ context.Context, input CreateNodeInput) (Node, error) {
	node := Node{ID: uuid.New(), ParentID: input.ParentID, Kind: input.Kind, Name: input.Name, Slug: input.Slug, SortOrder: input.SortOrder}
	f.nodes[node.ID] = node
	return node, nil
}
