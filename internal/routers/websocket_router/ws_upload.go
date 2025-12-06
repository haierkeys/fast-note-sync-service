package websocket_router

import (
	"context"
	"os"
	"time"
)

// BinaryChunkSession stores the state of an active upload
type BinaryChunkSession struct {
	ID          string             // 上传会话ID
	Vault       string             // 仓库标识
	Path        string             // 路径
	PathHash    string             // 路径哈希
	ContentHash string             // 内容哈希(可选)
	Ctime       int64              // 创建时间戳
	Mtime       int64              // 修改时间戳
	Size        int64              // 文件总大小
	TotalChunks int64              // 总分块数
	ChunkSize   int64              // 每个分块大小
	SavePath    string             // 临时保存路径
	FileHandle  *os.File           // 文件句柄
	CreatedAt   time.Time          // 会话创建时间
	CancelFunc  context.CancelFunc // 用于取消超时定时器
}

// UploadBinary handles binary chunks
// Protocol: [sessionID (36 bytes)][ChunkIndex (4 bytes BigEndian)][Data...]
