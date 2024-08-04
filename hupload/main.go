package main

import (
	"embed"
	"errors"
	"log/slog"
	"os"

	"github.com/ybizeul/hupload/internal/config"
	"github.com/ybizeul/hupload/pkg/apiws"
)

//go:embed admin-ui
var uiFS embed.FS

var api *apiws.APIWS

func main() {
	initLogging()
	slog.Info("Start Hupload")

	// Create configuration struct
	cfgPath := os.Getenv("CONFIG")
	if cfgPath == "" {
		cfgPath = "config.yml"
	}
	c := config.Config{
		Path: cfgPath,
	}

	// Load configuration
	err := c.Load()
	if err != nil {
		if errors.Is(err, config.ErrorConfigNoSuchFile) {
			slog.Warn("No configuration file found, using default values", slog.String("path", cfgPath))
		} else {
			panic(err)
		}
	}

	// Create API web service with the embedded UI
	api, err = apiws.New(uiFS, c.Values)
	if err != nil {
		panic(err)
	}

	// Create storage backend from configuration
	b, err := c.Storage()
	if err != nil {
		panic(err)
	}

	// Set as current storage backend for the application
	api.SetStorage(b)

	// Create backend from configuration
	a, err := c.Authentication()
	if err != nil {
		panic(err)
	}

	// Set as current backend for the application
	api.SetAuthentication(a)

	// Start the web server
	startWebServer(api)
}
