package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS returns a Gin middleware that sets Cross-Origin Resource Sharing headers.
//
// This is essential for GraphQL APIs because browser clients (React, Vue, etc.)
// call POST /graphql from a different origin (e.g. localhost:3000 → localhost:8080).
// Without these headers the browser blocks the request before it even leaves the tab.
//
// Why this matters more for GraphQL than REST:
// REST APIs often serve server-rendered pages on the same origin, so CORS is
// sometimes not needed. GraphQL APIs are almost always called from a separate
// frontend app, so CORS configuration is nearly always required.
func CORS(allowedOrigins string) gin.HandlerFunc {
	origins := strings.Split(allowedOrigins, ",")
	originSet := make(map[string]bool, len(origins))
	for _, o := range origins {
		originSet[strings.TrimSpace(o)] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Allow all origins if configured with "*"
		if originSet["*"] || originSet[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400") // cache preflight for 24h

		// Handle preflight requests (browser sends OPTIONS before the real POST)
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}