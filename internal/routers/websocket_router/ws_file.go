package websocket_router

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/logger"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"

	"go.uber.org/zap"
)

// FileWSHandler WebSocket file handler
// FileWSHandler WebSocket 文件处理器
// Uses App Container to inject dependencies
// 使用 App Container 注入依赖
type FileWSHandler struct {
	*WSHandler
}

// NewFileWSHandler creates FileWSHandler instance
// NewFileWSHandler 创建 FileWSHandler 实例
func NewFileWSHandler(a *app.App) *FileWSHandler {
	return &FileWSHandler{
		WSHandler: NewWSHandler(a),
	}
}

type FileUploadBinaryChunkSession struct {
	ID             string             // Session ID // 会话 ID
	Vault          string             // Vault Name // 仓库名称
	Path           string             // File Path // 文件路径
	PathHash       string             // File Path Hash // 文件路径哈希值
	ContentHash    string             // File Content Hash // 文件内容哈希值
	Ctime          int64              // Creation time // 创建时间
	Mtime          int64              // Modification time // 修改时间
	Size           int64              // File size // 文件大小
	TotalChunks    int64              // Total chunks // 总分块数
	UploadedChunks int64              // Uploaded chunks // 已上传分块数
	UploadedBytes  int64              // Uploaded bytes // 已上传字节数
	ChunkSize      int64              // Chunk size // 分块大小
	SavePath       string             // Temp save path // 临时保存路径
	FileHandle     *os.File           // File handle // 文件句柄
	mu             sync.Mutex         // Mutex to protect concurrent operations // 互斥锁，保护并发操作
	CreatedAt      time.Time          // Created time // 创建时间
	CancelFunc     context.CancelFunc // Cancel function for timeout control // 取消函数，用于超时控制
}

func (s *FileUploadBinaryChunkSession) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Cancel timeout timer
	// 取消超时定时器
	if s.CancelFunc != nil {
		s.CancelFunc()
		s.CancelFunc = nil
	}

	if s.FileHandle != nil {
		if err := s.FileHandle.Close(); err != nil {
			zap.L().Warn("cleanup: failed to close file handle",
				zap.String(logger.FieldSessionID, s.ID),
				zap.String(logger.FieldPath, s.Path),
				zap.String(logger.FieldMethod, "FileUploadBinaryChunkSession.Cleanup"),
				zap.Error(err),
			)
		}
		s.FileHandle = nil
	}
	if s.SavePath != "" {
		if err := os.Remove(s.SavePath); err != nil {
			zap.L().Warn("cleanup: failed to remove temp file",
				zap.String(logger.FieldSessionID, s.ID),
				zap.String(logger.FieldPath, s.SavePath),
				zap.String(logger.FieldMethod, "FileUploadBinaryChunkSession.Cleanup"),
				zap.Error(err),
			)
		}
		s.SavePath = ""
	}
}

// FileDownloadChunkSession defines the session state for file chunk download
// Used to track progress and file info for large file chunk downloads
// FileDownloadChunkSession 定义文件分块下载的会话状态。
// 用于跟踪大文件分块下载的进度和文件信息。
type FileDownloadChunkSession struct {
	SessionID   string // Session ID // 会话 ID
	Path        string // File path (for logging) // 文件路径(用于日志)
	Size        int64  // File size // 文件大小
	TotalChunks int64  // Total chunks // 总分块数
	ChunkSize   int64  // Chunk size // 分块大小
	SavePath    string // File actual save path // 文件实际保存路径
}

// FileUploadCheck checks file upload request, initializes upload session or confirms no upload needed
// FileUploadCheck 检查文件上传请求，初始化上传会话或确认无需上传。
func (h *FileWSHandler) FileUploadCheck(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.FileUpdateCheckRequest{}

	// Bind and validate parameters
	// 绑定并验证参数
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.file.FileUploadCheck.BindAndValid")
		return
	}

	// Required parameter validation
	// 必填参数校验
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

	ctx := c.Context()

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "FileUploadCheck", params.Path, params.Vault)

	// Check and create vault, internally uses SF to merge concurrent requests, avoiding duplicate creation issues
	// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	// 检查文件更新状态
	updateMode, fileSvc, err := h.App.FileService.UploadCheck(ctx, c.User.UID, params)

	if err != nil {
		h.respondError(c, code.ErrorFileUploadCheckFailed, err, "websocket_router.file.FileUploadCheck.UploadCheck")
		return
	}

	// UpdateContent 或 Create 模式，需要客户端上传文件
	switch updateMode {
	case "UpdateContent", "Create":

		session, err := h.handleFileUploadSessionCreate(c, params.Vault, params.Path, params.PathHash, params.ContentHash, params.Size, params.Ctime, params.Mtime)
		if err != nil {
			h.respondError(c, code.ErrorFileUploadCheckFailed, err, "websocket_router.file.FileUploadCheck.handleFileUploadSessionCreate")
			return
		}

		c.ToResponse(code.Success.WithData(
			dto.FileSyncUploadMessage{
				Path:      session.Path,
				SessionID: session.ID,
				ChunkSize: session.ChunkSize,
			},
		).WithVault(session.Vault), dto.FileUpload)
		return

	case "UpdateMtime":
		// 当用户 mtime 小于服务端 mtime 时，通知用户更新mtime
		c.ToResponse(code.Success.WithData(
			dto.FileSyncMtimeMessage{
				Path:  fileSvc.Path,
				Ctime: fileSvc.Ctime,
				Mtime: fileSvc.Mtime,
			},
		).WithVault(params.Vault), dto.FileSyncMtime)
		return
	default:
		// 无需更新
		c.ToResponse(code.SuccessNoUpdate)
	}
}

