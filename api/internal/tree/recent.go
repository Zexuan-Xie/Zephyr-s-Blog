package tree

import (
	"context"
)

const (
	DefaultRecentLimit = 24
	MaxRecentLimit     = 100
)

type RecentRepository interface {
	RecentFiles(ctx context.Context, limit, offset int) ([]FileEntry, error)
}

type RecentService struct {
	repo RecentRepository
}

func NewRecentService(repo RecentRepository) *RecentService {
	return &RecentService{repo: repo}
}

func (s *RecentService) Recent(ctx context.Context, limit, offset int) (FileEntryList, error) {
	if limit <= 0 {
		limit = DefaultRecentLimit
	}
	if limit > MaxRecentLimit {
		limit = MaxRecentLimit
	}
	if offset < 0 {
		offset = 0
	}
	items, err := s.repo.RecentFiles(ctx, limit, offset)
	if err != nil {
		return FileEntryList{}, err
	}
	return FileEntryList{Items: items}, nil
}

func (r *SQLRepository) RecentFiles(ctx context.Context, limit, offset int) ([]FileEntry, error) {
	const query = nodePathsCTE + `
		select p.id, p.parent_id, p.kind, p.name, p.slug, p.path, p.sort_order, p.created_at, p.updated_at,
			0 as child_directory_count,
			0 as child_file_count,
			fc.content_format, fc.status, coalesce(fc.keywords, '{}'::text[]) as keywords, fc.published_at,
			coalesce((select count(*) from likes l where l.target_type = 'file' and l.target_id = p.id), 0) as like_count,
			coalesce((select count(*) from comments c where c.file_node_id = p.id and c.deleted_at is null), 0) as comment_count,
			fc.search_text
		from node_paths p
		join file_contents fc on fc.node_id = p.id
		where p.kind = 'file' and fc.status = 'published'
		order by p.updated_at desc, fc.published_at desc nulls last, p.name
		limit $1 offset $2`
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]FileEntry, 0)
	for rows.Next() {
		entry, err := scanDirectoryChild(rows)
		if err != nil {
			return nil, err
		}
		file, ok := entry.(FileEntry)
		if ok {
			items = append(items, file)
		}
	}
	return items, rows.Err()
}

var _ RecentRepository = (*SQLRepository)(nil)
