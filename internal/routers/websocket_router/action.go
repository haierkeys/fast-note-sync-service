package websocket_router

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
	// ClientReceiveInfo client info action
	// ClientReceiveInfo 客户端信息接收动作
	ClientReceiveInfo WebSocketReceiveAction = "ClientInfo"
	// ClientReceiveAuth client authorization action
	// ClientReceiveAuth 客户端鉴权接收动作
	ClientReceiveAuth WebSocketReceiveAction = "Authorization"

	// ClientInfo client info ack action
	// ClientInfo 客户端信息确认发送动作
	ClientInfo WebSocketSendAction = "ClientInfo"

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
	// SettingReceiveClear clear all settings request
	// SettingReceiveClear 清理所有设置请求
	SettingReceiveClear WebSocketReceiveAction = "SettingClear"
	// SettingReceiveRePush setting missing pull request
	// SettingReceiveRePush 配置缺失请求拉取
	SettingReceiveRePush WebSocketReceiveAction = "SettingRePush"
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
	// FolderModifyAck folder modify operation ack
	// FolderModifyAck 文件夹修改操作 ack
	FolderModifyAck WebSocketSendAction = "FolderModifyAck"
	// FolderRenameAck folder rename operation ack
	// FolderRenameAck 文件夹重命名操作 ack
	FolderRenameAck WebSocketSendAction = "FolderRenameAck"
	// FolderDeleteAck folder delete operation ack
	// FolderDeleteAck 文件夹删除操作 ack
	FolderDeleteAck WebSocketSendAction = "FolderDeleteAck"
	// FolderSyncBatchAck folder sync batch receive ack
	// FolderSyncBatchAck 文件夹分批同步接收确认，服务端接收到中间批次后发回客户端
	FolderSyncBatchAck WebSocketSendAction = "FolderSyncBatchAck"

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
	// NoteModifyAck note modify operation ack
	// NoteModifyAck 笔记修改操作 ack
	NoteModifyAck WebSocketSendAction = "NoteModifyAck"
	// NoteRenameAck note rename operation ack
	// NoteRenameAck 笔记重命名操作 ack
	NoteRenameAck WebSocketSendAction = "NoteRenameAck"
	// NoteDeleteAck note delete operation ack
	// NoteDeleteAck 笔记删除操作 ack
	NoteDeleteAck WebSocketSendAction = "NoteDeleteAck"
	// NoteSyncBatchAck note sync batch receive ack
	// NoteSyncBatchAck 笔记分批同步接收确认，服务端接收到中间批次后发回客户端
	NoteSyncBatchAck WebSocketSendAction = "NoteSyncBatchAck"

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
	// FileRenameAck file rename operation ack
	// FileRenameAck 文件重命名操作 ack
	FileRenameAck WebSocketSendAction = "FileRenameAck"
	// FileUploadAck file upload complete ack
	// FileUploadAck 文件上传完成 ack
	FileUploadAck WebSocketSendAction = "FileUploadAck"
	// FileDeleteAck file delete operation ack
	// FileDeleteAck 文件删除操作 ack
	FileDeleteAck WebSocketSendAction = "FileDeleteAck"
	// FileSyncBatchAck file sync batch receive ack
	// FileSyncBatchAck 附件分批同步接收确认，服务端接收到中间批次后发回客户端
	FileSyncBatchAck WebSocketSendAction = "FileSyncBatchAck"

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
	// SettingSyncClear sync clear all settings
	// SettingSyncClear 同步清理所有设置
	SettingSyncClear WebSocketSendAction = "SettingSyncClear"
	// SettingModifyAck setting modify operation ack
	// SettingModifyAck 设置修改操作 ack
	SettingModifyAck WebSocketSendAction = "SettingModifyAck"
	// SettingDeleteAck setting delete operation ack
	// SettingDeleteAck 设置删除操作 ack
	SettingDeleteAck WebSocketSendAction = "SettingDeleteAck"
	// SettingSyncBatchAck setting sync batch receive ack
	// SettingSyncBatchAck 配置分批同步接收确认，服务端接收到中间批次后发回客户端
	SettingSyncBatchAck WebSocketSendAction = "SettingSyncBatchAck"

	// ---------------- Share ----------------

	// ShareSyncRefresh notify clients to refresh share state
	// ShareSyncRefresh 通知客户端刷新分享状态
	ShareSyncRefresh WebSocketSendAction = "ShareSyncRefresh"

	// ---------------- Page Sync ----------------

	// NoteSyncPage note sync page message
	// NoteSyncPage 笔记同步分页消息
	NoteSyncPage WebSocketSendAction = "NoteSyncPage"
	// FileSyncPage file sync page message
	// FileSyncPage 文件同步分页消息
	FileSyncPage WebSocketSendAction = "FileSyncPage"
	// SettingSyncPage setting sync page message
	// SettingSyncPage 配置同步分页消息
	SettingSyncPage WebSocketSendAction = "SettingSyncPage"
	// FolderSyncPage folder sync page message
	// FolderSyncPage 文件夹同步分页消息
	FolderSyncPage WebSocketSendAction = "FolderSyncPage"

	// NoteSyncPageAck note sync page ack request
	// NoteSyncPageAck 笔记同步分页确认接收
	NoteSyncPageAck WebSocketReceiveAction = "NoteSyncPageAck"
	// FileSyncPageAck file sync page ack request
	// FileSyncPageAck 文件同步分页确认接收
	FileSyncPageAck WebSocketReceiveAction = "FileSyncPageAck"
	// SettingSyncPageAck setting sync page ack request
	// SettingSyncPageAck 配置同步分页确认接收
	SettingSyncPageAck WebSocketReceiveAction = "SettingSyncPageAck"
	// FolderSyncPageAck folder sync page ack request
	// FolderSyncPageAck 文件夹同步分页确认接收
	FolderSyncPageAck WebSocketReceiveAction = "FolderSyncPageAck"
)

// WSQueuedMessage represents a message item to be sent
// WSQueuedMessage used to collect messages during sync process, and sent together in SyncEnd message
// WSQueuedMessage 表示待发送的消息项
// WSQueuedMessage 用于在同步过程中收集消息,在 SyncEnd 消息中统一合并发送
type WSQueuedMessage struct {
	Action  string `json:"action"`  // Message action/type // 消息动作/类型
	Data    any    `json:"data"`    // Message data payload // 消息数据负载
	Context string `json:"context"` // Context // 上下文
}
