package storage

import "errors"

var (
	ErrInvalidShareName    = errors.New("invalid share name")
	ErrShareNotFound       = errors.New("share not found")
	ErrMaxShareSizeReached = errors.New("Max share size reached")

	ErrItemNotFound = errors.New("item not found")
)
