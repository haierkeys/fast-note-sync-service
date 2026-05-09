package app

import (
	"strings"
)

// VerifyPermissions verifies if the given scope matches the required protocol, client, and function.
// VerifyPermissions 验证给定的作用域是否匹配所需的协议、客户端和功能。
// scope: The permission string (e.g. "p:rest c:webgui f:note_r") // 权限范围
// p: The protocol of the current request (e.g. "rest", "ws", "mcp") // 当前请求的协议
// c: The client of the current request (e.g. "webgui", "obsidian", "mobile") // 当前请求的客户端
// f: The function being accessed (e.g. "note_r", "note_w"). If empty, it means the resource is not restricted by function level. // 访问的功能。如果为空，表示该资源不受功能级别限制。
func VerifyPermissions(scope string, p string, c string, f string) bool {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		return true
	}

	parts := strings.Split(scope, " ")
	
	var scopeP, scopeC, scopeF string

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, "p:") {
			scopeP = part[2:]
		} else if strings.HasPrefix(part, "c:") {
			scopeC = part[2:]
		} else if strings.HasPrefix(part, "f:") {
			scopeF = part[2:]
		}
	}

	matchP := (scopeP == "*" || strings.EqualFold(scopeP, p))
	matchC := (scopeC == "*" || strings.EqualFold(scopeC, c))
	
	matchF := true
	if f != "" {
		matchF = (scopeF == "*" || strings.EqualFold(scopeF, f))
	}

	return matchP && matchC && matchF
}
