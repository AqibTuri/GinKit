// Package app is the composition root: it connects infrastructure (DB) to domain services and HTTP.
// Flow: Open Postgres → build repositories → build services → build per-feature Handlers → router.NewEngine.
// Anything imported here is "allowed" to know about concrete implementations; inner layers stay narrower.
package app

import (
	"fmt"
	"log/slog"

	"gin-api/internal/admin"
	"gin-api/internal/auth"
	"gin-api/internal/config"
	"gin-api/internal/database"
	"gin-api/internal/health"
	"gin-api/internal/repository"
	"gin-api/internal/router"
	"gin-api/internal/service"
	"gin-api/internal/upload"
	"gin-api/internal/user"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	Config *config.Config
	DB     *gorm.DB
	Engine *gin.Engine
}

func New(cfg *config.Config) (*App, error) {
	debug := cfg.Env == "development"
	db, err := database.Open(cfg.DatabaseURL, debug)
	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}

	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiry)
	userSvc := service.NewUserService(userRepo)

	h := &router.Handlers{
		Health: health.NewHandler(),
		Auth:   auth.NewHandler(authSvc),
		User:   user.NewHandler(userSvc),
		Upload: upload.NewHandler(cfg),
		Admin:  admin.NewHandler(),
	}

	engine := router.NewEngine(cfg, h)
	slog.Info("app wired", "env", cfg.Env, "gin_mode", cfg.GinMode)
	return &App{Config: cfg, DB: db, Engine: engine}, nil
}
