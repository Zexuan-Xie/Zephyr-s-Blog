package comments

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"time"

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

func (r *SQLRepository) PublishedFileExists(ctx context.Context, fileID uuid.UUID) (bool, error) {
	const query = `
		select exists(
			select 1
			from nodes n
			join file_contents fc on fc.node_id = n.id and fc.status = 'published'
			where n.id = $1 and n.kind = 'file'
		)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, fileID).Scan(&exists)
	return exists, err
}

func (r *SQLRepository) ListThread(ctx context.Context, fileID uuid.UUID, viewerID *uuid.UUID) (Thread, error) {
	rows, err := r.pool.Query(ctx, listCommentsQuery, fileID, uuidArg(viewerID))
	if err != nil {
		return Thread{}, err
	}
	defer rows.Close()

	commentsByID := map[uuid.UUID]*Comment{}
	order := make([]uuid.UUID, 0)
	for rows.Next() {
		comment, err := scanComment(rows)
		if err != nil {
			return Thread{}, err
		}
		comment.Replies = []Comment{}
		commentsByID[comment.ID] = &comment
		order = append(order, comment.ID)
	}
	if err := rows.Err(); err != nil {
		return Thread{}, err
	}

	thread := Thread{FileID: fileID, Comments: []Comment{}}
	for _, id := range order {
		comment := commentsByID[id]
		if comment.ParentID == nil {
			thread.Comments = append(thread.Comments, *comment)
			continue
		}
		parent, ok := commentsByID[*comment.ParentID]
		if !ok {
			continue
		}
		parent.Replies = append(parent.Replies, *comment)
	}
	sort.SliceStable(thread.Comments, func(i, j int) bool {
		return thread.Comments[i].CreatedAt.Before(thread.Comments[j].CreatedAt)
	})
	for i := range thread.Comments {
		sort.SliceStable(thread.Comments[i].Replies, func(a, b int) bool {
			return thread.Comments[i].Replies[a].CreatedAt.Before(thread.Comments[i].Replies[b].CreatedAt)
		})
	}
	return thread, nil
}

func (r *SQLRepository) FindComment(ctx context.Context, commentID uuid.UUID) (Comment, error) {
	comment, err := scanComment(r.pool.QueryRow(ctx, findCommentQuery, commentID, nil))
	if errors.Is(err, pgx.ErrNoRows) {
		return Comment{}, ErrCommentNotFound
	}
	return comment, err
}

func (r *SQLRepository) InsertComment(ctx context.Context, fileID uuid.UUID, userID uuid.UUID, input CreateInput) (Comment, error) {
	comment, err := scanComment(r.pool.QueryRow(ctx, insertCommentQuery, fileID, userID, input.ParentID, input.ReplyToUserID, input.Body, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Comment{}, ErrCommentNotFound
	}
	return comment, err
}

func (r *SQLRepository) SoftDeleteComment(ctx context.Context, commentID uuid.UUID, deletedBy uuid.UUID) (Comment, error) {
	comment, err := scanComment(r.pool.QueryRow(ctx, softDeleteCommentQuery, commentID, deletedBy, deletedBy))
	if errors.Is(err, pgx.ErrNoRows) {
		return Comment{}, ErrCommentNotFound
	}
	return comment, err
}

const commentSelectColumns = `
	c.id,
	c.file_node_id,
	c.parent_id,
	c.reply_to_user_id,
	u.id as user_id,
	coalesce(nullif(u.display_name, ''), u.email) as display_name,
	case when c.deleted_at is null then c.body else '' end as body,
	c.created_at,
	c.updated_at,
	c.deleted_at,
	(c.deleted_at is not null) as deleted,
	coalesce((select count(*) from likes l where l.target_type = 'comment' and l.target_id = c.id), 0) as like_count,
	case when $2::uuid is null then false else exists(
		select 1 from likes viewer_like
		where viewer_like.target_type = 'comment'
			and viewer_like.target_id = c.id
			and viewer_like.user_id = $2::uuid
	) end as viewer_has_liked`

const listCommentsQuery = `
	select ` + commentSelectColumns + `
	from comments c
	join users u on u.id = c.user_id
	where c.file_node_id = $1
	order by coalesce(c.parent_id, c.id), c.parent_id nulls first, c.created_at`

const findCommentQuery = `
	select ` + commentSelectColumns + `
	from comments c
	join users u on u.id = c.user_id
	where c.id = $1`

const insertCommentQuery = `
	with inserted as (
		insert into comments (file_node_id, user_id, parent_id, reply_to_user_id, body)
		values ($1, $2, $3, $4, $5)
		returning *
	)
	select ` + commentSelectColumns + `
	from inserted c
	join users u on u.id = c.user_id`

const softDeleteCommentQuery = `
	with updated as (
		update comments
		set deleted_at = coalesce(deleted_at, now()),
			deleted_by = coalesce(deleted_by, $2),
			updated_at = now()
		where id = $1
		returning *
	)
	select ` + commentSelectColumns + `
	from updated c
	join users u on u.id = c.user_id`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanComment(row rowScanner) (Comment, error) {
	var comment Comment
	var parentID uuid.NullUUID
	var replyToUserID uuid.NullUUID
	var deletedAt sql.NullTime
	if err := row.Scan(
		&comment.ID,
		&comment.FileNodeID,
		&parentID,
		&replyToUserID,
		&comment.User.ID,
		&comment.User.DisplayName,
		&comment.Body,
		&comment.CreatedAt,
		&comment.UpdatedAt,
		&deletedAt,
		&comment.Deleted,
		&comment.LikeCount,
		&comment.ViewerHasLiked,
	); err != nil {
		return Comment{}, err
	}
	if parentID.Valid {
		comment.ParentID = &parentID.UUID
	}
	if replyToUserID.Valid {
		comment.ReplyToUserID = &replyToUserID.UUID
	}
	if deletedAt.Valid {
		deletedAtValue := deletedAt.Time.UTC().Truncate(time.Microsecond)
		comment.DeletedAt = &deletedAtValue
	}
	return comment, nil
}

func uuidArg(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return *id
}
