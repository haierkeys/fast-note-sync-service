package websocket_router

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"github.com/haierkeys/fast-note-sync-service/pkg/fileurl"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"

	"go.uber.org/zap"
)

// BinaryChunkSession stores the state of an active upload
type FileUploadBinaryChunkSession struct {
	ID             string             // 上传会话ID
	Vault          string             // 仓库标识
	Path           string             // 路径
	PathHash       string             // 路径哈希
	ContentHash    string             // 内容哈希(可选)
	Ctime          int64              // 创建时间戳
	Mtime          int64              // 修改时间戳
	Size           int64              // 文件总大小
	TotalChunks    int64              // 总分块数
	UploadedChunks int64              // 已上传分块数
	ChunkSize      int64              // 每个分块大小
	SavePath       string             // 临时保存路径
	FileHandle     *os.File           // 文件句柄
	mu             sync.Mutex         // 互斥锁
	CreatedAt      time.Time          // 会话创建时间
	CancelFunc     context.CancelFunc // 用于取消超时定时器
}

type FileUploadCompleteParams struct {
	SessionID string `json:"sessionID" binding:"required"`
}

// cleanupSession 清理单个上传会话
// 函数名: cleanupSession
// 函数使用说明: 关闭文件句柄、删除临时文件、从会话map中移除
// 参数说明:
//   - c *app.WebsocketClient: WebSocket 客户端连接
//   - sessionID string: 会话ID
func cleanupSession(c *app.WebsocketClient, sessionID string) {
	c.BinaryMu.Lock()
	binarySession, exists := c.BinaryChunkSessions[sessionID]
	if exists {
		delete(c.BinaryChunkSessions, sessionID)
	}
	c.BinaryMu.Unlock()

	if !exists {
		return
	}

	session := binarySession.(*FileUploadBinaryChunkSession)

	// 取消超时定时器
	if session.CancelFunc != nil {
		session.CancelFunc()
	}

	// 关闭文件句柄
	if session.FileHandle != nil {
		if err := session.FileHandle.Close(); err != nil {
			global.Logger.Warn("cleanupSession: failed to close file handle",
				zap.String("sessionID", sessionID),
				zap.Error(err))
		}
	}

	// 删除临时文件
	if session.SavePath != "" {
		if err := os.Remove(session.SavePath); err != nil && !os.IsNotExist(err) {
			global.Logger.Warn("cleanupSession: failed to remove temp file",
				zap.String("sessionID", sessionID),
				zap.String("path", session.SavePath),
				zap.Error(err))
		}
	}

	global.Logger.Info("cleanupSession: session cleaned up",
		zap.String("sessionID", sessionID),
		zap.String("path", session.Path))
}

// startSessionTimeout 启动会话超时定时器
// 函数名: startSessionTimeout
// 函数使用说明: 创建一个定时器,在超时后自动清理会话
// 参数说明:
//   - c *app.WebsocketClient: WebSocket 客户端连接
//   - sessionID string: 会话ID
//   - timeout time.Duration: 超时时长
//
// 返回值说明:
//   - context.CancelFunc: 用于取消超时的函数
func startSessionTimeout(c *app.WebsocketClient, sessionID string, timeout time.Duration) context.CancelFunc {
	if timeout <= 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		select {
		case <-timer.C:
			// 超时,清理会话
			global.Logger.Warn("startSessionTimeout: session timeout, cleaning up",
				zap.String("sessionID", sessionID),
				zap.Duration("timeout", timeout))
			cleanupSession(c, sessionID)
		case <-ctx.Done():
			// 被取消,正常完成
			return
		}
	}()

	return cancel
}

