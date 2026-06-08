package oauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStytchClient_AuthorizeStart_sendsOAuthRequest(t *testing.T) {
	var gotPath string
	var gotAuth string
	var gotBody stytchAuthorizeRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("Decode request body error = %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"request_id":"req-1",
			"user_id":"fns:1",
			"client":{"client_id":"client-1","client_name":"ChatGPT"},
			"consent_required":true,
			"scope_results":[{"scope":"notes:read","description":"Read notes","is_grantable":true}],
			"status_code":200
		}`))
	}))
	defer server.Close()

	client := NewStytchClient(StytchClientConfig{
		Domain:    server.URL,
		ProjectID: "project-test",
		Secret:    "secret-test",
	})

	resp, err := client.AuthorizeStart(context.Background(), StytchAuthorizeParams{
		ClientID:            "client-1",
		RedirectURI:         "https://chatgpt.com/connector/oauth/callback",
		ResponseType:        "code",
		Scopes:              []string{"openid", "email", "notes:read"},
		State:               "state-1",
		CodeChallenge:       "challenge-1",
		CodeChallengeMethod: "S256",
		Resource:            "https://obsidian-fns.kahub.in/api/mcp",
		UserID:              "fns:1",
	})
	if err != nil {
		t.Fatalf("AuthorizeStart() error = %v", err)
	}

	if gotPath != "/v1/idp/oauth/authorize/start" {
		t.Fatalf("request path = %q, want /v1/idp/oauth/authorize/start", gotPath)
	}
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("project-test:secret-test"))
	if gotAuth != wantAuth {
		t.Fatalf("Authorization header = %q, want %q", gotAuth, wantAuth)
	}
	if gotBody.ClientID != "client-1" ||
		gotBody.RedirectURI != "https://chatgpt.com/connector/oauth/callback" ||
		gotBody.ResponseType != "code" ||
		gotBody.UserID != "fns:1" {
		t.Fatalf("request body = %#v, missing OAuth parameters", gotBody)
	}
	if gotBody.State != "" || gotBody.CodeChallenge != "" || len(gotBody.Resources) != 0 {
		t.Fatalf("request body = %#v, start request should not include submit-only OAuth parameters", gotBody)
	}
	if len(gotBody.Scopes) != 3 || gotBody.Scopes[2] != "notes:read" {
		t.Fatalf("request scopes = %#v, want requested scopes", gotBody.Scopes)
	}
	if !resp.ConsentRequired || len(resp.ScopeResults) != 1 {
		t.Fatalf("AuthorizeStart() response = %#v, want consent details", resp)
	}
}

func TestStytchClient_AuthorizeSubmit_returnsRedirectURI(t *testing.T) {
	var gotBody stytchAuthorizeRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/idp/oauth/authorize" {
			t.Fatalf("request path = %q, want /v1/idp/oauth/authorize", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("Decode request body error = %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"request_id":"req-2",
			"redirect_uri":"https://chatgpt.com/connector/oauth/callback?code=code-1&state=state-1",
			"authorization_code":"code-1",
			"status_code":200
		}`))
	}))
	defer server.Close()

	client := NewStytchClient(StytchClientConfig{
		Domain:    server.URL,
		ProjectID: "project-test",
		Secret:    "secret-test",
	})

	resp, err := client.AuthorizeSubmit(context.Background(), StytchAuthorizeParams{
		ClientID:            "client-1",
		RedirectURI:         "https://chatgpt.com/connector/oauth/callback",
		ResponseType:        "code",
		Scopes:              []string{"openid", "email", "notes:read"},
		State:               "state-1",
		CodeChallenge:       "challenge-1",
		CodeChallengeMethod: "S256",
		Resource:            "https://obsidian-fns.kahub.in/api/mcp",
		UserID:              "fns:1",
		ConsentGranted:      true,
	})
	if err != nil {
		t.Fatalf("AuthorizeSubmit() error = %v", err)
	}

	if !gotBody.ConsentGranted ||
		gotBody.State != "state-1" ||
		gotBody.CodeChallenge != "challenge-1" ||
		len(gotBody.Resources) != 1 ||
		gotBody.Resources[0] != "https://obsidian-fns.kahub.in/api/mcp" {
		t.Fatalf("request body = %#v, want consent and PKCE parameters", gotBody)
	}
	if resp.RedirectURI != "https://chatgpt.com/connector/oauth/callback?code=code-1&state=state-1" {
		t.Fatalf("RedirectURI = %q", resp.RedirectURI)
	}
}

func TestStytchClient_AuthorizeStart_sendsB2BOAuthRequest(t *testing.T) {
	var gotPath string
	var gotBody stytchAuthorizeRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("Decode request body error = %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"request_id":"req-3",
			"member_id":"member-live-1",
			"client":{"client_id":"client-1","client_name":"ChatGPT"},
			"consent_required":true,
			"scope_results":[{"scope":"notes:read","description":"Read notes","is_grantable":true}],
			"status_code":200
		}`))
	}))
	defer server.Close()

	client := NewStytchClient(StytchClientConfig{
		Domain:    server.URL,
		ProjectID: "project-test",
		Secret:    "secret-test",
		Kind:      StytchKindB2B,
	})

	_, err := client.AuthorizeStart(context.Background(), StytchAuthorizeParams{
		ClientID:       "client-1",
		RedirectURI:    "https://chatgpt.com/connector/oauth/callback",
		ResponseType:   "code",
		Scopes:         []string{"openid", "email", "notes:read"},
		OrganizationID: "organization-live-1",
		MemberID:       "member-live-1",
	})
	if err != nil {
		t.Fatalf("AuthorizeStart() error = %v", err)
	}

	if gotPath != "/v1/b2b/idp/oauth/authorize/start" {
		t.Fatalf("request path = %q, want /v1/b2b/idp/oauth/authorize/start", gotPath)
	}
	if gotBody.UserID != "" {
		t.Fatalf("UserID = %q, want empty for B2B request", gotBody.UserID)
	}
	if gotBody.OrganizationID != "organization-live-1" || gotBody.MemberID != "member-live-1" {
		t.Fatalf("request body = %#v, want organization_id and member_id", gotBody)
	}
}
