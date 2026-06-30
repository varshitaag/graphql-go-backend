package graph

// scalar.go
//
// GraphQL has 5 built-in scalar types: Int, Float, String, Boolean, ID.
// Any other data type you want to transport (dates, UUIDs, JSON blobs,
// currency amounts) must be defined as a custom scalar.
//
// This file is a common addition to any real GraphQL project.
// REST had no equivalent — you just serialised whatever you wanted into JSON.
//
// Our Book already uses String for createdAt/updatedAt (simplest approach).
// This file shows how you'd add a proper DateTime scalar if you wanted
// typed date handling on the client side.

import (
	"fmt"
	"time"
)

// DateTime is a custom scalar that serialises time.Time as an ISO-8601 string.
// To activate it:
//   1. Add `scalar DateTime` to schema.graphql
//   2. Change createdAt/updatedAt field types from String! to DateTime!
//   3. Use MarshalDateTime / UnmarshalDateTime in the resolver

// MarshalDateTime converts a Go time.Time to the ISO-8601 string
// that gets sent to the client in the JSON response.
func MarshalDateTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// UnmarshalDateTime converts an ISO-8601 string (received from a client
// as a variable) back into a Go time.Time.
func UnmarshalDateTime(v interface{}) (time.Time, error) {
	str, ok := v.(string)
	if !ok {
		return time.Time{}, fmt.Errorf("DateTime must be a string, got %T", v)
	}
	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid DateTime format %q, expected RFC3339", str)
	}
	return t, nil
}

// ── other common custom scalars you might add ─────────────────────────────────
//
// scalar UUID    — for string UUIDs, validated on input
// scalar JSON    — for arbitrary JSON blobs (map[string]interface{})
// scalar Upload  — for file uploads via multipart form
// scalar Long    — for int64 (GraphQL Int is only 32-bit)
//
// Each needs:
//   - `scalar XYZ` in schema.graphql
//   - A Marshal + Unmarshal function here
//   - Registration in gqlgen.yml (if using code generation)