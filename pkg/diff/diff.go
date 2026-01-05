package diff

import "github.com/sergi/go-diff/diffmatchpatch"

func MergeTexts(base, pc1, pc2 string, pc1First bool) (merged string, success bool) {
	dmp := diffmatchpatch.New()

	// 计算 PC1 相对于 base 的 diff,并过滤删除操作
	pc1Diffs := dmp.DiffMain(base, pc1, false)
	pc1DiffsNoDelete := make([]diffmatchpatch.Diff, 0)
	for _, diff := range pc1Diffs {
		if diff.Type != diffmatchpatch.DiffDelete {
			pc1DiffsNoDelete = append(pc1DiffsNoDelete, diff)
		}
	}
	pc1Patches := dmp.PatchMake(base, pc1DiffsNoDelete)

	// 计算 PC2 相对于 base 的 diff,并过滤删除操作
	pc2Diffs := dmp.DiffMain(base, pc2, false)
	pc2DiffsNoDelete := make([]diffmatchpatch.Diff, 0)
	for _, diff := range pc2Diffs {
		if diff.Type != diffmatchpatch.DiffDelete {
			pc2DiffsNoDelete = append(pc2DiffsNoDelete, diff)
		}
	}
	pc2Patches := dmp.PatchMake(base, pc2DiffsNoDelete)

	// 根据 pc1First 参数决定应用顺序
	var step1Result string
	var step1Success []bool
	var step2Success []bool

	if pc1First {
		// 先应用 PC1,再应用 PC2
		step1Result, step1Success = dmp.PatchApply(pc1Patches, base)
		merged, step2Success = dmp.PatchApply(pc2Patches, step1Result)
	} else {
		// 先应用 PC2,再应用 PC1
		step1Result, step1Success = dmp.PatchApply(pc2Patches, base)
		merged, step2Success = dmp.PatchApply(pc1Patches, step1Result)
	}

	// 检查是否所有补丁都成功应用
	success = true
	for _, s := range step1Success {
		if !s {
			success = false
			break
		}
	}
	if success {
		for _, s := range step2Success {
			if !s {
				success = false
				break
			}
		}
	}

	return merged, success
}
