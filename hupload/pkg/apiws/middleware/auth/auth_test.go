package auth

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ybizeul/hupload/pkg/apiws/authentication"
)

func TestServeNextAuthenticated(t *testing.T) {
	successMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ServeNextAuthenticated("user", next, w, r)
		})
	}

	fn1 := func(w http.ResponseWriter, r *http.Request) {
		s := r.Context().Value(authentication.AuthStatusKey).(authentication.AuthStatus)
		c := s.Error
		if c != nil {
			t.Errorf("Expected nil, got %v", c)
		}
		if !s.Authenticated {
			t.Errorf("Expected AuthStatusSuccess, got %v", c)
		}
		u := s.User
		if u != "user" {
			t.Errorf("Expected admin, got %v", u)
		}
	}

	h1 := successMiddleware(http.HandlerFunc(fn1))

	req, _ := http.NewRequest("GET", "https://example.com/", nil)

	h1.ServeHTTP(nil, req)
}

var fakeError = errors.New("Some Error")

func TestServeNextAuthFailed(t *testing.T) {
	successMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ServeNextError(next, w, r, fakeError)
		})
	}

	fn1 := func(w http.ResponseWriter, r *http.Request) {
		s := r.Context().Value(authentication.AuthStatusKey).(authentication.AuthStatus)
		if s.Error == nil {
			t.Errorf("Expected error, got nil")
		} else {
			if !errors.Is(s.Error, fakeError) {
				t.Errorf("Expected fakeError, got %v", s.Error)
			}
		}
	}

	h1 := successMiddleware(http.HandlerFunc(fn1))

	req, _ := http.NewRequest("GET", "https://example.com/", nil)

	h1.ServeHTTP(nil, req)
}

func TestConfirmAuthentication(t *testing.T) {
	successMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ServeNextAuthenticated("user", next, w, r)
		})
	}

	fn1 := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	}

	c := ConfirmAuthenticator{}

	h1 := successMiddleware(c.Middleware(http.HandlerFunc(fn1)))

	req, _ := http.NewRequest("GET", "https://example.com/", nil)

	r := httptest.NewRecorder()

	h1.ServeHTTP(r, req)

	if r.Code != http.StatusOK {
		t.Errorf("Expected 200, got %v", r.Code)
	}
}

func TestFailedAuthentication(t *testing.T) {
	successMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ServeNextError(next, w, r, fakeError)
		})
	}

	fn1 := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	}

	c := ConfirmAuthenticator{}

	h1 := successMiddleware(c.Middleware(http.HandlerFunc(fn1)))

	req, _ := http.NewRequest("GET", "https://example.com/", nil)

	r := httptest.NewRecorder()

	h1.ServeHTTP(r, req)

	if r.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %v", r.Code)
	}
	b, _ := io.ReadAll(r.Result().Body)
	if string(b) != "{\"errors\":[\"Some Error\"]}" {
		t.Errorf("Expected Unauthorized, got %v", string(b))
	}
}
