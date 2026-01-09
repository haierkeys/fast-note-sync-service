package websocket_router

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"

	"go.uber.org/zap"
)

// FileWSHandler WebSocket 文件处理器
// 使用 App Container 注入依赖
type FileWSHandler struct {
	*WSHandler
}

// NewFileWSHandler 创建 FileWSHandler 实例
func NewFileWSHandler(a *app.App) *FileWSHandler {
	return &FileWSHandler{
		WSHandler: NewWSHandler(a),
	}
}

type FileMessage struct {
	Path             string `json:"path" form:"path"`                     // 路径信息（文件路径）
	PathHash         string `json:"pathHash" form:"pathHash"`             // 路径哈希值，用于快速查找
	ContentHash      string `json:"contentHash" form:"contentHash"`       // 内容哈希，用于判定内容是否变更
	SavePath         string `json:"savePath" form:"savePath"  binding:""` // 文件保存路径
	Size             int64  `json:"size" form:"size"`                     // 文件大小
	Ctime            int64  `json:"ctime" form:"ctime"`                   // 创建时间戳（秒）
	Mtime            int64  `json:"mtime" form:"mtime"`                   // 文件修改时间戳（秒）
	UpdatedTimestamp int64  `json:"lastTime" form:"updatedTimestamp"`     // 记录更新时间戳（用于同步）
}

type FileUploadBinaryChunkSession struct {
	ID             string             // 会话 ID
	Vault          string             // 仓库名称
	Path           string             // 文件路径
	PathHash       string             // 文件路径哈希值
	ContentHash    string             // 文件内容哈希值
	Ctime          int64              // 创建时间
	Mtime          int64              // 修改时间
	Size           int64              // 文件大小
	TotalChunks    int64              // 总分块数
	UploadedChunks int64              // 已上传分块数
	UploadedBytes  int64              // 已上传字节数
	ChunkSize      int64              // 分块大小
	SavePath       string             // 临时保存路径
	FileHandle     *os.File           // 文件句柄
	mu             sync.Mutex         // 互斥锁，保护并发操作
	CreatedAt      time.Time          // 创建时间
	CancelFunc     context.CancelFunc // 取消函数，用于超时控制
}

func (s *FileUploadBinaryChunkSession) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.FileHandle != nil {
		_ = s.FileHandle.Close()
		s.FileHandle = nil
	}
	if s.SavePath != "" {
		_ = os.Remove(s.SavePath)
		s.SavePath = ""
	}
}

// FileDownloadChunkSession 定义文件分块下载的会话状态。
// 用于跟踪大文件分块下载的进度和文件信息。
type FileDownloadChunkSession struct {
	SessionID   string // 会话 ID
	Path        string // 文件路径(用于日志)
	Size        int64  // 文件大小
	TotalChunks int64  // 总分块数
	ChunkSize   int64  // 分块大小
	SavePath    string // 文件实际保存路径
}

// FileSyncEndMessage 定义文件同步结束时的消息结构。
type FileSyncEndMessage struct {
	LastTime           int64           `json:"lastTime" form:"lastTime"`                     // 最后同步时间
	NeedUploadCount    int64           `json:"needUploadCount" form:"needUploadCount"`       // 需要上传的数量
	NeedModifyCount    int64           `json:"needModifyCount" form:"needModifyCount"`       // 需要修改的数量
	NeedSyncMtimeCount int64           `json:"needSyncMtimeCount" form:"needSyncMtimeCount"` // 需要同步修改时间的数量
	NeedDeleteCount    int64           `json:"needDeleteCount" form:"needDeleteCount"`       // 需要删除的数量
	Messages           []queuedMessage `json:"messages"`                                     // 合并的消息队列
}

// FileUploadMessage 定义服务端通知客户端需要上传文件的消息结构。
type FileUploadMessage struct {
	Path      string `json:"path"`      // 文件路径
	SessionID string `json:"sessionId"` // 会话 ID
	ChunkSize int64  `json:"chunkSize"` // 分块大小
}