// FileUploadChunkBinary handles binary data for file chunk upload.
// FileUploadChunkBinary 处理文件分块上传的二进制数据。
func (h *FileWSHandler) FileUploadChunkBinary(c *pkgapp.WebsocketClient, data []byte) {
	// Check if context is cancelled (connection might be closed)
	// 检查 context 是否已取消（连接可能已关闭）
	select {
	case <-c.Context().Done():
		h.logInfo(c, "FileUploadChunkBinary: context cancelled, skipping chunk processing")
		return
	default:
	}

	if len(data) < 40 {
		h.logError(c, "websocket_router.file.FileUploadChunkBinary", fmt.Errorf("invalid data length: %d", len(data)))
		return
	}

	// Parse session ID and chunk index
	// 解析会话 ID 和分块索引
	sessionID := string(data[:36])
	chunkIndex := binary.BigEndian.Uint32(data[36:40])
	chunkData := data[40:]

	// Get session from global server (supports cross-connection)
	// 从全局服务器获取会话 (支持跨连接)
	binarySession := c.Server.GetSession(c.User.ID, sessionID)

	if binarySession == nil {
		h.logError(c, "websocket_router.file.FileUploadChunkBinary", fmt.Errorf("session not found: %s", sessionID))
		c.ToResponse(code.ErrorFileUploadSessionNotFound.WithData(map[string]string{
			"sessionID": sessionID,
		}))
		return
	}

	session := binarySession.(*FileUploadBinaryChunkSession)
	if session == nil {
		h.logError(c, "websocket_router.file.FileUploadChunkBinary", fmt.Errorf("session is nil: %s", sessionID))
		c.ToResponse(code.ErrorFileUploadSessionNotFound.WithData(map[string]string{
			"sessionID": sessionID,
		}))
		return
	}

	// Calculate write offset and write data
	// 计算写入偏移量并写入数据
	offset := int64(chunkIndex) * int64(session.ChunkSize)

	if _, err := session.FileHandle.WriteAt(chunkData, offset); err != nil {
		// Cleanup session on fatal error
		// 致命错误时清理会话
		h.handleFileUploadSessionCleanup(c, sessionID)
		h.respondErrorWithData(c, code.ErrorFileUploadFailed, err, map[string]string{"sessionID": sessionID}, "websocket_router.file.FileUploadChunkBinary.WriteAt")
		return
	}

	// Update uploaded count and bytes
	// 更新已上传计数和字节数
	session.mu.Lock()
	session.UploadedChunks++
	session.UploadedBytes += int64(len(chunkData))
	uploadedBytes := session.UploadedBytes
	session.mu.Unlock()

	// Check if all data has been uploaded (judged by bytes)
	// 检查是否所有数据都已上传(根据字节数判断)
	if uploadedBytes >= session.Size {

		h.logInfo(c, "FileUploadComplete: upload finished",
			zap.String("sessionID", sessionID),
			zap.String("path", session.Path),
			zap.Int64("uploadedBytes", uploadedBytes),
			zap.Int64("totalSize", session.Size),
			zap.Int64("uploadedChunks", session.UploadedChunks),
			zap.Int64("totalChunks", session.TotalChunks))

		// Get and remove session from global server
		// 获取并从全局服务器移除会话
		c.Server.RemoveSession(c.User.ID, session.ID)

		// Cancel timeout timer
		// 取消超时定时器
		if session.CancelFunc != nil {
			session.CancelFunc()
		}

		// Close temp file handle
		// 关闭临时文件句柄
		if err := session.FileHandle.Close(); err != nil {
			h.respondErrorWithData(c, code.ErrorFileUploadFailed, err, map[string]string{"sessionID": sessionID}, "websocket_router.file.FileUploadChunkBinary.Close")
			return
		}
		session.FileHandle = nil // Avoid double closing during cleanup // 避免再次清理时重复关闭

		ctx := c.Context()

		// Check and create vault, internally uses SF to merge concurrent requests, avoiding duplicate creation issues
		// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
		h.App.VaultService.GetOrCreate(ctx, c.User.UID, session.Vault)

		svcParams := &dto.FileUpdateRequest{
			Vault:       session.Vault,
			Path:        session.Path,
			PathHash:    session.PathHash,
			ContentHash: session.ContentHash,
			SavePath:    session.SavePath, // 将临时文件的完整路径作为 SavePath 传入
			Size:        session.Size,
			Ctime:       session.Ctime,
			Mtime:       session.Mtime,
		}

		// Update or create file record (DAO layer will automatically move temp file from SavePath to f_{id} folder)
		// 更新或创建文件记录 (DAO 层会自动将 SavePath 里的临时文件移动到 f_{id} 文件夹)
		_, fileSvc, err := h.App.FileService.UploadComplete(ctx, c.User.UID, svcParams)

		if err != nil {
			h.respondError(c, code.ErrorFileModifyOrCreateFailed, err, "websocket_router.file.FileUploadChunkBinary.UploadComplete")
			return
		}

		if fileSvc == nil {
			h.logInfo(c, "FileUploadChunkBinary: fileSvc is nil, skipping broadcast", zap.String("path", session.Path))
			c.ToResponse(code.Success)
			return
		}

		c.ToResponse(code.Success)
		// Broadcast file update message
		// 广播文件更新消息
		c.BroadcastResponse(code.Success.WithData(
			dto.FileSyncModifyMessage{
				Path:             fileSvc.Path,
				PathHash:         fileSvc.PathHash,
				ContentHash:      fileSvc.ContentHash,
				Size:             fileSvc.Size,
				Ctime:            fileSvc.Ctime,
				Mtime:            fileSvc.Mtime,
				UpdatedTimestamp: fileSvc.UpdatedTimestamp,
			},
		).WithVault(session.Vault), true, dto.FileSyncUpdate)
	}
}

