package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ybizeul/hupload/pkg/apiws/authentication"
)

type AuthMiddleware interface {
	Middleware(http.Handler) http.Handler
}

// serveNextAuthenticated adds a passes w and r to next middleware after adding
// successful authentication context key/value
func ServeNextAuthenticated(user string, next http.Handler, w http.ResponseWriter, r *http.Request) {
	s, ok := r.Context().Value(authentication.AuthStatusKey).(authentication.AuthStatus)
	if !ok {
		s = authentication.AuthStatus{}
	}
	if user != "" {
		s.User = user
	}
	s.Authenticated = true
	ctx := context.WithValue(r.Context(), authentication.AuthStatusKey, s)
	next.ServeHTTP(w, r.WithContext(ctx))

	// if user == "" {
	// 	next.ServeHTTP(w,
	// 		r.WithContext(
	// 			context.WithValue(
	// 				r.Context(),
	// 				AuthStatus,
	// 				AuthStatusSuccess,
	// 			),
	// 		),
	// 	)
	// } else {
	// 	next.ServeHTTP(w,
	// 		r.WithContext(
	// 			context.WithValue(
	// 				context.WithValue(
	// 					r.Context(),
	// 					AuthStatus,
	// 					AuthStatusSuccess),
	// 				AuthUser,
	// 				user,
	// 			),
	// 		),
	// 	)
	// }
}

// serveNextError adds a passes w and r to next middleware after adding
// failed authentication context key/value
// any previously defined err is wrapped around err
func ServeNextError(next http.Handler, w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		err = errors.New("unknown error")
	}
	s, ok := r.Context().Value(authentication.AuthStatusKey).(authentication.AuthStatus)
	var c context.Context
	if ok {
		s.Error = errors.Join(s.Error, err)
		s.Authenticated = false
	} else {
		s = authentication.AuthStatus{Error: err}
	}
	c = context.WithValue(r.Context(), authentication.AuthStatusKey, s)
	next.ServeHTTP(w, r.WithContext(c))
}

type ConfirmAuthenticator struct {
	Realm string
}

func (a *ConfirmAuthenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s, ok := r.Context().Value(authentication.AuthStatusKey).(authentication.AuthStatus)
		if ok && s.Authenticated {
			// if r.URL.Path == "/oidc" {
			// 	http.Redirect(w, r, "/shares", http.StatusFound)
			// 	return
			// }
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Add("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"%s\"", a.Realm))
		w.WriteHeader(http.StatusUnauthorized)
		if s.Error != nil {
			errs := struct {
				Errors []string `json:"errors"`
			}{
				Errors: strings.Split(s.Error.Error(), "\n"),
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

func AuthForRequest(r *http.Request) (string, bool) {
	s, ok := r.Context().Value(authentication.AuthStatusKey).(authentication.AuthStatus)
	if !ok {
		return "", false
	}
	return s.User, s.Authenticated
}
