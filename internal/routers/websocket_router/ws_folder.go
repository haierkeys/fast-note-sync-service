package websocket_router

import (
	"time"

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
	valid, errs := c.BindAndValidWithAction(msg.Type, msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.folder.FolderSync.BindAndValid", msg)
		return
	}

	// 分批协议：totalBatches > 1 时归集到缓存，集齐后统一执行差量同步
	// Batch protocol: accumulate into cache when totalBatches > 1, then run diff sync when all collected
	if params.TotalBatches > 1 {
		entry := syncBatchGetOrCreate(params.Context, "folder", params.TotalBatches)

		entry.mu.Lock()
		for _, f := range params.Folders {
			entry.Items = append(entry.Items, f)
		}
		entry.ReceivedCount++
		for _, df := range params.DelFolders {
			entry.DelItems = append(entry.DelItems, df)
		}
		for _, mf := range params.MissingFolders {
			entry.MissingItems = append(entry.MissingItems, mf)
		}
		entry.UpdatedAt = time.Now()
		received := entry.ReceivedCount
		total := entry.TotalBatches
		entry.mu.Unlock()

		if received < total {
			c.ToResponse(code.Success.WithData(map[string]interface{}{
				"context":    params.Context,
				"batchIndex": params.BatchIndex,
			}).WithVault(params.Vault).WithContext(params.Context), FolderSyncBatchAck)
			return
		}

		syncBatchDelete(params.Context, "folder")
		allFolders := make([]dto.FolderSyncCheckRequest, 0, len(entry.Items))
		for _, item := range entry.Items {
			allFolders = append(allFolders, item.(dto.FolderSyncCheckRequest))
		}
		params.Folders = allFolders

		allDelFolders := make([]dto.FolderSyncDelFolder, 0, len(entry.DelItems))
		for _, item := range entry.DelItems {
			allDelFolders = append(allDelFolders, item.(dto.FolderSyncDelFolder))
		}
		params.DelFolders = allDelFolders

		allMissingFolders := make([]dto.FolderSyncDelFolder, 0, len(entry.MissingItems))
		for _, item := range entry.MissingItems {
			allMissingFolders = append(allMissingFolders, item.(dto.FolderSyncDelFolder))
		}
		params.MissingFolders = allMissingFolders
	}

	h.doFolderSync(c, params)
}

