package middleware

import (
	"strings"

	"gin-api/internal/apperrors"
	"gin-api/internal/authcookie"
	"gin-api/internal/config"
	"gin-api/internal/domain"
	"gin-api/internal/service"
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

func bearerToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if h == "" || !strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return ""
	}
	return strings.TrimSpace(h[7:])
}

func accessTokenFromRequest(c *gin.Context, cookieName string) string {
	if t, err := c.Cookie(cookieName); err == nil && t != "" {
		return t
	}
	return bearerToken(c)
}

func refreshTokenFromRequest(c *gin.Context, cookieName string) string {
	t, err := c.Cookie(cookieName)
	if err != nil || t == "" {
		return ""
	}
	return t
}

func setClaims(c *gin.Context, claims *jwtutil.Claims) {
	c.Set(CtxUserID, claims.UserID)
	c.Set(CtxEmail, claims.Email)
	c.Set(CtxRole, claims.Role)
}

// JWTAuth reads an access JWT from the access cookie or Authorization: Bearer, validates it,
// and stores user id/email/role in Gin context. If the access token is missing or invalid but a valid
// refresh cookie exists, it issues new tokens (rotation), sets cookies, and continues.
func JWTAuth(cfg *config.Config, authSvc *service.AuthService) gin.HandlerFunc {
	key := []byte(cfg.JWTSecret)
	cs := cfg.AuthCookieSettings()
	return func(c *gin.Context) {
		raw := accessTokenFromRequest(c, cfg.CookieAccessName)
		claims, err := jwtutil.ParseAccess(key, raw)
		if err == nil {
			setClaims(c, claims)
			c.Next()
			return
		}

		refRaw := refreshTokenFromRequest(c, cfg.CookieRefreshName)
		if refRaw == "" {
			response.Error(c, apperrors.ErrUnauthorized)
			c.Abort()
			return
		}
		access, refresh, aMax, rMax, err := authSvc.Refresh(c.Request.Context(), refRaw)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}
		authcookie.Set(c, cs, access, refresh, int(aMax), int(rMax))

		claims, err = jwtutil.ParseAccess(key, access)
		if err != nil {
			response.Error(c, apperrors.ErrUnauthorized)
			c.Abort()
			return
		}
		setClaims(c, claims)
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