// FileDownloadMessage 定义服务端通知客户端准备下载文件的消息结构。
type FileDownloadMessage struct {
	Path        string `json:"path"`        // 文件路径
	Ctime       int64  `json:"ctime"`       // 创建时间
	Mtime       int64  `json:"mtime"`       // 修改时间
	SessionID   string `json:"sessionId"`   // 会话 ID
	ChunkSize   int64  `json:"chunkSize"`   // 分块大小
	TotalChunks int64  `json:"totalChunks"` // 总分块数
	Size        int64  `json:"size"`        // 文件总大小
}

// FileSyncMtimeMessage 定义文件元数据更新消息结构。
type FileSyncMtimeMessage struct {
	Path  string `json:"path"`   // 文件路径
	Ctime int64  `json:"ctime" ` // 创建时间
	Mtime int64  `json:"mtime" ` // 修改时间
}

type FileDeleteMessage struct {
	Path string `json:"path" form:"path"` // 路径信息（文件路径）
}

// FileUploadCheck 检查文件上传请求，初始化上传会话或确认无需上传。
func (h *FileWSHandler) FileUploadCheck(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.FileUpdateCheckRequest{}

	// 绑定并验证参数
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.logError(c, "websocket_router.file.FileUploadCheck.BindAndValid", errs)
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

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

	// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	// 检查文件更新状态
	updateMode, fileSvc, err := h.App.FileService.UploadCheck(ctx, c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorFileUploadCheckFailed.WithDetails(err.Error()))
		return
	}

	// UpdateContent 或 Create 模式，需要客户端上传文件
	switch updateMode {
	case "UpdateContent", "Create":

		fileUploadMessage, err := h.handleFileUploadSessionCreate(c, params.Vault, params.Path, params.PathHash, params.ContentHash, params.Size, params.Ctime, params.Mtime)
		if err != nil {
			c.ToResponse(code.ErrorFileUploadCheckFailed.WithDetails(err.Error()))
			return
		}

		c.ToResponse(code.Success.WithData(fileUploadMessage), "FileUpload")
		return

	case "UpdateMtime":
		// 当用户 mtime 小于服务端 mtime 时，通知用户更新mtime
		fileSyncMtimeMessage := &FileSyncMtimeMessage{
			Path:  fileSvc.Path,
			Ctime: fileSvc.Ctime,
			Mtime: fileSvc.Mtime,
		}
		c.ToResponse(code.Success.WithData(fileSyncMtimeMessage), "FileSyncMtime")
		return
	default:
		// 无需更新
		c.ToResponse(code.SuccessNoUpdate)
	}
}

