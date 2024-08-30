package main

import (
	"embed"
	"log/slog"
	"os"

	"github.com/ybizeul/hupload/internal/config"
)

//go:embed admin-ui
var uiFS embed.FS

//var api *apiws.APIWS

//var cfg config.Config

func main() {
	initLogging()
	slog.Info("Start Hupload")

	// Create configuration struct
	cfgPath := os.Getenv("CONFIG")
	if cfgPath == "" {
		cfgPath = "config.yml"
	}
	cfg := config.Config{
		Path: cfgPath,
	}

	h, err := NewHupload(&cfg)
	if err != nil {
		panic(err)
	}

	// Start the web server
	h.Start()
}
