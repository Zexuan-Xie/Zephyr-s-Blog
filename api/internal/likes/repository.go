package likes

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SQLRepository struct {
	pool *pgxpool.Pool
}

func NewSQLRepository(pool *pgxpool.Pool) *SQLRepository {
	return &SQLRepository{pool: pool}
}

func (r *SQLRepository) FileTargetExists(ctx context.Context, fileID uuid.UUID) (bool, error) {
	const query = `
		select exists(
			select 1
			from nodes n
			join published_file_contents pfc on pfc.node_id = n.id and pfc.visible
			where n.id = $1 and n.kind = 'file'
		)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, fileID).Scan(&exists)
	return exists, err
}

func (r *SQLRepository) CommentTargetState(ctx context.Context, commentID uuid.UUID) (bool, bool, error) {
	const query = `select deleted_at is not null from comments where id = $1`
	var deleted bool
	if err := r.pool.QueryRow(ctx, query, commentID).Scan(&deleted); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, false, nil
		}
		return false, false, err
	}
	return true, deleted, nil
}

func (r *SQLRepository) UpsertLike(ctx context.Context, userID uuid.UUID, target Target) error {
	_, err := r.pool.Exec(ctx, `
		insert into likes (user_id, target_type, target_id)
		values ($1, $2, $3)
		on conflict (user_id, target_type, target_id) do nothing`,
		userID, target.Type, target.ID)
	return err
}

func (r *SQLRepository) DeleteLike(ctx context.Context, userID uuid.UUID, target Target) error {
	_, err := r.pool.Exec(ctx, `delete from likes where user_id = $1 and target_type = $2 and target_id = $3`, userID, target.Type, target.ID)
	return err
}

func (r *SQLRepository) LikeState(ctx context.Context, userID uuid.UUID, target Target) (State, error) {
	const query = `
		select
			exists(select 1 from likes where user_id = $1 and target_type = $2 and target_id = $3) as liked,
			coalesce((select count(*) from likes where target_type = $2 and target_id = $3), 0) as like_count`
	var state State
	err := r.pool.QueryRow(ctx, query, userID, target.Type, target.ID).Scan(&state.Liked, &state.LikeCount)
	return state, err
}