// FileUploadChunkBinary 处理文件分块上传的二进制数据。
func (h *FileWSHandler) FileUploadChunkBinary(c *pkgapp.WebsocketClient, data []byte) {
	if len(data) < 40 {
		h.logError(c, "websocket_router.file.FileUploadChunkBinary", fmt.Errorf("invalid data length: %d", len(data)))
		return
	}

	// 解析会话 ID 和分块索引
	sessionID := string(data[:36])
	chunkIndex := binary.BigEndian.Uint32(data[36:40])
	chunkData := data[40:]

	c.BinaryMu.Lock()
	binarySession, exists := c.BinaryChunkSessions[sessionID]
	c.BinaryMu.Unlock()

	if !exists {
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

	// 计算写入偏移量并写入数据
	offset := int64(chunkIndex) * int64(session.ChunkSize)

	if _, err := session.FileHandle.WriteAt(chunkData, offset); err != nil {
		h.logError(c, "websocket_router.file.FileUploadChunkBinary.WriteAt", err)
		c.ToResponse(code.ErrorFileUploadFailed.WithData(map[string]string{
			"sessionID": sessionID,
		}).WithDetails(err.Error()))
		return
	}

	// 更新已上传计数和字节数
	session.mu.Lock()
	session.UploadedChunks++
	session.UploadedBytes += int64(len(chunkData))
	uploadedBytes := session.UploadedBytes
	session.mu.Unlock()

	// 检查是否所有数据都已上传(根据字节数判断)
	if uploadedBytes >= session.Size {

		h.logInfo(c, "FileUploadComplete: upload finished",
			zap.String("sessionID", sessionID),
			zap.String("path", session.Path),
			zap.Int64("uploadedBytes", uploadedBytes),
			zap.Int64("totalSize", session.Size),
			zap.Int64("uploadedChunks", session.UploadedChunks),
			zap.Int64("totalChunks", session.TotalChunks))

		// 获取并移除会话
		c.BinaryMu.Lock()
		delete(c.BinaryChunkSessions, session.ID)
		c.BinaryMu.Unlock()

		// 取消超时定时器
		if session.CancelFunc != nil {
			session.CancelFunc()
		}

		// 关闭临时文件句柄
		if err := session.FileHandle.Close(); err != nil {
			h.logError(c, "FileUploadComplete: failed to close file handle", err)
			c.ToResponse(code.ErrorFileUploadFailed.WithData(map[string]string{
				"sessionID": sessionID,
			}).WithDetails(err.Error()))
			return
		}
		session.FileHandle = nil // 避免再次清理时重复关闭

		ctx := c.Context()

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

		// 更新或创建文件记录 (DAO 层会自动将 SavePath 里的临时文件移动到 f_{id} 文件夹)
		_, fileSvc, err := h.App.FileService.UploadComplete(ctx, c.User.UID, svcParams)

		if err != nil {
			c.ToResponse(code.ErrorFileModifyOrCreateFailed.WithDetails(err.Error()))
			return
		}

		// 用于给客户端
		//sessionID := uuid.New().String()

		fileMsg := &FileMessage{
			Path:             fileSvc.Path,
			PathHash:         fileSvc.PathHash,
			ContentHash:      fileSvc.ContentHash,
			SavePath:         fileSvc.SavePath,
			Size:             fileSvc.Size,
			Ctime:            fileSvc.Ctime,
			Mtime:            fileSvc.Mtime,
			UpdatedTimestamp: fileSvc.UpdatedTimestamp,
		}

		c.ToResponse(code.Success)
		// 广播文件更新消息
		c.BroadcastResponse(code.Success.WithData(fileMsg).WithVault(session.Vault), true, "FileSyncUpdate")
	}
}

// FileDelete 处理文件删除请求。
func (h *FileWSHandler) FileDelete(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.FileDeleteRequest{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.logError(c, "websocket_router.file.FileDelete.BindAndValid", errs)
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	ctx := c.Context()

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "FileDelete", params.Path, params.Vault)

	// 获取或创建仓库
	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	// 执行删除逻辑
	fileSvc, err := h.App.FileService.Delete(ctx, c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorFileDeleteFailed.WithDetails(err.Error()))
		return
	}

	c.ToResponse(code.Success)

	fileDeleteMessage := convert.StructAssign(fileSvc, &FileDeleteMessage{}).(*FileDeleteMessage)
	// 广播文件删除消息
	c.BroadcastResponse(code.Success.WithData(fileDeleteMessage).WithVault(params.Vault), true, "FileSyncDelete")
}

