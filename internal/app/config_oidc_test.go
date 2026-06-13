package app

import (
	"strings"
	"testing"
)

func TestLoadConfig_OIDCDefaults(t *testing.T) {
	cfg, _, err := LoadConfig(writeTestConfig(t, "{}"))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.OIDC.Enabled {
		t.Fatalf("OIDC.Enabled = true, want false")
	}
	if cfg.OIDC.DisplayName != "Login with OIDC" {
		t.Fatalf("OIDC.DisplayName = %q, want Login with OIDC", cfg.OIDC.DisplayName)
	}
	if cfg.OIDC.CallbackPath != "/api/user/auth/oidc/callback" {
		t.Fatalf("OIDC.CallbackPath = %q, want /api/user/auth/oidc/callback", cfg.OIDC.CallbackPath)
	}
	if cfg.OIDC.UserMapping.SubjectClaim != "sub" {
		t.Fatalf("OIDC.UserMapping.SubjectClaim = %q, want sub", cfg.OIDC.UserMapping.SubjectClaim)
	}
	if cfg.OIDC.UserMapping.EmailClaim != "email" {
		t.Fatalf("OIDC.UserMapping.EmailClaim = %q, want email", cfg.OIDC.UserMapping.EmailClaim)
	}
	if got, want := strings.Join(cfg.OIDC.Scopes, " "), "openid profile email"; got != want {
		t.Fatalf("OIDC.Scopes = %q, want %q", got, want)
	}
}

func TestLoadConfig_OIDCEnabledRequiresProviderFields(t *testing.T) {
	_, _, err := LoadConfig(writeTestConfig(t, `
oidc:
  enabled: true
`))
	if err == nil {
		t.Fatal("LoadConfig() error = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "oidc.issuer") ||
		!strings.Contains(err.Error(), "oidc.client-id") ||
		!strings.Contains(err.Error(), "oidc.client-secret") ||
		!strings.Contains(err.Error(), "oidc.redirect-url") {
		t.Fatalf("LoadConfig() error = %v, want missing oidc fields", err)
	}
}

func TestLoadConfig_OIDCEnabledAcceptsProviderFields(t *testing.T) {
	cfg, _, err := LoadConfig(writeTestConfig(t, `
oidc:
  enabled: true
  display-name: Sign in with Casdoor
  issuer: http://localhost:8000
  client-id: fns
  client-secret: secret
  redirect-url: http://localhost:9000/api/user/auth/oidc/callback
  auto-register: true
  user-mapping:
    subject-claim: sub
    email-claim: email
    username-claim: preferred_username
    display-name-claim: name
`))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if !cfg.OIDC.Enabled {
		t.Fatalf("OIDC.Enabled = false, want true")
	}
	if cfg.OIDC.DisplayName != "Sign in with Casdoor" {
		t.Fatalf("OIDC.DisplayName = %q", cfg.OIDC.DisplayName)
	}
	if !cfg.OIDC.AutoRegister {
		t.Fatalf("OIDC.AutoRegister = false, want true")
	}
}