// FileDelete handles file deletion request.
// FileDelete 处理文件删除请求。
func (h *FileWSHandler) FileDelete(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.FileDeleteRequest{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.file.FileDelete.BindAndValid")
		return
	}

	ctx := c.Context()

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "FileDelete", params.Path, params.Vault)

	// 获取或创建仓库
	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	// Execute deletion logic
	// 执行删除逻辑
	fileSvc, err := h.App.FileService.Delete(ctx, c.User.UID, params)

	if err != nil {
		h.respondError(c, code.ErrorFileDeleteFailed, err, "websocket_router.file.FileDelete.Delete")
		return
	}

	c.ToResponse(code.Success)

	// Broadcast file deletion message
	// 广播文件删除消息
	c.BroadcastResponse(code.Success.WithData(
		dto.FileSyncDeleteMessage{
			Path:     fileSvc.Path,
			PathHash: fileSvc.PathHash,
			Ctime:    fileSvc.Ctime,
			Mtime:    fileSvc.Mtime,
			Size:     fileSvc.Size,
		},
	).WithVault(params.Vault), true, dto.FileSyncDelete)
}

// FileRename handles file rename request.
// FileRename 处理文件重命名请求。
func (h *FileWSHandler) FileRename(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.FileRenameRequest{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.file.FileRename.BindAndValid")
		return
	}

	uid := c.User.UID
	oldFile, newFile, err := h.App.FileService.WithClient(c.ClientName, c.ClientVersion).Rename(c.Context(), uid, params)
	if err != nil {
		h.respondError(c, code.ErrorFileRenameFailed, err, "websocket_router.file.FileRename.Rename")
		return
	}

	c.ToResponse(code.Success)

	c.BroadcastResponse(code.Success.WithData(
		dto.FileSyncRenameMessage{
			Path:        newFile.Path,
			PathHash:    newFile.PathHash,
			ContentHash: newFile.ContentHash,
			Ctime:       newFile.Ctime,
			Mtime:       newFile.Mtime,
			Size:        newFile.Size,
			OldPath:     oldFile.Path,
			OldPathHash: oldFile.PathHash,
		},
	).WithVault(params.Vault), true, dto.FileSyncRename)
}

