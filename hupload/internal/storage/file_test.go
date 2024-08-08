package storage

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestCreateShare(t *testing.T) {
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

	err := f.CreateShare("test", "admin", 10)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Ch
	_, err = os.Stat("data/test")
	if err != nil {
		t.Errorf("Expected share directory to be created")
	}

	metadata_f, err := os.Open("data/test/.metadata")

	if err != nil {
		t.Errorf("Expected metadata to be written")
	}

	expect := Share{
		Name:     "test",
		Owner:    "admin",
		Validity: 10,
	}

	var got Share
	err = yaml.NewDecoder(metadata_f).Decode(&got)

	if err != nil {
		t.Errorf("Expected metadata to be decoded")
	}

	got.DateCreated = time.Time{}
	if !reflect.DeepEqual(expect, got) {
		t.Errorf("Expected %v, got %v", expect, got)
	}
}

func TestCreateItem(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("data")
	})

	c := FileStorageConfig{
		Path:         "data",
		MaxFileSize:  1,
		MaxShareSize: 2,
	}

	f := NewFileStorage(c)
	if f == nil {
		t.Errorf("Expected FileStorage to be created")
	}

	f.initialize()

	err := f.CreateShare("test", "admin", 10)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	reader := bufio.NewReader(bytes.NewReader([]byte("test")))
	item, err := f.CreateItem("test", "test.txt", reader)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test item result
	if item.ItemInfo.Size != 4 {
		t.Errorf("Expected 4, got %v", item.ItemInfo.Size)
	}

	// Test file on disk
	content, _ := os.ReadFile("data/test/test.txt")
	if !bytes.Equal(content, []byte("test")) {
		t.Errorf("Expected test, got %v", string(content))
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

	_ = f.CreateShare("test", "admin", 10)

	err := f.DeleteShare("test")
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

	err := f.CreateShare("test", "admin", 10)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	createFile("2mb", 2)

	r, _ := os.Open("2mb")
	reader := bufio.NewReader(r)
	_, err = f.CreateItem("test", "test.txt", reader)
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

	err := f.CreateShare("test", "admin", 10)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	createFile("3mb", 3)

	r, _ := os.Open("3mb")
	defer r.Close()

	reader := bufio.NewReader(r)
	_, err = f.CreateItem("test", "test.txt", reader)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	s, _ := os.Open("3mb")
	defer s.Close()

	reader = bufio.NewReader(s)
	_, err = f.CreateItem("test", "test2.txt", reader)

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
		Validity:    10,
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
			Validity:    10,
			Size:        4,
			Count:       1,
			DateCreated: parsedTime,
		},
		{
			Name:        "test2",
			Owner:       "admin",
			Validity:    10,
			Size:        4,
			Count:       1,
			DateCreated: parsedTime,
		},
	}

	if !reflect.DeepEqual(shares, want) {
		t.Errorf("Expected %v, got %v", want, shares)
	}
}
