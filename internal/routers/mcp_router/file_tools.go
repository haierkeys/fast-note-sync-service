package mcp_router

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"io"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"github.com/mark3labs/mcp-go/mcp"
	mcpsrv "github.com/mark3labs/mcp-go/server"
)

func registerFileTools(srv *mcpsrv.MCPServer, appContainer *app.App, wss *pkgapp.WebsocketServer) {
	fileSvc := appContainer.FileService

	// 1. List Files
	toolListFiles := mcp.NewTool("file_list",
		mcp.WithDescription("List files in a vault"),
		mcp.WithString("vault", mcp.Required(), mcp.Description("Vault name")),
		mcp.WithString("keyword", mcp.Description("Search keyword")),
	)
	srv.AddTool(toolListFiles, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		keyword, _ := args["keyword"].(string)

		pager := &pkgapp.Pager{
			Page:     pkgapp.GetPage(1),
			PageSize: pkgapp.GetPageSize(100),
		}
		files, _, err := fileSvc.List(ctx, uid, &dto.FileListRequest{
			Vault:   vault,
			Keyword: keyword,
		}, pager)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resStr := fmt.Sprintf("Found %d files:\n", len(files))
		for _, f := range files {
			resStr += fmt.Sprintf("- %s (Size: %d)\n", f.Path, f.Size)
		}
		return mcp.NewToolResultText(resStr), nil
	})

	// 2. Get File Info
	toolGetFileInfo := mcp.NewTool("file_get_info",
		mcp.WithDescription("Get file metadata information"),
		mcp.WithString("vault", mcp.Required(), mcp.Description("Vault name")),
		mcp.WithString("path", mcp.Required(), mcp.Description("File path")),
	)
	srv.AddTool(toolGetFileInfo, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		path, _ := args["path"].(string)

		file, err := fileSvc.Get(ctx, uid, &dto.FileGetRequest{
			Vault:    vault,
			Path:     path,
			PathHash: util.EncodeHash32(path),
		})

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resStr := fmt.Sprintf("File path: %s\nSize: %d bytes\nMtime: %d", file.Path, file.Size, file.Mtime)
		return mcp.NewToolResultText(resStr), nil
	})

	// 3. Get File Content (Read File)
	toolGetContent := mcp.NewTool("file_read",
		mcp.WithDescription("Read file content and return. Returned as base64 string because it might be binary."),
		mcp.WithString("vault", mcp.Required(), mcp.Description("Vault name")),
		mcp.WithString("path", mcp.Required(), mcp.Description("File path")),
	)
	srv.AddTool(toolGetContent, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		path, _ := args["path"].(string)

		reader, _, _, _, err := fileSvc.GetContent(ctx, uid, &dto.FileGetRequest{
			Vault:    vault,
			Path:     path,
			PathHash: util.EncodeHash32(path),
		})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		defer reader.Close()

		data, err := io.ReadAll(reader)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		b64 := base64.StdEncoding.EncodeToString(data)
		return mcp.NewToolResultText(fmt.Sprintf("Base64 Content follows:\n%s", b64)), nil
	})

	// 4. Delete File
	toolDeleteFile := mcp.NewTool("file_delete",
		mcp.WithDescription("Delete a file"),
		mcp.WithString("vault", mcp.Required(), mcp.Description("Vault name")),
		mcp.WithString("path", mcp.Required(), mcp.Description("File path")),
	)
	srv.AddTool(toolDeleteFile, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		path, _ := args["path"].(string)

		file, err := fileSvc.Delete(ctx, uid, &dto.FileDeleteRequest{
			Vault:    vault,
			Path:     path,
			PathHash: util.EncodeHash32(path),
		})

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		wss.BroadcastToUser(uid, code.Success.WithData(file).WithVault(vault), "FileSyncDelete")
		return mcp.NewToolResultText(fmt.Sprintf("Deleted file: %s", file.Path)), nil
	})

	// 5. Rename File
	toolRenameFile := mcp.NewTool("file_rename",
		mcp.WithDescription("Rename a file"),
		mcp.WithString("vault", mcp.Required(), mcp.Description("Vault name")),
		mcp.WithString("oldPath", mcp.Required(), mcp.Description("Old file path")),
		mcp.WithString("newPath", mcp.Required(), mcp.Description("New file path")),
	)
	srv.AddTool(toolRenameFile, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		oldPath, _ := args["oldPath"].(string)
		newPath, _ := args["newPath"].(string)

		oldFile, newFile, err := fileSvc.Rename(ctx, uid, &dto.FileRenameRequest{
			Vault:       vault,
			OldPath:     oldPath,
			OldPathHash: util.EncodeHash32(oldPath),
			Path:        newPath,
			PathHash:    util.EncodeHash32(newPath),
		})

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		wss.BroadcastToUser(uid, code.Success.WithData(dto.FileSyncRenameMessage{
			Path:             newFile.Path,
			PathHash:         newFile.PathHash,
			ContentHash:      newFile.ContentHash,
			Ctime:            newFile.Ctime,
			Mtime:            newFile.Mtime,
			Size:             newFile.Size,
			UpdatedTimestamp: newFile.UpdatedTimestamp,
			OldPath:          oldFile.Path,
			OldPathHash:      oldFile.PathHash,
		}).WithVault(vault), "FileSyncRename")
		return mcp.NewToolResultText(fmt.Sprintf("Renamed file from %s to %s", oldFile.Path, newFile.Path)), nil
	})

	// 1. Restore File
	toolRestoreFile := mcp.NewTool("file_restore",
		mcp.WithDescription("Restore a deleted file from recycle bin"),
		mcp.WithString("vault", mcp.Required(), mcp.Description("Vault name")),
		mcp.WithString("path", mcp.Required(), mcp.Description("File path")),
	)
	srv.AddTool(toolRestoreFile, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		path, _ := args["path"].(string)

		file, err := fileSvc.Restore(ctx, uid, &dto.FileRestoreRequest{
			Vault:    vault,
			Path:     path,
			PathHash: util.EncodeHash32(path),
		})

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		wss.BroadcastToUser(uid, code.Success.WithData(file).WithVault(vault), "FileSyncUpdate")
		return mcp.NewToolResultText(fmt.Sprintf("Restored file: %s", file.Path)), nil
	})

	// 2. Recycle Clear File
	toolRecycleClearFile := mcp.NewTool("file_recycle_clear",
		mcp.WithDescription("Permanently delete a file from recycle bin (or all if path is empty)"),
		mcp.WithString("vault", mcp.Required(), mcp.Description("Vault name")),
		mcp.WithString("path", mcp.Description("File path. If empty, potentially clear all")),
	)
	srv.AddTool(toolRecycleClearFile, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		path, _ := args["path"].(string)

		err := fileSvc.RecycleClear(ctx, uid, &dto.FileRecycleClearRequest{
			Vault:    vault,
			Path:     path,
			PathHash: util.EncodeHash32(path),
		})

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText("Recycle clear successful"), nil
	})
}
