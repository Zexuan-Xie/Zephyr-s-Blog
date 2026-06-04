package tree

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetNode(ctx context.Context, nodeID uuid.UUID) (Node, error) {
	const query = `
with recursive ancestry as (
  select id, parent_id, kind, name, slug, sort_order, created_at, updated_at, array[slug] as slugs
    from nodes where id = $1
  union all
  select parent.id, parent.parent_id, parent.kind, parent.name, parent.slug, parent.sort_order,
         parent.created_at, parent.updated_at, array_prepend(parent.slug, ancestry.slugs)
    from nodes parent join ancestry on ancestry.parent_id = parent.id
)
select id, parent_id, kind, name, slug, sort_order, created_at, updated_at,
       '/' || array_to_string(slugs, '/') as path
from ancestry
order by array_length(slugs, 1) desc
limit 1`

	var node Node
	var kind string
	if err := r.pool.QueryRow(ctx, query, nodeID).Scan(
		&node.ID,
		&node.ParentID,
		&kind,
		&node.Name,
		&node.Slug,
		&node.SortOrder,
		&node.CreatedAt,
		&node.UpdatedAt,
		&node.Path,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Node{}, ErrNodeNotFound
		}
		return Node{}, err
	}
	node.Kind = NodeKind(kind)
	node.Path = NormalizePath(node.Path)
	return node, nil
}

func (r *Repository) GetFileContent(ctx context.Context, nodeID uuid.UUID) (FileContent, error) {
	const query = `
select node_id, content_format, keywords, body_raw, body_html, search_text, status,
       published_at, embedding_status, embedding_model, embedding_error, embedding_updated_at
from file_contents
where node_id = $1`
	content, err := r.scanFileContent(r.pool.QueryRow(ctx, query, nodeID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FileContent{}, ErrFileContentNotFound
		}
		return FileContent{}, err
	}
	return content, nil
}

func (r *Repository) UpsertFileContent(ctx context.Context, nodeID uuid.UUID, input UpsertFileContentInput) (FileContent, error) {
	const query = `
insert into file_contents (node_id, content_format, keywords, body_raw, body_html, search_text, embedding_status)
values ($1, $2, $3, $4, $5, $6, 'pending')
on conflict (node_id) do update set
  content_format = excluded.content_format,
  keywords = excluded.keywords,
  body_raw = excluded.body_raw,
  body_html = excluded.body_html,
  search_text = excluded.search_text,
  embedding_status = 'pending',
  embedding_error = null,
  embedding_updated_at = null
returning node_id, content_format, keywords, body_raw, body_html, search_text, status,
          published_at, embedding_status, embedding_model, embedding_error, embedding_updated_at`
	return r.scanFileContent(r.pool.QueryRow(ctx, query, nodeID, input.ContentFormat, input.Keywords, input.BodyRaw, input.BodyHTML, input.SearchText))
}

func (r *Repository) PublishFile(ctx context.Context, nodeID uuid.UUID) (FileContent, error) {
	const query = `
update file_contents
set status = 'published',
    published_at = coalesce(published_at, now())
where node_id = $1
returning node_id, content_format, keywords, body_raw, body_html, search_text, status,
          published_at, embedding_status, embedding_model, embedding_error, embedding_updated_at`
	content, err := r.scanFileContent(r.pool.QueryRow(ctx, query, nodeID))
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return FileContent{}, ErrFileContentNotFound
	}
	return content, err
}

func (r *Repository) UnpublishFile(ctx context.Context, nodeID uuid.UUID) (FileContent, error) {
	const query = `
update file_contents
set status = 'draft'
where node_id = $1
returning node_id, content_format, keywords, body_raw, body_html, search_text, status,
          published_at, embedding_status, embedding_model, embedding_error, embedding_updated_at`
	content, err := r.scanFileContent(r.pool.QueryRow(ctx, query, nodeID))
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return FileContent{}, ErrFileContentNotFound
	}
	return content, err
}

func (r *Repository) DeleteNode(ctx context.Context, nodeID uuid.UUID) error {
	commandTag, err := r.pool.Exec(ctx, `delete from nodes where id = $1`, nodeID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return ErrNodeNotFound
	}
	return nil
}

