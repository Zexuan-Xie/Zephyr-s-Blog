package assets

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const uniqueViolationCode = "23505"

type SQLRepository struct {
	pool          *pgxpool.Pool
	publicBaseURL string
}

func NewSQLRepository(pool *pgxpool.Pool, publicBaseURL string) *SQLRepository {
	if publicBaseURL == "" {
		publicBaseURL = "/api/assets"
	}
	return &SQLRepository{pool: pool, publicBaseURL: publicBaseURL}
}

func (r *SQLRepository) FileTargetExists(ctx context.Context, fileID uuid.UUID) (bool, error) {
	const query = `select exists(select 1 from nodes where id = $1 and kind = 'file')`
	var exists bool
	err := r.pool.QueryRow(ctx, query, fileID).Scan(&exists)
	return exists, err
}

func (r *SQLRepository) FileAssetTotalBytes(ctx context.Context, fileID uuid.UUID) (int64, error) {
	const query = `select coalesce(sum(size_bytes), 0) from file_assets where file_node_id = $1`
	var total int64
	err := r.pool.QueryRow(ctx, query, fileID).Scan(&total)
	return total, err
}

func (r *SQLRepository) CreateAsset(ctx context.Context, asset FileAsset) (FileAsset, error) {
	const query = `
		insert into file_assets (id, file_node_id, filename, mime_type, size_bytes, storage_provider, storage_key)
		values ($1, $2, $3, $4, $5, $6, $7)
		returning id, file_node_id, filename, mime_type, size_bytes, storage_provider, storage_key, created_at`
	created, err := scanAsset(r.pool.QueryRow(ctx, query,
		asset.ID, asset.FileID, asset.Filename, asset.MIMEType, asset.SizeBytes, asset.StorageProvider, asset.StorageKey,
	), r.publicBaseURL)
	if err != nil {
		return FileAsset{}, mapRepositoryError(err)
	}
	return created, nil
}

func (r *SQLRepository) FindPublishedAsset(ctx context.Context, assetID uuid.UUID, filename string) (FileAsset, error) {
	const query = `
		select a.id, a.file_node_id, a.filename, a.mime_type, a.size_bytes, a.storage_provider, a.storage_key, a.state, a.published_asset_id, a.created_at
		from file_assets a
		join nodes n on n.id = a.file_node_id and n.kind = 'file'
		join published_file_contents pfc on pfc.node_id = n.id and pfc.visible
		where a.id = $1 and a.filename = $2`
	asset, err := scanAsset(r.pool.QueryRow(ctx, query, assetID, filename), r.publicBaseURL)
	if errors.Is(err, pgx.ErrNoRows) {
		return FileAsset{}, ErrAssetNotFound
	}
	return asset, err
}

func (r *SQLRepository) FindDraftAsset(ctx context.Context, assetID uuid.UUID, filename string) (FileAsset, error) {
	const query = `
		select a.id, a.file_node_id, a.filename, a.mime_type, a.size_bytes, a.storage_provider, a.storage_key, a.state, a.published_asset_id, a.created_at
		from file_assets a
		where a.id = $1 and a.filename = $2 and a.state in ('draft','draft_and_published')`
	asset, err := scanAsset(r.pool.QueryRow(ctx, query, assetID, filename), r.publicBaseURL)
	if errors.Is(err, pgx.ErrNoRows) {
		return FileAsset{}, ErrAssetNotFound
	}
	return asset, err
}

func (r *SQLRepository) ListAssetState(ctx context.Context, fileID uuid.UUID) ([]FileAsset, []FileAsset, error) {
	const query = `
		select id, file_node_id, filename, mime_type, size_bytes, storage_provider, storage_key, state, published_asset_id, created_at
		from file_assets
		where file_node_id = $1
		order by created_at, filename`
	rows, err := r.pool.Query(ctx, query, fileID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	draft := []FileAsset{}
	published := []FileAsset{}
	for rows.Next() {
		asset, err := scanAsset(rows, r.publicBaseURL)
		if err != nil {
			return nil, nil, err
		}
		if asset.State == "draft" || asset.State == "draft_and_published" {
			draft = append(draft, asset)
		}
		if asset.State == "published" || asset.State == "draft_and_published" {
			published = append(published, asset)
		}
	}
	return draft, published, rows.Err()
}

func (r *SQLRepository) PromoteDraftAssets(ctx context.Context, fileID uuid.UUID) ([]FileAsset, error) {
	_, err := r.pool.Exec(ctx, `update file_assets set state = 'draft_and_published' where file_node_id = $1`, fileID)
	if err != nil {
		return nil, err
	}
	_, published, err := r.ListAssetState(ctx, fileID)
	return published, err
}

func (r *SQLRepository) DeleteAsset(ctx context.Context, assetID uuid.UUID) (FileAsset, error) {
	const query = `
		delete from file_assets
		where id = $1 and state = 'draft'
		returning id, file_node_id, filename, mime_type, size_bytes, storage_provider, storage_key, state, published_asset_id, created_at`
	asset, err := scanAsset(r.pool.QueryRow(ctx, query, assetID), r.publicBaseURL)
	if errors.Is(err, pgx.ErrNoRows) {
		return FileAsset{}, ErrAssetNotFound
	}
	return asset, err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAsset(row rowScanner, publicBaseURL string) (FileAsset, error) {
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
	asset.PublicURL = publicBaseURL + "/" + asset.ID.String() + "/" + asset.Filename
	return asset, nil
}

func mapRepositoryError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
		return ErrDuplicateAssetName
	}
	return err
}
