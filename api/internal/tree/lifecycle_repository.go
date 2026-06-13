package tree

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *SQLRepository) GetNode(ctx context.Context, nodeID uuid.UUID) (Node, error) {
	const query = nodePathsCTE + `
		select id, parent_id, kind, name, slug, path, sort_order, created_at, updated_at
		from node_paths
		where id = $1`
	node, err := scanNode(r.pool.QueryRow(ctx, query, nodeID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Node{}, ErrNodeNotFound
	}
	return node, err
}

func (r *SQLRepository) GetFileContent(ctx context.Context, nodeID uuid.UUID) (FileContent, error) {
	const query = `
		select node_id, content_format, keywords, body_raw, body_html, search_text, status,
			published_at, embedding_model, embedding_status, embedding_error, embedding_updated_at
		from file_contents
		where node_id = $1`
	content, err := scanLifecycleFileContent(r.pool.QueryRow(ctx, query, nodeID))
	if errors.Is(err, pgx.ErrNoRows) {
		return FileContent{}, ErrFileContentNotFound
	}
	return content, err
}

func (r *SQLRepository) UpsertFileContent(ctx context.Context, nodeID uuid.UUID, input UpsertFileContentInput) (FileContent, error) {
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
			published_at, embedding_model, embedding_status, embedding_error, embedding_updated_at`
	return scanLifecycleFileContent(r.pool.QueryRow(ctx, query, nodeID, input.ContentFormat, input.Keywords, input.BodyRaw, input.BodyHTML, input.SearchText))
}

func (r *SQLRepository) PublishFile(ctx context.Context, nodeID uuid.UUID) (FileContent, error) {
	const query = `
		update file_contents
		set status = 'published',
			published_at = coalesce(published_at, now())
		where node_id = $1
		returning node_id, content_format, keywords, body_raw, body_html, search_text, status,
			published_at, embedding_model, embedding_status, embedding_error, embedding_updated_at`
	content, err := scanLifecycleFileContent(r.pool.QueryRow(ctx, query, nodeID))
	if errors.Is(err, pgx.ErrNoRows) {
		return FileContent{}, ErrFileContentNotFound
	}
	return content, err
}

func (r *SQLRepository) UnpublishFile(ctx context.Context, nodeID uuid.UUID) (FileContent, error) {
	const query = `
		update file_contents
		set status = 'draft'
		where node_id = $1
		returning node_id, content_format, keywords, body_raw, body_html, search_text, status,
			published_at, embedding_model, embedding_status, embedding_error, embedding_updated_at`
	content, err := scanLifecycleFileContent(r.pool.QueryRow(ctx, query, nodeID))
	if errors.Is(err, pgx.ErrNoRows) {
		return FileContent{}, ErrFileContentNotFound
	}
	return content, err
}

func (r *SQLRepository) DeleteNode(ctx context.Context, nodeID uuid.UUID) error {
	commandTag, err := r.pool.Exec(ctx, `delete from nodes where id = $1`, nodeID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return ErrNodeNotFound
	}
	return nil
}

func (r *SQLRepository) HasChildNodes(ctx context.Context, directoryID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `select exists(select 1 from nodes where parent_id = $1)`, directoryID).Scan(&exists)
	return exists, err
}

func (r *SQLRepository) HasPublishedDescendantFiles(ctx context.Context, directoryID uuid.UUID) (bool, error) {
	const query = `
		with recursive descendants as (
			select id, kind from nodes where parent_id = $1
			union all
			select child.id, child.kind
			from nodes child
			join descendants parent on child.parent_id = parent.id
		)
		select exists(
			select 1
			from descendants d
			join file_contents fc on fc.node_id = d.id
			where d.kind = 'file' and fc.status = 'published'
		)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, directoryID).Scan(&exists)
	return exists, err
}

func (r *SQLRepository) PublishedDescendantFilePaths(ctx context.Context, directoryID uuid.UUID) ([]PublishedFilePath, error) {
	const query = nodePathsCTE + `
		select paths.id, paths.path
		from node_paths paths
		join file_contents fc on fc.node_id = paths.id and fc.status = 'published'
		where paths.path like (
			select directory.path || '/%' from node_paths directory where directory.id = $1
		)
		order by paths.path`
	rows, err := r.pool.Query(ctx, query, directoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := make([]PublishedFilePath, 0)
	for rows.Next() {
		var file PublishedFilePath
		if err := rows.Scan(&file.NodeID, &file.Path); err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, rows.Err()
}

func (r *SQLRepository) UpdateRedirectTargets(ctx context.Context, nodeID uuid.UUID, finalPath string) error {
	_, err := r.pool.Exec(ctx, `update path_redirects set new_path = $2 where node_id = $1`, nodeID, normalizePath(finalPath))
	return err
}

func (r *SQLRepository) UpsertPathRedirect(ctx context.Context, oldPath, newPath string, nodeID uuid.UUID) error {
	oldPath = normalizePath(oldPath)
	newPath = normalizePath(newPath)
	if oldPath == newPath {
		return nil
	}
	_, err := r.pool.Exec(ctx, `
		insert into path_redirects (old_path, new_path, node_id)
		values ($1, $2, $3)
		on conflict (old_path) do update set new_path = excluded.new_path, node_id = excluded.node_id`,
		oldPath, newPath, nodeID)
	return err
}

func scanLifecycleFileContent(row rowScanner) (FileContent, error) {
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
		&content.EmbeddingModel,
		&embeddingStatus,
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
