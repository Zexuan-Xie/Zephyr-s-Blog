package tree

import (
	"context"
	"testing"
)

func TestRecentServiceNormalizesPagination(t *testing.T) {
	repo := &fakeRecentRepository{}
	service := NewRecentService(repo)

	result, err := service.Recent(context.Background(), 0, -1)
	if err != nil {
		t.Fatalf("Recent() error = %v", err)
	}
	if repo.limit != DefaultRecentLimit || repo.offset != 0 {
		t.Fatalf("pagination = %d/%d", repo.limit, repo.offset)
	}
	if result.Items == nil {
		t.Fatal("items must be a non-nil array")
	}
}

type fakeRecentRepository struct {
	limit  int
	offset int
}

func (f *fakeRecentRepository) RecentFiles(_ context.Context, limit, offset int) ([]FileEntry, error) {
	f.limit = limit
	f.offset = offset
	return []FileEntry{}, nil
}
