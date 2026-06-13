package tree

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestStage2CreateNodeGeneratesChineseURLPathAndSuffixesInitialConflicts(t *testing.T) {
	repo := newFakeAdminRepository()
	parentID := uuid.New()
	service := NewAdminService(repo, nil)

	first, err := service.CreateNode(context.Background(), CreateNodeInput{
		ParentID:      &parentID,
		Kind:          NodeKindFile,
		Name:          "  研究 笔记  ",
		ContentFormat: ContentFormatMarkdown,
	})
	if err != nil {
		t.Fatalf("first CreateNode() error = %v", err)
	}
	if first.Node.Slug != "研究-笔记" || first.Node.Path != "/研究-笔记" {
		t.Fatalf("first generated path = slug %q path %q, want Chinese-preserving /研究-笔记", first.Node.Slug, first.Node.Path)
	}

	second, err := service.CreateNode(context.Background(), CreateNodeInput{
		ParentID:      &parentID,
		Kind:          NodeKindFile,
		Name:          "研究 笔记",
		ContentFormat: ContentFormatMarkdown,
	})
	if err != nil {
		t.Fatalf("second CreateNode() error = %v", err)
	}
	if second.Node.Slug != "研究-笔记-2" || second.Node.Path != "/研究-笔记-2" {
		t.Fatalf("conflict suffix path = slug %q path %q, want /研究-笔记-2", second.Node.Slug, second.Node.Path)
	}
}

func TestStage2ExplicitURLPathConflictIsNotSilentlyRewritten(t *testing.T) {
	nodeID := uuid.New()
	repo := newFakeAdminRepository()
	repo.nodes[nodeID] = AdminNodeDetail{Node: Node{ID: nodeID, Kind: NodeKindDirectory, Name: "Notes", Slug: "notes", Path: "/notes"}}
	service := NewAdminService(repo, nil)
	urlPath := "existing"

	_, err := service.UpdateNode(context.Background(), nodeID, UpdateNodeInput{URLPath: &urlPath})
	if !errors.Is(err, ErrDuplicatePath) {
		t.Fatalf("UpdateNode explicit URLPath error = %v, want ErrDuplicatePath", err)
	}
	if got := repo.nodes[nodeID].Node.Path; got != "/notes" {
		t.Fatalf("explicit conflict changed path to %q, want original /notes", got)
	}
}

func TestStage2AdminTreeContractIncludesDraftAndPublishedOnlyStatuses(t *testing.T) {
	repo := newFakeAdminRepository()
	draftID := uuid.New()
	publishedID := uuid.New()
	repo.adminTree = []AdminTreeNode{
		{ID: draftID, Kind: NodeKindFile, Name: "草稿", URLPath: "/草稿", Status: PublishStatusDraft},
		{ID: publishedID, Kind: NodeKindFile, Name: "已发布", URLPath: "/已发布", Status: PublishStatusPublished},
	}

	result, err := NewAdminService(repo, nil).AdminTree(context.Background())
	if err != nil {
		t.Fatalf("AdminTree() error = %v", err)
	}
	if len(result.Nodes) != 2 {
		t.Fatalf("AdminTree nodes = %d, want draft and published", len(result.Nodes))
	}
	for _, node := range result.Nodes {
		if node.Status != PublishStatusDraft && node.Status != PublishStatusPublished {
			t.Fatalf("node %s status = %q, want draft/published only", node.ID, node.Status)
		}
	}
}

func TestStage2ReorderMoveAndDeleteContracts(t *testing.T) {
	repo := newFakeAdminRepository()
	service := NewAdminService(repo, nil)
	parentID := uuid.New()
	childA := uuid.New()
	childB := uuid.New()

	_, err := service.ReorderChildren(context.Background(), parentID, ReorderChildrenInput{
		ChildIDs:        []uuid.UUID{childB, childA},
		ExpectedVersion: 3,
	})
	if err != nil {
		t.Fatalf("ReorderChildren() error = %v", err)
	}
	if repo.reorderParent != parentID || repo.reorderVersion != 3 || len(repo.reorderChildren) != 2 || repo.reorderChildren[0] != childB {
		t.Fatalf("reorder call = parent %s children %#v version %d", repo.reorderParent, repo.reorderChildren, repo.reorderVersion)
	}

	nodeID := uuid.New()
	preview, err := service.PreviewMove(context.Background(), nodeID, MoveNodeInput{NewParentID: &parentID, ExpectedVersion: 4})
	if err != nil {
		t.Fatalf("PreviewMove() error = %v", err)
	}
	if preview.DestinationPath == "" || len(preview.Redirects) == 0 || len(preview.AffectedPaths) == 0 {
		t.Fatalf("move preview = %#v, want destination, redirects, affected paths", preview)
	}

	repo.deleteErr = ErrNonEmptyDirectoryDelete
	if err := service.DeleteNode(context.Background(), parentID); !errors.Is(err, ErrNonEmptyDirectoryDelete) {
		t.Fatalf("DeleteNode(non-empty directory) error = %v, want ErrNonEmptyDirectoryDelete", err)
	}
	repo.deleteErr = ErrPublishedFileDelete
	if err := service.DeleteNode(context.Background(), childA); !errors.Is(err, ErrPublishedFileDelete) {
		t.Fatalf("DeleteNode(published file) error = %v, want ErrPublishedFileDelete", err)
	}
}