// FileChunkDownload handles file chunk download request.
// Client requests file download via this interface, server creates download session and starts sending chunks.
// FileChunkDownload 处理文件分片下载请求。
// 客户端通过此接口请求下载文件,服务端创建下载会话并开始发送分片。
func (h *FileWSHandler) FileChunkDownload(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.FileGetRequest{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.file.FileChunkDownload.BindAndValid")
		return
	}

	ctx := c.Context()

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "FileChunkDownload", params.Path, params.Vault)

	// 获取或创建仓库
	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	// Get file info
	// 获取文件信息
	fileSvc, err := h.App.FileService.Get(ctx, c.User.UID, params)

	if err != nil {
		h.respondError(c, code.ErrorFileGetFailed, err, "websocket_router.file.FileChunkDownload.Get")
		return
	}

	// Check if file exists on disk
	// 检查文件是否存在于磁盘
	if _, err := os.Stat(fileSvc.SavePath); os.IsNotExist(err) {
		h.respondError(c, code.ErrorFileGetFailed, fmt.Errorf("file not found on disk: %s (pathHash: %s)", fileSvc.SavePath, fileSvc.PathHash), "websocket_router.file.FileChunkDownload.Stat")
		return
	}

	// Create download session
	// 创建下载会话
	sessionID := uuid.New().String()
	chunkSize := getChunkSizeFromConfig(h.App.Config()) // 从注入的配置获取

	// Calculate total chunks
	// 计算总分块数
	totalChunks := util.Ceil(fileSvc.Size, chunkSize)

	// Initialize download session
	// 初始化下载会话
	session := &FileDownloadChunkSession{
		SessionID:   sessionID,
		Path:        fileSvc.Path,
		Size:        fileSvc.Size,
		TotalChunks: totalChunks,
		ChunkSize:   chunkSize,
		SavePath:    fileSvc.SavePath,
	}

	// Send download ready message
	// 发送下载准备消息
	c.ToResponse(code.Success.WithData(
		dto.FileSyncDownloadMessage{
			Path:        fileSvc.Path,
			Ctime:       fileSvc.Ctime,
			Mtime:       fileSvc.Mtime,
			SessionID:   session.SessionID,
			ChunkSize:   session.ChunkSize,
			TotalChunks: session.TotalChunks,
			Size:        session.Size,
		},
	).WithVault(params.Vault), dto.FileSyncChunkDownload)

	// Start chunk sending, pass timeout and logger
	// 启动分片发送,传入超时时间和 logger
	go h.handleFileChunkDownloadSendChunks(c, session)
}

