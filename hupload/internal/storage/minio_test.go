package storage_test

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"path"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/ybizeul/hupload/internal/storage"
)

func createMinioBackend(t *testing.T) *storage.MinioBackend {
	c := storage.MinioStorageConfig{
		Endpoint:     os.Getenv("MINIO_ENDPOINT"),
		Region:       os.Getenv("MINIO_DEFAULT_REGION"),
		AWSKey:       os.Getenv("MINIO_ACCESS_KEY_ID"),
		AWSSecret:    os.Getenv("MINIO_SECRET_ACCESS_KEY"),
		UsePathStyle: true,
		Bucket:       os.Getenv("BUCKET"),
		MaxFileSize:  4,
		MaxShareSize: 5,
	}

	f := storage.NewMinioStorage(c)
	if f == nil {
		t.Errorf("Expected S3 Storage to be created")
	}

	return f
}

func TestMinioCreateShare(t *testing.T) {
	f := createMinioBackend(t)

	t.Cleanup(func() {
		_ = f.DeleteShare(context.Background(), "Test")
	})

	// Test create share
	_, err := f.CreateShare(context.Background(), "Test", "admin", storage.DefaultOptions())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestMinioUpdateShare(t *testing.T) {
	f := createMinioBackend(t)

	t.Cleanup(func() {
		_ = f.DeleteShare(context.Background(), "Test")
	})

	// Test create share
	_, err := f.CreateShare(context.Background(), "Test", "admin", storage.DefaultOptions())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test update share
	newOptions := &storage.Options{
		Validity:    20,
		Exposure:    "both",
		Message:     "message",
		Description: "description",
	}
	options, err := f.UpdateShare(context.Background(), "Test", newOptions, &map[string]int64{"test.txt": 1})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}

	if !reflect.DeepEqual(options, newOptions) {
		t.Errorf("Expected options to be %v, got %v", options, newOptions)
	}
}

func TestCreateMinioItem(t *testing.T) {
	f := createMinioBackend(t)

	t.Cleanup(func() {
		_ = f.DeleteShare(context.Background(), "Test")
	})

	share, err := f.CreateShare(context.Background(), "Test", "admin", storage.DefaultOptions())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}

	tests := []struct {
		FileName  string
		Bytes     []byte
		Downloads int64
	}{
		{
			FileName:  "test.txt",
			Bytes:     []byte("test"),
			Downloads: 5,
		},
		{
			FileName:  "test2.txt",
			Bytes:     []byte(""),
			Downloads: 0,
		},
	}
	for _, test := range tests {
		b := bufio.NewReader(bytes.NewBuffer(test.Bytes))

		// Test create item
		_, err = f.CreateItem(context.Background(), "Test", test.FileName, int64(len(test.Bytes)), b)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		time.Sleep(1 * time.Second)
	}

	// Reverse tests to have the latest uploaded first
	slices.Reverse(tests)

	share.Downloads = map[string]int64{
		"test.txt":  5,
		"test2.txt": 0,
	}
	_, err = f.UpdateShare(context.Background(), "Test", &share.Options, &share.Downloads)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}

	// Check items are in share
	items, err := f.ListShare(context.Background(), "Test")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}

	if len(items) != len(tests) {
		t.Errorf("Expected %d items, got %d", len(tests), len(items))
		return
	}

	for i, item := range items {
		if item.Path != path.Join("Test", tests[i].FileName) {
			t.Errorf("Expected item name to be %s, got %s", path.Join("Test", tests[i].FileName), item.Path)
		}
		if item.ItemInfo.Size != int64(len(tests[i].Bytes)) {
			t.Errorf("Expected item size to be %d, got %d", len(tests[i].Bytes), item.ItemInfo.Size)
		}
		if item.Downloads != tests[i].Downloads {
			t.Errorf("Expected item downloads to be %d, got %d", tests[i].Downloads, item.Downloads)
		}
	}
}

func TestMinioGetShare(t *testing.T) {
	f := createMinioBackend(t)

	t.Cleanup(func() {
		_ = f.DeleteShare(context.Background(), "Test")
	})

	options := storage.Options{
		Validity:    7,
		Exposure:    "upload",
		Message:     "message",
		Description: "description",
	}
	_, err := f.CreateShare(context.Background(), "Test", "admin", options)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}

	// Test get share
	got, err := f.GetShare(context.Background(), "Test")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}
	got.DateCreated = time.Time{}

	want := &storage.Share{
		Version:   1,
		Name:      "Test",
		Owner:     "admin",
		Options:   options,
		Downloads: map[string]int64{},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expected share to be %v, got %v", got, want)
	}
}

func TestMinioListShares(t *testing.T) {
	f := createMinioBackend(t)

	t.Cleanup(func() {
		_ = f.DeleteShare(context.Background(), "Test1")
		_ = f.DeleteShare(context.Background(), "Test2")
	})

	var (
		err error
	)

	_, err = f.CreateShare(context.Background(), "Test1", "admin", storage.DefaultOptions())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}
	time.Sleep(1 * time.Second)
	_, err = f.CreateShare(context.Background(), "Test2", "admin", storage.DefaultOptions())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}

	// Test list shares
	got, err := f.ListShares(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}

	for i := range got {
		got[i].DateCreated = time.Time{}
	}

	want := []storage.Share{
		*storage.NewShare().WithName("Test2").WithOwner("admin").WithOptions(storage.DefaultOptions()),
		*storage.NewShare().WithName("Test1").WithOwner("admin").WithOptions(storage.DefaultOptions()),
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expected shares to be %v, got %v", got, want)
	}
}
