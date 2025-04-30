package middleware

import (
	"net/http"
	"regexp"
)

func ShareNameCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		share := r.PathValue("share")

		if share == "" {
			http.Error(w, "share name is required", http.StatusBadRequest)
			return
		}
		if !isShareNameSafe(share) {
			writeError(w, http.StatusBadRequest, "invalid share name")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isShareNameSafe(n string) bool {
	m := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(n)
	return m
}
