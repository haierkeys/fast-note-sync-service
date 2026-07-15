package mcp_router

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"github.com/mark3labs/mcp-go/mcp"
	mcpsrv "github.com/mark3labs/mcp-go/server"
)

func registerFileTools(srv *mcpsrv.MCPServer, appContainer *app.App, wss *pkgapp.WebsocketServer) {
	fileSvc := appContainer.FileService
	cfg := appContainer.Config()

	// 1. List Files
	toolListFiles := mcp.NewTool("file_list",
		mcp.WithDescription("List files in a vault"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("keyword", mcp.Description("Search keyword")),
		mcp.WithOutputSchema[mcpFileListOutput](),
	)
	srv.AddTool(readOnlyMCPTool(toolListFiles, cfg, "files:read"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "file_r"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		if vault == "" || strings.EqualFold(vault, "default") {
			vault = getDefaultVaultName(ctx, appContainer)
		}
		if err := checkVaultAccess(ctx, vault); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		keyword, _ := args["keyword"].(string)

		pager := &pkgapp.Pager{
			Page:     pkgapp.GetPage(1),
			PageSize: pkgapp.GetPageSize(100),
		}
		files, _, err := fileSvc.WithClient(getClientInfoFromContext(ctx)).List(ctx, uid, &dto.FileListRequest{
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
		mcpFiles := make([]*dto.McpFileDTO, len(files))
		for i, f := range files {
			mcpFiles[i] = f.ToMcpFileDTO()
		}
		return mcp.NewToolResultStructured(mcpFileListOutput{
			Vault: vault,
			Count: len(files),
			Files: mcpFiles,
		}, resStr), nil
	})

	// 2. Get File Info
	toolGetFileInfo := mcp.NewTool("file_get_info",
		mcp.WithDescription("Get file metadata information"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("File path")),
		mcp.WithOutputSchema[mcpFileOutput](),
	)
	srv.AddTool(readOnlyMCPTool(toolGetFileInfo, cfg, "files:read"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "file_r"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		if vault == "" || strings.EqualFold(vault, "default") {
			vault = getDefaultVaultName(ctx, appContainer)
		}
		if err := checkVaultAccess(ctx, vault); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		path, _ := args["path"].(string)

		file, err := fileSvc.WithClient(getClientInfoFromContext(ctx)).Get(ctx, uid, &dto.FileGetRequest{
			Vault:    vault,
			Path:     path,
			PathHash: util.EncodeHash32(path),
		})

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resStr := fmt.Sprintf("File path: %s\nSize: %d bytes\nMtime: %d", file.Path, file.Size, file.Mtime)
		return mcp.NewToolResultStructured(mcpFileOutput{
			Vault: vault,
			File:  file.ToMcpFileDTO(),
		}, resStr), nil
	})

	// 3. Get File Content (Read File)
	toolGetContent := mcp.NewTool("file_read",
		mcp.WithDescription("Read file content and return. Returned as base64 string because it might be binary."),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("File path")),
		mcp.WithOutputSchema[mcpFileReadOutput](),
	)
	srv.AddTool(readOnlyMCPTool(toolGetContent, cfg, "files:read"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "file_r"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		if vault == "" || strings.EqualFold(vault, "default") {
			vault = getDefaultVaultName(ctx, appContainer)
		}
		if err := checkVaultAccess(ctx, vault); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		path, _ := args["path"].(string)

		reader, _, _, _, err := fileSvc.WithClient(getClientInfoFromContext(ctx)).GetContent(ctx, uid, &dto.FileGetRequest{
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
		return mcp.NewToolResultStructured(mcpFileReadOutput{
			Vault:         vault,
			Path:          path,
			ContentBase64: b64,
			Size:          len(data),
		}, fmt.Sprintf("Base64 Content follows:\n%s", b64)), nil
	})

	// 4. Delete File
	toolDeleteFile := mcp.NewTool("file_delete",
		mcp.WithDescription("Delete a file"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("File path")),
		mcp.WithOutputSchema[mcpFileMutationOutput](),
	)
	srv.AddTool(writeMCPTool(toolDeleteFile, cfg, true, "files:write"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "file_w"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		if vault == "" || strings.EqualFold(vault, "default") {
			vault = getDefaultVaultName(ctx, appContainer)
		}
		if err := checkVaultAccess(ctx, vault); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		path, _ := args["path"].(string)

		file, err := fileSvc.WithClient(getClientInfoFromContext(ctx)).Delete(ctx, uid, &dto.FileDeleteRequest{
			Vault:    vault,
			Path:     path,
			PathHash: util.EncodeHash32(path),
		})

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		wss.BroadcastToUser(uid, code.Success.WithData(file).WithVault(vault), "FileSyncDelete")
		fallback := fmt.Sprintf("Deleted file: %s", file.Path)
		return mcp.NewToolResultStructured(mcpFileMutationOutput{
			Vault:     vault,
			Operation: "delete",
			File:      file.ToMcpFileDTO(),
		}, fallback), nil
	})

	// 5. Rename File
	toolRenameFile := mcp.NewTool("file_rename",
		mcp.WithDescription("Rename a file"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("oldPath", mcp.Required(), mcp.Description("Old file path")),
		mcp.WithString("newPath", mcp.Required(), mcp.Description("New file path")),
		mcp.WithOutputSchema[mcpFileMutationOutput](),
	)
	srv.AddTool(writeMCPTool(toolRenameFile, cfg, true, "files:write"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "file_w"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		if vault == "" || strings.EqualFold(vault, "default") {
			vault = getDefaultVaultName(ctx, appContainer)
		}
		if err := checkVaultAccess(ctx, vault); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		oldPath, _ := args["oldPath"].(string)
		newPath, _ := args["newPath"].(string)

		oldFile, newFile, err := fileSvc.WithClient(getClientInfoFromContext(ctx)).Rename(ctx, uid, &dto.FileRenameRequest{
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
		fallback := fmt.Sprintf("Renamed file from %s to %s", oldFile.Path, newFile.Path)
		return mcp.NewToolResultStructured(mcpFileMutationOutput{
			Vault:     vault,
			Operation: "rename",
			OldFile:   oldFile.ToMcpFileDTO(),
			NewFile:   newFile.ToMcpFileDTO(),
		}, fallback), nil
	})

	// 1. Restore File
	toolRestoreFile := mcp.NewTool("file_restore",
		mcp.WithDescription("Restore a deleted file from recycle bin"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("File path")),
		mcp.WithOutputSchema[mcpFileMutationOutput](),
	)
	srv.AddTool(writeMCPTool(toolRestoreFile, cfg, true, "files:write"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "file_w"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		if vault == "" || strings.EqualFold(vault, "default") {
			vault = getDefaultVaultName(ctx, appContainer)
		}
		if err := checkVaultAccess(ctx, vault); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		path, _ := args["path"].(string)

		file, err := fileSvc.WithClient(getClientInfoFromContext(ctx)).Restore(ctx, uid, &dto.FileRestoreRequest{
			Vault:    vault,
			Path:     path,
			PathHash: util.EncodeHash32(path),
		})

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		wss.BroadcastToUser(uid, code.Success.WithData(file).WithVault(vault), "FileSyncUpdate")
		fallback := fmt.Sprintf("Restored file: %s", file.Path)
		return mcp.NewToolResultStructured(mcpFileMutationOutput{
			Vault:     vault,
			Operation: "restore",
			File:      file.ToMcpFileDTO(),
		}, fallback), nil
	})

	// 2. Recycle Clear File
	toolRecycleClearFile := mcp.NewTool("file_recycle_clear",
		mcp.WithDescription("Permanently delete a file from recycle bin (or all if path is empty)"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Description("File path. If empty, potentially clear all")),
		mcp.WithOutputSchema[mcpFileRecycleClearOutput](),
	)
	srv.AddTool(writeMCPTool(toolRecycleClearFile, cfg, true, "files:write"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "file_w"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		if vault == "" || strings.EqualFold(vault, "default") {
			vault = getDefaultVaultName(ctx, appContainer)
		}
		if err := checkVaultAccess(ctx, vault); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		path, _ := args["path"].(string)

		err := fileSvc.WithClient(getClientInfoFromContext(ctx)).RecycleClear(ctx, uid, &dto.FileRecycleClearRequest{
			Vault:    vault,
			Path:     path,
			PathHash: util.EncodeHash32(path),
		})

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultStructured(mcpFileRecycleClearOutput{
			Vault: vault,
			Path:  path,
		}, "Recycle clear successful"), nil
	})

	// 8. Write File (Create or update file/attachment via MCP)
	// 8. 写入文件（通过 MCP 新建或更新文件/附件）
	toolWriteFile := mcp.NewTool("file_write",
		mcp.WithDescription("Create or update a file (attachment) in the vault by uploading its base64 encoded content"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Target file path in the vault (e.g. 'images/my_pic.png')")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Base64 encoded file content")),
		mcp.WithOutputSchema[mcpFileWriteOutput](),
	)
	srv.AddTool(writeMCPTool(toolWriteFile, cfg, false, "files:write"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "file_w"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		if vault == "" || strings.EqualFold(vault, "default") {
			vault = getDefaultVaultName(ctx, appContainer)
		}
		if err := checkVaultAccess(ctx, vault); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		path, _ := args["path"].(string)
		b64Content, _ := args["content"].(string)

		if !util.ValidatePath(path) {
			return mcp.NewToolResultError("invalid file path"), nil
		}

		// Decode base64 content // 解码 Base64 内容
		data, err := base64.StdEncoding.DecodeString(b64Content)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to decode base64 content: %v", err)), nil
		}

		// Get temp directory path // 获取临时目录路径
		tempDir := appContainer.Config().App.TempPath
		if tempDir == "" {
			tempDir = "storage/temp"
		}
		_ = os.MkdirAll(tempDir, 0755)
		tempPath := filepath.Join(tempDir, uuid.New().String())

		// Write data to temp file // 将数据写入临时文件
		if err := os.WriteFile(tempPath, data, 0644); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to write temp file: %v", err)), nil
		}
		defer os.Remove(tempPath)

		now := time.Now().UnixMilli()
		params := &dto.FileUpdateRequest{
			Vault:       vault,
			Path:        path,
			PathHash:    util.EncodeHash32(path),
			ContentHash: util.EncodeHash32Bytes(data),
			SavePath:    tempPath,
			Size:        int64(len(data)),
			Ctime:       now,
			Mtime:       now,
		}

		cType, cName, cVer := getClientInfoFromContext(ctx)
		_, fileDTO, err := fileSvc.WithClient(cType, cName, cVer).UpdateOrCreate(ctx, uid, params, false)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Broadcast WebSocket event to sync other clients // 广播 WebSocket 事件以同步其他客户端
		wss.BroadcastToUser(uid, code.Success.WithData(dto.FileSyncModifyMessage{
			Path:             fileDTO.Path,
			PathHash:         fileDTO.PathHash,
			ContentHash:      fileDTO.ContentHash,
			Size:             fileDTO.Size,
			Ctime:            fileDTO.Ctime,
			Mtime:            fileDTO.Mtime,
			UpdatedTimestamp: fileDTO.UpdatedTimestamp,
		}).WithVault(vault), "FileSyncUpdate")

		return mcp.NewToolResultStructured(mcpFileWriteOutput{
			Vault: vault,
			File:  fileDTO.ToMcpFileDTO(),
		}, fmt.Sprintf("Successfully wrote file: %s (Size: %d bytes)", fileDTO.Path, fileDTO.Size)), nil
	})
}
