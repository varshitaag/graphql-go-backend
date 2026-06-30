package graph_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yourname/gin-graphql-api/graph"
	"github.com/yourname/gin-graphql-api/models"
	"github.com/yourname/gin-graphql-api/routes"
	"github.com/yourname/gin-graphql-api/store"
)

// ── resolver unit tests ───────────────────────────────────────────────────────
// These test the resolver functions directly, bypassing HTTP entirely.
// Much cleaner than testing via HTTP: no JSON parsing, no curl calls.

func TestResolver_Books(t *testing.T) {
	s := store.NewBookStore()
	r := &graph.Resolver{Store: s}

	books, err := r.Books(context.Background())
	if err != nil {
		t.Fatalf("Books() returned error: %v", err)
	}
	if len(books) == 0 {
		t.Fatal("Books() returned empty slice, expected seeded data")
	}
}

func TestResolver_CreateBook(t *testing.T) {
	s := store.NewBookStore()
	r := &graph.Resolver{Store: s}

	input := models.BookInput{
		Title:  "Clean Architecture",
		Author: "Robert C. Martin",
		Year:   2017,
		Price:  39.99,
	}

	book, err := r.CreateBook(context.Background(), input)
	if err != nil {
		t.Fatalf("CreateBook() error: %v", err)
	}
	if book.Title != input.Title {
		t.Errorf("expected title %q, got %q", input.Title, book.Title)
	}
	if book.ID == 0 {
		t.Error("expected non-zero ID")
	}
}

func TestResolver_CreateBook_InvalidYear(t *testing.T) {
	s := store.NewBookStore()
	r := &graph.Resolver{Store: s}

	_, err := r.CreateBook(context.Background(), models.BookInput{
		Title: "Bad Book", Author: "A", Year: 1200, Price: 10,
	})
	if err == nil {
		t.Error("expected validation error for year < 1450, got nil")
	}
}

func TestResolver_CreateBook_InvalidPrice(t *testing.T) {
	s := store.NewBookStore()
	r := &graph.Resolver{Store: s}

	_, err := r.CreateBook(context.Background(), models.BookInput{
		Title: "Free Book", Author: "A", Year: 2020, Price: 0,
	})
	if err == nil {
		t.Error("expected validation error for price = 0, got nil")
	}
}

func TestResolver_Book_NotFound(t *testing.T) {
	s := store.NewBookStore()
	r := &graph.Resolver{Store: s}

	book, err := r.Book(context.Background(), "9999")
	if err != nil {
		t.Fatalf("Book() on missing ID should return nil error, got %v", err)
	}
	if book != nil {
		t.Errorf("expected nil book for missing ID, got %+v", book)
	}
}

func TestResolver_DeleteBook(t *testing.T) {
	s := store.NewBookStore()
	r := &graph.Resolver{Store: s}

	// seeded book ID 1 should exist
	ok, err := r.DeleteBook(context.Background(), "1")
	if err != nil {
		t.Fatalf("DeleteBook() error: %v", err)
	}
	if !ok {
		t.Error("expected DeleteBook to return true")
	}

	// second delete should fail
	_, err = r.DeleteBook(context.Background(), "1")
	if err == nil {
		t.Error("expected error on second delete, got nil")
	}
}

// ── HTTP integration tests ────────────────────────────────────────────────────
// These fire real GraphQL queries over HTTP via httptest.

func setupRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	return routes.SetupRouter()
}

func gqlPost(t *testing.T, router *gin.Engine, query string, variables map[string]interface{}) map[string]interface{} {
	t.Helper()
	body, _ := json.Marshal(map[string]interface{}{
		"query":     query,
		"variables": variables,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 (GraphQL always returns 200), got %d", w.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v — body: %s", err, w.Body.String())
	}
	return result
}

func TestHTTP_QueryBooks(t *testing.T) {
	router := setupRouter(t)
	result := gqlPost(t, router, `{ books { id title author } }`, nil)

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("no data in response: %v", result)
	}
	books, ok := data["books"].([]interface{})
	if !ok || len(books) == 0 {
		t.Fatal("expected non-empty books list")
	}
}

func TestHTTP_MutationCreateBook(t *testing.T) {
	router := setupRouter(t)
	query := `mutation CreateBook($input: BookInput!) {
		createBook(input: $input) { id title price }
	}`
	vars := map[string]interface{}{
		"input": map[string]interface{}{
			"title":  "Test Book",
			"author": "Tester",
			"year":   2020,
			"price":  29.99,
		},
	}

	result := gqlPost(t, router, query, vars)

	if errs, ok := result["errors"]; ok {
		t.Fatalf("unexpected errors: %v", errs)
	}
	data := result["data"].(map[string]interface{})
	book := data["createBook"].(map[string]interface{})
	if book["title"] != "Test Book" {
		t.Errorf("expected title 'Test Book', got %v", book["title"])
	}
}

func TestHTTP_ValidationError_ReturnsHTTP200(t *testing.T) {
	router := setupRouter(t)

	// Invalid price — should get a 200 with errors in body, NOT a 400
	result := gqlPost(t, router, `mutation {
		createBook(input: { title: "X", author: "Y", year: 2020, price: 0 }) { id }
	}`, nil)

	if _, hasErrors := result["errors"]; !hasErrors {
		t.Fatal("expected errors array in response body for validation failure")
	}
	// HTTP status is already verified to be 200 inside gqlPost
}
