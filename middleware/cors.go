package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// defaultAllowedOrigins is used when CORS_ALLOWED_ORIGINS is not set, so the
// published portfolio console works without any deployment configuration.
var defaultAllowedOrigins = []string{
	"https://jeferson0306.github.io",
	"http://localhost:3000",
}

// allowedOrigins reads CORS_ALLOWED_ORIGINS as a comma-separated list, letting a
// new site be authorised by changing an environment variable rather than the code.
func allowedOrigins() []string {
	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	if strings.TrimSpace(raw) == "" {
		return defaultAllowedOrigins
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

// CORS answers browser preflights and echoes the request origin when it is on
// the allow list.
//
// The origin is echoed rather than answered with "*": a wildcard cannot be used
// alongside credentials, and echoing keeps that door open without reopening this
// file later. Vary: Origin is required either way, or a shared cache can hand one
// site the header computed for another.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		for _, allowed := range allowedOrigins() {
			if origin != "" && origin == allowed {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
				c.Header("Access-Control-Allow-Headers", "Content-Type")
				c.Header("Access-Control-Max-Age", "86400")
				break
			}
		}
		c.Header("Vary", "Origin")

		// A preflight is not a validation request and must not reach the handler.
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
