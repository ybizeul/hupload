package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ybizeul/hupload/pkg/apiws/authentication"
)

func TestBasicAuth(t *testing.T) {
	c := authentication.FileAuthenticationConfig{
		Path: "basic_testdata/users.yml",
	}

	a, err := authentication.NewAuthenticationFile(c)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	m := BasicAuthMiddleware{
		Authentication: a,
	}

	fn1 := func(w http.ResponseWriter, r *http.Request) {
		s, ok := r.Context().Value(authentication.AuthStatusKey).(authentication.AuthStatus)
		if !ok {
			t.Errorf("Expected AuthStatus, got nil")
		}
		if s.Error != nil {
			t.Errorf("Expected nil, got %v", s.Error)
		}
		if s.Authenticated == false {
			t.Errorf("Expected Success, got %t", s.Authenticated)
		}
		if s.User != "admin" {
			t.Errorf("Expected admin, got %v", s.User)
		}
	}

	h1 := m.Middleware(http.HandlerFunc(fn1))

	req, _ := http.NewRequest("GET", "https://example.com/", nil)
	req.SetBasicAuth("admin", "hupload")

	h1.ServeHTTP(nil, req)
}

func TestBasicWrongCredentials(t *testing.T) {
	confirmMiddleware := ConfirmAuthenticator{
		Realm: "test",
	}

	c := authentication.FileAuthenticationConfig{
		Path: "basic_testdata/users.yml",
	}

	a, err := authentication.NewAuthenticationFile(c)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	m := BasicAuthMiddleware{
		Authentication: a,
	}

	fn1 := func(w http.ResponseWriter, r *http.Request) {
		s, ok := r.Context().Value(authentication.AuthStatusKey).(authentication.AuthStatus)
		if !ok {
			t.Errorf("Expected AuthStatus, got nil")
		}
		if s.Error == nil {
			t.Errorf("Expected error, got nil")
		} else {
			if !errors.Is(s.Error, ErrBasicAuthAuthenticationFailed) {
				t.Errorf("Expected authentication failed, got %v", c)
			}
		}
	}

	h1 := confirmMiddleware.Middleware(m.Middleware(http.HandlerFunc(fn1)))

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "https://example.com/", nil)
	req.SetBasicAuth("admin", "wrong")

	h1.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %v", w.Code)
	}
}

func TestBasicAuthNoCredentials(t *testing.T) {

	confirmMiddleware := ConfirmAuthenticator{
		Realm: "test",
	}

	c := authentication.FileAuthenticationConfig{
		Path: "basic_testdata/users.yml",
	}

	a, err := authentication.NewAuthenticationFile(c)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	m := BasicAuthMiddleware{
		Authentication: a,
	}

	fn1 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// s, ok := r.Context().Value(authentication.AuthStatusKey).(authentication.AuthStatus)
		// if !ok {
		// 	t.Errorf("Expected AuthStatus, got nil")
		// }
		// if !errors.Is(s.Error, ErrBasicAuthNoCredentials) {
		// 	t.Errorf("Expected ErrBasicAuthNoCredentials, got %v", c)
		// }
	}

	h1 := confirmMiddleware.Middleware(m.Middleware(http.HandlerFunc(fn1)))

	req, _ := http.NewRequest("GET", "https://example.com/", nil)

	w := httptest.NewRecorder()
	h1.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %v", w.Code)
	}
}
