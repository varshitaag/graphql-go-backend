#!/bin/bash
# GraphQL API smoke tests
# Start the server first: go run main.go
# Then: chmod +x test_graphql.sh && ./test_graphql.sh

BASE="http://localhost:8080/api/v1/graphql"

# ─────────────────────────────────────────────────────────────
# CRITICAL DIFFERENCE FROM REST
#
# REST:  curl -X GET  /api/v1/books
#        curl -X GET  /api/v1/books/1
#        curl -X POST /api/v1/books    -d '{...}'
#        curl -X PUT  /api/v1/books/1  -d '{...}'
#        curl -X DELETE /api/v1/books/1
#
# GraphQL: ALWAYS curl -X POST /api/v1/graphql -d '{"query": "..."}'
#          The operation type (query vs mutation) is in the body, not the URL.
#          The client decides WHICH FIELDS to get back.
# ─────────────────────────────────────────────────────────────

echo "=== QUERY: Get all books (only title + author — client picks fields) ==="
curl -s -X POST "$BASE" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "{ books { id title author } }"
  }'
# Notice: no year, price, createdAt in response — we didn't ask for them.
# REST would always return the full object. This is GraphQL's key advantage.
echo -e "\n"

echo "=== QUERY: Get all books (full fields this time) ==="
curl -s -X POST "$BASE" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "{ books { id title author year price createdAt updatedAt } }"
  }'
echo -e "\n"

echo "=== QUERY: Get a single book by ID ==="
curl -s -X POST "$BASE" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "{ book(id: \"1\") { id title author year price } }"
  }'
echo -e "\n"

echo "=== QUERY: Non-existent book (returns null, not 404) ==="
curl -s -X POST "$BASE" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "{ book(id: \"9999\") { id title } }"
  }'
# Response: { "data": { "book": null } }  — no HTTP error, null in the data
echo -e "\n"

echo "=== MUTATION: Create a new book ==="
curl -s -X POST "$BASE" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { createBook(input: { title: \"The Pragmatic Programmer\", author: \"Hunt & Thomas\", year: 1999, price: 42.50 }) { id title author year price } }"
  }'
echo -e "\n"

echo "=== MUTATION: Create with variables (preferred over inline) ==="
# Using variables separates the query document from the data — cleaner and
# avoids injection issues. There is no REST equivalent of this pattern.
curl -s -X POST "$BASE" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation CreateBook($input: BookInput!) { createBook(input: $input) { id title price } }",
    "variables": {
      "input": {
        "title": "Refactoring",
        "author": "Martin Fowler",
        "year": 2018,
        "price": 47.99
      }
    },
    "operationName": "CreateBook"
  }'
echo -e "\n"

echo "=== MUTATION: Update a book ==="
curl -s -X POST "$BASE" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { updateBook(id: \"1\", input: { title: \"The Go Programming Language (2nd Ed)\", author: \"Donovan & Kernighan\", year: 2024, price: 44.99 }) { id title year price updatedAt } }"
  }'
echo -e "\n"

echo "=== MUTATION: Delete a book ==="
curl -s -X POST "$BASE" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { deleteBook(id: \"1\") }"
  }'
# Returns: { "data": { "deleteBook": true } }
echo -e "\n"

echo "=== MUTATION: Validation error (price = 0) ==="
curl -s -X POST "$BASE" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { createBook(input: { title: \"Free Book\", author: \"Author\", year: 2020, price: 0 }) { id } }"
  }'
# Returns: { "errors": [{ "message": "price must be greater than 0" }] }
# Note: HTTP status is still 200 — errors are in the body, not the status code.
echo -e "\n"

echo "=== MULTI-OPERATION: Fetch two specific books in one request ==="
# This is IMPOSSIBLE in REST without two separate HTTP calls.
# GraphQL lets you alias multiple operations into a single round-trip.
curl -s -X POST "$BASE" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "{ firstBook: book(id: \"1\") { title } secondBook: book(id: \"2\") { title } }"
  }'
echo -e "\n"