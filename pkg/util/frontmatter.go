// Package util provides common utility functions
package util

import (
	"strings"

	"gopkg.in/yaml.v3"
)

const frontmatterDelimiter = "---"

// ParseFrontmatter extracts YAML frontmatter from content
// Returns the parsed YAML as a map, the body (content after frontmatter), and whether frontmatter exists
func ParseFrontmatter(content string) (yamlData map[string]interface{}, body string, hasFrontmatter bool) {
	if content == "" {
		return nil, content, false
	}

	// Check if content starts with frontmatter delimiter
	if !strings.HasPrefix(content, frontmatterDelimiter+"\n") {
		return nil, content, false
	}

	// Find the closing delimiter
	rest := content[len(frontmatterDelimiter)+1:]
	endIndex := strings.Index(rest, "\n"+frontmatterDelimiter)
	if endIndex == -1 {
		return nil, content, false
	}

	// Extract frontmatter YAML
	yamlContent := rest[:endIndex]
	body = rest[endIndex+len("\n"+frontmatterDelimiter):]

	// Remove leading newline from body if present
	if strings.HasPrefix(body, "\n") {
		body = body[1:]
	}

	// Parse YAML
	yamlData = make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(yamlContent), &yamlData); err != nil {
		// If YAML parsing fails, return as if no frontmatter
		return nil, content, false
	}

	return yamlData, body, true
}

// MergeFrontmatter merges updates into existing frontmatter and removes specified keys
func MergeFrontmatter(existing, updates map[string]interface{}, removeKeys []string) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy existing values
	for k, v := range existing {
		result[k] = v
	}

	// Apply updates
	for k, v := range updates {
		result[k] = v
	}

	// Remove specified keys
	for _, key := range removeKeys {
		delete(result, key)
	}

	return result
}

// ReconstructContent rebuilds content with frontmatter
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
