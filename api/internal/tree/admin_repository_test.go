package tree

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestMapAdminRepositoryErrorMapsUniqueViolation(t *testing.T) {
	err := mapAdminRepositoryError(&pgconn.PgError{Code: uniqueViolationCode})
	if !errors.Is(err, ErrDuplicateSlug) {
		t.Fatalf("mapAdminRepositoryError() = %v, want ErrDuplicateSlug", err)
	}
}

func TestMapAdminRepositoryErrorPreservesOtherErrors(t *testing.T) {
	want := errors.New("database unavailable")
	if got := mapAdminRepositoryError(want); !errors.Is(got, want) {
		t.Fatalf("mapAdminRepositoryError() = %v, want %v", got, want)
	}
}
