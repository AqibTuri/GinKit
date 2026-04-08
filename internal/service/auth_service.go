package service

import (
	"context"
	"time"

	"gin-api/internal/apperrors"
	"gin-api/internal/domain"
	"gin-api/internal/repository"
	"gin-api/pkg/jwtutil"
	"gin-api/pkg/password"
	"github.com/google/uuid"
)

// AuthService implements registration and login (password hashing + JWT issuance).
type AuthService struct {
	users     *repository.UserRepository
	jwtSecret []byte
	jwtTTL    time.Duration
}

func NewAuthService(users *repository.UserRepository, jwtSecret string, jwtTTL time.Duration) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: []byte(jwtSecret),
		jwtTTL:    jwtTTL,
	}
}

func (s *AuthService) Register(ctx context.Context, email, plainPassword string) (*domain.User, error) {
	existing, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, apperrors.Wrap(err, "DB_ERROR", "failed to look up user", 500)
	}
	if existing != nil {
		return nil, apperrors.ErrConflict
	}
	hash, err := password.Hash(plainPassword)
	if err != nil {
		return nil, apperrors.Wrap(err, "HASH_ERROR", "could not process password", 500)
	}
	u := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hash,
		Role:         domain.RoleUser,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, apperrors.Wrap(err, "DB_ERROR", "could not create user", 500)
	}
	return u, nil
}

func (s *AuthService) Login(ctx context.Context, email, plainPassword string) (token string, expiresSec int64, err error) {
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return "", 0, apperrors.Wrap(err, "DB_ERROR", "failed to look up user", 500)
	}
	if u == nil || !password.Verify(u.PasswordHash, plainPassword) {
		return "", 0, apperrors.ErrInvalidCredentials
	}
	tok, err := jwtutil.Sign(s.jwtSecret, u, s.jwtTTL)
	if err != nil {
		return "", 0, apperrors.Wrap(err, "TOKEN_ERROR", "could not issue token", 500)
	}
	return tok, int64(s.jwtTTL.Seconds()), nil
}
