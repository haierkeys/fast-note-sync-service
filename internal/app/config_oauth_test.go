package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestConfig(t *testing.T, body string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func TestLoadConfig_OAuthDefaults(t *testing.T) {
	cfg, _, err := LoadConfig(writeTestConfig(t, "{}"))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.OAuth.Enabled {
		t.Fatalf("OAuth.Enabled = true, want false")
	}
	if !cfg.OAuth.AllowStaticFNSToken {
		t.Fatalf("OAuth.AllowStaticFNSToken = false, want true")
	}
	if cfg.OAuth.DefaultClient != "ChatGPT" {
		t.Fatalf("OAuth.DefaultClient = %q, want ChatGPT", cfg.OAuth.DefaultClient)
	}
	if cfg.OAuth.DefaultClientName != "ChatGPT" {
		t.Fatalf("OAuth.DefaultClientName = %q, want ChatGPT", cfg.OAuth.DefaultClientName)
	}
	if cfg.OAuth.SubjectMapping.Mode != "email_or_fixed_uid" {
		t.Fatalf("OAuth.SubjectMapping.Mode = %q, want email_or_fixed_uid", cfg.OAuth.SubjectMapping.Mode)
	}
	if cfg.OAuth.SubjectMapping.Claim != "email" {
		t.Fatalf("OAuth.SubjectMapping.Claim = %q, want email", cfg.OAuth.SubjectMapping.Claim)
	}
	if got, want := strings.Join(cfg.OAuth.ScopesSupported, " "), "notes:read notes:write files:read files:write vaults:read"; got != want {
		t.Fatalf("OAuth.ScopesSupported = %q, want %q", got, want)
	}
	if got, want := strings.Join(cfg.OAuth.RequiredScopes, " "), "notes:read files:read vaults:read"; got != want {
		t.Fatalf("OAuth.RequiredScopes = %q, want %q", got, want)
	}
}

func TestLoadConfig_OAuthEnabledRequiresMetadataFields(t *testing.T) {
	_, _, err := LoadConfig(writeTestConfig(t, `
oauth:
  enabled: true
`))
	if err == nil {
		t.Fatal("LoadConfig() error = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "oauth.resource") ||
		!strings.Contains(err.Error(), "oauth.issuer") ||
		!strings.Contains(err.Error(), "oauth.jwks-url") ||
		!strings.Contains(err.Error(), "oauth.authorization-servers") {
		t.Fatalf("LoadConfig() error = %v, want missing oauth fields", err)
	}
}

func TestLoadConfig_OAuthEnabledAcceptsRequiredMetadataFields(t *testing.T) {
	cfg, _, err := LoadConfig(writeTestConfig(t, `
oauth:
  enabled: true
  resource: https://notes.example.test/api/mcp
  issuer: https://auth.example.test
  jwks-url: https://auth.example.test/jwks.json
  authorization-servers:
    - https://auth.example.test
  audience:
    - https://notes.example.test/api/mcp
`))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if !cfg.OAuth.Enabled {
		t.Fatalf("OAuth.Enabled = false, want true")
	}
	if cfg.OAuth.Resource != "https://notes.example.test/api/mcp" {
		t.Fatalf("OAuth.Resource = %q", cfg.OAuth.Resource)
	}
	if len(cfg.OAuth.AuthorizationServers) != 1 || cfg.OAuth.AuthorizationServers[0] != "https://auth.example.test" {
		t.Fatalf("OAuth.AuthorizationServers = %#v", cfg.OAuth.AuthorizationServers)
	}
}

func TestLoadConfig_StytchOAuthRequiresStytchFields(t *testing.T) {
	_, _, err := LoadConfig(writeTestConfig(t, `
oauth:
  enabled: true
  resource: https://notes.example.test/api/mcp
  issuer: https://auth.example.test
  jwks-url: https://auth.example.test/jwks.json
  authorization-servers:
    - https://auth.example.test
  stytch:
    enabled: true
`))
	if err == nil {
		t.Fatal("LoadConfig() error = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "oauth.stytch.domain") ||
		!strings.Contains(err.Error(), "oauth.stytch.project-id") ||
		!strings.Contains(err.Error(), "oauth.stytch.secret") {
		t.Fatalf("LoadConfig() error = %v, want missing stytch fields", err)
	}
}

func TestLoadConfig_StytchB2BRequiresMemberFields(t *testing.T) {
	_, _, err := LoadConfig(writeTestConfig(t, `
oauth:
  enabled: true
  resource: https://notes.example.test/api/mcp
  issuer: https://auth.example.test
  jwks-url: https://auth.example.test/jwks.json
  authorization-servers:
    - https://auth.example.test
  stytch:
    enabled: true
    kind: b2b
    domain: https://auth.example.test
    project-id: project-test
    secret: secret-test
`))
	if err == nil {
		t.Fatal("LoadConfig() error = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "oauth.stytch.organization-id") ||
		!strings.Contains(err.Error(), "oauth.stytch.member-id") {
		t.Fatalf("LoadConfig() error = %v, want missing b2b member fields", err)
	}
}

func TestLoadConfig_OAuthAllowStaticFNSTokenCanBeDisabled(t *testing.T) {
	cfg, _, err := LoadConfig(writeTestConfig(t, `
oauth:
  allow-static-fns-token: false
`))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.OAuth.AllowStaticFNSToken {
		t.Fatalf("OAuth.AllowStaticFNSToken = true, want false")
	}
}

func TestLoadConfig_SampleConfigOAuthSection(t *testing.T) {
	cfg, _, err := LoadConfig("../../config/config.yaml")
	if err != nil {
		t.Fatalf("LoadConfig(sample) error = %v", err)
	}
	if cfg.OAuth.Enabled {
		t.Fatalf("OAuth.Enabled = true, want false")
	}
	if !cfg.OAuth.AllowStaticFNSToken {
		t.Fatalf("OAuth.AllowStaticFNSToken = false, want true")
	}
}
