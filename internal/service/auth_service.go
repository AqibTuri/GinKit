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
	users           *repository.UserRepository
	jwtSecret       []byte
	jwtTTL          time.Duration
	jwtRefreshTTL   time.Duration
}

func NewAuthService(users *repository.UserRepository, jwtSecret string, jwtTTL, jwtRefreshTTL time.Duration) *AuthService {
	return &AuthService{
		users:         users,
		jwtSecret:     []byte(jwtSecret),
		jwtTTL:        jwtTTL,
		jwtRefreshTTL: jwtRefreshTTL,
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

func (s *AuthService) Login(ctx context.Context, email, plainPassword string) (access, refresh string, accessExpSec, refreshExpSec int64, err error) {
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return "", "", 0, 0, apperrors.Wrap(err, "DB_ERROR", "failed to look up user", 500)
	}
	if u == nil || !password.Verify(u.PasswordHash, plainPassword) {
		return "", "", 0, 0, apperrors.ErrInvalidCredentials
	}
	acc, err := jwtutil.SignAccess(s.jwtSecret, u, s.jwtTTL)
	if err != nil {
		return "", "", 0, 0, apperrors.Wrap(err, "TOKEN_ERROR", "could not issue access token", 500)
	}
	ref, err := jwtutil.SignRefresh(s.jwtSecret, u, s.jwtRefreshTTL)
	if err != nil {
		return "", "", 0, 0, apperrors.Wrap(err, "TOKEN_ERROR", "could not issue refresh token", 500)
	}
	return acc, ref, int64(s.jwtTTL.Seconds()), int64(s.jwtRefreshTTL.Seconds()), nil
}

// Refresh validates a refresh JWT, ensures the user still exists, and returns a new access+refresh pair (rotation).
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (access, refresh string, accessExpSec, refreshExpSec int64, err error) {
	claims, err := jwtutil.ParseRefresh(s.jwtSecret, refreshToken)
	if err != nil {
		return "", "", 0, 0, apperrors.ErrUnauthorized
	}
	u, err := s.users.FindByID(ctx, claims.UserID)
	if err != nil {
		return "", "", 0, 0, apperrors.Wrap(err, "DB_ERROR", "failed to look up user", 500)
	}
	if u == nil {
		return "", "", 0, 0, apperrors.ErrUnauthorized
	}
	acc, err := jwtutil.SignAccess(s.jwtSecret, u, s.jwtTTL)
	if err != nil {
		return "", "", 0, 0, apperrors.Wrap(err, "TOKEN_ERROR", "could not issue access token", 500)
	}
	ref, err := jwtutil.SignRefresh(s.jwtSecret, u, s.jwtRefreshTTL)
	if err != nil {
		return "", "", 0, 0, apperrors.Wrap(err, "TOKEN_ERROR", "could not issue refresh token", 500)
	}
	return acc, ref, int64(s.jwtTTL.Seconds()), int64(s.jwtRefreshTTL.Seconds()), nil
}
