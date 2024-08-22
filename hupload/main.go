package main

import (
	"embed"
	"log/slog"
	"os"

	"github.com/ybizeul/hupload/internal/config"
	"github.com/ybizeul/hupload/pkg/apiws"
)

//go:embed admin-ui
var uiFS embed.FS

var api *apiws.APIWS

var cfg config.Config

func main() {
	initLogging()
	slog.Info("Start Hupload")

	// Create configuration struct
	cfgPath := os.Getenv("CONFIG")
	if cfgPath == "" {
		cfgPath = "config.yml"
	}
	cfg = config.Config{
		Path: cfgPath,
	}

	// Load configuration
	found, err := cfg.Load()
	if !found {
		slog.Warn("No configuration file found, using default values", slog.String("path", cfgPath))
	}
	if err != nil {
		panic(err)
	}

	// Run migration
	err = cfg.Storage.Migrate()
	if err != nil {
		panic(err)
	}

	// Create API web service with the embedded UI
	api, err = apiws.New(uiFS, cfg.Values)
	if err != nil {
		panic(err)
	}

	api.SetAuthentication(cfg.Authentication)

	// Start the web server
	setup(api)
	api.Start()
}
