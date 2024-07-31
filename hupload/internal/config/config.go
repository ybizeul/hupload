package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/ybizeul/hupload/pkg/apiws/authservice"
	"github.com/ybizeul/hupload/pkg/apiws/storageservice"
)

// Config is the internal representation of Hupload configuration file
type Config struct {
	Path   string
	Values map[string]any
}

// Load reads the configuration file and populates the Config struct
func (c *Config) Load() error {
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

// Backend returns the storage backend struct that will be used to create
// shared, store and retrieve content.
func (c *Config) Backend() (storageservice.StorageServiceInterface, error) {
	b, ok := c.Values["backend"].(map[string]any)
	if !ok {
		return c.DefaultBackend(), nil
	}
	backendType, ok := b["type"].(string)
	if !ok {
		return nil, errors.New("invalid backend")
	}
	switch backendType {

	case "file":
		return storageservice.NewFileBackend(b), nil
	}
	return nil, errors.New("unknown backend")
}

func (c *Config) DefaultBackend() storageservice.StorageServiceInterface {
	return storageservice.NewFileBackend(map[string]any{
		"options": map[string]any{
			"path": "data",
		},
	})
}

func (c *Config) AuthBackend() (authservice.AuthServiceInterface, error) {
	b, ok := c.Values["auth"].(map[string]any)
	if !ok {
		return authservice.NewAuthBackendDefault(), nil
	}
	backendType, ok := b["type"].(string)
	if !ok {
		return nil, errors.New("invalid backend")
	}
	switch backendType {
	case "file":
		return authservice.NewAuthBackendBasic(b), nil
	}
	return nil, errors.New("unknown backend")
}
