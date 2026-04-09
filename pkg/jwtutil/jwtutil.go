// Used by AuthService (Sign*) and middleware.JWTAuth (Parse*).
package jwtutil

import (
	"fmt"
	"time"

	"gin-api/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	TokenUseAccess  = "access"
	TokenUseRefresh = "refresh"
)

type Claims struct {
	UserID uuid.UUID `json:"uid"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	Use    string    `json:"use"`
	jwt.RegisteredClaims
}

func sign(secret []byte, user *domain.User, ttl time.Duration, use string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		Use:    use,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(secret)
}

// SignAccess issues a short-lived JWT for API requests.
func SignAccess(secret []byte, user *domain.User, ttl time.Duration) (string, error) {
	return sign(secret, user, ttl, TokenUseAccess)
}

// SignRefresh issues a long-lived JWT used only to obtain new access tokens.
func SignRefresh(secret []byte, user *domain.User, ttl time.Duration) (string, error) {
	return sign(secret, user, ttl, TokenUseRefresh)
}

func parse(secret []byte, tokenString string, wantUse string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	if claims.Use != wantUse {
		return nil, fmt.Errorf("wrong token kind")
	}
	return claims, nil
}

// ParseAccess validates an access JWT.
func ParseAccess(secret []byte, tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("missing token")
	}
	return parse(secret, tokenString, TokenUseAccess)
}

// ParseRefresh validates a refresh JWT.
func ParseRefresh(secret []byte, tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("missing token")
	}
	return parse(secret, tokenString, TokenUseRefresh)
}
