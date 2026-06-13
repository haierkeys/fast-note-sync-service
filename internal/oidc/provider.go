package oidc

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	internaloauth "github.com/haierkeys/fast-note-sync-service/internal/oauth"
)

type ProviderConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	HTTPClient   *http.Client
}

type Provider struct {
	config   ProviderConfig
	metadata providerMetadata
	verifier *internaloauth.JWTVerifier
	client   *http.Client
}

type Claims struct {
	Subject     string
	Email       string
	Username    string
	DisplayName string
	Raw         map[string]interface{}
}

type providerMetadata struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
}

func NewProvider(ctx context.Context, config ProviderConfig) (*Provider, error) {
	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	metadata, err := discover(ctx, client, config.Issuer)
	if err != nil {
		return nil, err
	}
	return &Provider{
		config:   config,
		metadata: metadata,
		verifier: internaloauth.NewJWTVerifier(internaloauth.JWTVerifierConfig{
			Issuer:     metadata.Issuer,
			Audience:   config.ClientID,
			JWKSURL:    metadata.JWKSURI,
			HTTPClient: client,
		}),
		client: client,
	}, nil
}

func (p *Provider) AuthCodeURL(state, nonce, codeVerifier string) string {
	values := url.Values{}
	values.Set("client_id", p.config.ClientID)
	values.Set("redirect_uri", p.config.RedirectURL)
	values.Set("response_type", "code")
	values.Set("scope", strings.Join(p.config.Scopes, " "))
	values.Set("state", state)
	values.Set("nonce", nonce)
	values.Set("code_challenge", codeChallenge(codeVerifier))
	values.Set("code_challenge_method", "S256")
	return p.metadata.AuthorizationEndpoint + "?" + values.Encode()
}

func (p *Provider) Exchange(ctx context.Context, code, codeVerifier, nonce string) (*Claims, error) {
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("code", code)
	values.Set("redirect_uri", p.config.RedirectURL)
	values.Set("client_id", p.config.ClientID)
	values.Set("client_secret", p.config.ClientSecret)
	values.Set("code_verifier", codeVerifier)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.metadata.TokenEndpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("oidc token endpoint status %d", resp.StatusCode)
	}

	var tokenResp struct {
		IDToken string `json:"id_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}
	if strings.TrimSpace(tokenResp.IDToken) == "" {
		return nil, fmt.Errorf("oidc token response missing id_token")
	}

	verified, err := p.verifier.Verify(ctx, tokenResp.IDToken)
	if err != nil {
		return nil, err
	}
	rawNonce, _ := verified.Raw["nonce"].(string)
	if rawNonce != nonce {
		return nil, fmt.Errorf("oidc nonce mismatch")
	}

	return &Claims{
		Subject:     strings.TrimSpace(verified.Subject),
		Email:       stringClaim(verified.Raw, "email"),
		Username:    stringClaim(verified.Raw, "preferred_username"),
		DisplayName: stringClaim(verified.Raw, "name"),
		Raw:         verified.Raw,
	}, nil
}

func discover(ctx context.Context, client *http.Client, issuer string) (providerMetadata, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(issuer, "/")+"/.well-known/openid-configuration", nil)
	if err != nil {
		return providerMetadata{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return providerMetadata{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return providerMetadata{}, fmt.Errorf("oidc discovery status %d", resp.StatusCode)
	}

	var metadata providerMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return providerMetadata{}, err
	}
	if strings.TrimSpace(metadata.Issuer) == "" {
		metadata.Issuer = strings.TrimRight(issuer, "/")
	}
	if metadata.AuthorizationEndpoint == "" || metadata.TokenEndpoint == "" || metadata.JWKSURI == "" {
		return providerMetadata{}, fmt.Errorf("oidc discovery metadata is incomplete")
	}
	return metadata, nil
}

func codeChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func stringClaim(claims map[string]interface{}, name string) string {
	value, _ := claims[name].(string)
	return strings.TrimSpace(value)
}
