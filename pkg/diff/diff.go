package diff

import (
	"github.com/sergi/go-diff/diffmatchpatch"
)

// MergeResult 合并结果
type MergeResult struct {
	Content      string // 合并后的内容
	HasConflict  bool   // 是否存在冲突
	ConflictInfo string // 冲突详情
}

// textRange 表示文本中的一个区域
type textRange struct {
	Start int
	End   int
}

// insertInfo 插入操作的详细信息
type insertInfo struct {
	Position int    // 插入位置
	Content  string // 插入内容
}

// MergeTexts 三方合并文本
// 重构后保留删除操作，检测删除-修改冲突
func MergeTexts(base, pc1, pc2 string, pc1First bool) (MergeResult, error) {
	// 快速路径：如果两端修改结果相同，直接返回
	if pc1 == pc2 {
		return MergeResult{
			Content:     pc1,
			HasConflict: false,
		}, nil
	}

	// 快速路径：如果一端没有修改，直接返回另一端
	if pc1 == base {
		return MergeResult{
			Content:     pc2,
			HasConflict: false,
		}, nil
	}
	if pc2 == base {
		return MergeResult{
			Content:     pc1,
			HasConflict: false,
		}, nil
	}

	dmp := diffmatchpatch.New()

	// 计算 PC1 相对于 base 的 diff（保留删除操作）
	pc1Diffs := dmp.DiffMain(base, pc1, false)
	pc1Patches := dmp.PatchMake(base, pc1Diffs)

	// 计算 PC2 相对于 base 的 diff（保留删除操作）
	pc2Diffs := dmp.DiffMain(base, pc2, false)
	pc2Patches := dmp.PatchMake(base, pc2Diffs)

	// 检测冲突
	if hasConflict(pc1Diffs, pc2Diffs) {
		return MergeResult{
			HasConflict:  true,
			ConflictInfo: "conflict detected",
		}, nil
	}

	// 根据 pc1First 参数决定应用顺序
	var step1Result string
	var step1Success []bool
	var step2Success []bool
	var merged string

	if pc1First {
		// 先应用 PC1，再应用 PC2
		step1Result, step1Success = dmp.PatchApply(pc1Patches, base)
		merged, step2Success = dmp.PatchApply(pc2Patches, step1Result)
	} else {
		// 先应用 PC2，再应用 PC1
		step1Result, step1Success = dmp.PatchApply(pc2Patches, base)
		merged, step2Success = dmp.PatchApply(pc1Patches, step1Result)
	}

	// 检查补丁应用是否成功
	for _, s := range step1Success {
		if !s {
			return MergeResult{
				HasConflict:  true,
				ConflictInfo: "patch apply failed in step 1",
			}, nil
		}
	}
	for _, s := range step2Success {
		if !s {
			return MergeResult{
				HasConflict:  true,
				ConflictInfo: "patch apply failed in step 2",
			}, nil
		}
	}

	return MergeResult{
		Content:     merged,
		HasConflict: false,
	}, nil
}


// hasConflict 检测合并冲突
// 采用基于行的冲突检测策略，更符合文本编辑的语义
//
// 冲突情况：
// 1. 两方都修改了同一行（修改-修改冲突）
// 2. 一方删除某行，另一方修改同一行（删除-修改冲突）
// 3. 两方从空文件开始各自添加不同内容（空文件冲突）
// 4. 两方在同一行末尾追加不同内容（追加冲突）
//
// 非冲突情况：
// - 两方在文件末尾各自添加新行（可以合并）
// - 两方删除相同行（结果一致）
// - 两方修改结果相同（结果一致）
func hasConflict(pc1Diffs, pc2Diffs []diffmatchpatch.Diff) bool {
	// 提取行级别的变更
	pc1Changes := extractLineChangesFromDiffs(pc1Diffs)
	pc2Changes := extractLineChangesFromDiffs(pc2Diffs)

	// 特殊情况：两方都只是在末尾添加新行，不是冲突
	pc1OnlyAppend := isOnlyAppendAtEnd(pc1Changes)
	pc2OnlyAppend := isOnlyAppendAtEnd(pc2Changes)
	if pc1OnlyAppend && pc2OnlyAppend {
		return false
	}

	// 检查是否有行级别的冲突
	for lineNum, change1 := range pc1Changes {
		if change2, exists := pc2Changes[lineNum]; exists {
			// 两方都操作了同一行

			// 如果两方都是在末尾添加新行（以换行符开头），不是冲突
			if change1.changeType == lineInserted && change2.changeType == lineInserted &&
				change1.isAtEnd && change2.isAtEnd &&
				len(change1.newContent) > 0 && change1.newContent[0] == '\n' &&
				len(change2.newContent) > 0 && change2.newContent[0] == '\n' {
				continue
			}

			// 如果两方都是删除同一行，不是冲突
			if change1.changeType == lineDeleted && change2.changeType == lineDeleted {
				continue
			}

			// 如果两方修改/插入结果相同，不是冲突
			if change1.newContent == change2.newContent {
				continue
			}

			// 其他情况都是冲突
			return true
		}
	}

	return false
}

// lineChangeType 行变更类型
type lineChangeType int

const (
	lineModified lineChangeType = iota // 行被修改
	lineDeleted                        // 行被删除
	lineInserted                       // 新插入的行
)

// lineChange 表示对某一行的变更
type lineChange struct {
	changeType lineChangeType
	newContent string // 修改后的内容（如果是修改或插入）
	isAtEnd    bool   // 是否是在文件末尾的操作
}

