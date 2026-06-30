package store

import (
	"errors"
	"sync"
	"time"

	"github.com/yourname/gin-graphql-api/models"
)

// ErrNotFound is returned when a book with the given ID doesn't exist.
var ErrNotFound = errors.New("book not found")

// BookStore is a simple thread-safe in-memory data store.
// In a real app you'd replace this with a database (Postgres, MySQL, etc.)
// — the handlers only depend on this interface-like struct, so swapping
// the storage layer later doesn't require touching handler code much.
type BookStore struct {
	mu     sync.RWMutex
	books  map[int]models.Book
	nextID int
}

// NewBookStore creates an empty store pre-loaded with a couple of sample books.
func NewBookStore() *BookStore {
	s := &BookStore{
		books:  make(map[int]models.Book),
		nextID: 1,
	}

	// seed data so the API is testable immediately
	s.Create(models.BookInput{Title: "The Go Programming Language", Author: "Donovan & Kernighan", Year: 2015, Price: 39.99})
	s.Create(models.BookInput{Title: "Clean Code", Author: "Robert C. Martin", Year: 2008, Price: 34.50})

	return s
}

// GetAll returns every book in the store.
func (s *BookStore) GetAll() []models.Book {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.Book, 0, len(s.books))
	for _, b := range s.books {
		result = append(result, b)
	}
	return result
}

// GetByID returns a single book, or ErrNotFound.
func (s *BookStore) GetByID(id int) (models.Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, ok := s.books[id]
	if !ok {
		return models.Book{}, ErrNotFound
	}
	return b, nil
}

// Create adds a new book and returns it with its assigned ID.
func (s *BookStore) Create(input models.BookInput) models.Book {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	book := models.Book{
		ID:        s.nextID,
		Title:     input.Title,
		Author:    input.Author,
		Year:      input.Year,
		Price:     input.Price,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.books[book.ID] = book
	s.nextID++
	return book
}

// Update replaces an existing book's fields. Returns ErrNotFound if missing.
func (s *BookStore) Update(id int, input models.BookInput) (models.Book, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.books[id]
	if !ok {
		return models.Book{}, ErrNotFound
	}

	existing.Title = input.Title
	existing.Author = input.Author
	existing.Year = input.Year
	existing.Price = input.Price
	existing.UpdatedAt = time.Now()

	s.books[id] = existing
	return existing, nil
}

// Delete removes a book by ID. Returns ErrNotFound if missing.
func (s *BookStore) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.books[id]; !ok {
		return ErrNotFound
	}
	delete(s.books, id)
	return nil
}