package api_router

import (
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/config"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
)

func TestBuildStytchAuthorizeParams_mapsFNSUserToExternalID(t *testing.T) {
	params := buildStytchAuthorizeParams(config.OAuthConfig{
		Resource: "https://obsidian-fns.kahub.in/api/mcp",
		Stytch: config.StytchOAuthConfig{
			UserIDPrefix: "fns:",
		},
	}, StytchOAuthAuthorizeRequest{
		ClientID:            "client-1",
		RedirectURI:         "https://chatgpt.com/connector/oauth/callback",
		Scope:               "openid email notes:read files:read",
		State:               "state-1",
		CodeChallenge:       "challenge-1",
		CodeChallengeMethod: "S256",
	}, &dto.UserDTO{UID: 1, Email: "user@example.com"}, true)

	if params.UserID != "fns:1" {
		t.Fatalf("UserID = %q, want fns:1", params.UserID)
	}
	if params.ResponseType != "code" {
		t.Fatalf("ResponseType = %q, want code", params.ResponseType)
	}
	if params.Resource != "https://obsidian-fns.kahub.in/api/mcp" {
		t.Fatalf("Resource = %q", params.Resource)
	}
	if len(params.Scopes) != 4 || params.Scopes[2] != "notes:read" {
		t.Fatalf("Scopes = %#v, want parsed OAuth scopes", params.Scopes)
	}
	if !params.ConsentGranted {
		t.Fatalf("ConsentGranted = false, want true")
	}
}

func TestBuildStytchAuthorizeParams_usesConfiguredConsumerUserID(t *testing.T) {
	params := buildStytchAuthorizeParams(config.OAuthConfig{
		Resource: "https://obsidian-fns.kahub.in/api/mcp",
		Stytch: config.StytchOAuthConfig{
			Kind:   "consumer",
			UserID: "user-live-e49e451d-ce2b-472b-823b-c0cd9e7ebb7b",
		},
	}, StytchOAuthAuthorizeRequest{
		ClientID:    "client-1",
		RedirectURI: "https://chatgpt.com/connector/oauth/callback",
		Scope:       "openid email notes:read",
	}, &dto.UserDTO{UID: 1, Email: "user@example.com"}, true)

	if params.UserID != "user-live-e49e451d-ce2b-472b-823b-c0cd9e7ebb7b" {
		t.Fatalf("UserID = %q, want configured Stytch user ID", params.UserID)
	}
}

func TestBuildStytchAuthorizeParams_usesB2BMemberMapping(t *testing.T) {
	params := buildStytchAuthorizeParams(config.OAuthConfig{
		Resource: "https://obsidian-fns.kahub.in/api/mcp",
		Stytch: config.StytchOAuthConfig{
			Kind:           "b2b",
			OrganizationID: "organization-live-b2053a81-7b99-41e2-bbdd-1b012eeb38cd",
			MemberID:       "member-live-fc16613a-67d4-4878-80b9-77371fad937e",
		},
	}, StytchOAuthAuthorizeRequest{
		ClientID:    "client-1",
		RedirectURI: "https://chatgpt.com/connector/oauth/callback",
		Scope:       "openid email notes:read",
	}, &dto.UserDTO{UID: 1, Email: "user@example.com"}, true)

	if params.UserID != "" {
		t.Fatalf("UserID = %q, want empty for B2B", params.UserID)
	}
	if params.OrganizationID != "organization-live-b2053a81-7b99-41e2-bbdd-1b012eeb38cd" {
		t.Fatalf("OrganizationID = %q", params.OrganizationID)
	}
	if params.MemberID != "member-live-fc16613a-67d4-4878-80b9-77371fad937e" {
		t.Fatalf("MemberID = %q", params.MemberID)
	}
}
