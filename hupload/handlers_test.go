package main

import (
	"bufio"
	"bytes"
	"context"
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
	"reflect"
	"testing"
	"time"

	"github.com/ybizeul/hupload/internal/config"
	"github.com/ybizeul/hupload/internal/storage"
)

var cfgs map[string]struct {
	Config  *config.Config
	Enabled bool
	Cleanup func(h *Hupload)
} = map[string]struct {
	Config  *config.Config
	Enabled bool
	Cleanup func(h *Hupload)
}{
	"file": {
		Config: &config.Config{
			Path: "handlers_testdata/config.yml",
		},
		Enabled: true,
		Cleanup: func(h *Hupload) {
			os.RemoveAll("tmptest")
		},
	},
	"s3": {
		Config: &config.Config{
			Path: "handlers_testdata/config-s3.yml",
		},
		Enabled: os.Getenv("TEST_ENABLE_S3") == "1",
		Cleanup: func(h *Hupload) {
			os.RemoveAll("tmptest")
		},
	},
}

// getHupload returns a new Hupload instance for testing.
func getHupload(t *testing.T, cfg *config.Config) *Hupload {

	h, err := NewHupload(cfg)
	if err != nil {
		t.Fatal(err)
		return nil
	}

	return h
}

// makeShare creates a new share with the given name and parameters.
func makeShare(t *testing.T, h *Hupload, name string, owner string, params storage.Options) *storage.Share {
	share, err := h.Config.Storage.CreateShare(context.Background(), name, owner, params)
	if err != nil {
		t.Error(err)
		return nil
	}
	return share
}

// makeItem creates a new item with the given name and size.
func makeItem(t *testing.T, h *Hupload, shareName, fileName string, size int) {
	_, err := h.Config.Storage.CreateItem(context.Background(), shareName, fileName, int64(size), bufio.NewReader(io.LimitReader(rand.Reader, int64(size))))
	if err != nil {
		t.Fatal(err)
	}
}

func mustUnmarshalJSON(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	return m
}

func TestCreateShare(t *testing.T) {
	for name, cfg := range cfgs {
		if !cfg.Enabled {
			continue
		}
		t.Run(name, func(t *testing.T) {
			h := getHupload(t, cfg.Config)
			t.Cleanup(func() { cfg.Cleanup(h) })

			api := h.API

			// Create a share without authentication should fail
			t.Run("Create a share without authentication should fail", func(t *testing.T) {
				var (
					req *http.Request
					w   *httptest.ResponseRecorder
				)

				req = httptest.NewRequest("POST", "/api/v1/shares", nil)

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusUnauthorized {
					t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
					return
				}
				var got map[string]any
				if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
					t.Errorf("Failed to decode JSON: %v", err)
					return
				}

				want := mustUnmarshalJSON(t, `{"errors":["JWTAuthMiddleware: no Authorization header"]}`)

				if !reflect.DeepEqual(want, got) {
					t.Errorf("Want %s, got %s", want, got)
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

				api.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
					return
				}

				var share map[string]any

				err := json.NewDecoder(w.Body).Decode(&share)
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
					return
				}

				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), share["name"].(string))
				})

				token = &w.Result().Cookies()[0].Value
				if token == nil {
					t.Errorf("Expected token to be created")
					return
				}
			})

			t.Run("Create a share with token should succeed", func(t *testing.T) {
				if token == nil {
					t.Fatal("No token created")
					return
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

				api.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
					return
				}

				var share map[string]any

				err := json.NewDecoder(w.Body).Decode(&share)

				if err != nil {
					t.Errorf("Expected no error, got %v", err)
					return
				}

				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), share["name"].(string))
				})
			})

			t.Run("Create a share with same name should fail", func(t *testing.T) {
				makeShare(t, h, "samename", "admin", storage.Options{})
				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), "samename")
				})
				var (
					req *http.Request
					w   *httptest.ResponseRecorder
				)
				req = httptest.NewRequest("POST", path.Join("/api/v1/shares", "samename"), nil)

				req.AddCookie(&http.Cookie{
					Name:  "X-Token",
					Value: *token,
				})

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusConflict {
					t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
					return
				}
			})

			t.Run("Create a share with specific name should succeed", func(t *testing.T) {
				var (
					req *http.Request
					w   *httptest.ResponseRecorder
				)
				req = httptest.NewRequest("POST", path.Join("/api/v1/shares", "randomname"), nil)
				req.SetBasicAuth("admin", "hupload")

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
					return
				}

				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), "randomname")
				})
			})

			t.Run("Create a share with invalid name should fail", func(t *testing.T) {
				var (
					req *http.Request
					w   *httptest.ResponseRecorder
				)
				req = httptest.NewRequest("POST", path.Join("/api/v1/shares", url.QueryEscape("../test")), nil)

				req.SetBasicAuth("admin", "hupload")

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusBadRequest {
					t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
					return
				}
			})
		})
	}
}

