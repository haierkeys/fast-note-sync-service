package mcp_router

import (
	"context"
	"fmt"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"github.com/mark3labs/mcp-go/mcp"
	mcpsrv "github.com/mark3labs/mcp-go/server"
)

func getInt64Arg(args map[string]interface{}, key string) int64 {
	if val, ok := args[key]; ok {
		if f, ok := val.(float64); ok {
			return int64(f)
		}
		if i, ok := val.(int64); ok {
			return i
		}
		if i, ok := val.(int); ok {
			return int64(i)
		}
	}
	return 0
}

func registerVaultTools(srv *mcpsrv.MCPServer, appContainer *app.App) {
	vaultSvc := appContainer.VaultService
	cfg := appContainer.Config()

	// 1. List Vaults
	toolListVaults := mcp.NewTool("vault_list",
		mcp.WithDescription("List all available note vaults"),
		mcp.WithOutputSchema[mcpVaultListOutput](),
	)
	srv.AddTool(readOnlyMCPTool(toolListVaults, cfg, "vaults:read"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_r"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		uid := getUIDFromContext(ctx)

		vaults, err := vaultSvc.List(ctx, uid)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Filter vaults by token allowlist (if restricted).
		// 按 token 白名单过滤笔记库列表（如有限制）。
		allowedVaults, _ := ctx.Value("vaults").(string)
		if allowedVaults != "" {
			filtered := vaults[:0]
			for _, v := range vaults {
				if util.VerifyVaultAccess(allowedVaults, v.Name) {
					filtered = append(filtered, v)
				}
			}
			vaults = filtered
		}

		if len(vaults) == 0 {
			return mcp.NewToolResultStructured(mcpVaultListOutput{
				Count:  0,
				Vaults: vaults,
			}, "No vaults found."), nil
		}

		resStr := fmt.Sprintf("Found %d vaults:\n", len(vaults))
		for _, v := range vaults {
			resStr += fmt.Sprintf("- %s (ID: %d) [Notes: %d, Files: %d]\n", v.Name, v.ID, v.NoteCount, v.FileCount)
		}
		return mcp.NewToolResultStructured(mcpVaultListOutput{
			Count:  len(vaults),
			Vaults: vaults,
		}, resStr), nil
	})

	// 2. Get Vault
	toolGetVault := mcp.NewTool("vault_get",
		mcp.WithDescription("Get details of a specific vault by ID"),
		mcp.WithOutputSchema[mcpVaultOutput](),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("Vault ID")),
	)
	srv.AddTool(readOnlyMCPTool(toolGetVault, cfg, "vaults:read"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_r"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		id := getInt64Arg(args, "id")

		vault, err := vaultSvc.Get(ctx, uid, id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Check vault access by name after fetching (ID-based lookup).
		// 查到库后按库名做 vault 白名单校验（ID 方式查询）。
		if err := checkVaultAccess(ctx, vault.Name); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resStr := fmt.Sprintf("Vault: %s\nID: %d\nNotes: %d\nFiles: %d\nTotal Size: %d", vault.Name, vault.ID, vault.NoteCount, vault.FileCount, vault.Size)
		return mcp.NewToolResultStructured(mcpVaultOutput{
			Vault: vault,
		}, resStr), nil
	})

	// 3. Create or Update Vault
	toolCreateUpdateVault := mcp.NewTool("vault_create_or_update",
		mcp.WithDescription("Create a new vault or update an existing vault (by passing 'id')"),
		mcp.WithOutputSchema[mcpVaultMutationOutput](),
		mcp.WithString("vault", mcp.Required(), mcp.Description("Vault name")),
		mcp.WithNumber("id", mcp.Description("Vault ID for update. Omit or 0 to create new vault.")),
	)
	srv.AddTool(writeMCPTool(toolCreateUpdateVault, cfg, false, "notes:write"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_w"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vaultName, _ := args["vault"].(string)
		id := getInt64Arg(args, "id")

		if id > 0 {
			// For update, fetch existing vault to get its name for allowlist check.
			// 更新时先查出旧库名做白名单校验。
			existing, err := vaultSvc.Get(ctx, uid, id)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if err := checkVaultAccess(ctx, existing.Name); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			vault, err := vaultSvc.Update(ctx, uid, id, vaultName)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			fallback := fmt.Sprintf("Vault updated: %s (ID: %d)", vault.Name, vault.ID)
			return mcp.NewToolResultStructured(mcpVaultMutationOutput{
				Operation: "update",
				Vault:     vault,
				ID:        vault.ID,
			}, fallback), nil
		} else {
			// Disallow creating new vaults when the token has a vault restriction.
			// 当 token 有库限制时，禁止创建新库（防止绕过限制）。
			allowedVaults, _ := ctx.Value("vaults").(string)
			if allowedVaults != "" {
				return mcp.NewToolResultError("vault access restricted: cannot create new vault with a restricted token"), nil
			}
			vault, err := vaultSvc.Create(ctx, uid, vaultName)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			fallback := fmt.Sprintf("Vault created: %s (ID: %d)", vault.Name, vault.ID)
			return mcp.NewToolResultStructured(mcpVaultMutationOutput{
				Operation: "create",
				Vault:     vault,
				ID:        vault.ID,
			}, fallback), nil
		}
	})

	// 4. Delete Vault
	toolDeleteVault := mcp.NewTool("vault_delete",
		mcp.WithDescription("Delete a vault by ID"),
		mcp.WithOutputSchema[mcpVaultMutationOutput](),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("Vault ID")),
	)
	srv.AddTool(writeMCPTool(toolDeleteVault, cfg, true, "notes:write"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_w"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		id := getInt64Arg(args, "id")

		// Fetch vault first to get its name for allowlist check.
		// 先查出库名做白名单校验。
		existing, err := vaultSvc.Get(ctx, uid, id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if err := checkVaultAccess(ctx, existing.Name); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		err = vaultSvc.Delete(ctx, uid, id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultStructured(mcpVaultMutationOutput{
			Operation: "delete",
			ID:        id,
		}, fmt.Sprintf("Deleted vault with ID: %d", id)), nil
	})
}
