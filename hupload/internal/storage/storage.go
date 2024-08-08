package storage

import (
	"bufio"
	"time"
)

type Share struct {
	Name        string    `json:"name"`
	DateCreated time.Time `json:"created"`
	Owner       string    `json:"owner"`
	Validity    int       `json:"validity"`

	Size  int64 `json:"size"`
	Count int64 `json:"count"`

	Valid bool `json:"isvalid"`
}

func (s *Share) IsValid() bool {
	if s.Validity == 0 {
		return true
	}
	validUntil := s.DateCreated.AddDate(0, 0, s.Validity)
	return validUntil.After(time.Now())
}

type Item struct {
	Path     string
	ItemInfo ItemInfo
}

type ItemInfo struct {
	Size         int64
	DateModified time.Time
}

// BackendInterface must be implemented by any backend
type Storage interface {
	// CreateShare creates a new share
	CreateShare(name, owner string, validity int) error

	// CreateItem creates a new item in a share
	CreateItem(share, item string, reader *bufio.Reader) (*Item, error)

	// GetShare returns the share identified by share
	GetShare(string) (*Share, error)

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

	// UpdateMetadata updates the metadata of share
	UpdateMetadata(string) error
}
