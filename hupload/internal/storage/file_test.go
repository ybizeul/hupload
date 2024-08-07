package storage

import (
	"os"
	"reflect"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func setupSuite() func(t *testing.T) {
	return func(t *testing.T) {
		os.RemoveAll("data")
	}
}
func TestCreateShare(t *testing.T) {
	tearDown := setupSuite()
	c := map[string]any{
		"path":         "data",
		"max_file_mb":  5,
		"max_share_mb": 10,
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
	tearDown(t)
}
