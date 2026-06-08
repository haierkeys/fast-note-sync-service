package mcp_router

import (
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/mark3labs/mcp-go/mcp"
)

type mcpToolMetadata struct {
	ReadOnly    bool
	Destructive bool
	OpenWorld   bool
	Idempotent  bool
	Scopes      []string
}

func withMCPToolMetadata(tool mcp.Tool, cfg *app.AppConfig, metadata mcpToolMetadata) mcp.Tool {
	tool.Annotations.ReadOnlyHint = boolPtr(metadata.ReadOnly)
	tool.Annotations.DestructiveHint = boolPtr(metadata.Destructive)
	tool.Annotations.OpenWorldHint = boolPtr(metadata.OpenWorld)
	tool.Annotations.IdempotentHint = boolPtr(metadata.Idempotent)
	if tool.OutputSchema.Type == "" && len(tool.RawOutputSchema) == 0 {
		tool.OutputSchema = mcp.ToolOutputSchema{
			Type:       "object",
			Properties: map[string]any{},
			Required:   []string{},
		}
	}

	if tool.Meta == nil {
		tool.Meta = &mcp.Meta{}
	}
	if tool.Meta.AdditionalFields == nil {
		tool.Meta.AdditionalFields = make(map[string]any)
	}
	tool.Meta.AdditionalFields["securitySchemes"] = []map[string]any{
		{
			"type":   "oauth2",
			"scopes": oauthScopesForTool(cfg, metadata.Scopes),
		},
	}

	return tool
}

func readOnlyMCPTool(tool mcp.Tool, cfg *app.AppConfig, scopes ...string) mcp.Tool {
	return withMCPToolMetadata(tool, cfg, mcpToolMetadata{
		ReadOnly:    true,
		Destructive: false,
		OpenWorld:   false,
		Idempotent:  true,
		Scopes:      scopes,
	})
}

func writeMCPTool(tool mcp.Tool, cfg *app.AppConfig, destructive bool, scopes ...string) mcp.Tool {
	return withMCPToolMetadata(tool, cfg, mcpToolMetadata{
		ReadOnly:    false,
		Destructive: destructive,
		OpenWorld:   false,
		Idempotent:  false,
		Scopes:      scopes,
	})
}

func oauthScopesForTool(cfg *app.AppConfig, requested []string) []string {
	if cfg != nil && cfg.OAuth.DefaultFNSScope != "" {
		return []string{}
	}
	if len(requested) == 0 {
		if cfg != nil && len(cfg.OAuth.RequiredScopes) > 0 {
			return append([]string(nil), cfg.OAuth.RequiredScopes...)
		}
		if cfg != nil && len(cfg.OAuth.ScopesSupported) > 0 {
			return append([]string(nil), cfg.OAuth.ScopesSupported...)
		}
		return []string{}
	}
	if cfg == nil || len(cfg.OAuth.ScopesSupported) == 0 {
		return append([]string(nil), requested...)
	}

	supported := make(map[string]struct{}, len(cfg.OAuth.ScopesSupported))
	for _, scope := range cfg.OAuth.ScopesSupported {
		supported[scope] = struct{}{}
	}

	scopes := make([]string, 0, len(requested))
	for _, scope := range requested {
		if _, ok := supported[scope]; ok {
			scopes = append(scopes, scope)
		}
	}
	return scopes
}

func boolPtr(value bool) *bool {
	return &value
}
