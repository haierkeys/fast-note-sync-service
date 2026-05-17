package util

import "strings"

// VerifyVaultAccess checks if a target vault is allowed under the restricted allowedVaults list.
// VerifyVaultAccess 检查目标笔记库是否在允许访问的限制笔记库列表中。
// allowedVaults: comma-separated vaults allowed (empty means no restriction) // 允许访问的笔记库（逗号分隔，为空表示不限制）
// targetVault: target vault to check // 待检查的目标笔记库
// return: true if access is allowed, false otherwise // 允许访问返回 true，否则返回 false
func VerifyVaultAccess(allowedVaults string, targetVault string) bool {
	if allowedVaults == "" {
		return true
	}
	if targetVault == "" {
		return false
	}
	vaults := strings.Split(allowedVaults, ",")
	for _, v := range vaults {
		if strings.TrimSpace(v) == targetVault {
			return true
		}
	}
	return false
}
