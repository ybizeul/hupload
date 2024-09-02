package authentication

import (
	"net/http"
	"regexp"
	"testing"
)

func TestDefaultAuthentication(t *testing.T) {
	a := NewAuthenticationDefault()

	p := a.Password

	m := regexp.MustCompile(`^[A-Za-z]{7}$`).MatchString(p)
	if !m {
		t.Errorf("Expected password to be 7 characters long, got %s", p)
	}

	r, _ := http.NewRequest("GET", "http://localhost:8080", nil)
	r.SetBasicAuth("admin", p)

	a.AuthenticateRequest(nil, r, func(ok bool, err error) {
		if !ok {
			t.Errorf("Expected true, got false")
		}

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}
