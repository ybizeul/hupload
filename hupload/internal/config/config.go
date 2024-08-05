package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/ybizeul/hupload/pkg/apiws/authentication"
	"github.com/ybizeul/hupload/pkg/apiws/storage"
)

type TypeOptions struct {
	Type    string
	Options map[string]any
}

type ConfigValues struct {
	Title          string
	Storage        TypeOptions `yaml:"storage"`
	Authentication TypeOptions `yaml:"auth"`
}

// Config is the internal representation of Hupload configuration file
type Config struct {
	Path   string
	Values ConfigValues
}

// Load reads the configuration file and populates the Config struct
func (c *Config) Load() (bool, error) {
	// Set default templating values
	c.Values = ConfigValues{
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

	// Open the configuration file
	f, err := os.Open(c.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return true, err
	}
	defer f.Close()

	// populate yaml content to Config struct
	err = yaml.NewDecoder(f).Decode(&c.Values)
	if err != nil {
		return true, err
	}

	return true, nil
}

// Storage returns the storage backend struct that will be used to create
// shares, store and retrieve content.
func (c *Config) Storage() (storage.Storage, error) {
	// Check if the configuration has a storage backend defined
	s := c.Values.Storage
	// if !ok {
	// 	return DefaultStorage(), nil
	// }

	// Check if storage type is valid
	if s.Type == "" {
		return nil, ErrMissingStorageBackendType
	}

	switch s.Type {
	case "file":
		return storage.NewFileStorage(s.Options), nil
	}

	return nil, ErrUnknownStorageBackend
}

// // If no storage configuration is defined, use the default one
// func DefaultStorage() storage.Storage {
// 	return storage.NewFileStorage(map[string]any{
// 		"options": map[string]any{
// 			"path": "data",
// 		},
// 	})
// }

// Authentication returns the authentication backend struct that will be used
// to authenticate users.
func (c *Config) Authentication() (authentication.Authentication, error) {
	// Check if the configuration has a authentication backend defined
	a := c.Values.Authentication

	// Check if authentication type is valid
	if a.Type == "" {
		return nil, ErrMissingAuthenticationBackendType
	}

	switch a.Type {
	case "file":
		return authentication.NewAuthenticationFile(a.Options)
	case "default":
		return authentication.NewAuthenticationDefault(), nil
	}

	return nil, ErrUnknownAuthenticationBackend
}

// // If no authentication configuration is defined, use the default one
// func DefaultAuthentication() authentication.Authentication {
// 	return authentication.NewAuthenticationDefault()
// }
