package router

import (
	"gin-api/internal/admin"
	"github.com/gin-gonic/gin"
)

// RegisterAdminRoutes mounts admin-only routes; parent group already chains JWTAuth + AdminOnly.
func RegisterAdminRoutes(g *gin.RouterGroup, h *admin.Handler) {
	g.GET("/ping", h.Ping)
}
