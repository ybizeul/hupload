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
func (c *Config) Load() error {
	c.Values = map[string]any{"title": "Hupload"}
	f, err := os.Open(c.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrorConfigNoSuchFile
		}
		return err
	}
	defer f.Close()
	err = yaml.NewDecoder(f).Decode(&c.Values)
	if err != nil {
		return err
	}
	return nil
}

// Storage returns the storage backend struct that will be used to create
// shares, store and retrieve content.
func (c *Config) Storage() (storage.StorageInterface, error) {
	b, ok := c.Values["storage"].(map[string]any)
	if !ok {
		return DefaultStorage(), nil
	}
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

func DefaultStorage() storage.StorageInterface {
	return storage.NewFileStorage(map[string]any{
		"options": map[string]any{
			"path": "data",
		},
	})
}

func (c *Config) Authentication() (authentication.AuthenticationInterface, error) {
	b, ok := c.Values["auth"].(map[string]any)
	if !ok {
		return DefaultAuthentication(), nil
	}
	backendType, ok := b["type"].(string)
	if !ok {
		return nil, errors.New("invalid backend")
	}
	switch backendType {
	case "file":
		return authentication.NewAuthenticationBasic(b), nil
	}
	return nil, errors.New("unknown backend")
}

func DefaultAuthentication() authentication.AuthenticationInterface {
	return authentication.NewAuthenticationDefault()
}