func TestUpdateShare(t *testing.T) {
	for name, cfg := range cfgs {
		if !cfg.Enabled {
			continue
		}
		for altUser := range 2 {
			var username = "admin"
			if altUser == 1 {
				username = "admin2"
			}

			t.Run(name, func(t *testing.T) {
				h := getHupload(t, cfg.Config)
				h.Config.Values.HideOtherShares = (altUser == 1)

				t.Cleanup(func() { cfg.Cleanup(h) })
				api := h.API

				makeShare(t, h, "testupdate", "admin", storage.Options{
					Exposure:    "upload",
					Validity:    7,
					Description: "description",
					Message:     "message",
				})
				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), "testupdate")
				})

				t.Run("Update share should succeed", func(t *testing.T) {
					var (
						req *http.Request
						w   *httptest.ResponseRecorder
					)

					payload := `{"exposure":"download","validity":10,"description":"new description","message":"new message"}`

					req = httptest.NewRequest("PATCH", path.Join("/api/v1/shares", "testupdate"), bytes.NewBufferString(payload))
					req.SetBasicAuth(username, "hupload")

					w = httptest.NewRecorder()

					api.ServeHTTP(w, req)

					if altUser == 1 {
						if w.Code != http.StatusForbidden {
							t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
						}
						return
					}
					if w.Code != http.StatusOK {
						t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
						return
					}

					var got, want map[string]any

					_ = json.NewDecoder(w.Body).Decode(&got)
					_ = json.NewDecoder(bytes.NewBufferString(payload)).Decode(&want)

					if !reflect.DeepEqual(got, want) {
						t.Errorf("Want %v, got %v", want, got)
					}
				})
				t.Run("Update share should fail without auth", func(t *testing.T) {
					var (
						req *http.Request
						w   *httptest.ResponseRecorder
					)

					j := `{"exposure":"download","validity":10,"description":"new description","message":"new message"}`

					req = httptest.NewRequest("PATCH", path.Join("/api/v1/shares", "testupdate"), bytes.NewBufferString(j))

					w = httptest.NewRecorder()

					api.ServeHTTP(w, req)

					if w.Code != http.StatusUnauthorized {
						t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
						return
					}
				})

				t.Run("Update share should fail on invalid share name", func(t *testing.T) {
					var (
						req *http.Request
						w   *httptest.ResponseRecorder
					)

					j := `{"exposure":"download","validity":10,"description":"new description","message":"new message"}`

					req = httptest.NewRequest("PATCH", path.Join("/api/v1/shares", url.QueryEscape("../test")), bytes.NewBufferString(j))
					req.SetBasicAuth(username, "hupload")

					w = httptest.NewRecorder()

					api.ServeHTTP(w, req)

					if w.Code != http.StatusBadRequest {
						t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
						return
					}
				})

				t.Run("Update share should fail on inexistant share", func(t *testing.T) {
					var (
						req *http.Request
						w   *httptest.ResponseRecorder
					)

					j := `{"exposure":"download","validity":10,"description":"new description","message":"new message"}`

					req = httptest.NewRequest("PATCH", path.Join("/api/v1/shares", "inexistant"), bytes.NewBufferString(j))
					req.SetBasicAuth(username, "hupload")

					w = httptest.NewRecorder()

					api.ServeHTTP(w, req)

					if w.Code != http.StatusNotFound {
						t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
						return
					}
				})
			})
		}
	}
}

