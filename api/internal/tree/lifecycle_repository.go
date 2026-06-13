package tree

import (
	"context"
	"database/sql"
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
		select node_id, revision, content_format, keywords, body_raw, body_html, search_text, status,
			published_at, last_saved_at, embedding_model, embedding_status, embedding_error, embedding_updated_at
		from file_contents
		where node_id = $1`
	content, err := scanLifecycleFileContent(r.pool.QueryRow(ctx, query, nodeID))
	if errors.Is(err, pgx.ErrNoRows) {
		return FileContent{}, ErrFileContentNotFound
	}
	return content, err
}

func (r *SQLRepository) GetFileVersionState(ctx context.Context, nodeID uuid.UUID) (FileVersionState, error) {
	current, err := r.GetFileContent(ctx, nodeID)
	if err != nil {
		return FileVersionState{}, err
	}
	state := FileVersionState{Current: current, DraftAssets: []FileAsset{}, PublishedAssets: []FileAsset{}}
	previous, err := r.getPreviousContent(ctx, nodeID)
	if err == nil {
		state.Previous = &previous
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return FileVersionState{}, err
	}
	published, err := r.PublishedContent(ctx, nodeID)
	if err == nil {
		state.Published = &published
		state.HasUnpublishedChanges = !published.Visible || published.SourceRevision != current.Revision
	} else if !errors.Is(err, ErrFileContentNotFound) {
		return FileVersionState{}, err
	}
	state.DraftAssets, err = r.listFileAssetsByState(ctx, nodeID, false)
	if err != nil {
		return FileVersionState{}, err
	}
	state.PublishedAssets, err = r.listFileAssetsByState(ctx, nodeID, true)
	if err != nil {
		return FileVersionState{}, err
	}
	return state, nil
}

func (r *SQLRepository) UpsertFileContent(ctx context.Context, nodeID uuid.UUID, input UpsertFileContentInput) (FileContent, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return FileContent{}, err
	}
	defer tx.Rollback(ctx)

	current, err := scanLifecycleFileContent(tx.QueryRow(ctx, `
		select node_id, revision, content_format, keywords, body_raw, body_html, search_text, status,
			published_at, last_saved_at, embedding_model, embedding_status, embedding_error, embedding_updated_at
		from file_contents where node_id = $1 for update`, nodeID))
	if errors.Is(err, pgx.ErrNoRows) {
		const insertQuery = `
			insert into file_contents (node_id, revision, content_format, keywords, body_raw, body_html, search_text, embedding_status, last_saved_at)
			values ($1, 1, $2, $3, $4, $5, $6, 'pending', now())
			returning node_id, revision, content_format, keywords, body_raw, body_html, search_text, status,
				published_at, last_saved_at, embedding_model, embedding_status, embedding_error, embedding_updated_at`
		created, scanErr := scanLifecycleFileContent(tx.QueryRow(ctx, insertQuery, nodeID, input.ContentFormat, input.Keywords, input.BodyRaw, input.BodyHTML, input.SearchText))
		if scanErr != nil {
			return FileContent{}, scanErr
		}
		if err := tx.Commit(ctx); err != nil {
			return FileContent{}, err
		}
		return created, nil
	}
	if err != nil {
		return FileContent{}, err
	}
	if input.ExpectedRevision > 0 && input.ExpectedRevision != current.Revision {
		return FileContent{}, ErrLostUpdate
	}
	if current.ContentFormat == input.ContentFormat && current.BodyRaw == input.BodyRaw && stringPtrEqual(current.BodyHTML, input.BodyHTML) && stringSlicesEqual(current.Keywords, input.Keywords) {
		return current, tx.Commit(ctx)
	}
	if _, err := tx.Exec(ctx, `
		insert into file_content_previous_versions (node_id, revision, content_format, keywords, body_raw, body_html, search_text, status, last_saved_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		on conflict (node_id) do update set
			revision = excluded.revision,
			content_format = excluded.content_format,
			keywords = excluded.keywords,
			body_raw = excluded.body_raw,
			body_html = excluded.body_html,
			search_text = excluded.search_text,
			status = excluded.status,
			last_saved_at = excluded.last_saved_at,
			created_at = now()`, current.NodeID, current.Revision, current.ContentFormat, current.Keywords, current.BodyRaw, current.BodyHTML, current.SearchText, current.Status, current.LastSavedAt); err != nil {
		return FileContent{}, err
	}
	const updateQuery = `
		update file_contents
		set revision = revision + 1,
			content_format = $2,
			keywords = $3,
			body_raw = $4,
			body_html = $5,
			search_text = $6,
			status = case when exists(select 1 from published_file_contents p where p.node_id = $1 and p.visible) then 'unpublished_changes' else 'draft' end,
			last_saved_at = now(),
			embedding_status = 'pending',
			embedding_error = null,
			embedding_updated_at = null
		where node_id = $1
		returning node_id, revision, content_format, keywords, body_raw, body_html, search_text, status,
			published_at, last_saved_at, embedding_model, embedding_status, embedding_error, embedding_updated_at`
	updated, err := scanLifecycleFileContent(tx.QueryRow(ctx, updateQuery, nodeID, input.ContentFormat, input.Keywords, input.BodyRaw, input.BodyHTML, input.SearchText))
	if err != nil {
		return FileContent{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return FileContent{}, err
	}
	return updated, nil
}

func (r *SQLRepository) RestorePreviousContent(ctx context.Context, nodeID uuid.UUID, expectedRevision int) (FileVersionState, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return FileVersionState{}, err
	}
	defer tx.Rollback(ctx)
	current, err := scanLifecycleFileContent(tx.QueryRow(ctx, `
		select node_id, revision, content_format, keywords, body_raw, body_html, search_text, status,
			published_at, last_saved_at, embedding_model, embedding_status, embedding_error, embedding_updated_at
		from file_contents where node_id = $1 for update`, nodeID))
	if errors.Is(err, pgx.ErrNoRows) {
		return FileVersionState{}, ErrFileContentNotFound
	}
	if err != nil {
		return FileVersionState{}, err
	}
	if expectedRevision > 0 && current.Revision != expectedRevision {
		return FileVersionState{}, ErrLostUpdate
	}
	previous, err := r.getPreviousContentTx(ctx, tx, nodeID)
	if errors.Is(err, pgx.ErrNoRows) {
		return FileVersionState{}, ErrFileContentNotFound
	}
	if err != nil {
		return FileVersionState{}, err
	}
	if _, err := tx.Exec(ctx, `
		update file_content_previous_versions
		set revision = $2, content_format = $3, keywords = $4, body_raw = $5, body_html = $6, search_text = $7, status = $8, last_saved_at = $9, created_at = now()
		where node_id = $1`, current.NodeID, current.Revision, current.ContentFormat, current.Keywords, current.BodyRaw, current.BodyHTML, current.SearchText, current.Status, current.LastSavedAt); err != nil {
		return FileVersionState{}, err
	}
	_, err = tx.Exec(ctx, `
		update file_contents
		set revision = $2, content_format = $3, keywords = $4, body_raw = $5, body_html = $6, search_text = $7,
			status = case when exists(select 1 from published_file_contents p where p.node_id = $1 and p.visible) then 'unpublished_changes' else 'draft' end,
			last_saved_at = now(), embedding_status = 'pending', embedding_error = null, embedding_updated_at = null
		where node_id = $1`, nodeID, current.Revision+1, previous.ContentFormat, previous.Keywords, previous.BodyRaw, previous.BodyHTML, previous.SearchText)
	if err != nil {
		return FileVersionState{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return FileVersionState{}, err
	}
	return r.GetFileVersionState(ctx, nodeID)
}

func (r *SQLRepository) PublishCurrentSnapshot(ctx context.Context, nodeID uuid.UUID, expectedRevision int) (PublishResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return PublishResult{}, err
	}
	defer tx.Rollback(ctx)
	current, err := scanLifecycleFileContent(tx.QueryRow(ctx, `
		select node_id, revision, content_format, keywords, body_raw, body_html, search_text, status,
			published_at, last_saved_at, embedding_model, embedding_status, embedding_error, embedding_updated_at
		from file_contents where node_id = $1 for update`, nodeID))
	if errors.Is(err, pgx.ErrNoRows) {
		return PublishResult{}, ErrFileContentNotFound
	}
	if err != nil {
		return PublishResult{}, err
	}
	if expectedRevision > 0 && current.Revision != expectedRevision {
		return PublishResult{}, ErrLostUpdate
	}
	published, err := scanPublishedContent(tx.QueryRow(ctx, `
		insert into published_file_contents (node_id, source_revision, content_format, keywords, body_raw, body_html, search_text, published_at, updated_at, visible, embedding_model, embedding_status, embedding_error, embedding_updated_at)
		values ($1, $2, $3, $4, $5, $6, $7, coalesce($8, now()), now(), true, $9, $10, $11, $12)
		on conflict (node_id) do update set
			source_revision = excluded.source_revision,
			content_format = excluded.content_format,
			keywords = excluded.keywords,
			body_raw = excluded.body_raw,
			body_html = excluded.body_html,
			search_text = excluded.search_text,
			updated_at = now(),
			visible = true,
			embedding_model = excluded.embedding_model,
			embedding_status = excluded.embedding_status,
			embedding_error = excluded.embedding_error,
			embedding_updated_at = excluded.embedding_updated_at
		returning node_id, source_revision, content_format, keywords, body_raw, body_html, search_text, published_at, updated_at, visible`,
		nodeID, current.Revision, current.ContentFormat, current.Keywords, current.BodyRaw, current.BodyHTML, current.SearchText, current.PublishedAt, current.EmbeddingModel, current.EmbeddingStatus, current.EmbeddingError, current.EmbeddingUpdatedAt))
	if err != nil {
		return PublishResult{}, err
	}
	if _, err := tx.Exec(ctx, `update file_contents set status = 'published', published_at = coalesce(published_at, now()) where node_id = $1`, nodeID); err != nil {
		return PublishResult{}, err
	}
	if _, err := tx.Exec(ctx, `delete from published_file_assets where file_node_id = $1`, nodeID); err != nil {
		return PublishResult{}, err
	}
	if _, err := tx.Exec(ctx, `
		insert into published_file_assets (asset_id, file_node_id, filename, mime_type, size_bytes, storage_provider, storage_key)
		select id, file_node_id, filename, mime_type, size_bytes, storage_provider, storage_key
		from file_assets where file_node_id = $1 and state in ('draft','draft_and_published','published')`, nodeID); err != nil {
		return PublishResult{}, err
	}
	if _, err := tx.Exec(ctx, `update file_assets set state = 'draft_and_published' where file_node_id = $1`, nodeID); err != nil {
		return PublishResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return PublishResult{}, err
	}
	current, err = r.GetFileContent(ctx, nodeID)
	if err != nil {
		return PublishResult{}, err
	}
	assets, err := r.listFileAssetsByState(ctx, nodeID, true)
	if err != nil {
		return PublishResult{}, err
	}
	return PublishResult{Current: current, Published: published, PromotedAssets: assets}, nil
}

func (r *SQLRepository) PublishedContent(ctx context.Context, nodeID uuid.UUID) (PublishedContent, error) {
	const query = `
		select node_id, source_revision, content_format, keywords, body_raw, body_html, search_text, published_at, updated_at, visible
		from published_file_contents
		where node_id = $1`
	published, err := scanPublishedContent(r.pool.QueryRow(ctx, query, nodeID))
	if errors.Is(err, pgx.ErrNoRows) {
		return PublishedContent{}, ErrFileContentNotFound
	}
	return published, err
}

func (r *SQLRepository) UnpublishFile(ctx context.Context, nodeID uuid.UUID, expectedRevision int) (FileContent, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return FileContent{}, err
	}
	defer tx.Rollback(ctx)

	current, err := scanLifecycleFileContent(tx.QueryRow(ctx, `
		select node_id, revision, content_format, keywords, body_raw, body_html, search_text, status,
			published_at, last_saved_at, embedding_model, embedding_status, embedding_error, embedding_updated_at
		from file_contents where node_id = $1 for update`, nodeID))
	if errors.Is(err, pgx.ErrNoRows) {
		return FileContent{}, ErrFileContentNotFound
	}
	if err != nil {
		return FileContent{}, err
	}
	if expectedRevision <= 0 || current.Revision != expectedRevision {
		return FileContent{}, ErrLostUpdate
	}

	content, err := scanLifecycleFileContent(tx.QueryRow(ctx, `
		update file_contents
		set status = 'draft'
		where node_id = $1
		returning node_id, revision, content_format, keywords, body_raw, body_html, search_text, status,
			published_at, last_saved_at, embedding_model, embedding_status, embedding_error, embedding_updated_at`, nodeID))
	if err != nil {
		return FileContent{}, err
	}
	if _, err := tx.Exec(ctx, `update published_file_contents set visible = false where node_id = $1`, nodeID); err != nil {
		return FileContent{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return FileContent{}, err
	}
	return content, nil
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
			join published_file_contents pfc on pfc.node_id = d.id and pfc.visible
			where d.kind = 'file'
		)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, directoryID).Scan(&exists)
	return exists, err
}

func (r *SQLRepository) PublishedDescendantFilePaths(ctx context.Context, directoryID uuid.UUID) ([]PublishedFilePath, error) {
	const query = nodePathsCTE + `
		select paths.id, paths.path
		from node_paths paths
		join published_file_contents pfc on pfc.node_id = paths.id and pfc.visible
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

func (r *SQLRepository) getPreviousContent(ctx context.Context, nodeID uuid.UUID) (FileContent, error) {
	return r.getPreviousContentTx(ctx, r.pool, nodeID)
}

func (r *SQLRepository) getPreviousContentTx(ctx context.Context, q queryRower, nodeID uuid.UUID) (FileContent, error) {
	const query = `
		select node_id, revision, content_format, keywords, body_raw, body_html, search_text, status,
			null::timestamptz as published_at, last_saved_at, null::text as embedding_model, 'pending'::text as embedding_status, null::text as embedding_error, null::timestamptz as embedding_updated_at
		from file_content_previous_versions
		where node_id = $1`
	return scanLifecycleFileContent(q.QueryRow(ctx, query, nodeID))
}

type queryRower interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func scanLifecycleFileContent(row rowScanner) (FileContent, error) {
	var content FileContent
	var contentFormat string
	var status string
	var bodyHTML sql.NullString
	var publishedAt sql.NullTime
	var embeddingModel sql.NullString
	var embeddingStatus string
	var embeddingError sql.NullString
	var embeddingUpdatedAt sql.NullTime
	if err := row.Scan(
		&content.NodeID,
		&content.Revision,
		&contentFormat,
		&content.Keywords,
		&content.BodyRaw,
		&bodyHTML,
		&content.SearchText,
		&status,
		&publishedAt,
		&content.LastSavedAt,
		&embeddingModel,
		&embeddingStatus,
		&embeddingError,
		&embeddingUpdatedAt,
	); err != nil {
		return FileContent{}, err
	}
	content.ContentFormat = ContentFormat(contentFormat)
	content.Status = PublishStatus(status)
	content.EmbeddingStatus = EmbeddingStatus(embeddingStatus)
	if bodyHTML.Valid {
		content.BodyHTML = &bodyHTML.String
	}
	if publishedAt.Valid {
		content.PublishedAt = &publishedAt.Time
	}
	if embeddingModel.Valid {
		content.EmbeddingModel = &embeddingModel.String
	}
	if embeddingError.Valid {
		content.EmbeddingError = &embeddingError.String
	}
	if embeddingUpdatedAt.Valid {
		content.EmbeddingUpdatedAt = &embeddingUpdatedAt.Time
	}
	return content, nil
}

func scanPublishedContent(row rowScanner) (PublishedContent, error) {
	var content PublishedContent
	var contentFormat string
	var bodyHTML sql.NullString
	if err := row.Scan(&content.NodeID, &content.SourceRevision, &contentFormat, &content.Keywords, &content.BodyRaw, &bodyHTML, &content.SearchText, &content.PublishedAt, &content.UpdatedAt, &content.Visible); err != nil {
		return PublishedContent{}, err
	}
	content.ContentFormat = ContentFormat(contentFormat)
	if bodyHTML.Valid {
		content.BodyHTML = &bodyHTML.String
	}
	return content, nil
}

func stringPtrEqual(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
