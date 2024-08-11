package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/ybizeul/hupload/internal/config"
	"github.com/ybizeul/hupload/internal/storage"
	"github.com/ybizeul/hupload/pkg/apiws"
)

func getAPIServer(t *testing.T) *apiws.APIWS {

	cfg = config.Config{
		Path: "routes_testdata/config.yml",
	}

	_, err := cfg.Load()

	if err != nil {
		t.Fatal(err)
		return nil
	}

	result, _ := apiws.New(nil, cfg.Values)

	result.SetAuthentication(cfg.Authentication)

	setup(result)
	return result
}

func TestCreateShare(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("tmptest")
	})

	api := getAPIServer(t)

	// Create a share without authentication should fail
	t.Run("Create a share without authentication should fail", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)
		req = httptest.NewRequest("POST", "/api/v1/shares", nil)

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
			return
		}
	})

	// Create a share with authentication should work
	var token *string
	t.Run("Create a share with authentication should succeed", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)
		req = httptest.NewRequest("POST", "/api/v1/shares", nil)

		req.SetBasicAuth("admin", "hupload")

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}

		var share *storage.Share

		_ = json.NewDecoder(w.Body).Decode(&share)

		_, err := os.Stat(path.Join("tmptest/data/", share.Name))
		if err != nil {
			t.Errorf("Expected share directory to be created")
			return
		}

		token = &w.Result().Cookies()[0].Value
	})

	var share *storage.Share
	t.Run("Create a share with token should succeed", func(t *testing.T) {
		if token == nil {
			t.Skip("No token created")
		}
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)
		req = httptest.NewRequest("POST", "/api/v1/shares", nil)

		req.AddCookie(&http.Cookie{
			Name:  "X-Token",
			Value: *token,
		})

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}

		_ = json.NewDecoder(w.Body).Decode(&share)

		_, err := os.Stat(path.Join("tmptest/data/", share.Name))
		if err != nil {
			t.Errorf("Expected share directory to be created")
			return
		}
	})

	t.Run("Create a share with same name should fail", func(t *testing.T) {
		if share == nil {
			t.Skip("No share created")
		}
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)
		req = httptest.NewRequest("PUT", path.Join("/api/v1/shares", share.Name), nil)

		req.AddCookie(&http.Cookie{
			Name:  "X-Token",
			Value: *token,
		})

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}
	})
}
