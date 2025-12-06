package websocket_router

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"go.uber.org/zap"
)

// BinaryChunkSession stores the state of an active upload
type BinaryChunkSession struct {
	ID          string   // 上传会话ID
	Vault       string   // 仓库标识
	Path        string   // 路径
	PathHash    string   // 路径哈希
	ContentHash string   // 内容哈希（可选）
	Ctime       int64    // 创建时间戳
	Mtime       int64    // 修改时间戳
	Size        int64    // 文件总大小
	TotalChunks int64    // 总分块数
	ChunkSize   int64    // 每个分块大小
	SavePath    string   // 临时保存路径
	FileHandle  *os.File // 文件句柄
}

// FileChunkStart handles the initialization of a file upload
func FileChunkStart(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &UploadInitParams{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("FileChunkUploadInit BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	sessionID := uuid.New().String()
	tempDir := global.Config.App.TempPath
	if tempDir == "" {
		tempDir = "storage/temp"
	}
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		global.Logger.Error("FileChunkUploadInit MkdirAll err", zap.Error(err))
		c.ToResponse(code.ErrorServerInternal)
		return
	}

	tempPath := filepath.Join(tempDir, fmt.Sprintf("%s_%s", sessionID, params.Filename))
	file, err := os.Create(tempPath)
	if err != nil {
		global.Logger.Error("FileChunkUploadInit Create file err", zap.Error(err))
		c.ToResponse(code.ErrorServerInternal)
		return
	}

	session := &BinaryChunkSession{
		ID:          sessionID,
		Path:        params.Filename,
		ContentHash: params.Hash,
		Size:        params.Size,
		TotalChunks: params.TotalChunks,
		ChunkSize:   1024 * 1024, // Default 1MB, or calculated
		SavePath:    tempPath,
		FileHandle:  file,
		Ctime:       time.Now().Unix(),
	}

	c.BinaryMu.Lock()
	c.BinaryChunkSessions[sessionID] = session
	c.BinaryMu.Unlock()

	response := map[string]interface{}{
		"sessionID": sessionID,
		"chunkSize": session.ChunkSize,
	}
	c.ToResponse(code.Success.WithData(response), "FileChunkUploadInit")
}

// UploadBinary handles binary chunks
// Protocol: [sessionID (36 bytes)][ChunkIndex (4 bytes BigEndian)][Data...]

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
