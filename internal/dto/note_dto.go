// Package dto Defines data transfer objects (request parameters and response structs)
// Package dto 定义数据传输对象（请求参数和响应结构体）
package dto

import (
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// NoteDTO Note data transfer object
// NoteDTO 笔记数据传输对象
type NoteDTO struct {
	ID               int64      `json:"-" form:"id"`
	Action           string     `json:"-" form:"action"`
	Path             string     `json:"path" form:"path"`
	PathHash         string     `json:"pathHash" form:"pathHash"`
	Content          string     `json:"content" form:"content"`
	ContentHash      string     `json:"contentHash" form:"contentHash"`
	Version          int64      `json:"version" form:"version"`
	Ctime            int64      `json:"ctime" form:"ctime"`
	Mtime            int64      `json:"mtime" form:"mtime"`
	Size             int64      `json:"size" form:"size"`
	UpdatedTimestamp int64      `json:"lastTime" form:"updatedTimestamp"`
	UpdatedAt        timex.Time `json:"-"`
	CreatedAt        timex.Time `json:"-"`
}

// NoteNoContentDTO Note DTO without content
// NoteNoContentDTO 不包含内容的笔记 DTO
type NoteNoContentDTO struct {
	ID               int64      `json:"-" form:"id"`
	Action           string     `json:"-" form:"action"`
	Path             string     `json:"path" form:"path"`
	PathHash         string     `json:"pathHash" form:"pathHash"`
	Version          int64      `json:"version" form:"version"`
	Ctime            int64      `json:"ctime" form:"ctime"`
	Mtime            int64      `json:"mtime" form:"mtime"`
	Size             int64      `json:"size" form:"size"`
	UpdatedTimestamp int64      `json:"lastTime" form:"updatedTimestamp"`
	UpdatedAt        timex.Time `json:"-"`
	CreatedAt        timex.Time `json:"-"`
}

// NoteUpdateCheckRequest Client request parameters for checking if updates are needed
// 客户端用于检查是否需要更新的请求参数
type NoteUpdateCheckRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required"`
	Path        string `json:"path" form:"path" binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"`
	ContentHash string `json:"contentHash" form:"contentHash" binding:""`
	Ctime       int64  `json:"ctime" form:"ctime" binding:"required"`
	Mtime       int64  `json:"mtime" form:"mtime" binding:"required"`
}

// NoteModifyOrCreateRequest Request parameters for creating or modifying a note
// 用于创建或修改笔记的请求参数
type NoteModifyOrCreateRequest struct {
	Vault           string `json:"vault" form:"vault" binding:"required"`
	Path            string `json:"path" form:"path" binding:"required"`
	PathHash        string `json:"pathHash" form:"pathHash"`
	SrcPath         string `json:"srcPath" form:"srcPath"`
	SrcPathHash     string `json:"srcPathHash" form:"srcPathHash"`
	BaseHash        string `json:"baseHash" form:"baseHash" binding:""`
	BaseHashMissing bool   `json:"baseHashMissing" form:"baseHashMissing"` // Marks if baseHash is unavailable
	Content         string `json:"content" form:"content" binding:""`
	ContentHash     string `json:"contentHash" form:"contentHash" binding:""`
	Ctime           int64  `json:"ctime" form:"ctime"`
	Mtime           int64  `json:"mtime" form:"mtime"`
	OldPath         string `json:"oldPath" form:"oldPath"`
	OldPathHash     string `json:"oldPathHash" form:"oldPathHash"`
	CreateOnly      bool   `json:"createOnly" form:"createOnly"` // If true, fail if note already exists
}

// ContentModifyRequest Request parameters for modifying content only
// 专用于只修改内容的请求参数
type ContentModifyRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required"`
	Path        string `json:"path" form:"path" binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"`
	Content     string `json:"content" form:"content" binding:"required"`
	ContentHash string `json:"contentHash" form:"contentHash" binding:"required"`
	Ctime       int64  `json:"ctime" form:"ctime" binding:"required"`
	Mtime       int64  `json:"mtime" form:"mtime" binding:"required"`
}

// NoteDeleteRequest Parameters required for deleting a note
// 删除笔记所需参数
type NoteDeleteRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash"`
}

