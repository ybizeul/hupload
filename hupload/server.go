package main

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"
	"strconv"

	"log/slog"

	"github.com/ybizeul/apiws"
	"github.com/ybizeul/hupload/internal/config"
	"github.com/ybizeul/hupload/middleware"
)

type Hupload struct {
	Config *config.Config
	API    *apiws.APIWS
}

func NewHupload(c *config.Config) (*Hupload, error) {

	// Load configuration
	found, err := c.Load()
	if !found {
		slog.Warn("No configuration file found, using default values", "path", c.Path)
	}
	if err != nil {
		return nil, err
	}

	// Run migration
	err = c.Storage.Migrate()
	if err != nil {
		return nil, err
	}

	// Create API web service with the embedded UI
	api, err := apiws.New(uiFS, c.Values)
	if err != nil {
		return nil, err
	}

	api.WithAuthentication(c.Authentication)
	result := &Hupload{
		Config: c,
		API:    api,
	}

	result.setup()

	return result, nil
}

func (h *Hupload) Start() {
	h.API.Start()
}

func (h *Hupload) setup() {

	api := h.API

	// Setup routes

	// Guests can access a share and post new files in it
	// That's Hupload principle, the security is based on the share name
	// which is usually a random string.

	api.AddPublicRoute("GET    /api/v1/shares/{share}", shareCheck(http.HandlerFunc(h.getShare)))
	api.AddPublicRoute("GET    /api/v1/shares/{share}/items", shareCheck(http.HandlerFunc(h.getShareItems)))
	api.AddPublicRoute("GET    /api/v1/shares/{share}/items/{item}", shareAndItemCheck(http.HandlerFunc(h.getItem)))

	api.AddPublicRoute("POST   /api/v1/shares/{share}/items/{item}", shareAndItemCheck(http.HandlerFunc(h.postItem)))
	api.AddPublicRoute("DELETE /api/v1/shares/{share}/items/{item}", shareAndItemCheck(http.HandlerFunc(h.deleteItem)))

	api.AddPublicRoute("GET    /d/{share}", shareCheck(http.HandlerFunc(h.downloadShare)))
	api.AddPublicRoute("GET    /d/{share}/{item}", shareAndItemCheck(http.HandlerFunc(h.getItem)))

	// Protected routes

	api.AddRoute("GET    /api/v1/defaults", http.HandlerFunc(h.getDefaults))

	api.AddRoute("GET    /api/v1/shares", http.HandlerFunc(h.getShares))
	api.AddRoute("POST   /api/v1/shares", http.HandlerFunc(h.postShare))
	api.AddRoute("POST   /api/v1/shares/{share}", shareCheck(http.HandlerFunc(h.postShare)))
	api.AddRoute("PATCH  /api/v1/shares/{share}", shareCheck(http.HandlerFunc(h.patchShare)))
	api.AddRoute("DELETE /api/v1/shares/{share}", shareCheck(http.HandlerFunc(h.deleteShare)))

	api.AddRoute("GET    /api/v1/messages/{index}", http.HandlerFunc(h.getMessage))
	api.AddRoute("GET    /api/v1/messages", http.HandlerFunc(h.getMessages))

	api.AddRoute("GET    /api/v1/version", http.HandlerFunc(h.getVersion))

	api.AddRoute("GET    /api/v1/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusBadRequest, "Error")
	}))

	if os.Getenv("HTTP_PORT") != "" {
		p, err := strconv.Atoi(os.Getenv("HTTP_PORT"))
		if err != nil {
			slog.Error("Unable to use HTTP_PORT")
			panic(err)
		}
		api.WithPort(p)
	}
}

func shareCheck(h http.Handler) http.Handler {
	return middleware.ShareNameCheckMiddleware(h)
}

func shareAndItemCheck(h http.Handler) http.Handler {
	return middleware.ShareNameCheckMiddleware(middleware.ItemNameCheckMiddleware(h))
}

func generateRandomString(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}
