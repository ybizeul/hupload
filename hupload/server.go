package main

import (
	"os"
	"strconv"

	"log/slog"

	"github.com/ybizeul/hupload/pkg/apiws"
)

func startWebServer(api *apiws.APIWS) {

	// Define authenticators for protected routes
	authenticators := []apiws.AuthMiddleware{
		apiws.BasicAuthMiddleware{
			AuthService: api.AuthService,
		},
		apiws.JWTAuthMiddleware{
			HMACSecret: os.Getenv("JWT_SECRET"),
		},
	}

	// Setup routes
	api.AddRoute("POST /api/v1/share/{share}/{item}", nil, postItem)
	api.AddRoute("GET /api/v1/share/{share}", nil, getShare)
	api.AddRoute("GET /api/v1/share/{share}/{item}", nil, getItem)

	// Protected routes
	api.AddRoute("POST /api/v1/login", authenticators, postLogin)
	api.AddRoute("POST /api/v1/share", authenticators, postShare)
	api.AddRoute("DELETE /api/v1/share/{share}", authenticators, deleteShare)
	api.AddRoute("PUT /api/v1/share/{share}", authenticators, putShare)
	api.AddRoute("GET /api/v1/share", authenticators, getShares)
	api.AddRoute("GET /api/v1/version", authenticators, getVersion)

	if os.Getenv("HTTP_PORT") != "" {
		p, err := strconv.Atoi(os.Getenv("HTTP_PORT"))
		if err != nil {
			slog.Error("Unable to use HTTP_PORT")
			panic(err)
		}
		api.HTTPPort = p
	}
	api.Start()
}
