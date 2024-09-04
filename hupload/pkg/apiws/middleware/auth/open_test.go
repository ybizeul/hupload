package auth

import (
	"net/http"
	"testing"

	"github.com/ybizeul/hupload/pkg/apiws/authentication"
)

func TestOpenAuthWithCredentials(t *testing.T) {

	m := OpenAuthMiddleware{}

	fn1 := func(w http.ResponseWriter, r *http.Request) {
		s, ok := r.Context().Value(authentication.AuthStatusKey).(authentication.AuthStatus)
		if !ok {
			t.Errorf("Expected AuthStatus, got nil")
		}
		if s.Error != nil {
			t.Errorf("Expected nil, got %v", s.Error)
		}
		if !s.Authenticated {
			t.Errorf("Expected AuthStatusSuccess, got %t", s.Authenticated)
		}
		if s.User != "" {
			t.Errorf("Expected nil, got %s", s.User)
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
		s, ok := r.Context().Value(authentication.AuthStatusKey).(authentication.AuthStatus)
		if !ok {
			t.Errorf("Expected AuthStatus, got nil")
		}
		if s.Error != nil {
			t.Errorf("Expected nil, got %v", s.Error)
		}
		if !s.Authenticated {
			t.Errorf("Expected AuthStatusSuccess, got %t", s.Authenticated)
		}
		if s.User != "" {
			t.Errorf("Expected nil, got %v", s.User)
		}
	}

	h1 := m.Middleware(http.HandlerFunc(fn1))

	req, _ := http.NewRequest("GET", "https://example.com/", nil)

	h1.ServeHTTP(nil, req)
}
