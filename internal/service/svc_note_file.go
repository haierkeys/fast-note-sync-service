package service

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/haierkeys/fast-note-sync-service/pkg/util"
)

// NoteGetFileContent 获取文件内容
// 函数名: NoteGetFileContent
// 函数使用说明: 根据请求参数获取笔记内容或附件文件内容并返回
// 参数说明:
//   - uid int64: 用户ID
//   - params *NoteGetRequestParams: 包含 vault/path/pathHash 的请求参数
//
// 返回值说明:
//   - []byte: 文件原始数据
//   - string: MIME 类型 (Content-Type)
//   - int64: mtime (Last-Modified)
//   - string: etag (Content-Hash)
//   - error: 出错时返回错误
func (svc *Service) NoteGetFileContent(uid int64, params *NoteGetRequestParams) ([]byte, string, int64, string, error) {
	// 1. 获取仓库 ID
	vaultID, err := svc.VaultIdGetByName(params.Vault, uid)
	if err != nil {
		return nil, "", 0, "", fmt.Errorf("获取仓库失败: %w", err)
	}

	// 2. 确认路径哈希
	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// 3. 优先尝试从 Note 表获取 (笔记/文本内容)
	note, err := svc.dao.NoteGetAllByPathHash(params.PathHash, vaultID, uid)
	if err == nil && note != nil {
		// 笔记内容固定识别为 markdown
		return []byte(note.Content), "text/markdown; charset=utf-8", note.Mtime, note.ContentHash, nil
	}

	// 4. 尝试从 File 表获取 (附件/二进制文件)
	file, err := svc.dao.FileGetByPathHash(params.PathHash, vaultID, uid)
	if err == nil && file != nil && file.Action != "delete" {
		// 读取物理文件内容
		content, err := os.ReadFile(file.SavePath)
		if err != nil {
			return nil, "", 0, "", fmt.Errorf("读取物理文件失败: %w", err)
		}

		// 识别文件 MIME 类型
		ext := filepath.Ext(params.Path)
		contentType := mime.TypeByExtension(ext)
		if contentType == "" {
			// 如果扩展名识别不到, 进行内容嗅探
			contentType = http.DetectContentType(content)
		}

		// File 表没有 ContentHash 或不确定, 实时计算
		etag := util.EncodeHash32(string(content))

		return content, contentType, file.Mtime, etag, nil
	}

	return nil, "", 0, "", nil
}
