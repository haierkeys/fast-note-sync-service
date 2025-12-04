package websocket_router

import (
	"encoding/binary"
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

// UploadSession stores the state of an active upload
type BinaryChunkSession struct {
	UploadID    string
	Filename    string
	Hash        string
	TotalSize   int64
	TotalChunks int
	ChunkSize   int
	TempPath    string
	FileHandle  *os.File
	CreatedAt   time.Time
}

type UploadInitParams struct {
	Filename    string `json:"filename" binding:"required"`
	Hash        string `json:"hash" binding:"required"`
	Size        int64  `json:"size" binding:"required"`
	TotalChunks int    `json:"totalChunks" binding:"required"`
}

type UploadCompleteParams struct {
	UploadID string `json:"uploadId" binding:"required"`
}

// FileChunkUploadInit handles the initialization of a file upload
func FileChunkStart(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &UploadInitParams{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("FileChunkUploadInit BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uploadID := uuid.New().String()
	tempDir := global.Config.App.TempPath
	if tempDir == "" {
		tempDir = "storage/temp"
	}
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		global.Logger.Error("FileChunkUploadInit MkdirAll err", zap.Error(err))
		c.ToResponse(code.ErrorServerInternal)
		return
	}

	tempPath := filepath.Join(tempDir, fmt.Sprintf("%s_%s", uploadID, params.Filename))
	file, err := os.Create(tempPath)
	if err != nil {
		global.Logger.Error("FileChunkUploadInit Create file err", zap.Error(err))
		c.ToResponse(code.ErrorServerInternal)
		return
	}

	session := &BinaryChunkSession{
		UploadID:    uploadID,
		Filename:    params.Filename,
		Hash:        params.Hash,
		TotalSize:   params.Size,
		TotalChunks: params.TotalChunks,
		ChunkSize:   1024 * 1024, // Default 1MB, or calculated
		TempPath:    tempPath,
		FileHandle:  file,
		CreatedAt:   time.Now(),
	}

	c.BinaryMu.Lock()
	c.BinaryChunkSessions[uploadID] = session
	c.BinaryMu.Unlock()

	response := map[string]interface{}{
		"uploadId":  uploadID,
		"chunkSize": session.ChunkSize,
	}
	c.ToResponse(code.Success.WithData(response), "FileChunkUploadInit")
}

// UploadBinary handles binary chunks
// Protocol: [UploadID (36 bytes)][ChunkIndex (4 bytes BigEndian)][Data...]
func FileChunkBinary(c *app.WebsocketClient, data []byte) {
	if len(data) < 40 {
		global.Logger.Error("UploadBinary Invalid data length")
		return
	}

	uploadID := string(data[:36])
	chunkIndex := binary.BigEndian.Uint32(data[36:40])
	chunkData := data[40:]

	c.BinaryMu.Lock()
	binarySession, exists := c.BinaryChunkSessions[uploadID]

	c.BinaryMu.Unlock()
	session := binarySession.(*BinaryChunkSession)

	if !exists {
		global.Logger.Error("UploadBinary Session not found", zap.String("uploadId", uploadID))
		return
	}

	// Write to file
	// Note: For simplicity, we assume sequential upload or use Seek.
	// Using Seek is safer for parallel chunks.
	offset := int64(chunkIndex) * int64(session.ChunkSize)

	if _, err := session.FileHandle.Seek(offset, 0); err != nil {
		global.Logger.Error("UploadBinary Seek err", zap.Error(err))
		return
	}

	if _, err := session.FileHandle.Write(chunkData); err != nil {
		global.Logger.Error("UploadBinary Write err", zap.Error(err))
		return
	}
}

// FileChunkUploadComplete handles the completion of a file upload
func FileChunkEnd(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &UploadCompleteParams{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("FileChunkUploadComplete BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	c.BinaryMu.Lock()
	binarySession, exists := c.BinaryChunkSessions[params.UploadID]
	delete(c.BinaryChunkSessions, params.UploadID) // Remove session
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

	finalPath := filepath.Join(finalDir, fmt.Sprintf("%d_%s", time.Now().Unix(), session.Filename))

	if err := os.Rename(session.TempPath, finalPath); err != nil {
		// Try copy if rename fails (different volume)
		if err := copyFile(session.TempPath, finalPath); err != nil {
			global.Logger.Error("FileChunkUploadComplete Move file err", zap.Error(err))
			c.ToResponse(code.ErrorServerInternal)
			return
		}
		os.Remove(session.TempPath)
	}

	response := map[string]interface{}{
		"path": finalPath,
		"url":  "/uploads/" + filepath.Base(finalPath), // Mock URL
	}
	c.ToResponse(code.Success.WithData(response), "FileChunkUploadComplete")
}

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
