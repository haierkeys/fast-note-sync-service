// Package dto Defines data transfer objects (request parameters and response structs)
// Package dto 定义数据传输对象（请求参数和响应结构体）
package dto

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

// FileUpdateCheckRequest Client request parameters for checking if updates are needed
// 客户端用于检查是否需要更新的请求参数
type FileUpdateCheckRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required" example:"MyVault"`        // Vault name // 保险库名称
	Path        string `json:"path" form:"path" binding:"required" example:"Image.png"`        // File path // 文件路径
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required" example:"fhash123"` // Path hash // 路径哈希
	ContentHash string `json:"contentHash" form:"contentHash" binding:"" example:"chash456"`   // Content hash // 内容哈希
	Size        int64  `json:"size" form:"size" binding:"" example:"1024"`                     // File size // 文件大小
	Ctime       int64  `json:"ctime" form:"ctime" binding:"required" example:"1700000000"`     // Creation timestamp // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime" binding:"required" example:"1700000000"`     // Modification timestamp // 修改时间戳
}

// FileUpdateRequest Request parameters for creating or modifying a file
// 用于创建或修改文件的请求参数
type FileUpdateRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required" example:"MyVault"`          // Vault name // 保险库名称
	Path        string `json:"path" form:"path" binding:"required" example:"Image.png"`          // File path // 文件路径
	PathHash    string `json:"pathHash" form:"pathHash" example:"fhash123"`                      // Path hash // 路径哈希
	ContentHash string `json:"contentHash" form:"contentHash" binding:"" example:"chash456"`     // Content hash // 内容哈希
	SavePath    string `json:"savePath" form:"savePath" binding:"" example:"/uploads/Image.png"` // Save path on server // 服务器保存路径
	Size        int64  `json:"size" form:"size" example:"1024"`                                  // File size // 文件大小
	Ctime       int64  `json:"ctime" form:"ctime" example:"1700000000"`                          // Creation timestamp // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime" example:"1700000000"`                          // Modification timestamp // 修改时间戳
}

// FileDeleteRequest Parameters required for deleting a file
// 删除文件所需参数
type FileDeleteRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required" example:"MyVault"`        // Vault name // 保险库名称
	Path     string `json:"path" form:"path" binding:"required" example:"Image.png"`        // File path // 文件路径
	PathHash string `json:"pathHash" form:"pathHash" binding:"required" example:"fhash123"` // Path hash // 路径哈希
}

// FileRestoreRequest parameters for restoring a file
// FileRestoreRequest 恢复文件请求参数
type FileRestoreRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required" example:"MyVault"` // Vault name // 保险库名称
	Path     string `json:"path" form:"path" binding:"required" example:"Image.png"` // File path // 文件路径
	PathHash string `json:"pathHash" form:"pathHash" example:"fhash123"`             // Path hash // 路径哈希
}

// FileSyncCheckRequest/ Parameters for checking synchronization of a single record
// 同步检查单条记录的参数
type FileSyncCheckRequest struct {
	Path        string `json:"path" form:"path" example:"Image.png"`                           // File path // 文件路径
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required" example:"fhash123"` // Path hash // 路径哈希
	ContentHash string `json:"contentHash" form:"contentHash" binding:"" example:"chash456"`   // Content hash // 内容哈希
	Mtime       int64  `json:"mtime" form:"mtime" binding:"required" example:"1700000000"`     // Modification timestamp // 修改时间戳
	Ctime       int64  `json:"ctime" form:"ctime" example:"1700000000"`                        // Creation timestamp // 创建时间戳
	Size        int64  `json:"size" form:"size" example:"1024"`                                // File size // 文件大小
}

// FileSyncDelFile parameters for deleting a file during sync
// 同步删除文件参数
type FileSyncDelFile struct {
	Path     string `json:"path" form:"path" binding:"required" example:"DeletedFile.png"`   // File path // 文件路径
	PathHash string `json:"pathHash" form:"pathHash" binding:"required" example:"dfhash789"` // Path hash // 路径哈希
}

