package main

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
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

func makeShare(t *testing.T, name string, params ShareParameters) *storage.Share {
	share, err := cfg.Storage.CreateShare(name, "admin", storage.Options{Validity: params.Validity, Exposure: params.Exposure})
	if err != nil {
		t.Fatal(err)
	}
	return share
}

func makeItem(t *testing.T, shareName, fileName string, size int) {
	_, err := cfg.Storage.CreateItem(shareName, fileName, bufio.NewReader(io.LimitReader(rand.Reader, int64(size))))
	if err != nil {
		t.Fatal(err)
	}
}

// func createItem(t *testing.T, shareName, itemName string, size int) {
// 	fileSize := 1 * 1024 * 1024

// 	makeItem(t, shareName, itemName, fileSize)

// 	pr, ct := multipartWriter(fileSize)

// 	req := httptest.NewRequest("POST", path.Join("/api/v1/shares", "upload", "items", "newfile.txt"), pr)

// 	req.Header.Set("Content-Type", ct)

// 	w := httptest.NewRecorder()

// 	api.Mux.ServeHTTP(w, req)

// 	if w.Code != http.StatusOK {
// 		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
// 		return
// 	}
// }

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

	makeShare(t, "test", ShareParameters{
		Exposure: "upload",
		Validity: 7,
	})

	makeShare(t, "test2", ShareParameters{
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

		if share.Name != "test" || share.Options.Validity != 7 {
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

		if share.Name != "test" || share.Options.Validity != 7 {
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

	t.Run("Get share items should work", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)

		shareName := "itemstest"

		makeShare(t, shareName, ShareParameters{})
		makeItem(t, shareName, "newfile.txt", 1*1024*1024)
		makeItem(t, shareName, "newfile2.txt", 2*1024*1024)

		req = httptest.NewRequest("GET", path.Join("/api/v1/shares/", shareName, "items"), nil)

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}

		var result []storage.Item

		_ = json.NewDecoder(w.Body).Decode(&result)

		if len(result) != 2 {
			t.Errorf("Expected 2 items, got %d", len(result))
			return
		}

		if result[0].Path != path.Join(shareName, "newfile2.txt") {
			t.Errorf("Expected newfile2.txt, got %s", result[0].Path)
			return
		}
		if result[1].Path != path.Join(shareName, "newfile.txt") {
			t.Errorf("Expected newfile.txt, got %s", result[0].Path)
			return
		}

		if result[0].ItemInfo.Size != 2*1024*1024 {
			t.Errorf("Expected size 2*1024*1024, got %d", result[0].ItemInfo.Size)
			return
		}
		if result[1].ItemInfo.Size != 1*1024*1024 {
			t.Errorf("Expected size 1*1024*1024, got %d", result[0].ItemInfo.Size)
			return
		}
	})

	t.Run("Get share items on inexistant share shouldn't work", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)

		shareName := "notexistant"

		req = httptest.NewRequest("GET", path.Join("/api/v1/shares/", shareName, "items"), nil)

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
			return
		}
	})

	t.Run("Get share items on invalid share shouldn't work", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)

		shareName := url.QueryEscape("../notexistant")

		req = httptest.NewRequest("GET", path.Join("/api/v1/shares/", shareName, "items"), nil)

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			return
		}
	})
}

func TestDeleteShare(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("tmptest")
	})

	api := getAPIServer(t)

	t.Run("delete share should work", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)

		shareName := "deleteshare"
		makeShare(t, shareName, ShareParameters{Exposure: "download"})

		req = httptest.NewRequest("DELETE", path.Join("/api/v1/shares/", shareName), nil)
		req.SetBasicAuth("admin", "hupload")
		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}
	})

	t.Run("delete share unauthentication shouldn't work", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)

		shareName := "deleteshare"
		makeShare(t, shareName, ShareParameters{Exposure: "download"})

		req = httptest.NewRequest("DELETE", path.Join("/api/v1/shares/", shareName), nil)
		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
			return
		}
	})

	t.Run("delete share invalid share name shouldn't work", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)

		req = httptest.NewRequest("DELETE", path.Join("/api/v1/shares/", url.QueryEscape("../bogus")), nil)
		req.SetBasicAuth("admin", "hupload")

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			return
		}
	})

	t.Run("delete share inexistant share name shouldn't work", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)

		req = httptest.NewRequest("DELETE", path.Join("/api/v1/shares/", "nonexistant"), nil)
		req.SetBasicAuth("admin", "hupload")

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
			return
		}
	})
}

