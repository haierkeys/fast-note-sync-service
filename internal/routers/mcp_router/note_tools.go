package mcp_router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"github.com/mark3labs/mcp-go/mcp"
	mcpsrv "github.com/mark3labs/mcp-go/server"
)

// noteMCPError renders a service error for MCP clients. Application codes (notably the
// conditional-write conflicts 452/453) are returned as a machine-readable JSON payload so an
// agent can read the current revision/content and retry, instead of parsing a prose message.
// noteMCPError 把服务层错误返回给 MCP 客户端；条件写入冲突（452/453）以 JSON 形式返回，
// 便于 agent 读取当前版本号/内容后重试。
func noteMCPError(err error) *mcp.CallToolResult {
	var appErr *code.Code
	if errors.As(err, &appErr) {
		payload := map[string]interface{}{
			"code":    appErr.Code(),
			"status":  false,
			"message": appErr.Msg(),
		}
		if appErr.HaveData() {
			payload["data"] = appErr.Data()
		}
		if appErr.HaveDetails() {
			payload["details"] = strings.Join(appErr.Details(), ",")
		}
		if b, mErr := json.Marshal(payload); mErr == nil {
			return mcp.NewToolResultError(string(b))
		}
	}
	return mcp.NewToolResultError(err.Error())
}

// mcpRevisionArg extracts the optional `revision` argument. A missing or non-numeric value
// means "no revision supplied" and therefore a legacy unconditional write.
// mcpRevisionArg 读取可选的 revision 参数；缺省或非数字表示不提供版本号（旧式无条件写入）。
func mcpRevisionArg(args map[string]interface{}) *int64 {
	v, ok := args["revision"]
	if !ok || v == nil {
		return nil
	}
	switch n := v.(type) {
	case float64:
		rev := int64(n)
		return &rev
	case int64:
		rev := n
		return &rev
	case int:
		rev := int64(n)
		return &rev
	case json.Number:
		if rev, err := n.Int64(); err == nil {
			return &rev
		}
	}
	return nil
}

