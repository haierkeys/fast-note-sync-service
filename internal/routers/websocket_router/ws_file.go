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
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"github.com/haierkeys/fast-note-sync-service/pkg/fileurl"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"

	"go.uber.org/zap"
)

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

// FileUploadBinaryChunkSession 定义文件分块上传的会话状态。
// 用于跟踪大文件分块上传的进度和临时文件信息。
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
	ChunkSize      int64              // 分块大小
	SavePath       string             // 临时保存路径
	FileHandle     *os.File           // 文件句柄
	mu             sync.Mutex         // 互斥锁，保护并发操作
	CreatedAt      time.Time          // 创建时间
	CancelFunc     context.CancelFunc // 取消函数，用于超时控制
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
	Vault    string `json:"vault" form:"vault"`       // 仓库名称
	LastTime int64  `json:"lastTime" form:"lastTime"` // 最后同步时间
}

// FileNeedUploadMessage 定义服务端通知客户端需要上传文件的消息结构。
type FileUploadMessage struct {
	Path      string `json:"path"`      // 文件路径
	Ctime     int64  `json:"ctime" `    // 创建时间
	Mtime     int64  `json:"mtime" `    // 修改时间
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

// FileSyncNeedUploadMessage 定义文件同步中需要上传的文件消息结构。
type FileNeedUploadMessage struct {
	Path string `json:"path"` // 文件路径
}

type FileDeleteMessage struct {
	Path string `json:"path" form:"path"` // 路径信息（文件路径）
}

// FileUploadCheck 检查文件上传请求，初始化上传会话或确认无需上传。
func FileUploadCheck(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.FileUpdateCheckParams{}

	// 绑定并验证参数
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("websocket_router.file.FileUploadCheck.BindAndValid errs: %v", zap.Error(errs))
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

	svc := service.New(c.Ctx).WithSF(c.SF)

	// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
	svc.VaultGetOrCreate(params.Vault, c.User.UID)

	// 检查文件更新状态
	updateMode, fileSvc, err := svc.FileUpdateCheck(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorFileUploadCheckFailed.WithDetails(err.Error()))
		return
	}

	// UpdateContent 或 Create 模式，需要客户端上传文件
	if updateMode == "UpdateContent" || updateMode == "Create" {

		sessionID := uuid.New().String()
		tempDir := global.Config.App.TempPath
		if tempDir == "" {
			tempDir = "storage/temp"
		}
		// 创建临时目录
		if err := os.MkdirAll(tempDir, 0644); err != nil {
			global.Logger.Error("websocket_router.file.FileUploadCheck MkdirAll err", zap.Error(err))
			c.ToResponse(code.ErrorFileUploadCheckFailed.WithDetails(err.Error()))
			return
		}

		// 获取文件扩展名并生成临时路径
		fileExt := filepath.Ext(params.Path)
		tempPath := filepath.Join(tempDir, fmt.Sprintf("%s%s", sessionID, fileExt))
		file, err := os.Create(tempPath)
		if err != nil {
			global.Logger.Error("websocket_router.file.FileUploadCheck Create file err", zap.Error(err))
			c.ToResponse(code.ErrorFileUploadCheckFailed.WithDetails(err.Error()))
			return
		}

		// 初始化分块上传会话
		session := &FileUploadBinaryChunkSession{
			ID:          sessionID,
			Vault:       params.Vault,
			Path:        params.Path,
			PathHash:    params.PathHash,
			ContentHash: params.ContentHash,
			Size:        params.Size,
			Ctime:       params.Ctime,
			Mtime:       params.Mtime,
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
			timeout, err = time.ParseDuration(global.Config.App.UploadSessionTimeout)
			if err != nil {
				global.Logger.Warn("FileUploadCheck: invalid upload-session-timeout config, using default 5m",
					zap.String("config", global.Config.App.UploadSessionTimeout),
					zap.Error(err))
				timeout = 5 * time.Minute
			}
		} else {
			timeout = 5 * time.Minute
		}

		// 启动超时清理任务
		session.CancelFunc = startSessionTimeout(c, sessionID, timeout)

		// 注册会话
		c.BinaryMu.Lock()
		c.BinaryChunkSessions[sessionID] = session
		c.BinaryMu.Unlock()

		fileUploadMessage := &FileUploadMessage{
			Path:      params.Path,
			Ctime:     params.Ctime,
			Mtime:     params.Mtime,
			SessionID: session.ID,
			ChunkSize: session.ChunkSize,
		}

		c.ToResponse(code.Success.WithData(fileUploadMessage), "FileUpload")
		return

	} else if updateMode == "UpdateMtime" {
		// 仅更新修改时间
		fileSyncMtimeMessage := &FileSyncMtimeMessage{
			Path:  fileSvc.Path,
			Ctime: fileSvc.Ctime,
			Mtime: fileSvc.Mtime,
		}
		c.ToResponse(code.Success.WithData(fileSyncMtimeMessage), "FileSyncMtime")
		return
	}
	// 无需更新
	c.ToResponse(code.SuccessNoUpdate.Reset())
}

