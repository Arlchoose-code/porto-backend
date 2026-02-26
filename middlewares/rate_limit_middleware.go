package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	max      int
	window   time.Duration
}

var toolLimiter = &rateLimiter{
	requests: make(map[string][]time.Time),
	max:      20,
	window:   time.Minute,
}

// contactLimiter — lebih ketat: max 5 pesan per 10 menit per IP
var contactLimiter = &rateLimiter{
	requests: make(map[string][]time.Time),
	max:      5,
	window:   10 * time.Minute,
}

func init() {
	go toolLimiter.cleanup()
	go contactLimiter.cleanup()
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	var valid []time.Time
	for _, t := range rl.requests[ip] {
		if t.After(windowStart) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.max {
		rl.requests[ip] = valid
		return false
	}

	rl.requests[ip] = append(valid, now)
	return true
}

func (rl *rateLimiter) cleanup() {
	for {
		time.Sleep(5 * time.Minute)
		rl.mu.Lock()
		now := time.Now()
		for ip, times := range rl.requests {
			var valid []time.Time
			for _, t := range times {
				if t.After(now.Add(-rl.window)) {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rl.requests, ip)
			} else {
				rl.requests[ip] = valid
			}
		}
		rl.mu.Unlock()
	}
}

func ToolRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !toolLimiter.allow(c.ClientIP()) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "Too many requests (max 20/minute)",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func ContactRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !contactLimiter.allow(c.ClientIP()) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "Too many messages sent. Please wait a few minutes before trying again.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