func (r *Repository) HasPublishedDescendantFiles(ctx context.Context, directoryID uuid.UUID) (bool, error) {
	const query = `
with recursive descendants as (
  select id, kind from nodes where parent_id = $1
  union all
  select child.id, child.kind from nodes child join descendants parent on child.parent_id = parent.id
)
select exists(
  select 1
  from descendants d
  join file_contents fc on fc.node_id = d.id
  where d.kind = 'file' and fc.status = 'published'
)`
	var exists bool
	if err := r.pool.QueryRow(ctx, query, directoryID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *Repository) PublishedDescendantFilePaths(ctx context.Context, directoryID uuid.UUID) ([]PublishedFilePath, error) {
	const query = `
with recursive descendants as (
  select id, parent_id, kind, slug, array[slug] as relative_slugs
    from nodes where parent_id = $1
  union all
  select child.id, child.parent_id, child.kind, child.slug, descendants.relative_slugs || child.slug
    from nodes child join descendants on child.parent_id = descendants.id
), directory_path as (
  with recursive ancestry as (
    select id, parent_id, slug, array[slug] as slugs from nodes where id = $1
    union all
    select parent.id, parent.parent_id, parent.slug, array_prepend(parent.slug, ancestry.slugs)
      from nodes parent join ancestry on ancestry.parent_id = parent.id
  )
  select '/' || array_to_string(slugs, '/') as path
  from ancestry
  order by array_length(slugs, 1) desc
  limit 1
)
select d.id, directory_path.path || '/' || array_to_string(d.relative_slugs, '/') as path
from descendants d
cross join directory_path
join file_contents fc on fc.node_id = d.id
where d.kind = 'file' and fc.status = 'published'
order by path`
	rows, err := r.pool.Query(ctx, query, directoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []PublishedFilePath
	for rows.Next() {
		var file PublishedFilePath
		if err := rows.Scan(&file.NodeID, &file.Path); err != nil {
			return nil, err
		}
		file.Path = NormalizePath(file.Path)
		files = append(files, file)
	}
	return files, rows.Err()
}

func (r *Repository) UpdateRedirectTargets(ctx context.Context, nodeID uuid.UUID, finalPath string) error {
	_, err := r.pool.Exec(ctx, `update path_redirects set new_path = $2 where node_id = $1`, nodeID, NormalizePath(finalPath))
	return err
}

func (r *Repository) UpsertPathRedirect(ctx context.Context, oldPath, newPath string, nodeID uuid.UUID) error {
	oldPath = NormalizePath(oldPath)
	newPath = NormalizePath(newPath)
	if oldPath == newPath {
		return nil
	}
	_, err := r.pool.Exec(ctx, `
insert into path_redirects (old_path, new_path, node_id)
values ($1, $2, $3)
on conflict (old_path) do update set new_path = excluded.new_path, node_id = excluded.node_id`, oldPath, newPath, nodeID)
	return err
}

func (r *Repository) scanFileContent(row pgx.Row) (FileContent, error) {
	var content FileContent
	var contentFormat string
	var status string
	var embeddingStatus string
	if err := row.Scan(
		&content.NodeID,
		&contentFormat,
		&content.Keywords,
		&content.BodyRaw,
		&content.BodyHTML,
		&content.SearchText,
		&status,
		&content.PublishedAt,
		&embeddingStatus,
		&content.EmbeddingModel,
		&content.EmbeddingError,
		&content.EmbeddingUpdatedAt,
	); err != nil {
		return FileContent{}, err
	}
	content.ContentFormat = ContentFormat(contentFormat)
	content.Status = PublishStatus(status)
	content.EmbeddingStatus = EmbeddingStatus(embeddingStatus)
	return content, nil
}

func SearchTextFromParts(parts ...string) string {
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func ValidateContentFormat(format ContentFormat) error {
	switch format {
	case ContentFormatMarkdown, ContentFormatHTMLDocument:
		return nil
	default:
		return fmt.Errorf("invalid content_format %q", format)
	}
}
