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
			ServeNextError(next, w, r, errors.New("no authentication backend"))
			return
		}

		// If authentication has been sent, check credentials

		err := a.Authentication.AuthenticateRequest(nil, r)

		if err != nil {
			if errors.Is(err, authentication.ErrAuthenticationMissingCredentials) {
				ServeNextAuthenticated("", next, w, r)
				return
			}
			if errors.Is(err, authentication.ErrAuthenticationBadCredentials) {
				ServeNextError(next, w, r, ErrBasicAuthAuthenticationFailed)
				return
			}
			ServeNextError(next, w, r, err)
			return
		}

		qUser, _, _ := r.BasicAuth()
		ServeNextAuthenticated(qUser, next, w, r)
	})
}
