package router

import (
	"gin-api/internal/auth"
	"gin-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes maps HTTP verbs under /auth: public register/login; /me protected by JWTAuth(secret).
func RegisterAuthRoutes(g *gin.RouterGroup, secret string, h *auth.Handler) {
	g.POST("/register", h.Register)
	g.POST("/login", h.Login)
	g.GET("/me", middleware.JWTAuth(secret), h.Me)
}
