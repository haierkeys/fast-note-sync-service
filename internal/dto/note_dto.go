// Package dto Defines data transfer objects (request parameters and response structs)
// Package dto 定义数据传输对象（请求参数和响应结构体）
package dto

import (
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// NoteUpdateCheckRequest Client request parameters for checking if updates are needed
// 客户端用于检查是否需要更新的请求参数
type NoteUpdateCheckRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required"`       // Vault name // 保险库名称
	Path        string `json:"path" form:"path" binding:"required"`         // Note path // 笔记路径
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"` // Path hash // 路径哈希
	ContentHash string `json:"contentHash" form:"contentHash" binding:""`   // Content hash // 内容哈希
	Ctime       int64  `json:"ctime" form:"ctime" binding:"required"`       // Creation timestamp // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime" binding:"required"`       // Modification timestamp // 修改时间戳
}

// NoteModifyOrCreateRequest Request parameters for creating or modifying a note
// 用于创建或修改笔记的请求参数
type NoteModifyOrCreateRequest struct {
	Vault           string `json:"vault" form:"vault" binding:"required"`     // Vault name // 保险库名称
	Path            string `json:"path" form:"path" binding:"required"`       // Note path // 笔记路径
	PathHash        string `json:"pathHash" form:"pathHash"`                  // Path hash // 路径哈希
	SrcPath         string `json:"srcPath" form:"srcPath"`                    // Source path (for rename) // 源路径（用于重命名）
	SrcPathHash     string `json:"srcPathHash" form:"srcPathHash"`            // Source path hash // 源路径哈希
	BaseHash        string `json:"baseHash" form:"baseHash" binding:""`       // Base hash for sync // 同步基准哈希
	BaseHashMissing bool   `json:"baseHashMissing" form:"baseHashMissing"`    // Marks if baseHash is unavailable // 标记基准哈希是否缺失
	Content         string `json:"content" form:"content" binding:""`         // Note content // 笔记内容
	ContentHash     string `json:"contentHash" form:"contentHash" binding:""` // Content hash // 内容哈希
	Ctime           int64  `json:"ctime" form:"ctime"`                        // Creation timestamp // 创建时间戳
	Mtime           int64  `json:"mtime" form:"mtime"`                        // Modification timestamp // 修改时间戳
	OldPath         string `json:"oldPath" form:"oldPath"`                    // Old path // 旧路径
	OldPathHash     string `json:"oldPathHash" form:"oldPathHash"`            // Old path hash // 旧路径哈希
	CreateOnly      bool   `json:"createOnly" form:"createOnly"`              // If true, fail if note already exists // 如果为 true，笔记已存在则失败
}

// ContentModifyRequest Request parameters for modifying content only
// 专用于只修改内容的请求参数
type ContentModifyRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required"`             // Vault name // 保险库名称
	Path        string `json:"path" form:"path" binding:"required"`               // Note path // 笔记路径
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"`       // Path hash // 路径哈希
	Content     string `json:"content" form:"content" binding:"required"`         // Note content // 笔记内容
	ContentHash string `json:"contentHash" form:"contentHash" binding:"required"` // Content hash // 内容哈希
	Ctime       int64  `json:"ctime" form:"ctime" binding:"required"`             // Creation timestamp // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime" binding:"required"`             // Modification timestamp // 修改时间戳
}

// NoteDeleteRequest Parameters required for deleting a note
// 删除笔记所需参数
type NoteDeleteRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"` // Vault name // 保险库名称
	Path     string `json:"path" form:"path" binding:"required"`   // Note path // 笔记路径
	PathHash string `json:"pathHash" form:"pathHash"`              // Path hash // 路径哈希
}

// NoteRestoreRequest parameters for restoring a note
// NoteRestoreRequest 恢复笔记请求参数
type NoteRestoreRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"` // Vault name // 保险库名称
	Path     string `json:"path" form:"path" binding:"required"`   // Note path // 笔记路径
	PathHash string `json:"pathHash" form:"pathHash"`              // Path hash // 路径哈希
}

// NotePatchFrontmatterRequest parameters for patching note frontmatter
// NotePatchFrontmatterRequest 修改笔记 Frontmatter 请求参数
type NotePatchFrontmatterRequest struct {
	Vault    string                 `json:"vault" form:"vault" binding:"required"` // Vault name // 保险库名称
	Path     string                 `json:"path" form:"path" binding:"required"`   // Note path // 笔记路径
	PathHash string                 `json:"pathHash" form:"pathHash"`              // Path hash // 路径哈希
	Updates  map[string]interface{} `json:"updates"`                               // Fields to update // 待更新字段
	Remove   []string               `json:"remove"`                                // Fields to remove // 待移除字段
}

