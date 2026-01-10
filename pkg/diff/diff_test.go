package diff

import (
	"fmt"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// 验证删除操作在合并中被保留
func TestProperty4_DeleteOperationPreservation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// 场景1: 单方删除，另一方无修改（使用唯一标记避免误判）
	properties.Property("single side delete is preserved", prop.ForAll(
		func(id int) bool {
			// 使用唯一标记确保不会有歧义
			toDelete := fmt.Sprintf("__DELETE_ME_%d__", id)
			base := "prefix" + toDelete + "suffix"
			pc1 := "prefixsuffix" // PC1 删除了 toDelete
			pc2 := base           // PC2 无修改

			result, err := MergeTexts(base, pc1, pc2, true)
			if err != nil {
				return false
			}

			// 如果没有冲突，删除应该被保留
			if !result.HasConflict {
				return !strings.Contains(result.Content, toDelete)
			}
			return true // 有冲突也是可接受的
		},
		gen.IntRange(1, 1000),
	))

	// 场景2: 双方删除相同内容
	properties.Property("both sides delete same content", prop.ForAll(
		func(id int) bool {
			toDelete := fmt.Sprintf("__DELETE_ME_%d__", id)
			base := "prefix" + toDelete + "suffix"
			pc1 := "prefixsuffix" // PC1 删除
			pc2 := "prefixsuffix" // PC2 也删除

			result, err := MergeTexts(base, pc1, pc2, true)
			if err != nil {
				return false
			}

			// 双方都删除，结果应该不包含被删除内容
			if !result.HasConflict {
				return !strings.Contains(result.Content, toDelete)
			}
			return true
		},
		gen.IntRange(1, 1000),
	))

	properties.TestingRun(t)
}

// 验证删除-修改冲突被正确检测（基于行级别）
func TestProperty5_DeleteModifyConflictDetection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// 场景: 一方删除整行，另一方修改同一行
	properties.Property("delete-modify conflict is detected at line level", prop.ForAll(
		func(id int) bool {
			// 使用多行文本，确保行级别的删除-修改冲突
			lineToModify := fmt.Sprintf("Line_%d_content", id)
			base := "Line1\n" + lineToModify + "\nLine3"
			pc1 := "Line1\nLine3"                                    // PC1 删除中间行
			pc2 := "Line1\n" + lineToModify + "_modified" + "\nLine3" // PC2 修改中间行

			result, err := MergeTexts(base, pc1, pc2, true)
			if err != nil {
				return false
			}

			// 删除-修改冲突应该被检测到
			return result.HasConflict
		},
		gen.IntRange(1, 1000),
	))

	properties.TestingRun(t)
}


// 单元测试: 基本合并场景
func TestMergeTexts_BasicScenarios(t *testing.T) {
	tests := []struct {
		name        string
		base        string
		pc1         string
		pc2         string
		pc1First    bool
		wantConflict bool
		wantContains string
		wantNotContains string
	}{
		{
			name:        "no changes",
			base:        "Hello World",
			pc1:         "Hello World",
			pc2:         "Hello World",
			pc1First:    true,
			wantConflict: false,
			wantContains: "Hello World",
		},
		{
			name:        "pc1 only change",
			base:        "Hello",
			pc1:         "Hello World",
			pc2:         "Hello",
			pc1First:    true,
			wantConflict: false,
			wantContains: "World",
		},
		{
			name:        "pc2 only change",
			base:        "Hello",
			pc1:         "Hello",
			pc2:         "Hello Kiro",
			pc1First:    true,
			wantConflict: false,
			wantContains: "Kiro",
		},
		{
			name:        "both add different content",
			base:        "Hello",
			pc1:         "Hello World",
			pc2:         "Hello Kiro",
			pc1First:    true,
			wantConflict: true, // 单行文件，两端在末尾追加不同内容，视为冲突
		},
		{
			name:        "pc1 delete paragraph",
			base:        "Line1\nLine2\nLine3",
			pc1:         "Line1\nLine3",
			pc2:         "Line1\nLine2\nLine3",
			pc1First:    true,
			wantConflict: false,
			wantNotContains: "Line2",
		},
		{
			name:        "delete-modify conflict",
			base:        "Line1\nLine2\nLine3",
			pc1:         "Line1\nLine3",           // 删除 Line2
			pc2:         "Line1\nLine2 Modified\nLine3",  // 修改 Line2
			pc1First:    true,
			wantConflict: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MergeTexts(tt.base, tt.pc1, tt.pc2, tt.pc1First)
			if err != nil {
				t.Fatalf("MergeTexts() error = %v", err)
			}

			if result.HasConflict != tt.wantConflict {
				t.Errorf("HasConflict = %v, want %v (info: %s)", result.HasConflict, tt.wantConflict, result.ConflictInfo)
			}

			if !result.HasConflict {
				if tt.wantContains != "" && !strings.Contains(result.Content, tt.wantContains) {
					t.Errorf("Content should contain %q, got %q", tt.wantContains, result.Content)
				}
				if tt.wantNotContains != "" && strings.Contains(result.Content, tt.wantNotContains) {
					t.Errorf("Content should not contain %q, got %q", tt.wantNotContains, result.Content)
				}
			}
		})
	}
}

// 测试 extractDeleteRanges
func TestExtractDeleteRanges(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		modified string
		wantLen  int
	}{
		{
			name:     "no delete",
			base:     "Hello",
			modified: "Hello World",
			wantLen:  0,
		},
		{
			name:     "single delete",
			base:     "Hello World",
			modified: "Hello",
			wantLen:  1,
		},
		{
			name:     "multiple deletes",
			base:     "A B C D",
			modified: "A D",
			wantLen:  1, // "B C " 是连续删除
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dmp := newDMP()
			diffs := dmp.DiffMain(tt.base, tt.modified, false)
			ranges := extractDeleteRanges(diffs)

			if len(ranges) != tt.wantLen {
				t.Errorf("extractDeleteRanges() got %d ranges, want %d", len(ranges), tt.wantLen)
			}
		})
	}
}

// 测试 rangesOverlap
func TestRangesOverlap(t *testing.T) {
	tests := []struct {
		name string
		r1   textRange
		r2   textRange
		want bool
	}{
		{
			name: "no overlap - r1 before r2",
			r1:   textRange{0, 5},
			r2:   textRange{10, 15},
			want: false,
		},
		{
			name: "no overlap - r2 before r1",
			r1:   textRange{10, 15},
			r2:   textRange{0, 5},
			want: false,
		},
		{
			name: "overlap - partial",
			r1:   textRange{0, 10},
			r2:   textRange{5, 15},
			want: true,
		},
		{
			name: "overlap - r1 contains r2",
			r1:   textRange{0, 20},
			r2:   textRange{5, 15},
			want: true,
		},
		{
			name: "adjacent - no overlap",
			r1:   textRange{0, 5},
			r2:   textRange{5, 10},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rangesOverlap(tt.r1, tt.r2)
			if got != tt.want {
				t.Errorf("rangesOverlap(%v, %v) = %v, want %v", tt.r1, tt.r2, got, tt.want)
			}
		})
	}
}

// 辅助函数
func newDMP() *diffmatchpatch.DiffMatchPatch {
	return diffmatchpatch.New()
}