// FileChunkDownload 处理文件分片下载请求。
// 客户端通过此接口请求下载文件,服务端创建下载会话并开始发送分片。
func (h *FileWSHandler) FileChunkDownload(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.FileGetRequest{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.logError(c, "websocket_router.file.FileChunkDownload.BindAndValid", errs)
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	ctx := c.Context()

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "FileChunkDownload", params.Path, params.Vault)

	// 获取或创建仓库
	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	// 获取文件信息
	fileSvc, err := h.App.FileService.Get(ctx, c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorFileGetFailed.WithDetails(err.Error()))
		return
	}

	// 检查文件是否存在于磁盘
	if _, err := os.Stat(fileSvc.SavePath); os.IsNotExist(err) {
		h.logError(c, "websocket_router.file.FileChunkDownload",
			fmt.Errorf("file not found on disk: %s (pathHash: %s)", fileSvc.SavePath, fileSvc.PathHash))
		c.ToResponse(code.ErrorFileGetFailed.WithDetails("file not found on disk"))
		return
	}

	// 创建下载会话
	sessionID := uuid.New().String()
	chunkSize := getChunkSize() // 从配置获取

	// 计算总分块数
	totalChunks := util.Ceil(fileSvc.Size, chunkSize)

	// 初始化下载会话
	session := &FileDownloadChunkSession{
		SessionID:   sessionID,
		Path:        fileSvc.Path,
		Size:        fileSvc.Size,
		TotalChunks: totalChunks,
		ChunkSize:   chunkSize,
		SavePath:    fileSvc.SavePath,
	}

	// 构造下载消息
	fileDownloadMessage := &FileDownloadMessage{
		Path:        fileSvc.Path,
		Ctime:       fileSvc.Ctime,
		Mtime:       fileSvc.Mtime,
		SessionID:   session.SessionID,
		ChunkSize:   session.ChunkSize,
		TotalChunks: session.TotalChunks,
		Size:        session.Size,
	}

	// 发送下载准备消息
	c.ToResponse(code.Success.WithData(fileDownloadMessage), "FileSyncChunkDownload")

	// 启动分片发送,传入超时时间
	go handleFileChunkDownloadSendChunks(c, session)
}

