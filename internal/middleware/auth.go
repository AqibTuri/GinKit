package middleware

import (
	"strings"

	"gin-api/internal/apperrors"
	"gin-api/internal/domain"
	"gin-api/pkg/jwtutil"
	"gin-api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	CtxUserID = "auth_user_id"
	CtxEmail  = "auth_email"
	CtxRole   = "auth_role"
)

// JWTAuth validates Authorization: Bearer <jwt>, parses claims, stores user id/email/role in Gin context for handlers.
func JWTAuth(secret string) gin.HandlerFunc {
	key := []byte(secret)
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(strings.ToLower(h), "bearer ") {
			response.Error(c, apperrors.ErrUnauthorized)
			c.Abort()
			return
		}
		raw := strings.TrimSpace(h[7:])
		claims, err := jwtutil.Parse(key, raw)
		if err != nil {
			response.Error(c, apperrors.ErrUnauthorized)
			c.Abort()
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxEmail, claims.Email)
		c.Set(CtxRole, claims.Role)
		c.Next()
	}
}

// RequireRole returns middleware that allows only listed roles (guard).
func RequireRole(roles ...string) gin.HandlerFunc {
	allow := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allow[r] = struct{}{}
	}
	return func(c *gin.Context) {
		role, _ := c.Get(CtxRole)
		rs, ok := role.(string)
		if !ok {
			response.Error(c, apperrors.ErrForbidden)
			c.Abort()
			return
		}
		if _, ok := allow[rs]; !ok {
			response.Error(c, apperrors.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

func MustUserID(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(CtxUserID)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

func MustRole(c *gin.Context) string {
	v, _ := c.Get(CtxRole)
	rs, _ := v.(string)
	return rs
}

func MustEmail(c *gin.Context) string {
	v, _ := c.Get(CtxEmail)
	em, _ := v.(string)
	return em
}

// AdminOnly is a convenience guard for admin routes.
func AdminOnly() gin.HandlerFunc {
	return RequireRole(domain.RoleAdmin)
}