// FileSync batch checks if user files need update.
// Compares file list between client and server, deciding which files need upload, update or delete.
// FileSync 批量检测用户文件是否需要更新。
// 对比客户端和服务端的文件列表，决定哪些文件需要上传、更新或删除。
func (h *FileWSHandler) FileSync(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.FileSyncRequest{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.file.FileSync.BindAndValid")
		return
	}

	ctx := c.Context()

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "FileSync", "", params.Vault)

	// 获取或创建仓库
	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	// Get list of changed files after last sync
	// 获取最后一次同步后的变更文件列表
	list, err := h.App.FileService.ListByLastTime(ctx, c.User.UID, params)

	if err != nil {
		h.respondError(c, code.ErrorFileListFailed, err, "websocket_router.file.FileSync.ListByLastTime")
		return
	}

	// Build client file index
	// 构建客户端文件索引
	var cFiles map[string]dto.FileSyncCheckRequest = make(map[string]dto.FileSyncCheckRequest, 0)
	var cFilesKeys map[string]struct{} = make(map[string]struct{}, 0)

	if len(params.Files) > 0 {
		for _, file := range params.Files {
			cFiles[file.PathHash] = file
			cFilesKeys[file.PathHash] = struct{}{}
		}
	}

	// Create message queue for collecting all messages to be sent
	// 创建消息队列，用于收集所有待发送的消息
	var messageQueue []dto.WSQueuedMessage

	var lastTime int64
	var needUploadCount int64
	var needModifyCount int64
	var needSyncMtimeCount int64
	var needDeleteCount int64

	var cDelFilesKeys map[string]struct{} = make(map[string]struct{}, 0)

	// Handle files deleted by client
	// 处理客户端删除的文件
	if len(params.DelFiles) > 0 {
		for _, delFile := range params.DelFiles {
			// Check if file exists before deleting
			// 删除前检查文件是否存在
			getParams := &dto.FileGetRequest{
				Vault:    params.Vault,
				PathHash: delFile.PathHash,
			}
			checkFile, err := h.App.FileService.Get(ctx, c.User.UID, getParams)

			// If file exists, execute delete
			// 如果文件存在，执行删除
			if err == nil && checkFile != nil && checkFile.Action != "delete" {
				delParams := &dto.FileDeleteRequest{
					Vault:    params.Vault,
					Path:     delFile.Path,
					PathHash: delFile.PathHash,
				}
				fileSvc, err := h.App.FileService.Delete(ctx, c.User.UID, delParams)
				if err != nil {
					h.App.Logger().Error("websocket_router.file.FileSync.FileService.Delete",
						zap.String(logger.FieldTraceID, c.TraceID),
						zap.Int64(logger.FieldUID, c.User.UID),
						zap.String(logger.FieldPath, delFile.Path),
						zap.Error(err))
					continue
				}

				// Record PathHash deleted by client to avoid duplicate sending
				// 记录客户端已主动删除的 PathHash，避免重复下发
				cDelFilesKeys[delFile.PathHash] = struct{}{}

				// Broadcast deletion to other clients
				// 将删除消息广播给其他客户端
				c.BroadcastResponse(code.Success.WithData(
					dto.FileSyncDeleteMessage{
						Path:     fileSvc.Path,
						PathHash: fileSvc.PathHash,
						Ctime:    fileSvc.Ctime,
						Mtime:    fileSvc.Mtime,
						Size:     fileSvc.Size,
					},
				).WithVault(params.Vault), true, dto.FileSyncDelete)

			} else {
				// File does not exist, but we still need to record exclusion and broadcast delete message to ensure data consistency
				// 文件不存在，但仍需记录排除并广播删除消息，以确保数据一致性

				h.App.Logger().Debug("websocket_router.file.FileSync.FileService.Get check failed (not found or already deleted), broadcasting delete anyway",
					zap.String(logger.FieldTraceID, c.TraceID),
					zap.String("pathHash", delFile.PathHash))

				// Record PathHash
				// 记录 PathHash
				cDelFilesKeys[delFile.PathHash] = struct{}{}

				// Broadcast deletion with available info (Path/PathHash)
				// 使用现有信息(Path/PathHash)广播删除
				c.BroadcastResponse(code.Success.WithData(
					dto.FileSyncDeleteMessage{
						Path:     delFile.Path,
						PathHash: delFile.PathHash,
						Ctime:    0,
						Mtime:    0,
						Size:     0,
					},
				).WithVault(params.Vault), true, dto.FileSyncDelete)
			}
		}
	}

	// Handle files missing on client (only for incremental sync)
	// 处理客户端缺失的文件（仅限增量同步）
	if params.LastTime > 0 && len(params.MissingFiles) > 0 {
		for _, missingFile := range params.MissingFiles {
			getParams := &dto.FileGetRequest{
				Vault:    params.Vault,
				PathHash: missingFile.PathHash,
			}
			fileSvc, err := h.App.FileService.Get(ctx, c.User.UID, getParams)
			if err != nil {
				h.App.Logger().Warn("websocket_router.file.FileSync.FileService.Get",
					zap.String(logger.FieldTraceID, c.TraceID),
					zap.String("pathHash", missingFile.PathHash),
					zap.Error(err))
				continue
			}

			if fileSvc != nil && fileSvc.Action != "delete" {

				messageQueue = append(messageQueue, dto.WSQueuedMessage{
					Action: dto.FileSyncUpdate,
					Data: dto.FileSyncModifyMessage{
						Path:             fileSvc.Path,
						PathHash:         fileSvc.PathHash,
						ContentHash:      fileSvc.ContentHash,
						Size:             fileSvc.Size,
						Ctime:            fileSvc.Ctime,
						Mtime:            fileSvc.Mtime,
						UpdatedTimestamp: fileSvc.UpdatedTimestamp,
					},
				})
				needModifyCount++
				// 加入排除索引
				cDelFilesKeys[fileSvc.PathHash] = struct{}{}
			}
		}
	}

	// Iterate over server file list for processing
	// 遍历服务端文件列表进行处理
	for _, file := range list {
		// 如果该文件是客户端刚才通过参数告知删除的，则跳过下发
		if _, ok := cDelFilesKeys[file.PathHash]; ok {
			continue
		}

		if file.UpdatedTimestamp >= lastTime {
			lastTime = file.UpdatedTimestamp
		}

		if file.Action == "delete" {
			// Server already deleted, notify client to delete
			// 服务端已删除，通知客户端删除
			if _, ok := cFiles[file.PathHash]; ok {
				delete(cFilesKeys, file.PathHash)
				// 将消息添加到队列而非立即发送
				messageQueue = append(messageQueue, dto.WSQueuedMessage{
					Action: dto.FileSyncDelete,
					Data: dto.FileSyncDeleteMessage{
						Path:     file.Path,
						PathHash: file.PathHash,
						Ctime:    file.Ctime,
						Mtime:    file.Mtime,
						Size:     file.Size,
					},
				})
				needDeleteCount++
			}

		} else {

			if cFile, ok := cFiles[file.PathHash]; ok {
				// Client has this file
				// 客户端存在该文件
				delete(cFilesKeys, file.PathHash)

				if file.ContentHash == cFile.ContentHash && file.Mtime == cFile.Mtime {
					// Content and time match, no action
					// 内容与时间一致，无操作
					continue
				} else if file.ContentHash != cFile.ContentHash {
					// Content inconsistent
					// 内容不一致
					if file.Mtime > cFile.Mtime {
						// 服务端修改时间比客户端新, 通知客户端下载更新文件
						fileMessage := &dto.FileSyncModifyMessage{
							Path:             file.Path,
							PathHash:         file.PathHash,
							ContentHash:      file.ContentHash,
							Size:             file.Size,
							Ctime:            file.Ctime,
							Mtime:            file.Mtime,
							UpdatedTimestamp: file.UpdatedTimestamp,
						}
						// 将消息添加到队列而非立即发送
						messageQueue = append(messageQueue, dto.WSQueuedMessage{
							Action: dto.FileSyncUpdate,
							Data:   fileMessage,
						})
						needModifyCount++
					} else {
						// 服务端修改时间比客户端旧, 通知客户端上传文件
						session, ferr := h.handleFileUploadSessionCreate(c, params.Vault, cFile.Path, cFile.PathHash, cFile.ContentHash, cFile.Size, file.Ctime, cFile.Mtime)
						if ferr != nil {
							h.logError(c, "websocket_router.file.FileSync handleFileUploadSession err", ferr)
							continue
						}
						// 将消息添加到队列而非立即发送
						messageQueue = append(messageQueue, dto.WSQueuedMessage{
							Action: dto.FileUpload,
							Data: dto.FileSyncUploadMessage{
								Path:      session.Path,
								SessionID: session.ID,
								ChunkSize: session.ChunkSize,
							},
						})
						needUploadCount++
					}
				} else {
					// Content matches, but modification time differs, notify client to update file modification time
					// 内容一致, 但修改时间不一致, 通知客户端更新文件修改时间
					// Add message to queue instead of sending immediately
					// 将消息添加到队列而非立即发送
					messageQueue = append(messageQueue, dto.WSQueuedMessage{
						Action: dto.FileSyncMtime,
						Data: dto.FileSyncMtimeMessage{
							Path:  file.Path,
							Ctime: file.Ctime,
							Mtime: file.Mtime,
						},
					})
					needSyncMtimeCount++
				}
			} else {
				// File client doesn't have, notify client to download file
				// 客户端没有的文件, 通知客户端下载文件
				// 将消息添加到队列而非立即发送
				messageQueue = append(messageQueue, dto.WSQueuedMessage{
					Action: dto.FileSyncUpdate,
					Data: dto.FileSyncModifyMessage{
						Path:             file.Path,
						PathHash:         file.PathHash,
						ContentHash:      file.ContentHash,
						Size:             file.Size,
						Ctime:            file.Ctime,
						Mtime:            file.Mtime,
						UpdatedTimestamp: file.UpdatedTimestamp,
					},
				})
				needModifyCount++
			}
		}
	}

	if list == nil {
		lastTime = timex.Now().UnixMilli()
	}
	// Handle files that exist on client but not synced on server (request client upload)
	// 处理客户端存在但服务端未同步的文件（请求客户端上传）
	if len(cFilesKeys) > 0 {
		for pathHash := range cFilesKeys {
			file := cFiles[pathHash]
			// Create upload session and return FileUpload message
			// 创建上传会话并返回 FileUpload 消息
			session, ferr := h.handleFileUploadSessionCreate(c, params.Vault, file.Path, file.PathHash, file.ContentHash, file.Size, file.Ctime, file.Mtime)
			if ferr != nil {
				h.logError(c, "websocket_router.file.FileSync handleFileUploadSession err", ferr)
				continue
			}
			// Add message to queue instead of sending immediately
			// 将消息添加到队列而非立即发送
			messageQueue = append(messageQueue, dto.WSQueuedMessage{
				Action: dto.FileUpload,
				Data: dto.FileSyncUploadMessage{
					Path:      session.Path,
					SessionID: session.ID,
					ChunkSize: session.ChunkSize,
				},
			})
			needUploadCount++
		}
	}

	// Send FileSyncEnd message
	// 发送 FileSyncEnd 消息
	c.ToResponse(code.Success.WithData(
		dto.FileSyncEndMessage{
			LastTime:           lastTime,
			NeedUploadCount:    needUploadCount,
			NeedModifyCount:    needModifyCount,
			NeedSyncMtimeCount: needSyncMtimeCount,
			NeedDeleteCount:    needDeleteCount,
		},
	).WithVault(params.Vault).WithContext(params.Context), dto.FileSyncEnd)

	// Send queued messages individually
	// 逐条发送队列中的消息
	for _, item := range messageQueue {
		c.ToResponse(code.Success.WithData(item.Data).WithVault(params.Vault).WithContext(params.Context), item.Action)
	}
}

