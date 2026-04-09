// Package authcookie sets HttpOnly auth cookies so middleware and handlers stay decoupled.
package authcookie

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Settings names and flags for access/refresh cookies.
type Settings struct {
	AccessName  string
	RefreshName string
	Path        string
	Domain      string
	Secure      bool
	SameSite    http.SameSite
}

// Set writes access and refresh tokens as HttpOnly cookies (max age = seconds).
func Set(c *gin.Context, s Settings, access, refresh string, accessMaxAgeSec, refreshMaxAgeSec int) {
	c.SetSameSite(s.SameSite)
	c.SetCookie(s.AccessName, access, accessMaxAgeSec, s.Path, s.Domain, s.Secure, true)
	c.SetCookie(s.RefreshName, refresh, refreshMaxAgeSec, s.Path, s.Domain, s.Secure, true)
}

// Clear removes auth cookies (same names/path/domain/secure as Set).
func Clear(c *gin.Context, s Settings) {
	c.SetSameSite(s.SameSite)
	c.SetCookie(s.AccessName, "", -1, s.Path, s.Domain, s.Secure, true)
	c.SetCookie(s.RefreshName, "", -1, s.Path, s.Domain, s.Secure, true)
}
