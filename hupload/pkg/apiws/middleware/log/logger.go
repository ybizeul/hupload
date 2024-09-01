package logger

import (
	"log/slog"
	"net/http"
)

type APIWSLogger struct {
	handler http.Handler
}

func NewLogger(handlerToWrap http.Handler) *APIWSLogger {
	return &APIWSLogger{handlerToWrap}
}

func (a *APIWSLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	slog.Info("HTTP Request",
		// slog.String("remote_ip", r.RemoteAddr),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)

	a.handler.ServeHTTP(w, r)
}
