// Package util provides common utility functions
// Package util 提供通用工具函数
package util

import (
	"strings"

	"gopkg.in/yaml.v3"
)

const frontmatterDelimiter = "---"

// ParseFrontmatter extracts YAML frontmatter from content
// Returns the parsed YAML as a map, the body (content after frontmatter), and whether frontmatter exists
// ParseFrontmatter 从内容中提取 YAML frontmatter
// 返回解析后的 YAML map、正文（frontmatter 之后的内容）以及是否存在 frontmatter
func ParseFrontmatter(content string) (yamlData map[string]interface{}, body string, hasFrontmatter bool) {
	if content == "" {
		return nil, content, false
	}

	// Check if content starts with frontmatter delimiter
	// 检查内容是否以 frontmatter 分隔符开头
	if !strings.HasPrefix(content, frontmatterDelimiter+"\n") {
		return nil, content, false
	}

	// Find the closing delimiter
	// 查找结束分隔符
	rest := content[len(frontmatterDelimiter)+1:]
	endIndex := strings.Index(rest, "\n"+frontmatterDelimiter)
	if endIndex == -1 {
		return nil, content, false
	}

	// Extract frontmatter YAML
	// 提取 frontmatter YAML
	yamlContent := rest[:endIndex]
	body = rest[endIndex+len("\n"+frontmatterDelimiter):]

	// Remove leading newline from body if present
	// 如果 body 以换行符开头，将其移除
	body = strings.TrimPrefix(body, "\n")

	// Parse YAML
	// 解析 YAML
	yamlData = make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(yamlContent), &yamlData); err != nil {
		// If YAML parsing fails, return as if no frontmatter
		// 如果 YAML 解析失败，则当作没有 frontmatter 返回
		return nil, content, false
	}

	return yamlData, body, true
}

// MergeFrontmatter merges updates into existing frontmatter and removes specified keys
// MergeFrontmatter 将更新合并到现有的 frontmatter 中并移除指定的键
func MergeFrontmatter(existing, updates map[string]interface{}, removeKeys []string) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy existing values
	// 复制现有值
	for k, v := range existing {
		result[k] = v
	}

	// Apply updates
	// 应用更新
	for k, v := range updates {
		result[k] = v
	}

	// Remove specified keys
	// 移除指定的键
	for _, key := range removeKeys {
		delete(result, key)
	}

	return result
}

// ReconstructContent rebuilds content with frontmatter
// ReconstructContent 使用 frontmatter 重新构建内容
func ReconstructContent(yamlData map[string]interface{}, body string) string {
	if len(yamlData) == 0 {
		return body
	}

	yamlBytes, err := yaml.Marshal(yamlData)
	if err != nil {
		return body
	}

	var sb strings.Builder
	sb.WriteString(frontmatterDelimiter)
	sb.WriteString("\n")
	sb.Write(yamlBytes)
	sb.WriteString(frontmatterDelimiter)
	sb.WriteString("\n")
	sb.WriteString(body)

	return sb.String()
}
