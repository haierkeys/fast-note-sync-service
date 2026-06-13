package config

import (
	"fmt"
	"strings"
)

type OIDCUserMappingConfig struct {
	SubjectClaim     string `yaml:"subject-claim"`
	EmailClaim       string `yaml:"email-claim"`
	UsernameClaim    string `yaml:"username-claim"`
	DisplayNameClaim string `yaml:"display-name-claim"`
}

type OIDCConfig struct {
	Enabled      bool                  `yaml:"enabled" default:"false"`
	DisplayName  string                `yaml:"display-name"`
	Issuer       string                `yaml:"issuer"`
	ClientID     string                `yaml:"client-id"`
	ClientSecret string                `yaml:"client-secret"`
	RedirectURL  string                `yaml:"redirect-url"`
	CallbackPath string                `yaml:"callback-path"`
	Scopes       []string              `yaml:"scopes"`
	AutoRegister bool                  `yaml:"auto-register"`
	UserMapping  OIDCUserMappingConfig `yaml:"user-mapping"`
}

func (c *OIDCUserMappingConfig) SetDefaults() {
	if c.SubjectClaim == "" {
		c.SubjectClaim = "sub"
	}
	if c.EmailClaim == "" {
		c.EmailClaim = "email"
	}
	if c.UsernameClaim == "" {
		c.UsernameClaim = "preferred_username"
	}
	if c.DisplayNameClaim == "" {
		c.DisplayNameClaim = "name"
	}
}

func (c *OIDCConfig) Normalize() {
	if c.DisplayName == "" {
		c.DisplayName = "Login with OIDC"
	}
	if c.CallbackPath == "" {
		c.CallbackPath = "/api/user/auth/oidc/callback"
	}
	if len(c.Scopes) == 0 {
		c.Scopes = []string{"openid", "profile", "email"}
	}
	c.UserMapping.SetDefaults()
}

func (c OIDCConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	var missing []string
	if strings.TrimSpace(c.Issuer) == "" {
		missing = append(missing, "oidc.issuer")
	}
	if strings.TrimSpace(c.ClientID) == "" {
		missing = append(missing, "oidc.client-id")
	}
	if strings.TrimSpace(c.ClientSecret) == "" {
		missing = append(missing, "oidc.client-secret")
	}
	if strings.TrimSpace(c.RedirectURL) == "" {
		missing = append(missing, "oidc.redirect-url")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required oidc config: %s", strings.Join(missing, ", "))
	}
	return nil
}
