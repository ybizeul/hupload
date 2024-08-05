package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/ybizeul/hupload/pkg/apiws/authentication"
	"github.com/ybizeul/hupload/pkg/apiws/storage"
)

// Config is the internal representation of Hupload configuration file
type Config struct {
	Path   string
	Values map[string]any
}

// Load reads the configuration file and populates the Config struct
func (c *Config) Load() (bool, error) {
	// Set default templating values
	c.Values = map[string]any{"title": "Hupload"}

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
	b, ok := c.Values["storage"].(map[string]any)
	if !ok {
		return DefaultStorage(), nil
	}

	// Check if storage type is valid
	storageType, ok := b["type"].(string)
	if !ok {
		return nil, errors.New("invalid storage backend")
	}

	switch storageType {
	case "file":
		return storage.NewFileStorage(b), nil
	}

	return nil, errors.New("unknown backend")
}

// If no storage configuration is defined, use the default one
func DefaultStorage() storage.Storage {
	return storage.NewFileStorage(map[string]any{
		"options": map[string]any{
			"path": "data",
		},
	})
}

// Authentication returns the authentication backend struct that will be used
// to authenticate users.
func (c *Config) Authentication() (authentication.Authentication, error) {
	// Check if the configuration has a authentication backend defined
	b, ok := c.Values["auth"].(map[string]any)
	if !ok {
		return DefaultAuthentication(), nil
	}

	// Check if authentication type is valid
	backendType, ok := b["type"].(string)
	if !ok {
		return nil, errors.New("invalid backend")
	}

	switch backendType {
	case "file":
		return authentication.NewAuthenticationFile(b), nil
	}

	return nil, errors.New("unknown backend")
}

// If no authentication configuration is defined, use the default one
func DefaultAuthentication() authentication.Authentication {
	return authentication.NewAuthenticationDefault()
}
