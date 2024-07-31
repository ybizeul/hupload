package apiws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

type AuthMiddleware interface {
	Middleware(http.Handler) http.Handler
}

type ContextValue string

const (
	AuthStatus        ContextValue = "Authenticated"
	AuthUser          ContextValue = "User"
	AuthError         ContextValue = "Error"
	AuthStatusSuccess              = "Success"
)

// serveNextAuthenticated adds a passes w and r to next middleware after adding
// successful authentication context key/value
func serveNextAuthenticated(user string, next http.Handler, w http.ResponseWriter, r *http.Request) {
	next.ServeHTTP(w,
		r.WithContext(
			context.WithValue(
				context.WithValue(
					r.Context(),
					AuthStatus,
					AuthStatusSuccess),
				AuthUser,
				user,
			),
		),
	)
}

// serveNextError adds a passes w and r to next middleware after adding
// failed authentication context key/value
// any previously defined err is wrapped around err
func serveNextError(next http.Handler, w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		err = errors.New("unknown error")
	}

	e, ok := r.Context().Value(AuthError).(error)
	var c context.Context
	if ok {
		c = context.WithValue(r.Context(), AuthError, errors.Join(err, e))
	} else {
		c = context.WithValue(r.Context(), AuthError, err)
	}
	next.ServeHTTP(w, r.WithContext(c))
}

type ConfirmAuthenticator struct {
	Realm string
}

func (a *ConfirmAuthenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Context().Value(AuthStatus) == AuthStatusSuccess {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Add("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"%s\"", a.Realm))
		w.WriteHeader(http.StatusUnauthorized)
		e, ok := r.Context().Value(AuthError).(error)
		if ok {
			errs := struct {
				Errors []string `json:"errors"`
			}{
				Errors: strings.Split(e.Error(), "\n"),
			}
			b, _ := json.Marshal(errs)

			slog.Error("authentication failed", slog.Any("errors", b))

			_, _ = w.Write(b)
		} else {
			slog.Error("authentication failed")
			_, _ = w.Write([]byte("Unauthorized"))
			return
		}

	})
}
