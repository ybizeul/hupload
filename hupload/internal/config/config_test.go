package config

import (
	"os"
	"reflect"
	"testing"

	"github.com/ybizeul/apiws/authentication"
	"github.com/ybizeul/hupload/internal/storage"
)

func TestLoadEmptyConfig(t *testing.T) {
	t.Cleanup(func() {
		os.Remove("data")
	})

	c := Config{}
	present, err := c.Load()

	if present {
		t.Errorf("Expected config file to be missing")
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expect := ConfigValues{
		Title:               "Hupload",
		DefaultValidityDays: 7,
		DefaultExposure:     "upload",
		HideOtherShares:     false,
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
	t.Cleanup(func() {
		os.Remove("data")
	})

	c := Config{
		Path: "config_testdata/config.yml",
	}
	b, err := c.Load()

	if !b {
		t.Errorf("Expected config file to be found")
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expect := ConfigValues{
		Title:               "Hupload Title",
		DefaultValidityDays: 12,
		DefaultExposure:     "download",
		HideOtherShares:     true,
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
				"path": "config_testdata/users.yml",
			},
		},
		MessageTemplates: []MessageTemplate{
			{
				Title:   "Message title",
				Message: "Message content",
			},
		},
	}

	got := c.Values

	if !reflect.DeepEqual(expect, got) {
		t.Errorf("Expected %v, got %v", expect, got)
	}

	s := c.Storage

	if reflect.TypeOf(s).String() != "*storage.FileBackend" {
		t.Errorf("Expected *config.FileBackend, got %v", reflect.TypeOf(s).String())
	}

	if s.(*storage.FileBackend).Options.MaxFileSize != 500 {
		t.Errorf("Expected 500, got %v", s.(*storage.FileBackend).Options.MaxFileSize)
	}

	if s.(*storage.FileBackend).Options.MaxShareSize != 2000 {
		t.Errorf("Expected 2000, got %v", s.(*storage.FileBackend).Options.MaxShareSize)
	}

	a := c.Authentication

	if a == nil {
		t.Errorf("Expected authentication backend to be created")
	}
}

func TestLoadBadConfig(t *testing.T) {
	c := Config{
		Path: "config_testdata/config_bad_syntax.txt",
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
	t.Cleanup(func() {
		os.Remove("data")
	})

	c := Config{
		Path: "config_testdata/config_missing_users_file.yml",
	}
	b, err := c.Load()

	if err != authentication.ErrAuthenticationMissingUsersFile {
		t.Errorf("Expected ErrAuthenticationMissingUsersFile, got %v", err)
	}

	if !b {
		t.Errorf("Expected config file to be found")
	}
}
