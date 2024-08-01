package storageservice

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

const suffix = "_huploadtemp"

// FileBackend is a backend that stores files on the filesystem
type FileBackend struct {
	Options map[string]any
}

// NewFileBackend creates a new FileBackend, m is the configuration as found
// in Hupload configuration file.
func NewFileBackend(m map[string]any) *FileBackend {
	r := &FileBackend{
		Options: m["options"].(map[string]any),
	}
	r.initialize()

	return r
}

// initialize creates the root directory for the backend
func (b *FileBackend) initialize() {
	path := b.Options["path"].(string)
	if path == "" {
		panic("path is required")
	}
	err := os.MkdirAll(path, 0755)
	if err != nil {
		panic(err)
	}
}

func (b *FileBackend) CreateShare(s string, o string) error {
	_, err := os.Stat(path.Join(b.Options["path"].(string), s))
	if err == nil {
		return errors.New("share already exists")
	}

	err = os.Mkdir(path.Join(b.Options["path"].(string), s), 0755)
	if err != nil {
		return errors.New("cannot create share")
	}

	m := Share{
		Name:    s,
		Owner:   o,
		Created: time.Now(),
	}
	f, err := os.Create(path.Join(b.Options["path"].(string), s, ".metadata"))
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
	p := path.Join(b.Options["path"].(string), s, i)
	f, err := os.Create(p + suffix)
	if err != nil {
		return nil, errors.New("cannot create item")
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	if err != nil {
		os.Remove(p + suffix)
		return nil, errors.New("cannot copy item content")
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

func (b *FileBackend) ListShares() ([]Share, error) {
	d, err := os.ReadDir(b.Options["path"].(string))
	if err != nil {
		return nil, err
	}
	r := []Share{}
	for _, f := range d {
		if f.IsDir() {
			fm, err := os.Open(path.Join(b.Options["path"].(string), f.Name(), ".metadata"))
			if err != nil {
				continue
			}
			defer fm.Close()
			m := Share{}
			err = json.NewDecoder(fm).Decode(&m)
			if err != nil {
				return nil, err
			}
			r = append(r, m)
		}
	}
	sort.Slice(r, func(i, j int) bool {
		return r[i].Created.After(r[j].Created)
	})

	return r, nil
}

func (b *FileBackend) DeleteShare(s string) error {
	sharePath := path.Join(b.Options["path"].(string), s)
	err := os.RemoveAll(sharePath)
	if err != nil {
		return err
	}
	return nil
}

func (b *FileBackend) ListShare(s string) ([]Item, error) {
	d, err := os.ReadDir(path.Join(b.Options["path"].(string), s))
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
	return r, nil
}

func (b *FileBackend) GetItem(s string, i string) (*Item, error) {
	p := path.Join(b.Options["path"].(string), s, i)
	stat, err := os.Stat(p)
	if err != nil {
		return nil, err
	}
	return &Item{
		Path:     path.Join(s, i),
		ItemInfo: ItemInfo{Size: stat.Size()},
	}, nil
}

func (b *FileBackend) GetItemData(s string, i string) (*bufio.Reader, error) {
	p := path.Join(b.Options["path"].(string), s, i)
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	return bufio.NewReader(f), nil
}