func TestGetShare(t *testing.T) {
	for name, cfg := range cfgs {
		if !cfg.Enabled {
			continue
		}

		for altUser := range 2 {
			var username = "admin"
			if altUser == 1 {
				username = "admin2"
			}

			t.Run(name+" "+username, func(t *testing.T) {
				h := getHupload(t, cfg.Config)
				h.Config.Values.HideOtherShares = (altUser == 1)

				t.Cleanup(func() { cfg.Cleanup(h) })
				api := h.API

				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), "test")
					_ = h.Config.Storage.DeleteShare(context.Background(), "test2")
					_ = h.Config.Storage.DeleteShare(context.Background(), "test3")
				})

				makeShare(t, h, "test", "admin", storage.Options{
					Exposure:    "upload",
					Validity:    7,
					Description: "description",
					Message:     "message",
				})

				makeItem(t, h, "test", "file1", 1024)

				makeShare(t, h, "test2", "admin", storage.Options{
					Exposure:    "upload",
					Validity:    7,
					Description: "description",
					Message:     "message",
				})

				makeShare(t, h, "test3", "admin2", storage.Options{
					Exposure:    "download",
					Validity:    7,
					Description: "description",
					Message:     "message",
				})

				makeItem(t, h, "test3", "file1", 1024)

				t.Run("Get shares should succeed ("+name+")", func(t *testing.T) {
					var (
						req *http.Request
						w   *httptest.ResponseRecorder
					)
					req = httptest.NewRequest("GET", "/api/v1/shares", nil)
					req.SetBasicAuth(username, "hupload")
					w = httptest.NewRecorder()

					api.ServeHTTP(w, req)

					if w.Code != http.StatusOK {
						t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
						return
					}

					var shares []map[string]any

					_ = json.NewDecoder(w.Body).Decode(&shares)

					if altUser == 1 {
						if len(shares) != 1 {
							t.Errorf("Expected 1 shares, got %d", len(shares))
						}
						return
					}

					if len(shares) != 3 {
						t.Errorf("Expected 3 shares, got %d", len(shares))
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

					api.ServeHTTP(w, req)

					if w.Code != http.StatusUnauthorized {
						t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
						return
					}

					var got map[string]any
					if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
						t.Errorf("Failed to decode JSON: %v", err)
						return
					}
					want := mustUnmarshalJSON(t, `{"errors":["JWTAuthMiddleware: no Authorization header"]}`)
					if !reflect.DeepEqual(want, got) {
						t.Errorf("Want %s, got %s", want, got)
					}
				})

				t.Run("Get share should succeed", func(t *testing.T) {
					var (
						req *http.Request
						w   *httptest.ResponseRecorder
					)

					// Do download on the item
					req = httptest.NewRequest("GET", path.Join("/d/", "test3", "file1"), nil)
					w = httptest.NewRecorder()
					api.ServeHTTP(w, req)
					_, _ = io.Copy(io.Discard, w.Body)

					// fileReader, err := h.Config.Storage.GetItemData(context.Background(), "test", "file1")
					// if err != nil {
					// 	t.Errorf("Failed to get item data: %v", err)
					// 	return
					// }
					// _, _ = io.Copy(io.Discard, fileReader)
					// _ = fileReader.Close()

					tests := []struct {
						ShareName string
						Want      string
					}{
						{
							ShareName: "test",
							Want: `{
								"version":1,
								"name":"test",
								"owner":"admin",
								"options":{
									"validity":7,
									"exposure":"upload",
									"description":"description",
									"message":"message"
								},
								"count":1,
								"size":1024
							}`},
						{
							ShareName: "test3",
							Want: `{
								"version":1,
								"name":"test3",
								"owner":"admin2",
								"options":{
									"validity":7,
									"exposure":"download",
									"description":"description",
									"message":"message"
								},
								"count":1,
								"size":1024
							}`},
					}
					for _, tt := range tests {
						req = httptest.NewRequest("GET", path.Join("/api/v1/shares/", tt.ShareName), nil)
						req.SetBasicAuth(username, "hupload")
						w = httptest.NewRecorder()

						api.ServeHTTP(w, req)

						if w.Code != http.StatusOK {
							t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
							return
						}

						var share map[string]any

						_ = json.NewDecoder(w.Body).Decode(&share)

						delete(share, "created")

						want := mustUnmarshalJSON(t, tt.Want)

						if !reflect.DeepEqual(share, want) {
							t.Errorf("Want %v, got %v", want, share)
						}
					}
				})

				t.Run("Get share without authentication should succeed with filtered properties", func(t *testing.T) {
					var (
						req *http.Request
						w   *httptest.ResponseRecorder
					)
					req = httptest.NewRequest("GET", path.Join("/api/v1/shares/", "test"), nil)

					w = httptest.NewRecorder()

					api.ServeHTTP(w, req)

					if w.Code != http.StatusOK {
						t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
						return
					}

					var share map[string]any

					_ = json.NewDecoder(w.Body).Decode(&share)

					want := mustUnmarshalJSON(t, `
					{
						"name":"test",
						"options":{
							"exposure":"upload",
							"message":"message"
						}
					}`)

					if !reflect.DeepEqual(share, want) {
						t.Errorf("Want %v, got %v", want, share)
					}

					if !reflect.DeepEqual(share, want) {
						t.Errorf("Want %v, got %v", want, share)
					}
				})

				t.Run("Get share with invalid name should fail", func(t *testing.T) {
					var (
						req *http.Request
						w   *httptest.ResponseRecorder
					)
					req = httptest.NewRequest("GET", path.Join("/api/v1/shares/", url.QueryEscape("../test")), nil)

					w = httptest.NewRecorder()

					api.ServeHTTP(w, req)

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

					api.ServeHTTP(w, req)

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

					makeShare(t, h, shareName, "admin", storage.Options{Exposure: "download"})
					t.Cleanup(func() {
						_ = h.Config.Storage.DeleteShare(context.Background(), shareName)
					})

					makeItem(t, h, shareName, "newfile.txt", 1*1024*1024)
					time.Sleep(1 * time.Second)
					makeItem(t, h, shareName, "newfile2.txt", 2*1024*1024)

					// Download first item

					req = httptest.NewRequest("GET", path.Join("/d/", shareName, "newfile.txt"), nil)

					w = httptest.NewRecorder()

					api.ServeHTTP(w, req)

					_, _ = io.Copy(io.Discard, w.Body)

					// Without authentication
					req = httptest.NewRequest("GET", path.Join("/api/v1/shares/", shareName, "items"), nil)

					w = httptest.NewRecorder()

					api.ServeHTTP(w, req)

					if w.Code != http.StatusOK {
						t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
						return
					}

					var result []map[string]any

					_ = json.NewDecoder(w.Body).Decode(&result)

					if len(result) != 2 {
						t.Errorf("Expected 2 items, got %d", len(result))
						return
					}

					for _, item := range result {
						delete(item["ItemInfo"].(map[string]any), "created")
						delete(item["ItemInfo"].(map[string]any), "DateModified")
					}

					want := []map[string]any{
						mustUnmarshalJSON(t, `
							{
								"Path":"itemstest/newfile2.txt",
								"ItemInfo":{
									"Size":2097152
								}
							}`),
						mustUnmarshalJSON(t, `
							{
								"Path":"itemstest/newfile.txt",
								"ItemInfo":{
									"Size":1048576
								}
							}`),
					}

					if !reflect.DeepEqual(result, want) {
						t.Errorf("Want %v, got %v", want, result)
					}

					// With authentication
					req = httptest.NewRequest("GET", path.Join("/api/v1/shares/", shareName, "items"), nil)
					req.SetBasicAuth("admin", "hupload")
					w = httptest.NewRecorder()

					api.ServeHTTP(w, req)

					if w.Code != http.StatusOK {
						t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
						return
					}

					_ = json.NewDecoder(w.Body).Decode(&result)

					if len(result) != 2 {
						t.Errorf("Expected 2 items, got %d", len(result))
						return
					}

					for _, item := range result {
						delete(item["ItemInfo"].(map[string]any), "DateModified")
					}

					want = []map[string]any{
						mustUnmarshalJSON(t, `
							{
								"Path":"itemstest/newfile2.txt",
								"ItemInfo":{
									"Size":2097152
								}
							}`),
						mustUnmarshalJSON(t, `
							{
								"Path":"itemstest/newfile.txt",
								"Downloads":1,
								"ItemInfo":{
									"Size":1048576
								}
							}`),
					}

					if !reflect.DeepEqual(result, want) {
						t.Errorf("Want %v, got %v", want, result)
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

					api.ServeHTTP(w, req)

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

					api.ServeHTTP(w, req)

					if w.Code != http.StatusBadRequest {
						t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
						return
					}
				})
			})
		}
	}
}

