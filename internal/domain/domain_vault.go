package domain

import "time"

// Vault 仓库领域模型
type Vault struct {
	ID        int64
	UID       int64
	Name      string
	NoteCount int64
	NoteSize  int64
	FileCount int64
	FileSize  int64
	IsDeleted bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsEmpty 判断仓库是否为空
func (v *Vault) IsEmpty() bool {
	return v.NoteCount == 0 && v.FileCount == 0
}

// TotalSize 获取仓库总大小
func (v *Vault) TotalSize() int64 {
	return v.NoteSize + v.FileSize
}

// TotalCount 获取仓库总数量
func (v *Vault) TotalCount() int64 {
	return v.NoteCount + v.FileCount
}
