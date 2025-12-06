package websocket_router

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"

	"go.uber.org/zap"
)

type FileUploadCompleteParams struct {
	SessionID string `json:"sessionID" binding:"required"`
}

func FileUploadCheck(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.FileUpdateCheckParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("websocket_router.file.FileUploadCheck.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx).WithSF(c.SF)

	// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
	svc.VaultGetOrCreate(params.Vault, c.User.UID)

	updateMode, fileSvc, err := svc.FileUpdateCheck(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorNoteModifyFailed.WithDetails(err.Error()))
		return
	}

	// 需要客户端上传附件
	if updateMode == "UpdateContent" || updateMode == "Create" {

		sessionID := uuid.New().String()
		tempDir := global.Config.App.TempPath
		if tempDir == "" {
			tempDir = "storage/temp"
		}
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			global.Logger.Error("websocket_router.file.FileUploadCheck MkdirAll err", zap.Error(err))
			c.ToResponse(code.ErrorServerInternal)
			return
		}

		tempPath := filepath.Join(tempDir, fmt.Sprintf("%s_%s", sessionID, params.Path))
		file, err := os.Create(tempPath)
		if err != nil {
			global.Logger.Error("websocket_router.file.FileUploadCheck Create file err", zap.Error(err))
			c.ToResponse(code.ErrorServerInternal)
			return
		}

		session := &BinaryChunkSession{
			ID:          sessionID,
			Vault:       params.Path,
			Path:        params.Path,
			PathHash:    params.PathHash,
			ContentHash: params.ContentHash,
			Size:        params.Size,
			Ctime:       params.Ctime,
			Mtime:       params.Mtime,
			ChunkSize:   1024 * 1024, // Default 1MB, or calculated
			SavePath:    tempPath,
			FileHandle:  file,
		}
		session.ChunkSize = util.Ceil(session.Size, session.ChunkSize)

		c.BinaryMu.Lock()
		c.BinaryChunkSessions[sessionID] = session
		c.BinaryMu.Unlock()

		data := &service.FilePushMessage{
			Path:      fileSvc.Path,
			Ctime:     fileSvc.Ctime,
			Mtime:     fileSvc.Mtime,
			SessionID: session.ID,
			ChunkSize: session.ChunkSize,
		}
		c.ToResponse(code.Success.WithData(data), "FileNeedUpload")
		return

	} else if updateMode == "UpdateMtime" {

		data := &service.FileMtimePushMessage{
			Path:  fileSvc.Path,
			Ctime: fileSvc.Ctime,
			Mtime: fileSvc.Mtime,
		}
		c.ToResponse(code.Success.WithData(data), "FileSyncMtime")
		return
	}
	c.ToResponse(code.SuccessNoUpdate.Reset())
}

func FileUploadChunkBinary(c *app.WebsocketClient, data []byte) {
	if len(data) < 40 {
		global.Logger.Error("UploadBinary Invalid data length")
		return
	}

	sessionID := string(data[:36])
	chunkIndex := binary.BigEndian.Uint32(data[36:40])
	chunkData := data[40:]

	c.BinaryMu.Lock()
	binarySession, exists := c.BinaryChunkSessions[sessionID]
	c.BinaryMu.Unlock()
	session := binarySession.(*BinaryChunkSession)

	if !exists {
		global.Logger.Error("websocket_router.file.FileUploadChunkBinary Session not found", zap.String("sessionID", sessionID))
		return
	}

	// Write to file
	// Note: For simplicity, we assume sequential upload or use Seek.
	// Using Seek is safer for parallel chunks.
	offset := int64(chunkIndex) * int64(session.ChunkSize)

	if _, err := session.FileHandle.Seek(offset, 0); err != nil {
		global.Logger.Error("websocket_router.file.FileUploadChunkBinary FileHandle Seek err", zap.Error(err))
		return
	}

	if _, err := session.FileHandle.Write(chunkData); err != nil {
		global.Logger.Error("websocket_router.file.FileUploadChunkBinary FileHandle Write err", zap.Error(err))
		return
	}
}

