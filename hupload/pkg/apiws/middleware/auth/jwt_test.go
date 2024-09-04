package auth

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ybizeul/hupload/pkg/apiws/authentication"
)

func TestJWTAuth(t *testing.T) {

	secret := "secret"

	m := JWTAuthMiddleware{
		HMACSecret: secret,
	}

	long, short, err := m.generateTokens("admin")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
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
			t.Errorf("Expected AuthStatusSuccess, got %t", s.Authenticated)
		}
		if s.User != "admin" {
			t.Errorf("Expected admin, got %v", s.User)
		}
		_, _ = w.Write([]byte("OK"))
	}

	h1 := m.Middleware(http.HandlerFunc(fn1))

	req, _ := http.NewRequest("GET", "https://example.com/", nil)
	req.AddCookie(&http.Cookie{Name: "X-Token", Value: short})
	req.AddCookie(&http.Cookie{Name: "X-Token-Refresh", Value: long})

	r := httptest.NewRecorder()

	h1.ServeHTTP(r, req)

	if r.Code != http.StatusOK {
		t.Errorf("Expected OK, got %v", r.Code)
	}
	b, _ := io.ReadAll(r.Result().Body)
	if string(b) != "OK" {
		t.Errorf("Expected OK, got %v", string(b))
	}
}

func TestJWTAuthBadSecret(t *testing.T) {
	// First generate a valid token with a different secret
	bad_m := JWTAuthMiddleware{
		HMACSecret: "badSecret",
	}

	long, short, err := bad_m.generateTokens("admin")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test token with correct secret
	secret := "secret"

	m := JWTAuthMiddleware{
		HMACSecret: secret,
	}

	fn1 := func(w http.ResponseWriter, r *http.Request) {
		s, ok := r.Context().Value(authentication.AuthStatusKey).(authentication.AuthStatus)
		if !ok {
			t.Errorf("Expected AuthStatus, got nil")
		}
		if s.Error == nil {
			t.Errorf("Expected error, got nil")
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("UNAUTHORIZED"))
	}

	h1 := m.Middleware(http.HandlerFunc(fn1))

	req, _ := http.NewRequest("GET", "https://example.com/", nil)
	req.AddCookie(&http.Cookie{Name: "X-Token", Value: short})
	req.AddCookie(&http.Cookie{Name: "X-Token-Refresh", Value: long})

	r := httptest.NewRecorder()

	h1.ServeHTTP(r, req)

	if r.Code != http.StatusUnauthorized {
		t.Errorf("Expected Unauthorized, got %v", r.Code)
	}

	c := r.Result().Cookies()
	if len(c) != 2 {
		t.Errorf("Expected 2 cookies, got %v", len(c))
	}
	for _, v := range c {
		if v.Value != "deleted" {
			t.Errorf("Expected deleted, got %v", v.Value)
		}
		if v.Expires.Unix() != 0 {
			t.Errorf("Expected 0, got %v", v.Expires.Unix())
		}
	}
	b, _ := io.ReadAll(r.Result().Body)
	if string(b) != "UNAUTHORIZED" {
		t.Errorf("Expected UNAUTHORIZED, got %v", string(b))
	}
}
