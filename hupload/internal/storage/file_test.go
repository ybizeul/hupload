package storage

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"reflect"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func createFileBackend(t *testing.T) *FileBackend {
	c := FileStorageConfig{
		Path: "data",
	}

	f := NewFileStorage(c)
	if f == nil {
		t.Errorf("Expected FileStorage to be created")
	}

	f.initialize()

	return f
}
func TestCreateShare(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("data")
	})

	f := createFileBackend(t)

	tests := []struct {
		f      func() (*Share, error)
		expect Share
	}{
		{
			func() (*Share, error) {
				return f.CreateShare("test", "admin", Options{Validity: 10, Exposure: "upload"})
			},
			Share{
				Name:    "test",
				Owner:   "admin",
				Options: Options{Validity: 10, Exposure: "upload"},
			},
		},
		{
			func() (*Share, error) {
				return f.CreateShare("test", "admin", Options{Validity: 10, Exposure: "both"})
			},
			Share{
				Name:    "test",
				Owner:   "admin",
				Options: Options{Validity: 10, Exposure: "both"},
			},
		},
		{
			func() (*Share, error) {
				return f.CreateShare("test", "admin", Options{Validity: 10, Exposure: "download"})
			},
			Share{
				Name:    "test",
				Owner:   "admin",
				Options: Options{Validity: 10, Exposure: "download"},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Create share %+v", test.expect), func(t *testing.T) {

			share, err := test.f()
			share.DateCreated = time.Time{}
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			_, err = os.Stat(path.Join("data", share.Name))
			if err != nil {
				t.Errorf("Expected share directory to be created")
			}

			metadata_f, err := os.Open(path.Join("data", share.Name, ".metadata"))
			if err != nil {
				t.Errorf("Expected metadata to be written")
			}

			var got Share
			err = yaml.NewDecoder(metadata_f).Decode(&got)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			got.DateCreated = time.Time{}
			if !reflect.DeepEqual(&test.expect, &got) {
				t.Errorf("Expected %v, got %v", share, got)
			}
			os.RemoveAll(path.Join("data", share.Name))
		})
	}
}

func TestCreateItem(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("data")
	})

	f := createFileBackend(t)

	share, err := f.CreateShare("test", "admin", Options{Validity: 10, Exposure: "upload"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	reader := bufio.NewReader(bytes.NewReader([]byte("test")))
	item, err := f.CreateItem(share.Name, "test.txt", reader)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}

	// Test item result
	if item.ItemInfo.Size != 4 {
		t.Errorf("Expected 4, got %v", item.ItemInfo.Size)
		return
	}

	// Test file on disk
	content, _ := os.ReadFile("data/test/test.txt")
	if !bytes.Equal(content, []byte("test")) {
		t.Errorf("Expected test, got %v", string(content))
		return
	}
}

func TestDeleteItem(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("data")
	})

	f := createFileBackend(t)

	share, _ := f.CreateShare("test", "admin", Options{Validity: 10, Exposure: "upload"})

	t.Run("Delete inexistant item should fail", func(t *testing.T) {
		err := f.DeleteItem(share.Name, "test.txt")
		if err != ErrItemNotFound {
			t.Errorf("Expected ErrItemNotFound, got %v", err)
		}
	})

	reader := bufio.NewReader(bytes.NewReader([]byte("test")))
	_, _ = f.CreateItem(share.Name, "test.txt", reader)

	err := f.DeleteItem(share.Name, "test.txt")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	_, err = os.Stat(path.Join("data", share.Name, "test.txt"))
	if err == nil {
		t.Errorf("Expected item to be deleted")
	}
}
func TestDeleteShare(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("data")
	})

	c := FileStorageConfig{
		Path:         "data",
		MaxFileSize:  5,
		MaxShareSize: 10,
	}

	f := NewFileStorage(c)
	if f == nil {
		t.Errorf("Expected FileStorage to be created")
	}

	f.initialize()

	share, _ := f.CreateShare("test", "admin", Options{Validity: 10, Exposure: "upload"})

	err := f.DeleteShare(share.Name)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	_, err = os.Stat("data/test")
	if err == nil {
		t.Errorf("Expected share directory to be deleted")
	}
}

