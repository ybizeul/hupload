package storageservice

import (
	"bufio"
	"time"
)

type Share struct {
	Name    string    `json:"name"`
	Created time.Time `json:"created"`
	Owner   string    `json:"owner"`
}

type Item struct {
	Path     string
	ItemInfo ItemInfo
}
type ItemInfo struct {
	Size int64
}

// BackendInterface must be implemented by any backend
type StorageServiceInterface interface {
	// CreateShare creates a new share
	CreateShare(string, string) error

	// CreateItem creates a new item in a share
	CreateItem(string, string, *bufio.Reader) (*Item, error)

	// ListShares returns the list of shares available
	ListShares() ([]Share, error)

	// ListShare returns the list of items in a share
	ListShare(string) ([]Item, error)

	// ListShare returns the list of items in a share
	DeleteShare(string) error

	// GetItem returns the item identified by share and item
	GetItem(string, string) (*Item, error)

	// GetItem returns the item identified by share and item
	GetItemData(string, string) (*bufio.Reader, error)
}
