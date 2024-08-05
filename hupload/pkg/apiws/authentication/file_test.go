package authentication

import (
	"errors"
	"testing"
)

func TestAuthentication(t *testing.T) {
	m := map[string]any{
		"path": "../../../tests/config/users.yml",
	}
	a, err := NewAuthenticationFile(m)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	b, err := a.AuthenticateUser("admin", "hupload")
	if !b {
		t.Errorf("Expected true, got false")
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	b, err = a.AuthenticateUser("admin", "random")
	if b {
		t.Errorf("Expected false, got true")
	}
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	b, err = a.AuthenticateUser("nonexistant", "random")
	if b {
		t.Errorf("Expected false, got true")
	}
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestAuthenticationBadConfig(t *testing.T) {
	m := map[string]any{
		"path": 3,
	}
	_, err := NewAuthenticationFile(m)

	if errors.Is(err, ErrAuthenticationInvalidPath) {
		t.Errorf("Expected error, got nil")
	}
}

func TestAuthenticationInexistantUsersFile(t *testing.T) {
	m := map[string]any{
		"path": "nonexistant",
	}
	_, err := NewAuthenticationFile(m)

	if !errors.Is(err, ErrAuthenticationMissingUsersFile) {
		t.Errorf("Expected error, got nil")
	}
}