func TestDownloadShare(t *testing.T) {
	for name, cfg := range cfgs {
		if !cfg.Enabled {
			continue
		}

		shareName := "downloadshare"

		h := getHupload(t, cfg.Config)

		makeShare(t, h, shareName, "admin", storage.Options{Exposure: "download"})

		t.Cleanup(func() {
			_ = h.Config.Storage.DeleteShare(context.Background(), shareName)
			cfg.Cleanup(h)
		})

		makeItem(t, h, shareName, "newfile1.txt", 1*1024*1024)
		makeItem(t, h, shareName, "newfile2.txt", 1*1024*1024)

		t.Run(name, func(t *testing.T) {
			api := h.API

			req := httptest.NewRequest("GET", path.Join("/d/", shareName), nil)

			w := httptest.NewRecorder()

			api.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
				return
			}
		})

		t.Run(name+" mass download counter updated", func(t *testing.T) {
			storage := h.Config.Storage
			items, err := storage.ListShare(context.Background(), shareName)
			if err != nil {
				t.Fatal(err)
			}
			for _, item := range items {
				// Check if the mass download counter was updated
				if item.Downloads != 1 {
					t.Errorf("Expected mass download counter to be updated, got %d", item.Downloads)
				}
			}
		})
	}
}

func TestDeleteShare(t *testing.T) {
	for name, cfg := range cfgs {
		if !cfg.Enabled {
			continue
		}
		t.Run(name, func(t *testing.T) {
			h := getHupload(t, cfg.Config)
			t.Cleanup(func() { cfg.Cleanup(h) })
			api := h.API

			t.Run("delete share should work", func(t *testing.T) {
				var (
					req *http.Request
					w   *httptest.ResponseRecorder
				)

				shareName := "deleteshare"
				makeShare(t, h, shareName, "admin", storage.Options{Exposure: "download"})
				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), shareName)
				})
				req = httptest.NewRequest("DELETE", path.Join("/api/v1/shares/", shareName), nil)
				req.SetBasicAuth("admin", "hupload")
				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

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
				makeShare(t, h, shareName, "admin", storage.Options{Exposure: "download"})

				req = httptest.NewRequest("DELETE", path.Join("/api/v1/shares/", shareName), nil)
				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusUnauthorized {
					t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
					return
				}
				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), "deleteshare")
				})
			})

			t.Run("delete share invalid share name shouldn't work", func(t *testing.T) {
				var (
					req *http.Request
					w   *httptest.ResponseRecorder
				)

				req = httptest.NewRequest("DELETE", path.Join("/api/v1/shares/", url.QueryEscape("../bogus")), nil)
				req.SetBasicAuth("admin", "hupload")

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

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

				api.ServeHTTP(w, req)

				if w.Code != http.StatusNotFound {
					t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
					return
				}
			})
		})
	}
}