// FileUploadChunkBinary 处理文件分块上传的二进制数据。
func FileUploadChunkBinary(c *app.WebsocketClient, data []byte) {
	if len(data) < 40 {
		global.Logger.Error("UploadBinary Invalid data length")
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
		global.Logger.Error("websocket_router.file.FileUploadChunkBinary Session not found", zap.String("sessionID", sessionID))
		c.ToResponse(code.ErrorFileUploadSessionNotFound.WithData(map[string]string{
			"sessionID": sessionID,
		}))
		return
	}

	session := binarySession.(*FileUploadBinaryChunkSession)

	if session == nil {
		global.Logger.Error("websocket_router.file.FileUploadChunkBinary Session not found", zap.String("sessionID", sessionID))
		c.ToResponse(code.ErrorFileUploadSessionNotFound.WithData(map[string]string{
			"sessionID": sessionID,
		}))
		return
	}

	// 计算写入偏移量并写入数据
	offset := int64(chunkIndex) * int64(session.ChunkSize)

	if _, err := session.FileHandle.WriteAt(chunkData, offset); err != nil {
		global.Logger.Error("websocket_router.file.FileUploadChunkBinary FileHandle WriteAt err", zap.Error(err))
		c.ToResponse(code.ErrorFileUploadFailed.WithData(map[string]string{
			"sessionID": sessionID,
		}).WithDetails(err.Error()))
		return
	}

	// 更新已上传计数
	session.mu.Lock()
	session.UploadedChunks++
	session.mu.Unlock()

	// 检查是否所有分块都已上传
	if session.UploadedChunks == session.TotalChunks {

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
			global.Logger.Error("FileUploadComplete: failed to close file handle", zap.Error(err))
			c.ToResponse(code.ErrorFileUploadFailed.WithData(map[string]string{
				"sessionID": sessionID,
			}).WithDetails(err.Error()))
			return
		}

		baseUploadDir := global.Config.App.UploadSavePath
		if baseUploadDir == "" {
			baseUploadDir = "storage/uploads"
		}

		// 按月生成存储子目录
		dateDir := time.Now().Format("200601")
		finalDir := filepath.Join(baseUploadDir, dateDir)

		if err := os.MkdirAll(finalDir, 0644); err != nil {
			global.Logger.Error("websocket_router.file.FileUploadComplete MkdirAll err", zap.Error(err))
			c.ToResponse(code.ErrorFileUploadFailed.WithData(map[string]string{
				"sessionID": sessionID,
			}).WithDetails(err.Error()))
			return
		}

		// 生成唯一文件名
		originalName := filepath.Base(session.Path)
		ext := filepath.Ext(originalName)
		nameOnly := strings.TrimSuffix(originalName, ext)

		var finalPath string

		// 尝试生成不冲突的文件名
		for i := 0; i < 10; i++ {
			randStr := util.GetRandomString(8)
			tryFileName := fmt.Sprintf("%s_%s%s", nameOnly, randStr, ext)
			tryPath := filepath.Join(finalDir, tryFileName)

			if _, err := os.Stat(tryPath); os.IsNotExist(err) {
				finalPath = tryPath
				break
			}
		}

		if finalPath == "" {
			global.Logger.Error("websocket_router.file.FileUploadComplete Generate unique filename failed after retries")
			c.ToResponse(code.ErrorFileUploadFailed.WithData(map[string]string{
				"sessionID": sessionID,
			}))
			return
		}

		// 移动文件到最终路径
		if err := os.Rename(session.SavePath, finalPath); err != nil {
			// 跨设备移动失败时尝试复制并删除
			if err := fileurl.CopyFile(session.SavePath, finalPath); err != nil {
				global.Logger.Error("websocket_router.file.FileUploadComplete Move/Copy file err", zap.Error(err))
				c.ToResponse(code.ErrorFileUploadFailed.WithData(map[string]string{
					"sessionID": sessionID,
				}).WithDetails(err.Error()))
				return
			}
			os.Remove(session.SavePath)
		}

		svc := service.New(c.Ctx).WithSF(c.SF)

		// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
		svc.VaultGetOrCreate(session.Vault, c.User.UID)

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

		// 更新或创建文件记录
		_, fileSvc, err := svc.FileUpdateOrCreate(c.User.UID, svcParams, true)

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

		c.ToResponse(code.Success.Reset())
		// 广播文件更新消息
		c.BroadcastResponse(code.Success.Reset().WithData(fileMsg), true, "FileSyncUpdate")
	}
}

