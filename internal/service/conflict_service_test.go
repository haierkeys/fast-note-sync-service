package service

import (
	"regexp"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// 验证冲突文件路径格式正确

func TestProperty6_ConflictFilePathFormat(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// 冲突文件路径格式验证
	properties.Property("conflict path matches expected format", prop.ForAll(
		func(dir, filename, ext string) bool {
			// 构造原始路径
			var originalPath string
			if dir != "" {
				originalPath = dir + "/" + filename + ext
			} else {
				originalPath = filename + ext
			}

			// 使用 conflictService 的路径生成逻辑
			svc := &conflictService{}
			conflictPath := svc.generateConflictPath(originalPath)

			// 验证格式: {baseName}.conflict.{timestamp}{ext}
			// timestamp 格式: 20060102150405 (14位数字)
			pattern := regexp.MustCompile(`^(.+)\.conflict\.(\d{14})(\.[^.]+)?$`)
			matches := pattern.FindStringSubmatch(conflictPath)

			if matches == nil {
				t.Logf("Path doesn't match pattern: %s", conflictPath)
				return false
			}

			// 验证基础名称保留
			baseName := matches[1]
			expectedBase := strings.TrimSuffix(originalPath, ext)
			if baseName != expectedBase {
				t.Logf("Base name mismatch: got %s, want %s", baseName, expectedBase)
				return false
			}

			// 验证扩展名保留
			gotExt := matches[3]
			if gotExt != ext {
				t.Logf("Extension mismatch: got %s, want %s", gotExt, ext)
				return false
			}

			return true
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

// 单元测试: 冲突文件路径生成
func TestGenerateConflictPath(t *testing.T) {
	svc := &conflictService{}

	tests := []struct {
		name         string
		originalPath string
		wantContains string
		wantSuffix   string
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

			if !strings.Contains(got, tt.wantContains) {
				t.Errorf("generateConflictPath() = %v, want contains %v", got, tt.wantContains)
			}

			if tt.wantSuffix != "" && !strings.HasSuffix(got, tt.wantSuffix) {
				t.Errorf("generateConflictPath() = %v, want suffix %v", got, tt.wantSuffix)
			}

			// 验证时间戳格式 (14位数字)
			pattern := regexp.MustCompile(`\.conflict\.(\d{14})`)
			if !pattern.MatchString(got) {
				t.Errorf("generateConflictPath() = %v, timestamp format invalid", got)
			}
		})
	}
}