// cleanupSession cleans up discarded upload sessions due to completion or timeout.
// cleanupSession 清理因为完成或超时而废弃的上传会话。
func (h *FileWSHandler) handleFileUploadSessionCleanup(c *pkgapp.WebsocketClient, sessionID string) {
	binarySession := c.Server.GetSession(c.User.ID, sessionID)
	if binarySession == nil {
		return
	}
	c.Server.RemoveSession(c.User.ID, sessionID)

	session := binarySession.(pkgapp.SessionCleaner)
	session.Cleanup()

	h.logInfo(c, "cleanupSession: session cleaned up",
		zap.String("sessionID", sessionID))
}

func (h *FileWSHandler) handleFileUploadSessionTimeout(c *pkgapp.WebsocketClient, sessionID string, timeout time.Duration) context.CancelFunc {
	if timeout <= 0 {
		return nil
	}

	// Use context.Background() to ensure timeout timer is not affected by connection loss
	// This keeps the session valid during network fluctuation and reconnection
	// 使用 context.Background() 确保超时定时器不受连接断开影响
	// 这样在网络波动重连期间，会话依然有效
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		select {
		case <-timer.C:
			// Timeout triggered, clean up session
			// 超时触发，清理会话
			h.logWarn(c, "startSessionTimeout: session timeout, cleaning up",
				zap.String("sessionID", sessionID),
				zap.Duration("timeout", timeout))

			// Execute cleanup on timeout
			// 超时执行清理
			h.handleFileUploadSessionCleanup(c, sessionID)
		case <-ctx.Done():
			// Cancel timer
			// 取消定时器
			return
		}
	}()

	return cancel
}

