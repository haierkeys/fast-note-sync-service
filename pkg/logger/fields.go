package logger

// 统一的日志字段命名常量
// 用于确保整个项目中日志字段命名的一致性，便于日志查询和分析
const (
	// FieldTraceID 追踪 ID 字段
	FieldTraceID = "traceId"

	// FieldUID 用户 ID 字段
	FieldUID = "uid"

	// FieldAction 操作类型字段
	FieldAction = "action"

	// FieldPath 文件路径字段
	FieldPath = "path"

	// FieldVault 仓库名称字段
	FieldVault = "vault"

	// FieldDuration 耗时字段
	FieldDuration = "duration"

	// FieldSessionID 会话 ID 字段
	FieldSessionID = "sessionId"

	// FieldMethod 方法名称字段
	FieldMethod = "method"

	// FieldError 错误信息字段
	FieldError = "error"

	// FieldSize 文件大小字段
	FieldSize = "size"

	// FieldChunks 分块数量字段
	FieldChunks = "chunks"

	// FieldBucket 存储桶名称字段
	FieldBucket = "bucket"

	// FieldFileKey 文件键字段
	FieldFileKey = "fileKey"
)
