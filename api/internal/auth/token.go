package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"xlab-blog/api/internal/users"
)

const DefaultTokenTTL = 24 * time.Hour

type Claims struct {
	Role  users.Role `json:"role"`
	Email string     `json:"email"`
	jwt.RegisteredClaims
}

type TokenService struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewTokenService(secret string, ttl time.Duration) *TokenService {
	if ttl == 0 {
		ttl = DefaultTokenTTL
	}
	return &TokenService{
		secret: []byte(secret),
		ttl:    ttl,
		now:    time.Now,
	}
}

func (s *TokenService) Issue(user users.User) (string, error) {
	now := s.now().UTC()
	claims := Claims{
		Role:  user.Role,
		Email: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

func (s *TokenService) Parse(tokenString string) (Claims, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return Claims{}, err
	}
	if token == nil || !token.Valid {
		return Claims{}, errors.New("invalid token")
	}
	if _, err := uuid.Parse(claims.Subject); err != nil {
		return Claims{}, fmt.Errorf("invalid subject: %w", err)
	}
	if claims.Role != users.RoleAdmin && claims.Role != users.RoleReader {
		return Claims{}, errors.New("invalid role")
	}
	return claims, nil
}
