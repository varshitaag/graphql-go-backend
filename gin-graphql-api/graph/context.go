package graph

// context.go
//
// REST had gin.Context to carry per-request values (auth user, request ID).
// GraphQL resolvers receive a plain context.Context instead.
//
// This file defines typed keys and helpers so resolvers can access
// per-request data (e.g. the authenticated user, a logger, a trace ID)
// without importing gin or knowing anything about HTTP.
//
// Usage in middleware:
//   ctx = WithRequestID(ctx, "abc-123")
//
// Usage in a resolver:
//   id := RequestIDFromContext(ctx)

import "context"

// contextKey is an unexported type for context keys in this package.
// Using a named type prevents collisions with keys from other packages.
type contextKey string

const (
	contextKeyRequestID contextKey = "request_id"
	// Add more keys here as the app grows, e.g.:
	// contextKeyUserID   contextKey = "user_id"
	// contextKeyLogger   contextKey = "logger"
)

// WithRequestID stores a request ID in the context.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKeyRequestID, id)
}

// RequestIDFromContext retrieves the request ID, or "" if not set.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(contextKeyRequestID).(string); ok {
		return id
	}
	return ""
}

// ── example: how you'd add auth user ─────────────────────────────────────────
// Uncomment when you add authentication:
//
// type AuthUser struct {
//     ID    string
//     Email string
//     Role  string
// }
//
// func WithAuthUser(ctx context.Context, user *AuthUser) context.Context {
//     return context.WithValue(ctx, contextKeyUserID, user)
// }
//
// func AuthUserFromContext(ctx context.Context) (*AuthUser, bool) {
//     user, ok := ctx.Value(contextKeyUserID).(*AuthUser)
//     return user, ok
// }
//
// Then in a resolver:
// func (r *Resolver) CreateBook(ctx context.Context, input models.BookInput) (*models.Book, error) {
//     user, ok := AuthUserFromContext(ctx)
//     if !ok {
//         return nil, ValidationError("authentication required")
//     }
//     // use user.ID ...
// }
