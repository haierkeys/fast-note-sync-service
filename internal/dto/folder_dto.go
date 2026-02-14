package dto

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

// FolderGetRequest Request parameters for retrieving a folder
// 获取文件夹的请求参数
type FolderGetRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"` // Vault name // 保险库名称
	Path     string `json:"path" form:"path"`                      // Folder path // 文件夹路径
	PathHash string `json:"pathHash" form:"pathHash"`              // Path hash // 路径哈希
}

// FolderListRequest Request parameters for retrieving a folder list
// 获取文件夹列表的请求参数
type FolderListRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"` // Vault name // 保险库名称
	Path     string `json:"path" form:"path"`                      // Folder path // 文件夹路径
	PathHash string `json:"pathHash" form:"pathHash"`              // Path hash // 路径哈希
}

// FolderCreateRequest Request parameters for creating a folder
// 创建文件夹请求参数
type FolderCreateRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"` // Vault name // 保险库名称
	Path     string `json:"path" form:"path" binding:"required"`   // Folder path // 文件夹路径
	PathHash string `json:"pathHash" form:"pathHash"`              // Path hash // 路径哈希
}

// FolderDeleteRequest Request parameters for deleting a folder
// 删除文件夹请求参数
type FolderDeleteRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"` // Vault name // 保险库名称
	Path     string `json:"path" form:"path" binding:"required"`   // Folder path // 文件夹路径
	PathHash string `json:"pathHash" form:"pathHash"`              // Path hash // 路径哈希
}

// FolderSyncCheckRequest Parameters for single record check during synchronization
// 同步检查单条记录的参数
type FolderSyncCheckRequest struct {
	Path     string `json:"path" form:"path"`                            // Folder path // 文件夹路径
	PathHash string `json:"pathHash" form:"pathHash" binding:"required"` // Path hash // 路径哈希
	Mtime    int64  `json:"mtime" form:"mtime" binding:"required"`       // Modification timestamp // 修改时间戳
}

// FolderSyncDelFolder Parameters for deleting/missing folder during synchronization
// 同步删除/缺失文件夹的参数
type FolderSyncDelFolder struct {
	Path     string `json:"path" form:"path" binding:"required"`         // Folder path // 文件夹路径
	PathHash string `json:"pathHash" form:"pathHash" binding:"required"` // Path hash // 路径哈希
}

// FolderSyncRequest Synchronization request body
// 同步请求主体
type FolderSyncRequest struct {
	Vault          string                   `json:"vault" form:"vault" binding:"required"` // Vault name // 保险库名称
	LastTime       int64                    `json:"lastTime" form:"lastTime"`              // Last sync time // 最后同步时间
	Folders        []FolderSyncCheckRequest `json:"folders" form:"folders"`                // Folders to check // 待检查文件夹列表
	DelFolders     []FolderSyncDelFolder    `json:"delFolders" form:"delFolders"`          // Folders to delete // 待删除文件夹列表
	MissingFolders []FolderSyncDelFolder    `json:"missingFolders" form:"missingFolders"`  // Missing folders // 缺失文件夹列表
}

// FolderRenameRequest Request parameters for folder renaming
// 文件夹重命名请求参数
type FolderRenameRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required"`             // Vault name // 保险库名称
	Path        string `json:"path" form:"path" binding:"required"`               // Current path // 当前路径
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"`       // Current path hash // 当前路径哈希
	OldPath     string `json:"oldPath" form:"oldPath" binding:"required"`         // Old path // 旧路径
	OldPathHash string `json:"oldPathHash" form:"oldPathHash" binding:"required"` // Old path hash // 旧路径哈希
}

// FolderContentRequest Request parameters for retrieving folder contents
// 获取文件夹内容的请求参数
type FolderContentRequest struct {
	Vault     string `json:"vault" form:"vault" binding:"required"` // Vault name // 保险库名称
	Path      string `json:"path" form:"path"`                      // Folder path // 文件夹路径
	PathHash  string `json:"pathHash" form:"pathHash"`              // Path hash // 路径哈希
	SortBy    string `json:"sortBy" form:"sortBy"`                  // Sort by field // 排序字段
	SortOrder string `json:"sortOrder" form:"sortOrder"`            // Sort order // 排序顺序
}

// FolderTreeRequest Request parameters for retrieving the folder tree
// 获取文件夹树的请求参数
type FolderTreeRequest struct {
	Vault string `json:"vault" form:"vault" binding:"required"` // Vault name // 保险库名称
	Depth int    `json:"depth" form:"depth"`                    // Tree depth // 树深度
}

// ---------------- DTO / Response ----------------

// FolderDTO Folder data transfer object
// FolderDTO 文件夹数据传输对象
type FolderDTO struct {
	ID               int64      `json:"-" form:"id"`                      // Folder ID // 文件夹 ID
	Action           string     `json:"-" form:"action"`                  // Action // 动作
	Path             string     `json:"path" form:"path"`                 // Folder path // 文件夹路径
	PathHash         string     `json:"pathHash" form:"pathHash"`         // Path hash // 路径哈希值
	Level            int64      `json:"-" form:"level"`                   // Level // 层级
	FID              int64      `json:"-" form:"fid"`                     // Parent ID // 父 ID
	Ctime            int64      `json:"ctime" form:"ctime"`               // Creation timestamp // 创建时间戳
	Mtime            int64      `json:"mtime" form:"mtime"`               // Modification timestamp // 修改时间戳
	UpdatedTimestamp int64      `json:"lastTime" form:"updatedTimestamp"` // Record update timestamp // 记录更新时间戳
	UpdatedAt        timex.Time `json:"updatedAt"`                        // Updated at time // 更新时间
	CreatedAt        timex.Time `json:"createdAt"`                        // Created at time // 创建时间
}

// FolderTreeNode Folder tree node
// FolderTreeNode 文件夹树节点
type FolderTreeNode struct {
	Path      string            `json:"path"`               // Node path // 节点路径
	Name      string            `json:"name"`               // Node name // 节点名称
	NoteCount int               `json:"noteCount"`          // Note count // 笔记数量
	FileCount int               `json:"fileCount"`          // File count // 文件数量
	Children  []*FolderTreeNode `json:"children,omitempty"` // Child nodes // 子节点
}

// FolderTreeResponse Folder tree response structure
// FolderTreeResponse 文件夹树响应结构体
type FolderTreeResponse struct {
	Folders       []*FolderTreeNode `json:"folders"`       // Folder tree // 文件夹树
	RootNoteCount int               `json:"rootNoteCount"` // Note count in root // 根目录中的笔记数量
	RootFileCount int               `json:"rootFileCount"` // File count in root // 根目录中的文件数量
}