// FileSyncRequest Synchronization request body
// 同步请求主体
type FileSyncRequest struct {
	Context      string                 `json:"context" form:"context" binding:"required" example:"task123"` // Context // 上下文
	Vault        string                 `json:"vault" form:"vault" binding:"required" example:"MyVault"`     // Vault name // 保险库名称
	LastTime     int64                  `json:"lastTime" form:"lastTime" example:"1700000000"`               // Last sync time // 最后同步时间
	Files        []FileSyncCheckRequest `json:"files" form:"files"`                                          // Files to check // 待检查文件列表
	DelFiles     []FileSyncDelFile      `json:"delFiles" form:"delFiles"`                                    // Files to delete // 待删除文件列表
	MissingFiles []FileSyncDelFile      `json:"missingFiles" form:"missingFiles"`                            // Missing files // 缺失文件列表
}

// FileUploadCompleteRequest Parameters for file upload completion
// 文件上传完成参数
type FileUploadCompleteRequest struct {
	SessionID string `json:"sessionId" binding:"required" example:"sess_123456"` // Upload session ID // 上传会话 ID
}

// FileGetRequest Request parameters for retrieving a single file
// 用于获取单条文件的请求参数
type FileGetRequest struct {
	Vault     string `json:"vault" form:"vault" binding:"required" example:"MyVault"` // Vault name // 保险库名称
	Path      string `json:"path" form:"path" binding:"required" example:"Image.png"` // File path // 文件路径
	PathHash  string `json:"pathHash" form:"pathHash" example:"fhash123"`             // Path hash // 路径哈希
	IsRecycle bool   `json:"isRecycle" form:"isRecycle" example:"false"`              // Is in recycle bin // 是否在回收站
}

// FileListRequest Pagination parameters for retrieving the file list
// 获取文件列表的分页参数
type FileListRequest struct {
	Vault     string `json:"vault" form:"vault" binding:"required" example:"MyVault"` // Vault name // 保险库名称
	Keyword   string `json:"keyword" form:"keyword" example:"vacation"`               // Search keyword // 搜索关键词
	IsRecycle bool   `json:"isRecycle" form:"isRecycle" example:"false"`              // Is in recycle bin // 是否在回收站
	SortBy    string `json:"sortBy" form:"sortBy" example:"mtime"`                    // Sort by field // 排序字段
	SortOrder string `json:"sortOrder" form:"sortOrder" example:"desc"`               // Sort order // 排序顺序
}

// FileRenameRequest Parameters required for renaming a file
// 重命名文件所需参数
type FileRenameRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required" example:"MyVault"`               // Vault name // 保险库名称
	Path        string `json:"path" form:"path" binding:"required" example:"NewImage.png"`            // Current path // 当前路径
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required" example:"nfhash123"`       // Current path hash // 当前路径哈希
	OldPath     string `json:"oldPath" form:"oldPath" binding:"required" example:"OldImage.png"`      // Old path // 旧路径
	OldPathHash string `json:"oldPathHash" form:"oldPathHash" binding:"required" example:"ofhash456"` // Old path hash // 旧路径哈希
}

// ---------------- DTO / Response ----------------

// FileDTO File Data Transfer Object
// FileDTO 文件数据传输对象
type FileDTO struct {
	ID               int64      `json:"-"`                                // File ID // 文件 ID
	Action           string     `json:"-"`                                // Action // 动作
	Path             string     `json:"path" form:"path"`                 // File path // 文件路径
	PathHash         string     `json:"pathHash" form:"pathHash"`         // Path hash // 路径哈希
	ContentHash      string     `json:"contentHash" form:"contentHash"`   // Content hash // 内容哈希
	SavePath         string     `json:"-"`                                // Internal save path // 内部保存路径
	Size             int64      `json:"size" form:"size"`                 // File size // 文件大小
	Ctime            int64      `json:"ctime" form:"ctime"`               // Creation timestamp // 创建时间戳
	Mtime            int64      `json:"mtime" form:"mtime"`               // Modification timestamp // 修改时间戳
	UpdatedTimestamp int64      `json:"lastTime" form:"updatedTimestamp"` // Updated timestamp // 更新时间戳
	UpdatedAt        timex.Time `json:"updatedAt"`                        // Updated at time // 更新时间
	CreatedAt        timex.Time `json:"createdAt"`                        // Created at time // 创建时间
}
