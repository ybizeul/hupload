package authentication

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
)

// AuthenticationDefault is the default authentication when none has been found
// in configuration. Username is `admin` and password is a random 7 characters
type AuthenticationDefault struct {
	Password string
}

func NewAuthenticationDefault() *AuthenticationDefault {
	c := generateCode(7)
	slog.Info(fmt.Sprintf("Starting with default authentication backend. username: admin, password: %s", c))
	r := &AuthenticationDefault{
		Password: c,
	}

	return r
}

func (a *AuthenticationDefault) AuthenticateRequest(w http.ResponseWriter, r *http.Request, cb func(ok bool, err error)) {
	username, password, ok := r.BasicAuth()
	if !ok {
		cb(false, nil)
		return
	}
	if username == "admin" && password == a.Password {
		cb(true, nil)
		return
	}
	cb(false, nil)
}

func (o *AuthenticationDefault) CallbackFunc() (func(w http.ResponseWriter, r *http.Request), bool) {
	return nil, false
}

func generateCode(l int) string {
	code := ""

	for i := 0; i < l; i++ {
		c := rand.IntN(52)
		if c > 25 {
			c += 6
		}
		code += string(rune(c + 65))
	}
	return code
}
