package auth

import (
	"net/http"
	"strings"
)

// Middleware reads the "Authorization: Bearer <token>" header on every request.
// If a valid token is present, the user ID is attached to the request context.
// It never rejects a request outright — register/login must still work
// without a token. Protected resolvers check auth.UserIDFromContext themselves
// and return an Unauthorized error if it's missing.
func Middleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			tokenString := strings.TrimPrefix(header, "Bearer ")

			if tokenString != "" && tokenString != header {
				if userID, err := ParseToken(tokenString, jwtSecret); err == nil {
					r = r.WithContext(WithUserID(r.Context(), userID))
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
