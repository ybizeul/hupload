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

	err := a.AuthenticateRequest(nil, r)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
