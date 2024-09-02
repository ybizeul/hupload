package auth

import (
	"errors"
	"net/http"

	"github.com/ybizeul/hupload/pkg/apiws/authentication"
)

var (
	ErrBasicAuthNoCredentials        = errors.New("no basic authentication provided")
	ErrBasicAuthAuthenticationFailed = errors.New("authentication failed")
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
type BasicAuthMiddleware struct {
	Authentication authentication.Authentication
}

func (a BasicAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.Authentication == nil {
			serveNextError(next, w, r, errors.New("no authentication backend"))
			return
		}

		// If authentication has been sent, check credentials

		a.Authentication.AuthenticateRequest(nil, r, func(ok bool, err error) {
			if err != nil {
				if errors.Is(err, authentication.ErrAuthenticationMissingCredentials) {
					serveNextError(next, w, r, ErrBasicAuthNoCredentials)
					return
				}
				serveNextError(next, w, r, err)
				return
			}
			if !ok {
				serveNextError(next, w, r, ErrBasicAuthAuthenticationFailed)
				return
			} else {
				qUser, _, _ := r.BasicAuth()
				serveNextAuthenticated(qUser, next, w, r)
				return
			}
		})
	})
}
