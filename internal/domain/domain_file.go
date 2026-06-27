package domain

import "time"

// FileAction 定义文件操作类型
type FileAction string

const (
	FileActionCreate FileAction = "create"
	FileActionModify FileAction = "modify"
	FileActionDelete FileAction = "delete"
)

// File 文件领域模型
type File struct {
	ID               int64
	VaultID          int64
	Action           FileAction
	FID              int64
	Path             string
	PathHash         string
	ContentHash      string
	SavePath         string
	Rename           int64
	RenamedToID      int64  // 重命名目标记录ID: 当 Rename=1 时, 指向被创建的新记录
	Size             int64
	Ctime            int64
	Mtime            int64
	UpdatedTimestamp int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// IsDeleted 判断文件是否已删除
func (f *File) IsDeleted() bool {
	return f.Action == FileActionDelete
}

// IsCreated 判断文件是否为新建
func (f *File) IsCreated() bool {
	return f.Action == FileActionCreate
}

// IsModified 判断文件是否已修改
func (f *File) IsModified() bool {
	return f.Action == FileActionModify
}
