package main

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func makeShare(name string, params ShareParameters) *storage.Share {
	share, err := cfg.Storage.CreateShare(name, "admin", params.Validity, params.Exposure)
	if err != nil {
		return nil
	}
	return share
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

	t.Run("Create a share with sepcific name should succeed", func(t *testing.T) {
		if share == nil {
			t.Fatal("No share created")
		}
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)
		req = httptest.NewRequest("PUT", path.Join("/api/v1/shares", "randomname"), nil)

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

		_, err := os.Stat(path.Join("tmptest/data/", "randomname"))
		if err != nil {
			t.Errorf("Expected share directory to be created")
			return
		}
	})

	t.Run("Create a share with invalid name should fail", func(t *testing.T) {
		if share == nil {
			t.Fatal("No share created")
		}
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)
		req = httptest.NewRequest("PUT", path.Join("/api/v1/shares", url.QueryEscape("../test")), nil)

		req.AddCookie(&http.Cookie{
			Name:  "X-Token",
			Value: *token,
		})

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			return
		}
	})
}

func TestGetShare(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("tmptest")
	})

	api := getAPIServer(t)

	makeShare("test", ShareParameters{
		Exposure: "upload",
		Validity: 7,
	})

	makeShare("test2", ShareParameters{
		Exposure: "upload",
		Validity: 7,
	})

	t.Run("Get shares should succeed", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)
		req = httptest.NewRequest("GET", "/api/v1/shares", nil)
		req.SetBasicAuth("admin", "hupload")
		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}

		var shares []*storage.Share

		_ = json.NewDecoder(w.Body).Decode(&shares)

		if len(shares) != 2 {
			t.Errorf("Expected 2 shares, got %d", len(shares))
			return
		}
	})

	t.Run("Get shares without authentication should fail", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)
		req = httptest.NewRequest("GET", "/api/v1/shares", nil)

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
			return
		}
	})

	t.Run("Get share should succeed", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)
		req = httptest.NewRequest("GET", path.Join("/api/v1/shares/", "test"), nil)
		req.SetBasicAuth("admin", "hupload")
		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}

		var share *storage.Share

		_ = json.NewDecoder(w.Body).Decode(&share)

		if share.Name != "test" || share.Validity != 7 {
			t.Errorf("Share does not match created one")
		}
	})

	t.Run("Get share without authentication should succeed", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)
		req = httptest.NewRequest("GET", path.Join("/api/v1/shares/", "test"), nil)

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}

		var share *storage.Share

		_ = json.NewDecoder(w.Body).Decode(&share)

		if share.Name != "test" || share.Validity != 7 {
			t.Errorf("Share does not match created one")
		}
	})

	t.Run("Get share with invalid name should fail", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)
		req = httptest.NewRequest("GET", path.Join("/api/v1/shares/", url.QueryEscape("../test")), nil)

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}

		var result APIResult

		_ = json.NewDecoder(w.Body).Decode(&result)

		if result.Status != "error" {
			t.Errorf("Expected error, got %s", result.Status)
		}
	})

	t.Run("Get inexistant share should fail", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)
		req = httptest.NewRequest("GET", path.Join("/api/v1/shares/", "nonexistant"), nil)

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
			return
		}

		var result APIResult

		_ = json.NewDecoder(w.Body).Decode(&result)

		if result.Status != "error" {
			t.Errorf("Expected error, got %s", result.Status)
		}
	})
}

