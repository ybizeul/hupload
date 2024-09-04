package auth

import (
	"net/http"
)

type OpenAuthMiddleware struct {
}

func (a OpenAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeNextAuthenticated("", next, w, r)
	})
}
