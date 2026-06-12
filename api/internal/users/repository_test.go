package users

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestUpsertAdminMakesConfiguredSeedAuthoritative(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("pgx.Connect() error = %v", err)
	}
	defer conn.Close(ctx)

	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, `
		create temporary table users (
			id uuid primary key default gen_random_uuid(),
			email text not null unique,
			password_hash text not null,
			role text not null,
			display_name text,
			provider text not null,
			created_at timestamptz not null default now()
		) on commit drop`); err != nil {
		t.Fatalf("create temporary users table: %v", err)
	}
	if _, err := tx.Exec(ctx, `
		insert into users (email, password_hash, role, provider)
		values ('author@example.com', 'old-reader-hash', 'reader', 'github')`); err != nil {
		t.Fatalf("insert prior reader: %v", err)
	}

	repo := &SQLRepository{pool: tx}
	got, err := repo.UpsertAdmin(ctx, "author@example.com", "configured-author-hash")
	if err != nil {
		t.Fatalf("UpsertAdmin() error = %v", err)
	}
	if got.PasswordHash != "configured-author-hash" {
		t.Fatalf("password hash = %q, want configured-author-hash", got.PasswordHash)
	}
	if got.Role != RoleAdmin {
		t.Fatalf("role = %q, want admin", got.Role)
	}
	if got.Provider != "local" {
		t.Fatalf("provider = %q, want local", got.Provider)
	}
}
