package middleware

import (
	"errors"
	"log/slog"
	"runtime/debug"

	"gin-api/pkg/response"
	"github.com/gin-gonic/gin"
)

// Recovery turns panics into 500 + unified JSON (never crash the process on a bad handler).
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic recovered", "recover", r, "stack", string(debug.Stack()))
				response.Error(c, errors.New("internal server error"))
				c.Abort()
			}
		}()
		c.Next()
	}
}
