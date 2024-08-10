package apiws

import (
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/ybizeul/hupload/internal/config"
	"github.com/ybizeul/hupload/pkg/apiws/authentication"
	"github.com/ybizeul/hupload/pkg/apiws/middleware/auth"
	logger "github.com/ybizeul/hupload/pkg/apiws/middleware/log"
)

type APIWS struct {
	// StaticUI is the file system containing the static web directory.
	StaticUI fs.FS
	// HTTP port to listen on
	HTTPPort int
	// mux is the main ServeMux used by the API Web Server.
	// Public for integration tests
	Mux *http.ServeMux

	// TemplateData is used to customized some templated parts of the web UI.
	TemplateData config.ConfigValues

	// Authentication is the authentication backend
	Authentication authentication.Authentication
}

// New creates a new API Web Server. staticUI is the file system containing the
// web root directory.
func New(staticUI fs.FS, t config.ConfigValues) (*APIWS, error) {
	var f fs.FS = nil

	if staticUI != nil {
		d, err := fs.ReadDir(staticUI, ".")
		if err != nil {
			return nil, err
		}
		f, err = fs.Sub(staticUI, d[0].Name())
		if err != nil {
			return nil, err
		}
	}

	result := &APIWS{
		StaticUI:     f,
		HTTPPort:     8080,
		TemplateData: t,
		Mux:          http.NewServeMux(),
	}

	if f != nil {
		result.Mux.HandleFunc("GET /{path...}", func(w http.ResponseWriter, r *http.Request) {
			_, err := fs.Stat(result.StaticUI, r.URL.Path[1:])
			if err == nil {
				http.ServeFileFS(w, r, result.StaticUI, r.URL.Path[1:])
			} else {
				tmpl, err := template.New("index.html").ParseFS(result.StaticUI, "index.html")
				if err != nil {
					slog.Error("unable to parse template", slog.String("error", err.Error()))
				}
				err = tmpl.Execute(w, result.TemplateData)
				if err != nil {
					slog.Error("unable to execute template", slog.String("error", err.Error()))
				}
			}
		})
	}

	return result, nil
}

// SetAuthentication
func (a *APIWS) SetAuthentication(b authentication.Authentication) {
	a.Authentication = b
}

// AddRoute adds a new route to the API Web Server. pattern is the URL pattern
// to match. authenticators is a list of Authenticator to use to authenticate
// the request. handlerFunc is the function to call when the route is matched.
func (a *APIWS) AddRoute(pattern string, authenticators []auth.AuthMiddleware, handlerFunc func(w http.ResponseWriter, r *http.Request)) {
	if authenticators == nil {
		a.Mux.HandleFunc(pattern, handlerFunc)
	} else {
		var h http.Handler
		h = http.HandlerFunc(handlerFunc)
		c := auth.ConfirmAuthenticator{Realm: "Hupload"}
		h = c.Middleware(h)
		for i := range authenticators {
			h = authenticators[len(authenticators)-1-i].Middleware(h)
		}
		a.Mux.Handle(pattern, h)
	}
}

// Start starts the API Web Server.
func (a *APIWS) Start() {
	slog.Info(fmt.Sprintf("Starting web service on port %d", a.HTTPPort))

	err := http.ListenAndServe(fmt.Sprintf(":%d", a.HTTPPort), logger.NewLogger(a.Mux))
	if err != nil {
		slog.Error("unable to start http server", slog.String("error", err.Error()))
	}
}
