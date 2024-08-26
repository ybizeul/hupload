package storage

import "errors"

var (
	ErrInvalidShareName    = errors.New("invalid share name")
	ErrShareNotFound       = errors.New("share not found")
	ErrMaxShareSizeReached = errors.New("Max share size reached")
	ErrMaxFileSizeReached  = errors.New("Max file size reached")
	ErrShareAlreadyExists  = errors.New("share already exists")

	ErrItemNotFound    = errors.New("item not found")
	ErrInvalidItemName = errors.New("invalid item name")

	ErrEmptyFile = errors.New("empty file")
)
