package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// GraphQLLogger is a middleware that logs GraphQL requests with their
// operation name and type — much more useful than logging "POST /graphql"
// for every request, which tells you nothing about what was actually asked.
//
// REST loggers just log the method + URL:
//   GET  /api/v1/books       → immediately meaningful
//   POST /api/v1/books       → you know a book was created
//
// GraphQL default logs would show:
//   POST /graphql            → useless — every operation looks the same
//
// Our logger extracts the operation from the body:
//   [GraphQL] query "books" in 1.2ms
//   [GraphQL] mutation "createBook" in 2.4ms
func GraphQLLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Read the body (we need it for logging, but Gin will also need to read it)
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		// Restore the body so the actual handler can still read it
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Extract operation name from the GraphQL request body
		operationName, operationType := extractOperation(bodyBytes)

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		log.Printf("[GraphQL] %s %q | status %d | %v",
			operationType, operationName, status, duration)
	}
}

// extractOperation peeks at the request body to pull out the
// operation type (query/mutation) and operation name.
func extractOperation(body []byte) (name, opType string) {
	if len(body) == 0 {
		return "unknown", "unknown"
	}

	var req struct {
		Query         string `json:"query"`
		OperationName string `json:"operationName"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return "unknown", "unknown"
	}

	// Use the explicit operationName if provided
	if req.OperationName != "" {
		name = req.OperationName
	} else {
		// Fall back to first word after "query" or "mutation" in the query string
		name = extractFirstFieldName(req.Query)
	}

	// Detect operation type from the query string
	opType = "query"
	for i := 0; i < len(req.Query)-8; i++ {
		if req.Query[i] == 'm' && req.Query[i:i+8] == "mutation" {
			opType = "mutation"
			break
		}
	}

	return name, opType
}

// extractFirstFieldName naively extracts the first field name from a query
// document, e.g. "{ books { id title } }" → "books"
func extractFirstFieldName(query string) string {
	inBrace := false
	start := -1
	for i, ch := range query {
		if ch == '{' {
			inBrace = true
			continue
		}
		if inBrace && ch != ' ' && ch != '\n' && ch != '\t' && ch != '(' {
			if start == -1 {
				start = i
			}
		}
		if start != -1 && (ch == ' ' || ch == '\n' || ch == '(' || ch == '{') {
			return query[start:i]
		}
	}
	return "unknown"
}