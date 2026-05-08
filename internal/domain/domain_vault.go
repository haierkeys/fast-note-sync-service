package domain

import (
	"context"
	"time"
)

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

// VaultRepository 仓库仓储接口
type VaultRepository interface {
	// GetByID 根据ID获取仓库
	GetByID(ctx context.Context, id, uid int64) (*Vault, error)

	// GetByName 根据名称获取仓库
	GetByName(ctx context.Context, name string, uid int64) (*Vault, error)

	// Create 创建仓库
	Create(ctx context.Context, vault *Vault, uid int64) (*Vault, error)

	// Update 更新仓库
	Update(ctx context.Context, vault *Vault, uid int64) error

	// UpdateNoteCountSize 更新仓库的笔记数量和大小
	UpdateNoteCountSize(ctx context.Context, noteSize, noteCount, vaultID, uid int64) error

	// UpdateFileCountSize 更新仓库的文件数量和大小
	UpdateFileCountSize(ctx context.Context, fileSize, fileCount, vaultID, uid int64) error

	// List 获取仓库列表
	List(ctx context.Context, uid int64) ([]*Vault, error)

	// Delete 删除仓库（软删除）
	Delete(ctx context.Context, id, uid int64) error
}
