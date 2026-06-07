package oauth

import (
	"fmt"
	"strings"
)

var oauthScopeFunctions = map[string]string{
	"notes:read":  "note_r",
	"notes:write": "note_w",
	"files:read":  "file_r",
	"files:write": "file_w",
	"vaults:read": "note_r",
}

func MapOAuthScopesToFNS(client string, scopes []string) (string, error) {
	client = strings.TrimSpace(client)
	if client == "" {
		client = "*"
	}

	functions := make([]string, 0, len(scopes))
	seen := make(map[string]bool, len(scopes))
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		function, ok := oauthScopeFunctions[scope]
		if !ok {
			continue
		}
		if seen[function] {
			continue
		}
		seen[function] = true
		functions = append(functions, function)
	}

	if len(functions) == 0 {
		return "", fmt.Errorf("%w: no usable oauth scopes", ErrInsufficientScope)
	}

	return "p:mcp c:" + client + " f:" + strings.Join(functions, ","), nil
}
