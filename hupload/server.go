package main

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"
	"strconv"

	"log/slog"

	"github.com/ybizeul/hupload/internal/config"
	"github.com/ybizeul/hupload/pkg/apiws"
	"github.com/ybizeul/hupload/pkg/apiws/authentication"
	"github.com/ybizeul/hupload/pkg/apiws/middleware/auth"
)

type Hupload struct {
	Config *config.Config
	API    *apiws.APIWS
}

func NewHupload(c *config.Config) (*Hupload, error) {

	// Load configuration
	found, err := c.Load()
	if !found {
		slog.Warn("No configuration file found, using default values", slog.String("path", c.Path))
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

	api.SetAuthentication(c.Authentication)
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

	var authenticator auth.AuthMiddleware
	switch h.Config.Authentication.(type) {
	case *authentication.AuthenticationFile:
		authenticator = auth.BasicAuthMiddleware{
			Authentication: api.Authentication,
		}
	case *authentication.AuthenticationOIDC:
		authenticator = auth.OIDCAuthMiddleware{
			Authentication: api.Authentication,
		}
	}

	// Setup routes

	// Guests can access a share and post new files in it
	// That's Hupload principle, the security is based on the share name
	// which is usually a random string.

	api.AddPublicRoute("POST /api/v1/shares/{share}/items/{item}", authenticator, h.postItem)
	api.AddPublicRoute("GET /api/v1/shares/{share}/items", authenticator, h.getShareItems)
	api.AddPublicRoute("GET /api/v1/shares/{share}", authenticator, h.getShare)
	api.AddPublicRoute("GET /api/v1/shares/{share}/items/{item}", authenticator, h.getItem)
	api.AddPublicRoute("GET /d/{share}/{item}", authenticator, h.getItem)
	api.AddPublicRoute("DELETE /api/v1/shares/{share}/items/{item}", authenticator, h.deleteItem)

	// Protected routes
	api.AddRoute("GET /login", authenticator, h.postLogin)
	api.AddRoute("POST /api/v1/login", authenticator, h.postLogin)
	api.AddRoute("POST /api/v1/shares", authenticator, h.postShare)
	api.AddRoute("POST /api/v1/shares/{share}", authenticator, h.postShare)
	api.AddRoute("PATCH /api/v1/shares/{share}", authenticator, h.patchShare)
	api.AddRoute("DELETE /api/v1/shares/{share}", authenticator, h.deleteShare)
	api.AddRoute("GET /api/v1/shares", authenticator, h.getShares)
	api.AddRoute("GET /api/v1/version", authenticator, h.getVersion)

	api.AddRoute("GET /api/v1/*", authenticator, func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusBadRequest, "Error")
	})

	if os.Getenv("HTTP_PORT") != "" {
		p, err := strconv.Atoi(os.Getenv("HTTP_PORT"))
		if err != nil {
			slog.Error("Unable to use HTTP_PORT")
			panic(err)
		}
		api.HTTPPort = p
	}
}

func generateRandomString(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}
