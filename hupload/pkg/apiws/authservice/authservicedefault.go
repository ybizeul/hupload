package authservice

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
)

type AuthBackendDefault struct {
	Password string
}

func NewAuthBackendDefault() *AuthBackendDefault {
	c := generateCode(7)
	slog.Info(fmt.Sprintf("Starting with default authentication service. username: admin, password: %s", c))
	r := &AuthBackendDefault{
		Password: c,
	}

	return r
}
func (a *AuthBackendDefault) AuthenticateUser(username, password string) (bool, error) {
	if username == "admin" && password == a.Password {
		return true, nil
	}
	return false, nil
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