// NoteRestoreRequest parameters for restoring a note
// NoteRestoreRequest 恢复笔记请求参数
type NoteRestoreRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash"`
}

// NotePatchFrontmatterRequest parameters for patching note frontmatter
// NotePatchFrontmatterRequest 修改笔记 Frontmatter 请求参数
type NotePatchFrontmatterRequest struct {
	Vault    string                 `json:"vault" form:"vault" binding:"required"`
	Path     string                 `json:"path" form:"path" binding:"required"`
	PathHash string                 `json:"pathHash" form:"pathHash"`
	Updates  map[string]interface{} `json:"updates"`
	Remove   []string               `json:"remove"`
}

// NoteAppendRequest parameters for appending content to a note
// NoteAppendRequest 追加笔记内容请求参数
type NoteAppendRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash"`
	Content  string `json:"content" form:"content" binding:"required"`
}

// NotePrependRequest parameters for prepending content to a note
// NotePrependRequest 在笔记头部添加内容请求参数
type NotePrependRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash"`
	Content  string `json:"content" form:"content" binding:"required"`
}

// NoteReplaceRequest parameters for find/replace in a note
// NoteReplaceRequest 笔记查找替换请求参数
type NoteReplaceRequest struct {
	Vault         string `json:"vault" form:"vault" binding:"required"`
	Path          string `json:"path" form:"path" binding:"required"`
	PathHash      string `json:"pathHash" form:"pathHash"`
	Find          string `json:"find" form:"find" binding:"required"`
	Replace       string `json:"replace" form:"replace"`
	Regex         bool   `json:"regex" form:"regex"`
	All           bool   `json:"all" form:"all"`
	FailIfNoMatch bool   `json:"failIfNoMatch" form:"failIfNoMatch"`
}

// NoteReplaceResponse response for replace operation
// NoteReplaceResponse 替换操作响应
type NoteReplaceResponse struct {
	MatchCount int      `json:"matchCount"`
	Note       *NoteDTO `json:"note"`
}

// NoteMoveRequest parameters for moving a note
// NoteMoveRequest 移动笔记请求参数
type NoteMoveRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required"`
	Path        string `json:"path" form:"path" binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash"`
	Destination string `json:"destination" form:"destination" binding:"required"`
	Overwrite   bool   `json:"overwrite" form:"overwrite"`
}

// NoteLinkQueryRequest parameters for backlinks/outlinks query
// NoteLinkQueryRequest 反向链接/出链查询请求参数
type NoteLinkQueryRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash"`
}

// NoteLinkItem represents a link in backlinks/outlinks response
// NoteLinkItem 代表反向链接/出链响应中的链接项
type NoteLinkItem struct {
	Path     string `json:"path"`
	LinkText string `json:"linkText,omitempty"`
	Context  string `json:"context,omitempty"` // snippet around link
	IsEmbed  bool   `json:"isEmbed"`           // true if embed (![[...]]) vs regular link ([[...]])
}

// NoteSyncCheckRequest Parameters for checking synchronization of a single record
// 同步检查单条记录的参数
type NoteSyncCheckRequest struct {
	Path        string `json:"path" form:"path"`
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"`
	ContentHash string `json:"contentHash" form:"contentHash" binding:""`
	Mtime       int64  `json:"mtime" form:"mtime" binding:"required"`
}

type NoteSyncDelNote struct {
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash" binding:"required"`
}

// NoteSyncRequest Synchronization request body
// 同步请求主体
type NoteSyncRequest struct {
	Vault        string                 `json:"vault" form:"vault" binding:"required"`
	LastTime     int64                  `json:"lastTime" form:"lastTime"`
	Notes        []NoteSyncCheckRequest `json:"notes" form:"notes"`
	DelNotes     []NoteSyncDelNote      `json:"delNotes" form:"delNotes"`
	MissingNotes []NoteSyncDelNote      `json:"missingNotes" form:"missingNotes"`
}

// ModifyMtimeFilesRequest Request for querying modified files by mtime
// 用于按 mtime 查询修改文件
type ModifyMtimeFilesRequest struct {
	Vault string `json:"vault" form:"vault" binding:"required"`
	Mtime int64  `json:"mtime" form:"mtime"`
}

