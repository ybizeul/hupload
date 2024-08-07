package authentication

import (
	"errors"
	"testing"
)

func TestAuthentication(t *testing.T) {
	c := FileAuthenticationConfig{
		Path: "tests/users.yml",
	}

	a, err := NewAuthenticationFile(c)

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

func TestAuthenticationInexistantUsersFile(t *testing.T) {
	c := FileAuthenticationConfig{
		Path: "tests/users_inexistant.yml",
	}

	_, err := NewAuthenticationFile(c)

	if !errors.Is(err, ErrAuthenticationMissingUsersFile) {
		t.Errorf("Expected error, got nil")
	}
}
