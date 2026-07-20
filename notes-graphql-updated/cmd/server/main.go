package main

import (
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"notes-app/graph"
	"notes-app/internal/auth"
	"notes-app/internal/config"
	"notes-app/internal/database"
	"notes-app/internal/repository"
	"notes-app/internal/service"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()

	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	log.Printf("Google Client ID loaded: %q", cfg.GoogleClientID)

	// 2. Connect to Postgres
	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// 3. Wire up repositories -> services -> resolver
	noteRepo := repository.NewNoteRepository(db)
	noteService := service.NewNoteService(noteRepo)

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.GoogleClientID)

	resolver := graph.NewResolver(noteService, authService)

	// 4. Mount the GraphQL handler
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("Notes GraphQL Playground", "/query"))
	mux.Handle("/query", srv)

	// 5. Wrap everything in the JWT middleware. It attaches the user ID to
	// context when a valid "Authorization: Bearer <token>" header is present,
	// but never rejects a request outright — register/login must still work
	// without a token. Protected resolvers check for a user ID themselves.
	handlerWithAuth := auth.Middleware(cfg.JWTSecret)(mux)

	log.Printf("🚀 server ready at http://localhost:%s/", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, handlerWithAuth))
}
