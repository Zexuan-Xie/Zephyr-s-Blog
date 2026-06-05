package tree

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const uniqueViolationCode = "23505"

func (r *SQLRepository) GetAdminNode(ctx context.Context, nodeID uuid.UUID) (AdminNodeDetail, error) {
	node, err := r.GetNode(ctx, nodeID)
	if err != nil {
		return AdminNodeDetail{}, err
	}
	detail := AdminNodeDetail{
		Node:             node,
		Assets:           []FileAsset{},
		RedirectsCreated: []PathRedirect{},
	}
	if node.Kind == NodeKindFile {
		content, err := r.GetFileContent(ctx, nodeID)
		if err != nil {
			return AdminNodeDetail{}, err
		}
		detail.Content = &content
		assets, err := r.listFileAssets(ctx, nodeID)
		if err != nil {
			return AdminNodeDetail{}, err
		}
		detail.Assets = assets
	}
	return detail, nil
}

func (r *SQLRepository) CreateNode(ctx context.Context, input CreateNodeInput) (AdminNodeDetail, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return AdminNodeDetail{}, err
	}
	defer tx.Rollback(ctx)

	if err := validateParent(ctx, tx, input.ParentID); err != nil {
		return AdminNodeDetail{}, err
	}

	var nodeID uuid.UUID
	err = tx.QueryRow(ctx, `
		insert into nodes (parent_id, kind, name, slug, sort_order)
		values ($1, $2, $3, $4, $5)
		returning id`,
		uuidArg(input.ParentID), input.Kind, input.Name, input.Slug, input.SortOrder,
	).Scan(&nodeID)
	if err != nil {
		return AdminNodeDetail{}, mapAdminRepositoryError(err)
	}

	if input.Kind == NodeKindFile {
		_, err = tx.Exec(ctx, `
			insert into file_contents (node_id, content_format, body_raw)
			values ($1, $2, '')`,
			nodeID, input.ContentFormat,
		)
		if err != nil {
			return AdminNodeDetail{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return AdminNodeDetail{}, mapAdminRepositoryError(err)
	}
	return r.GetAdminNode(ctx, nodeID)
}

func (*SQLRepository) recordsPathChangesAtomically() {}

func (r *SQLRepository) UpdateNode(ctx context.Context, nodeID uuid.UUID, input UpdateNodeInput) (AdminNodeDetail, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return AdminNodeDetail{}, err
	}
	defer tx.Rollback(ctx)

	var lockedID uuid.UUID
	err = tx.QueryRow(ctx, `select id from nodes where id = $1 for update`, nodeID).Scan(&lockedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return AdminNodeDetail{}, ErrNodeNotFound
	}
	if err != nil {
		return AdminNodeDetail{}, err
	}

	current, err := getNodeInTransaction(ctx, tx, nodeID)
	if err != nil {
		return AdminNodeDetail{}, err
	}

	if input.ParentIDSet {
		if input.ParentID != nil && *input.ParentID == nodeID {
			return AdminNodeDetail{}, ErrNodeCycle
		}
		if err := validateParent(ctx, tx, input.ParentID); err != nil {
			return AdminNodeDetail{}, err
		}
		if input.ParentID != nil {
			cycle, err := wouldCreateNodeCycle(ctx, tx, nodeID, *input.ParentID)
			if err != nil {
				return AdminNodeDetail{}, err
			}
			if cycle {
				return AdminNodeDetail{}, ErrNodeCycle
			}
		}
	}

	var updatedID uuid.UUID
	err = tx.QueryRow(ctx, `
		update nodes
		set parent_id = case when $2::boolean then $3::uuid else parent_id end,
			name = coalesce($4::text, name),
			slug = coalesce($5::text, slug),
			sort_order = coalesce($6::integer, sort_order),
			updated_at = now()
		where id = $1
		returning id`,
		nodeID,
		input.ParentIDSet,
		uuidArg(input.ParentID),
		stringArg(input.Name),
		stringArg(input.Slug),
		intArg(input.SortOrder),
	).Scan(&updatedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return AdminNodeDetail{}, ErrNodeNotFound
	}
	if err != nil {
		return AdminNodeDetail{}, mapAdminRepositoryError(err)
	}

	updated, err := getNodeInTransaction(ctx, tx, updatedID)
	if err != nil {
		return AdminNodeDetail{}, err
	}
	if err := recordPathChangeInTransaction(ctx, tx, current, updated); err != nil {
		return AdminNodeDetail{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return AdminNodeDetail{}, mapAdminRepositoryError(err)
	}
	return r.GetAdminNode(ctx, updatedID)
}

func getNodeInTransaction(ctx context.Context, tx pgx.Tx, nodeID uuid.UUID) (Node, error) {
	const query = nodePathsCTE + `
		select id, parent_id, kind, name, slug, path, sort_order, created_at, updated_at
		from node_paths
		where id = $1`
	node, err := scanNode(tx.QueryRow(ctx, query, nodeID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Node{}, ErrNodeNotFound
	}
	return node, err
}

func recordPathChangeInTransaction(ctx context.Context, tx pgx.Tx, current Node, updated Node) error {
	oldPath := normalizePath(current.Path)
	newPath := normalizePath(updated.Path)
	if oldPath == "/" || newPath == "/" || oldPath == newPath {
		return nil
	}

	if updated.Kind == NodeKindFile {
		var status PublishStatus
		err := tx.QueryRow(ctx, `select status from file_contents where node_id = $1`, updated.ID).Scan(&status)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		if status != PublishStatusPublished {
			return nil
		}
		if err := updateRedirectTargetsInTransaction(ctx, tx, updated.ID, newPath); err != nil {
			return err
		}
		return upsertPathRedirectInTransaction(ctx, tx, oldPath, newPath, updated.ID)
	}

	const query = nodePathsCTE + `
		select paths.id, paths.path
		from node_paths paths
		join file_contents fc on fc.node_id = paths.id and fc.status = 'published'
		where paths.path like (
			select directory.path || '/%' from node_paths directory where directory.id = $1
		)
		order by paths.path`
	rows, err := tx.Query(ctx, query, updated.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var file PublishedFilePath
		if err := rows.Scan(&file.NodeID, &file.Path); err != nil {
			return err
		}
		finalPath := normalizePath(file.Path)
		oldFilePath := replacePathPrefix(finalPath, newPath, oldPath)
		if oldFilePath == finalPath {
			continue
		}
		if err := updateRedirectTargetsInTransaction(ctx, tx, file.NodeID, finalPath); err != nil {
			return err
		}
		if err := upsertPathRedirectInTransaction(ctx, tx, oldFilePath, finalPath, file.NodeID); err != nil {
			return err
		}
	}
	return rows.Err()
}

func updateRedirectTargetsInTransaction(ctx context.Context, tx pgx.Tx, nodeID uuid.UUID, finalPath string) error {
	_, err := tx.Exec(ctx, `update path_redirects set new_path = $2 where node_id = $1`, nodeID, normalizePath(finalPath))
	return err
}

func upsertPathRedirectInTransaction(ctx context.Context, tx pgx.Tx, oldPath, newPath string, nodeID uuid.UUID) error {
	oldPath = normalizePath(oldPath)
	newPath = normalizePath(newPath)
	if oldPath == newPath {
		return nil
	}
	_, err := tx.Exec(ctx, `
		insert into path_redirects (old_path, new_path, node_id)
		values ($1, $2, $3)
		on conflict (old_path) do update set new_path = excluded.new_path, node_id = excluded.node_id`,
		oldPath, newPath, nodeID)
	return err
}

func validateParent(ctx context.Context, tx pgx.Tx, parentID *uuid.UUID) error {
	if parentID == nil {
		return nil
	}
	var kind NodeKind
	err := tx.QueryRow(ctx, `select kind from nodes where id = $1`, *parentID).Scan(&kind)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrParentNotDirectory
	}
	if err != nil {
		return err
	}
	if kind != NodeKindDirectory {
		return ErrParentNotDirectory
	}
	return nil
}

func wouldCreateNodeCycle(ctx context.Context, tx pgx.Tx, nodeID, parentID uuid.UUID) (bool, error) {
	const query = `
		with recursive ancestors as (
			select id, parent_id from nodes where id = $1
			union all
			select parent.id, parent.parent_id
			from nodes parent
			join ancestors child on child.parent_id = parent.id
		)
		select exists(select 1 from ancestors where id = $2)`
	var exists bool
	err := tx.QueryRow(ctx, query, parentID, nodeID).Scan(&exists)
	return exists, err
}

func mapAdminRepositoryError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
		return ErrDuplicateSlug
	}
	return err
}

func stringArg(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func intArg(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}