// NoteAppendRequest parameters for appending content to a note
// NoteAppendRequest 追加笔记内容请求参数
type NoteAppendRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`     // Vault name // 保险库名称
	Path     string `json:"path" form:"path" binding:"required"`       // Note path // 笔记路径
	PathHash string `json:"pathHash" form:"pathHash"`                  // Path hash // 路径哈希
	Content  string `json:"content" form:"content" binding:"required"` // Content to append // 追加内容
}

// NotePrependRequest parameters for prepending content to a note
// NotePrependRequest 在笔记头部添加内容请求参数
type NotePrependRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`     // Vault name // 保险库名称
	Path     string `json:"path" form:"path" binding:"required"`       // Note path // 笔记路径
	PathHash string `json:"pathHash" form:"pathHash"`                  // Path hash // 路径哈希
	Content  string `json:"content" form:"content" binding:"required"` // Content to prepend // 头部添加内容
}

// NoteReplaceRequest parameters for find/replace in a note
// NoteReplaceRequest 笔记查找替换请求参数
type NoteReplaceRequest struct {
	Vault         string `json:"vault" form:"vault" binding:"required"` // Vault name // 保险库名称
	Path          string `json:"path" form:"path" binding:"required"`   // Note path // 笔记路径
	PathHash      string `json:"pathHash" form:"pathHash"`              // Path hash // 路径哈希
	Find          string `json:"find" form:"find" binding:"required"`   // String to find // 查找内容
	Replace       string `json:"replace" form:"replace"`                // String to replace with // 替换内容
	Regex         bool   `json:"regex" form:"regex"`                    // Use regex // 使用正则
	All           bool   `json:"all" form:"all"`                        // Replace all matches // 替换所有
	FailIfNoMatch bool   `json:"failIfNoMatch" form:"failIfNoMatch"`    // Fail if no match found // 若无匹配则失败
}

// NoteMoveRequest parameters for moving a note
// NoteMoveRequest 移动笔记请求参数
type NoteMoveRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required"`             // Vault name // 保险库名称
	Path        string `json:"path" form:"path" binding:"required"`               // Current path // 当前路径
	PathHash    string `json:"pathHash" form:"pathHash"`                          // Current path hash // 当前路径哈希
	Destination string `json:"destination" form:"destination" binding:"required"` // Destination path // 目标路径
	Overwrite   bool   `json:"overwrite" form:"overwrite"`                        // Overwrite existing // 覆盖现有
}

// NoteLinkQueryRequest parameters for backlinks/outlinks query
// NoteLinkQueryRequest 反向链接/出链查询请求参数
type NoteLinkQueryRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"` // Vault name // 保险库名称
	Path     string `json:"path" form:"path" binding:"required"`   // Note path // 笔记路径
	PathHash string `json:"pathHash" form:"pathHash"`              // Path hash // 路径哈希
}

// NoteSyncCheckRequest Parameters for checking synchronization of a single record
// 同步检查单条记录的参数
type NoteSyncCheckRequest struct {
	Path        string `json:"path" form:"path"`                            // Note path // 笔记路径
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"` // Path hash // 路径哈希
	ContentHash string `json:"contentHash" form:"contentHash" binding:""`   // Content hash // 内容哈希
	Mtime       int64  `json:"mtime" form:"mtime" binding:"required"`       // Modification timestamp // 修改时间戳
}

// NoteSyncDelNote parameters for deleting a note during sync
// 同步删除笔记参数
type NoteSyncDelNote struct {
	Path     string `json:"path" form:"path" binding:"required"`         // Note path // 笔记路径
	PathHash string `json:"pathHash" form:"pathHash" binding:"required"` // Path hash // 路径哈希
}

// NoteSyncRequest Synchronization request body
// 同步请求主体
type NoteSyncRequest struct {
	Vault        string                 `json:"vault" form:"vault" binding:"required"` // Vault name // 保险库名称
	LastTime     int64                  `json:"lastTime" form:"lastTime"`              // Last sync time // 最后同步时间
	Notes        []NoteSyncCheckRequest `json:"notes" form:"notes"`                    // Notes to check // 待检查笔记列表
	DelNotes     []NoteSyncDelNote      `json:"delNotes" form:"delNotes"`              // Notes to delete // 待删除笔记列表
	MissingNotes []NoteSyncDelNote      `json:"missingNotes" form:"missingNotes"`      // Missing notes // 缺失笔记列表
}

// ModifyMtimeFilesRequest Request for querying modified files by mtime
// 用于按 mtime 查询修改文件
type ModifyMtimeFilesRequest struct {
	Vault string `json:"vault" form:"vault" binding:"required"` // Vault name // 保险库名称
	Mtime int64  `json:"mtime" form:"mtime"`                    // Threshold modification timestamp // 修改时间戳阈值
}

