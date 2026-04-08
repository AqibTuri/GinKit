package router

import (
	"gin-api/internal/upload"
	"github.com/gin-gonic/gin"
)

// RegisterUploadRoutes mounts POST /upload (empty path → group prefix only). Parent adds JWTAuth.
func RegisterUploadRoutes(g *gin.RouterGroup, h *upload.Handler) {
	g.POST("", h.Upload)
}
