// Package domain 定义领域模型和接口
package domain

import "context"

// NoteRepository 笔记仓储接口
type NoteRepository interface {
	// GetByID 根据ID获取笔记
	GetByID(ctx context.Context, id, uid int64) (*Note, error)

	// GetByPathHash 根据路径哈希获取笔记（排除已删除）
	GetByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*Note, error)

	// GetByPathHashIncludeRecycle 根据路径哈希获取笔记（可选包含回收站）
	GetByPathHashIncludeRecycle(ctx context.Context, pathHash string, vaultID, uid int64, isRecycle bool) (*Note, error)

	// GetAllByPathHash 根据路径哈希获取笔记（包含所有状态）
	GetAllByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*Note, error)

	// GetByPath 根据路径获取笔记
	GetByPath(ctx context.Context, path string, vaultID, uid int64) (*Note, error)

	// Create 创建笔记
	Create(ctx context.Context, note *Note, uid int64) (*Note, error)

	// Update 更新笔记
	Update(ctx context.Context, note *Note, uid int64) (*Note, error)

	// UpdateDelete 更新笔记为删除状态
	UpdateDelete(ctx context.Context, note *Note, uid int64) error

	// UpdateMtime 更新笔记修改时间
	UpdateMtime(ctx context.Context, mtime int64, id, uid int64) error

	// UpdateActionMtime 更新笔记类型并修改时间
	UpdateActionMtime(ctx context.Context, action NoteAction, mtime int64, id, uid int64) error

	// UpdateSnapshot 更新笔记快照
	UpdateSnapshot(ctx context.Context, snapshot, snapshotHash string, version, id, uid int64) error

	// Delete 物理删除笔记
	Delete(ctx context.Context, id, uid int64) error

	// DeletePhysicalByTime 根据时间物理删除已标记删除的笔记
	DeletePhysicalByTime(ctx context.Context, timestamp, uid int64) error

	// DeletePhysicalByTimeAll 根据时间物理删除所有用户的已标记删除的笔记
	DeletePhysicalByTimeAll(ctx context.Context, timestamp int64) error

	// List 分页获取笔记列表
	// searchMode: path(默认), content, regex
	// sortBy: mtime(默认), ctime, path
	// sortOrder: desc(默认), asc
	List(ctx context.Context, vaultID int64, page, pageSize int, uid int64, keyword string, isRecycle bool, searchMode string, searchContent bool, sortBy string, sortOrder string) ([]*Note, error)

	// ListCount 获取笔记数量
	// searchMode: path(默认), content, regex
	ListCount(ctx context.Context, vaultID, uid int64, keyword string, isRecycle bool, searchMode string, searchContent bool) (int64, error)

	// ListByUpdatedTimestamp 根据更新时间戳获取笔记列表
	ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*Note, error)

	// ListContentUnchanged 获取内容未变更的笔记列表
	ListContentUnchanged(ctx context.Context, uid int64) ([]*Note, error)

	// CountSizeSum 获取笔记数量和大小总和
	CountSizeSum(ctx context.Context, vaultID, uid int64) (*CountSizeResult, error)
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

// UserRepository 用户仓储接口
type UserRepository interface {
	// GetByUID 根据UID获取用户
	GetByUID(ctx context.Context, uid int64) (*User, error)

	// GetByEmail 根据邮箱获取用户
	GetByEmail(ctx context.Context, email string) (*User, error)

	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*User, error)

	// Create 创建用户
	Create(ctx context.Context, user *User) (*User, error)

	// UpdatePassword 更新用户密码
	UpdatePassword(ctx context.Context, password string, uid int64) error

	// GetAllUIDs 获取所有用户UID
	GetAllUIDs(ctx context.Context) ([]int64, error)
}

