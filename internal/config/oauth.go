package config

import (
	"fmt"
	"strings"

	mcpserver "github.com/mark3labs/mcp-go/server"
	"gopkg.in/yaml.v3"
)

const (
	OAuthSubjectMappingEmail           = "email"
	OAuthSubjectMappingFixedUID        = "fixed_uid"
	OAuthSubjectMappingEmailOrFixedUID = "email_or_fixed_uid"

	DefaultOAuthSubjectMappingMode  = OAuthSubjectMappingEmailOrFixedUID
	DefaultOAuthSubjectMappingClaim = "email"
)

var (
	DefaultOAuthScopesSupported = []string{"notes:read", "notes:write", "files:read", "files:write", "vaults:read"}
	DefaultOAuthRequiredScopes  = []string{"notes:read", "files:read", "vaults:read"}
)

type OAuthSubjectMappingConfig struct {
	Mode     string `yaml:"mode"`
	Claim    string `yaml:"claim"`
	FixedUID int64  `yaml:"fixed-uid"`
}

type StytchOAuthConfig struct {
	Enabled        bool   `yaml:"enabled"`
	Kind           string `yaml:"kind"`
	Domain         string `yaml:"domain"`
	ProjectID      string `yaml:"project-id"`
	Secret         string `yaml:"secret"`
	UserID         string `yaml:"user-id"`
	UserIDPrefix   string `yaml:"user-id-prefix"`
	OrganizationID string `yaml:"organization-id"`
	MemberID       string `yaml:"member-id"`
}

type OAuthConfig struct {
	Enabled              bool                      `yaml:"enabled" default:"false"`
	Resource             string                    `yaml:"resource"`
	AuthorizationServers []string                  `yaml:"authorization-servers"`
	JWKSURL              string                    `yaml:"jwks-url"`
	Issuer               string                    `yaml:"issuer"`
	Audience             []string                  `yaml:"audience"`
	ScopesSupported      []string                  `yaml:"scopes-supported"`
	RequiredScopes       []string                  `yaml:"required-scopes"`
	ResourceName         string                    `yaml:"resource-name"`
	AllowStaticFNSToken  bool                      `yaml:"allow-static-fns-token"`
	DefaultClient        string                    `yaml:"default-client"`
	DefaultClientName    string                    `yaml:"default-client-name"`
	DefaultClientVersion string                    `yaml:"default-client-version"`
	DefaultVaultName     string                    `yaml:"default-vault-name"`
	SubjectMapping       OAuthSubjectMappingConfig `yaml:"subject-mapping"`
	DefaultFNSScope      string                    `yaml:"default-fns-scope"`
	Stytch               StytchOAuthConfig         `yaml:"stytch"`

	allowStaticFNSTokenSet bool `yaml:"-"`
}

func (c *OAuthSubjectMappingConfig) SetDefaults() {
	if c.Mode == "" {
		c.Mode = DefaultOAuthSubjectMappingMode
	}
	if c.Claim == "" {
		c.Claim = DefaultOAuthSubjectMappingClaim
	}
}

func (c *OAuthConfig) SetDefaults() {
	c.Normalize()
}

func (c *OAuthConfig) Normalize() {
	if !c.allowStaticFNSTokenSet && !c.AllowStaticFNSToken {
		c.AllowStaticFNSToken = true
	}
	if c.DefaultClient == "" {
		c.DefaultClient = "ChatGPT"
	}
	if c.DefaultClientName == "" {
		c.DefaultClientName = "ChatGPT"
	}
	if len(c.ScopesSupported) == 0 {
		c.ScopesSupported = append([]string(nil), DefaultOAuthScopesSupported...)
	}
	if len(c.RequiredScopes) == 0 {
		c.RequiredScopes = append([]string(nil), DefaultOAuthRequiredScopes...)
	}
	c.SubjectMapping.SetDefaults()
	if c.Stytch.Kind == "" {
		c.Stytch.Kind = "consumer"
	}
	if c.Stytch.UserIDPrefix == "" {
		c.Stytch.UserIDPrefix = "fns:"
	}
}

func (c *OAuthConfig) UnmarshalYAML(value *yaml.Node) error {
	type rawOAuthConfig OAuthConfig

	var raw rawOAuthConfig
	if err := value.Decode(&raw); err != nil {
		return err
	}

	for i := 0; i+1 < len(value.Content); i += 2 {
		if value.Content[i].Value == "allow-static-fns-token" {
			raw.allowStaticFNSTokenSet = true
			break
		}
	}

	*c = OAuthConfig(raw)
	return nil
}

func (c OAuthConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	var missing []string
	if strings.TrimSpace(c.Resource) == "" {
		missing = append(missing, "oauth.resource")
	}
	if strings.TrimSpace(c.Issuer) == "" {
		missing = append(missing, "oauth.issuer")
	}
	if strings.TrimSpace(c.JWKSURL) == "" {
		missing = append(missing, "oauth.jwks-url")
	}
	if len(c.AuthorizationServers) == 0 {
		missing = append(missing, "oauth.authorization-servers")
	}
	if c.Stytch.Enabled {
		if strings.TrimSpace(c.Stytch.Domain) == "" {
			missing = append(missing, "oauth.stytch.domain")
		}
		if strings.TrimSpace(c.Stytch.ProjectID) == "" {
			missing = append(missing, "oauth.stytch.project-id")
		}
		if strings.TrimSpace(c.Stytch.Secret) == "" {
			missing = append(missing, "oauth.stytch.secret")
		}
		if c.Stytch.Kind != "consumer" && c.Stytch.Kind != "b2b" {
			missing = append(missing, "oauth.stytch.kind")
		}
		if c.Stytch.Kind == "b2b" {
			if strings.TrimSpace(c.Stytch.OrganizationID) == "" {
				missing = append(missing, "oauth.stytch.organization-id")
			}
			if strings.TrimSpace(c.Stytch.MemberID) == "" {
				missing = append(missing, "oauth.stytch.member-id")
			}
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required oauth config: %s", strings.Join(missing, ", "))
	}
	return nil
}

func (c OAuthConfig) ProtectedResourceMetadata() mcpserver.ProtectedResourceMetadataConfig {
	scopesSupported := c.ScopesSupported
	if strings.TrimSpace(c.DefaultFNSScope) != "" {
		scopesSupported = nil
	}
	return mcpserver.ProtectedResourceMetadataConfig{
		Resource:               c.Resource,
		AuthorizationServers:   append([]string(nil), c.AuthorizationServers...),
		ScopesSupported:        append([]string(nil), scopesSupported...),
		BearerMethodsSupported: []string{"header"},
		ResourceName:           c.ResourceName,
		JWKSURI:                c.JWKSURL,
	}
}
