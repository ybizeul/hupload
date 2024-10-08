package storage_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/ybizeul/hupload/internal/storage"
)

func TestShouldBeValid(t *testing.T) {
	share := storage.Share{
		Name:        "test",
		Owner:       "admin",
		Options:     storage.Options{Validity: 10},
		DateCreated: time.Now().AddDate(0, 0, -5),
	}

	if share.IsValid() == false {
		t.Errorf("Expected share to be valid")
	}

	share.Options.Validity = 0

	if share.IsValid() == false {
		t.Errorf("Expected share to be valid")
	}
}

func TestShouldBeInvalid(t *testing.T) {
	share := storage.Share{
		Name:        "test",
		Owner:       "admin",
		Options:     storage.Options{Validity: 10},
		DateCreated: time.Now().AddDate(0, 0, -12),
	}

	if share.IsValid() == true {
		t.Errorf("Expected share to be invalid")
	}
}

func TestPublicShare(t *testing.T) {
	share := storage.Share{
		Name:  "test",
		Owner: "admin",
		Options: storage.Options{
			Validity:    10,
			Exposure:    "upload",
			Description: "test",
			Message:     "test",
		},
		DateCreated: time.Now(),
	}

	publicShare := share.PublicShare()

	want := &storage.PublicShare{
		Name: "test",
		Options: storage.PublicOptions{
			Exposure: "upload",
			Message:  "test",
		},
	}
	if reflect.DeepEqual(publicShare, want) == false {
		t.Errorf("Expected public share to be %v, got %v", want, publicShare)
	}
}
