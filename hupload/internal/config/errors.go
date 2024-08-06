package config

import "errors"

var (
	ErrorConfigNoSuchFile               = errors.New("missing configuration file")
	ErrMissingStorageBackendType        = errors.New("missing storage backend type")
	ErrMissingAuthenticationBackendType = errors.New("missing authentication backend type")
	ErrUnknownStorageBackend            = errors.New("unknown storage backend")
	ErrUnknownAuthenticationBackend     = errors.New("unknown authentication backend")
)