// handleFileUploadSession initializes a file upload session and returns upload message.
// handleFileUploadSession 初始化一个文件上传会话并返回上传消息.
func (h *FileWSHandler) handleFileUploadSessionCreate(c *pkgapp.WebsocketClient, vault, path, pathHash, contentHash string, size, ctime, mtime int64) (*FileUploadBinaryChunkSession, error) {
	sessionID := uuid.New().String()
	cfg := h.App.Config()
	tempDir := cfg.App.TempPath
	if tempDir == "" {
		tempDir = "storage/temp"
	}

	// Create temp directory
	// 创建临时目录
	if err := os.MkdirAll(tempDir, 0754); err != nil {
		h.logError(c, "websocket_router.file.handleFileUploadSession.MkdirAll", err)
		return nil, err
	}

	var tempPath string

	for i := 0; i < 10; i++ {
		tryFileName := uuid.New().String()
		tryPath := filepath.Join(tempDir, tryFileName)

		if _, err := os.Stat(tryPath); os.IsNotExist(err) {
			tempPath = tryPath
			break
		}
	}

	file, err := os.Create(tempPath)
	if err != nil {
		h.logError(c, "websocket_router.file.handleFileUploadSession.Create", err)
		return nil, err
	}

	// Initialize chunk upload session
	// 初始化分块上传会话
	session := &FileUploadBinaryChunkSession{
		ID:          sessionID,
		Vault:       vault,
		Path:        path,
		PathHash:    pathHash,
		ContentHash: contentHash,
		Size:        size,
		Ctime:       ctime,
		Mtime:       mtime,
		ChunkSize:   getChunkSizeFromConfig(cfg), // 从注入的配置获取
		SavePath:    tempPath,
		FileHandle:  file,
		CreatedAt:   time.Now(),
	}
	// Adjust chunk size based on file size
	// 根据文件大小调整分块大小
	session.TotalChunks = util.Ceil(session.Size, session.ChunkSize)

	// Configure timeout
	// 配置超时时间
	var timeout time.Duration
	if cfg.App.UploadSessionTimeout != "" && cfg.App.UploadSessionTimeout != "0" {
		var err error
		timeout, err = util.ParseDuration(cfg.App.UploadSessionTimeout)
		if err != nil {
			h.logWarn(c, "handleFileUploadSession: invalid upload-session-timeout config, using default 20m",
				zap.String("config", cfg.App.UploadSessionTimeout),
				zap.Error(err))
			timeout = 20 * time.Minute
		}
	} else {
		timeout = 20 * time.Minute
	}

	// Start timeout cleanup task
	// 启动超时清理任务
	session.CancelFunc = h.handleFileUploadSessionTimeout(c, sessionID, timeout)

	// Register to global server
	// 注册到全局服务器
	c.Server.SetSession(c.User.ID, session.ID, session)

	return session, nil
}

