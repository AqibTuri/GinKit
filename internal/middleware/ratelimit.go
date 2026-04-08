package middleware

import (
	"sync"

	"gin-api/pkg/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimit applies a token-bucket limiter per client IP (in-memory). For multi-instance production,
// replace with Redis-backed limiter using the same interface.
func RateLimit(rps float64, burst int) gin.HandlerFunc {
	var mu sync.Mutex
	limiters := make(map[string]*rate.Limiter)

	get := func(key string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()
		lim, ok := limiters[key]
		if !ok {
			lim = rate.NewLimiter(rate.Limit(rps), burst)
			limiters[key] = lim
		}
		return lim
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !get(ip).Allow() {
			response.TooManyRequests(c)
			c.Abort()
			return
		}
		c.Next()
	}
}