func TestGetItems(t *testing.T) {
	for name, cfg := range cfgs {
		if !cfg.Enabled {
			continue
		}
		t.Run(name, func(t *testing.T) {
			h := getHupload(t, cfg.Config)
			t.Cleanup(func() { cfg.Cleanup(h) })
			api := h.API

			t.Run("Get share item should work", func(t *testing.T) {
				var (
					req *http.Request
					w   *httptest.ResponseRecorder
				)

				for _, exp := range []string{"download", "both"} {
					shareName := "getitem" + exp
					fileSize := 1 * 1024 * 1024
					makeShare(t, h, shareName, "admin", storage.Options{Exposure: exp})
					t.Cleanup(func() {
						_ = h.Config.Storage.DeleteShare(context.Background(), shareName)
					})

					makeItem(t, h, shareName, "newfile.txt", fileSize)

					t.Cleanup(func() {
						_ = h.Config.Storage.DeleteShare(context.Background(), shareName)
					})

					req = httptest.NewRequest("GET", path.Join("/api/v1/shares/", shareName, "items", "newfile.txt"), nil)

					w = httptest.NewRecorder()

					api.ServeHTTP(w, req)

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

				req = httptest.NewRequest("GET", path.Join("/api/v1/shares/", shareName, "items", url.QueryEscape("item")), nil)

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

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

				makeShare(t, h, shareName, "admin", storage.Options{Exposure: "download"})
				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), shareName)
				})

				makeItem(t, h, shareName, "newfile.txt", 1*1024*1024)

				req = httptest.NewRequest("GET", path.Join("/api/v1/shares/", shareName, "items", url.QueryEscape(".metadata")), nil)

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

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

				makeShare(t, h, shareName, "admin", storage.Options{Exposure: "download"})
				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), shareName)
				})

				req = httptest.NewRequest("GET", path.Join("/api/v1/shares/", shareName, "items", "notexists"), nil)

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

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

				makeShare(t, h, shareName, "admin", storage.Options{Exposure: "upload"})
				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), shareName)
				})

				makeItem(t, h, shareName, "newfile.txt", 1*1024*1024)

				req = httptest.NewRequest("GET", path.Join("/api/v1/shares/", shareName, "items", "newfile.txt"), nil)
				req.SetBasicAuth("admin", "hupload")

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
					return
				}
			})
		})
	}
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
	for name, cfg := range cfgs {
		if !cfg.Enabled {
			continue
		}
		t.Run(name, func(t *testing.T) {
			h := getHupload(t, cfg.Config)
			t.Cleanup(func() { cfg.Cleanup(h) })
			api := h.API
			var (
				req *http.Request
				w   *httptest.ResponseRecorder
			)

			t.Run("Upload a file without authentication should work", func(t *testing.T) {
				// Create upload share
				makeShare(t, h, "upload", "admin", storage.Options{
					Exposure: "upload",
					Validity: 7,
				})

				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), "upload")
				})

				makeItem(t, h, "upload", "newfile.txt", 1*1024*1024)

				fileSize := 1 * 1024 * 1024

				pr, ct := multipartWriter(fileSize)

				req = httptest.NewRequest("POST", path.Join("/api/v1/shares", "upload", "items", "newfile.txt"), pr)

				req.Header.Set("Content-Type", ct)
				req.Header.Set("FileSize", fmt.Sprintf("%d", fileSize))

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
					return
				}
			})

			t.Run("Upload a file without authentication should not work (download share)", func(t *testing.T) {
				makeShare(t, h, "download", "admin", storage.Options{
					Exposure: "download",
					Validity: 7,
				})

				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), "download")
				})

				fileSize := 3 * 1024 * 1024
				pr, ct := multipartWriter(fileSize)

				req = httptest.NewRequest("POST", path.Join("/api/v1/shares", "download", "items", "newfile.txt"), pr)

				req.Header.Set("Content-Type", ct)

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusUnauthorized {
					t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
					return
				}

				// _, err := os.Stat(path.Join("tmptest/data/", "download", "newfile.txt"))
				// if err == nil {
				// 	t.Errorf("Expected file to not be created")
				// 	return
				// }
			})

			t.Run("Upload a file without authentication should work authenticated (download share)", func(t *testing.T) {
				shareName := "uploadondownloadwithauth"
				makeShare(t, h, shareName, "admin", storage.Options{
					Exposure: "download",
					Validity: 7,
				})

				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), shareName)
				})

				fileSize := 3 * 1024 * 1024
				pr, ct := multipartWriter(fileSize)

				req = httptest.NewRequest("POST", path.Join("/api/v1/shares", shareName, "items", "newfile.txt"), pr)

				req.SetBasicAuth("admin", "hupload")

				req.Header.Set("Content-Type", ct)
				req.Header.Set("FileSize", fmt.Sprintf("%d", fileSize))

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
					return
				}
			})

			t.Run("Upload a file too big should not work", func(t *testing.T) {
				makeShare(t, h, "toobig", "admin", storage.Options{
					Exposure: "upload",
					Validity: 7,
				})

				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), "toobig")
				})

				fileSize := 3*1024*1024 + 1
				pr, ct := multipartWriter(fileSize)

				req = httptest.NewRequest("POST", path.Join("/api/v1/shares", "toobig", "items", "newfile.txt"), pr)

				req.Header.Set("Content-Type", ct)
				req.Header.Set("FileSize", fmt.Sprintf("%d", fileSize))

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusInsufficientStorage {
					t.Errorf("Expected status %d, got %d", http.StatusInsufficientStorage, w.Code)
					return
				}
			})

			t.Run("Upload too much data on a share shouldn't work", func(t *testing.T) {
				makeShare(t, h, "sharetoobig", "admin", storage.Options{
					Exposure: "upload",
					Validity: 7,
				})

				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), "sharetoobig")
				})

				fileSize := 3 * 1024 * 1024
				pr, ct := multipartWriter(fileSize)

				req = httptest.NewRequest("POST", path.Join("/api/v1/shares", "sharetoobig", "items", "newfile.txt"), pr)

				req.Header.Set("Content-Type", ct)
				req.Header.Set("FileSize", fmt.Sprintf("%d", fileSize))

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				fileSize = 3 * 1024 * 1024
				pr, ct = multipartWriter(fileSize)

				req = httptest.NewRequest("POST", path.Join("/api/v1/shares", "sharetoobig", "items", "newfile2.txt"), pr)

				req.Header.Set("Content-Type", ct)
				req.Header.Set("FileSize", fmt.Sprintf("%d", fileSize))

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusInsufficientStorage {
					t.Errorf("Expected status %d, got %d", http.StatusInsufficientStorage, w.Code)
					return
				}
			})

			t.Run("Upload to invalid share should fail", func(t *testing.T) {
				fileSize := 3 * 1024 * 1024
				pr, ct := multipartWriter(fileSize)

				req = httptest.NewRequest("POST", path.Join("/api/v1/shares", url.QueryEscape("../test"), "items", "newfile.txt"), pr)

				req.Header.Set("Content-Type", ct)
				req.Header.Set("FileSize", fmt.Sprintf("%d", fileSize))

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusBadRequest {
					t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
					return
				}
			})

			t.Run("Upload to inexistant share should fail", func(t *testing.T) {
				fileSize := 3 * 1024 * 1024
				pr, ct := multipartWriter(fileSize)

				req = httptest.NewRequest("POST", path.Join("/api/v1/shares", "inexistant", "items", "newfile.txt"), pr)

				req.Header.Set("Content-Type", ct)
				req.Header.Set("FileSize", fmt.Sprintf("%d", fileSize))

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusNotFound {
					t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
					return
				}
			})

			t.Run("Upload malformed data should fail", func(t *testing.T) {
				makeShare(t, h, "malformed", "admin", storage.Options{})
				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), "malformed")
				})

				req = httptest.NewRequest("POST", path.Join("/api/v1/shares", "malformed", "items", "newfile.txt"), nil)

				req.Header.Set("Content-Type", "text/plain")
				req.Header.Set("FileSize", "10")

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusBadRequest {
					t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
					return
				}
			})

			t.Run("Upload a file without file size should fail", func(t *testing.T) {
				// Create upload share
				makeShare(t, h, "nofilesize", "admin", storage.Options{
					Exposure: "upload",
					Validity: 7,
				})

				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), "nofilesize")
				})

				fileSize := 1 * 1024 * 1024

				pr, ct := multipartWriter(fileSize)

				req = httptest.NewRequest("POST", path.Join("/api/v1/shares", "nofilesize", "items", "newfile.txt"), pr)

				req.Header.Set("Content-Type", ct)

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusBadRequest {
					t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
					return
				}
			})

			t.Run("Upload a file with invalid file name should fail", func(t *testing.T) {
				// Create upload share
				makeShare(t, h, "upload", "admin", storage.Options{
					Exposure: "upload",
					Validity: 7,
				})

				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), "upload")
				})

				fileSize := 1 * 1024 * 1024

				pr, ct := multipartWriter(fileSize)

				req = httptest.NewRequest("POST", path.Join("/api/v1/shares", "upload", "items", url.QueryEscape("../file.txt")), pr)

				req.Header.Set("Content-Type", ct)
				req.Header.Set("FileSize", fmt.Sprintf("%d", fileSize))

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusBadRequest {
					t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
					return
				}
			})
		})
	}
}

