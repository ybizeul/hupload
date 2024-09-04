package authentication

import "errors"

var ErrAuthenticationMissingUsersFile = errors.New("missing users file")
var ErrAuthenticationInvalidPath = errors.New("invalid path")

var ErrAuthenticationMissingCredentials = errors.New("no credentials provided in request")
var ErrAuthenticationBadCredentials = errors.New("invalid username or password")

var ErrAuthenticationRedirect = errors.New("redirect to authenticate")
