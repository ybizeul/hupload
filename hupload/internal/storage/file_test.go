package storage_test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/ybizeul/hupload/internal/storage"
	"gopkg.in/yaml.v3"
)

func createFileBackend(t *testing.T) *storage.FileBackend {
	c := storage.FileStorageConfig{
		Path:         "data",
		MaxFileSize:  4,
		MaxShareSize: 5,
	}

	f := storage.NewFileStorage(c)
	if f == nil {
		t.Errorf("Expected FileStorage to be created")
	}

	return f
}
func TestCreateShare(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("data")
	})

	f := createFileBackend(t)

	tests := []struct {
		f      func() (*storage.Share, error)
		expect storage.Share
	}{
		{
			func() (*storage.Share, error) {
				return f.CreateShare(context.Background(), "test", "admin", storage.Options{Validity: 10, Exposure: "upload"})
			},
			storage.Share{
				Version: 1,
				Name:    "test",
				Owner:   "admin",
				Options: storage.Options{Validity: 10, Exposure: "upload"},
			},
		},
		{
			func() (*storage.Share, error) {
				return f.CreateShare(context.Background(), "test", "admin", storage.Options{Validity: 10, Exposure: "both"})
			},
			storage.Share{
				Version: 1,
				Name:    "test",
				Owner:   "admin",
				Options: storage.Options{Validity: 10, Exposure: "both"},
			},
		},
		{
			func() (*storage.Share, error) {
				return f.CreateShare(context.Background(), "test", "admin", storage.Options{Validity: 10, Exposure: "download"})
			},
			storage.Share{
				Version: 1,
				Name:    "test",
				Owner:   "admin",
				Options: storage.Options{Validity: 10, Exposure: "download"},
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

			var got storage.Share
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

func TestUpdateShare(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("data")
	})

	f := createFileBackend(t)

	share, _ := f.CreateShare(context.Background(), "test", "admin", storage.Options{Validity: 10, Exposure: "upload", Description: "description", Message: "message"})

	newOptions := &storage.Options{
		Validity:    20,
		Exposure:    "both",
		Description: "new description",
		Message:     "new message",
	}

	o, err := f.UpdateShare(context.Background(), share.Name, newOptions)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if reflect.DeepEqual(o, newOptions) == false {
		t.Errorf("Expected %v, got %v", newOptions, o)
	}

	share.Options = *newOptions
	share2, _ := f.GetShare(context.Background(), share.Name)

	if !share.DateCreated.Equal(share2.DateCreated) {
		t.Errorf("Expected %v, got %v", share.DateCreated, share2.DateCreated)
	}

	share.DateCreated = time.Time{}
	share2.DateCreated = time.Time{}

	if reflect.DeepEqual(share, share2) == false {
		t.Errorf("Expected %+v, got %+v", share, share2)
	}
}
func TestCreateItem(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("data")
	})

	f := createFileBackend(t)

	share, err := f.CreateShare(context.Background(), "test", "admin", storage.Options{Validity: 10, Exposure: "upload"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	tests := []struct {
		FileName string
		Bytes    []byte
	}{
		{
			FileName: "test.txt",
			Bytes:    []byte("test"),
		},
		{
			FileName: "test2.txt",
			Bytes:    []byte(""),
		},
	}

	for _, test := range tests {
		reader := bufio.NewReader(bytes.NewReader(test.Bytes))
		item, err := f.CreateItem(context.Background(), share.Name, test.FileName, 0, reader)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
			return
		}

		// Test item result
		if item.ItemInfo.Size != int64(len(test.Bytes)) {
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
}

func TestDeleteItem(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("data")
	})

	f := createFileBackend(t)

	share, _ := f.CreateShare(context.Background(), "test", "admin", storage.Options{Validity: 10, Exposure: "upload"})

	t.Run("Delete inexistant item should fail", func(t *testing.T) {
		err := f.DeleteItem(context.Background(), share.Name, "test.txt")
		if err != storage.ErrItemNotFound {
			t.Errorf("Expected ErrItemNotFound, got %v", err)
		}
	})

	reader := bufio.NewReader(bytes.NewReader([]byte("test")))
	_, _ = f.CreateItem(context.Background(), share.Name, "test.txt", 0, reader)

	err := f.DeleteItem(context.Background(), share.Name, "test.txt")
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

	c := storage.FileStorageConfig{
		Path:         "data",
		MaxFileSize:  5,
		MaxShareSize: 10,
	}

	f := storage.NewFileStorage(c)
	if f == nil {
		t.Errorf("Expected FileStorage to be created")
	}

	share, _ := f.CreateShare(context.Background(), "test", "admin", storage.Options{Validity: 10, Exposure: "upload"})

	err := f.DeleteShare(context.Background(), share.Name)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	_, err = os.Stat("data/test")
	if err == nil {
		t.Errorf("Expected share directory to be deleted")
	}
}

func TestSafeShareName(t *testing.T) {
	if !storage.IsShareNameSafe("test") {
		t.Errorf("Expected true, got false")
	}

	if storage.IsShareNameSafe("Test/path") {
		t.Errorf("Expected true, got false")
	}

	if storage.IsShareNameSafe("share/../path") {
		t.Errorf("Expected true, got false")
	}
}

