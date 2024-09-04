package authentication

import (
	"net/http"
)

type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type AuthStatusKeyType string

var AuthStatusKey AuthStatusKeyType = "AuthStatus"

type AuthStatus struct {
	Authenticated bool
	User          string
	Error         error
}

// AuthenticationInterface must be implemented by the authentication backend
type Authentication interface {
	AuthenticateRequest(w http.ResponseWriter, r *http.Request) error
	CallbackFunc(http.Handler) (func(w http.ResponseWriter, r *http.Request), bool)
	ShowLoginForm() bool
	LoginURL() string
}
