package storage_test

import (
	"bufio"
	"bytes"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/ybizeul/hupload/internal/storage"
)

func createS3Backend(t *testing.T) *storage.S3Backend {
	c := storage.S3StorageConfig{
		Endpoint:     os.Getenv("AWS_ENDPOINT_URL"),
		Region:       os.Getenv("AWS_DEFAULT_REGION"),
		AWSKey:       os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecret:    os.Getenv("AWS_SECRET_ACCESS_KEY"),
		UsePathStyle: true,
		Bucket:       os.Getenv("BUCKET"),
		MaxFileSize:  4,
		MaxShareSize: 5,
	}

	f := storage.NewS3Storage(c)
	if f == nil {
		t.Errorf("Expected S3 Storage to be created")
	}

	return f
}

func TestS3CreateShare(t *testing.T) {
	f := createS3Backend(t)

	t.Cleanup(func() {
		_ = f.DeleteShare("Test")
	})

	// Test create share
	_, err := f.CreateShare("Test", "admin", storage.DefaultOptions())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestS3UpdateShare(t *testing.T) {
	f := createS3Backend(t)

	t.Cleanup(func() {
		_ = f.DeleteShare("Test")
	})

	// Test create share
	_, err := f.CreateShare("Test", "admin", storage.DefaultOptions())
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
	options, err := f.UpdateShare("Test", newOptions)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}

	if !reflect.DeepEqual(options, newOptions) {
		t.Errorf("Expected options to be %v, got %v", options, newOptions)
	}
}

func TestCreateS3Item(t *testing.T) {
	f := createS3Backend(t)

	t.Cleanup(func() {
		_ = f.DeleteShare("Test")
	})

	_, err := f.CreateShare("Test", "admin", storage.DefaultOptions())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}

	content := []byte("test")
	b := bufio.NewReader(bytes.NewBuffer([]byte("test")))

	// Test create item
	_, err = f.CreateItem("Test", "test.txt", int64(len(content)), b)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestS3GetShare(t *testing.T) {
	f := createS3Backend(t)

	t.Cleanup(func() {
		_ = f.DeleteShare("Test")
	})

	options := storage.Options{
		Validity:    7,
		Exposure:    "upload",
		Message:     "message",
		Description: "description",
	}
	_, err := f.CreateShare("Test", "admin", options)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}

	// Test get share
	got, err := f.GetShare("Test")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}
	got.DateCreated = time.Time{}

	want := &storage.Share{
		Version: 1,
		Name:    "Test",
		Owner:   "admin",
		Options: options,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expected share to be %v, got %v", got, want)
	}
}

func TestS3ListShares(t *testing.T) {
	f := createS3Backend(t)

	t.Cleanup(func() {
		_ = f.DeleteShare("Test1")
		_ = f.DeleteShare("Test2")
	})

	var (
		err error
	)

	_, err = f.CreateShare("Test1", "admin", storage.DefaultOptions())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}
	_, err = f.CreateShare("Test2", "admin", storage.DefaultOptions())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}

	// Test list shares
	got, err := f.ListShares()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}

	for i := range got {
		got[i].DateCreated = time.Time{}
	}

	want := []storage.Share{
		{
			Version: 1,
			Name:    "Test2",
			Owner:   "admin",
			Options: storage.DefaultOptions(),
		},
		{
			Version: 1,
			Name:    "Test1",
			Owner:   "admin",
			Options: storage.DefaultOptions(),
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expected shares to be %v, got %v", got, want)
	}
}
