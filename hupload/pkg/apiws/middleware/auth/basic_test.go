package auth

import (
	"errors"
	"net/http"
	"testing"

	"github.com/ybizeul/hupload/pkg/apiws/authentication"
)

func TestBasicAuth(t *testing.T) {
	c := authentication.FileAuthenticationConfig{
		Path: "tests/users.yml",
	}

	a, err := authentication.NewAuthenticationFile(c)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	m := BasicAuthMiddleware{
		Authentication: a,
	}

	fn1 := func(w http.ResponseWriter, r *http.Request) {
		c := r.Context().Value(AuthError)
		if c != nil {
			t.Errorf("Expected nil, got %v", c.(error))
		}
		c = r.Context().Value(AuthStatus)
		if c != AuthStatusSuccess {
			t.Errorf("Expected AuthStatusSuccess, got %v", c)
		}
		u := r.Context().Value(AuthUser)
		if u != "admin" {
			t.Errorf("Expected admin, got %v", u)
		}
	}

	h1 := m.Middleware(http.HandlerFunc(fn1))

	req, _ := http.NewRequest("GET", "https://example.com/", nil)
	req.SetBasicAuth("admin", "hupload")

	h1.ServeHTTP(nil, req)
}

func TestBasicWrongCredentials(t *testing.T) {
	c := authentication.FileAuthenticationConfig{
		Path: "tests/users.yml",
	}

	a, err := authentication.NewAuthenticationFile(c)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	m := BasicAuthMiddleware{
		Authentication: a,
	}

	fn1 := func(w http.ResponseWriter, r *http.Request) {
		c := r.Context().Value(AuthError)
		if c == nil {
			t.Errorf("Expected error, got nil")
		} else {
			if !errors.Is(c.(error), ErrBasicAuthAuthenticationFailed) {
				t.Errorf("Expected authentication failed, got %v", c)
			}
		}
	}

	h1 := m.Middleware(http.HandlerFunc(fn1))

	req, _ := http.NewRequest("GET", "https://example.com/", nil)
	req.SetBasicAuth("admin", "wrong")

	h1.ServeHTTP(nil, req)
}

func TestBasicAuthNoCredentials(t *testing.T) {
	c := authentication.FileAuthenticationConfig{
		Path: "tests/users.yml",
	}

	a, err := authentication.NewAuthenticationFile(c)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	m := BasicAuthMiddleware{
		Authentication: a,
	}

	fn1 := func(w http.ResponseWriter, r *http.Request) {
		c := r.Context().Value(AuthError).(error)
		if !errors.Is(c, ErrBasicAuthNoCredentials) {
			t.Errorf("Expected ErrBasicAuthNoCredentials, got %v", c)
		}
	}

	h1 := m.Middleware(http.HandlerFunc(fn1))

	req, _ := http.NewRequest("GET", "https://example.com/", nil)

	h1.ServeHTTP(nil, req)
}
