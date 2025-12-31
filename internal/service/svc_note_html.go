package service

import (
	"bytes"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"github.com/yuin/goldmark"
)

// NoteGenerateHTML 生成笔记的 HTML 代码
// 函数名: NoteGenerateHTML
// 函数使用说明: 根据笔记路径获取笔记内容,处理嵌入标签并将 Markdown 转换为 HTML 格式并返回
// 参数说明:
//   - uid int64: 用户ID
//   - params *NoteGetRequestParams: 包含 vault/path/pathHash 的请求参数
//
// 返回值说明:
//   - string: HTML 代码
//   - error: 出错时返回错误
func (svc *Service) NoteGenerateHTML(uid int64, params *NoteGetRequestParams) (string, error) {
	// 1. 获取主笔记内容
	note, err := svc.NoteGet(uid, params)
	if err != nil {
		return "", fmt.Errorf("获取笔记失败: %w", err)
	}

	if note == nil || note.Action == "delete" {
		return "", fmt.Errorf("笔记不存在或已删除")
	}

	// 2. 处理 Obsidian 嵌入语法 ![[]]
	// 使用递归处理以支持“画中画”效果, 默认深度限制为 3 层
	processedContent := svc.processMarkdownEmbeds(uid, params.Vault, note.Content, 0)

	// 3. 配置 goldmark HTML 渲染器
	md := goldmark.New(
		goldmark.WithExtensions(),
	)

	// 4. 转换 Markdown 到 HTML
	var buf bytes.Buffer
	if err := md.Convert([]byte(processedContent), &buf); err != nil {
		return "", fmt.Errorf("转换 Markdown 到 HTML 失败: %w", err)
	}

	return buf.String(), nil
}

// processMarkdownEmbeds 递归处理 Markdown 中的嵌入标签 ![[]]
func (svc *Service) processMarkdownEmbeds(uid int64, vault string, content string, depth int) string {
	if depth > 3 { // 防止循环嵌套溢出
		return content
	}

	// 匹配 ![[path|options]]
	re := regexp.MustCompile(`!\[\[(.+?)\]\]`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		inner := match[3 : len(match)-2] // 去掉 ![[ 和 ]]
		parts := strings.Split(inner, "|")
		pathWithAnchor := strings.TrimSpace(parts[0])
		options := parts[1:]

		// 剥离锚点 #标题 或 #^blockid (暂不支持精确定位, 只取文件路径)
		filePath := pathWithAnchor
		if idx := strings.Index(pathWithAnchor, "#"); idx != -1 {
			filePath = pathWithAnchor[:idx]
		}

		if filePath == "" {
			return match
		}

		// 检查路径后缀确定类型
		ext := strings.ToLower(filepath.Ext(filePath))
		isImage := strings.Contains(".jpg.jpeg.png.gif.webp.svg.bmp", ext) && ext != ""
		isPDF := ext == ".pdf"

		if isImage {
			// 处理图片嵌入
			apiUrl := fmt.Sprintf("/api/note/file?vault=%s&path=%s", url.QueryEscape(vault), url.QueryEscape(filePath))

			// 处理尺寸选项 |100 或 |100x200
			width := ""
			height := ""
			if len(options) > 0 {
				sizePart := options[0]
				if strings.Contains(sizePart, "x") {
					sizes := strings.Split(sizePart, "x")
					width = sizes[0]
					if len(sizes) > 1 {
						height = sizes[1]
					}
				} else {
					width = sizePart
				}
			}

			// 如果有尺寸信息, 使用 <img> 标签以支持尺寸控制
			if width != "" {
				imgTag := fmt.Sprintf(`<img src="%s" width="%s"`, apiUrl, width)
				if height != "" {
					imgTag += fmt.Sprintf(` height="%s"`, height)
				}
				imgTag += " />"
				return imgTag
			}

			// 默认返回标准 Markdown 图片语法
			return fmt.Sprintf("![%s](%s)", filePath, apiUrl)
		}

		if isPDF {
			// 处理 PDF 嵌入 (提供预览连接)
			apiUrl := fmt.Sprintf("/api/note/file?vault=%s&path=%s", url.QueryEscape(vault), url.QueryEscape(filePath))
			return fmt.Sprintf(`<embed src="%s" type="application/pdf" width="100%%" height="600px" />`, apiUrl)
		}

		// 处理笔记嵌入 (MD 文件或无后缀默认视为笔记)
		if ext == "" || ext == ".md" {
			params := &NoteGetRequestParams{
				Vault: vault,
				Path:  filePath,
			}
			// 递归获取引用的笔记内容
			embedNote, _, err := svc.NoteGetFileContent(uid, params)
			if err != nil || embedNote == nil {
				return fmt.Sprintf(`[Embed Missing: %s]`, filePath)
			}

			// 对嵌入内容继续进行解析 (递归)
			innerContent := svc.processMarkdownEmbeds(uid, vault, string(embedNote), depth+1)

			return fmt.Sprintf("\n<div class=\"embedded-note\">\n\n%s\n\n</div>\n", innerContent)
		}

		// 其他类型文件作为普通链接
		apiUrl := fmt.Sprintf("/api/note/file?vault=%s&path=%s", url.QueryEscape(vault), url.QueryEscape(filePath))
		return fmt.Sprintf("[%s](%s)", filePath, apiUrl)
	})
}

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
//   - error: 出错时返回错误
func (svc *Service) NoteGetFileContent(uid int64, params *NoteGetRequestParams) ([]byte, string, error) {
	// 1. 获取仓库 ID
	vaultID, err := svc.VaultIdGetByName(params.Vault, uid)
	if err != nil {
		return nil, "", fmt.Errorf("获取仓库失败: %w", err)
	}

	// 2. 确认路径哈希
	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// 3. 优先尝试从 Note 表获取 (笔记/文本内容)
	note, err := svc.dao.NoteGetByPathHash(params.PathHash, vaultID, uid)
	if err == nil && note != nil && note.Action != "delete" {
		// 笔记内容固定识别为 markdown
		return []byte(note.Content), "text/markdown; charset=utf-8", nil
	}

	// 4. 尝试从 File 表获取 (附件/二进制文件)
	file, err := svc.dao.FileGetByPathHash(params.PathHash, vaultID, uid)
	if err == nil && file != nil && file.Action != "delete" {
		// 读取物理文件内容
		content, err := os.ReadFile(file.SavePath)
		if err != nil {
			return nil, "", fmt.Errorf("读取物理文件失败: %w", err)
		}

		// 识别文件 MIME 类型
		ext := filepath.Ext(params.Path)
		contentType := mime.TypeByExtension(ext)
		if contentType == "" {
			// 如果扩展名识别不到, 进行内容嗅探
			contentType = http.DetectContentType(content)
		}

		return content, contentType, nil
	}

	return nil, "", nil
}
