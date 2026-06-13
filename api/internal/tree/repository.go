package tree

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const nodePathsCTE = `
	with recursive node_paths as (
		select id, parent_id, kind, name, slug, sort_order, created_at, updated_at,
			('/' || slug)::text as path
		from nodes
		where parent_id is null
		union all
		select n.id, n.parent_id, n.kind, n.name, n.slug, n.sort_order, n.created_at, n.updated_at,
			(np.path || '/' || n.slug)::text as path
		from nodes n
		join node_paths np on np.id = n.parent_id
	)`

type SQLRepository struct {
	pool *pgxpool.Pool
}

func NewSQLRepository(pool *pgxpool.Pool) *SQLRepository {
	return &SQLRepository{pool: pool}
}

func (r *SQLRepository) DirectoryPage(ctx context.Context, parentID *uuid.UUID) (DirectoryPage, error) {
	parentArg := uuidArg(parentID)
	var node *Node
	pagePath := "/"
	if parentID != nil {
		parent, err := r.findDirectoryByID(ctx, *parentID)
		if err != nil {
			return DirectoryPage{}, err
		}
		node = &parent
		pagePath = parent.Path
	}

	const query = nodePathsCTE + `
		select p.id, p.parent_id, p.kind, p.name, p.slug, p.path, p.sort_order, p.created_at, p.updated_at,
			coalesce((
				select count(*) from nodes child
				where child.parent_id = p.id and child.kind = 'directory'
			), 0) as child_directory_count,
			coalesce((
				select count(*) from nodes child
				join published_file_contents child_pfc on child_pfc.node_id = child.id and child_pfc.visible
				where child.parent_id = p.id and child.kind = 'file'
			), 0) as child_file_count,
			pfc.content_format, 'published'::text as status, coalesce(pfc.keywords, '{}'::text[]) as keywords, pfc.published_at,
			coalesce((select count(*) from likes l where l.target_type = 'file' and l.target_id = p.id), 0) as like_count,
			coalesce((select count(*) from comments c where c.file_node_id = p.id and c.deleted_at is null), 0) as comment_count,
			pfc.search_text
		from node_paths p
		left join published_file_contents pfc on pfc.node_id = p.id and pfc.visible
		where (($1::uuid is null and p.parent_id is null) or p.parent_id = $1::uuid)
			and (p.kind = 'directory' or pfc.node_id is not null)
		order by p.kind, p.sort_order, p.name, p.slug`

	rows, err := r.pool.Query(ctx, query, parentArg)
	if err != nil {
		return DirectoryPage{}, err
	}
	defer rows.Close()

	entries := make([]any, 0)
	for rows.Next() {
		entry, err := scanDirectoryChild(rows)
		if err != nil {
			return DirectoryPage{}, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return DirectoryPage{}, err
	}

	return DirectoryPage{Node: node, Path: pagePath, Entries: entries}, nil
}

func (r *SQLRepository) FilePage(ctx context.Context, node Node) (FilePage, error) {
	const query = `
		select pfc.node_id, pfc.source_revision, pfc.content_format, pfc.keywords, pfc.body_raw, pfc.body_html, pfc.search_text,
			'published'::text as status, pfc.published_at, pfc.updated_at as last_saved_at, pfc.embedding_model, pfc.embedding_status, pfc.embedding_error,
			pfc.embedding_updated_at,
			coalesce((select count(*) from likes l where l.target_type = 'file' and l.target_id = pfc.node_id), 0) as like_count,
			coalesce((select count(*) from comments c where c.file_node_id = pfc.node_id and c.deleted_at is null), 0) as comment_count
		from published_file_contents pfc
		where pfc.node_id = $1 and pfc.visible`

	content, likeCount, commentCount, err := scanFilePageContent(r.pool.QueryRow(ctx, query, node.ID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FilePage{}, ErrNotFound
		}
		return FilePage{}, err
	}
	assets, err := r.listFileAssetsByState(ctx, node.ID, true)
	if err != nil {
		return FilePage{}, err
	}
	return FilePage{
		Node:           node,
		Content:        content,
		KeywordsPublic: PublicKeywords(content.Keywords),
		LikeCount:      likeCount,
		ViewerHasLiked: false,
		CommentCount:   commentCount,
		Assets:         assets,
	}, nil
}

func (r *SQLRepository) FindNodeByParentAndSlug(ctx context.Context, parentID *uuid.UUID, slug string) (Node, error) {
	parentArg := uuidArg(parentID)
	const query = nodePathsCTE + `
		select id, parent_id, kind, name, slug, path, sort_order, created_at, updated_at
		from node_paths
		where (($1::uuid is null and parent_id is null) or parent_id = $1::uuid)
			and slug = $2`

	node, err := scanNode(r.pool.QueryRow(ctx, query, parentArg, slug))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Node{}, ErrNotFound
		}
		return Node{}, err
	}
	return node, nil
}