func TestDeleteItem(t *testing.T) {
	for name, cfg := range cfgs {
		if !cfg.Enabled {
			continue
		}
		t.Run(name, func(t *testing.T) {
			h := getHupload(t, cfg.Config)
			t.Cleanup(func() { cfg.Cleanup(h) })
			api := h.API

			var (
				req *http.Request
				w   *httptest.ResponseRecorder
			)

			t.Run("delete a file as admin should work", func(t *testing.T) {
				// Create upload share
				share := makeShare(t, h, "uploadadmin", "admin", storage.Options{
					Exposure: "upload",
					Validity: 7,
				})
				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), "uploadadmin")
				})

				makeItem(t, h, share.Name, "newfile.txt", 1*1024*1024)

				req = httptest.NewRequest("DELETE", path.Join("/api/v1/shares", share.Name, "items", "newfile.txt"), nil)

				req.SetBasicAuth("admin", "hupload")

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
					return
				}
			})
			t.Run("delete a file as guest should work on a upload share", func(t *testing.T) {
				// Create upload share
				share := makeShare(t, h, "upload", "admin", storage.Options{
					Exposure: "upload",
					Validity: 7,
				})
				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), "upload")
				})
				makeItem(t, h, share.Name, "newfile.txt", 1*1024*1024)

				req = httptest.NewRequest("DELETE", path.Join("/api/v1/shares", share.Name, "items", "newfile.txt"), nil)

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
					return
				}
			})

			t.Run("delete a file as guest should work on a both share", func(t *testing.T) {
				// Create upload share
				share := makeShare(t, h, "both", "admin", storage.Options{
					Exposure: "both",
					Validity: 7,
				})
				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), "both")
				})

				makeItem(t, h, share.Name, "newfile.txt", 1*1024*1024)

				req = httptest.NewRequest("DELETE", path.Join("/api/v1/shares", share.Name, "items", "newfile.txt"), nil)

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
					return
				}
			})

			t.Run("delete a file as guest should fail on a download share", func(t *testing.T) {
				// Create upload share
				share := makeShare(t, h, "download", "admin", storage.Options{
					Exposure: "download",
					Validity: 7,
				})
				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), "download")
				})
				makeItem(t, h, share.Name, "newfile.txt", 1*1024*1024)

				req = httptest.NewRequest("DELETE", path.Join("/api/v1/shares", share.Name, "items", "newfile.txt"), nil)

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusUnauthorized {
					t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
					return
				}
			})

			t.Run("delete a file on invalid share name should fail", func(t *testing.T) {
				// Create upload share

				req = httptest.NewRequest("DELETE", path.Join("/api/v1/shares", url.QueryEscape("../share"), "items", "newfile.txt"), nil)

				req.SetBasicAuth("admin", "hupload")

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusBadRequest {
					t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
					return
				}
			})

			t.Run("delete a file on inexistant share should fail", func(t *testing.T) {
				// Create upload share

				req = httptest.NewRequest("DELETE", path.Join("/api/v1/shares", "inexistant", "items", "newfile.txt"), nil)

				req.SetBasicAuth("admin", "hupload")

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusNotFound {
					t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
					return
				}
			})

			t.Run("delete an invalid file name should fail", func(t *testing.T) {
				makeShare(t, h, "invaliditem", "admin", storage.Options{})
				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), "invaliditem")
				})
				req = httptest.NewRequest("DELETE", path.Join("/api/v1/shares", "invaliditem", "items", url.QueryEscape("../file.txt")), nil)

				req.SetBasicAuth("admin", "hupload")

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusBadRequest {
					t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
					return
				}
			})

			t.Run("delete an inexistant file should fail", func(t *testing.T) {
				makeShare(t, h, "inexistantfile", "admin", storage.Options{})
				t.Cleanup(func() {
					_ = h.Config.Storage.DeleteShare(context.Background(), "inexistantfile")
				})
				req = httptest.NewRequest("DELETE", path.Join("/api/v1/shares", "inexistantfile", "items", "newfile.txt"), nil)

				req.SetBasicAuth("admin", "hupload")

				w = httptest.NewRecorder()

				api.ServeHTTP(w, req)

				if w.Code != http.StatusNotFound {
					t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
					return
				}

			})
		})
	}
}

