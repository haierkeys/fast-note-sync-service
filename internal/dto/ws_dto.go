package dto

// WebSocketMsgType WebSocket Binary message type
// WebSocket 二进制消息类型
type WebSocketMsgType = string

// VaultFileMsgType vault attachment message
// 笔记库附件消息
const VaultFileMsgType WebSocketMsgType = "00"

// WebSocketAction WebSocket text message type
// WebSocket 文本消息类型
type WebSocketAction = string

const (
	// Folder related
	// 文件夹相关

	// FolderSyncModify folder synchronization modification
	// FolderSyncModify 文件夹同步修改
	FolderSyncModify WebSocketAction = "FolderSyncModify"
	// FolderSyncDelete folder synchronization deletion
	// FolderSyncDelete 文件夹同步删除
	FolderSyncDelete WebSocketAction = "FolderSyncDelete"
	// FolderSyncEnd folder synchronization finished
	// FolderSyncEnd 文件夹同步结束
	FolderSyncEnd WebSocketAction = "FolderSyncEnd"


	// Note related
	// 笔记相关

	// NoteSyncModify note synchronization modification
	// NoteSyncModify 笔记同步修改
	NoteSyncModify WebSocketAction = "NoteSyncModify"
	// NoteSyncDelete note synchronization deletion
	// NoteSyncDelete 笔记同步删除
	NoteSyncDelete WebSocketAction = "NoteSyncDelete"
	// NoteSyncMtime note modification time synchronization
	// NoteSyncMtime 笔记修改时间同步
	NoteSyncMtime WebSocketAction = "NoteSyncMtime"
	// NoteSyncEnd note synchronization finished
	// NoteSyncEnd 笔记同步结束
	NoteSyncEnd WebSocketAction = "NoteSyncEnd"
	// NoteSyncNeedPush indicates client needs to push note content
	// NoteSyncNeedPush 表示客户端需要推送笔记内容
	NoteSyncNeedPush WebSocketAction = "NoteSyncNeedPush"

	// File related
	// 文件/附件相关

	// FileSyncUpdate file synchronization update
	// FileSyncUpdate 文件同步更新
	FileSyncUpdate WebSocketAction = "FileSyncUpdate"
	// FileSyncDelete file synchronization deletion
	// FileSyncDelete 文件同步删除
	FileSyncDelete WebSocketAction = "FileSyncDelete"
	// FileSyncMtime file modification time synchronization
	// FileSyncMtime 文件修改时间同步
	FileSyncMtime WebSocketAction = "FileSyncMtime"
	// FileSyncEnd file synchronization finished
	// FileSyncEnd 文件同步结束
	FileSyncEnd WebSocketAction = "FileSyncEnd"
	// FileUpload file upload action
	// FileUpload 文件上传动作
	FileUpload WebSocketAction = "FileUpload"
	// FileSyncChunkDownload file chunk download for sync
	// FileSyncChunkDownload 同步时的文件块下载
	FileSyncChunkDownload WebSocketAction = "FileSyncChunkDownload"

	// Setting related
	// 设置相关

	// SettingSyncModify setting synchronization modification
	// SettingSyncModify 设置同步修改
	SettingSyncModify WebSocketAction = "SettingSyncModify"
	// SettingSyncDelete setting synchronization deletion
	// SettingSyncDelete 设置同步删除
	SettingSyncDelete WebSocketAction = "SettingSyncDelete"
	// SettingSyncMtime setting modification time synchronization
	// SettingSyncMtime 设置修改时间同步
	SettingSyncMtime WebSocketAction = "SettingSyncMtime"
	// SettingSyncEnd setting synchronization finished
	// SettingSyncEnd 设置同步结束
	SettingSyncEnd WebSocketAction = "SettingSyncEnd"
	// SettingSyncNeedUpload indicates client needs to upload setting
	// SettingSyncNeedUpload 表示客户端需要上传设置
	SettingSyncNeedUpload WebSocketAction = "SettingSyncNeedUpload"
)
