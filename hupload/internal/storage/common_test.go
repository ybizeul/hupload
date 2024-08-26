package storage

import (
	"errors"
	"io"
	"os"
	"testing"
)

func readerForCapacity(capacity int) io.ReadCloser {
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		b := 1024
		w := 0
		for w < capacity {
			if w+b > capacity {
				b = capacity - w
			}
			_, _ = pw.Write(make([]byte, b))
			w += b
		}
	}()

	return pr
}

func TestFileOverflow(t *testing.T) {
	storages := []Storage{
		createFileBackend(t),
		createS3Backend(t),
	}

	t.Cleanup(func() {
		for i, _ := range storages {
			_ = storages[i].DeleteShare("test")
		}
	})

	for i := range storages {
		storage := storages[i]

		share, err := storage.CreateShare("test", "admin", Options{Validity: 10, Exposure: "upload"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		t.Run("Uploading a file too big should fail", func(t *testing.T) {
			fileSize := 5 * 1024 * 1024

			reader := readerForCapacity(fileSize)

			_, err = storage.CreateItem(share.Name, "test.txt", int64(fileSize), reader)
			reader.Close()

			if !errors.Is(err, ErrMaxFileSizeReached) {
				t.Errorf("Expected ErrMaxFileSizeReached, got %v", err)
			}
		})

		t.Run("Uploading a smaller file should succeed", func(t *testing.T) {
			fileSize := 3 * 1024 * 1024

			reader := readerForCapacity(fileSize)

			_, err = storage.CreateItem(share.Name, "test.txt", int64(fileSize), reader)
			reader.Close()

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})

		t.Run("Uploading another file should fail", func(t *testing.T) {
			fileSize := 3 * 1024 * 1024

			reader := readerForCapacity(fileSize)

			_, err = storage.CreateItem(share.Name, "test.txt", int64(fileSize), reader)
			reader.Close()

			if !errors.Is(err, ErrMaxShareSizeReached) {
				t.Errorf("Expected ErrMaxShareSizeReached, got %v", err)
			}
		})

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

	reader1 := readerForCapacity(3 * 1024 * 1024)
	defer reader1.Close()

	_, err = f.CreateItem(share.Name, "test.txt", 0, reader1)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	reader2 := readerForCapacity(3 * 1024 * 1024)
	defer reader2.Close()

	_, err = f.CreateItem(share.Name, "test2.txt", 0, reader2)

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
