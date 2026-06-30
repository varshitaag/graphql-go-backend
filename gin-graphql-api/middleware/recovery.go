package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery returns a middleware that catches panics and returns them
// as a proper GraphQL error response rather than a 500 HTML page.
//
// REST recovery: return { "error": "internal server error" } with HTTP 500
//
// GraphQL recovery: MUST return HTTP 200 with the error inside the body
// because GraphQL clients parse the JSON body, not the HTTP status code.
// A 500 with an HTML body breaks every GraphQL client library.
//
//	{ "errors": [{ "message": "internal server error" }] }
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] recovered: %v\n%s", r, debug.Stack())

				// GraphQL error envelope — HTTP 200 but errors in the body
				c.JSON(http.StatusOK, gin.H{
					"data": nil,
					"errors": []gin.H{
						{"message": "internal server error"},
					},
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
