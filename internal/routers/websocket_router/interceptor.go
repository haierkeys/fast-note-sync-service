package websocket_router

import (
	"strings"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/json"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewMessageInterceptor 创建 WebSocket 业务消息前置拦截器，依次执行认证、Vault 和 RBAC 检查。
// NewMessageInterceptor creates a WebSocket business message pre-handler interceptor
// that sequentially enforces auth, vault access, and RBAC checks.
func NewMessageInterceptor(appContainer *app.App) func(*pkgapp.WebsocketClient, *pkgapp.WebSocketMessage) bool {
	logger := appContainer.Logger()
	return func(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) bool {
		if !checkAuth(c, msg, logger) {
			return false
		}
		if !checkVaultAccess(c, msg, logger) {
			return false
		}
		if !checkRBAC(c, msg, logger) {
			return false
		}
		return true
	}
}

// checkAuth 验证用户是否已完成身份认证。
// checkAuth verifies that the client has a valid authenticated user session.
func checkAuth(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage, logger interface {
	Warn(string, ...zapcore.Field)
}) bool {
	if c.User == nil {
		logger.Warn("WS User not authenticated",
			zap.String("msgType", msg.Type),
			zap.String("traceId", c.TraceID))
		c.ToResponse(code.ErrorNotUserAuthToken)
		return false
	}
	return true
}

// checkVaultAccess 针对活跃 WebSocket 连接校验笔记库访问权限限制。
// checkVaultAccess validates that the requested vault is within the client's allowed scope.
func checkVaultAccess(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage, logger interface {
	Warn(string, ...zapcore.Field)
}) bool {
	if c.Vaults == "" {
		return true
	}
	var vaultInfo struct {
		Vault string `json:"vault"`
	}
	if err := json.Unmarshal(msg.Data, &vaultInfo); err != nil || vaultInfo.Vault == "" {
		return true
	}
	if !util.VerifyVaultAccess(c.Vaults, vaultInfo.Vault) {
		logger.Warn("WS OnMessage Vault Restricted",
			zap.String("Type", msg.Type),
			zap.String("uid", c.User.ID),
			zap.String("vault", vaultInfo.Vault))
		c.ToResponse(code.ErrorAuthTokenScopeRestricted.WithDetails("Vault access restricted: "+vaultInfo.Vault), msg.Type+"Ack")
		return false
	}
	return true
}

// checkRBAC 将操作映射到 RBAC 权限功能点并校验客户端权限，权限不足时触发回滚。
// checkRBAC maps the message type to an RBAC function and verifies the client's scope.
// On denial it delegates to handlePermissionDenied for client-side rollback.
func checkRBAC(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage, logger interface {
	Warn(string, ...zapcore.Field)
	Info(string, ...zapcore.Field)
}) bool {
	function := resolveRBACFunction(msg.Type)
	if function == "" {
		return true // 无需权限检查的消息类型，直接放行 / message type requires no RBAC check
	}
	if pkgapp.VerifyPermissions(c.Scope, "ws", c.ClientType(), function) {
		return true
	}
	logger.Warn("WS OnMessage Permission Denied",
		zap.String("Type", msg.Type),
		zap.String("uid", c.User.ID),
		zap.String("function", function))
	return handlePermissionDenied(c, msg, function, logger)
}

// resolveRBACFunction 将 WebSocket 消息类型映射到 RBAC 功能点字符串。
// resolveRBACFunction maps a WebSocket message type to its corresponding RBAC function key.
// Returns an empty string if no permission check is required for the given type.
func resolveRBACFunction(msgType string) string {
	switch msgType {
	case NoteReceiveSync, NoteReceiveCheck, NoteReceiveRePush, FolderReceiveSync:
		return "note_r"
	case NoteReceiveModify, NoteReceiveDelete, NoteReceiveRename, FolderReceiveModify, FolderReceiveDelete, FolderReceiveRename:
		return "note_w"
	case FileReceiveChunkDownload, FileReceiveRePush, FileReceiveSync:
		return "file_r"
	case FileReceiveUploadCheck, FileReceiveDelete, FileReceiveRename:
		return "file_w"
	case SettingReceiveSync, SettingReceiveCheck, SettingReceiveRePush:
		return "config_r"
	case SettingReceiveModify, SettingReceiveDelete, SettingReceiveClear:
		return "config_w"
	}
	return ""
}

// handlePermissionDenied 在权限拒绝后向客户端发送错误响应，并对写操作触发回滚同步。
// handlePermissionDenied sends an error response and, for write operations,
// triggers a compensating action (rename-back or re-push) to keep the client consistent.
// Always returns false to halt message processing.
func handlePermissionDenied(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage, function string, logger interface {
	Info(string, ...zapcore.Field)
}) bool {
	resPath := resolveResourcePath(msg)
	c.ToResponse(code.ErrorAuthTokenScopeRestricted.WithDetails("Permission denied: "+resPath), msg.Type+"Ack")

	if !strings.HasSuffix(function, "_w") {
		return false
	}

	// 写操作：优先尝试重命名回滚，失败则触发重推
	// Write operations: attempt rename-rollback first, fall back to re-push
	if strings.HasSuffix(msg.Type, "Rename") && rollbackRename(c, msg, function, logger) {
		return false
	}
	triggerRePush(c, msg, function, logger)
	return false
}

// resolveResourcePath 从消息数据中提取资源路径用于错误描述，优先取 path，其次取 name，最后回退到消息类型。
// resolveResourcePath extracts a human-readable resource identifier from message data,
// falling back to the message type when neither path nor name are present.
func resolveResourcePath(msg *pkgapp.WebSocketMessage) string {
	var pathInfo struct {
		Path string `json:"path"`
		Name string `json:"name"`
	}
	_ = json.Unmarshal(msg.Data, &pathInfo)
	if pathInfo.Path != "" {
		return pathInfo.Path
	}
	if pathInfo.Name != "" {
		return pathInfo.Name
	}
	return msg.Type
}

// rollbackRename 对重命名操作发送反向重命名消息，使客户端回退到原始路径。
// rollbackRename sends a compensating rename message so the client reverts to the original path.
// Returns true if the rollback message was successfully dispatched.
func rollbackRename(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage, function string, logger interface {
	Info(string, ...zapcore.Field)
}) bool {
	var renameData map[string]interface{}
	if err := json.Unmarshal(msg.Data, &renameData); err != nil {
		return false
	}

	vault, _ := renameData["vault"].(string)
	newPath, _ := renameData["path"].(string)
	newPathHash, _ := renameData["pathHash"].(string)
	oldPath, _ := renameData["oldPath"].(string)
	oldPathHash, _ := renameData["oldPathHash"].(string)

	if newPath == "" || oldPath == "" {
		return false
	}

	syncRenameAction := resolveSyncRenameAction(function, msg.Type)
	if syncRenameAction == "" {
		return false
	}

	rollbackData := map[string]interface{}{
		"path":        oldPath,
		"pathHash":    oldPathHash,
		"oldPath":     newPath,
		"oldPathHash": newPathHash,
	}
	c.ToResponse(code.Success.WithData(rollbackData).WithVault(vault), syncRenameAction)
	// 重命名回滚后无需再触发 RePush / No subsequent re-push needed after rename rollback
	return true
}

// resolveSyncRenameAction 根据 RBAC 功能点和消息类型确定用于回滚的同步重命名动作名称。
// resolveSyncRenameAction returns the compensating sync-rename action name for a given function and message type.
func resolveSyncRenameAction(function, msgType string) string {
	switch function {
	case "note_w":
		if strings.Contains(msgType, "Folder") {
			return FolderSyncRename
		}
		return NoteSyncRename
	case "file_w":
		return FileSyncRename
	}
	return ""
}

// triggerRePush 对写操作权限拒绝后触发对应的重推动作，保证客户端数据一致性。
// triggerRePush invokes the corresponding re-push handler to restore client-side consistency
// after a write operation is denied.
func triggerRePush(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage, function string, logger interface {
	Info(string, ...zapcore.Field)
}) {
	rePushAction := resolveRePushAction(function)
	if rePushAction == "" {
		return
	}
	if h, ok := c.Server.GetHandler(rePushAction); ok {
		logger.Info("WS Trigger RePush on permission denied",
			zap.String("action", rePushAction),
			zap.String("uid", c.User.ID))
		h(c, msg)
	}
}

// resolveRePushAction 根据 RBAC 功能点返回对应的重推动作名称。
// resolveRePushAction returns the re-push action name associated with the given RBAC write function.
func resolveRePushAction(function string) string {
	switch function {
	case "note_w":
		return NoteReceiveRePush
	case "file_w":
		return FileReceiveRePush
	case "config_w":
		return SettingReceiveRePush
	}
	return ""
}