// extractLineChangesFromDiffs 从 diff 中提取行级别的变更
func extractLineChangesFromDiffs(diffs []diffmatchpatch.Diff) map[int]lineChange {
	changes := make(map[int]lineChange)
	lineNum := 0
	totalLines := 0

	// 首先计算原文总行数
	for _, d := range diffs {
		if d.Type == diffmatchpatch.DiffEqual || d.Type == diffmatchpatch.DiffDelete {
			for _, ch := range d.Text {
				if ch == '\n' {
					totalLines++
				}
			}
		}
	}

	// 遍历 diff，记录每行的变更
	for i, d := range diffs {
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			for _, ch := range d.Text {
				if ch == '\n' {
					lineNum++
				}
			}
		case diffmatchpatch.DiffDelete:
			startLine := lineNum
			deletedText := d.Text

			// 检查是否紧跟插入（表示修改）
			isModify := i+1 < len(diffs) && diffs[i+1].Type == diffmatchpatch.DiffInsert
			var newContent string
			if isModify {
				newContent = diffs[i+1].Text
			}

			// 记录删除/修改影响的每一行
			currentLine := startLine
			for _, ch := range deletedText {
				if isModify {
					changes[currentLine] = lineChange{
						changeType: lineModified,
						newContent: newContent,
						isAtEnd:    currentLine >= totalLines-1,
					}
				} else {
					changes[currentLine] = lineChange{
						changeType: lineDeleted,
						isAtEnd:    currentLine >= totalLines-1,
					}
				}
				if ch == '\n' {
					currentLine++
					lineNum++
				}
			}
			// 如果删除的文本不以换行结尾，也要记录当前行
			if len(deletedText) > 0 && deletedText[len(deletedText)-1] != '\n' {
				if isModify {
					changes[currentLine] = lineChange{
						changeType: lineModified,
						newContent: newContent,
						isAtEnd:    currentLine >= totalLines-1,
					}
				} else {
					changes[currentLine] = lineChange{
						changeType: lineDeleted,
						isAtEnd:    currentLine >= totalLines-1,
					}
				}
			}

		case diffmatchpatch.DiffInsert:
			// 纯插入（前面不是删除）
			if i == 0 || diffs[i-1].Type != diffmatchpatch.DiffDelete {
				// 在当前行位置插入
				changes[lineNum] = lineChange{
					changeType: lineInserted,
					newContent: d.Text,
					isAtEnd:    lineNum >= totalLines,
				}
			}
		}
	}

	return changes
}

// isOnlyAppendAtEnd 检查变更是否只是在文件末尾添加新行
// 注意：在现有行末尾追加内容不算"末尾添加新行"
func isOnlyAppendAtEnd(changes map[int]lineChange) bool {
	if len(changes) == 0 {
		return false
	}
	for _, change := range changes {
		// 必须是插入类型
		if change.changeType != lineInserted {
			return false
		}
		// 必须是在末尾
		if !change.isAtEnd {
			return false
		}
		// 插入的内容必须以换行符开头（表示添加新行）
		// 如果不以换行符开头，说明是在现有行末尾追加内容
		if len(change.newContent) == 0 || change.newContent[0] != '\n' {
			return false
		}
	}
	return true
}

// extractInsertInfos 从 diff 列表中提取纯插入操作的位置和内容
// 纯插入是指前面没有删除操作的插入（不是替换的一部分）
// 注意：此函数保留用于其他用途，但不再用于冲突检测
func extractInsertInfos(diffs []diffmatchpatch.Diff) []insertInfo {
	var infos []insertInfo
	pos := 0

	for i, d := range diffs {
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			pos += len(d.Text)
		case diffmatchpatch.DiffDelete:
			pos += len(d.Text)
		case diffmatchpatch.DiffInsert:
			// 只有前一个不是删除时，才是纯插入
			if i == 0 || diffs[i-1].Type != diffmatchpatch.DiffDelete {
				infos = append(infos, insertInfo{
					Position: pos,
					Content:  d.Text,
				})
			}
		}
	}

	return infos
}

// extractDeleteRanges 从 diff 列表中提取删除操作的位置范围
// 注意：此函数保留用于其他用途
func extractDeleteRanges(diffs []diffmatchpatch.Diff) []textRange {
	var ranges []textRange
	pos := 0

	for _, d := range diffs {
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			pos += len(d.Text)
		case diffmatchpatch.DiffDelete:
			ranges = append(ranges, textRange{
				Start: pos,
				End:   pos + len(d.Text),
			})
			pos += len(d.Text)
		case diffmatchpatch.DiffInsert:
			// 插入操作不改变在原文中的位置
		}
	}

	return ranges
}

// extractModifyRanges 从 diff 列表中提取修改操作的位置范围
// 修改被定义为：删除后紧跟插入
// 注意：此函数保留用于其他用途
func extractModifyRanges(diffs []diffmatchpatch.Diff) []textRange {
	var ranges []textRange
	pos := 0

	for i, d := range diffs {
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			pos += len(d.Text)
		case diffmatchpatch.DiffDelete:
			// 检查是否紧跟插入（表示修改）
			if i+1 < len(diffs) && diffs[i+1].Type == diffmatchpatch.DiffInsert {
				ranges = append(ranges, textRange{
					Start: pos,
					End:   pos + len(d.Text),
				})
			}
			pos += len(d.Text)
		case diffmatchpatch.DiffInsert:
			// 插入操作不改变在原文中的位置
		}
	}

	return ranges
}

// rangesOverlap 检查两个范围是否重叠
func rangesOverlap(r1, r2 textRange) bool {
	return r1.Start < r2.End && r2.Start < r1.End
}