func TestGetItemData(t *testing.T) {
	c := storage.FileStorageConfig{
		Path:         "file_testdata/data",
		MaxFileSize:  1,
		MaxShareSize: 2,
	}

	f := storage.NewFileStorage(c)
	if f == nil {
		t.Errorf("Expected FileStorage to be created")
	}

	r, err := f.GetItemData(context.Background(), "test", "test.txt")
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
	c := storage.FileStorageConfig{
		Path:         "file_testdata/data",
		MaxFileSize:  1,
		MaxShareSize: 2,
	}

	f := storage.NewFileStorage(c)

	share, err := f.GetShare(context.Background(), "test")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	parsedTime, err := time.Parse("2006-01-02T15:04:05.99-07:00", "2024-08-08T16:20:25.231034+02:00")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	want := storage.Share{
		Name:        "test",
		Owner:       "admin",
		Options:     storage.Options{Validity: 10, Exposure: "upload", Description: "description", Message: "message"},
		Size:        4,
		Count:       1,
		DateCreated: parsedTime,
	}

	if !reflect.DeepEqual(*share, want) {
		t.Errorf("Expected %v, got %v", want, share)
	}
}

func TestListShare(t *testing.T) {
	c := storage.FileStorageConfig{
		Path:         "file_testdata/data",
		MaxFileSize:  1,
		MaxShareSize: 2,
	}

	f := storage.NewFileStorage(c)

	items, err := f.ListShare(context.Background(), "test")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	items[0].ItemInfo.DateModified = time.Time{}

	want := []storage.Item{
		{
			Path: "test/test.txt",
			ItemInfo: storage.ItemInfo{
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
	c := storage.FileStorageConfig{
		Path:         "file_testdata/data",
		MaxFileSize:  1,
		MaxShareSize: 2,
	}

	f := storage.NewFileStorage(c)

	shares, err := f.ListShares(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	parsedTime, err := time.Parse("2006-01-02T15:04:05.99-07:00", "2024-08-08T16:20:25.231034+02:00")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	want := []storage.Share{
		{
			Name:        "test",
			Owner:       "admin",
			Options:     storage.Options{Validity: 10, Exposure: "upload", Description: "description", Message: "message"},
			Size:        4,
			Count:       1,
			DateCreated: parsedTime,
		},
		{
			Name:        "test2",
			Owner:       "admin",
			Options:     storage.Options{Validity: 10, Exposure: "upload"},
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

	c := storage.FileStorageConfig{
		Path:         "file_testdata/data_old_copy",
		MaxFileSize:  1,
		MaxShareSize: 2,
	}

	f := storage.NewFileStorage(c)

	err = f.Migrate()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	parsedTime, _ := time.Parse("2006-01-02T15:04:05.99-07:00", "2024-08-08T16:20:25.231034+02:00")

	tests := []struct {
		name string
		want storage.Share
	}{
		// Migration of a v0 share
		{
			name: "test",
			want: storage.Share{
				Version:     1,
				Name:        "test",
				Owner:       "admin",
				Options:     storage.Options{Validity: 10},
				Size:        4,
				Count:       1,
				DateCreated: parsedTime,
			},
		},

		// Migration of a v1 share
		{
			name: "test2",
			want: storage.Share{
				Version:     1,
				Name:        "test2",
				Owner:       "admin",
				Options:     storage.Options{Validity: 10},
				Size:        4,
				Count:       1,
				DateCreated: parsedTime,
			},
		},

		// Migration of a v0 share that is actually a v1 share
		{
			name: "test3",
			want: storage.Share{
				Version: 1,
				Name:    "test3",
				Owner:   "admin",
				Options: storage.Options{
					Validity:    10,
					Exposure:    "both",
					Description: "desc",
					Message:     "message",
				},
				Size:        4,
				Count:       1,
				DateCreated: parsedTime,
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Get share %s", test.name), func(t *testing.T) {
			share, _ := f.GetShare(context.Background(), test.name)

			if !reflect.DeepEqual(*share, test.want) {
				t.Errorf("Expected %v, got %v", test.want, share)
			}
		})
	}
}

func TestShareWithDescriptionAndMessage(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("datadescription")
	})

	c := storage.FileStorageConfig{
		Path:         "datadescription",
		MaxFileSize:  1,
		MaxShareSize: 2,
	}

	f := storage.NewFileStorage(c)
	if f == nil {
		t.Errorf("Expected FileStorage to be created")
	}

	share, err := f.CreateShare(context.Background(), "test", "admin", storage.Options{Validity: 10, Exposure: "upload", Description: "test description", Message: "test message"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	metadata_f, err := os.Open(path.Join("datadescription", share.Name, ".metadata"))
	if err != nil {
		t.Errorf("Expected metadata to be written")
	}

	var got storage.Share
	err = yaml.NewDecoder(metadata_f).Decode(&got)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if got.Options.Description != "test description" {
		t.Errorf("Expected test description, got %v", got.Options.Description)
	}
	if got.Options.Message != "test message" {
		t.Errorf("Expected test message, got %v", got.Options.Message)
	}
}
