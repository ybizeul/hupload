package storage

import (
	"bufio"
	"io"
	"time"
)

type Options struct {
	Validity    int    `json:"validity,omitempty"`
	Exposure    string `json:"exposure"`
	Description string `json:"description,omitempty"`
	Message     string `json:"message"`
}

func DefaultOptions() Options {
	return Options{
		Validity: 7,
		Exposure: "upload",
	}
}

type Share struct {
	Version     int       `json:"version,omitempty"`
	Name        string    `json:"name"`
	DateCreated time.Time `json:"created,omitempty"`
	Owner       string    `json:"owner,omitempty"`
	Options     Options   `json:"options,omitempty"`

	Size  int64 `json:"size,omitempty"`
	Count int64 `json:"count,omitempty"`
}

func (s *Share) IsValid() bool {
	if s.Options.Validity == 0 {
		return true
	}

	validUntil := s.DateCreated.AddDate(0, 0, s.Options.Validity)

	return validUntil.After(time.Now())
}

type PublicShare struct {
	Name    string        `json:"name"`
	Options PublicOptions `json:"options,omitempty"`
}

type PublicOptions struct {
	Exposure string `json:"exposure"`
	Message  string `json:"message"`
}

func (s *Share) PublicShare() *PublicShare {
	return &PublicShare{
		Name: s.Name,
		Options: PublicOptions{
			Exposure: s.Options.Exposure,
			Message:  s.Options.Message,
		},
	}
}

func PublicShares(shares []Share) []PublicShare {
	publicShares := make([]PublicShare, 0)

	for _, s := range shares {
		publicShares = append(publicShares, *s.PublicShare())
	}

	return publicShares
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
	// Migrate will be called at initialization to give an opportunity to
	// the backend to migrate data from a previous version to the current one
	Migrate() error

	// CreateShare creates a new share
	CreateShare(name, owner string, options Options) (*Share, error)

	// UpdateShare updates an existing share
	UpdateShare(name string, options *Options) (*Options, error)

	// CreateItem creates a new item in a share
	CreateItem(share, item string, size int64, reader *bufio.Reader) (*Item, error)

	// CreateItem creates a new item in a share
	DeleteItem(share, item string) error

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
	GetItemData(string, string) (io.ReadCloser, error)
}
