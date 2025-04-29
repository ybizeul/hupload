package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"
)

const suffix = "_huploadtemp"

// FileStorageConfig is the configuration structure for the file backend
// Path is the root directory where shares and items are stored
// MaxFileSize is the maximum size in MB for an item
// MaxShareSize is the maximum size in MB for a share
type FileStorageConfig struct {
	Path         string `yaml:"path"`
	MaxFileSize  int64  `yaml:"max_file_mb"`
	MaxShareSize int64  `yaml:"max_share_mb"`
}

// FileBackend is a backend that stores files on the filesystem
// Options is the configuration for the file storage backend
// DefaultValidityDays is a global option in the configuration file that
// sets the default validity of a share in days
type FileBackend struct {
	Options             FileStorageConfig
	DefaultValidityDays int
}

// NewFileStorage creates a new FileBackend with the provided options o
func NewFileStorage(o FileStorageConfig) *FileBackend {
	r := FileBackend{
		Options: o,
	}

	r.initialize()

	return &r
}

// initialize creates the root directory for the backend and panics if it can't
// be created or if no path is provided.

func (b *FileBackend) initialize() {
	path := b.Options.Path
	if path == "" {
		panic("path is required")
	}
	err := os.MkdirAll(path, 0755)
	if err != nil {
		panic(err)
	}
}

// IsShareNameSafe checks if a share name is safe to use,, the primary goal is
// to make sure that no path traversal is possible
func IsShareNameSafe(n string) bool {
	m := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(n)
	return m
}

func isItemNameSafe(n string) bool {
	return !strings.HasPrefix(n, ".")
}

// Migrate moves all previous metadata versions to new version
func (b *FileBackend) Migrate() error {
	type ShareV1 struct {
		Version     int       `json:"version"`
		Name        string    `json:"name"`
		DateCreated time.Time `json:"created,omitempty"`
		Owner       string    `json:"owner,omitempty"`
		Validity    int       `json:"validity"`
		Exposure    string    `json:"exposure"`
		Options     Options   `json:"options,omitempty"`
		Size        int64     `json:"size,omitempty"`
		Count       int64     `json:"count,omitempty"`
	}

	d, err := os.ReadDir(b.Options.Path)
	if err != nil {
		return err
	}
	for _, d := range d {
		if !d.IsDir() {
			continue
		}

		fm, err := os.Open(path.Join(b.Options.Path, d.Name(), ".metadata"))
		if err != nil {
			continue
		}

		var m ShareV1
		err = json.NewDecoder(fm).Decode(&m)
		if err != nil || m.Version > 0 {
			continue
		}
		fm.Close()

		o := Options{}
		if m.Options != o {
			o = m.Options
		} else {
			o = Options{
				Validity: m.Validity,
				Exposure: m.Exposure,
			}
		}

		nm := Share{
			Version:     1,
			Name:        m.Name,
			DateCreated: m.DateCreated,
			Owner:       m.Owner,
			Options:     o,
			Size:        m.Size,
			Count:       m.Count,
		}

		fm, err = os.Create(path.Join(b.Options.Path, d.Name(), ".metadata"))
		if err != nil {
			continue
		}

		err = json.NewEncoder(fm).Encode(nm)
		if err != nil {
			continue
		}
		fm.Close()
	}
	return nil
}

// CreateShare creates a new share with the provided name, owner and validity
// in days. It returns an error if the share already exists or if the name is
// invalid. owner is only used to populate metadata.

