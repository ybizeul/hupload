package config

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ybizeul/apiws/auth"
	authmiddleware "github.com/ybizeul/apiws/auth/middleware"
	"github.com/ybizeul/apiws/auth/oidc"
)

var ErrInvalidAPIKey = errors.New("invalid API key")

// API key auth reuses existing backends and short-circuits with an authenticated
// context when a valid API key is provided.
type apiKeyAuth struct {
	next auth.Authentication
	keys map[string]struct{}
}

func newAPIKeyAuth(next auth.Authentication, keys []string) *apiKeyAuth {
	return &apiKeyAuth{next: next, keys: keySet(keys)}
}

func (a *apiKeyAuth) AuthMiddleware(next http.Handler) http.Handler {
	wrapped := next
	if a.next != nil {
		wrapped = a.next.AuthMiddleware(next)
	}

	if len(a.keys) == 0 {
		return wrapped
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey, present := apiKeyForRequest(r)
		if !present {
			wrapped.ServeHTTP(w, r)
			return
		}

		if _, ok := a.keys[apiKey]; !ok {
			authmiddleware.ServeNextError(next, w, r, ErrInvalidAPIKey)
			return
		}

		authmiddleware.ServeNextAuthenticated("api-key", next, w, r)
	})
}

// oidcAPIKeyAuth keeps OIDC optional interfaces available to apiws so login,
// logout and callback handlers keep working when API keys are enabled.
type oidcAPIKeyAuth struct {
	*apiKeyAuth
	oidc *oidc.OIDC
}

func newOIDCAPIKeyAuth(oidcBackend *oidc.OIDC, keys []string) *oidcAPIKeyAuth {
	return &oidcAPIKeyAuth{
		apiKeyAuth: newAPIKeyAuth(oidcBackend, keys),
		oidc:       oidcBackend,
	}
}

func (a *oidcAPIKeyAuth) CallbackHandler(h http.Handler) (pattern string, handler http.Handler) {
	return a.oidc.CallbackHandler(h)
}

func (a *oidcAPIKeyAuth) LoginHandler() (path string, skipForm bool, h http.Handler) {
	return a.oidc.LoginHandler()
}

func (a *oidcAPIKeyAuth) LogoutURL() string {
	return a.oidc.LogoutURL()
}

func keySet(keys []string) map[string]struct{} {
	result := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		result[key] = struct{}{}
	}
	return result
}

func apiKeyForRequest(r *http.Request) (string, bool) {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if authHeader == "" {
		return "", false
	}

	parts := strings.Fields(authHeader)
	if len(parts) != 2 {
		return "", false
	}

	scheme := strings.ToLower(parts[0])
	if scheme != "bearer" {
		return "", false
	}

	key := strings.TrimSpace(parts[1])
	if key == "" {
		return "", false
	}

	return key, true
}
