package config

import (
	"fmt"
	"regexp"
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
	Providers    []OIDCProviderConfig  `yaml:"providers"`
}

type OIDCProviderConfig struct {
	ID           string                `yaml:"id"`
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
	if len(c.Providers) == 0 && c.hasLegacyProvider() {
		c.Providers = []OIDCProviderConfig{{
			ID:           "default",
			DisplayName:  c.DisplayName,
			Issuer:       c.Issuer,
			ClientID:     c.ClientID,
			ClientSecret: c.ClientSecret,
			RedirectURL:  c.RedirectURL,
			CallbackPath: c.CallbackPath,
			Scopes:       c.Scopes,
			AutoRegister: c.AutoRegister,
			UserMapping:  c.UserMapping,
		}}
	}
	for i := range c.Providers {
		c.Providers[i].Normalize()
	}
}

func (c OIDCConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	if len(c.Providers) > 0 {
		return validateOIDCProviders(c.Providers)
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

func (c OIDCConfig) ProviderByID(id string) (OIDCProviderConfig, bool) {
	id = strings.TrimSpace(id)
	if id == "" && len(c.Providers) == 1 {
		return c.Providers[0], true
	}
	for _, provider := range c.Providers {
		if provider.ID == id {
			return provider, true
		}
	}
	return OIDCProviderConfig{}, false
}

func (c OIDCConfig) DefaultProvider() (OIDCProviderConfig, bool) {
	if len(c.Providers) == 0 {
		return OIDCProviderConfig{}, false
	}
	return c.Providers[0], true
}

func (c OIDCConfig) hasLegacyProvider() bool {
	return strings.TrimSpace(c.Issuer) != "" ||
		strings.TrimSpace(c.ClientID) != "" ||
		strings.TrimSpace(c.ClientSecret) != "" ||
		strings.TrimSpace(c.RedirectURL) != ""
}

func (c *OIDCProviderConfig) Normalize() {
	c.ID = strings.TrimSpace(c.ID)
	if c.ID == "" {
		c.ID = "default"
	}
	if c.DisplayName == "" {
		c.DisplayName = "Login with OIDC"
	}
	if c.CallbackPath == "" {
		c.CallbackPath = "/api/user/auth/oidc/callback/" + c.ID
		if c.ID == "default" {
			c.CallbackPath = "/api/user/auth/oidc/callback"
		}
	}
	if len(c.Scopes) == 0 {
		c.Scopes = []string{"openid", "profile", "email"}
	}
	c.UserMapping.SetDefaults()
}

var oidcProviderIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func validateOIDCProviders(providers []OIDCProviderConfig) error {
	seen := map[string]struct{}{}
	for i, provider := range providers {
		prefix := fmt.Sprintf("oidc.providers[%d]", i)
		var missing []string
		if strings.TrimSpace(provider.ID) == "" {
			missing = append(missing, prefix+".id")
		} else if !oidcProviderIDPattern.MatchString(provider.ID) {
			return fmt.Errorf("%s.id must contain only letters, numbers, underscores, or hyphens", prefix)
		} else if _, ok := seen[provider.ID]; ok {
			return fmt.Errorf("duplicate oidc provider id: %s", provider.ID)
		}
		seen[provider.ID] = struct{}{}
		if strings.TrimSpace(provider.Issuer) == "" {
			missing = append(missing, prefix+".issuer")
		}
		if strings.TrimSpace(provider.ClientID) == "" {
			missing = append(missing, prefix+".client-id")
		}
		if strings.TrimSpace(provider.ClientSecret) == "" {
			missing = append(missing, prefix+".client-secret")
		}
		if strings.TrimSpace(provider.RedirectURL) == "" {
			missing = append(missing, prefix+".redirect-url")
		}
		if len(missing) > 0 {
			return fmt.Errorf("missing required oidc config: %s", strings.Join(missing, ", "))
		}
	}
	return nil
}
