package oidc

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestProviderAuthCodeURLIncludesOIDCParameters(t *testing.T) {
	server, _ := newTestOIDCProvider(t, "nonce-1")
	defer server.Close()

	provider, err := NewProvider(context.Background(), ProviderConfig{
		Issuer:       server.URL,
		ClientID:     "fns",
		ClientSecret: "secret",
		RedirectURL:  "https://fns.example.com/api/user/auth/oidc/callback",
		Scopes:       []string{"openid", "email"},
	})
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	authURL := provider.AuthCodeURL("state-1", "nonce-1", "verifier-1")
	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("Parse(%q) error = %v", authURL, err)
	}
	values := parsed.Query()
	if got, want := parsed.Scheme+"://"+parsed.Host+parsed.Path, server.URL+"/authorize"; got != want {
		t.Fatalf("authorize endpoint = %q, want %q", got, want)
	}
	if values.Get("client_id") != "fns" {
		t.Fatalf("client_id = %q", values.Get("client_id"))
	}
	if values.Get("response_type") != "code" {
		t.Fatalf("response_type = %q", values.Get("response_type"))
	}
	if values.Get("scope") != "openid email" {
		t.Fatalf("scope = %q", values.Get("scope"))
	}
	if values.Get("state") != "state-1" || values.Get("nonce") != "nonce-1" {
		t.Fatalf("state/nonce = %q/%q", values.Get("state"), values.Get("nonce"))
	}
	if values.Get("code_challenge") == "" || values.Get("code_challenge_method") != "S256" {
		t.Fatalf("PKCE params missing: %s", values.Encode())
	}
}

func TestProviderExchangeVerifiesIDTokenAndNonce(t *testing.T) {
	server, _ := newTestOIDCProvider(t, "nonce-1")
	defer server.Close()

	provider, err := NewProvider(context.Background(), ProviderConfig{
		Issuer:       server.URL,
		ClientID:     "fns",
		ClientSecret: "secret",
		RedirectURL:  "https://fns.example.com/api/user/auth/oidc/callback",
		Scopes:       []string{"openid", "email", "profile"},
	})
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	claims, err := provider.Exchange(context.Background(), "code-1", "verifier-1", "nonce-1")
	if err != nil {
		t.Fatalf("Exchange() error = %v", err)
	}
	if claims.Subject != "oidc-subject-1" {
		t.Fatalf("claims.Subject = %q", claims.Subject)
	}
	if claims.Email != "oidc@example.com" {
		t.Fatalf("claims.Email = %q", claims.Email)
	}
	if claims.Username != "oidc-user" {
		t.Fatalf("claims.Username = %q", claims.Username)
	}
	if claims.DisplayName != "OIDC User" {
		t.Fatalf("claims.DisplayName = %q", claims.DisplayName)
	}
}

func TestProviderExchangeRejectsNonceMismatch(t *testing.T) {
	server, _ := newTestOIDCProvider(t, "nonce-from-token")
	defer server.Close()

	provider, err := NewProvider(context.Background(), ProviderConfig{
		Issuer:       server.URL,
		ClientID:     "fns",
		ClientSecret: "secret",
		RedirectURL:  "https://fns.example.com/api/user/auth/oidc/callback",
		Scopes:       []string{"openid", "email"},
	})
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	if _, err := provider.Exchange(context.Background(), "code-1", "verifier-1", "expected-nonce"); err == nil {
		t.Fatal("Exchange() error = nil, want nonce mismatch")
	}
}

func newTestOIDCProvider(t *testing.T, nonce string) (*httptest.Server, *rsa.PrivateKey) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	kid := "test-key"

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			writeJSON(t, w, map[string]any{
				"issuer":                 server.URL,
				"authorization_endpoint": server.URL + "/authorize",
				"token_endpoint":         server.URL + "/token",
				"jwks_uri":               server.URL + "/jwks",
			})
		case "/jwks":
			writeJSON(t, w, map[string]any{
				"keys": []map[string]any{{
					"kty": "RSA",
					"use": "sig",
					"alg": "RS256",
					"kid": kid,
					"n":   base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
					"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.PublicKey.E)).Bytes()),
				}},
			})
		case "/token":
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm() error = %v", err)
			}
			if r.Form.Get("code") != "code-1" {
				t.Fatalf("token code = %q", r.Form.Get("code"))
			}
			token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
				"iss":                server.URL,
				"aud":                "fns",
				"sub":                "oidc-subject-1",
				"email":              "oidc@example.com",
				"preferred_username": "oidc-user",
				"name":               "OIDC User",
				"nonce":              nonce,
				"iat":                time.Now().Unix(),
				"exp":                time.Now().Add(time.Hour).Unix(),
			})
			token.Header["kid"] = kid
			signed, err := token.SignedString(key)
			if err != nil {
				t.Fatalf("SignedString() error = %v", err)
			}
			writeJSON(t, w, map[string]any{
				"access_token": "access-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
				"id_token":     signed,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	return server, key
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
}
