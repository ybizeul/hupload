package authentication

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

type AuthenticationOIDCConfig struct {
	ProviderURL  string `yaml:"provider_url"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURL  string `yaml:"redirect_url"`
}

type AuthenticationOIDC struct {
	Options AuthenticationOIDCConfig

	Provider *oidc.Provider
	Config   oauth2.Config
}

func NewAuthenticationOIDC(o AuthenticationOIDCConfig) (*AuthenticationOIDC, error) {
	var err error
	result := &AuthenticationOIDC{
		Options: o,
	}

	result.Provider, err = oidc.NewProvider(context.Background(), result.Options.ProviderURL)
	if err != nil {
		return nil, err
	}

	// Configure an OpenID Connect aware OAuth2 client.
	result.Config = oauth2.Config{
		ClientID:     result.Options.ClientID,
		ClientSecret: result.Options.ClientSecret,
		RedirectURL:  result.Options.RedirectURL,

		// Discovery returns the OAuth2 endpoints.
		Endpoint: result.Provider.Endpoint(),

		// "openid" is a required scope for OpenID Connect flows.
		Scopes: []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return result, nil
}

func (o *AuthenticationOIDC) AuthenticateRequest(w http.ResponseWriter, r *http.Request, cb func(bool, error)) {
	if r.URL.Path == "/login" {
		http.Redirect(w, r, o.Config.AuthCodeURL("state"), http.StatusFound)
	} else {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(o.Config.AuthCodeURL("state")))
	}
	cb(false, ErrAuthenticationRedirect)
}

func (o *AuthenticationOIDC) CallbackFunc() (func(w http.ResponseWriter, r *http.Request), bool) {
	return func(w http.ResponseWriter, r *http.Request) {
		var verifier = o.Provider.Verifier(&oidc.Config{ClientID: o.Options.ClientID})

		// Verify state and errors.

		oauth2Token, err := o.Config.Exchange(r.Context(), r.URL.Query().Get("code"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(err)
			return
		}

		// Extract the ID Token from OAuth2 token.
		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(err)
			return
		}

		// Parse and verify ID Token payload.
		idToken, err := verifier.Verify(r.Context(), rawIDToken)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(err)
			return
		}

		// Extract custom claims
		var claims struct {
			Email    string `json:"email"`
			Verified bool   `json:"email_verified"`
		}
		if err := idToken.Claims(&claims); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(err)
			return
		}
	}, true
}
