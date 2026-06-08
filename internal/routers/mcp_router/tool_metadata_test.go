package mcp_router

import (
	"encoding/json"
	"testing"

	appconfig "github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestToolMetadataMarshalIncludesSecuritySchemesAndAnnotations(t *testing.T) {
	cfg := &appconfig.AppConfig{
		OAuth: config.OAuthConfig{
			Enabled:         true,
			ScopesSupported: []string{"notes:read", "notes:write"},
			RequiredScopes:  []string{"notes:read"},
		},
	}

	tools := []mcp.Tool{
		withMCPToolMetadata(mcp.NewTool("note_get"), cfg, mcpToolMetadata{
			ReadOnly:    true,
			Destructive: false,
			OpenWorld:   false,
			Scopes:      []string{"notes:read"},
		}),
		withMCPToolMetadata(mcp.NewTool("note_delete"), cfg, mcpToolMetadata{
			ReadOnly:    false,
			Destructive: true,
			OpenWorld:   false,
			Scopes:      []string{"notes:write"},
		}),
	}

	payload, err := json.Marshal(mcp.NewListToolsResult(tools, ""))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var body struct {
		Tools []struct {
			Name         string `json:"name"`
			OutputSchema struct {
				Type       string         `json:"type"`
				Properties map[string]any `json:"properties"`
				Required   []string       `json:"required"`
			} `json:"outputSchema"`
			Meta struct {
				SecuritySchemes []struct {
					Type   string   `json:"type"`
					Scopes []string `json:"scopes"`
				} `json:"securitySchemes"`
			} `json:"_meta"`
			Annotations struct {
				ReadOnlyHint    *bool `json:"readOnlyHint"`
				DestructiveHint *bool `json:"destructiveHint"`
				OpenWorldHint   *bool `json:"openWorldHint"`
			} `json:"annotations"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		t.Fatalf("Unmarshal() error = %v\npayload=%s", err, payload)
	}
	if len(body.Tools) != 2 {
		t.Fatalf("tools length = %d, want 2", len(body.Tools))
	}

	readTool := body.Tools[0]
	if readTool.OutputSchema.Type != "object" {
		t.Fatalf("read tool output schema type = %q, want object", readTool.OutputSchema.Type)
	}
	if len(readTool.Meta.SecuritySchemes) != 1 {
		t.Fatalf("read tool security schemes = %#v, want one OAuth scheme", readTool.Meta.SecuritySchemes)
	}
	if readTool.Meta.SecuritySchemes[0].Type != "oauth2" {
		t.Fatalf("read tool security scheme type = %q, want oauth2", readTool.Meta.SecuritySchemes[0].Type)
	}
	if got := readTool.Meta.SecuritySchemes[0].Scopes; len(got) != 1 || got[0] != "notes:read" {
		t.Fatalf("read tool scopes = %#v, want notes:read", got)
	}
	if readTool.Annotations.ReadOnlyHint == nil || !*readTool.Annotations.ReadOnlyHint {
		t.Fatalf("readOnlyHint = %#v, want true", readTool.Annotations.ReadOnlyHint)
	}
	if readTool.Annotations.DestructiveHint == nil || *readTool.Annotations.DestructiveHint {
		t.Fatalf("destructiveHint = %#v, want false", readTool.Annotations.DestructiveHint)
	}

	writeTool := body.Tools[1]
	if got := writeTool.Meta.SecuritySchemes[0].Scopes; len(got) != 1 || got[0] != "notes:write" {
		t.Fatalf("write tool scopes = %#v, want notes:write", got)
	}
	if writeTool.Annotations.ReadOnlyHint == nil || *writeTool.Annotations.ReadOnlyHint {
		t.Fatalf("write readOnlyHint = %#v, want false", writeTool.Annotations.ReadOnlyHint)
	}
	if writeTool.Annotations.DestructiveHint == nil || !*writeTool.Annotations.DestructiveHint {
		t.Fatalf("write destructiveHint = %#v, want true", writeTool.Annotations.DestructiveHint)
	}
}

func TestToolMetadataDefaultFNSScopeOmitsOAuthScopes(t *testing.T) {
	cfg := &appconfig.AppConfig{
		OAuth: config.OAuthConfig{
			Enabled:         true,
			ScopesSupported: []string{"notes:read"},
			DefaultFNSScope: "p:mcp c:* f:*",
		},
	}

	tool := withMCPToolMetadata(mcp.NewTool("note_get"), cfg, mcpToolMetadata{
		ReadOnly: true,
		Scopes:   []string{"notes:read"},
	})

	payload, err := json.Marshal(mcp.NewListToolsResult([]mcp.Tool{tool}, ""))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var body struct {
		Tools []struct {
			Meta struct {
				SecuritySchemes []struct {
					Scopes []string `json:"scopes"`
				} `json:"securitySchemes"`
			} `json:"_meta"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got := body.Tools[0].Meta.SecuritySchemes[0].Scopes; len(got) != 0 {
		t.Fatalf("scopes = %#v, want empty", got)
	}
}
