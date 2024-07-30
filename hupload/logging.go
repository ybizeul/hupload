package main

import (
	"log/slog"
	"os"
	"path/filepath"
)

var logLevel = new(slog.LevelVar)

func initLogging() {
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: logLevel, ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			source, _ := a.Value.Any().(*slog.Source)
			if source != nil {
				source.File = filepath.Base(source.File)
			}
		}
		return a
	}})
	slog.SetDefault(slog.New(h))

	if os.Getenv("DEBUG") != "" {
		logLevel.Set(slog.LevelDebug)
	}
}
