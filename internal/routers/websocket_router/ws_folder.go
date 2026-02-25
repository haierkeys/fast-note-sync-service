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
	var cFoldersKeys map[string]struct{} = make(map[string]struct{}, 0)
	if len(params.Folders) > 0 {
		for _, f := range params.Folders {
			cFolders[f.PathHash] = f
		}
		for _, folder := range params.Folders {
			cFoldersKeys[folder.PathHash] = struct{}{}
		}
	}

	var messageQueue []dto.WSQueuedMessage
	var needModifyCount int64
	var needDeleteCount int64
	var cDelFoldersKeys map[string]struct{} = make(map[string]struct{})

	// Handle deleted folders from client
	if len(params.DelFolders) > 0 {
		for _, delFolder := range params.DelFolders {

			// Check if folder exists before deleting
			checkFolder, err := h.App.FolderService.Get(ctx, uid, &dto.FolderGetRequest{
				Vault:    params.Vault,
				PathHash: delFolder.PathHash,
			})

			if err == nil && checkFolder != nil && checkFolder.Action != "delete" {
				delParams := &dto.FolderDeleteRequest{
					Vault:    params.Vault,
					Path:     delFolder.Path,
					PathHash: delFolder.PathHash,
				}
				folder, err := h.App.FolderService.Delete(ctx, uid, delParams)
				if err != nil {
					h.App.Logger().Error("websocket_router.folder.FolderSync.FolderService.Delete",
						zap.String(logpkg.FieldTraceID, c.TraceID),
						zap.Int64(logpkg.FieldUID, uid),
						zap.String(logpkg.FieldPath, delFolder.Path),
						zap.Error(err))
					continue
				}
				cDelFoldersKeys[delFolder.PathHash] = struct{}{}
				// Broadcast deletion to other clients
				c.BroadcastResponse(code.Success.WithData(
					dto.FolderSyncDeleteMessage{
						Path:     folder.Path,
						PathHash: folder.PathHash,
						Ctime:    folder.Ctime,
						Mtime:    folder.Mtime,
					},
				).WithVault(params.Vault).WithContext(params.Context), true, dto.FolderSyncDelete)
			} else {
				h.App.Logger().Debug("websocket_router.folder.FolderSync.FolderService.Get check failed (not found or already deleted), broadcasting delete anyway",
					zap.String(logpkg.FieldTraceID, c.TraceID),
					zap.String("pathHash", delFolder.PathHash))

				cDelFoldersKeys[delFolder.PathHash] = struct{}{}
				// Broadcast deletion with available info
				c.BroadcastResponse(code.Success.WithData(
					dto.FolderSyncDeleteMessage{
						Path:     delFolder.Path,
						PathHash: delFolder.PathHash,
						Ctime:    0,
						Mtime:    0,
					},
				).WithVault(params.Vault).WithContext(params.Context), true, dto.FolderSyncDelete)
			}

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
				h.App.Logger().Warn("websocket_router.folder.FolderSync.FolderService.Get",
					zap.String(logpkg.FieldTraceID, c.TraceID),
					zap.String("pathHash", missingFolder.PathHash),
					zap.Error(err))
				continue
			}
			if folder != nil && folder.Action != "delete" {
				messageQueue = append(messageQueue, dto.WSQueuedMessage{
					Action: dto.FolderSyncModify,
					Data: dto.FolderSyncModifyMessage{
						Path:             folder.Path,
						PathHash:         folder.PathHash,
						Ctime:            folder.Ctime,
						Mtime:            folder.Mtime,
						UpdatedTimestamp: folder.UpdatedTimestamp,
					},
				})
				needModifyCount++
				cDelFoldersKeys[folder.PathHash] = struct{}{}
			}
		}
	}

	// Get updated folders from server
	list, err := h.App.FolderService.ListByUpdatedTimestamp(ctx, uid, params.Vault, params.LastTime)
	if err != nil {
		h.respondError(c, code.ErrorFolderListFailed, err, "websocket_router.folder.FolderSync.ListByUpdatedTimestamp")
		return
	}

	for _, folder := range list {
		if _, ok := cDelFoldersKeys[folder.PathHash]; ok {
			continue
		}

		if folder.Action == "delete" {
			delete(cFoldersKeys, folder.PathHash)
			messageQueue = append(messageQueue, dto.WSQueuedMessage{
				Action: dto.FolderSyncDelete,
				Data: dto.FolderSyncDeleteMessage{
					Path:     folder.Path,
					PathHash: folder.PathHash,
					Ctime:    folder.Ctime,
					Mtime:    folder.Mtime,
				},
			})
			needDeleteCount++
		} else {
			delete(cFoldersKeys, folder.PathHash)
			_, exists := cFolders[folder.PathHash]
			if !exists {

				messageQueue = append(messageQueue, dto.WSQueuedMessage{
					Action: dto.FolderSyncModify,
					Data: dto.FolderSyncModifyMessage{
						Path:             folder.Path,
						PathHash:         folder.PathHash,
						Ctime:            folder.Ctime,
						Mtime:            folder.Mtime,
						UpdatedTimestamp: folder.UpdatedTimestamp,
					},
				})
				needModifyCount++
			}
		}
	}

	if len(cFoldersKeys) > 0 {
		for pathHash := range cFoldersKeys {
			folder := cFolders[pathHash]

			newFolder, err := h.App.FolderService.UpdateOrCreate(ctx, uid, &dto.FolderCreateRequest{
				Vault:    params.Vault,
				Path:     folder.Path,
				PathHash: folder.PathHash,
			})
			if err != nil {
				h.logError(c, "websocket_router.folder.FolderSync.UpdateOrCreate", err)
				continue
			}
			c.BroadcastResponse(code.Success.WithData(
				dto.FolderSyncModifyMessage{
					Path:             newFolder.Path,
					PathHash:         newFolder.PathHash,
					Ctime:            newFolder.Ctime,
					Mtime:            newFolder.Mtime,
					UpdatedTimestamp: newFolder.UpdatedTimestamp,
				},
			).WithVault(params.Vault).WithContext(params.Context), true, dto.FolderSyncModify)
		}
	}

	// Send FolderSyncEnd message
	// 发送 FolderSyncEnd 消息
	c.ToResponse(code.Success.WithData(&dto.FolderSyncEndMessage{
		LastTime:        timex.Now().UnixMilli(),
		NeedModifyCount: needModifyCount,
		NeedDeleteCount: needDeleteCount,
	}).WithVault(params.Vault).WithContext(params.Context), dto.FolderSyncEnd)

	// Send queued messages individually
	// 逐条发送队列中的消息
	for _, item := range messageQueue {
		c.ToResponse(code.Success.WithData(item.Data).WithVault(params.Vault).WithContext(params.Context), item.Action)
	}
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
		h.respondError(c, code.ErrorFolderModifyOrCreateFailed, err, "websocket_router.folder.FolderModify.UpdateOrCreate")
		return
	}

	c.ToResponse(code.Success)
	c.BroadcastResponse(code.Success.WithData(
		dto.FolderSyncModifyMessage{
			Path:             folder.Path,
			PathHash:         folder.PathHash,
			Ctime:            folder.Ctime,
			Mtime:            folder.Mtime,
			UpdatedTimestamp: folder.UpdatedTimestamp,
		},
	).WithVault(params.Vault), true, dto.FolderSyncModify)
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
	folder, err := h.App.FolderService.Delete(c.Context(), uid, params)
	if err != nil {
		h.respondError(c, code.ErrorFolderDeleteFailed, err, "websocket_router.folder.FolderDelete.Delete")
		return
	}

	c.ToResponse(code.Success)
	c.BroadcastResponse(code.Success.WithData(
		dto.FolderSyncDeleteMessage{
			Path:     folder.Path,
			PathHash: folder.PathHash,
			Ctime:    folder.Ctime,
			Mtime:    folder.Mtime,
		},
	).WithVault(params.Vault), true, dto.FolderSyncDelete)
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
		h.respondError(c, code.ErrorFolderRenameFailed, err, "websocket_router.folder.FolderRename.Rename")
		return
	}

	c.ToResponse(code.Success)

	// 如果 oldFolder 为空，说明是新增文件夹
	if oldFolder == nil {
		c.BroadcastResponse(code.Success.WithData(
			dto.FolderSyncModifyMessage{
				Path:             newFolder.Path,
				PathHash:         newFolder.PathHash,
				Ctime:            newFolder.Ctime,
				Mtime:            newFolder.Mtime,
				UpdatedTimestamp: newFolder.UpdatedTimestamp,
			},
		).WithVault(params.Vault), true, dto.FolderSyncModify)
		return
	}

	c.BroadcastResponse(code.Success.WithData(dto.FolderSyncRenameMessage{
		Path:        newFolder.Path,
		PathHash:    newFolder.PathHash,
		Ctime:       newFolder.Ctime,
		Mtime:       newFolder.Mtime,
		OldPath:     oldFolder.Path,
		OldPathHash: oldFolder.PathHash,
	}).WithVault(params.Vault), true, dto.FolderSyncRename)

}