// NoteGetRequest Request parameters for retrieving a single note
// 用于获取单条笔记的请求参数
type NoteGetRequest struct {
	Vault     string `json:"vault" form:"vault" binding:"required"` // Vault name // 保险库名称
	Path      string `json:"path" form:"path" binding:"required"`   // Note path // 笔记路径
	PathHash  string `json:"pathHash" form:"pathHash"`              // Path hash // 路径哈希
	IsRecycle bool   `json:"isRecycle" form:"isRecycle"`            // Is in recycle bin // 是否在回收站
}

// NoteRenameRequest Parameters required for renaming a note
// 重命名笔记所需参数
type NoteRenameRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required"`             // Vault name // 保险库名称
	Path        string `json:"path" form:"path" binding:"required"`               // Current path // 当前路径
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"`       // Current path hash // 当前路径哈希
	OldPath     string `json:"oldPath" form:"oldPath" binding:"required"`         // Old path // 旧路径
	OldPathHash string `json:"oldPathHash" form:"oldPathHash" binding:"required"` // Old path hash // 旧路径哈希
}

// NoteListRequest Pagination parameters for retrieving the note list
// 获取笔记列表的分页参数
type NoteListRequest struct {
	Vault         string `json:"vault" form:"vault" binding:"required"` // Vault name // 保险库名称
	Keyword       string `json:"keyword" form:"keyword"`                // Search keyword // 搜索关键词
	IsRecycle     bool   `json:"isRecycle" form:"isRecycle"`            // Is in recycle bin // 是否在回收站
	SearchMode    string `json:"searchMode" form:"searchMode"`          // Search mode (path, content, regex) // 搜索模式（路径、内容、正则）
	SearchContent bool   `json:"searchContent" form:"searchContent"`    // Whether to search content // 是否搜索内容
	SortBy        string `json:"sortBy" form:"sortBy"`                  // Sort by field // 排序字段
	SortOrder     string `json:"sortOrder" form:"sortOrder"`            // Sort order // 排序顺序
}

// NoteHistoryListRequest Note history list request parameters
// 笔记历史列表请求参数
type NoteHistoryListRequest struct {
	Vault     string `json:"vault" form:"vault" binding:"required"` // Vault name // 保险库名称
	Path      string `json:"path" form:"path" binding:"required"`   // Note path // 笔记路径
	PathHash  string `json:"pathHash" form:"pathHash"`              // Path hash // 路径哈希
	IsRecycle bool   `json:"isRecycle" form:"isRecycle"`            // Is in recycle bin // 是否在回收站
}

// NoteHistoryRestoreRequest Request parameters for restoring a historical version
// 历史版本恢复请求参数
type NoteHistoryRestoreRequest struct {
	Vault     string `json:"vault" form:"vault" binding:"required"`         // Vault name // 保险库名称
	HistoryID int64  `json:"historyId" form:"historyId" binding:"required"` // History version ID // 历史版本 ID
}

// ---------------- DTO / Response ----------------

// NoteDTO Note data transfer object
// NoteDTO 笔记数据传输对象
type NoteDTO struct {
	ID               int64      `json:"-" form:"id"`                    // Note ID // 笔记 ID
	Action           string     `json:"-" form:"action"`                // Action // 动作
	Path             string     `json:"path" form:"path"`               // Note path // 笔记路径
	PathHash         string     `json:"pathHash" form:"pathHash"`       // Path hash // 路径哈希
	Content          string     `json:"content" form:"content"`         // Note content // 笔记内容
	ContentHash      string     `json:"contentHash" form:"contentHash"` // Content hash // 内容哈希
	Version          int64      `json:"version" form:"version"`         // Version number // 版本号
	Ctime            int64      `json:"ctime" form:"ctime"`             // Creation timestamp // 创建时间戳
	Mtime            int64      `json:"mtime" form:"mtime"`             // Modification timestamp // 修改时间戳
	Size             int64      `json:"size" form:"size"`               // Note size // 笔记大小
	UpdatedTimestamp int64      `json:"lastTime"`                       // Record update timestamp // 记录更新时间戳
	UpdatedAt        timex.Time `json:"updatedAt"`                      // Updated at time // 更新时间
	CreatedAt        timex.Time `json:"createdAt"`                      // Created at time // 创建时间
}