func TestGetItems(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("tmptest")
	})

	api := getAPIServer(t)

	t.Run("Get share item should work", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)

		for _, exp := range []string{"download", "both"} {
			shareName := "getitem" + exp
			fileSize := 1 * 1024 * 1024
			makeShare(t, shareName, ShareParameters{Exposure: exp})
			makeItem(t, shareName, "newfile.txt", fileSize)

			req = httptest.NewRequest("GET", path.Join("/api/v1/shares/", shareName, "items", "newfile.txt"), nil)

			w = httptest.NewRecorder()

			api.Mux.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
				return
			}

			if w.Result().Header.Get("Content-Length") != fmt.Sprintf("%d", fileSize) {
				t.Errorf("Expected size %d, got %s", fileSize, w.Result().Header.Get("Content-Size"))
				return
			}
		}
	})

	t.Run("Get share items on inexistant share shouldn't work", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)

		shareName := "inexistant"

		req = httptest.NewRequest("GET", path.Join("/api/v1/shares/", shareName, "items", url.QueryEscape(".metadata")), nil)

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
			return
		}
	})

	t.Run("Get share items on invalid share item shouldn't work", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)

		shareName := "invaliditem"

		makeShare(t, shareName, ShareParameters{Exposure: "download"})
		makeItem(t, shareName, "newfile.txt", 1*1024*1024)

		req = httptest.NewRequest("GET", path.Join("/api/v1/shares/", shareName, "items", url.QueryEscape(".metadata")), nil)

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			return
		}
	})

	t.Run("Get share item that does not exists shouldn't work", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)

		shareName := "notexistitem"

		makeShare(t, shareName, ShareParameters{Exposure: "download"})

		req = httptest.NewRequest("GET", path.Join("/api/v1/shares/", shareName, "items", "notexists"), nil)

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
			return
		}
	})

	t.Run("Get share item upload authenticated should work", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)

		shareName := "uploadauth"

		makeShare(t, shareName, ShareParameters{Exposure: "upload"})
		makeItem(t, shareName, "newfile.txt", 1*1024*1024)

		req = httptest.NewRequest("GET", path.Join("/api/v1/shares/", shareName, "items", "newfile.txt"), nil)
		req.SetBasicAuth("admin", "hupload")

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}
	})
}

