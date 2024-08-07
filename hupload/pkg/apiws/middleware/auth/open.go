package auth

import (
	"net/http"

	"github.com/ybizeul/hupload/pkg/apiws/authentication"
)

type OpenAuthMiddleware struct {
	Authentication authentication.Authentication
}

func (a OpenAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveNextAuthenticated("", next, w, r)
	})
}
