package auth

import (
	"errors"
	"net/http"

	"github.com/ybizeul/hupload/pkg/apiws/authentication"
)

// BasicAuthenticator uses a password file to authenticate users, like :
//   - username: admin
//     password: $2y$10$AJEytAoJfc4yQjUS8/cG6eXADlgK/Dt3AvdB0boPJ7EcHofewGQIK
//
// To has a password, you can use htpasswd command :
//
// ❯ htpasswd -bnBC 10 "" hupload
// :$2y$10$AJEytAoJfc4yQjUS8/cG6eXADlgK/Dt3AvdB0boPJ7EcHofewGQIK
//
// and remove the leading `:` from the hash
type OIDCAuthMiddleware struct {
	Authentication authentication.Authentication
}

func (a OIDCAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.Authentication == nil {
			serveNextError(next, w, r, errors.New("no authentication backend"))
			return
		}

		// If authentication has been sent, check credentials

		a.Authentication.AuthenticateRequest(w, r, func(ok bool, err error) {
			if err == authentication.ErrAuthenticationRedirect {
				return
			}
			if ok {
				// TODO
				serveNextAuthenticated("admin", next, w, r)
			} else {
				serveNextError(next, w, r, err)
			}
		})
	})
}