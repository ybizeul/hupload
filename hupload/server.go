package main

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"
	"strconv"

	"log/slog"

	"github.com/ybizeul/hupload/pkg/apiws"
	"github.com/ybizeul/hupload/pkg/apiws/middleware/auth"
)

func setup(api *apiws.APIWS) {

	// Get JWT_SECRET
	hmac := os.Getenv("JWT_SECRET")
	if len(hmac) == 0 {
		hmac = generateRandomString(32)
	}

	// Define authenticators for protected routes
	authenticators := []auth.AuthMiddleware{
		auth.BasicAuthMiddleware{
			Authentication: api.Authentication,
		},
		auth.JWTAuthMiddleware{
			HMACSecret: hmac,
		},
	}

	authenticatorsOpen := []auth.AuthMiddleware{
		auth.OpenAuthMiddleware{},
		auth.BasicAuthMiddleware{
			Authentication: api.Authentication,
		},
		auth.JWTAuthMiddleware{
			HMACSecret: hmac,
		},
	}

	// Setup routes

	// Guests can access a share and post new files in it
	// That's Hupload principle, the security is based on the share name
	// which is usually a random string.

	api.AddRoute("POST /api/v1/shares/{share}/items/{item}", authenticatorsOpen, postItem)
	api.AddRoute("GET /api/v1/shares/{share}/items", authenticatorsOpen, getShareItems)
	api.AddRoute("GET /api/v1/shares/{share}", authenticatorsOpen, getShare)
	api.AddRoute("GET /api/v1/shares/{share}/items/{item}", authenticatorsOpen, getItem)
	api.AddRoute("DELETE /api/v1/shares/{share}/items/{item}", authenticatorsOpen, deleteItem)

	// Protected routes
	api.AddRoute("POST /api/v1/login", authenticators, postLogin)
	api.AddRoute("POST /api/v1/shares", authenticators, postShare)
	api.AddRoute("POST /api/v1/shares/{share}", authenticators, postShare)
	api.AddRoute("DELETE /api/v1/shares/{share}", authenticators, deleteShare)
	//api.AddRoute("PUT /api/v1/shares/{share}", authenticators, putShare)
	api.AddRoute("GET /api/v1/shares", authenticators, getShares)
	api.AddRoute("GET /api/v1/version", authenticators, getVersion)

	api.AddRoute("GET /api/v1/*", authenticators, func(w http.ResponseWriter, r *http.Request) {
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
