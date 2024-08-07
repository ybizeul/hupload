package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/ybizeul/hupload/internal/storage"
	"github.com/ybizeul/hupload/pkg/apiws/authentication"
)

type TypeOptions struct {
	Type    string
	Options map[string]any
}

type ConfigValues struct {
	Title               string
	DefaultValidityDays int         `yaml:"availability_days"`
	Storage             TypeOptions `yaml:"storage"`
	Authentication      TypeOptions `yaml:"auth"`
}

// Config is the internal representation of Hupload configuration file
type Config struct {
	Path   string
	Values ConfigValues

	Storage        storage.Storage
	Authentication authentication.Authentication
}

// Load reads the configuration file and populates the Config struct
func (c *Config) Load() (fileExists bool, err error) {
	// Set default templating values
	c.Values = ConfigValues{
		Title:               "Hupload",
		DefaultValidityDays: 7,
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

	fileExists = true

	defer func() {
		if err != nil {
			return
		}

		c.Storage, err = c.storage()
		if err != nil {
			return
		}

		c.Authentication, err = c.authentication()
		if err != nil {
			return
		}
	}()

	// Open the configuration file
	f, err2 := os.Open(c.Path)
	if err2 != nil {
		if errors.Is(err2, os.ErrNotExist) {
			fileExists = false
			return
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

// storage returns the storage backend struct that will be used to create
// shares, store and retrieve content.

func (c *Config) storage() (storage.Storage, error) {
	s := c.Values.Storage
	if s.Type == "" {
		return nil, ErrMissingStorageBackendType
	}

	switch c.Values.Storage.Type {
	case "file":
		var options storage.FileStorageConfig
		b, err := yaml.Marshal(s.Options)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(b, &options)
		if err != nil {
			return nil, err
		}

		return storage.NewFileStorage(options), nil
	}

	return nil, ErrUnknownStorageBackend
}

// authentication returns the authentication backend struct that will be used
// to authenticate users.
func (c *Config) authentication() (authentication.Authentication, error) {
	a := c.Values.Authentication

	switch a.Type {
	case "file":
		var options authentication.FileAuthenticationConfig

		b, err := yaml.Marshal(a.Options)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(b, &options)
		if err != nil {
			return nil, err
		}
		return authentication.NewAuthenticationFile(options)
	case "default":
		return authentication.NewAuthenticationDefault(), nil
	case "":
		return nil, ErrMissingAuthenticationBackendType
	}

	return nil, ErrUnknownAuthenticationBackend
}
