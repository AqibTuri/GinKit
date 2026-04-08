package service

import (
	"context"

	"gin-api/internal/apperrors"
	"gin-api/internal/domain"
	"gin-api/internal/repository"
	"github.com/google/uuid"
)

// UserService loads users by id and exposes CanView for authorization checks in handlers.
type UserService struct {
	users *repository.UserRepository
}

func NewUserService(users *repository.UserRepository) *UserService {
	return &UserService{users: users}
}

// GetByID returns a user if found; caller enforces authorization (self or admin).
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "DB_ERROR", "failed to load user", 500)
	}
	if u == nil {
		return nil, apperrors.ErrNotFound
	}
	return u, nil
}

func (s *UserService) CanView(actorID uuid.UUID, actorRole string, target *domain.User) bool {
	if actorRole == domain.RoleAdmin {
		return true
	}
	return actorID == target.ID
}
