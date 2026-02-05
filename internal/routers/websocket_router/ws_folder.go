package websocket_router

import (
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	logpkg "github.com/haierkeys/fast-note-sync-service/pkg/logger"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"go.uber.org/zap"
)

type FolderWSHandler struct {
	*WSHandler
}

func NewFolderWSHandler(a *app.App) *FolderWSHandler {
	return &FolderWSHandler{WSHandler: NewWSHandler(a)}
}

// FolderSync handles folder synchronization
func (h *FolderWSHandler) FolderSync(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.FolderSyncRequest{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.folder.FolderSync.BindAndValid")
		return
	}

	ctx := c.Context()
	uid := c.User.UID

	// Check and create vault
	h.App.VaultService.GetOrCreate(ctx, uid, params.Vault)

	var cFolders map[string]dto.FolderSyncCheckRequest = make(map[string]dto.FolderSyncCheckRequest)
	if len(params.Folders) > 0 {
		for _, f := range params.Folders {
			cFolders[f.PathHash] = f
		}
	}

	var messageQueue []dto.WSQueuedMessage
	var needModifyCount int64
	var needDeleteCount int64
	var cDelFoldersKeys map[string]struct{} = make(map[string]struct{})

	// Handle deleted folders from client
	if len(params.DelFolders) > 0 {
		for _, delFolder := range params.DelFolders {
			delParams := &dto.FolderDeleteRequest{
				Vault:    params.Vault,
				Path:     delFolder.Path,
				PathHash: delFolder.PathHash,
			}
			err := h.App.FolderService.Delete(ctx, uid, delParams)
			if err != nil {
				h.App.Logger().Error("failed to delete folder during sync",
					zap.String(logpkg.FieldTraceID, c.TraceID),
					zap.Int64(logpkg.FieldUID, uid),
					zap.String(logpkg.FieldPath, delFolder.Path),
					zap.Error(err))
				continue
			}
			cDelFoldersKeys[delFolder.PathHash] = struct{}{}
			// Broadcast deletion to other clients
			c.BroadcastResponse(code.Success.WithData(delFolder).WithVault(params.Vault), true, dto.FolderSyncDelete)
		}
	}

	// Handle missing folders on client
	if params.LastTime > 0 && len(params.MissingFolders) > 0 {
		for _, missingFolder := range params.MissingFolders {
			folder, err := h.App.FolderService.Get(ctx, uid, &dto.FolderGetRequest{
				Vault:    params.Vault,
				Path:     missingFolder.Path,
				PathHash: missingFolder.PathHash,
			})
			if err != nil {
				h.App.Logger().Warn("failed to fetch missing folder during sync",
					zap.String(logpkg.FieldTraceID, c.TraceID),
					zap.String("pathHash", missingFolder.PathHash),
					zap.Error(err))
				continue
			}
			if folder != nil && folder.Action != "delete" {
				messageQueue = append(messageQueue, dto.WSQueuedMessage{
					Action: dto.FolderSyncModify,
					Data:   folder,
				})
				needModifyCount++
				cDelFoldersKeys[folder.PathHash] = struct{}{}
			}
		}
	}

	// Get updated folders from server
	list, err := h.App.FolderService.ListByUpdatedTimestamp(ctx, uid, params.Vault, params.LastTime)
	if err != nil {
		h.respondError(c, code.ErrorDBQuery, err, "FolderSync")
		return
	}

	for _, folder := range list {
		if _, ok := cDelFoldersKeys[folder.PathHash]; ok {
			continue
		}

		if folder.Action == "delete" {
			messageQueue = append(messageQueue, dto.WSQueuedMessage{
				Action: dto.FolderSyncDelete,
				Data:   folder,
			})
			needDeleteCount++
		} else {
			// Compare with client version
			_, exists := cFolders[folder.PathHash]
			if !exists {
				messageQueue = append(messageQueue, dto.WSQueuedMessage{
					Action: dto.FolderSyncModify,
					Data:   folder,
				})
				needModifyCount++
			}
		}
	}

	c.ToResponse(code.Success.WithData(&dto.FolderSyncEndMessage{
		LastTime:        timex.Now().UnixMilli(),
		NeedModifyCount: needModifyCount,
		NeedDeleteCount: needDeleteCount,
		Messages:        messageQueue,
	}).WithVault(params.Vault), dto.FolderSyncEnd)
}

// FolderModify handles folder modification/creation
func (h *FolderWSHandler) FolderModify(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.FolderCreateRequest{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.folder.FolderModify.BindAndValid")
		return
	}

	uid := c.User.UID
	folder, err := h.App.FolderService.UpdateOrCreate(c.Context(), uid, params)
	if err != nil {
		h.respondError(c, code.ErrorDBQuery, err, "FolderModify")
		return
	}

	c.ToResponse(code.Success.WithData(folder))
	c.BroadcastResponse(code.Success.WithData(folder).WithVault(params.Vault), true, dto.FolderSyncModify)
}

// FolderDelete handles folder deletion
// 删除
func (h *FolderWSHandler) FolderDelete(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.FolderDeleteRequest{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.folder.FolderDelete.BindAndValid")
		return
	}

	uid := c.User.UID
	err := h.App.FolderService.Delete(c.Context(), uid, params)
	if err != nil {
		h.respondError(c, code.ErrorDBQuery, err, "FolderDelete")
		return
	}

	c.ToResponse(code.Success)
	c.BroadcastResponse(code.Success.WithData(params).WithVault(params.Vault), true, dto.FolderSyncDelete)
}

// FolderRename handles folder renaming
// 重命名文件夹
func (h *FolderWSHandler) FolderRename(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.FolderRenameRequest{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.folder.FolderRename.BindAndValid")
		return
	}

	uid := c.User.UID
	oldFolder, newFolder, err := h.App.FolderService.Rename(c.Context(), uid, params)
	if err != nil {
		h.respondError(c, code.ErrorDBQuery, err, "FolderRename")
		return
	}

	c.ToResponse(code.Success.WithData(newFolder))

	c.BroadcastResponse(code.Success.WithData(dto.FolderSyncRenameMessage{
		Path:        newFolder.Path,
		PathHash:    newFolder.PathHash,
		Ctime:       newFolder.Ctime,
		Mtime:       newFolder.Mtime,
		OldPath:     oldFolder.Path,
		OldPathHash: oldFolder.PathHash,
	}).WithVault(params.Vault), true, dto.FolderSyncRename)

}
