package config

import (
	"errors"
	"reflect"
	"testing"

	"github.com/ybizeul/hupload/pkg/apiws/authentication"
	"github.com/ybizeul/hupload/pkg/apiws/storage"
)

func TestLoadEmptyConfig(t *testing.T) {
	c := Config{}
	c.Load()

	expect := ConfigValues{
		Title: "Hupload",
		Storage: TypeOptions{
			Type: "file",
			Options: map[string]any{
				"path": "data",
			},
		},
		Authentication: TypeOptions{
			Type: "default",
		},
	}

	got := c.Values

	if !reflect.DeepEqual(expect, got) {
		t.Errorf("Expected %v, got %v", expect, got)
	}
}

func TestLoadGoodConfig(t *testing.T) {
	c := Config{
		Path: "../../tests/config/config.yml",
	}
	b, err := c.Load()

	if !b {
		t.Errorf("Expected config file to be found")
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expect := ConfigValues{
		Title: "Hupload",
		Storage: TypeOptions{
			Type: "file",
			Options: map[string]any{
				"path":         "data",
				"max_file_mb":  500,
				"max_share_mb": 2000,
			},
		},
		Authentication: TypeOptions{
			Type: "file",
			Options: map[string]any{
				"path": "../../tests/config/users.yml",
			},
		},
	}

	got := c.Values

	if !reflect.DeepEqual(expect, got) {
		t.Errorf("Expected %v, got %v", expect, got)
	}

	s, err := c.Storage()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if reflect.TypeOf(s).String() != "*storage.FileBackend" {
		t.Errorf("Expected *config.FileBackend, got %v", reflect.TypeOf(s).String())
	}

	if s.(*storage.FileBackend).Options["max_file_mb"] != 500 {
		t.Errorf("Expected 500, got %v", s.(*storage.FileBackend).Options["max_file_mb"])
	}

	if s.(*storage.FileBackend).Options["max_share_mb"] != 2000 {
		t.Errorf("Expected 2000, got %v", s.(*storage.FileBackend).Options["max_share_mb"])
	}

	_, err = c.Authentication()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestLoadBadConfig(t *testing.T) {
	c := Config{
		Path: "../../tests/config/config_bad_syntax.txt",
	}
	b, err := c.Load()

	if !b {
		t.Errorf("Expected config file to be found")
	}

	if err == nil {
		t.Errorf("Expected error to be returned")
	}

}

func TestMissingUsersFile(t *testing.T) {
	c := Config{
		Path: "../../tests/config/config_missing_users_file.yml",
	}
	b, _ := c.Load()

	if !b {
		t.Errorf("Expected config file to be found")
	}

	_, err := c.Authentication()

	if !errors.Is(err, authentication.ErrAuthenticationMissingUsersFile) {
		t.Errorf("Expected ErrAuthenticationMissingUsersFile to be returned")
	}

}