// multipartWriter returns a reader and a multipart.Writer for a body with one
// attachment of size size.
func multipartWriter(size int) (io.Reader, *multipart.Writer) {
	// writer will be returned to the caller
	var writer *multipart.Writer

	// multipart body will be piped from multipart.Writer to pr
	// pr will be returned to the caller
	pr, pw := io.Pipe()

	// we need to synchronize and wait for writer to be created
	wait := make(chan struct{})

	// we need a go function here to avoid multipart.NewWriter(pw) bloking
	go func() {
		// data will be fed by chunks of 1024 bytes
		chunk := make([]byte, 1024)

		// create a new multipart.Writer
		writer = multipart.NewWriter(pw)

		// signal that writer is created so parent function can return
		wait <- struct{}{}

		// Create the form file
		dataPart, _ := writer.CreateFormFile("data", "file.txt")

		// write size bytes to dataart
		for i := 0; i < size; i += 1024 {
			_, _ = dataPart.Write(chunk)
		}

		// close the writer
		writer.Close()
	}()
	<-wait

	// return to caller
	return pr, writer
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

	t.Run("Upload a file without authentication should work", func(t *testing.T) {
		// Create upload share
		makeShare("upload", ShareParameters{
			Exposure: "upload",
			Validity: 7,
		})

		fileSize := 1 * 1024 * 1024

		pr, writer := multipartWriter(fileSize)

		req = httptest.NewRequest("POST", path.Join("/api/v1/shares", "upload", "items", "newfile.txt"), pr)

		req.Header.Set("Content-Type", writer.FormDataContentType())

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}

		stat, err := os.Stat(path.Join("tmptest/data/", "upload", "newfile.txt"))
		if err != nil {
			t.Errorf("Expected file to be created")
			return
		}
		if stat.Size() != int64(fileSize) {
			t.Errorf("Expected file size to be %d, got %d", fileSize, stat.Size())
			return
		}
	})

	t.Run("Upload a file without authentication should not work (download share)", func(t *testing.T) {
		makeShare("download", ShareParameters{
			Exposure: "download",
			Validity: 7,
		})

		// body := new(bytes.Buffer)

		// writer := multipart.NewWriter(body)
		// // create a new form-data header name data and filename data.txt
		// dataPart, _ := writer.CreateFormFile("data", "file.txt")

		// fileSize := 3 * 1024 * 1024

		// finished := make(chan bool)
		// go func() {
		// 	_ = writeData(dataPart, fileSize)
		// 	writer.Close()
		// 	finished <- true
		// }()
		// <-finished

		fileSize := 3 * 1024 * 1024
		pr, writer := multipartWriter(fileSize)

		req = httptest.NewRequest("POST", path.Join("/api/v1/shares", "download", "items", "newfile.txt"), pr)

		req.Header.Set("Content-Type", writer.FormDataContentType())

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
			return
		}

		_, err := os.Stat(path.Join("tmptest/data/", "download", "newfile.txt"))
		if err == nil {
			t.Errorf("Expected file to not be created")
			return
		}
	})

	t.Run("Upload a file too big should not work", func(t *testing.T) {
		makeShare("toobig", ShareParameters{
			Exposure: "upload",
			Validity: 7,
		})

		// writer := multipart.NewWriter(body)
		// // create a new form-data header name data and filename data.txt
		// dataPart, _ := writer.CreateFormFile("data", "file.txt")

		fileSize := 3*1024*1024 + 1
		pr, writer := multipartWriter(fileSize)

		req = httptest.NewRequest("POST", path.Join("/api/v1/shares", "toobig", "items", "newfile.txt"), pr)

		req.Header.Set("Content-Type", writer.FormDataContentType())

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusInsufficientStorage {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}

		_, err := os.Stat(path.Join("tmptest/data/", "toobig", "newfile.txt"))
		if err == nil {
			t.Errorf("Expected file to be deleted")
			return
		}
	})

	t.Run("Upload too much data on a share shouldn't work", func(t *testing.T) {
		makeShare("sharetoobig", ShareParameters{
			Exposure: "upload",
			Validity: 7,
		})

		fileSize := 3 * 1024 * 1024
		pr, writer := multipartWriter(fileSize)

		req = httptest.NewRequest("POST", path.Join("/api/v1/shares", "sharetoobig", "items", "newfile.txt"), pr)

		req.Header.Set("Content-Type", writer.FormDataContentType())

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		fileSize = 3 * 1024 * 1024
		pr, writer = multipartWriter(fileSize)

		req = httptest.NewRequest("POST", path.Join("/api/v1/shares", "sharetoobig", "items", "newfile2.txt"), pr)

		req.Header.Set("Content-Type", writer.FormDataContentType())

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusInsufficientStorage {
			t.Errorf("Expected status %d, got %d", http.StatusInsufficientStorage, w.Code)
			return
		}

		_, err := os.Stat(path.Join("tmptest/data/", "sharetoobig", "newfile2.txt"))
		if err == nil {
			t.Errorf("Expected file to be deleted")
			return
		}
	})
}

func TestVersion(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("tmptest")
	})

	api := getAPIServer(t)

	t.Run("Get version with authentication should succceed", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)
		req = httptest.NewRequest("GET", "/api/v1/version", nil)
		req.SetBasicAuth("admin", "hupload")
		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}

		var v struct {
			Version string `json:"version"`
		}
		_ = json.NewDecoder(w.Body).Decode(&v)

		if v.Version != "0.0.0" {
			t.Errorf("Expected version 0.0.0, got %s", v.Version)
			return
		}
	})

	t.Run("Get version without authentication should fail", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)
		req = httptest.NewRequest("GET", "/api/v1/version", nil)

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}

		var v struct {
			Version string `json:"version"`
		}
		_ = json.NewDecoder(w.Body).Decode(&v)

		if v.Version != "" {
			t.Errorf("Expected version \"\", got %s", v.Version)
			return
		}
	})
}
