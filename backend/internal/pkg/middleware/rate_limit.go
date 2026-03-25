package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type bucket struct {
	hits []time.Time
}

type MemoryRateLimiter struct {
	mu      sync.Mutex
	store   map[string]*bucket
	limit   int
	window  time.Duration
	keyFunc func(*gin.Context) string
}

func NewMemoryRateLimiter(limit int, window time.Duration, keyFunc func(*gin.Context) string) *MemoryRateLimiter {
	if limit <= 0 {
		limit = 60
	}
	if window <= 0 {
		window = time.Minute
	}
	if keyFunc == nil {
		keyFunc = func(c *gin.Context) string {
			return c.ClientIP()
		}
	}
	return &MemoryRateLimiter{
		store:   make(map[string]*bucket),
		limit:   limit,
		window:  window,
		keyFunc: keyFunc,
	}
}

func (r *MemoryRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := r.keyFunc(c)
		if key == "" {
			key = c.ClientIP()
		}
		now := time.Now()
		cutoff := now.Add(-r.window)

		r.mu.Lock()
		b := r.store[key]
		if b == nil {
			b = &bucket{hits: make([]time.Time, 0, r.limit)}
			r.store[key] = b
		}
		filtered := b.hits[:0]
		for _, ts := range b.hits {
			if ts.After(cutoff) {
				filtered = append(filtered, ts)
			}
		}
		b.hits = filtered
		if len(b.hits) >= r.limit {
			retryAfter := int(r.window.Seconds())
			r.mu.Unlock()
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			c.AbortWithStatusJSON(429, gin.H{"error": "too many requests"})
			return
		}
		b.hits = append(b.hits, now)
		r.mu.Unlock()
		c.Next()
	}
}

func KeyByIP(c *gin.Context) string {
	return c.ClientIP()
}

func KeyByUserOrIP(c *gin.Context) string {
	if userID := c.GetInt64("userID"); userID > 0 {
		return fmt.Sprintf("u:%d", userID)
	}
	if uid := c.GetUint64("user_id"); uid > 0 {
		return fmt.Sprintf("u:%d", uid)
	}
	return "ip:" + c.ClientIP()
}