func registerNoteTools(srv *mcpsrv.MCPServer, appContainer *app.App, wss *pkgapp.WebsocketServer) {
	noteSvc := appContainer.NoteService
	cfg := appContainer.Config()

	// 1. List Notes
	toolListNotes := mcp.NewTool("note_list",
		mcp.WithDescription("List or search notes in a vault. Use this to find a note by title or keyword before calling note_get."),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("keyword", mcp.Description("Search keyword")),
		mcp.WithString("searchMode", mcp.Description("Where to match the keyword: 'path' (default) searches note paths and filenames; 'content' searches inside note bodies using full-text search.")),
		mcp.WithOutputSchema[mcpNoteListOutput](),
	)
	srv.AddTool(readOnlyMCPTool(toolListNotes, cfg, "notes:read"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_r"); err != nil {
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
		searchMode, _ := args["searchMode"].(string)

		pager := &pkgapp.Pager{
			Page:     pkgapp.GetPage(1),
			PageSize: pkgapp.GetPageSize(100),
		}
		notes, _, err := noteSvc.WithClient(getClientInfoFromContext(ctx)).List(ctx, uid, &dto.NoteListRequest{
			Vault:      vault,
			Keyword:    keyword,
			SearchMode: searchMode,
		}, pager)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resStr := fmt.Sprintf("Found %d notes:\n", len(notes))
		for _, n := range notes {
			resStr += fmt.Sprintf("- %s (ID: %d, Size: %d, Mtime: %d)\n", n.Path, n.ID, n.Size, n.Mtime)
		}
		mcpNotes := make([]*dto.McpNoteNoContentDTO, len(notes))
		for i, n := range notes {
			mcpNotes[i] = n.ToMcpNoteNoContentDTO()
		}
		return mcp.NewToolResultStructured(mcpNoteListOutput{
			Vault: vault,
			Count: len(notes),
			Notes: mcpNotes,
		}, resStr), nil
	})

	// 2. Get Note
	toolGetNote := mcp.NewTool("note_get",
		mcp.WithDescription("Get the full content of a single note. Requires the EXACT vault-relative file path."),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Exact vault-relative path to the note.")),
		mcp.WithOutputSchema[mcpNoteOutput](),
	)
	srv.AddTool(readOnlyMCPTool(toolGetNote, cfg, "notes:read"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_r"); err != nil {
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
		pathHash := util.EncodeHash32(path)

		note, err := noteSvc.WithClient(getClientInfoFromContext(ctx)).Get(ctx, uid, &dto.NoteGetRequest{
			Vault:    vault,
			Path:     path,
			PathHash: pathHash,
		})

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultStructured(mcpNoteOutput{
			Vault: vault,
			Note:  note.ToMcpNoteDTO(),
		}, note.Content), nil
	})

	// 3. Create or Update Note
	toolCreateUpdateNote := mcp.NewTool("note_create_or_update",
		mcp.WithDescription("Create or update a note"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Note path")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Note content")),
		mcp.WithNumber("revision", mcp.Description("Optional concurrency guard: the `revision` you read from note_get/note_list. The write is applied only if the note is still at that revision; otherwise it fails with a conflict (code 452) and nothing is changed. Omit for a legacy unconditional write.")),
		mcp.WithOutputSchema[mcpNoteMutationOutput](),
	)
	srv.AddTool(writeMCPTool(toolCreateUpdateNote, cfg, false, "notes:write"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_w"); err != nil {
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
		content, _ := args["content"].(string)
		pathHash := util.EncodeHash32(path)
		contentHash := util.EncodeHash32(content)

		now := time.Now().UnixMilli()
		_, note, err := noteSvc.WithClient(getClientInfoFromContext(ctx)).ModifyOrCreate(ctx, uid, &dto.NoteModifyOrCreateRequest{
			Vault:       vault,
			Path:        path,
			PathHash:    pathHash,
			Content:     content,
			ContentHash: contentHash,
			Mtime:       now,
			Ctime:       now,
			Revision:    mcpRevisionArg(args),
		}, false)

		if err != nil {
			return noteMCPError(err), nil
		}

		wss.BroadcastToUser(uid, code.Success.WithData(note).WithVault(vault), "NoteSyncModify")
		fallback := fmt.Sprintf("Successfully saved note: %s (Version: %d)", note.Path, note.Version)
		return mcp.NewToolResultStructured(mcpNoteMutationOutput{
			Vault:     vault,
			Operation: "create_or_update",
			Note:      note.ToMcpNoteDTO(),
		}, fallback), nil
	})

	// 4. Delete Note
	toolDeleteNote := mcp.NewTool("note_delete",
		mcp.WithDescription("Delete a note"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Note path")),
		mcp.WithOutputSchema[mcpNoteMutationOutput](),
		mcp.WithNumber("revision", mcp.Description("Optional concurrency guard: the `revision` you read from note_get/note_list. The write is applied only if the note is still at that revision; otherwise it fails with a conflict (code 452) and nothing is changed. Omit for a legacy unconditional write.")),
	)
	srv.AddTool(writeMCPTool(toolDeleteNote, cfg, true, "notes:write"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_w"); err != nil {
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
		pathHash := util.EncodeHash32(path)

		note, err := noteSvc.WithClient(getClientInfoFromContext(ctx)).Delete(ctx, uid, &dto.NoteDeleteRequest{
			Vault:    vault,
			Path:     path,
			PathHash: pathHash,
			Revision: mcpRevisionArg(args),
		})

		if err != nil {
			return noteMCPError(err), nil
		}

		wss.BroadcastToUser(uid, code.Success.WithData(note).WithVault(vault), "NoteSyncDelete")
		fallback := fmt.Sprintf("Deleted note: %s", note.Path)
		return mcp.NewToolResultStructured(mcpNoteMutationOutput{
			Vault:     vault,
			Operation: "delete",
			Note:      note.ToMcpNoteDTO(),
		}, fallback), nil
	})

	// 5. Rename Note
	toolRenameNote := mcp.NewTool("note_rename",
		mcp.WithDescription("Rename a note"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("oldPath", mcp.Required(), mcp.Description("Old note path")),
		mcp.WithString("newPath", mcp.Required(), mcp.Description("New note path")),
		mcp.WithNumber("revision", mcp.Description("Optional concurrency guard: the `revision` you read from note_get/note_list. The write is applied only if the note is still at that revision; otherwise it fails with a conflict (code 452) and nothing is changed. Omit for a legacy unconditional write.")),
		mcp.WithOutputSchema[mcpNoteMutationOutput](),
	)
	srv.AddTool(writeMCPTool(toolRenameNote, cfg, true, "notes:write"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_w"); err != nil {
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

		oldNote, newNote, err := noteSvc.WithClient(getClientInfoFromContext(ctx)).Rename(ctx, uid, &dto.NoteRenameRequest{
			Vault:       vault,
			OldPath:     oldPath,
			OldPathHash: util.EncodeHash32(oldPath),
			Path:        newPath,
			PathHash:    util.EncodeHash32(newPath),
			Revision:    mcpRevisionArg(args),
		})

		if err != nil {
			return noteMCPError(err), nil
		}

		wss.BroadcastToUser(uid, code.Success.WithData(dto.NoteSyncRenameMessage{
			Path:             newNote.Path,
			PathHash:         newNote.PathHash,
			ContentHash:      newNote.ContentHash,
			Ctime:            newNote.Ctime,
			Mtime:            newNote.Mtime,
			Size:             newNote.Size,
			OldPath:          oldNote.Path,
			OldPathHash:      oldNote.PathHash,
			UpdatedTimestamp: newNote.UpdatedTimestamp,
		}).WithVault(vault), "NoteSyncRename")
		fallback := fmt.Sprintf("Renamed note from %s to %s", oldNote.Path, newNote.Path)
		return mcp.NewToolResultStructured(mcpNoteMutationOutput{
			Vault:     vault,
			Operation: "rename",
			OldNote:   oldNote.ToMcpNoteDTO(),
			NewNote:   newNote.ToMcpNoteDTO(),
		}, fallback), nil
	})

	// 1. Restore Note
	toolRestoreNote := mcp.NewTool("note_restore",
		mcp.WithDescription("Restore a deleted note from recycle bin"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Note path")),
		mcp.WithOutputSchema[mcpNoteMutationOutput](),
		mcp.WithNumber("revision", mcp.Description("Optional concurrency guard: the `revision` you read from note_get/note_list. The write is applied only if the note is still at that revision; otherwise it fails with a conflict (code 452) and nothing is changed. Omit for a legacy unconditional write.")),
	)
	srv.AddTool(writeMCPTool(toolRestoreNote, cfg, true, "notes:write"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_w"); err != nil {
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

		note, err := noteSvc.WithClient(getClientInfoFromContext(ctx)).Restore(ctx, uid, &dto.NoteRestoreRequest{
			Vault:    vault,
			Path:     path,
			PathHash: util.EncodeHash32(path),
			Revision: mcpRevisionArg(args),
		})

		if err != nil {
			return noteMCPError(err), nil
		}

		wss.BroadcastToUser(uid, code.Success.WithData(note).WithVault(vault), "NoteSyncModify")
		fallback := fmt.Sprintf("Restored note: %s", note.Path)
		return mcp.NewToolResultStructured(mcpNoteMutationOutput{
			Vault:     vault,
			Operation: "restore",
			Note:      note.ToMcpNoteDTO(),
		}, fallback), nil
	})

	// 2. Recycle Clear Note
	toolRecycleClear := mcp.NewTool("note_recycle_clear",
		mcp.WithDescription("Permanently delete a note from recycle bin (or all if path is empty)"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Description("Note path. If empty, potentially clear all (based on service logic)")),
		mcp.WithOutputSchema[mcpNoteRecycleClearOutput](),
	)
	srv.AddTool(writeMCPTool(toolRecycleClear, cfg, true, "notes:write"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_w"); err != nil {
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

		err := noteSvc.WithClient(getClientInfoFromContext(ctx)).RecycleClear(ctx, uid, &dto.NoteRecycleClearRequest{
			Vault:    vault,
			Path:     path,
			PathHash: util.EncodeHash32(path),
		})

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultStructured(mcpNoteRecycleClearOutput{
			Vault: vault,
			Path:  path,
		}, "Recycle clear successful"), nil
	})

	// 3. Patch Frontmatter
	toolPatchFrontmatter := mcp.NewTool("note_patch_frontmatter",
		mcp.WithDescription("Patch (update or remove) frontmatter of a note"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Note path")),
		mcp.WithString("updates", mcp.Description("JSON string for fields to update (e.g. {\"tags\":[\"t1\"]})")),
		mcp.WithString("remove", mcp.Description("JSON string array for fields to remove (e.g. [\"old_tag\"])")),
		mcp.WithNumber("revision", mcp.Description("Optional concurrency guard: the `revision` you read from note_get/note_list. The write is applied only if the note is still at that revision; otherwise it fails with a conflict (code 452) and nothing is changed. Omit for a legacy unconditional write.")),
		mcp.WithOutputSchema[mcpNoteMutationOutput](),
	)
	srv.AddTool(writeMCPTool(toolPatchFrontmatter, cfg, false, "notes:write"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_w"); err != nil {
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
		updatesStr, _ := args["updates"].(string)
		removeStr, _ := args["remove"].(string)

		var updates map[string]interface{}
		if updatesStr != "" {
			if err := json.Unmarshal([]byte(updatesStr), &updates); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid JSON for updates: %v", err)), nil
			}
		}

		var remove []string
		if removeStr != "" {
			if err := json.Unmarshal([]byte(removeStr), &remove); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid JSON for remove: %v", err)), nil
			}
		}

		note, err := noteSvc.WithClient(getClientInfoFromContext(ctx)).PatchFrontmatter(ctx, uid, &dto.NotePatchFrontmatterRequest{
			Vault:    vault,
			Path:     path,
			PathHash: util.EncodeHash32(path),
			Updates:  updates,
			Remove:   remove,
			Revision: mcpRevisionArg(args),
		})

		if err != nil {
			return noteMCPError(err), nil
		}

		wss.BroadcastToUser(uid, code.Success.WithData(note).WithVault(vault), "NoteSyncModify")
		fallback := fmt.Sprintf("Frontmatter patched for %s", note.Path)
		return mcp.NewToolResultStructured(mcpNoteMutationOutput{
			Vault:     vault,
			Operation: "patch_frontmatter",
			Note:      note.ToMcpNoteDTO(),
		}, fallback), nil
	})

	// 4. Append
	toolAppend := mcp.NewTool("note_append",
		mcp.WithDescription("Append content to the end of a note"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Note path")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Content to append")),
		mcp.WithNumber("revision", mcp.Description("Optional concurrency guard: the `revision` you read from note_get/note_list. The write is applied only if the note is still at that revision; otherwise it fails with a conflict (code 452) and nothing is changed. Omit for a legacy unconditional write.")),
		mcp.WithOutputSchema[mcpNoteMutationOutput](),
	)
	srv.AddTool(writeMCPTool(toolAppend, cfg, false, "notes:write"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_w"); err != nil {
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
		content, _ := args["content"].(string)

		note, err := noteSvc.WithClient(getClientInfoFromContext(ctx)).AppendContent(ctx, uid, &dto.NoteAppendRequest{
			Vault:    vault,
			Path:     path,
			PathHash: util.EncodeHash32(path),
			Content:  content,
			Revision: mcpRevisionArg(args),
		})

		if err != nil {
			return noteMCPError(err), nil
		}

		wss.BroadcastToUser(uid, code.Success.WithData(note).WithVault(vault), "NoteSyncModify")
		fallback := fmt.Sprintf("Appended content to %s", note.Path)
		return mcp.NewToolResultStructured(mcpNoteMutationOutput{
			Vault:     vault,
			Operation: "append",
			Note:      note.ToMcpNoteDTO(),
		}, fallback), nil
	})

	// 5. Prepend
	toolPrepend := mcp.NewTool("note_prepend",
		mcp.WithDescription("Prepend content to the beginning of a note (after frontmatter)"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Note path")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Content to prepend")),
		mcp.WithNumber("revision", mcp.Description("Optional concurrency guard: the `revision` you read from note_get/note_list. The write is applied only if the note is still at that revision; otherwise it fails with a conflict (code 452) and nothing is changed. Omit for a legacy unconditional write.")),
		mcp.WithOutputSchema[mcpNoteMutationOutput](),
	)
	srv.AddTool(writeMCPTool(toolPrepend, cfg, false, "notes:write"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_w"); err != nil {
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
		content, _ := args["content"].(string)

		note, err := noteSvc.WithClient(getClientInfoFromContext(ctx)).PrependContent(ctx, uid, &dto.NotePrependRequest{
			Vault:    vault,
			Path:     path,
			PathHash: util.EncodeHash32(path),
			Content:  content,
			Revision: mcpRevisionArg(args),
		})

		if err != nil {
			return noteMCPError(err), nil
		}

		wss.BroadcastToUser(uid, code.Success.WithData(note).WithVault(vault), "NoteSyncModify")
		fallback := fmt.Sprintf("Prepended content to %s", note.Path)
		return mcp.NewToolResultStructured(mcpNoteMutationOutput{
			Vault:     vault,
			Operation: "prepend",
			Note:      note.ToMcpNoteDTO(),
		}, fallback), nil
	})

	// 6. Replace
	toolReplace := mcp.NewTool("note_replace",
		mcp.WithDescription("Find and replace text in a note"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Note path")),
		mcp.WithString("find", mcp.Required(), mcp.Description("Content to find")),
		mcp.WithString("replace", mcp.Required(), mcp.Description("Content to replace with")),
		mcp.WithNumber("revision", mcp.Description("Optional concurrency guard: the `revision` you read from note_get/note_list. The write is applied only if the note is still at that revision; otherwise it fails with a conflict (code 452) and nothing is changed. Omit for a legacy unconditional write.")),
		mcp.WithBoolean("regex", mcp.Description("Use regex matching (default false)")),
		mcp.WithBoolean("all", mcp.Description("Replace all matches (default true)")),
		mcp.WithBoolean("failIfNoMatch", mcp.Description("Fail if no match (default true)")),
		mcp.WithOutputSchema[mcpNoteReplaceOutput](),
	)
	srv.AddTool(writeMCPTool(toolReplace, cfg, true, "notes:write"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_w"); err != nil {
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
		find, _ := args["find"].(string)
		replace, _ := args["replace"].(string)
		regex, okRegex := args["regex"].(bool)
		if !okRegex {
			regex = false
		}
		all, okAll := args["all"].(bool)
		if !okAll {
			all = true
		}
		failIfNoMatch, okFail := args["failIfNoMatch"].(bool)
		if !okFail {
			failIfNoMatch = true
		}

		res, err := noteSvc.WithClient(getClientInfoFromContext(ctx)).ReplaceContent(ctx, uid, &dto.NoteReplaceRequest{
			Vault:         vault,
			Path:          path,
			PathHash:      util.EncodeHash32(path),
			Find:          find,
			Replace:       replace,
			Regex:         regex,
			All:           all,
			FailIfNoMatch: failIfNoMatch,
			Revision:      mcpRevisionArg(args),
		})

		if err != nil {
			return noteMCPError(err), nil
		}

		wss.BroadcastToUser(uid, code.Success.WithData(res.Note).WithVault(vault), "NoteSyncModify")
		fallback := fmt.Sprintf("Replaced %d occurrences", res.MatchCount)
		return mcp.NewToolResultStructured(mcpNoteReplaceOutput{
			Vault:      vault,
			MatchCount: res.MatchCount,
			Note:       res.Note.ToMcpNoteDTO(),
		}, fallback), nil
	})

	// 7. Get Backlinks
	toolGetBacklinks := mcp.NewTool("note_get_backlinks",
		mcp.WithDescription("Get backlinks to a note"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Note path")),
		mcp.WithOutputSchema[mcpNoteLinksOutput](),
	)
	srv.AddTool(readOnlyMCPTool(toolGetBacklinks, cfg, "notes:read"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_r"); err != nil {
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

		linkSvc := appContainer.NoteLinkService
		links, err := linkSvc.GetBacklinks(ctx, uid, &dto.NoteLinkQueryRequest{
			Vault:    vault,
			Path:     path,
			PathHash: util.EncodeHash32(path),
		})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		b, err := json.Marshal(links)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultStructured(mcpNoteLinksOutput{
			Vault: vault,
			Path:  path,
			Count: len(links),
			Links: links,
		}, string(b)), nil
	})

	// 8. Get Outlinks
	toolGetOutlinks := mcp.NewTool("note_get_outlinks",
		mcp.WithDescription("Get outlinks from a note"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Note path")),
		mcp.WithOutputSchema[mcpNoteLinksOutput](),
	)
	srv.AddTool(readOnlyMCPTool(toolGetOutlinks, cfg, "notes:read"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := checkPermission(ctx, "note_r"); err != nil {
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

		linkSvc := appContainer.NoteLinkService
		links, err := linkSvc.GetOutlinks(ctx, uid, &dto.NoteLinkQueryRequest{
			Vault:    vault,
			Path:     path,
			PathHash: util.EncodeHash32(path),
		})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		b, err := json.Marshal(links)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultStructured(mcpNoteLinksOutput{
			Vault: vault,
			Path:  path,
			Count: len(links),
			Links: links,
		}, string(b)), nil
	})
}
