package apiws

import (
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/ybizeul/hupload/pkg/apiws/authentication"
	"github.com/ybizeul/hupload/pkg/apiws/middleware/auth"
	logger "github.com/ybizeul/hupload/pkg/apiws/middleware/log"
	"github.com/ybizeul/hupload/pkg/apiws/storage"
)

type APIWS struct {
	StaticUI fs.FS
	HTTPPort int
	mux      *http.ServeMux

	TemplateData any

	Storage        storage.StorageInterface
	Authentication authentication.AuthenticationInterface
}

// New creates a new API Web Server. staticUI is the file system containing the
// web root directory.
func New(staticUI fs.FS, t any) (*APIWS, error) {
	d, err := fs.ReadDir(staticUI, ".")
	if err != nil {
		return nil, err
	}
	f, err := fs.Sub(staticUI, d[0].Name())
	if err != nil {
		return nil, err
	}
	return &APIWS{
		StaticUI:     f,
		HTTPPort:     8080,
		TemplateData: t,
		mux:          http.NewServeMux(),
	}, nil
}

func (a *APIWS) SetStorage(b storage.StorageInterface) {
	a.Storage = b
}

func (a *APIWS) SetAuthentication(b authentication.AuthenticationInterface) {
	a.Authentication = b
}

// AddRoute adds a new route to the API Web Server. pattern is the URL pattern
// to match. authenticators is a list of Authenticator to use to authenticate
// the request. handlerFunc is the function to call when the route is matched.
func (a *APIWS) AddRoute(pattern string, authenticators []auth.AuthMiddleware, handlerFunc func(w http.ResponseWriter, r *http.Request)) {
	if authenticators == nil {
		a.mux.HandleFunc(pattern, handlerFunc)
	} else {
		var h http.Handler
		h = http.HandlerFunc(handlerFunc)
		c := auth.ConfirmAuthenticator{Realm: "Hupload"}
		h = c.Middleware(h)
		for i := range authenticators {
			h = authenticators[len(authenticators)-1-i].Middleware(h)
		}
		a.mux.Handle(pattern, h)
	}
}

// Start starts the API Web Server.
func (a *APIWS) Start() {

	a.mux.HandleFunc("GET /{path...}", func(w http.ResponseWriter, r *http.Request) {
		_, err := fs.Stat(a.StaticUI, r.URL.Path[1:])
		if err == nil {
			http.ServeFileFS(w, r, a.StaticUI, r.URL.Path[1:])
		} else {
			tmpl, err := template.New("index.html").ParseFS(a.StaticUI, "index.html")
			if err != nil {
				slog.Error("unable to parse template", slog.String("error", err.Error()))
			}
			err = tmpl.Execute(w, a.TemplateData)
			if err != nil {
				slog.Error("unable to execute template", slog.String("error", err.Error()))
			}
		}
	})

	slog.Info(fmt.Sprintf("Starting web service on port %d", a.HTTPPort))
	err := http.ListenAndServe(fmt.Sprintf(":%d", a.HTTPPort), logger.NewLogger(a.mux))
	if err != nil {
		slog.Error("unable to start http server", slog.String("error", err.Error()))
	}
}
