package dto

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

// McpNoteDTO is a MCP-specific note DTO with time fields as string
// McpNoteDTO 是 MCP 专用的笔记 DTO，时间字段使用 string 类型以匹配 MCP output schema
type McpNoteDTO struct {
	Path             string `json:"path"`          // Note path // 笔记路径
	PathHash         string `json:"pathHash"`      // Path hash // 路径哈希
	Content          string `json:"content"`       // Note content // 笔记内容
	ContentHash      string `json:"contentHash"`   // Content hash // 内容哈希
	Version          int64  `json:"version"`       // Version number // 版本号
	Ctime            int64  `json:"ctime"`         // Creation timestamp // 创建时间戳
	Mtime            int64  `json:"mtime"`         // Modification timestamp // 修改时间戳
	Size             int64  `json:"size"`          // Note size // 笔记大小
	ClientName       string `json:"clientName"`    // Client name // 客户端名称
	ClientType       string `json:"clientType"`    // Client type // 客户端类型
	ClientVersion    string `json:"clientVersion"` // Client version // 客户端版本
	UpdatedTimestamp int64  `json:"lastTime"`      // Record update timestamp // 记录更新时间戳
	UpdatedAt        string `json:"updatedAt"`     // Updated at time // 更新时间
	CreatedAt        string `json:"createdAt"`     // Created at time // 创建时间
}

// McpNoteNoContentDTO is a MCP-specific note DTO without content
// McpNoteNoContentDTO 是 MCP 专用的不含内容的笔记 DTO
type McpNoteNoContentDTO struct {
	Path             string `json:"path"`          // Note path // 笔记路径
	PathHash         string `json:"pathHash"`      // Path hash // 路径哈希
	Version          int64  `json:"version"`       // Version number // 版本号
	Ctime            int64  `json:"ctime"`         // Creation timestamp // 创建时间戳
	Mtime            int64  `json:"mtime"`         // Modification timestamp // 修改时间戳
	Size             int64  `json:"size"`          // Note size // 笔记大小
	ClientName       string `json:"clientName"`    // Client name // 客户端名称
	ClientType       string `json:"clientType"`    // Client type // 客户端类型
	ClientVersion    string `json:"clientVersion"` // Client version // 客户端版本
	UpdatedTimestamp int64  `json:"lastTime"`      // Record update timestamp // 记录更新时间戳
	UpdatedAt        string `json:"updatedAt"`     // Updated at time // 更新时间
	CreatedAt        string `json:"createdAt"`     // Created at time // 创建时间
}

// McpFileDTO is a MCP-specific file DTO with time fields as string
// McpFileDTO 是 MCP 专用的文件 DTO，时间字段使用 string 类型以匹配 MCP output schema
type McpFileDTO struct {
	Path             string `json:"path"`        // File path // 文件路径
	PathHash         string `json:"pathHash"`    // Path hash // 路径哈希
	ContentHash      string `json:"contentHash"` // Content hash // 内容哈希
	Rename           int64  `json:"rename"`      // Rename flag // 重命名标记
	Size             int64  `json:"size"`        // File size // 文件大小
	Ctime            int64  `json:"ctime"`       // Creation timestamp // 创建时间戳
	Mtime            int64  `json:"mtime"`       // Modification timestamp // 修改时间戳
	UpdatedTimestamp int64  `json:"lastTime"`    // Updated timestamp // 更新时间戳
	UpdatedAt        string `json:"updatedAt"`   // Updated at time // 更新时间
	CreatedAt        string `json:"createdAt"`   // Created at time // 创建时间
}

// formatMcpTime formats timex.Time to string, returning empty string for zero time values
// formatMcpTime 将 timex.Time 格式化为字符串，零值返回空字符串以匹配 MCP output schema 的 type: string
func formatMcpTime(t timex.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.String()
}

// ToMcpNoteDTO converts NoteDTO to McpNoteDTO for MCP output schema compatibility
// ToMcpNoteDTO 将 NoteDTO 转换为 McpNoteDTO，以兼容 MCP output schema
func (note *NoteDTO) ToMcpNoteDTO() *McpNoteDTO {
	if note == nil {
		return nil
	}
	return &McpNoteDTO{
		Path:             note.Path,
		PathHash:         note.PathHash,
		Content:          note.Content,
		ContentHash:      note.ContentHash,
		Version:          note.Version,
		Ctime:            note.Ctime,
		Mtime:            note.Mtime,
		Size:             note.Size,
		ClientName:       note.ClientName,
		ClientType:       note.ClientType,
		ClientVersion:    note.ClientVersion,
		UpdatedTimestamp: note.UpdatedTimestamp,
		UpdatedAt:        formatMcpTime(note.UpdatedAt),
		CreatedAt:        formatMcpTime(note.CreatedAt),
	}
}

// ToMcpNoteNoContentDTO converts NoteNoContentDTO to McpNoteNoContentDTO
// ToMcpNoteNoContentDTO 将 NoteNoContentDTO 转换为 McpNoteNoContentDTO
func (note *NoteNoContentDTO) ToMcpNoteNoContentDTO() *McpNoteNoContentDTO {
	if note == nil {
		return nil
	}
	return &McpNoteNoContentDTO{
		Path:             note.Path,
		PathHash:         note.PathHash,
		Version:          note.Version,
		Ctime:            note.Ctime,
		Mtime:            note.Mtime,
		Size:             note.Size,
		ClientName:       note.ClientName,
		ClientType:       note.ClientType,
		ClientVersion:    note.ClientVersion,
		UpdatedTimestamp: note.UpdatedTimestamp,
		UpdatedAt:        formatMcpTime(note.UpdatedAt),
		CreatedAt:        formatMcpTime(note.CreatedAt),
	}
}

