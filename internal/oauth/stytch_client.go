package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	StytchKindConsumer = "consumer"
	StytchKindB2B      = "b2b"
)

type StytchClientConfig struct {
	Domain     string
	ProjectID  string
	Secret     string
	Kind       string
	HTTPClient *http.Client
}

type StytchClient struct {
	domain    string
	projectID string
	secret    string
	kind      string
	client    *http.Client
}

type StytchAuthorizeParams struct {
	ClientID            string
	RedirectURI         string
	ResponseType        string
	Scopes              []string
	State               string
	Nonce               string
	CodeChallenge       string
	CodeChallengeMethod string
	Resource            string
	UserID              string
	OrganizationID      string
	MemberID            string
	ConsentGranted      bool
}

type StytchAuthorizeStartResponse struct {
	RequestID       string              `json:"request_id"`
	UserID          string              `json:"user_id"`
	Client          StytchOAuthClient   `json:"client"`
	ConsentRequired bool                `json:"consent_required"`
	ScopeResults    []StytchScopeResult `json:"scope_results"`
	StatusCode      int                 `json:"status_code"`
}

type StytchAuthorizeSubmitResponse struct {
	RequestID         string `json:"request_id"`
	RedirectURI       string `json:"redirect_uri"`
	AuthorizationCode string `json:"authorization_code"`
	StatusCode        int    `json:"status_code"`
}

type StytchOAuthClient struct {
	ClientID          string `json:"client_id"`
	ClientName        string `json:"client_name"`
	ClientDescription string `json:"client_description"`
	ClientType        string `json:"client_type"`
	LogoURL           string `json:"logo_url"`
}

type StytchScopeResult struct {
	Scope       string `json:"scope"`
	Description string `json:"description"`
	IsGrantable bool   `json:"is_grantable"`
}

type stytchAuthorizeRequest struct {
	ClientID            string   `json:"client_id"`
	RedirectURI         string   `json:"redirect_uri"`
	ResponseType        string   `json:"response_type"`
	Scopes              []string `json:"scopes"`
	UserID              string   `json:"user_id,omitempty"`
	OrganizationID      string   `json:"organization_id,omitempty"`
	MemberID            string   `json:"member_id,omitempty"`
	SessionToken        string   `json:"session_token,omitempty"`
	SessionJWT          string   `json:"session_jwt,omitempty"`
	Prompt              string   `json:"prompt,omitempty"`
	State               string   `json:"state,omitempty"`
	Nonce               string   `json:"nonce,omitempty"`
	CodeChallenge       string   `json:"code_challenge,omitempty"`
	Resources           []string `json:"resources,omitempty"`
	ConsentGranted      bool     `json:"consent_granted,omitempty"`
}

func NewStytchClient(cfg StytchClientConfig) *StytchClient {
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	kind := cfg.Kind
	if kind == "" {
		kind = StytchKindConsumer
	}

	return &StytchClient{
		domain:    strings.TrimRight(cfg.Domain, "/"),
		projectID: cfg.ProjectID,
		secret:    cfg.Secret,
		kind:      kind,
		client:    client,
	}
}

func (c *StytchClient) AuthorizeStart(ctx context.Context, params StytchAuthorizeParams) (*StytchAuthorizeStartResponse, error) {
	var out StytchAuthorizeStartResponse
	if err := c.post(ctx, c.authorizeStartPath(), params.toRequest(false), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *StytchClient) AuthorizeSubmit(ctx context.Context, params StytchAuthorizeParams) (*StytchAuthorizeSubmitResponse, error) {
	var out StytchAuthorizeSubmitResponse
	if err := c.post(ctx, c.authorizeSubmitPath(), params.toRequest(true), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (p StytchAuthorizeParams) toRequest(includeConsent bool) stytchAuthorizeRequest {
	req := stytchAuthorizeRequest{
		ClientID:            p.ClientID,
		RedirectURI:         p.RedirectURI,
		ResponseType:        p.ResponseType,
		Scopes:              append([]string(nil), p.Scopes...),
		UserID:              p.UserID,
		OrganizationID:      p.OrganizationID,
		MemberID:            p.MemberID,
	}
	if includeConsent {
		req.ConsentGranted = p.ConsentGranted
		req.State = p.State
		req.Nonce = p.Nonce
		req.CodeChallenge = p.CodeChallenge
		if p.Resource != "" {
			req.Resources = []string{p.Resource}
		}
	}
	return req
}

func (c *StytchClient) authorizeStartPath() string {
	if c.kind == StytchKindB2B {
		return "/v1/b2b/idp/oauth/authorize/start"
	}
	return "/v1/idp/oauth/authorize/start"
}

func (c *StytchClient) authorizeSubmitPath() string {
	if c.kind == StytchKindB2B {
		return "/v1/b2b/idp/oauth/authorize"
	}
	return "/v1/idp/oauth/authorize"
}

func (c *StytchClient) post(ctx context.Context, path string, body any, out any) error {
	if c.domain == "" || c.projectID == "" || c.secret == "" {
		return fmt.Errorf("stytch client config is incomplete")
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.domain+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.projectID, c.secret)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("stytch request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	if err := json.Unmarshal(responseBody, out); err != nil {
		return err
	}
	return nil
}
