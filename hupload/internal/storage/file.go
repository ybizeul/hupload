package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

const suffix = "_huploadtemp"

type FileStorageConfig struct {
	Path         string `yaml:"path"`
	MaxFileSize  int64  `yaml:"max_file_mb"`
	MaxShareSize int64  `yaml:"max_share_mb"`
}

// FileBackend is a backend that stores files on the filesystem
type FileBackend struct {
	Options             FileStorageConfig
	DefaultValidityDays int
}

// NewFileStorage creates a new FileBackend, m is the configuration as found
// in Hupload configuration file.
func NewFileStorage(m map[string]any) *FileBackend {
	b, err := yaml.Marshal(m)
	if err != nil {
		return nil
	}

	var r FileBackend

	err = yaml.Unmarshal(b, &r.Options)
	if err != nil {
		return nil
	}

	r.initialize()

	return &r
}

// initialize creates the root directory for the backend
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

func (b *FileBackend) CreateShare(name, owner string, validity int) error {
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

func (b *FileBackend) CreateItem(s string, i string, r *bufio.Reader) (*Item, error) {
	p := path.Join(b.Options.Path, s, i)
	f, err := os.Create(p + suffix)
	if err != nil {
		return nil, errors.New("cannot create item")
	}
	defer f.Close()

	share, err := b.GetShare(s)

	src := r

	maxWrite := int64(0)

	maxShare := b.Options.MaxShareSize * 1024 * 1024
	if maxShare > 0 {
		maxWrite = maxShare - share.Size
		if maxWrite <= 0 {
			return nil, errors.New("Max share capacity already reached")
		}
	}

	maxItem := b.Options.MaxFileSize * 1024 * 1024
	if maxItem > 0 {
		if maxWrite > maxItem || maxWrite == 0 {
			maxWrite = maxItem
		}
	}

	if maxWrite > 0 {
		// Check if there is a max bytes associated to share
		if err != nil {
			return nil, errors.New("cannot get share")
		}
		if maxWrite <= 0 {
			return nil, errors.New("Max share capacity already reached")
		}

		src = bufio.NewReader(io.LimitReader(r, maxWrite))
	}

	written, err := io.Copy(f, src)
	if err != nil {
		os.Remove(p + suffix)
		return nil, errors.New("cannot copy item content")
	}

	if written == 0 || written == maxWrite {
		os.Remove(p + suffix)
		return nil, errors.New("File too big or share is full")
	}

	err = os.Rename(p+suffix, p)
	if err != nil {
		return nil, errors.New("cannot rename item to final destination")
	}

	item, err := b.GetItem(s, i)
	if err != nil {
		return nil, errors.New("cannot get added item")
	}
	return item, nil
}

func (b *FileBackend) GetShare(s string) (*Share, error) {
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
	m.Valid = m.IsValid()

	return &m, nil
}

func (b *FileBackend) ListShares() ([]Share, error) {
	d, err := os.ReadDir(b.Options.Path)
	if err != nil {
		return nil, err
	}
	var r []Share

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

func (b *FileBackend) DeleteShare(s string) error {
	sharePath := path.Join(b.Options.Path, s)
	err := os.RemoveAll(sharePath)
	if err != nil {
		return err
	}
	return nil
}

func (b *FileBackend) ListShare(s string) ([]Item, error) {
	d, err := os.ReadDir(path.Join(b.Options.Path, s))
	if err != nil {
		return nil, err
	}
	r := []Item{}
	for _, f := range d {
		if strings.HasPrefix(f.Name(), ".") || strings.HasSuffix(f.Name(), suffix) {
			continue
		}
		i, err := b.GetItem(s, f.Name())
		if err != nil {
			return nil, err
		}
		r = append(r, *i)
	}

	sort.Slice(r, func(i, j int) bool {
		return r[i].ItemInfo.DateModified.After(r[j].ItemInfo.DateModified)
	})

	return r, nil
}

func (b *FileBackend) GetItem(s string, i string) (*Item, error) {
	p := path.Join(b.Options.Path, s, i)
	stat, err := os.Stat(p)
	if err != nil {
		return nil, err
	}
	return &Item{
		Path:     path.Join(s, i),
		ItemInfo: ItemInfo{Size: stat.Size(), DateModified: stat.ModTime()},
	}, nil
}

func (b *FileBackend) GetItemData(s string, i string) (*bufio.Reader, error) {
	p := path.Join(b.Options.Path, s, i)
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	return bufio.NewReader(f), nil
}

func (b *FileBackend) UpdateMetadata(s string) error {
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