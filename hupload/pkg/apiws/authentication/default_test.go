package authentication

import (
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

	b, err := a.AuthenticateUser("admin", p)

	if !b {
		t.Errorf("Expected true, got false")
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
