package apiws

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/ybizeul/hupload/pkg/apiws/authservice"
	"github.com/ybizeul/hupload/pkg/apiws/storageservice"
)

type APIWS struct {
	StaticUI fs.FS
	HTTPPort int
	mux      *http.ServeMux

	StorageService storageservice.StorageServiceInterface
	AuthService    authservice.AuthServiceInterface
}

// New creates a new API Web Server. staticUI is the file system containing the
// web root directory.
func New(staticUI fs.FS) (*APIWS, error) {
	d, err := fs.ReadDir(staticUI, ".")
	if err != nil {
		return nil, err
	}
	f, err := fs.Sub(staticUI, d[0].Name())
	if err != nil {
		return nil, err
	}
	return &APIWS{
		StaticUI: f,
		HTTPPort: 8080,
		mux:      http.NewServeMux(),
	}, nil
}

func (a *APIWS) SetStorageService(b storageservice.StorageServiceInterface) {
	a.StorageService = b
}

func (a *APIWS) SetAuthService(b authservice.AuthServiceInterface) {
	a.AuthService = b
}

// AddRoute adds a new route to the API Web Server. pattern is the URL pattern
// to match. authenticators is a list of Authenticator to use to authenticate
// the request. handlerFunc is the function to call when the route is matched.
func (a *APIWS) AddRoute(pattern string, authenticators []AuthMiddleware, handlerFunc func(w http.ResponseWriter, r *http.Request)) {
	if authenticators == nil {
		a.mux.HandleFunc(pattern, handlerFunc)
	} else {
		var h http.Handler
		h = http.HandlerFunc(handlerFunc)
		c := ConfirmAuthenticator{Realm: "Hupload"}
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
			http.ServeFileFS(w, r, a.StaticUI, "index.html")
		}
	})

	slog.Info(fmt.Sprintf("Starting web service on port %d", a.HTTPPort))
	err := http.ListenAndServe(fmt.Sprintf(":%d", a.HTTPPort), newLogger(a.mux))
	if err != nil {
		slog.Error("unable to start http server", slog.String("error", err.Error()))
	}
}

func UserForRequest(r *http.Request) string {
	user, ok := r.Context().Value(AuthUser).(string)
	if !ok {
		slog.Error("putShare", slog.String("error", "no user in context"))
		return ""
	}
	return user
}
