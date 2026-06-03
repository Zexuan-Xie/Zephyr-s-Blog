package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/google/uuid"

	"xlab-blog/api/internal/users"
)

const MinPasswordLength = 8

type Service struct {
	repo   users.Repository
	tokens *TokenService
}

type AuthResult struct {
	Token string     `json:"token"`
	User  users.User `json:"user"`
}

func NewService(repo users.Repository, tokens *TokenService) *Service {
	return &Service{repo: repo, tokens: tokens}
}

func (s *Service) Register(ctx context.Context, email, password string, displayName *string) (AuthResult, error) {
	email = normalizeEmail(email)
	if err := validateEmail(email); err != nil {
		return AuthResult{}, err
	}
	if err := validatePassword(password); err != nil {
		return AuthResult{}, err
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		return AuthResult{}, fmt.Errorf("hash password: %w", err)
	}
	user, err := s.repo.CreateReader(ctx, email, passwordHash, displayName)
	if err != nil {
		return AuthResult{}, err
	}
	return s.resultFor(user)
}

func (s *Service) Login(ctx context.Context, email, password string) (AuthResult, error) {
	email = normalizeEmail(email)
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			return AuthResult{}, users.ErrInvalidCredential
		}
		return AuthResult{}, err
	}
	if !VerifyPassword(password, user.PasswordHash) {
		return AuthResult{}, users.ErrInvalidCredential
	}
	return s.resultFor(user)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (users.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) SeedAdmin(ctx context.Context, email, password string) error {
	if email == "" && password == "" {
		return nil
	}
	email = normalizeEmail(email)
	if err := validateEmail(email); err != nil {
		return err
	}
	if err := validatePassword(password); err != nil {
		return err
	}
	passwordHash, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}
	_, err = s.repo.UpsertAdmin(ctx, email, passwordHash)
	return err
}

func (s *Service) resultFor(user users.User) (AuthResult, error) {
	token, err := s.tokens.Issue(user)
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{Token: token, User: user}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validateEmail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email: %w", err)
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	}
	return nil
}