// FileDelete 处理文件删除请求。
func FileDelete(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.FileDeleteParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.file.FileDelete.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx).WithSF(c.SF)

	// 获取或创建仓库
	svc.VaultGetOrCreate(params.Vault, c.User.UID)

	// 执行删除逻辑
	fileSvc, err := svc.FileDelete(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorFileDeleteFailed.WithDetails(err.Error()))
		return
	}

	c.ToResponse(code.Success.Reset())

	fileDeleteMessage := convert.StructAssign(fileSvc, &FileDeleteMessage{}).(*FileDeleteMessage)
	// 广播文件删除消息
	c.BroadcastResponse(code.Success.Reset().WithData(fileDeleteMessage), true, "FileSyncDelete")
}

// FileChunkDownload 处理文件分片下载请求。
// 客户端通过此接口请求下载文件,服务端创建下载会话并开始发送分片。
func FileChunkDownload(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.FileGetParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("websocket_router.file.FileChunkDownload.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx).WithSF(c.SF)

	// 获取或创建仓库
	svc.VaultGetOrCreate(params.Vault, c.User.UID)

	// 获取文件信息
	fileSvc, err := svc.FileGet(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorFileGetFailed.WithDetails(err.Error()))
		return
	}

	// 检查文件是否存在于磁盘
	if _, err := os.Stat(fileSvc.SavePath); os.IsNotExist(err) {
		global.Logger.Error("websocket_router.file.FileChunkDownload file not found on disk",
			zap.String("path", fileSvc.SavePath),
			zap.String("pathHash", fileSvc.PathHash))
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
	go sendFileChunks(c, session)
}

// sendFileChunks 执行文件分片发送。
// 在独立的 goroutine 中运行,读取文件并通过 WebSocket 发送二进制分片。
func sendFileChunks(c *app.WebsocketClient, session *FileDownloadChunkSession) {
	// 创建超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

// FileSync 批量检测用户文件是否需要更新。
// 对比客户端和服务端的文件列表，决定哪些文件需要上传、更新或删除。
func FileSync(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.FileSyncParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.file.FileSync.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx).WithSF(c.SF)

	// 获取或创建仓库
	svc.VaultGetOrCreate(params.Vault, c.User.UID)

	// 获取最后一次同步后的变更文件列表
	list, err := svc.FileListByLastTime(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorFileListFailed.WithDetails(err.Error()))
		return
	}

	// 构建客户端文件索引
	var cFiles map[string]service.FileSyncCheckRequestParams = make(map[string]service.FileSyncCheckRequestParams, 0)
	var cFilesKeys map[string]struct{} = make(map[string]struct{}, 0)

	if len(params.Files) > 0 {
		for _, file := range params.Files {
			cFiles[file.PathHash] = file
			cFilesKeys[file.PathHash] = struct{}{}
		}
	}

	var lastTime int64

	// 遍历服务端文件列表进行处理
	for _, file := range list {
		if file.UpdatedTimestamp >= lastTime {
			lastTime = file.UpdatedTimestamp
		}

		if file.Action == "delete" {
			// 服务端已删除，通知客户端删除
			if _, ok := cFiles[file.PathHash]; ok {
				c.ToResponse(code.Success.WithData(file), "FileSyncDelete")
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
						c.ToResponse(code.Success.WithData(fileMessage), "FileSyncUpdate")
					} else {
						// 服务端修改时间比客户端旧, 通知客户端上传文件
						fileNeedUploadMessage := &FileNeedUploadMessage{
							Path: file.Path,
						}
						c.ToResponse(code.Success.WithData(fileNeedUploadMessage), "FileNeedUpload")
					}
				} else {
					// 内容一致, 但修改时间不一致, 通知客户端更新文件修改时间
					fileSyncMtimeMessage := &FileSyncMtimeMessage{
						Path:  file.Path,
						Ctime: file.Ctime,
						Mtime: file.Mtime,
					}
					c.ToResponse(code.Success.WithData(fileSyncMtimeMessage), "FileSyncMtime")
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
				c.ToResponse(code.Success.WithData(fileMessage), "FileSyncUpdate")
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
			FileCheck := convert.StructAssign(&file, &FileNeedUploadMessage{}).(*FileNeedUploadMessage)
			c.ToResponse(code.Success.WithData(FileCheck), "FileNeedUpload")
		}
	}

	message := &FileSyncEndMessage{
		Vault:    params.Vault,
		LastTime: lastTime,
	}
	// 发送同步结束消息
	c.ToResponse(code.Success.WithData(message), "FileSyncEnd")
}

// cleanupSession 清理因为完成或超时而废弃的上传会话。
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

	// 停止超时定时器
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

	// 清理临时文件
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

// startSessionTimeout 启动会话超时定时器。
// 返回一个取消函数，调用后可阻止超时清理的执行。
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
			// 超时触发，清理会话
			global.Logger.Warn("startSessionTimeout: session timeout, cleaning up",
				zap.String("sessionID", sessionID),
				zap.Duration("timeout", timeout))

			// 超时执行清理
			cleanupSession(c, sessionID)
		case <-ctx.Done():
			// 取消定时器
			return
		}
	}()

	return cancel
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
