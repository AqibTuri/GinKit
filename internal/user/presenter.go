package user

import "gin-api/internal/domain"

// PresentPublic strips internal fields; never add PasswordHash here.
func PresentPublic(u *domain.User) UserOut {
	if u == nil {
		return UserOut{}
	}
	return UserOut{
		ID:    u.ID.String(),
		Email: u.Email,
		Role:  u.Role,
	}
}