// NoteGetRequest Request parameters for retrieving a single note
// 用于获取单条笔记的请求参数
type NoteGetRequest struct {
	Vault     string `json:"vault" form:"vault" binding:"required"`
	Path      string `json:"path" form:"path" binding:"required"`
	PathHash  string `json:"pathHash" form:"pathHash"`
	IsRecycle bool   `json:"isRecycle" form:"isRecycle"`
}

// NoteRenameRequest Parameters required for renaming a note
// 重命名笔记所需参数
type NoteRenameRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required"`
	Path        string `json:"path" form:"path" binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"`
	OldPath     string `json:"oldPath" form:"oldPath" binding:"required"`
	OldPathHash string `json:"oldPathHash" form:"oldPathHash" binding:"required"`
}

// NoteListRequest Pagination parameters for retrieving the note list
// 获取笔记列表的分页参数
type NoteListRequest struct {
	Vault         string `json:"vault" form:"vault" binding:"required"`
	Keyword       string `json:"keyword" form:"keyword"`
	IsRecycle     bool   `json:"isRecycle" form:"isRecycle"`
	SearchMode    string `json:"searchMode" form:"searchMode"`       // Search mode: path(default), content, regex
	SearchContent bool   `json:"searchContent" form:"searchContent"` // Whether to search content
	SortBy        string `json:"sortBy" form:"sortBy"`               // Sorting field: mtime(default), ctime, path
	SortOrder     string `json:"sortOrder" form:"sortOrder"`         // Sorting order: desc(default), asc
}

// NoteWithFileLinksResponse Note response structure with file links
// 带有文件链接的笔记响应结构体
type NoteWithFileLinksResponse struct {
	ID               int64             `json:"id"`
	Path             string            `json:"path"`
	PathHash         string            `json:"pathHash"`
	Content          string            `json:"content"`
	ContentHash      string            `json:"contentHash"`
	FileLinks        map[string]string `json:"fileLinks"`
	Version          int64             `json:"version"`
	Ctime            int64             `json:"ctime"`
	Mtime            int64             `json:"mtime"`
	UpdatedTimestamp int64             `json:"updatedTimestamp"`
	UpdatedAt        interface{}       `json:"updatedAt"`
	CreatedAt        interface{}       `json:"createdAt"`
}

// NoteHistoryListRequest Note history list request parameters
// 笔记历史列表请求参数
type NoteHistoryListRequest struct {
	Vault     string `json:"vault" form:"vault" binding:"required"`
	Path      string `json:"path" form:"path" binding:"required"`
	PathHash  string `json:"pathHash" form:"pathHash"`
	IsRecycle bool   `json:"isRecycle" form:"isRecycle"`
}

// NoteHistoryDTO Note history data transfer object
// 笔记历史数据传输对象
type NoteHistoryDTO struct {
	ID          int64                 `json:"id" form:"id"`
	NoteID      int64                 `json:"noteId" form:"noteId"`
	VaultID     int64                 `json:"vaultId" form:"vaultId"`
	Path        string                `json:"path" form:"path"`
	Diffs       []diffmatchpatch.Diff `json:"diffs"`
	Content     string                `json:"content" form:"content"`
	ContentHash string                `json:"contentHash" form:"contentHash"`
	ClientName  string                `json:"clientName" form:"clientName"`
	Version     int64                 `json:"version" form:"version"`
	CreatedAt   timex.Time            `json:"createdAt" form:"createdAt"`
}

// NoteHistoryNoContentDTO Note history DTO without content
// 不包含内容的笔记历史 DTO
type NoteHistoryNoContentDTO struct {
	ID         int64      `json:"id" form:"id"`
	NoteID     int64      `json:"noteId" form:"noteId"`
	VaultID    int64      `json:"vaultId" form:"vaultId"`
	Path       string     `json:"path" form:"path"`
	ClientName string     `json:"clientName" form:"clientName"`
	Version    int64      `json:"version" form:"version"`
	CreatedAt  timex.Time `json:"createdAt" form:"createdAt"`
}

// NoteHistoryRestoreRequest Request parameters for restoring a historical version
// 历史版本恢复请求参数
type NoteHistoryRestoreRequest struct {
	Vault     string `json:"vault" form:"vault" binding:"required"`
	HistoryID int64  `json:"historyId" form:"historyId" binding:"required"`
}