// FileSync 批量检测用户文件是否需要更新。
// 对比客户端和服务端的文件列表，决定哪些文件需要上传、更新或删除。
func (h *FileWSHandler) FileSync(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.FileSyncRequest{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.logError(c, "websocket_router.file.FileSync.BindAndValid", errs)
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	ctx := c.Context()

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "FileSync", "", params.Vault)

	// 获取或创建仓库
	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	// 获取最后一次同步后的变更文件列表
	list, err := h.App.FileService.Sync(ctx, c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorFileListFailed.WithDetails(err.Error()))
		return
	}

	// 构建客户端文件索引
	var cFiles map[string]dto.FileSyncCheckRequest = make(map[string]dto.FileSyncCheckRequest, 0)
	var cFilesKeys map[string]struct{} = make(map[string]struct{}, 0)

	if len(params.Files) > 0 {
		for _, file := range params.Files {
			cFiles[file.PathHash] = file
			cFilesKeys[file.PathHash] = struct{}{}
		}
	}

	// 创建消息队列，用于收集所有待发送的消息
	var messageQueue []queuedMessage

	var lastTime int64
	var needUploadCount int64
	var needModifyCount int64
	var needSyncMtimeCount int64
	var needDeleteCount int64

	// 遍历服务端文件列表进行处理
	for _, file := range list {
		if file.UpdatedTimestamp >= lastTime {
			lastTime = file.UpdatedTimestamp
		}

		if file.Action == "delete" {
			// 服务端已删除，通知客户端删除
			if _, ok := cFiles[file.PathHash]; ok {
				delete(cFilesKeys, file.PathHash)
				// 将消息添加到队列而非立即发送
				messageQueue = append(messageQueue, queuedMessage{
					Action: "FileSyncDelete",
					Data:   file,
				})
				needDeleteCount++
			}

		} else {

			if cFile, ok := cFiles[file.PathHash]; ok {
				// 客户端存在该文件
				delete(cFilesKeys, file.PathHash)

				if file.ContentHash == cFile.ContentHash && file.Mtime == cFile.Mtime {
					// 内容与时间一致，无操作
					continue
				} else if file.ContentHash != cFile.ContentHash {
					// 内容不一致
					if file.Mtime > cFile.Mtime {
						// 服务端修改时间比客户端新, 通知客户端下载更新文件
						fileMessage := &FileMessage{
							Path:             file.Path,
							PathHash:         file.PathHash,
							ContentHash:      file.ContentHash,
							SavePath:         file.SavePath,
							Size:             file.Size,
							Ctime:            file.Ctime,
							Mtime:            file.Mtime,
							UpdatedTimestamp: file.UpdatedTimestamp,
						}
						// 将消息添加到队列而非立即发送
						messageQueue = append(messageQueue, queuedMessage{
							Action: "FileSyncUpdate",
							Data:   fileMessage,
						})
						needModifyCount++
					} else {
						// 服务端修改时间比客户端旧, 通知客户端上传文件
						fileUploadMessage, ferr := h.handleFileUploadSessionCreate(c, params.Vault, cFile.Path, cFile.PathHash, cFile.ContentHash, cFile.Size, file.Ctime, cFile.Mtime)
						if ferr != nil {
							global.Logger.Error("websocket_router.file.FileSync handleFileUploadSession err", zap.Error(ferr))
							continue
						}
						// 将消息添加到队列而非立即发送
						messageQueue = append(messageQueue, queuedMessage{
							Action: "FileUpload",
							Data:   fileUploadMessage,
						})
						needUploadCount++
					}
				} else {
					// 内容一致, 但修改时间不一致, 通知客户端更新文件修改时间
					fileSyncMtimeMessage := &FileSyncMtimeMessage{
						Path:  file.Path,
						Ctime: file.Ctime,
						Mtime: file.Mtime,
					}
					// 将消息添加到队列而非立即发送
					messageQueue = append(messageQueue, queuedMessage{
						Action: "FileSyncMtime",
						Data:   fileSyncMtimeMessage,
					})
					needSyncMtimeCount++
				}
			} else {
				// 客户端没有的文件, 通知客户端下载文件
				fileMessage := &FileMessage{
					Path:             file.Path,
					PathHash:         file.PathHash,
					ContentHash:      file.ContentHash,
					SavePath:         file.SavePath,
					Size:             file.Size,
					Ctime:            file.Ctime,
					Mtime:            file.Mtime,
					UpdatedTimestamp: file.UpdatedTimestamp,
				}
				// 将消息添加到队列而非立即发送
				messageQueue = append(messageQueue, queuedMessage{
					Action: "FileSyncUpdate",
					Data:   fileMessage,
				})
				needModifyCount++
			}
		}
	}

	if list == nil {
		lastTime = timex.Now().UnixMilli()
	}
	// 处理客户端存在但服务端未同步的文件（请求客户端上传）
	if len(cFilesKeys) > 0 {
		for pathHash := range cFilesKeys {
			file := cFiles[pathHash]
			// 创建上传会话并返回 FileUpload 消息
			fileUploadMessage, ferr := h.handleFileUploadSessionCreate(c, params.Vault, file.Path, file.PathHash, file.ContentHash, file.Size, 0, file.Mtime)
			if ferr != nil {
				global.Logger.Error("websocket_router.file.FileSync handleFileUploadSession err", zap.Error(ferr))
				continue
			}
			// 将消息添加到队列而非立即发送
			messageQueue = append(messageQueue, queuedMessage{
				Action: "FileUpload",
				Data:   fileUploadMessage,
			})
			needUploadCount++
		}
	}

	// 发送 FileSyncEnd 消息，包含所有合并的消息
	message := &FileSyncEndMessage{
		LastTime:           lastTime,
		NeedUploadCount:    needUploadCount,
		NeedModifyCount:    needModifyCount,
		NeedSyncMtimeCount: needSyncMtimeCount,
		NeedDeleteCount:    needDeleteCount,
		Messages:           messageQueue,
	}

	c.ToResponse(code.Success.WithData(message).WithVault(params.Vault), "FileSyncEnd")
}