func FileUploadCheck(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.FileUpdateCheckParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("websocket_router.file.FileUploadCheck.BindAndValid errs: %v", zap.Error(errs))
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

		// 获取原文件扩展名
		fileExt := filepath.Ext(params.Path)
		// 使用 sessionID 作为随机文件名,保持原文件扩展名
		tempPath := filepath.Join(tempDir, fmt.Sprintf("%s%s", sessionID, fileExt))
		file, err := os.Create(tempPath)
		if err != nil {
			global.Logger.Error("websocket_router.file.FileUploadCheck Create file err", zap.Error(err))
			c.ToResponse(code.ErrorServerInternal)
			return
		}

		session := &FileUploadBinaryChunkSession{
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
			CreatedAt:   time.Now(),
		}
		session.ChunkSize = util.Ceil(session.Size, session.ChunkSize)

		// 解析超时配置
		var timeout time.Duration
		if global.Config.App.UploadSessionTimeout != "" && global.Config.App.UploadSessionTimeout != "0" {
			var err error
			timeout, err = time.ParseDuration(global.Config.App.UploadSessionTimeout)
			if err != nil {
				global.Logger.Warn("FileUploadCheck: invalid upload-session-timeout config, using default 5m",
					zap.String("config", global.Config.App.UploadSessionTimeout),
					zap.Error(err))
				timeout = 5 * time.Minute
			}
		} else {
			timeout = 5 * time.Minute // 默认5分钟
		}

		// 启动超时定时器
		session.CancelFunc = startSessionTimeout(c, sessionID, timeout)

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

	if !exists {
		global.Logger.Error("websocket_router.file.FileUploadChunkBinary Session not found", zap.String("sessionID", sessionID))
		return
	}

	session := binarySession.(*FileUploadBinaryChunkSession)

	offset := int64(chunkIndex) * int64(session.ChunkSize)

	if _, err := session.FileHandle.WriteAt(chunkData, offset); err != nil {
		global.Logger.Error("websocket_router.file.FileUploadChunkBinary FileHandle WriteAt err", zap.Error(err))
		return
	}

	// 更新已上传分块数
	session.mu.Lock()
	session.UploadedChunks++
	session.mu.Unlock()
}

// FileUploadComplete handles the completion of a file upload
func FileUploadComplete(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &FileUploadCompleteParams{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("websocket_router.file.FileUploadComplete BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	//获取用户上传数据
	c.BinaryMu.Lock()
	binarySession, exists := c.BinaryChunkSessions[params.SessionID]
	if exists {
		delete(c.BinaryChunkSessions, params.SessionID) // Remove session
	}
	c.BinaryMu.Unlock()

	if !exists {
		c.ToResponse(code.ErrorInvalidParams.WithDetails("Session not found"))
		return
	}

	session := binarySession.(*FileUploadBinaryChunkSession)

	// 取消超时定时器
	if session.CancelFunc != nil {
		session.CancelFunc()
	}

	// 关闭文件句柄
	if err := session.FileHandle.Close(); err != nil {
		global.Logger.Error("FileUploadComplete: failed to close file handle", zap.Error(err))
		c.ToResponse(code.ErrorServerInternal)
		return
	}

	baseUploadDir := global.Config.App.UploadSavePath
	if baseUploadDir == "" {
		baseUploadDir = "storage/uploads"
	}

	// 生成日期子目录 (格式: YYMMDD, 如 240520)
	dateDir := time.Now().Format("200601")
	finalDir := filepath.Join(baseUploadDir, dateDir)

	// 创建目标目录
	if err := os.MkdirAll(finalDir, 0755); err != nil {
		global.Logger.Error("websocket_router.file.FileUploadComplete MkdirAll err", zap.Error(err))
		c.ToResponse(code.ErrorServerInternal)
		return
	}

	// 生成唯一文件名 (循环检查直到文件名不冲突)
	originalName := filepath.Base(session.Path)
	ext := filepath.Ext(originalName)
	nameOnly := strings.TrimSuffix(originalName, ext)

	var finalPath string

	// 最多重试 10 次，防止极端情况死循环
	for i := 0; i < 10; i++ {
		// 使用 util 包生成随机字符串 (例如长度为 8)
		randStr := util.GetRandomString(8)

		// 拼接文件名: 原文件名_随机串.后缀
		tryFileName := fmt.Sprintf("%s_%s%s", nameOnly, randStr, ext)
		tryPath := filepath.Join(finalDir, tryFileName)

		// 判断文件是否存在
		// 如果返回 IsNotExist，说明文件名可用，跳出循环
		if _, err := os.Stat(tryPath); os.IsNotExist(err) {
			finalPath = tryPath
			break
		}
	}

	// 如果重试 10 次后 finalPath 依然为空，报错
	if finalPath == "" {
		global.Logger.Error("websocket_router.file.FileUploadComplete Generate unique filename failed after retries")
		c.ToResponse(code.ErrorServerInternal)
		return
	}

	// 移动文件 (优先 Rename，失败则 Copy+Remove)
	if err := os.Rename(session.SavePath, finalPath); err != nil {
		// Rename 失败（通常因跨磁盘分区），使用流式复制
		if err := fileurl.CopyFile(session.SavePath, finalPath); err != nil {
			global.Logger.Error("websocket_router.file.FileUploadComplete Move/Copy file err", zap.Error(err))
			c.ToResponse(code.ErrorServerInternal)
			return
		}
		// 复制成功后删除原临时文件
		os.Remove(session.SavePath)
	}

	svc := service.New(c.Ctx).WithSF(c.SF)

	svcParams := &service.FileUpdateParams{
		Vault:       session.Vault,
		Path:        session.Path,
		PathHash:    session.PathHash,
		ContentHash: session.ContentHash,
		SavePath:    finalPath,
		Size:        session.Size,
		Ctime:       session.Ctime,
		Mtime:       session.Mtime,
	}

	_, fileSvc, err := svc.FileUpdateOrCreate(c.User.UID, svcParams, true)

	if err != nil {
		c.ToResponse(code.ErrorServerInternal.WithDetails(err.Error()))
		return
	}

	c.ToResponse(code.Success.Reset())
	c.BroadcastResponse(code.Success.Reset().WithData(fileSvc), true, "FileSyncUpload")
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
	params := &service.FileDeleteParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.file.FileDelete.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx).WithSF(c.SF)

	// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
	svc.VaultGetOrCreate(params.Vault, c.User.UID)

	fileSvc, err := svc.FileDelete(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorNoteDeleteFailed.WithDetails(err.Error()))
		return
	}

	c.ToResponse(code.Success.Reset())
	c.BroadcastResponse(code.Success.Reset().WithData(fileSvc), true, "FileSyncDelete")
}