func (r *SQLRepository) RedirectPath(ctx context.Context, oldPath string) (string, error) {
	const query = `
		select pr.new_path
		from path_redirects pr
		join nodes n on n.id = pr.node_id and n.kind = 'file'
		join published_file_contents pfc on pfc.node_id = n.id and pfc.visible
		where pr.old_path = $1`
	var newPath string
	if err := r.pool.QueryRow(ctx, query, oldPath).Scan(&newPath); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return newPath, nil
}

func (r *SQLRepository) findDirectoryByID(ctx context.Context, id uuid.UUID) (Node, error) {
	const query = nodePathsCTE + `
		select id, parent_id, kind, name, slug, path, sort_order, created_at, updated_at
		from node_paths
		where id = $1 and kind = 'directory'`
	node, err := scanNode(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Node{}, ErrNotFound
		}
		return Node{}, err
	}
	return node, nil
}

func uuidArg(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return *id
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanNode(row rowScanner) (Node, error) {
	var node Node
	var parentID uuid.NullUUID
	var kind string
	if err := row.Scan(&node.ID, &parentID, &kind, &node.Name, &node.Slug, &node.Path, &node.SortOrder, &node.CreatedAt, &node.UpdatedAt); err != nil {
		return Node{}, err
	}
	if parentID.Valid {
		node.ParentID = &parentID.UUID
	}
	node.Kind = NodeKind(kind)
	return node, nil
}

func scanDirectoryChild(row rowScanner) (any, error) {
	var node Node
	var parentID uuid.NullUUID
	var kind string
	var childDirectoryCount int
	var childFileCount int
	var contentFormat sql.NullString
	var status sql.NullString
	var keywords []string
	var publishedAt sql.NullTime
	var likeCount int
	var commentCount int
	var searchText sql.NullString

	if err := row.Scan(
		&node.ID,
		&parentID,
		&kind,
		&node.Name,
		&node.Slug,
		&node.Path,
		&node.SortOrder,
		&node.CreatedAt,
		&node.UpdatedAt,
		&childDirectoryCount,
		&childFileCount,
		&contentFormat,
		&status,
		&keywords,
		&publishedAt,
		&likeCount,
		&commentCount,
		&searchText,
	); err != nil {
		return nil, err
	}
	if parentID.Valid {
		node.ParentID = &parentID.UUID
	}
	node.Kind = NodeKind(kind)

	if node.Kind == NodeKindDirectory {
		return DirectoryEntry{Node: node, ChildDirectoryCount: childDirectoryCount, ChildFileCount: childFileCount}, nil
	}

	var published *time.Time
	if publishedAt.Valid {
		published = &publishedAt.Time
	}
	readingTime := readingTimeMinutes(searchText.String)
	return FileEntry{
		Node:               node,
		ContentFormat:      ContentFormat(contentFormat.String),
		Status:             PublishStatus(status.String),
		Keywords:           keywords,
		PublishedAt:        published,
		LikeCount:          likeCount,
		CommentCount:       commentCount,
		ReadingTimeMinutes: &readingTime,
	}, nil
}

func scanFilePageContent(row rowScanner) (FileContent, int, int, error) {
	var content FileContent
	var contentFormat string
	var status string
	var bodyHTML sql.NullString
	var publishedAt sql.NullTime
	var embeddingModel sql.NullString
	var embeddingStatus string
	var embeddingError sql.NullString
	var embeddingUpdatedAt sql.NullTime
	var likeCount int
	var commentCount int
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
		&likeCount,
		&commentCount,
	); err != nil {
		return FileContent{}, 0, 0, err
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
	return content, likeCount, commentCount, nil
}

func readingTimeMinutes(text string) int {
	words := len(strings.Fields(text))
	if words == 0 {
		return 1
	}
	minutes := words / 200
	if words%200 != 0 {
		minutes++
	}
	if minutes < 1 {
		return 1
	}
	return minutes
}

func (r *SQLRepository) listFileAssets(ctx context.Context, fileID uuid.UUID) ([]FileAsset, error) {
	return r.listFileAssetsByState(ctx, fileID, false)
}

func (r *SQLRepository) listFileAssetsByState(ctx context.Context, fileID uuid.UUID, publishedOnly bool) ([]FileAsset, error) {
	query := `
		select id, file_node_id, filename, mime_type, size_bytes, storage_provider, storage_key, state, published_asset_id, created_at
		from file_assets
		where file_node_id = $1`
	if publishedOnly {
		query += ` and state in ('published','draft_and_published')`
	}
	query += ` order by created_at, filename`
	rows, err := r.pool.Query(ctx, query, fileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	assets := make([]FileAsset, 0)
	for rows.Next() {
		asset, err := scanFileAsset(rows)
		if err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}
	return assets, rows.Err()
}

func scanFileAsset(row rowScanner) (FileAsset, error) {
	var asset FileAsset
	if err := row.Scan(
		&asset.ID,
		&asset.FileID,
		&asset.Filename,
		&asset.MIMEType,
		&asset.SizeBytes,
		&asset.StorageProvider,
		&asset.StorageKey,
		&asset.State,
		&asset.PublishedAssetID,
		&asset.CreatedAt,
	); err != nil {
		return FileAsset{}, err
	}
	asset.PublicURL = "/api/assets/" + asset.ID.String() + "/" + asset.Filename
	return asset, nil
}
