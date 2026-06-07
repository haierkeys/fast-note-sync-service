package routers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/config"
)

func TestOAuthMetadataRoutes_Output(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	registerOAuthMetadataRoutesWithConfig(r, config.OAuthConfig{
		Enabled:              true,
		Resource:             "https://notes.example.test/api/mcp",
		AuthorizationServers: []string{"https://auth.example.test"},
		JWKSURL:              "https://auth.example.test/jwks.json",
		ScopesSupported:      []string{"notes:read", "files:read"},
		ResourceName:         "Fast Note Sync MCP",
	})

	for _, path := range []string{
		"/.well-known/oauth-protected-resource",
		"/.well-known/oauth-protected-resource/api/mcp",
	} {
		t.Run(path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200, body=%s", w.Code, w.Body.String())
			}
			if got := w.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type = %q, want application/json", got)
			}

			var body struct {
				Resource             string   `json:"resource"`
				AuthorizationServers []string `json:"authorization_servers"`
				JWKSURI              string   `json:"jwks_uri"`
				ScopesSupported      []string `json:"scopes_supported"`
				ResourceName         string   `json:"resource_name"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if body.Resource != "https://notes.example.test/api/mcp" {
				t.Fatalf("resource = %q", body.Resource)
			}
			if len(body.AuthorizationServers) != 1 || body.AuthorizationServers[0] != "https://auth.example.test" {
				t.Fatalf("authorization_servers = %#v", body.AuthorizationServers)
			}
			if body.JWKSURI != "https://auth.example.test/jwks.json" {
				t.Fatalf("jwks_uri = %q", body.JWKSURI)
			}
			if len(body.ScopesSupported) != 2 || body.ScopesSupported[0] != "notes:read" || body.ScopesSupported[1] != "files:read" {
				t.Fatalf("scopes_supported = %#v", body.ScopesSupported)
			}
			if body.ResourceName != "Fast Note Sync MCP" {
				t.Fatalf("resource_name = %q", body.ResourceName)
			}
		})
	}
}

func TestOAuthMetadataRoutes_DefaultFNSScopeOmitsScopesSupported(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	registerOAuthMetadataRoutesWithConfig(r, config.OAuthConfig{
		Enabled:              true,
		Resource:             "https://notes.example.test/api/mcp",
		AuthorizationServers: []string{"https://auth.example.test"},
		JWKSURL:              "https://auth.example.test/jwks.json",
		ScopesSupported:      []string{"notes:read"},
		DefaultFNSScope:      "p:mcp c:* f:*",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-protected-resource/api/mcp", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", w.Code, w.Body.String())
	}

	var body struct {
		ScopesSupported []string `json:"scopes_supported"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if len(body.ScopesSupported) != 0 {
		t.Fatalf("scopes_supported = %#v, want empty", body.ScopesSupported)
	}
}

func TestOAuthMetadataRoutes_Disabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	registerOAuthMetadataRoutesWithConfig(r, config.OAuthConfig{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-protected-resource", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}
