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
	// Get JWT_SECRET
	hmac := os.Getenv("JWT_SECRET")
	if len(hmac) == 0 {
		hmac = generateRandomString(32)
	}

	// Define authenticators for protected routes
	// authenticators := []auth.AuthMiddleware{
	// 	auth.BasicAuthMiddleware{
	// 		Authentication: api.Authentication,
	// 	},
	// 	auth.JWTAuthMiddleware{
	// 		HMACSecret: hmac,
	// 	},
	// }
	authenticators := []auth.AuthMiddleware{
		auth.OIDCAuthMiddleware{
			Authentication: api.Authentication,
		},
	}

	// authenticatorsOpen := []auth.AuthMiddleware{
	// 	auth.OpenAuthMiddleware{},
	// 	auth.BasicAuthMiddleware{
	// 		Authentication: api.Authentication,
	// 	},
	// 	auth.JWTAuthMiddleware{
	// 		HMACSecret: hmac,
	// 	},
	// }
	authenticatorsOpen := []auth.AuthMiddleware{
		auth.OpenAuthMiddleware{},
		auth.OIDCAuthMiddleware{
			Authentication: api.Authentication,
		},
	}
	// Setup routes

	// Guests can access a share and post new files in it
	// That's Hupload principle, the security is based on the share name
	// which is usually a random string.

	api.AddRoute("POST   /api/v1/shares/{share}/items/{item}", authenticatorsOpen, h.postItem)
	api.AddRoute("GET    /api/v1/shares/{share}/items", authenticatorsOpen, h.getShareItems)
	api.AddRoute("GET    /api/v1/shares/{share}", authenticatorsOpen, h.getShare)
	api.AddRoute("GET    /api/v1/shares/{share}/items/{item}", authenticatorsOpen, h.getItem)
	api.AddRoute("GET    /d/{share}/{item}", authenticatorsOpen, h.getItem)
	api.AddRoute("DELETE /api/v1/shares/{share}/items/{item}", authenticatorsOpen, h.deleteItem)

	// Protected routes
	api.AddRoute("GET    /login", authenticators, h.postLogin)
	api.AddRoute("POST   /api/v1/login", authenticators, h.postLogin)
	api.AddRoute("POST   /api/v1/shares", authenticators, h.postShare)
	api.AddRoute("POST   /api/v1/shares/{share}", authenticators, h.postShare)
	api.AddRoute("PATCH  /api/v1/shares/{share}", authenticators, h.patchShare)
	api.AddRoute("DELETE /api/v1/shares/{share}", authenticators, h.deleteShare)
	api.AddRoute("GET    /api/v1/shares", authenticators, h.getShares)
	api.AddRoute("GET    /api/v1/version", authenticators, h.getVersion)

	api.AddRoute("GET    /api/v1/*", authenticators, func(w http.ResponseWriter, r *http.Request) {
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
