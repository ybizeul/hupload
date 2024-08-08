package auth

import (
	"net/http"
	"testing"
)

func TestOpenAuthWithCredentials(t *testing.T) {

	m := OpenAuthMiddleware{}

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
		if u != nil {
			t.Errorf("Expected nil, got %v", u)
		}
	}

	h1 := m.Middleware(http.HandlerFunc(fn1))

	req, _ := http.NewRequest("GET", "https://example.com/", nil)
	req.SetBasicAuth("admin", "hupload")

	h1.ServeHTTP(nil, req)
}

func TestOpenAuthWithoutCredentials(t *testing.T) {

	m := OpenAuthMiddleware{}

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
		if u != nil {
			t.Errorf("Expected nil, got %v", u)
		}
	}

	h1 := m.Middleware(http.HandlerFunc(fn1))

	req, _ := http.NewRequest("GET", "https://example.com/", nil)

	h1.ServeHTTP(nil, req)
}
