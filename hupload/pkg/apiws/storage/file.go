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

// FileBackend is a backend that stores files on the filesystem
type FileBackend struct {
	Options map[string]any
}

// NewFileStorage creates a new FileBackend, m is the configuration as found
// in Hupload configuration file.
func NewFileStorage(m map[string]any) *FileBackend {
	b, err := yaml.Marshal(m)
	if err != nil {
		return nil
	}

	r := &FileBackend{}

	err = yaml.Unmarshal(b, &r.Options)
	if err != nil {
		return nil
	}

	r.initialize()

	return r
}

// initialize creates the root directory for the backend
func (b *FileBackend) initialize() {
	path := b.stringOption("path")
	if path == "" {
		panic("path is required")
	}
	err := os.MkdirAll(path, 0755)
	if err != nil {
		panic(err)
	}
}

func (b *FileBackend) stringOption(o string) string {
	v, ok := b.Options[o].(string)
	if !ok {
		return ""
	}
	return v
}

func (b *FileBackend) int64Option(o string) int64 {
	v, ok := b.Options[o].(int)
	if !ok {
		return 0
	}
	return int64(v)
}

func (b *FileBackend) CreateShare(s string, o string) error {
	_, err := os.Stat(path.Join(b.stringOption("path"), s))
	if err == nil {
		return errors.New("share already exists")
	}

	p := path.Join(b.stringOption("path"), s)
	err = os.Mkdir(p, 0755)
	if err != nil {
		slog.Error("cannot create share", slog.String("error", err.Error()), slog.String("path", p))
		return errors.New("cannot create share")
	}

	m := Share{
		Name:        s,
		Owner:       o,
		DateCreated: time.Now(),
	}

	f, err := os.Create(path.Join(b.stringOption("path"), s, ".metadata"))
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
	p := path.Join(b.stringOption("path"), s, i)
	f, err := os.Create(p + suffix)
	if err != nil {
		return nil, errors.New("cannot create item")
	}
	defer f.Close()

	share, err := b.GetShare(s)

	src := r

	maxWrite := int64(0)

	maxShare := b.int64Option("max_share_mb") * 1024 * 1024
	if maxShare > 0 {
		maxWrite = maxShare - share.Size
		if maxWrite <= 0 {
			return nil, errors.New("Max share capacity already reached")
		}
	}

	maxItem := b.int64Option("max_file_mb") * 1024 * 1024
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
	fm, err := os.Open(path.Join(b.stringOption("path"), s, ".metadata"))
	if err != nil {
		return nil, err
	}
	defer fm.Close()

	m := Share{}
	err = json.NewDecoder(fm).Decode(&m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (b *FileBackend) ListShares() ([]Share, error) {
	d, err := os.ReadDir(b.stringOption("path"))
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

func (b *FileBackend) DeleteShare(s string) error {
	sharePath := path.Join(b.stringOption("path"), s)
	err := os.RemoveAll(sharePath)
	if err != nil {
		return err
	}
	return nil
}

func (b *FileBackend) ListShare(s string) ([]Item, error) {
	d, err := os.ReadDir(path.Join(b.stringOption("path"), s))
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
	p := path.Join(b.stringOption("path"), s, i)
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
	p := path.Join(b.stringOption("path"), s, i)
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	return bufio.NewReader(f), nil
}

func (b *FileBackend) UpdateMetadata(s string) error {
	sd, err := os.ReadDir(path.Join(b.stringOption("path"), s))
	if err != nil {
		return err
	}

	fm, err := os.OpenFile(path.Join(b.stringOption("path"), s, ".metadata"), os.O_RDWR, 0644)
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