// FileSync 处理全量或增量笔记同步
// 函数名: FileSync
// 函数使用说明: 根据客户端提供的本地文件列表与服务器端最近更新列表比较，决定返回哪些文件需要上传、需要同步 mtime、需要删除或需要更新；最后返回同步结束消息。
// 参数说明:
//   - c *app.WebsocketClient: 当前 WebSocket 客户端连接，包含上下文与响应发送能力。
//   - msg *app.WebSocketMessage: 接收到的同步请求，包含客户端的文件摘要和同步起始时间等信息。
//
// 返回值说明:
//   - 无
func FileSync(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.FileSyncRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.file.FileSync.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx).WithSF(c.SF)

	// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
	svc.VaultGetOrCreate(params.Vault, c.User.UID)

	list, err := svc.FileListByLastTime(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorNoteModifyFailed.WithDetails(err.Error()))
		return
	}

	var cFiles map[string]service.FileSyncCheckRequestParams = make(map[string]service.FileSyncCheckRequestParams, 0)
	var cFilesKeys map[string]struct{} = make(map[string]struct{}, 0)

	if len(params.Files) > 0 {
		for _, file := range params.Files {
			cFiles[file.PathHash] = file
			cFilesKeys[file.PathHash] = struct{}{}
		}
	}

	var lastTime int64

	for _, file := range list {
		if file.UpdatedTimestamp >= lastTime {
			lastTime = file.UpdatedTimestamp
		}

		if file.Action == "delete" {
			c.ToResponse(code.Success.WithData(file), "FileSyncDelete")
		} else {
			if cFile, ok := cFiles[file.PathHash]; ok {

				delete(cFilesKeys, file.PathHash)

				if file.ContentHash == cFile.ContentHash && file.Mtime == cFile.Mtime {
					continue
				} else if file.ContentHash != cFile.ContentHash {
					FileCheck := convert.StructAssign(file, &service.FileMtimePushMessage{}).(*service.FileMtimePushMessage)
					c.ToResponse(code.Success.WithData(FileCheck), "FileSyncNeedPush")
				} else {
					FileCheck := convert.StructAssign(file, &service.FileMtimePushMessage{}).(*service.FileMtimePushMessage)
					c.ToResponse(code.Success.WithData(FileCheck), "FileSyncMtime")
				}
			} else {
				c.ToResponse(code.Success.WithData(file), "FileSyncModify")
			}
		}
	}

	if list == nil {
		lastTime = timex.Now().UnixMilli()
	}
	if len(cFilesKeys) > 0 {
		for pathHash := range cFilesKeys {
			file := cFiles[pathHash]
			FileCheck := convert.StructAssign(&file, &service.FileMtimePushMessage{}).(*service.FileMtimePushMessage)
			c.ToResponse(code.Success.WithData(FileCheck), "FileSyncNeedPush")
		}
	}

	message := &service.FileSyncEndMessage{
		Vault:    params.Vault,
		LastTime: lastTime,
	}
	c.ToResponse(code.Success.WithData(message), "FileSyncEnd")
}
