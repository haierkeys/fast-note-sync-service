package mcp_router

import (
	"context"
	"fmt"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
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

	// 1. List Vaults
	toolListVaults := mcp.NewTool("vault_list",
		mcp.WithDescription("List all available note vaults"),
	)
	srv.AddTool(toolListVaults, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_r"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		uid := getUIDFromContext(ctx)

		vaults, err := vaultSvc.List(ctx, uid)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if len(vaults) == 0 {
			return mcp.NewToolResultText("No vaults found."), nil
		}

		resStr := fmt.Sprintf("Found %d vaults:\n", len(vaults))
		for _, v := range vaults {
			resStr += fmt.Sprintf("- %s (ID: %d) [Notes: %d, Files: %d]\n", v.Name, v.ID, v.NoteCount, v.FileCount)
		}
		return mcp.NewToolResultText(resStr), nil
	})

	// 2. Get Vault
	toolGetVault := mcp.NewTool("vault_get",
		mcp.WithDescription("Get details of a specific vault by ID"),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("Vault ID")),
	)
	srv.AddTool(toolGetVault, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		resStr := fmt.Sprintf("Vault: %s\nID: %d\nNotes: %d\nFiles: %d\nTotal Size: %d", vault.Name, vault.ID, vault.NoteCount, vault.FileCount, vault.Size)
		return mcp.NewToolResultText(resStr), nil
	})

	// 3. Create or Update Vault
	toolCreateUpdateVault := mcp.NewTool("vault_create_or_update",
		mcp.WithDescription("Create a new vault or update an existing vault (by passing 'id')"),
		mcp.WithString("vault", mcp.Required(), mcp.Description("Vault name")),
		mcp.WithNumber("id", mcp.Description("Vault ID for update. Omit or 0 to create new vault.")),
	)
	srv.AddTool(toolCreateUpdateVault, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_w"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vaultName, _ := args["vault"].(string)
		id := getInt64Arg(args, "id")

		if id > 0 {
			vault, err := vaultSvc.Update(ctx, uid, id, vaultName)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Vault updated: %s (ID: %d)", vault.Name, vault.ID)), nil
		} else {
			vault, err := vaultSvc.Create(ctx, uid, vaultName)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Vault created: %s (ID: %d)", vault.Name, vault.ID)), nil
		}
	})

	// 4. Delete Vault
	toolDeleteVault := mcp.NewTool("vault_delete",
		mcp.WithDescription("Delete a vault by ID"),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("Vault ID")),
	)
	srv.AddTool(toolDeleteVault, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_w"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		id := getInt64Arg(args, "id")

		err := vaultSvc.Delete(ctx, uid, id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Deleted vault with ID: %d", id)), nil
	})
}
