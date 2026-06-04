package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"xlab-blog/api/internal/users"
)

func TestPasswordHashAndVerify(t *testing.T) {
	hash, err := HashPassword("correct-horse")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == "correct-horse" {
		t.Fatal("password was not hashed")
	}
	if !VerifyPassword("correct-horse", hash) {
		t.Fatal("expected valid password to verify")
	}
	if VerifyPassword("wrong", hash) {
		t.Fatal("expected invalid password to fail")
	}
}

func TestTokenIssueParseAndRejectTamper(t *testing.T) {
	tokenService := NewTokenService("secret", time.Hour)
	user := users.User{ID: uuid.New(), Email: "reader@example.com", Role: users.RoleReader}

	token, err := tokenService.Issue(user)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	claims, err := tokenService.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if claims.Subject != user.ID.String() || claims.Role != user.Role || claims.Email != user.Email {
		t.Fatalf("claims = %#v, want user claims", claims)
	}

	tampered := tamperTokenSignature(t, token)
	if _, err := tokenService.Parse(tampered); err == nil {
		t.Fatal("expected tampered token to fail")
	}
}

func TestTokenRejectsExpired(t *testing.T) {
	tokenService := NewTokenService("secret", time.Hour)
	tokenService.now = func() time.Time { return time.Unix(1000, 0) }
	user := users.User{ID: uuid.New(), Email: "reader@example.com", Role: users.RoleReader}

	token, err := tokenService.Issue(user)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	tokenService.now = func() time.Time { return time.Unix(1000, 0).Add(2 * time.Hour) }
	if _, err := tokenService.Parse(token); err == nil {
		t.Fatal("expected expired token to fail")
	}
}

func tamperTokenSignature(t *testing.T, token string) string {
	t.Helper()
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[2] == "" {
		t.Fatalf("unexpected token format: %q", token)
	}
	replacement := "A"
	if strings.HasPrefix(parts[2], replacement) {
		replacement = "B"
	}
	parts[2] = replacement + parts[2][1:]
	return strings.Join(parts, ".")
}