func TestMessages(t *testing.T) {
	c := &config.Config{
		Path: "handlers_testdata/config.yml",
	}

	h := getHupload(t, c)

	t.Run("Get messages should work", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/messages", nil)

		req.SetBasicAuth("admin", "hupload")

		w := httptest.NewRecorder()

		h.API.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}

		got := []string{}
		_ = json.NewDecoder(w.Body).Decode(&got)

		want := []string{"Message title"}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("Expected %v, got %v", want, got)
			return
		}
	})

	t.Run("Get message without auth should fail", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/messages", nil)

		w := httptest.NewRecorder()

		h.API.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}

		type btype struct {
			Errors []string `json:"errors"`
		}

		got := btype{}

		_ = json.NewDecoder(w.Body).Decode(&got)

		want := btype{
			Errors: []string{"JWTAuthMiddleware: no Authorization header"},
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("Expected %v, got %v", want, got)
			return
		}
	})

	t.Run("Get message should work", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/messages/1", nil)

		req.SetBasicAuth("admin", "hupload")

		w := httptest.NewRecorder()

		h.API.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}

		var got *config.MessageTemplate

		_ = json.NewDecoder(w.Body).Decode(&got)

		want := &config.MessageTemplate{
			Title:   "Message title",
			Message: "Message content",
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("Expected %v, got %v", want, got)
			return
		}
	})

	t.Run("Get message without auth should fail", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/messages/1", nil)

		w := httptest.NewRecorder()

		h.API.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
			return
		}

		type btype struct {
			Errors []string `json:"errors"`
		}

		got := btype{}

		_ = json.NewDecoder(w.Body).Decode(&got)

		want := btype{
			Errors: []string{"JWTAuthMiddleware: no Authorization header"},
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("Expected %v, got %v", want, got)
			return
		}
	})
}

