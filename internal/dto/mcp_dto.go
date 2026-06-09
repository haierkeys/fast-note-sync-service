package dto

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

// McpNoteDTO is a MCP-specific note DTO with time fields as string
// McpNoteDTO 是 MCP 专用的笔记 DTO，时间字段使用 string 类型以匹配 MCP output schema
type McpNoteDTO struct {
	Path             string `json:"path"`             // Note path // 笔记路径
	PathHash         string `json:"pathHash"`         // Path hash // 路径哈希
	Content          string `json:"content"`          // Note content // 笔记内容
	ContentHash      string `json:"contentHash"`      // Content hash // 内容哈希
	Version          int64  `json:"version"`          // Version number // 版本号
	Ctime            int64  `json:"ctime"`            // Creation timestamp // 创建时间戳
	Mtime            int64  `json:"mtime"`            // Modification timestamp // 修改时间戳
	Size             int64  `json:"size"`             // Note size // 笔记大小
	ClientName       string `json:"clientName"`       // Client name // 客户端名称
	ClientType       string `json:"clientType"`       // Client type // 客户端类型
	ClientVersion    string `json:"clientVersion"`    // Client version // 客户端版本
	UpdatedTimestamp int64  `json:"lastTime"`         // Record update timestamp // 记录更新时间戳
	UpdatedAt        string `json:"updatedAt"`        // Updated at time // 更新时间
	CreatedAt        string `json:"createdAt"`        // Created at time // 创建时间
}

// McpNoteNoContentDTO is a MCP-specific note DTO without content
// McpNoteNoContentDTO 是 MCP 专用的不含内容的笔记 DTO
type McpNoteNoContentDTO struct {
	Path             string `json:"path"`             // Note path // 笔记路径
	PathHash         string `json:"pathHash"`         // Path hash // 路径哈希
	Version          int64  `json:"version"`          // Version number // 版本号
	Ctime            int64  `json:"ctime"`            // Creation timestamp // 创建时间戳
	Mtime            int64  `json:"mtime"`            // Modification timestamp // 修改时间戳
	Size             int64  `json:"size"`             // Note size // 笔记大小
	ClientName       string `json:"clientName"`       // Client name // 客户端名称
	ClientType       string `json:"clientType"`       // Client type // 客户端类型
	ClientVersion    string `json:"clientVersion"`    // Client version // 客户端版本
	UpdatedTimestamp int64  `json:"lastTime"`         // Record update timestamp // 记录更新时间戳
	UpdatedAt        string `json:"updatedAt"`        // Updated at time // 更新时间
	CreatedAt        string `json:"createdAt"`        // Created at time // 创建时间
}

// McpFileDTO is a MCP-specific file DTO with time fields as string
// McpFileDTO 是 MCP 专用的文件 DTO，时间字段使用 string 类型以匹配 MCP output schema
type McpFileDTO struct {
	Path             string `json:"path"`             // File path // 文件路径
	PathHash         string `json:"pathHash"`         // Path hash // 路径哈希
	ContentHash      string `json:"contentHash"`      // Content hash // 内容哈希
	Rename           int64  `json:"rename"`           // Rename flag // 重命名标记
	Size             int64  `json:"size"`             // File size // 文件大小
	Ctime            int64  `json:"ctime"`            // Creation timestamp // 创建时间戳
	Mtime            int64  `json:"mtime"`            // Modification timestamp // 修改时间戳
	UpdatedTimestamp int64  `json:"lastTime"`         // Updated timestamp // 更新时间戳
	UpdatedAt        string `json:"updatedAt"`        // Updated at time // 更新时间
	CreatedAt        string `json:"createdAt"`        // Created at time // 创建时间
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