// handleFileChunkDownloadSendChunks executes file chunk sending.
// Runs in independent goroutine, reads file and sends binary chunks via WebSocket.
// handleFileChunkDownloadSendChunks 执行文件分片发送。
// 在独立的 goroutine 中运行,读取文件并通过 WebSocket 发送二进制分片。
func (h *FileWSHandler) handleFileChunkDownloadSendChunks(c *pkgapp.WebsocketClient, session *FileDownloadChunkSession) {
	logger := h.App.Logger()
	// Get timeout from config, default 1 hour
	// 从配置获取超时时间，默认 1 小时
	timeout := 1 * time.Hour
	cfg := h.App.Config()
	if cfg.App.DownloadSessionTimeout != "" && cfg.App.DownloadSessionTimeout != "0" {
		if t, err := util.ParseDuration(cfg.App.DownloadSessionTimeout); err == nil {
			timeout = t
		} else {
			LogWarnWithLogger(logger, c, "sendFileChunks: invalid download-session-timeout config, using default 1h",
				zap.String("config", cfg.App.DownloadSessionTimeout),
				zap.Error(err))
		}
	}

	// Create timeout context, based on WebSocket connection context
	// 创建超时上下文，基于 WebSocket 连接的 context
	ctx, cancel := context.WithTimeout(c.Context(), timeout)
	defer cancel()

	// Open file
	// 打开文件
	file, err := os.Open(session.SavePath)
	if err != nil {
		LogErrorWithLogger(logger, c, "sendFileChunks: failed to open file", err)
		c.ToResponse(code.ErrorFileGetFailed.WithDetails("failed to open file"))
		return
	}
	defer file.Close()

	LogInfoWithLogger(logger, c, "sendFileChunks: starting file download",
		zap.String("sessionID", session.SessionID),
		zap.String("path", session.Path),
		zap.Int64("size", session.Size),
		zap.Int64("totalChunks", session.TotalChunks))

	// Loop to send chunks
	// 循环发送分片
	for chunkIndex := int64(0); chunkIndex < session.TotalChunks; chunkIndex++ {
		// Check timeout
		// 检查超时
		select {
		case <-ctx.Done():
			LogWarnWithLogger(logger, c, "sendFileChunks: download timeout",
				zap.String("sessionID", session.SessionID),
				zap.Int64("sentChunks", chunkIndex),
				zap.Int64("totalChunks", session.TotalChunks))
			return
		default:
		}

		// Calculate current chunk size
		// 计算当前分片的大小
		chunkStart := chunkIndex * session.ChunkSize
		chunkEnd := chunkStart + session.ChunkSize
		if chunkEnd > session.Size {
			chunkEnd = session.Size
		}
		currentChunkSize := chunkEnd - chunkStart

		// 读取分片数据
		chunkData := make([]byte, currentChunkSize)
		n, err := file.ReadAt(chunkData, chunkStart)
		if err != nil && err.Error() != "EOF" {
			LogErrorWithLogger(logger, c, "sendFileChunks: failed to read chunk", err)
			c.ToResponse(code.ErrorFileGetFailed.WithDetails("failed to read file chunk"))
			return
		}

		// 构造二进制消息
		// 格式: [36 bytes session_id][4 bytes chunk_index][chunk_data]
		headerSize := 40
		packet := make([]byte, headerSize+n)

		// 1. Session ID (36 bytes)
		copy(packet[0:36], []byte(session.SessionID))

		// 2. Chunk Index (4 bytes, Big Endian)
		binary.BigEndian.PutUint32(packet[36:40], uint32(chunkIndex))

		// 3. Chunk Data
		copy(packet[40:], chunkData[:n])

		// 发送二进制消息
		err = c.SendBinary(dto.VaultFileMsgType, packet)
		if err != nil {
			LogErrorWithLogger(logger, c, "sendFileChunks: failed to send chunk", err)
			return
		}

		// 每发送 100 个分片记录一次日志
		if (chunkIndex+1)%100 == 0 || chunkIndex == session.TotalChunks-1 {
			LogInfoWithLogger(logger, c, "sendFileChunks: progress",
				zap.String("sessionID", session.SessionID),
				zap.Int64("sent", chunkIndex+1),
				zap.Int64("total", session.TotalChunks))
		}
	}

	LogInfoWithLogger(logger, c, "sendFileChunks: download completed",
		zap.String("sessionID", session.SessionID),
		zap.String("path", session.Path),
		zap.Int64("totalChunks", session.TotalChunks))
}

func (h *FileWSHandler) FileRePush(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.FileGetRequest{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.file.FileRePush.BindAndValid")
		return
	}

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "FileRePush", params.Path, params.Vault)

	ctx := c.Context()
	// 获取或创建仓库
	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	fileSvc, err := h.App.FileService.Get(ctx, c.User.UID, params)
	if err != nil {
		h.respondError(c, code.ErrorFileGetFailed, err, "websocket_router.file.FileRePush.Get")
		return
	}

	if fileSvc != nil && fileSvc.Action != "delete" {
		c.ToResponse(code.Success.WithData(
			dto.FileSyncModifyMessage{
				Path:             fileSvc.Path,
				PathHash:         fileSvc.PathHash,
				ContentHash:      fileSvc.ContentHash,
				Size:             fileSvc.Size,
				Ctime:            fileSvc.Ctime,
				Mtime:            fileSvc.Mtime,
				UpdatedTimestamp: fileSvc.UpdatedTimestamp,
			},
		).WithVault(params.Vault), dto.FileSyncUpdate)
	} else {
		c.ToResponse(code.ErrorFileGetFailed)
	}
}

// getChunkSizeFromConfig 从注入的配置获取分片大小, 默认为 512KB
func getChunkSizeFromConfig(cfg *app.AppConfig) int64 {
	return util.ParseSize(cfg.App.FileChunkSize, 1024*512)
}