// FileUploadComplete handles the completion of a file upload
func FileUploadComplete(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &FileUploadCompleteParams{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("FileChunkUploadComplete BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	c.BinaryMu.Lock()
	binarySession, exists := c.BinaryChunkSessions[params.SessionID]
	delete(c.BinaryChunkSessions, params.SessionID) // Remove session
	c.BinaryMu.Unlock()
	session := binarySession.(*BinaryChunkSession)

	if !exists {
		c.ToResponse(code.ErrorInvalidParams.WithDetails("Session not found"))
		return
	}

	session.FileHandle.Close()

	// Move to final destination
	finalDir := "storage/uploads"
	if err := os.MkdirAll(finalDir, 0755); err != nil {
		global.Logger.Error("FileChunkUploadComplete MkdirAll err", zap.Error(err))
		c.ToResponse(code.ErrorServerInternal)
		return
	}

	finalPath := filepath.Join(finalDir, fmt.Sprintf("%d_%s", time.Now().Unix(), session.Path))

	if err := os.Rename(session.SavePath, finalPath); err != nil {
		// Try copy if rename fails (different volume)
		if err := copyFile(session.SavePath, finalPath); err != nil {
			global.Logger.Error("FileChunkUploadComplete Move file err", zap.Error(err))
			c.ToResponse(code.ErrorServerInternal)
			return
		}
		os.Remove(session.SavePath)
	}

	response := map[string]interface{}{
		"path": finalPath,
		"url":  "/uploads/" + filepath.Base(finalPath), // Mock URL
	}
	c.ToResponse(code.Success.WithData(response), "FileChunkUploadComplete")
}

// NoteModify 处理文件修改的 WebSocket 消息
// 函数名: NoteModify
// 函数使用说明: 处理客户端发送的笔记修改或创建消息，进行参数校验、更新检查并在需要时写回数据库或通知其他客户端。
// 参数说明:
//   - c *app.WebsocketClient: 当前 WebSocket 客户端连接，包含上下文、用户信息、发送响应等能力。
//   - msg *app.WebSocketMessage: 接收到的 WebSocket 消息，包含消息数据和类型。
//
// 返回值说明:
//   - 无
func FileUpload(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.FileUpdateParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.note.NoteModify.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}
	if params.PathHash == "" {
		c.ToResponse(code.ErrorInvalidParams.WithDetails("pathHash is required"))
		return
	}
	if params.ContentHash == "" {
		c.ToResponse(code.ErrorInvalidParams.WithDetails("contentHash is required"))
		return
	}
	if params.Mtime == 0 {
		c.ToResponse(code.ErrorInvalidParams.WithDetails("mtime is required"))
		return
	}
	if params.Ctime == 0 {
		c.ToResponse(code.ErrorInvalidParams.WithDetails("ctime is required"))
		return
	}

	svc := service.New(c.Ctx).WithSF(c.SF)

	// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
	svc.VaultGetOrCreate(params.Vault, c.User.UID)

	checkParams := convert.StructAssign(params, &service.FileUpdateCheckParams{}).(*service.FileUpdateCheckParams)
	isNew, isNeedUpdate, isNeedSyncMtime, _, err := svc.FileUpdateCheck(c.User.UID, checkParams)

	if err != nil {
		c.ToResponse(code.ErrorNoteModifyFailed.WithDetails(err.Error()))
		return
	}

	var file *service.File

	if isNew || isNeedSyncMtime || isNeedUpdate {
		_, file, err = svc.FileModifyOrCreate(c.User.UID, params, true)
		if err != nil {
			c.ToResponse(code.ErrorNoteModifyFailed.WithDetails(err.Error()))
			return
		}
		// 通知所有客户端更新mtime
		if isNeedSyncMtime {
			c.ToResponse(code.Success.Reset())
			c.BroadcastResponse(code.Success.Reset().WithData(file), false, "NoteSyncModify")
			return
		} else if isNeedUpdate || isNew {

			c.ToResponse(code.Success.Reset())
			c.BroadcastResponse(code.Success.Reset().WithData(file), true, "NoteSyncModify")
			return
		}
	} else {
		c.ToResponse(code.SuccessNoUpdate.Reset())
		return
	}

}

// NoteDelete 处理文件删除的 WebSocket 消息
// 函数名: NoteDelete
// 函数使用说明: 接收客户端的笔记删除请求，执行删除操作并通知其他客户端同步删除事件。
// 参数说明:
//   - c *app.WebsocketClient: 当前 WebSocket 客户端连接，包含发送响应与广播能力。
//   - msg *app.WebSocketMessage: 接收到的删除请求消息，包含要删除的笔记标识等参数。
//
// 返回值说明:
//   - 无
func FileDelete(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.NoteDeleteRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.note.NoteDelete.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx).WithSF(c.SF)

	// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
	svc.VaultGetOrCreate(params.Vault, c.User.UID)

	note, err := svc.NoteDelete(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorNoteDeleteFailed.WithDetails(err.Error()))
		return
	}

	c.ToResponse(code.Success)
	c.BroadcastResponse(code.Success.WithData(note), true, "NoteSyncDelete")
}

