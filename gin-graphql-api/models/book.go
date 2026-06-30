package models

import "time"

// Book represents a book resource in our API.
// JSON tags control how fields are (de)serialized.
// `binding` tags are used by Gin for request validation.
type Book struct {
	ID        int       `json:"id"`
	Title     string    `json:"title" binding:"required"`
	Author    string    `json:"author" binding:"required"`
	Year      int       `json:"year" binding:"required,gte=1450,lte=2100"`
	Price     float64   `json:"price" binding:"required,gt=0"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BookInput is the shape we accept on create/update requests.
// Keeping a separate input struct (without ID/timestamps) is a common
// Go REST pattern — it stops clients from setting server-controlled fields.
type BookInput struct {
	Title  string  `json:"title" binding:"required"`
	Author string  `json:"author" binding:"required"`
	Year   int     `json:"year" binding:"required,gte=1450,lte=2100"`
	Price  float64 `json:"price" binding:"required,gt=0"`
}
