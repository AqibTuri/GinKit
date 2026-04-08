// Package middleware holds Gin handlers that run before/after route handlers (cross-cutting concerns).
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const HeaderRequestID = "X-Request-ID"
const CtxRequestID = "request_id"

// RequestID ensures X-Request-ID on response; generates UUID if client omitted it (correlate logs/traces).
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(HeaderRequestID)
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Writer.Header().Set(HeaderRequestID, rid)
		c.Set(CtxRequestID, rid)
		c.Next()
	}
}
