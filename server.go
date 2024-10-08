package bstore

import (
	"container/list"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type IPRateLimiter struct {
	ips        map[string]*list.Element
	timestamps *list.List
	capacity   int
	mu         sync.Mutex
}

type timestampEntry struct {
	ip    string
	stamp int64
}

func NewIPRateLimiter(capacity int) *IPRateLimiter {
	return &IPRateLimiter{
		ips:        make(map[string]*list.Element),
		timestamps: list.New(),
		capacity:   capacity,
	}
}

func (bstore *ServerCfg) Cors(r *gin.Engine) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     bstore.CORS.AllowOrigins,
		AllowMethods:     bstore.CORS.AllowMethods,
		AllowHeaders:     bstore.CORS.AllowHeaders,
		ExposeHeaders:    bstore.CORS.ExposeHeaders,
		AllowCredentials: bstore.CORS.AllowCredentials,
		MaxAge:           time.Duration(bstore.CORS.MaxAge),
	}))
}

func (bstore *ServerCfg) Middleware(r *gin.Engine) {
	r.Use(bstore.check_valid_path())
	if bstore.MWare.RateLimit.Enabled {
		rateLimiter := NewIPRateLimiter(int(bstore.MWare.RateLimitCapacity))
		r.Use(checkIPRateLimit(rateLimiter, bstore.MWare.RateLimit.MaxRequests, time.Duration(bstore.MWare.RateLimit.Duration)))
	}
	r.Use(validateReadWriteKey(bstore.GetRWKey()))
}

func (rl *IPRateLimiter) CheckRateLimit(ip string, limit int64, per time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now().UnixNano()

	if elem, exists := rl.ips[ip]; exists {
		entry := elem.Value.(*timestampEntry)
		if now-entry.stamp < per.Nanoseconds() {
			return false
		}
		entry.stamp = now
		rl.timestamps.MoveToFront(elem)
	} else {
		if len(rl.ips) >= rl.capacity {
			oldest := rl.timestamps.Back()
			if oldest != nil {
				oldestEntry := oldest.Value.(*timestampEntry)
				delete(rl.ips, oldestEntry.ip)
				rl.timestamps.Remove(oldest)
			}
		}
		entry := &timestampEntry{ip: ip, stamp: now}
		elem := rl.timestamps.PushFront(entry)
		rl.ips[ip] = elem
	}

	return true
}

func validateReadWriteKey(validKey string) gin.HandlerFunc {
	protectedPaths := []string{
		"/api/upload/",
		"/api/download/",
		"/api/delete/",
		"/api/list/",
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		needsValidation := false
		for _, protectedPath := range protectedPaths {
			if strings.HasPrefix(path, protectedPath) {
				needsValidation = true
				break
			}
		}

		if !needsValidation {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			c.Abort()
			return
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return
		}

		key := strings.TrimPrefix(authHeader, bearerPrefix)
		if key != validKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid read_write key"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func checkIPRateLimit(rl *IPRateLimiter, maxRequests int64, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !rl.CheckRateLimit(ip, maxRequests, duration) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func (bstore *ServerCfg) check_valid_path() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if len(path) > bstore.MWare.MaxPathLength {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Path too long"})
			c.Abort()
			return
		}

		if bstore.MWare.OnlyBstorePaths {
			// serve public files, request to direct public files will be blocked.
			if strings.HasPrefix(path, "/bstore") {
				c.Next()
				return
			}

			validPaths := []string{
				"/api/upload/",
				"/api/download/",
				"/api/delete/",
				"/api/list/",
			}

			for _, validPath := range validPaths {
				if strings.HasPrefix(path, validPath) {
					c.Next()
					return
				}
			}

			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid path"})
			c.Abort()
		}

		c.Next()
	}
}
