package storage

import (
	"bufio"
	"bytes"
	"reflect"
	"testing"
	"time"
)

func createS3Backend(t *testing.T) *S3Backend {
	c := S3StorageConfig{
		AWSKey:    "SJdTNdqZZgb2pOj84Xka",
		AWSSecret: "IeVcaEdKMY6SSh1conEA1eyqVX5PzgI0GIbqkxco",
		Bucket:    "hupload",
	}

	f := NewS3Storage(c)
	if f == nil {
		t.Errorf("Expected FileStorage to be created")
	}

	//f.initialize()

	return f
}

func TestS3CreateShare(t *testing.T) {
	f := createS3Backend(t)

	t.Cleanup(func() {
		_ = f.DeleteShare("Test")
	})

	// Test create share
	_, err := f.CreateShare("Test", "admin", DefaultOptions())
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
	_, err := f.CreateShare("Test", "admin", DefaultOptions())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test update share
	newOptions := &Options{
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

	_, err := f.CreateShare("Test", "admin", DefaultOptions())
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

	options := Options{
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

	want := &Share{
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

	_, err = f.CreateShare("Test1", "admin", DefaultOptions())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}
	_, err = f.CreateShare("Test2", "admin", DefaultOptions())
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

	want := []Share{
		{
			Name:    "Test1",
			Owner:   "admin",
			Options: DefaultOptions(),
		},
		{
			Name:    "Test2",
			Owner:   "admin",
			Options: DefaultOptions(),
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expected shares to be %v, got %v", got, want)
	}
}