// NoteNoContentDTO Note DTO without content
// NoteNoContentDTO 不包含内容的笔记 DTO
type NoteNoContentDTO struct {
	ID               int64      `json:"-" form:"id"`                      // Note ID // 笔记 ID
	Action           string     `json:"-" form:"action"`                  // Action // 动作
	Path             string     `json:"path" form:"path"`                 // Note path // 笔记路径
	PathHash         string     `json:"pathHash" form:"pathHash"`         // Path hash // 路径哈希
	Version          int64      `json:"version" form:"version"`           // Version number // 版本号
	Ctime            int64      `json:"ctime" form:"ctime"`               // Creation timestamp // 创建时间戳
	Mtime            int64      `json:"mtime" form:"mtime"`               // Modification timestamp // 修改时间戳
	Size             int64      `json:"size" form:"size"`                 // Note size // 笔记大小
	UpdatedTimestamp int64      `json:"lastTime" form:"updatedTimestamp"` // Record update timestamp // 记录更新时间戳
	UpdatedAt        timex.Time `json:"updatedAt"`                        // Updated at time // 更新时间
	CreatedAt        timex.Time `json:"createdAt"`                        // Created at time // 创建时间
}

// NoteReplaceResponse response for replace operation
// NoteReplaceResponse 替换操作响应
type NoteReplaceResponse struct {
	MatchCount int      `json:"matchCount"` // Number of matches found // 匹配数量
	Note       *NoteDTO `json:"note"`       // Updated note data // 更新后的笔记数据
}

// NoteLinkItem represents a link in backlinks/outlinks response
// NoteLinkItem 代表反向链接/出链响应中的链接项
type NoteLinkItem struct {
	Path     string `json:"path"`               // Target path // 目标路径
	LinkText string `json:"linkText,omitempty"` // Raw link text (optional) // 原始链接文本（可选）
	Context  string `json:"context,omitempty"`  // Text context around link // 链接文本上下文
	IsEmbed  bool   `json:"isEmbed"`            // Is it an embed (![[...]]) // 是否为嵌入
}

// NoteWithFileLinksResponse Note response structure with file links
// 带有文件链接的笔记响应结构体
type NoteWithFileLinksResponse struct {
	ID               int64             `json:"-"`           // Note ID // 笔记 ID
	Path             string            `json:"path"`        // Note path // 笔记路径
	PathHash         string            `json:"pathHash"`    // Path hash // 路径哈希
	Content          string            `json:"content"`     // Note content // 笔记内容
	ContentHash      string            `json:"contentHash"` // Content hash // 内容哈希
	FileLinks        map[string]string `json:"fileLinks"`   // Map of file link to actual path // 文件链接到实际路径的映射
	Version          int64             `json:"version"`     // Version number // 版本号
	Ctime            int64             `json:"ctime"`       // Creation timestamp // 创建时间戳
	Mtime            int64             `json:"mtime"`       // Modification timestamp // 修改时间戳
	UpdatedTimestamp int64             `json:"lastTime"`    // Record update timestamp // 记录更新时间戳
	UpdatedAt        interface{}       `json:"updatedAt"`   // Updated at time // 更新时间
	CreatedAt        interface{}       `json:"createdAt"`   // Created at time // 创建时间
}

// NoteHistoryDTO Note history data transfer object
// 笔记历史数据传输对象
type NoteHistoryDTO struct {
	ID          int64                 `json:"id" form:"id"`                   // History entry ID // 历史项 ID
	NoteID      int64                 `json:"noteId" form:"noteId"`           // Associated note ID // 笔记 ID
	VaultID     int64                 `json:"vaultId" form:"vaultId"`         // Associated vault ID // 保险库 ID
	Path        string                `json:"path" form:"path"`               // Note path at that time // 当时的笔记路径
	Diffs       []diffmatchpatch.Diff `json:"diffs"`                          // Text differences // 文本差异内容
	Content     string                `json:"content" form:"content"`         // Full historical content // 完整历史内容
	ContentHash string                `json:"contentHash" form:"contentHash"` // Content hash // 内容哈希
	ClientName  string                `json:"clientName" form:"clientName"`   // Client that made changes // 产生变更的客户端
	Version     int64                 `json:"version" form:"version"`         // Historical version number // 历史版本号
	CreatedAt   timex.Time            `json:"createdAt" form:"createdAt"`     // Creation time of this version // 此版本的创建时间
}

// NoteHistoryNoContentDTO Note history DTO without content
// 不包含内容的笔记历史 DTO
type NoteHistoryNoContentDTO struct {
	ID         int64      `json:"id" form:"id"`                 // History entry ID // 历史项 ID
	NoteID     int64      `json:"noteId" form:"noteId"`         // Associated note ID // 笔记 ID
	VaultID    int64      `json:"vaultId" form:"vaultId"`       // Associated vault ID // 保险库 ID
	Path       string     `json:"path" form:"path"`             // Note path at that time // 当时的笔记路径
	ClientName string     `json:"clientName" form:"clientName"` // Client that made changes // 产生变更的客户端
	Version    int64      `json:"version" form:"version"`       // Historical version number // 历史版本号
	CreatedAt  timex.Time `json:"createdAt" form:"createdAt"`   // Creation time of this version // 此版本的创建时间
}