func TestFileOverflow(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("data")
		os.RemoveAll("2mb")
	})

	c := FileStorageConfig{
		Path:         "data",
		MaxFileSize:  1,
		MaxShareSize: 3,
	}

	f := NewFileStorage(c)
	if f == nil {
		t.Errorf("Expected FileStorage to be created")
	}

	f.initialize()

	share, err := f.CreateShare("test", "admin", Options{Validity: 10, Exposure: "upload"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	createFile("2mb", 2)

	r, _ := os.Open("2mb")
	reader := bufio.NewReader(r)
	_, err = f.CreateItem(share.Name, "test.txt", reader)
	defer r.Close()

	if !errors.Is(err, ErrMaxShareSizeReached) {
		t.Errorf("Expected ErrMaxShareSizeReached, got %v", err)
	}
}

func TestShareOverflow(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("data")
		os.RemoveAll("3mb")
	})

	c := FileStorageConfig{
		Path:         "data",
		MaxFileSize:  4,
		MaxShareSize: 5,
	}

	f := NewFileStorage(c)
	if f == nil {
		t.Errorf("Expected FileStorage to be created")
	}

	f.initialize()

	share, err := f.CreateShare("test", "admin", Options{Validity: 10, Exposure: "upload"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	createFile("3mb", 3)

	r, _ := os.Open("3mb")
	defer r.Close()

	reader := bufio.NewReader(r)
	_, err = f.CreateItem(share.Name, "test.txt", reader)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	s, _ := os.Open("3mb")
	defer s.Close()

	reader = bufio.NewReader(s)
	_, err = f.CreateItem(share.Name, "test2.txt", reader)

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestSafeShareName(t *testing.T) {
	if !isShareNameSafe("test") {
		t.Errorf("Expected true, got false")
	}

	if isShareNameSafe("Test/path") {
		t.Errorf("Expected true, got false")
	}

	if isShareNameSafe("share/../path") {
		t.Errorf("Expected true, got false")
	}
}

func createFile(path string, size int) {
	s := int64(size * 1024 * 1024)
	fd, err := os.Create(path)
	if err != nil {
		log.Fatal("Failed to create output")
	}
	_, err = fd.Seek(s-1, 0)
	if err != nil {
		log.Fatal("Failed to seek")
	}
	_, err = fd.Write([]byte{0})
	if err != nil {
		log.Fatal("Write failed")
	}
	err = fd.Close()
	if err != nil {
		log.Fatal("Failed to close file")
	}
}

func TestGetItemData(t *testing.T) {
	c := FileStorageConfig{
		Path:         "file_testdata/data",
		MaxFileSize:  1,
		MaxShareSize: 2,
	}

	f := NewFileStorage(c)
	if f == nil {
		t.Errorf("Expected FileStorage to be created")
	}

	f.initialize()

	r, err := f.GetItemData("test", "test.txt")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	b, err := io.ReadAll(r)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	r.Close()

	if !bytes.Equal(b, []byte("test")) {
		t.Errorf("Expected test, got %v", string(b))
	}
}

func TestGetShare(t *testing.T) {
	c := FileStorageConfig{
		Path:         "file_testdata/data",
		MaxFileSize:  1,
		MaxShareSize: 2,
	}

	f := NewFileStorage(c)

	share, err := f.GetShare("test")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	parsedTime, err := time.Parse("2006-01-02T15:04:05.99-07:00", "2024-08-08T16:20:25.231034+02:00")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	want := Share{
		Name:        "test",
		Owner:       "admin",
		Options:     Options{Validity: 10, Exposure: "upload"},
		Size:        4,
		Count:       1,
		DateCreated: parsedTime,
	}

	if !reflect.DeepEqual(*share, want) {
		t.Errorf("Expected %v, got %v", want, share)
	}
}

func TestListShare(t *testing.T) {
	c := FileStorageConfig{
		Path:         "file_testdata/data",
		MaxFileSize:  1,
		MaxShareSize: 2,
	}

	f := NewFileStorage(c)

	items, err := f.ListShare("test")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	items[0].ItemInfo.DateModified = time.Time{}

	want := []Item{
		{
			Path: "test/test.txt",
			ItemInfo: ItemInfo{
				Size:         4,
				DateModified: time.Time{},
			},
		},
	}

	if !reflect.DeepEqual(items, want) {
		t.Errorf("Expected %v, got %v", want, items)
	}
}

func TestListShares(t *testing.T) {
	c := FileStorageConfig{
		Path:         "file_testdata/data",
		MaxFileSize:  1,
		MaxShareSize: 2,
	}

	f := NewFileStorage(c)

	shares, err := f.ListShares()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	parsedTime, err := time.Parse("2006-01-02T15:04:05.99-07:00", "2024-08-08T16:20:25.231034+02:00")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	want := []Share{
		{
			Name:        "test",
			Owner:       "admin",
			Options:     Options{Validity: 10, Exposure: "upload"},
			Size:        4,
			Count:       1,
			DateCreated: parsedTime,
		},
		{
			Name:        "test2",
			Owner:       "admin",
			Options:     Options{Validity: 10, Exposure: "upload"},
			Size:        4,
			Count:       1,
			DateCreated: parsedTime,
		},
	}

	if !reflect.DeepEqual(shares, want) {
		t.Errorf("Expected %v, got %v", want, shares)
	}
}

func TestMigrate(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("file_testdata/data_old_copy")
	})

	cmd := exec.Command("cp", "-r", "file_testdata/data_old", "file_testdata/data_old_copy")
	err := cmd.Run()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	c := FileStorageConfig{
		Path:         "file_testdata/data_old_copy",
		MaxFileSize:  1,
		MaxShareSize: 2,
	}

	f := NewFileStorage(c)

	err = f.Migrate()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	share, _ := f.GetShare("test")

	parsedTime, _ := time.Parse("2006-01-02T15:04:05.99-07:00", "2024-08-08T16:20:25.231034+02:00")

	expect := Share{
		Version:     1,
		Name:        "test",
		Owner:       "admin",
		Options:     Options{Validity: 10},
		Size:        4,
		Count:       1,
		DateCreated: parsedTime,
	}

	if !reflect.DeepEqual(*share, expect) {
		t.Errorf("Expected %v, got %v", expect, share)
	}

	share, _ = f.GetShare("test2")

	expect = Share{
		Version:     1,
		Name:        "test2",
		Owner:       "admin",
		Options:     Options{Validity: 10},
		Size:        4,
		Count:       1,
		DateCreated: parsedTime,
	}

	if !reflect.DeepEqual(*share, expect) {
		t.Errorf("Expected %v, got %v", expect, share)
	}
}