// NoteSync 处理全量或增量笔记同步
// 函数名: NoteSync
// 函数使用说明: 根据客户端提供的本地笔记列表与服务器端最近更新列表比较，决定返回哪些笔记需要上传、需要同步 mtime、需要删除或需要更新；最后返回同步结束消息。
// 参数说明:
//   - c *app.WebsocketClient: 当前 WebSocket 客户端连接，包含上下文与响应发送能力。
//   - msg *app.WebSocketMessage: 接收到的同步请求，包含客户端的笔记摘要和同步起始时间等信息。
//
// 返回值说明:
//   - 无
func FileSync(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.NoteSyncRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.note.NoteModify.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx).WithSF(c.SF)

	// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
	svc.VaultGetOrCreate(params.Vault, c.User.UID)

	list, err := svc.NoteListByLastTime(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorNoteModifyFailed.WithDetails(err.Error()))
		return
	}

	var cNotes map[string]service.NoteSyncCheckRequestParams = make(map[string]service.NoteSyncCheckRequestParams, 0)
	var cNotesKeys map[string]struct{} = make(map[string]struct{}, 0)

	if len(params.Notes) > 0 {
		for _, note := range params.Notes {
			cNotes[note.PathHash] = note
			cNotesKeys[note.PathHash] = struct{}{}
		}
	}

	var lastTime int64

	for _, note := range list {
		if note.UpdatedTimestamp >= lastTime {
			lastTime = note.UpdatedTimestamp
		}

		if note.Action == "delete" {
			c.ToResponse(code.Success.WithData(note), "NoteSyncDelete")
		} else {
			if cNote, ok := cNotes[note.PathHash]; ok {

				delete(cNotesKeys, note.PathHash)

				if note.ContentHash == cNote.ContentHash && note.Mtime == cNote.Mtime {
					continue
				} else if note.ContentHash != cNote.ContentHash {
					NoteCheck := convert.StructAssign(note, &service.NoteSyncNeedPushMessage{}).(*service.NoteSyncNeedPushMessage)
					c.ToResponse(code.Success.WithData(NoteCheck), "NoteSyncNeedPush")
				} else {
					NoteCheck := convert.StructAssign(note, &service.NoteSyncMtimeMessage{}).(*service.NoteSyncMtimeMessage)
					c.ToResponse(code.Success.WithData(NoteCheck), "NoteSyncMtime")
				}
			} else {
				c.ToResponse(code.Success.WithData(note), "NoteSyncModify")
			}
		}
	}

	if list == nil {
		lastTime = timex.Now().UnixMilli()
	}
	if len(cNotesKeys) > 0 {
		for pathHash := range cNotesKeys {
			note := cNotes[pathHash]
			NoteCheck := convert.StructAssign(&note, &service.NoteSyncNeedPushMessage{}).(*service.NoteSyncNeedPushMessage)
			c.ToResponse(code.Success.WithData(NoteCheck), "NoteSyncNeedPush")
		}
	}

	message := &service.NoteSyncEndMessage{
		Vault:    params.Vault,
		LastTime: lastTime,
	}
	c.ToResponse(code.Success.WithData(message), "NoteSyncEnd")
}