// FileRepository 文件仓储接口
type FileRepository interface {
	// GetByID 根据 ID 获取文件
	GetByID(ctx context.Context, id, uid int64) (*File, error)

	// GetByPathHash 根据路径哈希获取文件
	GetByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*File, error)

	// GetByPath 根据路径获取文件
	GetByPath(ctx context.Context, path string, vaultID, uid int64) (*File, error)

	// GetByPathLike 根据路径后缀获取文件
	GetByPathLike(ctx context.Context, path string, vaultID, uid int64) (*File, error)

	// Create 创建文件
	Create(ctx context.Context, file *File, uid int64) (*File, error)

	// Update 更新文件
	Update(ctx context.Context, file *File, uid int64) (*File, error)

	// UpdateMtime 更新文件修改时间
	UpdateMtime(ctx context.Context, mtime int64, id, uid int64) error

	// Delete 物理删除文件
	Delete(ctx context.Context, id, uid int64) error

	// DeletePhysicalByTime 根据时间物理删除已标记删除的文件
	DeletePhysicalByTime(ctx context.Context, timestamp, uid int64) error

	// DeletePhysicalByTimeAll 根据时间物理删除所有用户的已标记删除的文件
	DeletePhysicalByTimeAll(ctx context.Context, timestamp int64) error

	// List 分页获取文件列表
	List(ctx context.Context, vaultID int64, page, pageSize int, uid int64, keyword string, isRecycle bool, sortBy string, sortOrder string) ([]*File, error)

	// ListCount 获取文件数量
	ListCount(ctx context.Context, vaultID, uid int64, keyword string, isRecycle bool) (int64, error)

	// ListByUpdatedTimestamp 根据更新时间戳获取文件列表
	ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*File, error)

	// ListByMtime 根据修改时间戳获取文件列表
	ListByMtime(ctx context.Context, timestamp, vaultID, uid int64) ([]*File, error)

	// CountSizeSum 获取文件数量和大小总和
	CountSizeSum(ctx context.Context, vaultID, uid int64) (*CountSizeResult, error)
}

// SettingRepository 配置仓储接口
type SettingRepository interface {
	// GetByPathHash 根据路径哈希获取配置
	GetByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*Setting, error)

	// Create 创建配置
	Create(ctx context.Context, setting *Setting, uid int64) (*Setting, error)

	// Update 更新配置
	Update(ctx context.Context, setting *Setting, uid int64) (*Setting, error)

	// UpdateMtime 更新配置修改时间
	UpdateMtime(ctx context.Context, mtime int64, id, uid int64) error

	// Delete 物理删除配置
	Delete(ctx context.Context, id, uid int64) error

	// DeletePhysicalByTime 根据时间物理删除已标记删除的配置
	DeletePhysicalByTime(ctx context.Context, timestamp, uid int64) error

	// DeletePhysicalByTimeAll 根据时间物理删除所有用户的已标记删除的配置
	DeletePhysicalByTimeAll(ctx context.Context, timestamp int64) error

	// ListByUpdatedTimestamp 根据更新时间戳获取配置列表
	ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*Setting, error)
}

// NoteHistoryRepository 笔记历史仓储接口
type NoteHistoryRepository interface {
	// GetByID 根据ID获取历史记录
	GetByID(ctx context.Context, id, uid int64) (*NoteHistory, error)

	// GetByNoteIDAndHash 根据笔记ID和内容哈希获取历史记录
	GetByNoteIDAndHash(ctx context.Context, noteID int64, contentHash string, uid int64) (*NoteHistory, error)

	// Create 创建历史记录
	Create(ctx context.Context, history *NoteHistory, uid int64) (*NoteHistory, error)

	// ListByNoteID 根据笔记ID获取历史记录列表
	ListByNoteID(ctx context.Context, noteID int64, page, pageSize int, uid int64) ([]*NoteHistory, int64, error)

	// GetLatestVersion 获取笔记的最新版本号
	GetLatestVersion(ctx context.Context, noteID, uid int64) (int64, error)

	// Migrate 迁移历史记录（更新 NoteID）
	Migrate(ctx context.Context, oldNoteID, newNoteID, uid int64) error

	// GetNoteIDsWithOldHistory 获取有旧历史记录的笔记ID列表
	// cutoffTime: 截止时间戳（毫秒），返回有早于此时间历史记录的笔记ID
	GetNoteIDsWithOldHistory(ctx context.Context, cutoffTime int64, uid int64) ([]int64, error)

	// DeleteOldVersions 删除旧版本历史记录，保留最近 N 个版本
	// noteID: 笔记ID
	// cutoffTime: 截止时间戳（毫秒），删除早于此时间的记录
	// keepVersions: 保留的最近版本数量
	DeleteOldVersions(ctx context.Context, noteID int64, cutoffTime int64, keepVersions int, uid int64) error

	// Delete 删除指定ID的历史记录
	Delete(ctx context.Context, id, uid int64) error
}
