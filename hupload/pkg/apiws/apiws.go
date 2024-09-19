package apiws

import (
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"

	"github.com/ybizeul/hupload/pkg/apiws/authentication"
	"github.com/ybizeul/hupload/pkg/apiws/middleware/auth"
	logger "github.com/ybizeul/hupload/pkg/apiws/middleware/log"
	"gopkg.in/square/go-jose.v2/json"
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
	TemplateData any

	// Authentication is the authentication backend
	Authentication authentication.Authentication
}

// New creates a new API Web Server. staticUI is the file system containing the
// web root directory.
func New(staticUI fs.FS, templateData any) (*APIWS, error) {
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
		TemplateData: templateData,
		Mux:          http.NewServeMux(),
	}

	if result.StaticUI != nil {
		result.Mux.HandleFunc("GET /{path...}", func(w http.ResponseWriter, r *http.Request) {
			_, err := fs.Stat(result.StaticUI, r.URL.Path[1:])
			if err == nil {
				if path.Ext(r.URL.Path) == ".html" {
					tmpl, err := template.New(r.URL.Path[1:]).ParseFS(result.StaticUI, r.URL.Path[1:])
					if err != nil {
						slog.Error("unable to parse template", slog.String("error", err.Error()))
					}
					err = tmpl.Execute(w, result.TemplateData)
					if err != nil {
						slog.Error("unable to execute template", slog.String("error", err.Error()))
					}
				} else {
					http.ServeFileFS(w, r, result.StaticUI, r.URL.Path[1:])
				}
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

	result.AddPublicRoute("GET /auth", nil, func(w http.ResponseWriter, r *http.Request) {
		user, _ := auth.AuthForRequest(r)
		response := struct {
			User          string `json:"user"`
			ShowLoginForm bool   `json:"showLoginForm"`
			LoginURL      string `json:"loginUrl"`
		}{
			User:          user,
			ShowLoginForm: result.Authentication.ShowLoginForm(),
			LoginURL:      result.Authentication.LoginURL(),
		}
		_ = json.NewEncoder(w).Encode(response)
	})

	return result, nil
}

// SetAuthentication
func (a *APIWS) SetAuthentication(b authentication.Authentication) {
	a.Authentication = b
}

// AddRoute adds a new route to the API Web Server. pattern is the URL pattern
// to match. authenticators is a list of Authenticator to use to authenticate
// the request. handlerFunc is the function to call when the route is matched.
func (a *APIWS) AddRoute(pattern string, authenticator auth.AuthMiddleware, handlerFunc func(w http.ResponseWriter, r *http.Request)) {
	j := auth.JWTAuthMiddleware{
		HMACSecret: os.Getenv("JWT_SECRET"),
	}
	c := auth.ConfirmAuthenticator{Realm: "Hupload"}
	a.Mux.Handle(pattern,
		authenticator.Middleware(
			j.Middleware(
				c.Middleware(http.HandlerFunc(handlerFunc)))))
}

func (a *APIWS) AddPublicRoute(pattern string, authenticator auth.AuthMiddleware, handlerFunc func(w http.ResponseWriter, r *http.Request)) {
	j := auth.JWTAuthMiddleware{
		HMACSecret: os.Getenv("JWT_SECRET"),
	}
	c := auth.ConfirmAuthenticator{Realm: "Hupload"}
	o := auth.OpenAuthMiddleware{}

	if authenticator == nil {
		a.Mux.Handle(pattern,
			j.Middleware(
				o.Middleware(
					c.Middleware(http.HandlerFunc(handlerFunc)))))
	} else {
		a.Mux.Handle(pattern,
			authenticator.Middleware(
				j.Middleware(
					o.Middleware(
						c.Middleware(http.HandlerFunc(handlerFunc))))))
	}
}

// Start starts the API Web Server.
func (a *APIWS) Start() {
	slog.Info(fmt.Sprintf("Starting web service on port %d", a.HTTPPort))

	// Check if we have a callback function for this authentication
	if _, ok := a.Authentication.CallbackFunc(nil); ok {
		// If there is, define action to redirect to "/shares"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s, ok := r.Context().Value(authentication.AuthStatusKey).(authentication.AuthStatus)
			if ok && s.Authenticated {
				http.Redirect(w, r, "/shares", http.StatusFound)
				return
			}
			if r.URL.Query().Get("error") != "" {
				http.Error(w, r.URL.Query().Get("error"), http.StatusUnauthorized)
				return
			}
		})
		m := auth.NewJWTAuthMiddleware(os.Getenv("JWT_SECRET"))
		f, _ := a.Authentication.CallbackFunc(m.Middleware(handler))
		a.Mux.HandleFunc("GET /oidc", f)
	}

	err := http.ListenAndServe(fmt.Sprintf(":%d", a.HTTPPort), logger.NewLogger(a.Mux))
	if err != nil {
		slog.Error("unable to start http server", slog.String("error", err.Error()))
	}
}
