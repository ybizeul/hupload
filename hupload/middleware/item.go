package middleware

import (
	"net/http"
	"strings"
)

func ItemNameCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		item := r.PathValue("item")

		if item == "" {
			writeError(w, http.StatusBadRequest, "item name is required")
			return
		}
		if !isItemNameSafe(item) {
			writeError(w, http.StatusBadRequest, "invalid item name")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isItemNameSafe(n string) bool {
	return !strings.HasPrefix(n, ".")
}