func (b *FileBackend) CreateShare(ctx context.Context, name, owner string, options Options) (*Share, error) {
	if !IsShareNameSafe(name) {
		return nil, ErrInvalidShareName
	}

	_, err := os.Stat(path.Join(b.Options.Path, name))
	if err == nil {
		return nil, ErrShareAlreadyExists
	}

	p := path.Join(b.Options.Path, name)
	err = os.Mkdir(p, 0755)
	if err != nil {
		slog.Error("cannot create share", slog.String("error", err.Error()), slog.String("path", p))
		return nil, err
	}

	if options.Exposure == "" {
		options.Exposure = "upload"
	}

	m := Share{
		Version:     1,
		Name:        name,
		Owner:       owner,
		DateCreated: time.Now(),
		Options:     options,
	}

	f, err := os.Create(path.Join(b.Options.Path, name, ".metadata"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(m)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

// UpdateShare updates the metadata of a share with the provided name. It returns
// an error if the share does not exist or if the name is invalid.

func (b *FileBackend) UpdateShare(ctx context.Context, name string, options *Options) (*Options, error) {
	if !IsShareNameSafe(name) {
		return nil, ErrInvalidShareName
	}

	m, err := b.GetShare(ctx, name)
	if err != nil {
		return nil, err
	}

	m.Options = *options

	f, err := os.Create(path.Join(b.Options.Path, name, ".metadata"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(m)
	if err != nil {
		return nil, err
	}

	return &m.Options, nil
}

// CreateItem creates a new item in the provided share with the provided name
// and content. It returns an error if the share does not exist, or if the item
// doesn't fit in the share or if the share is full. The content is read from
// the provided bufio.Reader.

func (b *FileBackend) CreateItem(ctx context.Context, s string, i string, size int64, r io.Reader) (*Item, error) {
	if !IsShareNameSafe(s) {
		return nil, ErrInvalidShareName
	}

	// Get Share metadata
	share, err := b.GetShare(ctx, s)
	if err != nil {
		return nil, err
	}

	// Check amount of free capacity in share according to current limits
	maxWrite := int64(0)

	maxShare := b.Options.MaxShareSize * 1024 * 1024
	if maxShare > 0 {
		maxWrite = maxShare - share.Size
		if maxWrite <= 0 {
			return nil, ErrMaxShareSizeReached
		}
	}

	maxItem := b.Options.MaxFileSize * 1024 * 1024
	if maxItem > 0 {
		if size > 0 && maxItem < size {
			return nil, ErrMaxFileSizeReached
		}
		if maxWrite > maxItem || maxWrite == 0 {
			maxWrite = maxItem
		}
	}

	// maxWrite is the actual allowed size for the item, so we fix the limit
	// to one more byte
	if maxWrite > 0 {
		maxWrite++

		if size > 0 && maxWrite < size {
			return nil, ErrMaxShareSizeReached
		}
	}

	// path.Join("/", i) is used to avoid path traversal
	p := path.Join(b.Options.Path, s, path.Join("/", i))
	f, err := os.Create(p + suffix)
	if err != nil {
		slog.Error("cannot create item", slog.String("error", err.Error()), slog.String("path", p))
		return nil, err
	}
	defer f.Close()

	src := r
	// Substitute bufio.Reader with a limited reader
	if maxWrite != 0 {
		src = bufio.NewReader(io.LimitReader(r, maxWrite))
	}

	written, err := io.Copy(f, src)
	if err != nil {
		os.Remove(p + suffix)
		return nil, err
	}

	// If nothing was written or if the max write limit was reached, remove the
	// temporary file and return an error
	if written == maxWrite {
		os.Remove(p + suffix)
		return nil, ErrMaxShareSizeReached
	}

	err = os.Rename(p+suffix, p)
	if err != nil {
		return nil, err
	}

	item, err := b.GetItem(ctx, s, i)
	if err != nil {
		return nil, err
	}

	err = b.updateMetadata(s)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (b *FileBackend) DeleteItem(ctx context.Context, s string, i string) error {
	if !IsShareNameSafe(s) {
		return ErrInvalidShareName
	}

	if !isItemNameSafe(i) {
		return ErrInvalidItemName
	}

	p := path.Join(b.Options.Path, s, path.Join("/", i))
	err := os.Remove(p)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrItemNotFound
		}
		return err
	}

	err = b.updateMetadata(s)
	if err != nil {
		return err
	}

	return nil
}

// GetShare retrieves the metadata for a share with the provided name. It
// returns an error if the share does not exist or if the name is invalid.

func (b *FileBackend) GetShare(ctx context.Context, s string) (*Share, error) {
	if !IsShareNameSafe(s) {
		return nil, ErrInvalidShareName
	}
	fm, err := os.Open(path.Join(b.Options.Path, s, ".metadata"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrShareNotFound
		}
		return nil, err
	}
	defer fm.Close()

	var m Share
	err = json.NewDecoder(fm).Decode(&m)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

// ListShares returns the list of shares available in the backend. It returns
// an error if the directory can't be read or if the metadata file can't be
// decoded.
// The returned shares are sorted by creation date, newest first.

func (b *FileBackend) ListShares(ctx context.Context) ([]Share, error) {
	d, err := os.ReadDir(b.Options.Path)
	if err != nil {
		return nil, err
	}
	r := []Share{}

	// Shares loop
	for _, f := range d {
		if f.IsDir() {
			m, err := b.GetShare(ctx, f.Name())
			if err != nil {
				continue
			}
			if os.Getenv("DEMO") == "true" {
				referenceDate := time.Date(2024, time.September, 1, 16, 48, 0, 0, time.UTC)
				delta := time.Since(referenceDate)
				m.DateCreated = m.DateCreated.Add(delta)
			}
			r = append(r, *m)
		}
	}
	sort.Slice(r, func(i, j int) bool {
		return r[i].DateCreated.After(r[j].DateCreated)
	})

	return r, nil
}

// DeleteShare removes a share and all its content from the backend. It returns
// an error if the share does not exist or if the name is invalid.

func (b *FileBackend) DeleteShare(ctx context.Context, s string) error {
	if !IsShareNameSafe(s) {
		return ErrInvalidShareName
	}
	sharePath := path.Join(b.Options.Path, s)

	_, err := os.Stat(sharePath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrShareNotFound
		}
		return err
	}
	err = os.RemoveAll(sharePath)
	if err != nil {
		return err
	}
	return nil
}

// ListShare returns the list of items in a share. It returns an error if the
// share does not exist or if the name is invalid. The items are sorted by
// modification date, newest first.
// .* files and temporary upload files are excluded from the result.

func (b *FileBackend) ListShare(ctx context.Context, s string) ([]Item, error) {
	if !IsShareNameSafe(s) {
		return nil, ErrInvalidShareName
	}

	d, err := os.ReadDir(path.Join(b.Options.Path, s))
	if err != nil {
		return nil, err
	}

	r := []Item{}
	for _, f := range d {
		// Ignore dotfiles and temporary files
		if strings.HasPrefix(f.Name(), ".") || strings.HasSuffix(f.Name(), suffix) {
			continue
		}

		i, err := b.GetItem(ctx, s, f.Name())
		if err != nil {
			return nil, err
		}

		r = append(r, *i)
	}

	// Sort items by modification date, newest first
	sort.Slice(r, func(i, j int) bool {
		return r[i].ItemInfo.DateModified.After(r[j].ItemInfo.DateModified)
	})

	return r, nil
}

// GetItem retrieves the metadata for an item in a share. It returns an error if
// the share or the item do not exist or if the share name is invalid.
func (b *FileBackend) GetItem(ctx context.Context, s string, i string) (*Item, error) {
	if !IsShareNameSafe(s) {
		return nil, ErrInvalidShareName
	}

	if !isItemNameSafe(i) {
		return nil, ErrInvalidItemName
	}

	// path.Join("/", i) is used to avoid path traversal
	p := path.Join(b.Options.Path, s, path.Join("/", i))

	stat, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrItemNotFound
		}
		return nil, err
	}

	return &Item{
		Path:     path.Join(s, i),
		ItemInfo: ItemInfo{Size: stat.Size(), DateModified: stat.ModTime()},
	}, nil
}

// GetItemData retrieves the content of an item in a share. It returns an error
// if the share or the item do not exist or if the share name is invalid.
func (b *FileBackend) GetItemData(ctx context.Context, s string, i string) (io.ReadCloser, error) {
	if !IsShareNameSafe(s) {
		return nil, ErrInvalidShareName
	}

	if !isItemNameSafe(i) {
		return nil, ErrInvalidItemName
	}

	// path.Join("/", i) is used to avoid path traversal
	p := path.Join(b.Options.Path, s, path.Join("/", i))

	f, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrItemNotFound
		}
		return nil, err
	}
	return f, nil
}

func (b *FileBackend) updateMetadata(s string) error {
	if !IsShareNameSafe(s) {
		return ErrInvalidShareName
	}
	sd, err := os.ReadDir(path.Join(b.Options.Path, s))
	if err != nil {
		return err
	}

	fm, err := os.OpenFile(path.Join(b.Options.Path, s, ".metadata"), os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer fm.Close()

	m := Share{}
	err = json.NewDecoder(fm).Decode(&m)
	if err != nil {
		return err
	}

	m.Size = 0
	m.Count = 0

	// Share content loop
	for _, i := range sd {
		if strings.HasPrefix(i.Name(), ".") {
			continue
		}
		info, err := i.Info()
		if err != nil {
			slog.Error("cannot get file info", slog.String("error", err.Error()))
			continue
		}
		m.Size += info.Size()
		m.Count += 1
	}

	_, err = fm.Seek(0, 0)
	if err != nil {
		return err
	}

	err = json.NewEncoder(fm).Encode(m)
	if err != nil {
		return err
	}

	return nil
}
