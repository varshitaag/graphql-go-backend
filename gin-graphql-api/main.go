package main

// ─────────────────────────────────────────────────────────────────────────────
// main.go
//
// IDENTICAL to the REST version in structure.
// The complexity difference between REST and GraphQL lives entirely
// in graph/ and routes/ — main.go doesn't care.
// ─────────────────────────────────────────────────────────────────────────────

import (
	"log"

	"github.com/yourname/gin-graphql-api/routes"
)

func main() {
	router := routes.SetupRouter()

	log.Println("Server starting on http://localhost:8080")
	log.Println("GraphiQL explorer: http://localhost:8080/api/v1/graphql")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
