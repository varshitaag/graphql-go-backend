package routes

// ─────────────────────────────────────────────────────────────────────────────
// router.go
//
// REST equivalent: routes/router.go (same file, very different content)
//
// KEY DIFFERENCE — compare side by side:
//
//   REST router registered FIVE route handlers:
//     books.GET("",      bookHandler.GetBooks)
//     books.GET("/:id",  bookHandler.GetBook)
//     books.POST("",     bookHandler.CreateBook)
//     books.PUT("/:id",  bookHandler.UpdateBook)
//     books.DELETE("/:id", bookHandler.DeleteBook)
//
//   GraphQL router registers ONE route:
//     api.POST("/graphql", gqlHandler.ServeHTTP)
//
//   That single route handles ALL operations. The "routing" from
//   operation → resolver now happens inside the GraphQL engine,
//   not at the HTTP layer.
// ─────────────────────────────────────────────────────────────────────────────

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourname/gin-graphql-api/graph"
	"github.com/yourname/gin-graphql-api/store"
)

// SetupRouter builds and returns the configured Gin engine.
func SetupRouter() *gin.Engine {
	router := gin.Default()

	// Wire up dependencies (same pattern as REST)
	bookStore := store.NewBookStore()
	resolver := &graph.Resolver{Store: bookStore}

	// Build the GraphQL handler (parses & validates the schema at startup)
	gqlHandler, err := graph.NewHandler(resolver)
	if err != nil {
		log.Fatalf("failed to build GraphQL handler: %v", err)
	}

	// Health check (unchanged from REST)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api/v1")
	{
		// ONE endpoint replaces all five REST routes.
		// All queries and mutations go through here.
		api.POST("/graphql", gqlHandler.ServeHTTP)

		// Optional: serve GraphiQL (in-browser query explorer) on GET.
		// This is the GraphQL equivalent of Postman for REST.
		api.GET("/graphql", func(c *gin.Context) {
			c.Data(http.StatusOK, "text/html; charset=utf-8", graphiqlHTML)
		})
	}

	return router
}

// graphiqlHTML is a minimal GraphiQL explorer page.
// Visit http://localhost:8080/api/v1/graphql in your browser to use it.
// There is no REST equivalent — REST APIs are tested with curl/Postman.
var graphiqlHTML = []byte(`<!DOCTYPE html>
<html>
<head>
  <title>GraphiQL — Books API</title>
  <link rel="stylesheet" href="https://unpkg.com/graphiql/graphiql.min.css"/>
</head>
<body style="margin:0">
  <div id="graphiql" style="height:100vh"></div>
  <script src="https://unpkg.com/react/umd/react.development.js"></script>
  <script src="https://unpkg.com/react-dom/umd/react-dom.development.js"></script>
  <script src="https://unpkg.com/graphiql/graphiql.min.js"></script>
  <script>
    const fetcher = GraphiQL.createFetcher({ url: '/api/v1/graphql' });
    ReactDOM.render(React.createElement(GraphiQL, { fetcher }), document.getElementById('graphiql'));
  </script>
</body>
</html>`)