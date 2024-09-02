package authentication

import "net/http"

type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// AuthenticationInterface must be implemented by the authentication backend
type Authentication interface {
	AuthenticateRequest(w http.ResponseWriter, r *http.Request, cb func(bool, error))
	CallbackFunc() (func(w http.ResponseWriter, r *http.Request), bool)
}
