package storage

import "errors"

var (
	ErrInvalidShareName    = errors.New("invalid share name")
	ErrShareNotFound       = errors.New("share not found")
	ErrMaxShareSizeReached = errors.New("Max share size reached")
	ErrShareAlreadyExists  = errors.New("share already exists")

	ErrItemNotFound    = errors.New("item not found")
	ErrInvalidItemName = errors.New("invalid item name")
)