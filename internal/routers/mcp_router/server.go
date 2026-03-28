package mcp_router

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/mark3labs/mcp-go/mcp"
	mcpsrv "github.com/mark3labs/mcp-go/server"
)

func getUIDFromContext(ctx context.Context) int64 {
	if val := ctx.Value("uid"); val != nil {
		if uid, ok := val.(int64); ok {
			return uid
		}
	}
	return 1
}

func getDefaultVaultName(ctx context.Context, appContainer *app.App) string {
	// 1. From context (Header X-Default-Vault-Name)
	if val := ctx.Value("default_vault_name"); val != nil {
		if name, ok := val.(string); ok && name != "" {
			return name
		}
	}

	uid := getUIDFromContext(ctx)

	// 2. From user settings (placeholder, assuming there might be a default vault setting)
	// We can try to list vaults and pick the first one as a fallback for now
	vaults, err := appContainer.VaultService.List(ctx, uid)
	if err == nil && len(vaults) > 0 {
		return vaults[0].Name
	}

	return "Default"
}

func getArgs(req mcp.CallToolRequest) map[string]interface{} {
	if req.Params.Arguments != nil {
		if args, ok := req.Params.Arguments.(map[string]interface{}); ok {
			return args
		}
	}
	return make(map[string]interface{})
}

func NewMCPServer(appContainer *app.App, wss *pkgapp.WebsocketServer) *mcpsrv.MCPServer {
	// Create MCP server
	srv := mcpsrv.NewMCPServer(
		"fast-note-sync-service",
		appContainer.Version().Version,
	)

	// Note Tools
	registerNoteTools(srv, appContainer, wss)

	// File Tools
	registerFileTools(srv, appContainer, wss)

	// Vault Tools
	registerVaultTools(srv, appContainer)

	return srv
}
