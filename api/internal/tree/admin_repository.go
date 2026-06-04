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

func (r *SQLRepository) UpdateNode(ctx context.Context, nodeID uuid.UUID, input UpdateNodeInput) (AdminNodeDetail, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return AdminNodeDetail{}, err
	}
	defer tx.Rollback(ctx)

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

	if err := tx.Commit(ctx); err != nil {
		return AdminNodeDetail{}, mapAdminRepositoryError(err)
	}
	return r.GetAdminNode(ctx, updatedID)
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
