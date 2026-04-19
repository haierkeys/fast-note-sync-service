// Package service implements the business logic layer.
// Package service 实现业务逻辑层。
package service

import (
	"regexp"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
)

// TestProperty_ConflictFilePathFormat uses property-based testing to validate that
// the conflict path format is always correct regardless of input.
// TestProperty_ConflictFilePathFormat 使用基于属性的测试验证，无论输入如何，冲突路径格式始终正确。
func TestProperty_ConflictFilePathFormat(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: generated conflict path must match {base}.conflict.{14-digit-timestamp}{ext}
	// 属性：生成的冲突路径必须符合 {base}.conflict.{14位时间戳}{ext} 格式
	properties.Property("conflict path matches expected format", prop.ForAll(
		func(dir, filename, ext string) bool {
			var originalPath string
			if dir != "" {
				originalPath = dir + "/" + filename + ext
			} else {
				originalPath = filename + ext
			}

			svc := &conflictService{}
			conflictPath := svc.generateConflictPath(originalPath)

			// Pattern: {baseName}.conflict.{14-digit-timestamp}{ext}
			// 格式: {基础名}.conflict.{14位数字时间戳}{扩展名}
			pattern := regexp.MustCompile(`^(.+)\.conflict\.(\d{14})(\.[^.]+)?$`)
			matches := pattern.FindStringSubmatch(conflictPath)
			if matches == nil {
				return false
			}

			// Base name must be preserved.
			// 基础名称必须被保留。
			baseName := matches[1]
			expectedBase := strings.TrimSuffix(originalPath, ext)
			if baseName != expectedBase {
				return false
			}

			// Extension must be preserved.
			// 扩展名必须被保留。
			gotExt := matches[3]
			return gotExt == ext
		},
		gen.AlphaString().SuchThat(func(s string) bool {
			return !strings.Contains(s, ".") && !strings.Contains(s, "/")
		}),
		gen.AlphaString().SuchThat(func(s string) bool {
			return len(s) > 0 && !strings.Contains(s, ".") && !strings.Contains(s, "/")
		}),
		gen.OneConstOf(".md", ".txt", ".json", ""),
	))

	properties.TestingRun(t)
}

// TestGenerateConflictPath verifies that generateConflictPath produces correctly formatted paths.
// TestGenerateConflictPath 验证 generateConflictPath 生成格式正确的路径。
func TestGenerateConflictPath(t *testing.T) {
	svc := &conflictService{}

	tests := []struct {
		name         string // test case name / 测试用例名称
		originalPath string // input path / 输入路径
		wantContains string // expected substring / 期望包含的子字符串
		wantSuffix   string // expected suffix / 期望的后缀
	}{
		{
			name:         "markdown file",
			originalPath: "notes/test.md",
			wantContains: "notes/test.conflict.",
			wantSuffix:   ".md",
		},
		{
			name:         "nested path",
			originalPath: "folder/subfolder/note.md",
			wantContains: "folder/subfolder/note.conflict.",
			wantSuffix:   ".md",
		},
		{
			name:         "no extension",
			originalPath: "README",
			wantContains: "README.conflict.",
			wantSuffix:   "",
		},
		{
			name:         "txt file",
			originalPath: "data.txt",
			wantContains: "data.conflict.",
			wantSuffix:   ".txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.generateConflictPath(tt.originalPath)

			assert.Contains(t, got, tt.wantContains, "path should contain expected base")

			if tt.wantSuffix != "" {
				assert.True(t, strings.HasSuffix(got, tt.wantSuffix),
					"expected suffix %q but got %q", tt.wantSuffix, got)
			}

			// Timestamp must be a 14-digit number.
			// 时间戳必须是 14 位数字。
			pattern := regexp.MustCompile(`\.conflict\.(\d{14})`)
			assert.Regexp(t, pattern, got, "conflict path should contain 14-digit timestamp")
		})
	}
}
