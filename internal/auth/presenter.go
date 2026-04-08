package auth

import (
	"gin-api/internal/domain"
	"github.com/google/uuid"
)

// presenter.go — converts internal models / claims to auth.UserOut / TokenOut (safe for clients).

// PresentUser maps a domain user to API JSON (no password hash, ever).
func PresentUser(u *domain.User) UserOut {
	if u == nil {
		return UserOut{}
	}
	return UserOut{
		ID:    u.ID.String(),
		Email: u.Email,
		Role:  u.Role,
	}
}

// PresentMe builds the current-user payload from token claims (no DB round-trip).
func PresentMe(id uuid.UUID, email, role string) UserOut {
	return UserOut{
		ID:    id.String(),
		Email: email,
		Role:  role,
	}
}

func PresentToken(access string, expiresSec int64) TokenOut {
	return TokenOut{
		AccessToken:      access,
		ExpiresInSeconds: expiresSec,
	}
}
