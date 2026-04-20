package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipEntry struct {
	mu          sync.Mutex
	count       int
	windowStart time.Time
}

// RateLimit returns a per-IP fixed-window rate limiter middleware.
// Each IP is limited to maxRequests per minute.
func RateLimit(maxRequests int) gin.HandlerFunc {
	var store sync.Map
	const window = time.Minute

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		v, _ := store.LoadOrStore(ip, &ipEntry{windowStart: now})
		entry := v.(*ipEntry)

		entry.mu.Lock()
		if now.Sub(entry.windowStart) >= window {
			entry.count = 0
			entry.windowStart = now
		}
		entry.count++
		exceeded := entry.count > maxRequests
		entry.mu.Unlock()

		if exceeded {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "Terlalu banyak percobaan. Coba lagi dalam beberapa saat.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
