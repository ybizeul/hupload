package storage

import (
	"testing"
	"time"
)

func TestShouldBeValid(t *testing.T) {
	share := Share{
		Name:        "test",
		Owner:       "admin",
		Validity:    10,
		DateCreated: time.Now().AddDate(0, 0, -5),
	}

	if share.IsValid() == false {
		t.Errorf("Expected share to be valid")
	}

	share.Validity = 0

	if share.IsValid() == false {
		t.Errorf("Expected share to be valid")
	}
}

func TestShouldBeInvalid(t *testing.T) {
	share := Share{
		Name:        "test",
		Owner:       "admin",
		Validity:    10,
		DateCreated: time.Now().AddDate(0, 0, -12),
	}

	if share.IsValid() == true {
		t.Errorf("Expected share to be invalid")
	}
}