// cleanupSession 清理因为完成或超时而废弃的上传会话。
func (h *FileWSHandler) handleFileUploadSessionCleanup(c *pkgapp.WebsocketClient, sessionID string) {
	c.BinaryMu.Lock()
	binarySession, exists := c.BinaryChunkSessions[sessionID]
	if exists {
		delete(c.BinaryChunkSessions, sessionID)
	}
	c.BinaryMu.Unlock()

	if !exists {
		return
	}

	session := binarySession.(pkgapp.SessionCleaner)
	session.Cleanup()

	h.logInfo(c, "cleanupSession: session cleaned up",
		zap.String("sessionID", sessionID))
}

// startSessionTimeout 启动会话超时定时器。
// 返回一个取消函数，调用后可阻止超时清理的执行。
func (h *FileWSHandler) handleFileUploadSessionTimeout(c *pkgapp.WebsocketClient, sessionID string, timeout time.Duration) context.CancelFunc {
	if timeout <= 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(c.Context())

	go func() {
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		select {
		case <-timer.C:
			// 超时触发，清理会话
			LogWarnWithTrace(c, "startSessionTimeout: session timeout, cleaning up",
				zap.String("sessionID", sessionID),
				zap.Duration("timeout", timeout))

			// 超时执行清理
			h.handleFileUploadSessionCleanup(c, sessionID)
		case <-ctx.Done():
			// 取消定时器
			return
		}
	}()

	return cancel
}

// handleFileUploadSession 初始化一个文件上传会话并返回上传消息。
func (h *FileWSHandler) handleFileUploadSessionCreate(c *pkgapp.WebsocketClient, vault, path, pathHash, contentHash string, size, ctime, mtime int64) (*FileUploadMessage, error) {
	sessionID := uuid.New().String()
	tempDir := global.Config.App.TempPath
	if tempDir == "" {
		tempDir = "storage/temp"
	}

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
		ChunkSize:   getChunkSize(), // 从配置获取
		SavePath:    tempPath,
		FileHandle:  file,
		CreatedAt:   time.Now(),
	}
	// 根据文件大小调整分块大小
	session.TotalChunks = util.Ceil(session.Size, session.ChunkSize)

	// 配置超时时间
	var timeout time.Duration
	if global.Config.App.UploadSessionTimeout != "" && global.Config.App.UploadSessionTimeout != "0" {
		var err error
		timeout, err = util.ParseDuration(global.Config.App.UploadSessionTimeout)
		if err != nil {
			LogWarnWithTrace(c, "handleFileUploadSession: invalid upload-session-timeout config, using default 20m",
				zap.String("config", global.Config.App.UploadSessionTimeout),
				zap.Error(err))
			timeout = 20 * time.Minute
		}
	} else {
		timeout = 20 * time.Minute
	}

	// 启动超时清理任务
	session.CancelFunc = h.handleFileUploadSessionTimeout(c, sessionID, timeout)

	// 注册会话
	c.BinaryMu.Lock()
	c.BinaryChunkSessions[sessionID] = session
	c.BinaryMu.Unlock()

	return &FileUploadMessage{
		Path:      path,
		SessionID: session.ID,
		ChunkSize: session.ChunkSize,
	}, nil
}