func TestVersion(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("tmptest")
	})

	h := getHupload(t, cfgs["file"].Config)
	api := h.API

	t.Run("Get version with authentication should succceed", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)
		req = httptest.NewRequest("GET", "/api/v1/version", nil)
		req.SetBasicAuth("admin", "hupload")
		w = httptest.NewRecorder()

		api.ServeHTTP(w, req)

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

		api.ServeHTTP(w, req)

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

func TestDefaults(t *testing.T) {
	h := getHupload(t, cfgs["file"].Config)
	api := h.API

	t.Run("Get defaults with authentication should succceed", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)
		req = httptest.NewRequest("GET", "/api/v1/defaults", nil)
		req.SetBasicAuth("admin", "hupload")
		w = httptest.NewRecorder()

		api.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}

		got := struct {
			Validity int    `json:"validity"`
			Exposure string `json:"exposure"`
		}{}

		json.NewDecoder(w.Body).Decode(&got)

		want := struct {
			Validity int    `json:"validity"`
			Exposure string `json:"exposure"`
		}{
			Validity: 12,
			Exposure: "download",
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("Expected %v, got %v", want, got)
			return
		}
	})

	t.Run("Get defaults without authentication should fail", func(t *testing.T) {
		var (
			req *http.Request
			w   *httptest.ResponseRecorder
		)
		req = httptest.NewRequest("GET", "/api/v1/defaults", nil)

		w = httptest.NewRecorder()

		api.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
			return
		}
	})
}
