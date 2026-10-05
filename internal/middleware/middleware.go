package middleware

import (
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	ALLOWED_ORIGIN = "ALLOWED_ORIGIN"

	corsMaxAge = "86400" // browsers may cache preflight responses for 24 hours
)

// LoggingMiddleware logs method, path (without query string), status and duration.
// The log line is written in a defer, so a request that panics is logged too (as 500)
// regardless of the order in which Logging and Recovery are registered.
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			status := c.Writer.Status()
			r := recover()
			if r != nil {
				status = http.StatusInternalServerError
			}
			log.Printf("%s %s %d %v", c.Request.Method, c.Request.URL.Path, status, time.Since(start))
			if r != nil {
				panic(r) // let RecoveryMiddleware produce the response
			}
		}()
		c.Next()
	}
}

// RecoveryMiddleware turns a panic into a 500 response. The panic value and stack trace
// are written to the log only; the client gets a generic message.
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v\n%s", err, debug.Stack())
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":    "internal_server_error",
						"message": "Internal server error",
					},
				})
			}
		}()
		c.Next()
	}
}

// CORSMiddleware sets CORS headers. If ALLOWED_ORIGIN is set, only that origin is allowed;
// otherwise the request Origin is reflected (or "*" when there is none), which is
// convenient for development but should not be used in production.
func CORSMiddleware() gin.HandlerFunc {
	allowedOrigin := os.Getenv(ALLOWED_ORIGIN)

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}
		if allowedOrigin != "" {
			origin = allowedOrigin
		}

		h := c.Writer.Header()
		h.Set("Access-Control-Allow-Origin", origin)
		h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		h.Set("Access-Control-Max-Age", corsMaxAge)
		h.Add("Vary", "Origin") // the response depends on the Origin header, so caches must key on it

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}