package mcp_router

import "github.com/haierkeys/fast-note-sync-service/internal/dto"

type mcpNoteListOutput struct {
	Vault string                     `json:"vault"`
	Count int                        `json:"count"`
	Notes []*dto.McpNoteNoContentDTO `json:"notes"`
}

type mcpNoteOutput struct {
	Vault string          `json:"vault"`
	Note  *dto.McpNoteDTO `json:"note"`
}

type mcpNoteMutationOutput struct {
	Vault     string          `json:"vault"`
	Operation string          `json:"operation"`
	Note      *dto.McpNoteDTO `json:"note,omitempty"`
	OldNote   *dto.McpNoteDTO `json:"oldNote,omitempty"`
	NewNote   *dto.McpNoteDTO `json:"newNote,omitempty"`
}

type mcpNoteRecycleClearOutput struct {
	Vault string `json:"vault"`
	Path  string `json:"path,omitempty"`
}

type mcpNoteReplaceOutput struct {
	Vault      string          `json:"vault"`
	MatchCount int             `json:"matchCount"`
	Note       *dto.McpNoteDTO `json:"note"`
}

type mcpNoteLinksOutput struct {
	Vault string              `json:"vault"`
	Path  string              `json:"path"`
	Count int                 `json:"count"`
	Links []*dto.NoteLinkItem `json:"links"`
}

type mcpFileListOutput struct {
	Vault string            `json:"vault"`
	Count int               `json:"count"`
	Files []*dto.McpFileDTO `json:"files"`
}

type mcpFileOutput struct {
	Vault string          `json:"vault"`
	File  *dto.McpFileDTO `json:"file"`
}

type mcpFileReadOutput struct {
	Vault         string `json:"vault"`
	Path          string `json:"path"`
	ContentBase64 string `json:"contentBase64"`
	Size          int    `json:"size"`
}

type mcpFileMutationOutput struct {
	Vault     string          `json:"vault"`
	Operation string          `json:"operation"`
	File      *dto.McpFileDTO `json:"file,omitempty"`
	OldFile   *dto.McpFileDTO `json:"oldFile,omitempty"`
	NewFile   *dto.McpFileDTO `json:"newFile,omitempty"`
}

type mcpFileRecycleClearOutput struct {
	Vault string `json:"vault"`
	Path  string `json:"path,omitempty"`
}

type mcpFileWriteOutput struct {
	Vault string          `json:"vault"`
	File  *dto.McpFileDTO `json:"file"`
}

type mcpFolderMutationOutput struct {
	Vault     string         `json:"vault"`
	Operation string         `json:"operation"`
	Folder    *dto.FolderDTO `json:"folder,omitempty"`
}

type mcpVaultListOutput struct {
	Count  int             `json:"count"`
	Vaults []*dto.VaultDTO `json:"vaults"`
}

type mcpVaultOutput struct {
	Vault *dto.VaultDTO `json:"vault"`
}

type mcpVaultMutationOutput struct {
	Operation string        `json:"operation"`
	Vault     *dto.VaultDTO `json:"vault,omitempty"`
	ID        int64         `json:"id,omitempty"`
}
