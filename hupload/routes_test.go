package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
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
		if token == nil {
			t.Errorf("Expected token to be created")
			return
		}
	})

	var share *storage.Share
	t.Run("Create a share with token should succeed", func(t *testing.T) {
		if token == nil {
			t.Fatal("No token created")
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

		if share == nil {
			t.Errorf("Expected token to be created")
			return
		}

		_, err := os.Stat(path.Join("tmptest/data/", share.Name))
		if err != nil {
			t.Errorf("Expected share directory to be created")
			return
		}
	})

	t.Run("Create a share with same name should fail", func(t *testing.T) {
		if share == nil {
			t.Fatal("No share created")
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

func TestUpload(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("tmptest")
	})

	api := getAPIServer(t)

	var (
		req *http.Request
		w   *httptest.ResponseRecorder
	)

	// Create upload share
	req = httptest.NewRequest("POST", "/api/v1/shares", nil)

	req.SetBasicAuth("admin", "hupload")

	w = httptest.NewRecorder()

	api.Mux.ServeHTTP(w, req)

	var uploadShare *storage.Share

	_ = json.NewDecoder(w.Body).Decode(&uploadShare)

	t.Run("Upload a file without authentication should work", func(t *testing.T) {
		body := new(bytes.Buffer)

		writer := multipart.NewWriter(body)
		// create a new form-data header name data and filename data.txt
		dataPart, _ := writer.CreateFormFile("data", "file.txt")

		// copy file content into multipart section dataPart
		f, _ := os.Open("routes_testdata/file.txt")
		_, _ = io.Copy(dataPart, f)
		writer.Close()

		req = httptest.NewRequest("POST", path.Join("/api/v1/shares", uploadShare.Name, "items", "newfile.txt"), body)

		req.Header.Set("Content-Type", writer.FormDataContentType())

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}

		stat, err := os.Stat(path.Join("tmptest/data/", uploadShare.Name, "newfile.txt"))
		if err != nil {
			t.Errorf("Expected file to be created")
			return
		}
		if stat.Size() != 4 {
			t.Errorf("Expected file size to be 4, got %d", stat.Size())
			return
		}
	})

	// Create download share
	p := ShareParameters{
		Exposure: "download",
	}

	b, _ := json.Marshal(p)

	reader := bytes.NewReader(b)

	req = httptest.NewRequest("POST", "/api/v1/shares", reader)

	req.SetBasicAuth("admin", "hupload")

	w = httptest.NewRecorder()

	api.Mux.ServeHTTP(w, req)

	var downloadShare *storage.Share

	_ = json.NewDecoder(w.Body).Decode(&downloadShare)

	t.Run("Upload a file without authentication should not work (download share)", func(t *testing.T) {
		body := new(bytes.Buffer)

		writer := multipart.NewWriter(body)
		// create a new form-data header name data and filename data.txt
		dataPart, _ := writer.CreateFormFile("data", "file.txt")

		// copy file content into multipart section dataPart
		f, _ := os.Open("routes_testdata/file.txt")
		_, _ = io.Copy(dataPart, f)
		writer.Close()

		req = httptest.NewRequest("POST", path.Join("/api/v1/shares", downloadShare.Name, "items", "newfile.txt"), body)

		req.Header.Set("Content-Type", writer.FormDataContentType())

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
			return
		}

		_, err := os.Stat(path.Join("tmptest/data/", downloadShare.Name, "newfile.txt"))
		if err == nil {
			t.Errorf("Expected file to not be created")
			return
		}
	})
}
