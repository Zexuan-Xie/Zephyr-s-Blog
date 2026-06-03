package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"xlab-blog/api/internal/users"
)

type fakeUserRepo struct {
	byID    map[uuid.UUID]users.User
	byEmail map[string]users.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byID: map[uuid.UUID]users.User{}, byEmail: map[string]users.User{}}
}

func (r *fakeUserRepo) CreateReader(_ context.Context, email, passwordHash string, displayName *string) (users.User, error) {
	if _, ok := r.byEmail[email]; ok {
		return users.User{}, users.ErrEmailExists
	}
	user := users.User{ID: uuid.New(), Email: email, PasswordHash: passwordHash, Role: users.RoleReader, DisplayName: displayName, Provider: "local", CreatedAt: time.Now()}
	r.byID[user.ID] = user
	r.byEmail[email] = user
	return user, nil
}

func (r *fakeUserRepo) FindByEmail(_ context.Context, email string) (users.User, error) {
	user, ok := r.byEmail[email]
	if !ok {
		return users.User{}, users.ErrUserNotFound
	}
	return user, nil
}

func (r *fakeUserRepo) FindByID(_ context.Context, id uuid.UUID) (users.User, error) {
	user, ok := r.byID[id]
	if !ok {
		return users.User{}, users.ErrUserNotFound
	}
	return user, nil
}

func (r *fakeUserRepo) UpsertAdmin(_ context.Context, email, passwordHash string) (users.User, error) {
	if user, ok := r.byEmail[email]; ok {
		user.Role = users.RoleAdmin
		r.byEmail[email] = user
		r.byID[user.ID] = user
		return user, nil
	}
	user := users.User{ID: uuid.New(), Email: email, PasswordHash: passwordHash, Role: users.RoleAdmin, Provider: "local", CreatedAt: time.Now()}
	r.byID[user.ID] = user
	r.byEmail[email] = user
	return user, nil
}

func TestRegisterCreatesReader(t *testing.T) {
	service := NewService(newFakeUserRepo(), NewTokenService("secret", time.Hour))
	result, err := service.Register(context.Background(), "Reader@Example.com", "long-password", nil)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if result.User.Role != users.RoleReader {
		t.Fatalf("role = %q, want reader", result.User.Role)
	}
	if result.User.Email != "reader@example.com" {
		t.Fatalf("email normalized to %q", result.User.Email)
	}
	if result.Token == "" {
		t.Fatal("missing token")
	}
}

func TestRegisterRejectsDuplicateEmail(t *testing.T) {
	repo := newFakeUserRepo()
	service := NewService(repo, NewTokenService("secret", time.Hour))
	_, err := service.Register(context.Background(), "reader@example.com", "long-password", nil)
	if err != nil {
		t.Fatalf("first Register() error = %v", err)
	}
	_, err = service.Register(context.Background(), "reader@example.com", "long-password", nil)
	if !errors.Is(err, users.ErrEmailExists) {
		t.Fatalf("duplicate Register() error = %v, want ErrEmailExists", err)
	}
}

func TestLoginRejectsInvalidCredentials(t *testing.T) {
	service := NewService(newFakeUserRepo(), NewTokenService("secret", time.Hour))
	_, err := service.Register(context.Background(), "reader@example.com", "long-password", nil)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	_, err = service.Login(context.Background(), "reader@example.com", "wrong-password")
	if !errors.Is(err, users.ErrInvalidCredential) {
		t.Fatalf("Login() error = %v, want ErrInvalidCredential", err)
	}
}

func TestSeedAdminCreatesOrUpgradesAdmin(t *testing.T) {
	repo := newFakeUserRepo()
	service := NewService(repo, NewTokenService("secret", time.Hour))
	_, err := service.Register(context.Background(), "admin@example.com", "reader-password", nil)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if err := service.SeedAdmin(context.Background(), "admin@example.com", "admin-password"); err != nil {
		t.Fatalf("SeedAdmin() error = %v", err)
	}
	user, err := repo.FindByEmail(context.Background(), "admin@example.com")
	if err != nil {
		t.Fatalf("FindByEmail() error = %v", err)
	}
	if user.Role != users.RoleAdmin {
		t.Fatalf("role = %q, want admin", user.Role)
	}
}
