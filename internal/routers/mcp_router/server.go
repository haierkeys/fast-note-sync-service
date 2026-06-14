package mcp_router

import (
	"context"
	"fmt"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
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

func getClientInfoFromContext(ctx context.Context) (string, string, string) {
	var cType, cName, cVer string
	if val := ctx.Value("client_type"); val != nil {
		cType, _ = val.(string)
	}
	if val := ctx.Value("client_name"); val != nil {
		cName, _ = val.(string)
	}
	if val := ctx.Value("client_version"); val != nil {
		cVer, _ = val.(string)
	}
	return cType, cName, cVer
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

func checkPermission(ctx context.Context, function string) error {
	scope, _ := ctx.Value("scope").(string)
	cType, _, _ := getClientInfoFromContext(ctx)
	if !pkgapp.VerifyPermissions(scope, "mcp", cType, function) {
		return fmt.Errorf("permission denied: %s", function)
	}
	return nil
}

// checkVaultAccess verifies whether the target vault is permitted by the token's vault allowlist.
// If the token has no vault restriction (vaults == ""), all vaults are allowed.
// checkVaultAccess 验证目标笔记库是否在 token 的允许列表中。
// 如果 token 没有配置 vault 限制（vaults 为空），则允许访问所有库。
func checkVaultAccess(ctx context.Context, vault string) error {
	allowedVaults, _ := ctx.Value("vaults").(string)
	if allowedVaults == "" {
		return nil
	}
	if !util.VerifyVaultAccess(allowedVaults, vault) {
		return fmt.Errorf("vault access restricted: %s", vault)
	}
	return nil
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

	// Folder Tools
	registerFolderTools(srv, appContainer, wss)

	// Vault Tools
	registerVaultTools(srv, appContainer)

	return srv
}
