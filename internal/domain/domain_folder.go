package domain

import "time"

// FolderAction 定义文件夹操作类型
type FolderAction string

const (
	FolderActionCreate FolderAction = "create"
	FolderActionModify FolderAction = "modify"
	FolderActionDelete FolderAction = "delete"
)

// Folder 文件夹领域模型
type Folder struct {
	ID               int64
	VaultID          int64
	Action           FolderAction
	Path             string
	PathHash         string
	Level            int64
	FID              int64
	Ctime            int64
	Mtime            int64
	UpdatedTimestamp int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// IsDeleted 判断文件夹是否已删除
func (f *Folder) IsDeleted() bool {
	return f.Action == FolderActionDelete
}
