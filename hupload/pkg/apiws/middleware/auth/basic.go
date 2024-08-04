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
type BasicAuthMiddleware struct {
	Authentication authentication.AuthenticationInterface
}

func (a BasicAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Collect authentication from request
		qUser, qPasswd, ok := r.BasicAuth()

		// If authentication has been sent, check credentials
		if ok {
			b, err := a.Authentication.AuthenticateUser(qUser, qPasswd)
			if err != nil {
				serveNextError(next, w, r, err)
				return
			}
			if !b {
				serveNextError(next, w, r, errors.New("authentication failed"))
				return
			} else {
				serveNextAuthenticated(qUser, next, w, r)
				return
			}
		}

		// Fail by default
		serveNextError(next, w, r, errors.New("no basic authentication provided"))
	})
}
