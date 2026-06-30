package graph

// ─────────────────────────────────────────────────────────────────────────────
// handler.go
//
// REST equivalent: handlers/book_handler.go  (but very different role)
//
// In REST, each handler function mapped directly to one HTTP route.
// In GraphQL, there is ONE HTTP handler for ALL operations. It:
//   1. Receives every request on POST /graphql
//   2. Parses the GraphQL query/mutation document from the JSON body
//   3. Validates it against the schema
//   4. Calls the right resolver function(s)
//   5. Returns a JSON response shaped exactly like the query asked for
//
// We're building this "by hand" using the low-level gqlparser library
// so you can see what's happening under the hood. In production you'd
// typically use gqlgen's generated ExecutableSchema instead.
// ─────────────────────────────────────────────────────────────────────────────

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/yourname/gin-graphql-api/models"
)

// gqlRequest is the shape of every GraphQL HTTP request body.
//
// REST equivalent: there was no equivalent — each REST endpoint had
// its own distinct request format (/books POST vs /books/:id PUT).
//
// In GraphQL, every single operation arrives in this same envelope:
//   { "query": "...", "variables": {...}, "operationName": "..." }
type gqlRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operationName"`
}

// gqlResponse is the standard GraphQL response envelope.
// "data" holds the result; "errors" holds any errors.
// Both can be present simultaneously (partial success is valid in GraphQL).
type gqlResponse struct {
	Data   interface{}            `json:"data,omitempty"`
	Errors []*gqlerror.Error      `json:"errors,omitempty"`
}

// Handler is the single HTTP handler that serves all GraphQL operations.
// Mount it at POST /graphql in routes/router.go.
type Handler struct {
	resolver *Resolver
	schema   *ast.Schema
}

// NewHandler builds the GraphQL handler, parsing the schema once at startup.
func NewHandler(r *Resolver) (*Handler, error) {
	// Parse the .graphql schema file into an AST.
	// REST had no equivalent step — routes were just registered in the router.
	schemaDoc, gqlErr := gqlparser.LoadSchema(&ast.Source{
		Name:  "schema.graphql",
		Input: schemaSDL, // defined at bottom of this file
	})
	if gqlErr != nil {
		return nil, fmt.Errorf("failed to parse schema: %v", gqlErr)
	}

	return &Handler{resolver: r, schema: schemaDoc}, nil
}

// ServeHTTP is the single entry point for ALL GraphQL traffic.
//
// REST equivalent: Gin called a different handler function per route.
//   GET  /books     → handler.GetBooks(c)
//   POST /books     → handler.CreateBook(c)
//   etc.
//
// GraphQL: every request hits THIS function. We inspect the parsed
// query/mutation document to decide which resolver to call.
func (h *Handler) ServeHTTP(c *gin.Context) {
	// 1. Parse the request body
	var req gqlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gqlResponse{
			Errors: gqlerror.List{{Message: "invalid request body: " + err.Error()}},
		})
		return
	}

	// 2. Parse + validate the GraphQL document against the schema
	//    (REST did nothing like this; routing was pure URL matching)
	doc, listErr := gqlparser.LoadQuery(h.schema, req.Query)
	if listErr != nil {
		c.JSON(http.StatusOK, gqlResponse{Errors: listErr})
		return
	}

	// 3. Execute — walk the operation and call the correct resolver(s)
	ctx := c.Request.Context()
	data, execErrs := h.execute(ctx, doc, req.Variables)

	// 4. Return the standard GraphQL envelope
	//    Status is always 200 OK in GraphQL (errors go inside the body)
	resp := gqlResponse{Data: data}
	if len(execErrs) > 0 {
		resp.Errors = execErrs
	}
	c.JSON(http.StatusOK, resp)
}

