package graph

// errors.go
//
// In REST, error responses were ad hoc — each handler called
//   c.JSON(http.StatusXXX, gin.H{"error": "message"})
// with whatever shape felt right at the time.
//
// In GraphQL, errors always go inside the "errors" array and the
// spec defines fields every error object should have:
//   { "message": "...", "locations": [...], "path": [...], "extensions": {...} }
//
// This file centralises error construction so every error in the API
// has a consistent shape and a machine-readable "code" in extensions.
// Frontend apps can switch on the code to show the right UI.

import "fmt"

// GraphQLError is a structured error with a code for the client to act on.
type GraphQLError struct {
	message string
	code    string
}

func (e *GraphQLError) Error() string {
	return e.message
}

// Code returns the machine-readable error code string.
func (e *GraphQLError) Code() string {
	return e.code
}

// ── constructors ──────────────────────────────────────────────────────────────

// NotFound returns a standard "not found" error.
// REST equivalent: c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
func NotFound(resource string, id interface{}) *GraphQLError {
	return &GraphQLError{
		message: fmt.Sprintf("%s with id %v not found", resource, id),
		code:    "NOT_FOUND",
	}
}

// ValidationError returns a validation failure error.
// REST equivalent: c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
func ValidationError(msg string) *GraphQLError {
	return &GraphQLError{
		message: msg,
		code:    "VALIDATION_ERROR",
	}
}

// InvalidArgument returns an error for malformed input arguments.
func InvalidArgument(field, reason string) *GraphQLError {
	return &GraphQLError{
		message: fmt.Sprintf("invalid argument %q: %s", field, reason),
		code:    "BAD_USER_INPUT",
	}
}

// Internal returns a generic internal error (hides implementation details).
// REST equivalent: c.JSON(http.StatusInternalServerError, gin.H{"error": "..."})
func Internal(msg string) *GraphQLError {
	return &GraphQLError{
		message: "internal error: " + msg,
		code:    "INTERNAL_ERROR",
	}
}
