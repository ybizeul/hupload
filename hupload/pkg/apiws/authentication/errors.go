package authentication

import "errors"

var ErrAuthenticationMissingUsersFile = errors.New("missing users file")
var ErrAuthenticationInvalidPath = errors.New("invalid path")
