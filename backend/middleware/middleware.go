package middleware

import (
	"log"
	"net/http"
)

// CORS Middleware for handling cross origin requests
// This is needed because the back-end and front-end are on different ports
func CORSMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("Executing middleware")

			// Define allowed origins for use by cors middleware
			allowedOrigins := []string{
				"http://localhost:3000",
			}

			origin := r.Header.Get("Origin")

			// Check if the origin is in the allowed list
			for _, o := range allowedOrigins {
				if o == origin {
					log.Println("Allowed Origin:", origin)

					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Credentials", "true") // Enable because using cookies and session-based auth
					break
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token")

			// Handle Preflight Requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
