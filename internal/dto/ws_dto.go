package dto

// WebSocketMsgType WebSocket Binary message type
// WebSocket 二进制消息类型
type WebSocketMsgType = string

// VaultFileMsgType vault attachment message
// 笔记库附件消息
const VaultFileMsgType WebSocketMsgType = "00"

// WebSocketReceiveAction WebSocket text receive action type
// WebSocket 文本接收动作类型
type WebSocketReceiveAction = string

// WebSocketSendAction WebSocket text send action type
// WebSocket 文本发送动作类型
type WebSocketSendAction = string

const (

	// ---------------- Folder ----------------

	// FolderReceiveSync folder synchronization request
	// FolderReceiveSync 文件夹同步请求
	FolderReceiveSync WebSocketReceiveAction = "FolderSync"
	// FolderReceiveModify folder modify or create request
	// FolderReceiveModify 文件夹修改或创建请求
	FolderReceiveModify WebSocketReceiveAction = "FolderModify"
	// FolderReceiveDelete folder delete request
	// FolderReceiveDelete 文件夹删除请求
	FolderReceiveDelete WebSocketReceiveAction = "FolderDelete"
	// FolderReceiveRename folder rename request
	// FolderReceiveRename 文件夹重命名请求
	FolderReceiveRename WebSocketReceiveAction = "FolderRename"

	// ---------------- Note ----------------

	// NoteReceiveSync note synchronization request
	// NoteReceiveSync 笔记同步请求
	NoteReceiveSync WebSocketReceiveAction = "NoteSync"
	// NoteReceiveModify note modify or create request
	// NoteReceiveModify 笔记修改或创建请求
	NoteReceiveModify WebSocketReceiveAction = "NoteModify"
	// NoteReceiveDelete note delete request
	// NoteReceiveDelete 笔记删除请求
	NoteReceiveDelete WebSocketReceiveAction = "NoteDelete"
	// NoteReceiveRename note rename request
	// NoteReceiveRename 笔记重命名请求
	NoteReceiveRename WebSocketReceiveAction = "NoteRename"
	// NoteReceiveCheck note modification check request
	// NoteReceiveCheck 笔记修改检查请求
	NoteReceiveCheck WebSocketReceiveAction = "NoteCheck"
	// NoteReceiveRePush Note missing pull request
	// NoteReceiveRePush 笔记缺失请求拉取
	NoteReceiveRePush WebSocketReceiveAction = "NoteRePush"

	// ---------------- File ----------------

	// FileReceiveSync file synchronization request
	// FileReceiveSync 文件同步请求
	FileReceiveSync WebSocketReceiveAction = "FileSync"
	// FileReceiveUploadCheck file upload pre-check request
	// FileReceiveUploadCheck 文件上传前检查请求
	FileReceiveUploadCheck WebSocketReceiveAction = "FileUploadCheck"
	// FileReceiveDelete file delete request
	// FileReceiveDelete 文件删除请求
	FileReceiveDelete WebSocketReceiveAction = "FileDelete"
	// FileReceiveRename file rename request
	// FileReceiveRename 文件重命名请求
	FileReceiveRename WebSocketReceiveAction = "FileRename"
	// FileReceiveChunkDownload file chunk download request
	// FileReceiveChunkDownload 文件分片下载请求
	FileReceiveChunkDownload WebSocketReceiveAction = "FileChunkDownload"
	// FileReceiveRePush file missing pull request
	// FileReceiveRePush 文件缺失请求拉取
	FileReceiveRePush WebSocketReceiveAction = "FileRePush"

	// ---------------- Setting ----------------

	// SettingReceiveSync setting synchronization request
	// SettingReceiveSync 设置同步请求
	SettingReceiveSync WebSocketReceiveAction = "SettingSync"
	// SettingReceiveModify setting modify or create request
	// SettingReceiveModify 设置修改或创建请求
	SettingReceiveModify WebSocketReceiveAction = "SettingModify"
	// SettingReceiveDelete setting delete request
	// SettingReceiveDelete 设置删除请求
	SettingReceiveDelete WebSocketReceiveAction = "SettingDelete"
	// SettingReceiveCheck setting modification check request
	// SettingReceiveCheck 设置修改检查请求
	SettingReceiveCheck WebSocketReceiveAction = "SettingCheck"
)