// multipartWriter returns a reader and a multipart.Writer for a body with one
// attachment of size size.
func multipartWriter(size int) (io.Reader, string) {
	pr, pw := io.Pipe()

	writer := multipart.NewWriter(pw)

	ct := writer.FormDataContentType()

	chunk := make([]byte, 1024)

	go func() {
		defer pw.Close()

		dataPart, err := writer.CreateFormFile("data", "file.txt")
		if err != nil {
			return
		}

		for i := 0; i < size; i += len(chunk) {
			if size-i < len(chunk) {
				chunk = make([]byte, size-i)
			}
			if _, err := dataPart.Write(chunk); err != nil {
				return
			}
		}
		writer.Close()
	}()

	// return to caller
	return pr, ct
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
		makeShare(t, "upload", ShareParameters{
			Exposure: "upload",
			Validity: 7,
		})

		makeItem(t, "upload", "newfile.txt", 1*1024*1024)

		fileSize := 1 * 1024 * 1024

		pr, ct := multipartWriter(fileSize)

		req = httptest.NewRequest("POST", path.Join("/api/v1/shares", "upload", "items", "newfile.txt"), pr)

		req.Header.Set("Content-Type", ct)

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
		makeShare(t, "download", ShareParameters{
			Exposure: "download",
			Validity: 7,
		})

		fileSize := 3 * 1024 * 1024
		pr, ct := multipartWriter(fileSize)

		req = httptest.NewRequest("POST", path.Join("/api/v1/shares", "download", "items", "newfile.txt"), pr)

		req.Header.Set("Content-Type", ct)

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

	t.Run("Upload a file without authentication should work authenticated (download share)", func(t *testing.T) {
		shareName := "uploadondownloadwithauth"
		makeShare(t, shareName, ShareParameters{
			Exposure: "download",
			Validity: 7,
		})

		fileSize := 3 * 1024 * 1024
		pr, ct := multipartWriter(fileSize)

		req = httptest.NewRequest("POST", path.Join("/api/v1/shares", shareName, "items", "newfile.txt"), pr)

		req.SetBasicAuth("admin", "hupload")

		req.Header.Set("Content-Type", ct)

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}

		_, err := os.Stat(path.Join("tmptest/data/", shareName, "newfile.txt"))
		if err != nil {
			t.Errorf("Expected file to be created")
			return
		}
	})

	t.Run("Upload a file too big should not work", func(t *testing.T) {
		makeShare(t, "toobig", ShareParameters{
			Exposure: "upload",
			Validity: 7,
		})

		// writer := multipart.NewWriter(body)
		// // create a new form-data header name data and filename data.txt
		// dataPart, _ := writer.CreateFormFile("data", "file.txt")

		fileSize := 3*1024*1024 + 1
		pr, ct := multipartWriter(fileSize)

		req = httptest.NewRequest("POST", path.Join("/api/v1/shares", "toobig", "items", "newfile.txt"), pr)

		req.Header.Set("Content-Type", ct)

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
		makeShare(t, "sharetoobig", ShareParameters{
			Exposure: "upload",
			Validity: 7,
		})

		fileSize := 3 * 1024 * 1024
		pr, ct := multipartWriter(fileSize)

		req = httptest.NewRequest("POST", path.Join("/api/v1/shares", "sharetoobig", "items", "newfile.txt"), pr)

		req.Header.Set("Content-Type", ct)

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		fileSize = 3 * 1024 * 1024
		pr, ct = multipartWriter(fileSize)

		req = httptest.NewRequest("POST", path.Join("/api/v1/shares", "sharetoobig", "items", "newfile2.txt"), pr)

		req.Header.Set("Content-Type", ct)

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

func TestDeleteItem(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("testdelete")
	})

	api := getAPIServer(t)

	var (
		req *http.Request
		w   *httptest.ResponseRecorder
	)

	t.Run("delete a file as admin should work", func(t *testing.T) {
		t.Cleanup(func() {
			os.RemoveAll("tmptest/data/uploadadmin")
		})

		// Create upload share
		share := makeShare(t, "uploadadmin", ShareParameters{
			Exposure: "upload",
			Validity: 7,
		})

		makeItem(t, share.Name, "newfile.txt", 1*1024*1024)

		req = httptest.NewRequest("DELETE", path.Join("/api/v1/shares", share.Name, "items", "newfile.txt"), nil)

		req.SetBasicAuth("admin", "hupload")

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}
	})
	t.Run("delete a file as guest should work on a upload share", func(t *testing.T) {
		t.Cleanup(func() {
			os.RemoveAll("tmptest/data/upload")
		})
		// Create upload share
		share := makeShare(t, "upload", ShareParameters{
			Exposure: "upload",
			Validity: 7,
		})

		makeItem(t, share.Name, "newfile.txt", 1*1024*1024)

		req = httptest.NewRequest("DELETE", path.Join("/api/v1/shares", share.Name, "items", "newfile.txt"), nil)

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}
	})

	t.Run("delete a file as guest should work on a both share", func(t *testing.T) {
		t.Cleanup(func() {
			os.RemoveAll("tmptest/data/both")
		})
		// Create upload share
		share := makeShare(t, "both", ShareParameters{
			Exposure: "both",
			Validity: 7,
		})

		makeItem(t, share.Name, "newfile.txt", 1*1024*1024)

		req = httptest.NewRequest("DELETE", path.Join("/api/v1/shares", share.Name, "items", "newfile.txt"), nil)

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}
	})

	t.Run("delete a file as guest should fail on a download share", func(t *testing.T) {
		t.Cleanup(func() {
			os.RemoveAll("tmptest/data/download")
		})
		// Create upload share
		share := makeShare(t, "download", ShareParameters{
			Exposure: "download",
			Validity: 7,
		})

		makeItem(t, share.Name, "newfile.txt", 1*1024*1024)

		req = httptest.NewRequest("DELETE", path.Join("/api/v1/shares", share.Name, "items", "newfile.txt"), nil)

		w = httptest.NewRecorder()

		api.Mux.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
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