// execute walks the GraphQL document and dispatches to resolvers.
func (h *Handler) execute(
	ctx context.Context,
	doc *ast.QueryDocument,
	vars map[string]interface{},
) (interface{}, gqlerror.List) {

	if len(doc.Operations) == 0 {
		return nil, gqlerror.List{{Message: "no operations in document"}}
	}

	op := doc.Operations[0]
	result := map[string]interface{}{}

	for _, sel := range op.SelectionSet {
		field, ok := sel.(*ast.Field)
		if !ok {
			continue
		}

		var value interface{}
		var err error

		switch op.Operation {
		case ast.Query:
			value, err = h.resolveQuery(ctx, field, vars)
		case ast.Mutation:
			value, err = h.resolveMutation(ctx, field, vars)
		default:
			return nil, gqlerror.List{{Message: "subscriptions not supported"}}
		}

		if err != nil {
			return nil, gqlerror.List{{Message: err.Error()}}
		}
		result[field.Alias] = value
	}

	return result, nil
}

// resolveQuery dispatches query fields (reads) to the right resolver method.
//
// REST equivalent: the Gin router matched URL+verb to a function.
// Here we match the GraphQL field name to a resolver function.
func (h *Handler) resolveQuery(ctx context.Context, field *ast.Field, vars map[string]interface{}) (interface{}, error) {
	switch field.Name {
	case "books":
		return h.resolver.Books(ctx)

	case "book":
		id := argString(field, "id", vars)
		return h.resolver.Book(ctx, id)

	default:
		return nil, fmt.Errorf("unknown query field: %s", field.Name)
	}
}

// resolveMutation dispatches mutation fields (writes) to the right resolver.
func (h *Handler) resolveMutation(ctx context.Context, field *ast.Field, vars map[string]interface{}) (interface{}, error) {
	switch field.Name {
	case "createBook":
		input := argInput(field, vars)
		return h.resolver.CreateBook(ctx, input)

	case "updateBook":
		id := argString(field, "id", vars)
		input := argInput(field, vars)
		return h.resolver.UpdateBook(ctx, id, input)

	case "deleteBook":
		id := argString(field, "id", vars)
		return h.resolver.DeleteBook(ctx, id)

	default:
		return nil, fmt.Errorf("unknown mutation field: %s", field.Name)
	}
}

// ── argument helpers ─────────────────────────────────────────────────────────

// argString extracts a scalar string argument from a field.
// Variables take precedence over inline literal values.
func argString(field *ast.Field, name string, vars map[string]interface{}) string {
	for _, arg := range field.Arguments {
		if arg.Name != name {
			continue
		}
		if arg.Value.Kind == ast.Variable {
			if v, ok := vars[arg.Value.Raw]; ok {
				return fmt.Sprintf("%v", v)
			}
		}
		return arg.Value.Raw
	}
	return ""
}

// argInput extracts a BookInput argument, supporting both inline objects
// and $variable references.
func argInput(field *ast.Field, vars map[string]interface{}) models.BookInput {
	var input models.BookInput
	for _, arg := range field.Arguments {
		if arg.Name != "input" {
			continue
		}

		// Variable reference: { input: $myInput }
		if arg.Value.Kind == ast.Variable {
			if raw, ok := vars[arg.Value.Raw]; ok {
				b, _ := json.Marshal(raw)
				_ = json.Unmarshal(b, &input)
			}
			return input
		}

		// Inline object: { input: { title: "...", ... } }
		for _, child := range arg.Value.Children {
			switch child.Name {
			case "title":
				input.Title = child.Value.Raw
			case "author":
				input.Author = child.Value.Raw
			case "year":
				fmt.Sscanf(child.Value.Raw, "%d", &input.Year)
			case "price":
				fmt.Sscanf(child.Value.Raw, "%f", &input.Price)
			}
		}
	}
	return input
}

// schemaSDL is the raw GraphQL schema embedded as a Go string.
// In production you'd use go:embed to load schema.graphql from disk.
const schemaSDL = `
type Book {
    id:        ID!
    title:     String!
    author:    String!
    year:      Int!
    price:     Float!
    createdAt: String!
    updatedAt: String!
}

input BookInput {
    title:  String!
    author: String!
    year:   Int!
    price:  Float!
}

type Query {
    books:         [Book!]!
    book(id: ID!): Book
}

type Mutation {
    createBook(input: BookInput!): Book!
    updateBook(id: ID!, input: BookInput!): Book
    deleteBook(id: ID!): Boolean!
}
`