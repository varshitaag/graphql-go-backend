package graph

// ─────────────────────────────────────────────────────────────────────────────
// resolver.go
//
// REST equivalent: handlers/book_handler.go
//
// KEY DIFFERENCE:
//   REST  → one Go function per HTTP verb+route (GetBook, CreateBook, etc.)
//           each function calls c.JSON(status, payload) to write the response.
//
//   GraphQL → one Go function per schema field in Query/Mutation.
//             Functions just RETURN a value (or error). The GraphQL runtime
//             handles serialisation, field selection, and HTTP status codes.
//             No c.JSON, no http.StatusNotFound — pure business logic only.
// ─────────────────────────────────────────────────────────────────────────────

import (
	"context"
	"fmt"
	"strconv"

	"github.com/yourname/gin-graphql-api/models"
	"github.com/yourname/gin-graphql-api/store"
)

// Resolver is the root dependency-injection struct.
//
// REST equivalent: BookHandler struct in handlers/book_handler.go
// Same idea — hold the store here, pass it to every resolver method.
type Resolver struct {
	Store *store.BookStore
}

// ── Query resolvers ──────────────────────────────────────────────────────────
// These handle the "Query" block from schema.graphql.

// Books resolves the `books` query field.
//
// REST equivalent: BookHandler.GetBooks — handled GET /api/v1/books
//
// DIFFERENCE: No gin.Context, no c.JSON call.
// We just return the slice. gqlgen marshals it to JSON automatically,
// and includes ONLY the fields the client asked for in the query document.
func (r *Resolver) Books(ctx context.Context) ([]*models.Book, error) {
	books := r.Store.GetAll()

	// Convert []Book → []*Book (gqlgen needs pointers for nullable types)
	result := make([]*models.Book, len(books))
	for i := range books {
		b := books[i]
		result[i] = &b
	}
	return result, nil
}

// Book resolves the `book(id: ID!)` query field.
//
// REST equivalent: BookHandler.GetBook — handled GET /api/v1/books/:id
//
// DIFFERENCE: In REST, :id came from c.Param("id") and a 404 response
// was sent manually with c.JSON(http.StatusNotFound, ...).
// Here, we return (nil, nil) if not found — GraphQL maps that to
// a null field in the response, and no error is thrown.
// For a hard error (like invalid id format) we return (nil, err).
func (r *Resolver) Book(ctx context.Context, id string) (*models.Book, error) {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id %q: must be an integer", id)
	}

	book, err := r.Store.GetByID(intID)
	if err != nil {
		// Not found → return null (nil, nil), not an error
		return nil, nil
	}
	return &book, nil
}

// ── Mutation resolvers ───────────────────────────────────────────────────────
// These handle the "Mutation" block from schema.graphql.

// CreateBook resolves the `createBook(input: BookInput!)` mutation.
//
// REST equivalent: BookHandler.CreateBook — handled POST /api/v1/books
//
// DIFFERENCE: In REST, the request body was bound via c.ShouldBindJSON(&input)
// and validation used struct `binding` tags. Here, gqlgen has already parsed
// and type-checked the arguments before this function is called — the `input`
// argument arrives as a ready-to-use Go struct. No manual binding needed.
func (r *Resolver) CreateBook(ctx context.Context, input models.BookInput) (*models.Book, error) {
	if err := validateBookInput(input); err != nil {
		return nil, err
	}
	book := r.Store.Create(input)
	return &book, nil
}

// UpdateBook resolves the `updateBook(id: ID!, input: BookInput!)` mutation.
//
// REST equivalent: BookHandler.UpdateBook — handled PUT /api/v1/books/:id
func (r *Resolver) UpdateBook(ctx context.Context, id string, input models.BookInput) (*models.Book, error) {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id %q", id)
	}

	if err := validateBookInput(input); err != nil {
		return nil, err
	}

	book, err := r.Store.Update(intID, input)
	if err != nil {
		return nil, fmt.Errorf("book %d not found", intID)
	}
	return &book, nil
}

// DeleteBook resolves the `deleteBook(id: ID!)` mutation.
//
// REST equivalent: BookHandler.DeleteBook — handled DELETE /api/v1/books/:id
//
// DIFFERENCE: REST returned 204 No Content (empty body on success).
// GraphQL mutations must return something — we return Boolean! (true = deleted).
func (r *Resolver) DeleteBook(ctx context.Context, id string) (bool, error) {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return false, fmt.Errorf("invalid id %q", id)
	}

	if err := r.Store.Delete(intID); err != nil {
		return false, fmt.Errorf("book %d not found", intID)
	}
	return true, nil
}

// ── validation helper ────────────────────────────────────────────────────────
// REST used struct binding tags for this. In GraphQL, non-null fields (String!)
// are enforced by the schema, but business rules (year range, price > 0)
// still need manual validation in the resolver.
func validateBookInput(input models.BookInput) error {
	if input.Year < 1450 || input.Year > 2100 {
		return fmt.Errorf("year must be between 1450 and 2100, got %d", input.Year)
	}
	if input.Price <= 0 {
		return fmt.Errorf("price must be greater than 0, got %.2f", input.Price)
	}
	return nil
}
