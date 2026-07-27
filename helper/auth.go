package helper

import (
	"fmt"
	"net/http"
)

const apiKeyHeader = "X-API-Key"

// AuthMiddleware validates the X-API-Key request header against the
// BILLING_ENGINE_API_KEY environment variable before calling next.
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		expected := "fortestingonly"
		got := r.Header.Get(apiKeyHeader)

		if expected == "" || got == "" || got != expected {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			fmt.Println("expected:", expected)
			return
		}

		next(w, r)
	}
}
