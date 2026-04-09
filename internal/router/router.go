// Package router registers all HTTP routes and middleware order for the Gin engine.
// Order matters: global middleware runs first; then /api/v1 group adds rate limit; some groups add JWTAuth (cookie or Bearer, with silent refresh).
// Split files: *_routes.go keep URL→handler mapping small and feature-specific.
package router

import (
	"gin-api/internal/admin"
	"gin-api/internal/auth"
	"gin-api/internal/config"
	"gin-api/internal/health"
	"gin-api/internal/middleware"
	"gin-api/internal/service"
	"gin-api/internal/upload"
	"gin-api/internal/user"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Handlers groups HTTP handlers for wiring (constructor injection from app layer).
type Handlers struct {
	Health *health.Handler
	Auth   *auth.Handler
	User   *user.Handler
	Upload *upload.Handler
	Admin  *admin.Handler
}

// NewEngine builds the Gin engine with global middleware and versioned API groups.
func NewEngine(cfg *config.Config, authSvc *service.AuthService, h *Handlers) *gin.Engine {
	gin.SetMode(cfg.GinMode)
	r := gin.New()
	// Global chain (every request): identify → panic-safe JSON → access log
	r.Use(middleware.RequestID())
	r.Use(middleware.Recovery())
	r.Use(gin.Logger())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Outside /api/v1: probes for orchestrators (K8s, load balancers)
	r.GET("/health/live", h.Health.Live)
	r.GET("/health/ready", h.Health.Ready)

	v1 := r.Group("/api/v1")
	v1.Use(middleware.RateLimit(cfg.RateLimitRPS, cfg.RateBurst)) // per-IP token bucket

	authG := v1.Group("/auth")
	RegisterAuthRoutes(authG, cfg, authSvc, h.Auth)

	users := v1.Group("/users", middleware.JWTAuth(cfg, authSvc))
	RegisterUserRoutes(users, h.User)

	uploadG := v1.Group("/upload", middleware.JWTAuth(cfg, authSvc))
	RegisterUploadRoutes(uploadG, h.Upload)

	adminG := v1.Group("/admin", middleware.JWTAuth(cfg, authSvc), middleware.AdminOnly())
	RegisterAdminRoutes(adminG, h.Admin)

	return r
}
