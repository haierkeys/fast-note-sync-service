package logger

// Unified log field naming constants
// 统一的日志字段命名常量
// Used to ensure consistency of log field naming across the project, facilitating log query and analysis
// 用于确保整个项目中日志字段命名的一致性，便于日志查询和分析
const (
	// FieldTraceID Trace ID field // 追踪 ID 字段
	FieldTraceID = "traceId"

	// FieldUID User ID field // 用户 ID 字段
	FieldUID = "uid"

	// FieldAction Action type field // 操作类型字段
	FieldAction = "action"

	// FieldPath File path field // 文件路径字段
	FieldPath = "path"

	// FieldVault Vault name field // 仓库名称字段
	FieldVault = "vault"

	// FieldDuration Time elapsed field // 耗时字段
	FieldDuration = "duration"

	// FieldSessionID Session ID field // 会话 ID 字段
	FieldSessionID = "sessionId"

	// FieldMethod Method name field // 方法名称字段
	FieldMethod = "method"

	// FieldError Error message field // 错误信息字段
	FieldError = "error"

	// FieldSize File size field // 文件大小字段
	FieldSize = "size"

	// FieldChunks Chunks count field // 分块数量字段
	FieldChunks = "chunks"

	// FieldBucket Storage bucket name field // 存储桶名称字段
	FieldBucket = "bucket"

	// FieldFileKey File key field // 文件键字段
	FieldFileKey = "fileKey"
)
