package mcp_router

import (
	"context"
	"encoding/json"
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

func registerNoteTools(srv *mcpsrv.MCPServer, appContainer *app.App, wss *pkgapp.WebsocketServer) {
	noteSvc := appContainer.NoteService

	// 1. List Notes
	toolListNotes := mcp.NewTool("note_list",
		mcp.WithDescription("List notes in a vault"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("keyword", mcp.Description("Search keyword")),
	)
	srv.AddTool(toolListNotes, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		if vault == "" || strings.EqualFold(vault, "default") {
			vault = getDefaultVaultName(ctx, appContainer)
		}
		keyword, _ := args["keyword"].(string)

		pager := &pkgapp.Pager{
			Page:     pkgapp.GetPage(1),
			PageSize: pkgapp.GetPageSize(100),
		}
		notes, _, err := noteSvc.WithClient(getClientInfoFromContext(ctx)).List(ctx, uid, &dto.NoteListRequest{
			Vault:   vault,
			Keyword: keyword,
		}, pager)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resStr := fmt.Sprintf("Found %d notes:\n", len(notes))
		for _, n := range notes {
			resStr += fmt.Sprintf("- %s (ID: %d, Size: %d, Mtime: %d)\n", n.Path, n.ID, n.Size, n.Mtime)
		}
		return mcp.NewToolResultText(resStr), nil
	})

	// 2. Get Note
	toolGetNote := mcp.NewTool("note_get",
		mcp.WithDescription("Get a single note by path"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Note path")),
	)
	srv.AddTool(toolGetNote, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		if vault == "" || strings.EqualFold(vault, "default") {
			vault = getDefaultVaultName(ctx, appContainer)
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

		return mcp.NewToolResultText(note.Content), nil
	})

	// 3. Create or Update Note
	toolCreateUpdateNote := mcp.NewTool("note_create_or_update",
		mcp.WithDescription("Create or update a note"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Note path")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Note content")),
	)
	srv.AddTool(toolCreateUpdateNote, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		if vault == "" || strings.EqualFold(vault, "default") {
			vault = getDefaultVaultName(ctx, appContainer)
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
		}, false)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		wss.BroadcastToUser(uid, code.Success.WithData(note).WithVault(vault), "NoteSyncModify")
		return mcp.NewToolResultText(fmt.Sprintf("Successfully saved note: %s (Version: %d)", note.Path, note.Version)), nil
	})

	// 4. Delete Note
	toolDeleteNote := mcp.NewTool("note_delete",
		mcp.WithDescription("Delete a note"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Note path")),
	)
	srv.AddTool(toolDeleteNote, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		if vault == "" || strings.EqualFold(vault, "default") {
			vault = getDefaultVaultName(ctx, appContainer)
		}
		path, _ := args["path"].(string)
		pathHash := util.EncodeHash32(path)

		note, err := noteSvc.WithClient(getClientInfoFromContext(ctx)).Delete(ctx, uid, &dto.NoteDeleteRequest{
			Vault:    vault,
			Path:     path,
			PathHash: pathHash,
		})

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		wss.BroadcastToUser(uid, code.Success.WithData(note).WithVault(vault), "NoteSyncDelete")
		return mcp.NewToolResultText(fmt.Sprintf("Deleted note: %s", note.Path)), nil
	})

	// 5. Rename Note
	toolRenameNote := mcp.NewTool("note_rename",
		mcp.WithDescription("Rename a note"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("oldPath", mcp.Required(), mcp.Description("Old note path")),
		mcp.WithString("newPath", mcp.Required(), mcp.Description("New note path")),
	)
	srv.AddTool(toolRenameNote, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		if vault == "" || strings.EqualFold(vault, "default") {
			vault = getDefaultVaultName(ctx, appContainer)
		}
		oldPath, _ := args["oldPath"].(string)
		newPath, _ := args["newPath"].(string)

		oldNote, newNote, err := noteSvc.WithClient(getClientInfoFromContext(ctx)).Rename(ctx, uid, &dto.NoteRenameRequest{
			Vault:       vault,
			OldPath:     oldPath,
			OldPathHash: util.EncodeHash32(oldPath),
			Path:        newPath,
			PathHash:    util.EncodeHash32(newPath),
		})

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
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
		return mcp.NewToolResultText(fmt.Sprintf("Renamed note from %s to %s", oldNote.Path, newNote.Path)), nil
	})

	// 1. Restore Note
	toolRestoreNote := mcp.NewTool("note_restore",
		mcp.WithDescription("Restore a deleted note from recycle bin"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Note path")),
	)
	srv.AddTool(toolRestoreNote, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		if vault == "" || strings.EqualFold(vault, "default") {
			vault = getDefaultVaultName(ctx, appContainer)
		}
		path, _ := args["path"].(string)

		note, err := noteSvc.WithClient(getClientInfoFromContext(ctx)).Restore(ctx, uid, &dto.NoteRestoreRequest{
			Vault:    vault,
			Path:     path,
			PathHash: util.EncodeHash32(path),
		})

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		wss.BroadcastToUser(uid, code.Success.WithData(note).WithVault(vault), "NoteSyncModify")
		return mcp.NewToolResultText(fmt.Sprintf("Restored note: %s", note.Path)), nil
	})

	// 2. Recycle Clear Note
	toolRecycleClear := mcp.NewTool("note_recycle_clear",
		mcp.WithDescription("Permanently delete a note from recycle bin (or all if path is empty)"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Description("Note path. If empty, potentially clear all (based on service logic)")),
	)
	srv.AddTool(toolRecycleClear, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		if vault == "" || strings.EqualFold(vault, "default") {
			vault = getDefaultVaultName(ctx, appContainer)
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

		return mcp.NewToolResultText("Recycle clear successful"), nil
	})

	// 3. Patch Frontmatter
	toolPatchFrontmatter := mcp.NewTool("note_patch_frontmatter",
		mcp.WithDescription("Patch (update or remove) frontmatter of a note"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Note path")),
		mcp.WithString("updates", mcp.Description("JSON string for fields to update (e.g. {\"tags\":[\"t1\"]})")),
		mcp.WithString("remove", mcp.Description("JSON string array for fields to remove (e.g. [\"old_tag\"])")),
	)
	srv.AddTool(toolPatchFrontmatter, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		if vault == "" || strings.EqualFold(vault, "default") {
			vault = getDefaultVaultName(ctx, appContainer)
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
		})

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		wss.BroadcastToUser(uid, code.Success.WithData(note).WithVault(vault), "NoteSyncModify")
		return mcp.NewToolResultText(fmt.Sprintf("Frontmatter patched for %s", note.Path)), nil
	})

	// 4. Append
	toolAppend := mcp.NewTool("note_append",
		mcp.WithDescription("Append content to the end of a note"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Note path")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Content to append")),
	)
	srv.AddTool(toolAppend, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		if vault == "" || strings.EqualFold(vault, "default") {
			vault = getDefaultVaultName(ctx, appContainer)
		}
		path, _ := args["path"].(string)
		content, _ := args["content"].(string)

		note, err := noteSvc.WithClient(getClientInfoFromContext(ctx)).AppendContent(ctx, uid, &dto.NoteAppendRequest{
			Vault:    vault,
			Path:     path,
			PathHash: util.EncodeHash32(path),
			Content:  content,
		})

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		wss.BroadcastToUser(uid, code.Success.WithData(note).WithVault(vault), "NoteSyncModify")
		return mcp.NewToolResultText(fmt.Sprintf("Appended content to %s", note.Path)), nil
	})

	// 5. Prepend
	toolPrepend := mcp.NewTool("note_prepend",
		mcp.WithDescription("Prepend content to the beginning of a note (after frontmatter)"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Note path")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Content to prepend")),
	)
	srv.AddTool(toolPrepend, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		if vault == "" || strings.EqualFold(vault, "default") {
			vault = getDefaultVaultName(ctx, appContainer)
		}
		path, _ := args["path"].(string)
		content, _ := args["content"].(string)

		note, err := noteSvc.WithClient(getClientInfoFromContext(ctx)).PrependContent(ctx, uid, &dto.NotePrependRequest{
			Vault:    vault,
			Path:     path,
			PathHash: util.EncodeHash32(path),
			Content:  content,
		})

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		wss.BroadcastToUser(uid, code.Success.WithData(note).WithVault(vault), "NoteSyncModify")
		return mcp.NewToolResultText(fmt.Sprintf("Prepended content to %s", note.Path)), nil
	})

	// 6. Replace
	toolReplace := mcp.NewTool("note_replace",
		mcp.WithDescription("Find and replace text in a note"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Note path")),
		mcp.WithString("find", mcp.Required(), mcp.Description("Content to find")),
		mcp.WithString("replace", mcp.Required(), mcp.Description("Content to replace with")),
		mcp.WithBoolean("regex", mcp.Description("Use regex matching (default false)")),
		mcp.WithBoolean("all", mcp.Description("Replace all matches (default true)")),
		mcp.WithBoolean("failIfNoMatch", mcp.Description("Fail if no match (default true)")),
	)
	srv.AddTool(toolReplace, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		if vault == "" || strings.EqualFold(vault, "default") {
			vault = getDefaultVaultName(ctx, appContainer)
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
		})

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		wss.BroadcastToUser(uid, code.Success.WithData(res.Note).WithVault(vault), "NoteSyncModify")
		return mcp.NewToolResultText(fmt.Sprintf("Replaced %d occurrences", res.MatchCount)), nil
	})

	// 7. Get Backlinks
	toolGetBacklinks := mcp.NewTool("note_get_backlinks",
		mcp.WithDescription("Get backlinks to a note"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Note path")),
	)
	srv.AddTool(toolGetBacklinks, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		if vault == "" || strings.EqualFold(vault, "default") {
			vault = getDefaultVaultName(ctx, appContainer)
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

		return mcp.NewToolResultText(string(b)), nil
	})

	// 8. Get Outlinks
	toolGetOutlinks := mcp.NewTool("note_get_outlinks",
		mcp.WithDescription("Get outlinks from a note"),
		mcp.WithString("vault", mcp.Description("Vault name. Omitting this or providing 'default' will use the client-configured default vault.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Note path")),
	)
	srv.AddTool(toolGetOutlinks, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uid := getUIDFromContext(ctx)
		args := getArgs(req)
		vault, _ := args["vault"].(string)
		if vault == "" || strings.EqualFold(vault, "default") {
			vault = getDefaultVaultName(ctx, appContainer)
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

		return mcp.NewToolResultText(string(b)), nil
	})
}
