package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/ybizeul/hupload/internal/storage"
	"github.com/ybizeul/hupload/pkg/apiws/authentication"
)

// TypeOption is a Type as a string and a map of options for the yaml
// configuration file.
// It is used for storage and authentication backends so new backends can be
// added without modifying the code.
// When loading configuration, the Type is matched in a switch statement and
// the corresponding backend is created. Options are then marshalled and
// unmarshalled to the configuration struct of the corresponding backend.
type TypeOptions struct {
	Type    string
	Options map[string]any
}

// ConfigValues is the struct that will be populated by the yaml configuration
// file.
// It will also be sent as-is to the templating engine to render the variables
// in the HTML templates, i.e. {{.Title}} will be replaced by the value of
// ConfigValues.Title.
type ConfigValues struct {
	Title               string
	DefaultValidityDays int               `yaml:"availability_days"`
	Storage             TypeOptions       `yaml:"storage"`
	Authentication      TypeOptions       `yaml:"auth"`
	MessageTemplates    []MessageTemplate `yaml:"messages"`
}

// Config is the internal representation of Hupload configuration file at path
// Path. Storage and Authentication are interfaces to the actual backends used
// to store shares data and authenticate users.
type Config struct {
	Path   string
	Values ConfigValues

	Storage        storage.Storage
	Authentication authentication.Authentication
}

// Load reads the configuration file and populates the Config struct
// accordingly. The actual creation of the storage and authentication backends
// is done into the defer block so that if no error is populated in err, the
// Config struct is fully populated with default backends if none had been
// defined during execution, like when the config file is missing.
// fileExists is a boolean that is set to false if the configuration file is
// missing so appropriate action can be taken by the caller.

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
// shares, store and retrieve content. It requires that the configuration
// struct has been populated with the values from the yaml configuration file.
// it returns a storage.Storage interface and an error if something failed

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

	case "s3":
		var options storage.S3StorageConfig
		b, err := yaml.Marshal(s.Options)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(b, &options)
		if err != nil {
			return nil, err
		}

		if options.Region == "" {
			if os.Getenv("AWS_DEFAULT_REGION") == "" {
				panic("missing region parameter or AWS_DEFAULT_REGION environment for s3 config !")
			}
			options.Region = os.Getenv("AWS_DEFAULT_REGION")
		}

		if options.AWSKey == "" {
			if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
				panic("missing aws_key parameter or AWS_ACCESS_KEY_ID environment for s3 config !")
			}
			options.AWSKey = os.Getenv("AWS_ACCESS_KEY_ID")
		}

		if options.AWSSecret == "" {
			if os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
				panic("missing aws_secret parameter or AWS_SECRET_ACCESS_KEY environment for s3 config !")
			}
			options.AWSSecret = os.Getenv("AWS_SECRET_ACCESS_KEY")
		}

		if options.Bucket == "" {
			if os.Getenv("BUCKET") == "" {
				panic("missing bucket parameter or BUCKET environment for s3 config !")
			}
			options.Bucket = os.Getenv("BUCKET")
		}

		return storage.NewS3Storage(options), nil
	}

	return nil, ErrUnknownStorageBackend
}

// authentication returns the authentication backend struct that will be used to
// authenticate users. It requires that the configuration struct has been
// populated with the values from the yaml configuration file. It returns an
// authentication.Authentication interface and an error if something failed

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
	case "oidc":
		var options authentication.AuthenticationOIDCConfig

		b, err := yaml.Marshal(a.Options)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(b, &options)
		if err != nil {
			return nil, err
		}
		return authentication.NewAuthenticationOIDC(options)
	case "default":
		return authentication.NewAuthenticationDefault(), nil
	}

	return nil, ErrUnknownAuthenticationBackend
}
