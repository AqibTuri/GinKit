package router

import (
	"gin-api/internal/auth"
	"gin-api/internal/config"
	"gin-api/internal/middleware"
	"gin-api/internal/service"
	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes maps HTTP verbs under /auth: public register/login/refresh/logout; /me uses JWTAuth.
func RegisterAuthRoutes(g *gin.RouterGroup, cfg *config.Config, authSvc *service.AuthService, h *auth.Handler) {
	g.POST("/register", h.Register)
	g.POST("/login", h.Login)
	g.POST("/refresh", h.Refresh)
	g.POST("/logout", h.Logout)
	g.GET("/me", middleware.JWTAuth(cfg, authSvc), h.Me)
}