const (
	// ---------------- Folder ----------------

	// FolderSyncModify folder synchronization modification
	// FolderSyncModify 文件夹同步修改
	FolderSyncModify WebSocketSendAction = "FolderSyncModify"
	// FolderSyncDelete folder synchronization deletion
	// FolderSyncDelete 文件夹同步删除
	FolderSyncDelete WebSocketSendAction = "FolderSyncDelete"
	// FolderSyncEnd folder synchronization finished
	// FolderSyncEnd 文件夹同步结束
	FolderSyncEnd WebSocketSendAction = "FolderSyncEnd"
	// FolderRename folder rename action
	// FolderRename 文件夹重命名动作
	FolderSyncRename WebSocketSendAction = "FolderSyncRename"

	// ---------------- Note ----------------

	// NoteSyncModify note synchronization modification
	// NoteSyncModify 笔记同步修改
	NoteSyncModify WebSocketSendAction = "NoteSyncModify"
	// NoteSyncDelete note synchronization deletion
	// NoteSyncDelete 笔记同步删除
	NoteSyncDelete WebSocketSendAction = "NoteSyncDelete"
	// NoteSyncRename note synchronization rename
	// NoteSyncRename 笔记同步重命名
	NoteSyncRename WebSocketSendAction = "NoteSyncRename"
	// NoteSyncMtime note modification time synchronization
	// NoteSyncMtime 笔记修改时间同步
	NoteSyncMtime WebSocketSendAction = "NoteSyncMtime"
	// NoteSyncEnd note synchronization finished
	// NoteSyncEnd 笔记同步结束
	NoteSyncEnd WebSocketSendAction = "NoteSyncEnd"
	// NoteSyncNeedPush indicates client needs to push note content
	// NoteSyncNeedPush 表示客户端需要推送笔记内容
	NoteSyncNeedPush WebSocketSendAction = "NoteSyncNeedPush"

	// ---------------- File ----------------

	// FileSyncUpdate file synchronization update
	// FileSyncUpdate 文件同步更新
	FileSyncUpdate WebSocketSendAction = "FileSyncUpdate"
	// FileSyncDelete file synchronization deletion
	// FileSyncDelete 文件同步删除
	FileSyncDelete WebSocketSendAction = "FileSyncDelete"
	// FileSyncRename file synchronization rename
	// FileSyncRename 文件同步重命名
	FileSyncRename WebSocketSendAction = "FileSyncRename"
	// FileSyncMtime file modification time synchronization
	// FileSyncMtime 文件修改时间同步
	FileSyncMtime WebSocketSendAction = "FileSyncMtime"
	// FileSyncEnd file synchronization finished
	// FileSyncEnd 文件同步结束
	FileSyncEnd WebSocketSendAction = "FileSyncEnd"
	// FileUpload file upload action
	// FileUpload 文件上传动作
	FileUpload WebSocketSendAction = "FileUpload"
	// FileSyncChunkDownload file chunk download for sync
	// FileSyncChunkDownload 同步时的文件块下载
	FileSyncChunkDownload WebSocketSendAction = "FileSyncChunkDownload"

	// ---------------- Setting ----------------

	// SettingSyncModify setting synchronization modification
	// SettingSyncModify 设置同步修改
	SettingSyncModify WebSocketSendAction = "SettingSyncModify"
	// SettingSyncDelete setting synchronization deletion
	// SettingSyncDelete 设置同步删除
	SettingSyncDelete WebSocketSendAction = "SettingSyncDelete"
	// SettingSyncMtime setting modification time synchronization
	// SettingSyncMtime 设置修改时间同步
	SettingSyncMtime WebSocketSendAction = "SettingSyncMtime"
	// SettingSyncEnd setting synchronization finished
	// SettingSyncEnd 设置同步结束
	SettingSyncEnd WebSocketSendAction = "SettingSyncEnd"
	// SettingSyncNeedUpload indicates client needs to upload setting
	// SettingSyncNeedUpload 表示客户端需要上传设置
	SettingSyncNeedUpload WebSocketSendAction = "SettingSyncNeedUpload"
)

// WSQueuedMessage represents a message item to be sent
// Used to collect messages during sync process, and sent together in SyncEnd message
// WSQueuedMessage 表示待发送的消息项
// 用于在同步过程中收集消息,在 SyncEnd 消息中统一合并发送
type WSQueuedMessage struct {
	Action string `json:"action"` // Message type // 消息类型
	Data   any    `json:"data"`   // Message content // 消息内容
}
