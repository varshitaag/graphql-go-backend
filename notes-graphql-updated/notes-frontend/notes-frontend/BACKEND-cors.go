package auth

import "net/http"

// CORS allows requests from your React dev server, which runs on a
// different origin (e.g. http://localhost:5173) than the Go server
// (http://localhost:8080). Browsers block cross-origin requests by
// default — this middleware adds the headers that opt back in.
//
// allowedOrigin should match your frontend's exact URL, e.g.
// "http://localhost:5173". Using "*" works for quick local testing but
// won't work once you add cookies/credentials, and shouldn't be used in
// production.
func CORS(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Browsers send a preflight OPTIONS request before the real one,
			// just to check these headers — it expects an empty 200 back,
			// not the actual GraphQL response.
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