// sendFileChunks 执行文件分片发送。
// 在独立的 goroutine 中运行,读取文件并通过 WebSocket 发送二进制分片。
func handleFileChunkDownloadSendChunks(c *pkgapp.WebsocketClient, session *FileDownloadChunkSession) {
	// 创建超时上下文，基于 WebSocket 连接的 context
	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	// 打开文件
	file, err := os.Open(session.SavePath)
	if err != nil {
		global.Logger.Error("sendFileChunks: failed to open file",
			zap.String("sessionID", session.SessionID),
			zap.String("path", session.SavePath),
			zap.Error(err))
		c.ToResponse(code.ErrorFileGetFailed.WithDetails("failed to open file"))
		return
	}
	defer file.Close()

	global.Logger.Info("sendFileChunks: starting file download",
		zap.String("sessionID", session.SessionID),
		zap.String("path", session.Path),
		zap.Int64("size", session.Size),
		zap.Int64("totalChunks", session.TotalChunks))

	// 循环发送分片
	for chunkIndex := int64(0); chunkIndex < session.TotalChunks; chunkIndex++ {
		// 检查超时
		select {
		case <-ctx.Done():
			global.Logger.Warn("sendFileChunks: download timeout",
				zap.String("sessionID", session.SessionID),
				zap.Int64("sentChunks", chunkIndex),
				zap.Int64("totalChunks", session.TotalChunks))
			return
		default:
		}

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
			global.Logger.Error("sendFileChunks: failed to read chunk",
				zap.String("sessionID", session.SessionID),
				zap.Int64("chunkIndex", chunkIndex),
				zap.Error(err))
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
		err = c.SendBinary(VaultFileSync, packet)
		if err != nil {
			global.Logger.Error("sendFileChunks: failed to send chunk",
				zap.String("sessionID", session.SessionID),
				zap.Int64("chunkIndex", chunkIndex),
				zap.Error(err))
			return
		}

		// 每发送 100 个分片记录一次日志
		if (chunkIndex+1)%100 == 0 || chunkIndex == session.TotalChunks-1 {
			global.Logger.Info("sendFileChunks: progress",
				zap.String("sessionID", session.SessionID),
				zap.Int64("sent", chunkIndex+1),
				zap.Int64("total", session.TotalChunks))
		}
	}

	global.Logger.Info("sendFileChunks: download completed",
		zap.String("sessionID", session.SessionID),
		zap.String("path", session.Path),
		zap.Int64("totalChunks", session.TotalChunks))
}

// getChunkSize 获取配置的分片大小, 默认为 1MB
func getChunkSize() int64 {
	// 从全局配置中读取文件分片大小的设置字符串
	sizeStr := global.Config.App.FileChunkSize
	// 如果配置为空，则直接按照默认值 1MB 处理
	if sizeStr == "" {
		return 1024 * 512 // 默认 512KB
	}

	// 预处理字符串：去除首尾空格并转换为大写，以便统一处理后缀（如 mb, MB, Mb 等）
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))
	// 定义基础倍数，默认为 1（即单位为字节 B）
	var multiplier int64 = 1

	// 判断是否包含 MB 后缀
	if strings.HasSuffix(sizeStr, "MB") {
		multiplier = 1024 * 1024                    // 如果是 MB，倍数为 1024*1024
		sizeStr = strings.TrimSuffix(sizeStr, "MB") // 去除后缀，只保留数字部分字符串
	} else if strings.HasSuffix(sizeStr, "KB") {
		// 判断是否包含 KB 后缀
		multiplier = 1024                           // 如果是 KB，倍数为 1024
		sizeStr = strings.TrimSuffix(sizeStr, "KB") // 去除后缀
	} else if strings.HasSuffix(sizeStr, "B") {
		// 判断是否包含 B 后缀
		multiplier = 1                             // 如果是 B，倍数为 1
		sizeStr = strings.TrimSuffix(sizeStr, "B") // 去除后缀
	}

	// 解析剩余的数字字符串为 int64 整数
	// 再次 trim 是为了防止 "1 MB" 这种中间有空格的情况被去除后缀后变成 "1 " 导致解析失败
	size, err := strconv.ParseInt(strings.TrimSpace(sizeStr), 10, 64)

	// 如果解析出错（例如包含非数字字符）或者数值小于等于0
	if err != nil || size <= 0 {
		return 1024 * 512 // 解析失败或配置了无效值，回退到默认值 512KB
	}

	// 返回最终计算出的字节大小（数字 * 倍数）
	return size * multiplier
}
