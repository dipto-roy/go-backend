package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/dip-roy/go-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type bucket struct {
	tokens   float64
	lastSeen time.Time
	mu       sync.Mutex
}

type rateLimiter struct {
	buckets map[string]*bucket
	mu      sync.RWMutex
	rate    float64
	burst   float64
}

func newRateLimiter(rps, burst float64) *rateLimiter {
	rl := &rateLimiter{
		buckets: make(map[string]*bucket),
		rate:    rps,
		burst:   burst,
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	b, ok := rl.buckets[ip]
	if !ok {
		b = &bucket{tokens: rl.burst, lastSeen: time.Now()}
		rl.buckets[ip] = b
	}
	rl.mu.Unlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastSeen).Seconds()
	b.tokens += elapsed * rl.rate
	if b.tokens > rl.burst {
		b.tokens = rl.burst
	}
	b.lastSeen = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		for ip, b := range rl.buckets {
			b.mu.Lock()
			if time.Since(b.lastSeen) > 10*time.Minute {
				delete(rl.buckets, ip)
			}
			b.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

func RateLimit(rps, burst int64) gin.HandlerFunc {
	rl := newRateLimiter(float64(rps), float64(burst))
	return func(c *gin.Context) {
		if !rl.allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, response.Response{
				Success: false,
				Error:   &response.ErrorBody{Code: "RATE_LIMITED", Message: "too many requests"},
			})
			return
		}
		c.Next()
	}
}
