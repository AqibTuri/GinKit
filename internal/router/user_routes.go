package router

import (
	"gin-api/internal/user"
	"github.com/gin-gonic/gin"
)

// RegisterUserRoutes mounts /users/:id. Parent group must already apply JWTAuth (see router.go).
func RegisterUserRoutes(g *gin.RouterGroup, h *user.Handler) {
	g.GET("/:id", h.GetByID)
}
