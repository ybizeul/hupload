package authentication

import (
	"errors"
	"net/http"
	"testing"
)

func TestAuthentication(t *testing.T) {
	c := FileAuthenticationConfig{
		Path: "file_testdata/users.yml",
	}

	a, err := NewAuthenticationFile(c)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	r, _ := http.NewRequest("GET", "http://localhost:8080", nil)
	r.SetBasicAuth("admin", "hupload")

	a.AuthenticateRequest(nil, r, func(ok bool, err error) {
		if !ok {
			t.Errorf("Expected true, got false")
		}
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	r, _ = http.NewRequest("GET", "http://localhost:8080", nil)
	r.SetBasicAuth("admin", "random")

	a.AuthenticateRequest(nil, r, func(ok bool, err error) {
		if ok {
			t.Errorf("Expected false, got true")
		}
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	r, _ = http.NewRequest("GET", "http://localhost:8080", nil)
	r.SetBasicAuth("nonexistant", "random")

	a.AuthenticateRequest(nil, r, func(ok bool, err error) {
		if ok {
			t.Errorf("Expected false, got true")
		}
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

func TestAuthenticationInexistantUsersFile(t *testing.T) {
	c := FileAuthenticationConfig{
		Path: "file_testdata/users_inexistant.yml",
	}

	_, err := NewAuthenticationFile(c)

	if !errors.Is(err, ErrAuthenticationMissingUsersFile) {
		t.Errorf("Expected error, got nil")
	}
}