// doFolderSync 执行文件夹差量同步核心逻辑（原 FolderSync 函数体）
// doFolderSync runs the core folder differential sync logic
func (h *FolderWSHandler) doFolderSync(c *pkgapp.WebsocketClient, params *dto.FolderSyncRequest) {
	ctx := c.Context()
	uid := c.User.UID

	// Check and create vault
	h.App.VaultService.GetOrCreate(ctx, uid, params.Vault)

	folderSvc := h.App.GetFolderService(c.ClientType, c.ClientName, c.ClientVersion)

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
		hasWritePermission := pkgapp.VerifyPermissions(c.Scope, "ws", c.ClientType, "note_w")

		for _, delFolder := range params.DelFolders {

			// Check if folder exists before deleting
			checkFolder, err := folderSvc.Get(ctx, uid, &dto.FolderGetRequest{
				Vault:    params.Vault,
				PathHash: delFolder.PathHash,
			})

			if err == nil && checkFolder != nil && checkFolder.Action != "delete" {
				if !hasWritePermission {
					h.App.Logger().Warn("websocket_router.folder.FolderSync: permission denied for deletion",
						zap.String(logpkg.FieldTraceID, c.TraceID),
						zap.Int64(logpkg.FieldUID, uid),
						zap.String(logpkg.FieldPath, delFolder.Path))
					continue
				}

				delParams := &dto.FolderDeleteRequest{
					Vault:    params.Vault,
					Path:     delFolder.Path,
					PathHash: delFolder.PathHash,
				}
				folder, err := folderSvc.Delete(ctx, uid, delParams)
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
						Path:             folder.Path,
						PathHash:         folder.PathHash,
						Ctime:            folder.Ctime,
						Mtime:            folder.Mtime,
						UpdatedTimestamp: folder.UpdatedTimestamp,
					},
				).WithVault(params.Vault).WithContext(params.Context), true, FolderSyncDelete)
			} else {
				h.App.Logger().Debug("websocket_router.folder.FolderSync.FolderService.Get check failed (not found or already deleted), broadcasting delete anyway",
					zap.String(logpkg.FieldTraceID, c.TraceID),
					zap.String("pathHash", delFolder.PathHash))

				cDelFoldersKeys[delFolder.PathHash] = struct{}{}
				// Broadcast deletion with available info
				c.BroadcastResponse(code.Success.WithData(
					dto.FolderSyncDeleteMessage{
						Path:             delFolder.Path,
						PathHash:         delFolder.PathHash,
						Ctime:            0,
						Mtime:            0,
						UpdatedTimestamp: 0,
					},
				).WithVault(params.Vault).WithContext(params.Context), true, FolderSyncDelete)
			}

		}
	}

	// Handle missing folders on client
	if params.LastTime > 0 && len(params.MissingFolders) > 0 {
		for _, missingFolder := range params.MissingFolders {
			folder, err := folderSvc.Get(ctx, uid, &dto.FolderGetRequest{
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
						Context: params.Context,
					Action: FolderSyncModify,
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
	// 获取服务端更新的文件夹列表

	// Record sync start time before querying to avoid missing writes that occur during query processing.
	// 查询前记录同步开始时间，防止查询处理期间的写入被遗漏（经典增量同步快照时间戳方案）。
	syncStartTime := timex.Now().UnixMilli()

	list, err := folderSvc.ListByUpdatedTimestamp(ctx, uid, params.Vault, params.LastTime)
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
						Context: params.Context,
				Action: FolderSyncDelete,
				Data: dto.FolderSyncDeleteMessage{
					Path:             folder.Path,
					PathHash:         folder.PathHash,
					Ctime:            folder.Ctime,
					Mtime:            folder.Mtime,
					UpdatedTimestamp: folder.UpdatedTimestamp,
				},
			})
			needDeleteCount++
		} else {
			delete(cFoldersKeys, folder.PathHash)
			_, exists := cFolders[folder.PathHash]
			if !exists {

				messageQueue = append(messageQueue, dto.WSQueuedMessage{
						Context: params.Context,
					Action: FolderSyncModify,
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

			newFolder, err := folderSvc.UpdateOrCreate(ctx, uid, &dto.FolderCreateRequest{
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
			).WithVault(params.Vault).WithContext(params.Context), true, FolderSyncModify)
		}
	}

	// Send FolderSyncEnd message
	// 发送 FolderSyncEnd 消息
	c.ToResponse(code.Success.WithData(&dto.FolderSyncEndMessage{
		// Use syncStartTime (recorded before query) to prevent writes during query processing from being missed.
		// 使用查询前记录 of syncStartTime，防止查询处理期间的写入在下次增量同步时被永久遗漏。
		LastTime:        syncStartTime,
		NeedModifyCount: needModifyCount,
		NeedDeleteCount: needDeleteCount,
	}).WithVault(params.Vault).WithContext(params.Context), FolderSyncEnd)

	// 在 End 消息后，启动受控分页发送流程
	if len(messageQueue) > 0 {
		pageSize := h.App.Config().App.SyncDownChunkNum
		if pageSize <= 0 {
			pageSize = 50 // 默认值防呆
		}
		entry := &syncDownloadEntry{
			Context:      params.Context,
			TypeName:     "folder",
			Vault:        params.Vault,
			MessageQueue: messageQueue,
			PageSize:     pageSize,
			CurrentPage:  0,
		}
		syncDownloadStore(params.Context, "folder", entry)
		// 默认不自动发送，等待客户端拉取
	}
}

// FolderSyncPageAck handles WebSocket messages for client page ACK
// FolderSyncPageAck 处理客户端发来的分页下载 ACK 消息
func (h *FolderWSHandler) FolderSyncPageAck(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.SyncPageAckRequest{}
	valid, errs := c.BindAndValidWithAction(msg.Type, msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.folder.FolderSyncPageAck.BindAndValid", msg)
		return
	}

	entry, ok := syncDownloadGet(params.Context, "folder")
	if !ok {
		h.App.Logger().Warn("FolderSyncPageAck: sync download entry not found",
			zap.String(logpkg.FieldTraceID, c.TraceID),
			zap.String("context", params.Context))
		return
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	if params.PageIndex == -1 {
		sendSyncPage(c, entry)
		return
	}

	if params.PageIndex != entry.CurrentPage {
		h.App.Logger().Warn("FolderSyncPageAck: page index mismatch",
			zap.String(logpkg.FieldTraceID, c.TraceID),
			zap.Int("expected", entry.CurrentPage),
			zap.Int("got", params.PageIndex))
		return
	}

	start := entry.CurrentPage * entry.PageSize
	end := start + entry.PageSize
	if end >= len(entry.MessageQueue) {
		syncDownloadDelete(params.Context, "folder")
		h.App.Logger().Info("FolderSyncPageAck: sync finished, cache cleared",
			zap.String(logpkg.FieldTraceID, c.TraceID),
			zap.String("context", params.Context))
		return
	}

	entry.CurrentPage++
	entry.UpdatedAt = time.Now()
	sendSyncPage(c, entry)
}


// FolderModify handles folder modification/creation
func (h *FolderWSHandler) FolderModify(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.FolderCreateRequest{}
	valid, errs := c.BindAndValidWithAction(msg.Type, msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.folder.FolderModify.BindAndValid", msg)
		return
	}

	uid := c.User.UID
	folder, err := h.App.GetFolderService(c.ClientType, c.ClientName, c.ClientVersion).UpdateOrCreate(c.Context(), uid, params)
	if err != nil {
		h.respondError(c, code.ErrorFolderModifyOrCreateFailed, err, "websocket_router.folder.FolderModify.UpdateOrCreate")
		return
	}

	c.ToResponse(code.Success.WithData(dto.FolderModifyAckMessage{
		LastTime: folder.UpdatedTimestamp,
		Path:     folder.Path,
		PathHash: folder.PathHash,
	}).WithVault(params.Vault).WithContext(params.Context), string(FolderModifyAck))
	c.BroadcastResponse(code.Success.WithData(
		dto.FolderSyncModifyMessage{
			Path:             folder.Path,
			PathHash:         folder.PathHash,
			Ctime:            folder.Ctime,
			Mtime:            folder.Mtime,
			UpdatedTimestamp: folder.UpdatedTimestamp,
		},
	).WithVault(params.Vault), true, FolderSyncModify)
}

// FolderDelete handles folder deletion
// 删除
func (h *FolderWSHandler) FolderDelete(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.FolderDeleteRequest{}
	valid, errs := c.BindAndValidWithAction(msg.Type, msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.folder.FolderDelete.BindAndValid", msg)
		return
	}

	uid := c.User.UID
	folder, err := h.App.GetFolderService(c.ClientType, c.ClientName, c.ClientVersion).Delete(c.Context(), uid, params)
	if err != nil {
		h.respondError(c, code.ErrorFolderDeleteFailed, err, "websocket_router.folder.FolderDelete.Delete")
		return
	}

	c.ToResponse(code.Success.WithData(dto.FolderDeleteAckMessage{
		LastTime: folder.UpdatedTimestamp,
		Path:     folder.Path,
		PathHash: folder.PathHash,
	}).WithVault(params.Vault).WithContext(params.Context), string(FolderDeleteAck))
	c.BroadcastResponse(code.Success.WithData(
		dto.FolderSyncDeleteMessage{
			Path:             folder.Path,
			PathHash:         folder.PathHash,
			Ctime:            folder.Ctime,
			Mtime:            folder.Mtime,
			UpdatedTimestamp: folder.UpdatedTimestamp,
		},
	).WithVault(params.Vault), true, FolderSyncDelete)
}

// FolderRename handles folder renaming
// 重命名文件夹
func (h *FolderWSHandler) FolderRename(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.FolderRenameRequest{}
	valid, errs := c.BindAndValidWithAction(msg.Type, msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.folder.FolderRename.BindAndValid", msg)
		return
	}

	uid := c.User.UID
	folderSvc := h.App.GetFolderService(c.ClientType, c.ClientName, c.ClientVersion)
	oldFolder, newFolder, err := folderSvc.Rename(c.Context(), uid, params)
	if err != nil {
		h.respondError(c, code.ErrorFolderRenameFailed, err, "websocket_router.folder.FolderRename.Rename")
		return
	}

	c.ToResponse(code.Success.WithData(dto.FolderRenameAckMessage{
		LastTime: newFolder.UpdatedTimestamp,
		Path:     newFolder.Path,
		PathHash: newFolder.PathHash,
	}).WithVault(params.Vault).WithContext(params.Context), string(FolderRenameAck))

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
		).WithVault(params.Vault), true, FolderSyncModify)
		return
	}

	c.BroadcastResponse(code.Success.WithData(dto.FolderSyncRenameMessage{
		Path:             newFolder.Path,
		PathHash:         newFolder.PathHash,
		Ctime:            newFolder.Ctime,
		Mtime:            newFolder.Mtime,
		OldPath:          oldFolder.Path,
		OldPathHash:      oldFolder.PathHash,
		UpdatedTimestamp: newFolder.UpdatedTimestamp,
	}).WithVault(params.Vault), true, FolderSyncRename)

}
