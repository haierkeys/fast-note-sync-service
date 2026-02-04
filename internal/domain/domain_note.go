// Package domain 定义领域模型和接口
package domain

import "time"

// NoteAction 定义笔记操作类型
type NoteAction string

const (
	NoteActionCreate NoteAction = "create"
	NoteActionModify NoteAction = "modify"
	NoteActionDelete NoteAction = "delete"
)

// Note 笔记领域模型
type Note struct {
	ID                      int64
	VaultID                 int64
	Action                  NoteAction
	Rename                  int64
	FID                     int64
	Path                    string
	PathHash                string
	Content                 string
	ContentHash             string
	ContentLastSnapshot     string
	ContentLastSnapshotHash string
	Version                 int64
	ClientName              string
	Size                    int64
	Ctime                   int64
	Mtime                   int64
	UpdatedTimestamp        int64
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

// CountSizeResult 统计结果
type CountSizeResult struct {
	Count int64
	Size  int64
}

// IsDeleted 判断笔记是否已删除
func (n *Note) IsDeleted() bool {
	return n.Action == NoteActionDelete
}

// IsCreated 判断笔记是否为新建
func (n *Note) IsCreated() bool {
	return n.Action == NoteActionCreate
}

// IsModified 判断笔记是否已修改
func (n *Note) IsModified() bool {
	return n.Action == NoteActionModify
}
