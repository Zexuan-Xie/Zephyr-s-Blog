package users

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const uniqueViolationCode = "23505"

type SQLRepository struct {
	pool rowQuerier
}

func NewSQLRepository(pool *pgxpool.Pool) *SQLRepository {
	return &SQLRepository{pool: pool}
}

type rowQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (r *SQLRepository) CreateReader(ctx context.Context, email, passwordHash string, displayName *string) (User, error) {
	const query = `
		insert into users (email, password_hash, role, display_name, provider)
		values ($1, $2, 'reader', $3, 'local')
		returning id, email, password_hash, role, display_name, provider, created_at`

	user, err := scanUser(r.pool.QueryRow(ctx, query, email, passwordHash, displayName))
	if err != nil {
		if isUniqueViolation(err) {
			return User{}, ErrEmailExists
		}
		return User{}, err
	}
	return user, nil
}

func (r *SQLRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	const query = `
		select id, email, password_hash, role, display_name, provider, created_at
		from users
		where email = $1`

	user, err := scanUser(r.pool.QueryRow(ctx, query, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}
	return user, nil
}

func (r *SQLRepository) FindByID(ctx context.Context, id uuid.UUID) (User, error) {
	const query = `
		select id, email, password_hash, role, display_name, provider, created_at
		from users
		where id = $1`

	user, err := scanUser(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}
	return user, nil
}

func (r *SQLRepository) UpsertAdmin(ctx context.Context, email, passwordHash string) (User, error) {
	const query = `
		insert into users (email, password_hash, role, provider)
		values ($1, $2, 'admin', 'local')
		on conflict (email) do update set
			password_hash = excluded.password_hash,
			role = excluded.role,
			provider = excluded.provider
		returning id, email, password_hash, role, display_name, provider, created_at`

	return scanUser(r.pool.QueryRow(ctx, query, email, passwordHash))
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner) (User, error) {
	var user User
	var displayName sql.NullString
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &displayName, &user.Provider, &user.CreatedAt); err != nil {
		return User{}, err
	}
	if displayName.Valid {
		user.DisplayName = &displayName.String
	}
	return user, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode
}
