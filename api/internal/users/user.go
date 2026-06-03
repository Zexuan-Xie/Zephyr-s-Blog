package users

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleReader Role = "reader"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	DisplayName  *string   `json:"display_name,omitempty"`
	Provider     string    `json:"provider"`
	CreatedAt    time.Time `json:"created_at"`
}

var (
	ErrEmailExists       = errors.New("email already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidCredential = errors.New("invalid credentials")
)

type Repository interface {
	CreateReader(ctx context.Context, email, passwordHash string, displayName *string) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByID(ctx context.Context, id uuid.UUID) (User, error)
	UpsertAdmin(ctx context.Context, email, passwordHash string) (User, error)
}
