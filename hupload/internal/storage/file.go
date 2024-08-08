package storage

import (
	"bufio"
	"encoding/json"
	"errors"
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

// isShareNameSafe checks if a share name is safe to use,, the primary goal is
// to make sure that no path traversal is possible
func isShareNameSafe(n string) bool {
	m := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(n)
	return m
}

// CreateShare creates a new share with the provided name, owner and validity
// in days. It returns an error if the share already exists or if the name is
// invalid. owner is only used to populate metadata.

func (b *FileBackend) CreateShare(name, owner string, validity int) error {
	if !isShareNameSafe(name) {
		return errors.New("invalid share name")
	}
	_, err := os.Stat(path.Join(b.Options.Path, name))
	if err == nil {
		return errors.New("share already exists")
	}

	p := path.Join(b.Options.Path, name)
	err = os.Mkdir(p, 0755)
	if err != nil {
		slog.Error("cannot create share", slog.String("error", err.Error()), slog.String("path", p))
		return errors.New("cannot create share")
	}

	m := Share{
		Name:        name,
		Owner:       owner,
		DateCreated: time.Now(),
		Validity:    validity,
	}

	f, err := os.Create(path.Join(b.Options.Path, name, ".metadata"))
	if err != nil {
		return err
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(m)
	if err != nil {
		return err
	}

	return nil
}

// CreateItem creates a new item in the provided share with the provided name
// and content. It returns an error if the share does not exist, or if the item
// doesn't fit in the share or if the share is full. The content is read from
// the provided bufio.Reader.
var (
	ErrMaxShareSizeReached = errors.New("Max share size reached")
)

func (b *FileBackend) CreateItem(s string, i string, r *bufio.Reader) (*Item, error) {
	if !isShareNameSafe(s) {
		return nil, errors.New("invalid share name")
	}

	// Get Share metadata
	share, err := b.GetShare(s)
	if err != nil {
		return nil, errors.New("cannot get share")
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
		if maxWrite > maxItem || maxWrite == 0 {
			maxWrite = maxItem
		}
	}

	// maxWrite is the actual allowed size for the item, so we fix the limit
	// to one more byte
	if maxWrite > 0 {
		maxWrite++
	}

	// path.Join("/", i) is used to avoid path traversal
	p := path.Join(b.Options.Path, s, path.Join("/", i))
	f, err := os.Create(p + suffix)
	if err != nil {
		slog.Error("cannot create item", slog.String("error", err.Error()), slog.String("path", p))
		return nil, errors.New("cannot create item")
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
		return nil, errors.New("cannot copy item content")
	}

	// If nothing was written or if the max write limit was reached, remove the
	// temporary file and return an error
	if written == 0 || written == maxWrite {
		os.Remove(p + suffix)
		return nil, ErrMaxShareSizeReached
	}

	err = os.Rename(p+suffix, p)
	if err != nil {
		return nil, err
	}

	item, err := b.GetItem(s, i)
	if err != nil {
		return nil, err
	}

	err = b.updateMetadata(s)
	if err != nil {
		return nil, err
	}

	return item, nil
}

// GetShare retrieves the metadata for a share with the provided name. It
// returns an error if the share does not exist or if the name is invalid.

func (b *FileBackend) GetShare(s string) (*Share, error) {
	if !isShareNameSafe(s) {
		return nil, errors.New("invalid share name")
	}
	fm, err := os.Open(path.Join(b.Options.Path, s, ".metadata"))
	if err != nil {
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

func (b *FileBackend) ListShares() ([]Share, error) {
	d, err := os.ReadDir(b.Options.Path)
	if err != nil {
		return nil, err
	}
	r := []Share{}

	// Shares loop
	for _, f := range d {
		if f.IsDir() {
			m, err := b.GetShare(f.Name())
			if err != nil {
				continue
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

func (b *FileBackend) DeleteShare(s string) error {
	if !isShareNameSafe(s) {
		return errors.New("invalid share name")
	}
	sharePath := path.Join(b.Options.Path, s)
	err := os.RemoveAll(sharePath)
	if err != nil {
		return err
	}
	return nil
}

// ListShare returns the list of items in a share. It returns an error if the
// share does not exist or if the name is invalid. The items are sorted by
// modification date, newest first.
// .* files and temporary upload files are excluded from the result.

func (b *FileBackend) ListShare(s string) ([]Item, error) {
	if !isShareNameSafe(s) {
		return []Item{}, errors.New("invalid share name")
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

		i, err := b.GetItem(s, f.Name())
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
func (b *FileBackend) GetItem(s string, i string) (*Item, error) {
	if !isShareNameSafe(s) {
		return nil, errors.New("invalid share name")
	}

	// path.Join("/", i) is used to avoid path traversal
	p := path.Join(b.Options.Path, s, path.Join("/", i))

	stat, err := os.Stat(p)
	if err != nil {
		return nil, err
	}

	return &Item{
		Path:     path.Join(s, i),
		ItemInfo: ItemInfo{Size: stat.Size(), DateModified: stat.ModTime()},
	}, nil
}

// GetItemData retrieves the content of an item in a share. It returns an error
// if the share or the item do not exist or if the share name is invalid.
func (b *FileBackend) GetItemData(s string, i string) (io.ReadCloser, error) {
	if !isShareNameSafe(s) {
		return nil, errors.New("invalid share name")
	}

	// path.Join("/", i) is used to avoid path traversal
	p := path.Join(b.Options.Path, s, path.Join("/", i))

	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (b *FileBackend) updateMetadata(s string) error {
	if !isShareNameSafe(s) {
		return errors.New("invalid share name")
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