// ToMcpFileDTO converts FileDTO to McpFileDTO for MCP output schema compatibility
// ToMcpFileDTO 将 FileDTO 转换为 McpFileDTO，以兼容 MCP output schema
func (file *FileDTO) ToMcpFileDTO() *McpFileDTO {
	if file == nil {
		return nil
	}
	return &McpFileDTO{
		Path:             file.Path,
		PathHash:         file.PathHash,
		ContentHash:      file.ContentHash,
		Rename:           file.Rename,
		Size:             file.Size,
		Ctime:            file.Ctime,
		Mtime:            file.Mtime,
		UpdatedTimestamp: file.UpdatedTimestamp,
		UpdatedAt:        formatMcpTime(file.UpdatedAt),
		CreatedAt:        formatMcpTime(file.CreatedAt),
	}
}

// McpNoteHitDTO is a MCP-specific search hit: note metadata plus a short
// context snippet around the first keyword occurrence. The full body is
// deliberately omitted so a search never blows up the agent's context window;
// call note_get with the returned path to read the whole note.
// McpNoteHitDTO 是 MCP 专用的搜索命中项：笔记元数据 + 关键词附近的上下文片段。
// 故意不返回正文，避免一次搜索撑爆 Agent 上下文；需要全文时用返回的 path 调 note_get。
type McpNoteHitDTO struct {
	Path             string `json:"path"`      // Note path // 笔记路径
	PathHash         string `json:"pathHash"`  // Path hash // 路径哈希
	Snippet          string `json:"snippet"`   // Context around the first match // 首个命中位置附近的上下文
	MatchLine        int    `json:"matchLine"` // 1-based line number of the first match, 0 if not found // 首个命中所在行号(从1起)，未找到为0
	Size             int64  `json:"size"`      // Note size in bytes // 笔记字节数
	Version          int64  `json:"version"`   // Version number // 版本号
	Mtime            int64  `json:"mtime"`     // Modification timestamp // 修改时间戳
	UpdatedTimestamp int64  `json:"lastTime"`  // Record update timestamp // 记录更新时间戳
	UpdatedAt        string `json:"updatedAt"` // Updated at time // 更新时间
}

// McpNoteSearchOutput is the structured output of note_search
// McpNoteSearchOutput 是 note_search 的结构化输出
type McpNoteSearchOutput struct {
	Vault    string           `json:"vault"`
	Keyword  string           `json:"keyword"`
	Mode     string           `json:"mode"`  // "fulltext" or "path" // 全文或路径
	Count    int              `json:"count"` // hits returned in this page // 本页返回条数
	Total    int              `json:"total"` // total matches // 命中总数
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
	HasMore  bool             `json:"hasMore"`
	Notes    []*McpNoteHitDTO `json:"notes"`
}

// McpNoteLinkDTO is one link edge in the vault link graph
// McpNoteLinkDTO 是笔记链接图的一条边
type McpNoteLinkDTO struct {
	Path     string `json:"path"`     // The other end of the link // 链接另一端的路径
	LinkText string `json:"linkText"` // Link text as written (alias) // 书写时的链接文本(别名)
	PathHash string `json:"pathHash,omitempty"`
	Broken   bool   `json:"broken"` // Target note does not exist // 目标笔记不存在
}

// McpNoteBrokenLinkDTO is one broken wikilink with its origin
// McpNoteBrokenLinkDTO 是一条失效的双链及其来源
type McpNoteBrokenLinkDTO struct {
	SourcePath string `json:"sourcePath"` // Note containing the broken link // 含失效链接的笔记
	LinkText   string `json:"linkText"`
}

// McpNoteTagListOutput is the structured output of note_list_tags
// McpNoteTagListOutput 是 note_list_tags 的结构化输出
type McpNoteTagListOutput struct {
	Vault string       `json:"vault"`
	Scope string       `json:"scope"`
	Count int          `json:"count"`
	Tags  []*McpTagDTO `json:"tags"`
}

// McpTagDTO is a tag with its usage count
// McpTagDTO 是标签及其引用次数
type McpTagDTO struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// McpNoteBrokenLinksOutput is the structured output of note_list_broken_links
// McpNoteBrokenLinksOutput 是 note_list_broken_links 的结构化输出
type McpNoteBrokenLinksOutput struct {
	Vault      string                  `json:"vault"`
	Count      int                     `json:"count"`
	TotalLinks int                     `json:"totalLinks"` // Every link edge scanned // 扫描到的链接边总数
	Broken     []*McpNoteBrokenLinkDTO `json:"broken"`
}

// McpNoteOrphanListOutput is the structured output of note_list_orphans
// McpNoteOrphanListOutput 是 note_list_orphans 的结构化输出
type McpNoteOrphanListOutput struct {
	Vault      string   `json:"vault"`
	Count      int      `json:"count"`
	TotalNotes int      `json:"totalNotes"` // Live notes scanned // 扫描到的存活笔记数
	Notes      []string `json:"notes"`
}

// McpVaultStatsOutput is the structured output of vault_stats
// McpVaultStatsOutput 是 vault_stats 的结构化输出
type McpVaultStatsOutput struct {
	Vault     string `json:"vault"`
	NoteCount int64  `json:"noteCount"`
	NoteSize  int64  `json:"noteSize"`
	FileCount int64  `json:"fileCount"`
	FileSize  int64  `json:"fileSize"`
	TotalSize int64  `json:"totalSize"`
	LinkCount int    `json:"linkCount"`
	TagCount  int    `json:"tagCount"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
