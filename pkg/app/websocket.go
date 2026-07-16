package app

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/json"
	"github.com/haierkeys/fast-note-sync-service/pkg/logger"
	"github.com/haierkeys/fast-note-sync-service/pkg/safego"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"golang.org/x/sync/singleflight"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	validatorV10 "github.com/go-playground/validator/v10"
	"github.com/lxzan/gws"
	"go.uber.org/zap"
)

type LogType string

const (
	WSPingInterval         = 25
	WSPingWait             = 60
	WSPingWriteTimeout     = 10      // WritePing write timeout (seconds), must < PingInterval // WritePing 写超时（秒），需小于 PingInterval
	LogInfo        LogType = "info"
	LogError       LogType = "error"
	LogWarn        LogType = "warn"
	LogDebug       LogType = "debug"
)

// traceIDKeyType used to store Trace ID in context
// traceIDKeyType 用于在 context 中存储 Trace ID
type traceIDKeyType struct{}

// TraceIDKey is the key to store Trace ID in context
// TraceIDKey 是 context 中存储 Trace ID 的 key
var TraceIDKey = traceIDKeyType{}

// GetTraceID gets Trace ID from context
// GetTraceID 从 context 中获取 Trace ID
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// generateTraceID generates a new Trace ID
// generateTraceID 生成新的 Trace ID
func generateTraceID() string {
	return uuid.New().String()
}

// extractOrGenerateTraceID extracts or generates Trace ID from HTTP request
// extractOrGenerateTraceID 从 HTTP 请求中提取或生成 Trace ID
func extractOrGenerateTraceID(c *gin.Context) string {
	// Try to get from Header
	// extractOrGenerateTraceID 尝试从 Header 中获取
	if traceID := c.GetHeader("X-Trace-ID"); traceID != "" {
		return traceID
	}
	if traceID := c.GetHeader("X-Request-ID"); traceID != "" {
		return traceID
	}
	// Generate new Trace ID
	// 生成新的 Trace ID
	return generateTraceID()
}

// wsLogger is the logger used by WebSocket module (injected via App Container)
// wsLogger 是 WebSocket 模块使用的日志器（通过 App Container 注入）
var wsLogger *zap.Logger

// wsProductionMode marks whether it is production mode (injected via App Container)
// wsProductionMode 标记是否为生产模式（通过 App Container 注入）
var wsProductionMode bool

// SetWSLogger sets the logger for WebSocket module
// SetWSLogger 设置 WebSocket 模块的日志器
func SetWSLogger(logger *zap.Logger) {
	wsLogger = logger
}

// SetWSProductionMode sets the production mode flag for WebSocket module
// SetWSProductionMode 设置 WebSocket 模块的生产模式标记
func SetWSProductionMode(production bool) {
	wsProductionMode = production
}

// isDevelopmentMode checks if it is development environment
// isDevelopmentMode 检查是否为开发环境
// Output colored console logs in development environment
// 开发环境下会输出彩色控制台日志
func isDevelopmentMode() bool {
	return !wsProductionMode
}

// log records logs
// log 记录日志
// t: log type
// t: 日志类型
// msg: log message
// msg: 日志消息
// fields: zap log fields
// fields: zap 日志字段
func log(t LogType, msg string, fields ...zap.Field) {
	if wsLogger == nil {
		return
	}
	switch t {
	case LogError:
		wsLogger.Error(msg, fields...)
	case LogWarn:
		wsLogger.Warn(msg, fields...)
	case LogInfo:
		wsLogger.Info(msg, fields...)
	case LogDebug:
		wsLogger.Debug(msg, fields...)
	}
}

// logWithTraceID records logs, including Trace ID
// logWithTraceID 记录日志，包含 Trace ID
func logWithTraceID(t LogType, traceID string, msg string, fields ...zap.Field) {
	if traceID != "" {
		fields = append([]zap.Field{zap.String("traceId", traceID)}, fields...)
	}
	log(t, msg, fields...)
}

// NoteModifyLog records WebSocket operation logs
// NoteModifyLog 记录 WebSocket 操作日志
// Supports both structured logs and development environment colored output
// 同时支持结构化日志和开发环境彩色输出
// traceID: trace ID
// traceID: 追踪 ID
// uid: user ID
// uid: 用户 ID
// action: name of the operation executed
// action: 执行的操作名称
// params: variadic parameters, usually the first is Path, the second is Vault
// params: 可变参数，通常第一个为 Path，第二个为 Vault
func NoteModifyLog(traceID string, uid int64, action string, params ...string) {
	var path, vault string

	if len(params) > 0 {
		path = params[0]
	}

	if len(params) > 1 {
		vault = params[1]
	}

	// Structured log output (for log aggregation and analysis)
	// 结构化日志输出（用于日志聚合和分析）
	if wsLogger != nil {
		wsLogger.Info("WebSocket action",
			zap.String(logger.FieldTraceID, traceID),
			zap.Int64(logger.FieldUID, uid),
			zap.String(logger.FieldAction, action),
			zap.String(logger.FieldVault, vault),
			zap.String(logger.FieldPath, path),
		)
	}

	// Keep colored console output in development environment for easy local debugging
	// 开发环境保留彩色控制台输出，便于本地调试
	if isDevelopmentMode() {
		printColoredLog(uid, action, traceID, vault, path)
	}
}

// printColoredLog outputs colored logs (development environment only)
// printColoredLog 输出彩色日志（仅开发环境）
// Use ANSI escape codes to achieve colored output
// 使用 ANSI 转义码实现彩色输出
func printColoredLog(uid int64, action, traceID, vault, path string) {
	str := fmt.Sprintf("[WS] | \033[30;43m %d \033[0m\033[97;44m %s \033[0m", uid, action)

	if traceID != "" && len(traceID) >= 8 {
		str += fmt.Sprintf("\033[90m[%s]\033[0m ", traceID[:8]) // Only display the first 8 digits to keep it concise
		// Only display the first 8 digits to keep it concise
		// 只显示前8位以保持简洁
	}

	if vault != "" {
		str += fmt.Sprintf("\033[32m %s \033[0m", vault)
	}

	if path != "" {
		str += fmt.Sprintf("\033[32m %s \033[0m", path)
	}

	fmt.Println(str)
}

type WebSocketMessage struct {
	Type string `json:"type"` // Operation type, e.g., "upload", "update", "delete" // 操作类型，例如 "upload", "update", "delete"
	Data []byte `json:"data"` // File data (only used for upload and update) // 文件数据（仅在上传和更新时使用）
}

type ClientInfoMessage struct {
	Name                string `json:"name"`                // Client name // 客户端名称
	Version             string `json:"version"`             // Client version // 客户端版本
	Type                string `json:"type"`                // Client type "web" | "desktop" | "mobile" | "obsidianPlugin" // 客户端类型 "web" | "desktop" | "mobile" | "obsidianPlugin"
	IsDesktop           bool   `json:"isDesktop"`           // Is desktop // 是否为桌面端
	IsMobile            bool   `json:"isMobile"`            // Is mobile // 是否为移动端
	IsPhone             bool   `json:"isPhone"`             // Is phone // 是否为手机
	IsTablet            bool   `json:"isTablet"`            // Is tablet // 是否为平板
	IsMacOS             bool   `json:"isMacOS"`             // Is macOS // 是否为 macOS
	IsWin               bool   `json:"isWin"`               // Is Windows // 是否为 Windows
	IsLinux             bool   `json:"isLinux"`             // Is Linux // 是否为 Linux
	OfflineSyncStrategy string `json:"offlineSyncStrategy"` // Offline device sync strategy "newTimeMerge" | "ignoreTimeMerge" // 离线设备同步策略 "newTimeMerge" | "ignoreTimeMerge"
	Protobuf            bool   `json:"protobuf"`            // Use protobuf // 是否使用 protobuf
}

type WSConfig struct {
	GWSOption    gws.ServerOption
	PingInterval time.Duration
	PingWait     time.Duration
	// WriteTimeout application-layer write deadline for outbound messages (ToResponse/
	// BroadcastResponse/SendBinary etc.); already resolved by the caller (nil-vs-zero
	// distinguished at the config layer), so 0 here means "no deadline" (old behavior)
	// and is NOT replaced by an internal default.
	// WriteTimeout 应用层出站消息（ToResponse/BroadcastResponse/SendBinary 等）的写超时；
	// 调用方已在配置层解析好 nil 与显式 0 的区别，这里 0 就表示"不设超时"（旧行为），
	// 不会再被内部默认值覆盖。
	WriteTimeout time.Duration
}

// SessionCleaner interface, used to clean up session resources when the connection is disconnected
// SessionCleaner 接口，用于在连接断开时清理会话资源
type SessionCleaner interface {
	Cleanup()
}

// PathHashGetter interface, used to identify session by file path hash
// PathHashGetter 接口，用于通过文件路径哈希标识会话
type PathHashGetter interface {
	GetPathHash() string
}

// SessionCreatedAtGetter interface for sessions that track creation time
// SessionCreatedAtGetter 接口，用于获取会话创建时间
type SessionCreatedAtGetter interface {
	GetCreatedAt() time.Time
}

// DiffMergeEntry represents an entry in DiffMergePaths
// DiffMergeEntry 表示 DiffMergePaths 中的条目
// Contains creation timestamp for timeout cleanup mechanism
// 包含创建时间戳，用于超时清理机制
type DiffMergeEntry struct {
	CreatedAt time.Time // Entry creation time // 条目创建时间
}

// WebsocketClient struct to store each WebSocket connection and its associated state
// WebsocketClient 结构体来存储每个 WebSocket 连接及其相关状态
type WebsocketClient struct {
	conn                *gws.Conn                 // Underlying WebSocket connection handle // WebSocket 底层连接句柄
	done                chan struct{}             // Close signal channel, used for graceful shutdown // 关闭信号通道，用于优雅关闭读/写协程
	app                 AppContainer              // App Container reference // App Container 引用
	Server              *WebsocketServer          // WebSocket server reference // WebSocket 服务器引用，用于访问全局状态（如会话）
	Ctx                 *gin.Context              // Original HTTP upgrade request context // 原始 HTTP 升级请求的上下文
	WsCtx               context.Context           // Long-lifecycle context for WebSocket connection // WebSocket 连接的长生命周期 context
	WsCancel            context.CancelFunc        // Used to cancel WsCtx // 用于取消 WsCtx
	TraceID             string                    // Trace ID of the connection // 连接的追踪 ID
	User                *UserEntity               // Authenticated user info // 已认证用户信息，通常在握手阶段绑定
	UserClients         ConnStorage               // User connection pool // 用户连接池，支持多设备在线时广播或单点通信
	SF                  *singleflight.Group       // Concurrency control // 并发控制：相同 key 的请求只执行一次，其余等待结果
	BinaryMu            sync.Mutex                // Synchronization lock when reading and writing data // 用于读写数据时的同步锁 (不再保护 map 存储)
	infoMu              sync.RWMutex              // Guards the client-reported connection metadata below (written once by ClientInfo, read concurrently under gws ParallelEnabled) // 保护下方客户端上报的连接元数据（由 ClientInfo 写入一次，gws ParallelEnabled 下会被并发读取）
	clientName          string                    // Client name (e.g., "Mac", "Windows", "iPhone"); access via ClientName() // 客户端名称 (例如 "Mac", "Windows", "iPhone")；请通过 ClientName() 访问
	clientType          string                    // Client type "web" | "desktop" | "mobile" | "obsidianPlugin"; access via ClientType() // 客户端类型；请通过 ClientType() 访问
	clientPlatform      map[string]bool           // Client platform details; access via ClientPlatform() // 客户端平台详情；请通过 ClientPlatform() 访问
	clientVersion       string                    // Client version number (e.g., "1.2.4"); access via ClientVersion() // 客户端版本号；请通过 ClientVersion() 访问
	offlineSyncStrategy string                    // Offline device sync strategy "newTimeMerge" | "ignoreTimeMerge"; access via OfflineSyncStrategy() // 离线设备同步策略；请通过 OfflineSyncStrategy() 访问
	useProtobuf         bool                      // Whether to use protobuf protocol; access via UseProtobuf() // 是否使用 protobuf 协议；请通过 UseProtobuf() 访问
	StartTime           timex.Time                // Connection start time // 连接开始时间
	IsFirstSync         bool                      // Whether it's the first sync // 是否是第一次同步过
	DiffMergePaths      map[string]DiffMergeEntry // File paths needing merging // 需要合并的文件路径，包含创建时间用于超时清理
	DiffMergePathsMu    sync.RWMutex              // Mutex lock to prevent concurrency conflicts // 互斥锁，防止并发冲突
	failCount           atomic.Int32              // Consecutive broadcast failure counter; connection closed when exceeding threshold // 连续广播失败计数器，超过阈值时主动关闭连接
	lastPongAt          atomic.Int64                    // Unix timestamp of last received pong; used to detect zombie connections // 最后一次收到 pong 的 Unix 时间戳，用于检测僵尸连接
	TokenID             int64                     // Bound Token ID // 绑定的令牌 ID
	Scope               string                    // Token Scope // 令牌权限范围
	Vaults              string                    // Restrict Vaults // 限制笔记库
	Lang                string                    // Language preference // 语言偏好
	Protocol            string                    // Protocol "protobuf" or other // 协议 "protobuf" 或其他
	ProtoVersion        int                       // Client-declared handshake protocol version, from URL query "pv"; >=2 means the client supports v2 negotiation (negotiation block in auth response, window pipelining, early pb upgrade) // 客户端声明的握手协议版本，来自 URL query "pv"；>=2 表示客户端支持 v2 协商（auth 响应携带协商块、窗口流水线、pb 提前升级）
	PbEnabled           bool                      // Client's local protobufEnabled setting, from URL query "pb" (1/0); only meaningful when ProtoVersion>=2 // 客户端本地 protobufEnabled 设置，来自 URL query "pb"（1/0）；仅在 ProtoVersion>=2 时有意义
	currentAction       string                    // Current action type being processed // Current action type being processed // 当前正在处理的动作类型
}

// ClientName returns the client-reported name (e.g. "Mac", "Windows", "iPhone").
// Safe for concurrent use; guarded against the concurrent write in ClientInfo().
// ClientName 返回客户端上报的名称（例如 "Mac"、"Windows"、"iPhone"）。
// 并发安全，防止与 ClientInfo() 中的并发写发生竞争。
func (c *WebsocketClient) ClientName() string {
	c.infoMu.RLock()
	defer c.infoMu.RUnlock()
	return c.clientName
}

// ClientType returns the client-reported type ("web" | "desktop" | "mobile" | "obsidianPlugin").
// ClientType 返回客户端上报的类型（"web" | "desktop" | "mobile" | "obsidianPlugin"）。
func (c *WebsocketClient) ClientType() string {
	c.infoMu.RLock()
	defer c.infoMu.RUnlock()
	return c.clientType
}

// ClientVersion returns the client-reported version (e.g. "1.2.4").
// ClientVersion 返回客户端上报的版本号（例如 "1.2.4"）。
func (c *WebsocketClient) ClientVersion() string {
	c.infoMu.RLock()
	defer c.infoMu.RUnlock()
	return c.clientVersion
}

// ClientPlatform returns the client-reported platform flags. The returned map is only ever
// replaced wholesale (never mutated in place) by ClientInfo(), so it is safe to read after
// this call returns even though the map itself isn't copied.
// ClientPlatform 返回客户端上报的平台标记。ClientInfo() 只会整体替换该 map（不会原地修改），
// 因此即便没有拷贝该 map，本调用返回后继续读取也是安全的。
func (c *WebsocketClient) ClientPlatform() map[string]bool {
	c.infoMu.RLock()
	defer c.infoMu.RUnlock()
	return c.clientPlatform
}

// OfflineSyncStrategy returns the client-reported offline sync strategy.
// OfflineSyncStrategy 返回客户端上报的离线同步策略。
func (c *WebsocketClient) OfflineSyncStrategy() string {
	c.infoMu.RLock()
	defer c.infoMu.RUnlock()
	return c.offlineSyncStrategy
}

// UseProtobuf reports whether this connection negotiated the protobuf protocol.
// UseProtobuf 返回该连接是否已协商使用 protobuf 协议。
func (c *WebsocketClient) UseProtobuf() bool {
	c.infoMu.RLock()
	defer c.infoMu.RUnlock()
	return c.useProtobuf
}

// setClientInfo atomically updates all client-reported connection metadata under infoMu,
// so concurrent readers (e.g. other goroutines processing messages on the same connection
// under gws ParallelEnabled) never observe a partially-updated state.
// setClientInfo 在 infoMu 保护下原子更新全部客户端上报的连接元数据，
// 使并发读方（例如 gws ParallelEnabled 下处理同一连接其他消息的 goroutine）
// 不会看到只更新了一部分字段的中间状态。
func (c *WebsocketClient) setClientInfo(name, clientType, version string, platform map[string]bool, offlineSyncStrategy string, useProtobuf bool) {
	c.infoMu.Lock()
	defer c.infoMu.Unlock()
	c.clientName = name
	c.clientType = clientType
	c.clientVersion = version
	c.clientPlatform = platform
	c.offlineSyncStrategy = offlineSyncStrategy
	c.useProtobuf = useProtobuf
}

// setUseProtobuf sets only the useProtobuf field, guarded by the same infoMu as setClientInfo.
// Used for the v2 handshake merge's early pb upgrade (§5.1): the connection may switch to
// protobuf right after the auth response, before ClientInfo arrives. ClientInfo's later call
// to setClientInfo() re-sets useProtobuf to the same value, so this is idempotent with it.
// setUseProtobuf 仅设置 useProtobuf 字段，与 setClientInfo 共用同一把 infoMu。
// 用于 v2 握手合并的 pb 提前升级（§5.1）：连接可能在 auth 响应后、ClientInfo 到达前就切换到 protobuf。
// ClientInfo 随后调用 setClientInfo() 会把 useProtobuf 重新设置为相同的值，因此与之幂等。
func (c *WebsocketClient) setUseProtobuf(useProtobuf bool) {
	c.infoMu.Lock()
	defer c.infoMu.Unlock()
	c.useProtobuf = useProtobuf
}

// initContext initializes the context for the WebSocket connection
// initContext 初始化 WebSocket 连接的 context
// Called when building connection
// 在连接建立时调用
func (c *WebsocketClient) initContext(traceID string) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, TraceIDKey, traceID)
	c.WsCtx, c.WsCancel = context.WithCancel(ctx)
	c.TraceID = traceID
}

// cancelContext cancels the context for the WebSocket connection
// cancelContext 取消 WebSocket 连接的 context
// Called when closing connection
// 在连接关闭时调用
func (c *WebsocketClient) cancelContext() {
	if c.WsCancel != nil {
		c.WsCancel()
	}
}

// Context returns the context of the WebSocket connection
// Context 返回 WebSocket 连接的 context
// Used for all operations requiring context (database queries, external calls, etc.)
// 用于所有需要 context 的操作（数据库查询、外部调用等）
func (c *WebsocketClient) Context() context.Context {
	if c.WsCtx == nil {
		panic("WebsocketClient.WsCtx is not initialized")
	}
	return c.WsCtx
}

// WithTimeout creates a sub context with timeout
// WithTimeout 创建带超时的子 context
// Used for operations requiring timeout control
// 用于需要超时控制的操作
func (c *WebsocketClient) WithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.WsCtx, timeout)
}

// CleanupExpiredDiffMergePaths cleans up expired DiffMergePaths entries
// CleanupExpiredDiffMergePaths 清理过期的 DiffMergePaths 条目
// timeout: timeout duration, entries exceeding this duration will be deleted
// timeout: 超时时间，超过此时间的条目将被删除
func (c *WebsocketClient) CleanupExpiredDiffMergePaths(timeout time.Duration) int {
	c.DiffMergePathsMu.Lock()
	defer c.DiffMergePathsMu.Unlock()

	if c.DiffMergePaths == nil {
		return 0
	}

	now := time.Now()
	cleanedCount := 0
	for path, entry := range c.DiffMergePaths {
		if now.Sub(entry.CreatedAt) > timeout {
			delete(c.DiffMergePaths, path)
			cleanedCount++
		}
	}
	return cleanedCount
}

// ClearAllDiffMergePaths cleans up all DiffMergePaths entries
// ClearAllDiffMergePaths 清理所有 DiffMergePaths 条目
// Called when closing connection
// 在连接关闭时调用
func (c *WebsocketClient) ClearAllDiffMergePaths() int {
	c.DiffMergePathsMu.Lock()
	defer c.DiffMergePathsMu.Unlock()

	if c.DiffMergePaths == nil {
		return 0
	}

	count := len(c.DiffMergePaths)
	c.DiffMergePaths = make(map[string]DiffMergeEntry)
	return count
}

// WebSocket version of parameter binding and validation utility functions based on global validator
// 基于全局验证器的 WebSocket 版本参数绑定和验证工具函数
func (c *WebsocketClient) BindAndValid(data []byte, obj any) (bool, ValidErrors) {
	return c.BindAndValidWithAction(c.currentAction, data, obj)
}

// BindAndValidWithAction WebSocket version of parameter binding and validation with specific action to avoid race condition
// 带特定动作名称的 WebSocket 版本参数校验，避免并发消息下的 c.currentAction 竞态冲突
func (c *WebsocketClient) BindAndValidWithAction(action string, data []byte, obj any) (bool, ValidErrors) {
	var errs ValidErrors

	if c.UseProtobuf() && c.Server.ProtobufDecoder != nil {
		decoded, err := c.Server.ProtobufDecoder(action, data, obj)
		if err != nil {
			errs = append(errs, &ValidError{
				Key:     "body",
				Message: "Protobuf decode failed: " + err.Error(),
			})
			return false, errs
		}
		if decoded {
			validator := c.app.Validator()
			if validator == nil {
				return true, nil
			}
			if err := validator.ValidateStruct(obj); err != nil {
				if validationErrors, ok := err.(validatorV10.ValidationErrors); ok {
					v := c.Ctx.Value("trans")
					trans := v.(ut.Translator)
					for _, validationErr := range validationErrors {
						translatedMsg := validationErr.Translate(trans)
						errs = append(errs, &ValidError{
							Key:     validationErr.Field(),
							Message: translatedMsg,
						})
					}
				}
				return false, errs
			}
			return true, nil
		}
	}

	// Step 1: JSON deserialization (can be replaced by other formats)
	// BindAndValid Step 1: JSON 反序列化（可替换成其他格式）
	if err := json.Unmarshal(data, obj); err != nil {
		// Decoding error handling
		// BindAndValid 解码错误处理
		errs = append(errs, &ValidError{
			Key:     "body",
			Message: "Invalid message format",
		})
		return false, errs
	}

	// Step 2: Parameter validation
	// Step 2: 参数验证
	validator := c.app.Validator()
	if validator == nil {
		return true, nil
	}
	if err := validator.ValidateStruct(obj); err != nil {
		// If verification fails, check error type
		// 如果验证失败，检查错误类型
		if validationErrors, ok := err.(validatorV10.ValidationErrors); ok {
			// Get translator
			// 获取翻译器
			v := c.Ctx.Value("trans")
			trans := v.(ut.Translator)

			// Iterate through validation errors and translate them
			// 遍历验证错误并进行翻译
			for _, validationErr := range validationErrors {
				translatedMsg := validationErr.Translate(trans) // Translate error message
				// Translate error message
				// 翻译错误消息
				errs = append(errs, &ValidError{
					Key:     validationErr.Field(),
					Message: translatedMsg,
				})
			}
		}
		return false, errs // Return validation error
		// 返回验证错误
	}
	return true, nil
}

// Send Ping message regularly
// 定期发送 Ping 消息
func (c *WebsocketClient) PingLoop(PingInterval time.Duration) {
	ticker := time.NewTicker(PingInterval * time.Second) // Send Ping every 25 seconds // 每 25 秒发送一次 Ping
	defer ticker.Stop()

	// Periodic cleanup of expired conflict merge paths
	// 定期清理已过期的冲突合并路径
	cleanupTicker := time.NewTicker(10 * time.Minute)
	defer cleanupTicker.Stop()

	// Initialize last pong time to now (connection just established)
	// 初始化最后 pong 时间为当前（连接刚建立）
	c.lastPongAt.Store(time.Now().Unix())
	// Track whether we've sent a ping and are waiting for a pong
	// 跟踪是否已发送 ping 并等待 pong
	pingSent := false

	for {
		select {
		case <-c.done:
			log(LogInfo, "WS Client Close Ping")
			return
		case <-ticker.C:
			if c.conn == nil {
				return
			}
			// If we sent a ping last cycle but never received a pong, the connection is likely dead.
			// If we sent a ping last cycle but never received a pong, force close.
			// 如果上一轮发了 ping 但没收到 pong，连接可能已死，强制关闭。
			if pingSent {
				lastPong := c.lastPongAt.Load()
				elapsed := time.Since(time.Unix(lastPong, 0))
				if elapsed > time.Duration(WSPingWait)*time.Second {
					log(LogWarn, "WS Client: no pong received within deadline, force closing",
						zap.Duration("sinceLastPong", elapsed),
						zap.String("traceID", c.TraceID))
					_ = c.conn.NetConn().Close()
					return
				}
			}
			// Set write deadline to prevent WritePing from blocking indefinitely on dead connections.
			// 设置写超时，防止 WritePing 在死连接上永久阻塞
			_ = c.conn.NetConn().SetWriteDeadline(time.Now().Add(WSPingWriteTimeout * time.Second))
			if err := c.conn.WritePing(nil); err != nil {
				// Normal error when the connection is closed, lower log level
				// 连接关闭时的正常错误，降低日志级别
				if strings.Contains(err.Error(), "use of closed network connection") {
					log(LogDebug, "WS Client Ping: connection closed")
				} else {
					log(LogError, "WS Client Ping err ", zap.Error(err))
					// Force close the underlying connection to trigger gws OnClose callback,
					// release all resources (buffers, goroutines, worker pool slots).
					// 强制关闭底层连接，触发 gws OnClose 回调，释放所有资源（缓冲区、goroutine、Worker Pool 槽位）
					_ = c.conn.NetConn().Close()
				}
				pingSent = false
				return
			}
			// Reset write deadline after successful ping.
			// Ping 成功后重置写超时
			_ = c.conn.NetConn().SetWriteDeadline(time.Time{})
			pingSent = true
			// log(LogInfo, "WS Client Ping", zap.String("uid", c.User.ID))
		case <-cleanupTicker.C:
			// Cleanup items expired for more than 1 hour
			// 清理过期超过 1 小时的项
			if count := c.CleanupExpiredDiffMergePaths(1 * time.Hour); count > 0 {
				log(LogDebug, "PingLoop: cleaned up expired DiffMergePaths",
					zap.Int("count", count),
					zap.String("traceID", c.TraceID))
			}
		}
	}
}

// ToResponse converts the result to JSON format and sends it to the client
// ToResponse 将结果转换为 JSON 格式并发送给客户端
func (c *WebsocketClient) ToResponse(code *code.Code, action ...string) {

	var actionType string
	if len(action) > 0 {
		actionType = action[0]
	}

	var responseBytes []byte

	content := Res{
		Code:    code.Code(),
		Status:  code.Status(),
		Message: code.MsgIn(c.Lang),
		Data:    code.Data(),
	}

	if code.HaveDetails() {
		content.Details = strings.Join(code.Details(), ",")
	}

	if code.HaveVault() {
		content.Vault = code.Vault()
	}
	if code.HaveContext() {
		content.Context = code.Context()
	}
	if code.HavePath() {
		content.Path = code.Path()
	}
	if code.HavePageIndex() {
		content.PageIndex = code.PageIndex()
	}

	if c.app.IsReturnSuccess() || actionType != "" || code.Code() > 200 || code.HaveData() || code.HaveDetails() {
		if c.UseProtobuf() && c.Server.ProtobufEncoder != nil && actionType != "" {
			pbBytes, err := c.Server.ProtobufEncoder(actionType, &content)
			if err == nil {
				c.writeMessage(gws.OpcodeBinary, pbBytes)
				return
			}
			log(LogError, "WS Protobuf encode failed, falling back to JSON", zap.Error(err), zap.String("uid", func() string {
				if c.User != nil {
					return c.User.ID
				}
				return "Guest"
			}()))
		}

		responseBytes, _ = json.Marshal(content)
		if actionType != "" {
			responseBytes = []byte(fmt.Sprintf(`%s|%s`, actionType, string(responseBytes)))
		}
		c.send(responseBytes, false, false)
	}
}

// BroadcastResponse converts the result to JSON format and broadcasts it to all connected clients of the current user
// BroadcastResponse 将结果转换为 JSON 格式并广播给当前用户的所有连接客户端
//
// Parameters:
// 参数:
//
//	code: business response status code object, including status code, message and data
//	code: 业务响应状态码对象，包含状态码、消息和数据
//	options: optional parameter list
//	options: 可选参数列表
//	  - options[0] (bool):   required, whether to exclude the current client (true: exclude self, false: broadcast to all ends)
//	  - options[0] (bool):   必填，是否排除当前客户端 (true: 排除自己, false: 广播给所有端)
//	  - options[1] (string): optional, identification of action type (ActionType), used for clients to distinguish message types
//	  - options[1] (string): 选填，动作类型标识 (ActionType)，用于客户端区分消息类型
func (c *WebsocketClient) BroadcastResponse(code *code.Code, options ...any) {

	var actionType string
	if len(options) > 1 {
		actionType = options[1].(string)
	}

	if len(c.UserClients) <= 0 {
		return
	}

	content := Res{
		Code:    code.Code(),
		Status:  code.Status(),
		Message: code.MsgIn(c.Lang),
		Data:    code.Data(),
	}

	if code.HaveDetails() {
		content.Details = strings.Join(code.Details(), ",")
	}

	if code.HaveVault() {
		content.Vault = code.Vault()
	}

	if code.HaveContext() {
		content.Context = code.Context()
	}

	c.sendBroadcast(&content, actionType, options[0].(bool))
}

func (c *WebsocketClient) send(responseBytes []byte, isBroadcast bool, isExcludeSelf bool) {
	c.sendMessage(responseBytes)
}

func (c *WebsocketClient) sendMessage(payload []byte) {
	c.writeMessage(gws.OpcodeText, payload)
}

// writeMessage writes an application-layer message under the configured write deadline
// (ws-write-timeout, default 10s), so a stalled/zombie connection cannot block WriteMessage
// indefinitely and stall the write lock (see P9). The deadline is cleared after the write
// completes, matching PingLoop's SetWriteDeadline/clear usage.
// writeMessage 在配置的应用层写超时（ws-write-timeout，默认 10s）保护下写入消息，
// 避免僵尸/卡顿连接让 WriteMessage 无限阻塞并拖住写锁（见 P9）。写完后清空 deadline，
// 用法与 PingLoop 的 SetWriteDeadline/清空一致。
func (c *WebsocketClient) writeMessage(opcode gws.Opcode, payload []byte) error {
	if c.conn == nil {
		return fmt.Errorf("connection is nil")
	}
	timeout := c.Server.config.WriteTimeout
	if timeout > 0 {
		_ = c.conn.NetConn().SetWriteDeadline(time.Now().Add(timeout))
	}
	err := c.conn.WriteMessage(opcode, payload)
	if timeout > 0 {
		_ = c.conn.NetConn().SetWriteDeadline(time.Time{})
	}
	return err
}

func (c *WebsocketClient) sendBroadcast(content *Res, actionType string, isExcludeSelf bool) {
	// 持锁期间只拷贝目标连接列表，随后立即释放锁——WriteMessage 本身可能阻塞（慢设备/网络抖动），
	// 不应该在持有 c.Server.mu 期间发生，否则会拖慢该用户下所有其他并发操作。
	// Only copy the target connection list while holding the lock, then release it right
	// away — WriteMessage can block (slow device / network jitter) and must not happen
	// while holding c.Server.mu, or it stalls every other concurrent operation for this user.
	c.Server.mu.RLock()
	targets := make([]*WebsocketClient, 0, len(c.UserClients))
	for _, uc := range c.UserClients {
		if uc.conn == nil {
			continue
		}
		if isExcludeSelf && uc.conn == c.conn {
			continue
		}
		targets = append(targets, uc)
	}
	c.Server.mu.RUnlock()

	if len(targets) == 0 {
		return
	}

	var jsonBytes []byte
	if actionType != "" {
		mBytes, _ := json.Marshal(content)
		jsonBytes = []byte(fmt.Sprintf(`%s|%s`, actionType, string(mBytes)))
	} else {
		jsonBytes, _ = json.Marshal(content)
	}

	// 逐连接并发扇出：gws Conn.WriteMessage 内部对同一连接的写入用 c.mu 做了互斥
	// （已查证 github.com/lxzan/gws@v1.9.1 writer.go doWrite），不同连接之间互不影响，
	// 因此可以安全地并发写入，避免一台慢设备拖慢同用户下其他设备的广播。
	// Fan out concurrently per connection: gws Conn.WriteMessage internally serializes
	// writes to the same connection via c.mu (verified in
	// github.com/lxzan/gws@v1.9.1 writer.go doWrite); different connections don't share
	// that lock, so concurrent writes across connections are safe and prevent one slow
	// device from stalling the broadcast to the user's other devices.
	var wg sync.WaitGroup
	for _, uc := range targets {
		wg.Add(1)
		safego.Go(wsLogger, func() {
			defer wg.Done()

			var err error
			if uc.UseProtobuf() && uc.Server.ProtobufEncoder != nil && actionType != "" {
				var pbBytes []byte
				pbBytes, err = uc.Server.ProtobufEncoder(actionType, content)
				if err == nil {
					err = uc.writeMessage(gws.OpcodeBinary, pbBytes)
				}
			} else {
				err = uc.writeMessage(gws.OpcodeText, jsonBytes)
			}

			if err != nil {
				if uc.failCount.Add(1) == 4 {
					uc.conn.WriteClose(1000, []byte("broadcast failed"))
				}
			} else {
				uc.failCount.Store(0)
			}
		})
	}
	wg.Wait()
}

// SendBinary sends binary messages
// SendBinary 发送二进制消息
// prefix: 2-byte prefix
// prefix: 2字节前缀
func (c *WebsocketClient) SendBinary(prefix string, payload []byte) error {
	if c.conn == nil {
		return fmt.Errorf("connection is nil")
	}
	if len(prefix) != 2 {
		return fmt.Errorf("prefix must be 2 bytes")
	}
	// Concat prefix and data
	// 拼接前缀和数据
	data := make([]byte, 2+len(payload))
	copy(data[0:2], prefix)
	copy(data[2:], payload)
	return c.writeMessage(gws.OpcodeBinary, data)
}

// ------------------------------------> WebsocketServer

type ConnStorage = map[*gws.Conn]*WebsocketClient

// AppContainer defines App Container interface, used to decouple pkg/app and internal/app
// AppContainer 定义 App Container 接口，用于解耦 pkg/app 和 internal/app
// This interface allows WebsocketServer to use App Container functions without circular dependency
// 这个接口允许 WebsocketServer 使用 App Container 的功能而不产生循环依赖
type AppContainer interface {
	// Logger gets logger
	// Logger 获取日志器
	Logger() *zap.Logger
	// SubmitTask submits task to Worker Pool
	// SubmitTask 提交任务到 Worker Pool
	SubmitTask(ctx context.Context, task func(context.Context) error) error
	// SubmitTaskAsync submits task to Worker Pool asynchronously (without waiting for results)
	// SubmitTaskAsync 异步提交任务到 Worker Pool（不等待结果）
	SubmitTaskAsync(ctx context.Context, task func(context.Context) error) error
	// Version gets version info
	// Version 获取版本信息
	Version() VersionInfo
	// CheckVersion checks version
	// CheckVersion 检查版本
	CheckVersion(pluginVersion string) CheckVersionInfo
	// Validator gets validator (may be nil)
	// Validator 获取验证器（可能为 nil）
	Validator() ValidatorInterface
	// IsReturnSuccess whether to return success response
	// IsReturnSuccess 是否返回成功响应
	IsReturnSuccess() bool
	// GetAuthTokenKey gets Token key
	// GetAuthTokenKey 获取 Token 密钥
	GetAuthTokenKey() string
	// IsProductionMode whether it is production mode
	// IsProductionMode 是否为生产模式
	IsProductionMode() bool
	// GetTokenService gets token service for RBAC
	// GetTokenService 获取 Token 服务
	GetTokenService() any // Use any to avoid circular dependency, then type assert in use
	// SyncChunkNums returns the configured upload/download sync batch sizes (SyncUpChunkNum,
	// SyncDownChunkNum), used for v2 handshake negotiation (§2.3).
	// SyncChunkNums 返回配置的上传/下载同步分批大小（SyncUpChunkNum、SyncDownChunkNum），
	// 用于 v2 握手协商（§2.3）。
	SyncChunkNums() (up int, down int)
	// PipelineWindows returns the clamped upload/download pipeline window sizes for v2 handshake
	// negotiation (§2.3); 0 means the window is disabled (stop-and-wait) for that direction.
	// PipelineWindows 返回用于 v2 握手协商（§2.3）的、经过钳制的上下行流水线窗口大小；
	// 0 表示该方向禁用窗口（stop-and-wait）。
	PipelineWindows() (up int, down int)
}

// ValidatorInterface validator interface
// ValidatorInterface 验证器接口
type ValidatorInterface interface {
	ValidateStruct(obj interface{}) error
}

type WebsocketServer struct {
	app               AppContainer // App Container (Required) // App Container（必须）
	handlers           map[string]func(*WebsocketClient, *WebSocketMessage)
	noAuthHandlers     map[string]func(*WebsocketClient, *WebSocketMessage) // Handlers that do not require user authentication // 免登录鉴权消息处理器集合
	interceptors       []func(*WebsocketClient, *WebSocketMessage) bool     // Pre-handler interceptor chain // 消息前置拦截器链
	userVerifyHandler  func(*WebsocketClient, int64) (*UserSelectEntity, error)
	tokenVerifyHandler func(ctx context.Context, uid int64, tokenID int64, nonce string, reqClientType, reqClientName, reqClientVersion, reqUserAgent, reqIP string) (string, string, error)
	binaryHandlers    map[string]func(*WebsocketClient, []byte) // Binary message handler map: prefix -> handler // 二进制消息处理器映射 prefix -> handler
	clients           ConnStorage
	userClients       map[string]ConnStorage
	connWg            sync.WaitGroup
	mu                sync.RWMutex
	up                *gws.Upgrader
	config            *WSConfig
	// Global session management (UID -> SessionID -> Session)
	// 全局会话管理 (UID -> SessionID -> Session)
	binaryChunkSessions map[string]map[string]any
	sessionsMu          sync.RWMutex
	EnvelopeDecoder     func(data []byte) (string, []byte, error)               // Protobuf envelope decoder // Protobuf 信封解包钩子
	ProtobufDecoder     func(action string, data []byte, obj any) (bool, error) // Protobuf decoder hook // Protobuf 解码钩子
	ProtobufEncoder     func(action string, res *Res) ([]byte, error)           // Protobuf encoder hook // Protobuf 编码钩子
}

// WSClientInfo WebSocket client information for API responses
// WSClientInfo 用于 API 响应的 WebSocket 客户端信息
type WSClientInfo struct {
	UID           string          `json:"uid"`
	Nickname      string          `json:"nickname"`
	ClientName    string          `json:"clientName"`
	ClientType    string          `json:"clientType"`
	ClientVersion string          `json:"clientVersion"`
	PlatformInfo  map[string]bool `json:"platformInfo"`
	RemoteAddr    string          `json:"remoteAddr"`
	StartTime     timex.Time      `json:"startTime"`
	TraceID       string          `json:"traceId"`
	TokenID       int64           `json:"tokenId"`
}

// GetClients returns information of all currently connected WebSocket clients
// GetClients 返回所有当前已连接的 WebSocket 客户端信息
func (w *WebsocketServer) GetClients() []WSClientInfo {
	w.mu.RLock()
	defer w.mu.RUnlock()
	clients := make([]WSClientInfo, 0, len(w.clients))
	for _, c := range w.clients {
		info := WSClientInfo{
			ClientName:    c.ClientName(),
			ClientType:    c.ClientType(),
			ClientVersion: c.ClientVersion(),
			PlatformInfo:  c.ClientPlatform(),
			RemoteAddr:    c.conn.RemoteAddr().String(),
			StartTime:     c.StartTime,
			TraceID:       c.TraceID,
			TokenID:       c.TokenID,
		}
		if c.User != nil {
			info.UID = c.User.ID
			info.Nickname = c.User.Nickname
		}
		clients = append(clients, info)
	}
	return clients
}

// KickClient closes a WebSocket connection by TraceID
// KickClient 通过 TraceID 关闭 WebSocket 连接
func (w *WebsocketServer) KickClient(traceID string) bool {
	w.mu.RLock()
	client, ok := w.clientsByTraceID(traceID)
	w.mu.RUnlock()

	if !ok {
		return false
	}

	if client.conn != nil {
		client.conn.WriteClose(1000, []byte("kicked by admin"))
	}
	return true
}

// clientsByTraceID finds a client by TraceID (helper, requires mu lock)
func (w *WebsocketServer) clientsByTraceID(traceID string) (*WebsocketClient, bool) {
	for _, c := range w.clients {
		if c.TraceID == traceID {
			return c, true
		}
	}
	return nil, false
}

// NewWebsocketServer creates WebSocket server instance
// NewWebsocketServer 创建 WebSocket 服务器实例
// c: WebSocket config // c: WebSocket 配置
// app: App Container (Required) // app: App Container（必须）
func NewWebsocketServer(c WSConfig, app AppContainer) *WebsocketServer {
	if app == nil {
		panic("AppContainer is required for WebsocketServer")
	}
	if c.PingInterval == 0 {
		c.PingInterval = WSPingInterval
	}
	if c.PingWait == 0 {
		c.PingWait = WSPingWait
	}

	// Set logger for WebSocket module
	// 设置 WebSocket 模块的日志器
	SetWSLogger(app.Logger())
	// Set production mode flag for WebSocket module
	// 设置 WebSocket 模块的生产模式标记
	SetWSProductionMode(app.IsProductionMode())

	wss := &WebsocketServer{
		app:                 app,
		handlers:            make(map[string]func(*WebsocketClient, *WebSocketMessage)),
		noAuthHandlers:      make(map[string]func(*WebsocketClient, *WebSocketMessage)),
		interceptors:        make([]func(*WebsocketClient, *WebSocketMessage) bool, 0),
		binaryHandlers:      make(map[string]func(*WebsocketClient, []byte)),
		clients:             make(ConnStorage),
		userClients:         make(map[string]ConnStorage),
		config:              &c,
		binaryChunkSessions: make(map[string]map[string]any),
	}

	// Register built-in unauthenticated handlers
	// 自动注册系统内置免登录鉴权的消息处理器
	wss.UseWithoutAuth("Authorization", wss.Authorization)
	wss.UseWithoutAuth("ClientInfo", wss.ClientInfo)

	return wss
}

// App gets App Container
// App 获取 App Container
func (w *WebsocketServer) App() AppContainer {
	return w.app
}

func (w *WebsocketServer) Upgrade() {
	w.up = gws.NewUpgrader(w, &w.config.GWSOption)
}

func (w *WebsocketServer) Run() gin.HandlerFunc {

	return func(c *gin.Context) {

		w.Upgrade()
		socket, err := w.up.Upgrade(c.Writer, c.Request)
		if err != nil {
			log(LogError, "WS Start err", zap.Error(err))
			return
		}

		// Extract or generate Trace ID from HTTP request
		// 从 HTTP 请求中提取或生成 Trace ID
		traceID := extractOrGenerateTraceID(c)

		client := &WebsocketClient{
			conn:      socket,
			done:      make(chan struct{}),
			app:       w.app,
			Server:    w,
			Ctx:       c,
			SF:        new(singleflight.Group),
			StartTime: timex.Now(),
		}

		// Extract client info from query parameters
		// 从查询参数中提取客户端信息
		// 连接刚创建，尚未发布给其他 goroutine，此处直接写入无需加锁
		// Connection was just created and not yet published to other goroutines, so a direct write here needs no locking
		client.clientType = c.Query("client")
		client.clientName = c.Query("clientName")
		client.clientVersion = c.Query("clientVersion")
		client.Protocol = c.Query("protocol")

		// v2 handshake capability declaration (§2.2): pv = protocol version the client
		// supports, pb = client's local protobufEnabled setting. Missing/invalid values
		// leave ProtoVersion at its zero value, which downstream negotiation logic treats
		// the same as an old (v1) client.
		// v2 握手能力声明（§2.2）：pv = 客户端支持的协议版本，pb = 客户端本地 protobufEnabled 设置。
		// 缺失或非法值使 ProtoVersion 保持零值，下游协商逻辑会将其视为旧（v1）客户端。
		if pv, err := strconv.Atoi(c.Query("pv")); err == nil {
			client.ProtoVersion = pv
		}
		client.PbEnabled = c.Query("pb") == "1"

		// Extract language preference
		// 提取语言偏好
		lang := c.Query("lang")
		if lang == "" {
			lang = c.GetHeader("lang")
		}
		client.Lang = strings.ToLower(strings.ReplaceAll(lang, "-", "_"))

		// Initialize long-lifecycle context for WebSocket connection
		// 初始化 WebSocket 连接的长生命周期 context
		client.initContext(traceID)

		w.AddClient(client)
		w.connWg.Add(1)
		log(LogInfo, "WS Start",
			zap.String("type", "ReadLoop"),
			zap.String("traceID", traceID),
			zap.String("client", client.ClientType()),
			zap.String("clientName", client.ClientName()),
			zap.String("clientVersion", client.ClientVersion()),
		)
		go socket.ReadLoop()
	}
}

func (w *WebsocketServer) Use(action string, handler func(*WebsocketClient, *WebSocketMessage)) {
	w.handlers[action] = handler
}

// UseWithoutAuth registers a message handler that does not require user authentication
// UseWithoutAuth 注册无需用户登录鉴权的消息处理器
func (w *WebsocketServer) UseWithoutAuth(action string, handler func(*WebsocketClient, *WebSocketMessage)) {
	w.noAuthHandlers[action] = handler
}

// UseInterceptor registers a pre-handler interceptor
// UseInterceptor 注册消息前置拦截器
func (w *WebsocketServer) UseInterceptor(interceptor func(*WebsocketClient, *WebSocketMessage) bool) {
	w.interceptors = append(w.interceptors, interceptor)
}

// GetHandler returns the handler for a specific action
// GetHandler 返回指定动作的消息处理器
func (w *WebsocketServer) GetHandler(action string) (func(*WebsocketClient, *WebSocketMessage), bool) {
	h, ok := w.handlers[action]
	return h, ok
}

func (w *WebsocketServer) UseUserVerify(handler func(*WebsocketClient, int64) (*UserSelectEntity, error)) {
	w.userVerifyHandler = handler
}

func (w *WebsocketServer) UseTokenVerify(handler func(ctx context.Context, uid int64, tokenID int64, nonce string, reqClientType, reqClientName, reqClientVersion, reqUserAgent, reqIP string) (string, string, error)) {
	w.tokenVerifyHandler = handler
}

func (w *WebsocketServer) UseBinary(prefix string, handler func(*WebsocketClient, []byte)) {
	if len(prefix) != 2 {
		panic("binary message prefix must be 2 characters")
	}
	w.binaryHandlers[prefix] = handler
}

func (w *WebsocketServer) Authorization(c *WebsocketClient, msg *WebSocketMessage) {

	secretKey := w.app.GetAuthTokenKey()
	if user, err := ParseTokenWithKey(string(msg.Data), secretKey); err != nil {
		log(LogError, "WS Authorization FAILD", zap.Error(err))
		if appErr, ok := err.(*code.Code); ok {
			c.ToResponse(appErr, "Authorization")
		} else {
			c.ToResponse(code.ErrorInvalidUserAuthToken, "Authorization")
		}
		time.Sleep(2 * time.Second)
		c.conn.WriteClose(1000, []byte("AuthorizationFaild"))
	} else {

		uid, err := strconv.ParseInt(user.ID, 10, 64)
		if err != nil {
			log(LogError, "WS Authorization FAILD", zap.Error(err))
			c.ToResponse(code.ErrorInvalidUserAuthToken, "Authorization")
			time.Sleep(2 * time.Second)
			c.conn.WriteClose(1000, []byte("AuthorizationFaild"))
			return
		}

		// Verify 3D RBAC permissions via injected handler
		// 通过注入的处理函数验证 3D RBAC 权限
		if w.tokenVerifyHandler != nil {
			reqClientType := c.Ctx.GetHeader("x-client")
			if reqClientType == "" {
				reqClientType = c.Ctx.Query("client")
			}
			reqUserAgent := c.Ctx.GetHeader("User-Agent")
			reqIP := c.Ctx.ClientIP()

			scope, vaults, err := w.tokenVerifyHandler(c.Context(), uid, user.TokenID, user.Nonce, reqClientType, c.ClientName(), c.ClientVersion(), reqUserAgent, reqIP)
			if err != nil {
				log(LogError, "WS Authorization FAILD: Token verify failed", zap.Error(err))
				if appErr, ok := err.(*code.Code); ok {
					c.ToResponse(appErr, "Authorization")
				} else {
					c.ToResponse(code.ErrorInvalidUserAuthToken, "Authorization")
				}
				time.Sleep(2 * time.Second)
				c.conn.WriteClose(1000, []byte("AuthorizationFaild"))
				return
			}
			c.Scope = scope
			c.Vaults = vaults
		}

		// Mandatorily verify user validity
		// 用户有效性强制验证
		userSelect, err := w.userVerifyHandler(c, uid)
		if userSelect == nil || err != nil {
			log(LogError, "WS Authorization FAILD USER Not Exist", zap.Error(err))
			if appErr, ok := err.(*code.Code); ok {
				c.ToResponse(appErr, "Authorization")
			} else {
				c.ToResponse(code.ErrorInvalidUserAuthToken, "Authorization")
			}
			time.Sleep(2 * time.Second)
			c.conn.WriteClose(1000, []byte("AuthorizationFaild"))
			return
		}

		user.Nickname = userSelect.Nickname
		c.TokenID = user.TokenID

		log(LogInfo, "WS Authorization", zap.String("uid", user.ID), zap.String("Nickname", user.Nickname), zap.Int64("TokenID", c.TokenID))
		c.User = user
		c.UserClients = w.AddUserClient(c)

		versionInfo := w.app.Version()

		// Handshake merge (§2.3/§5.1): auth response is always JSON text (useProtobuf is only
		// ever set by ClientInfo/pv2-early-upgrade, never before this point), so adding keys
		// here is a plain JSON-object addition — old clients (no pv or pv<2) JSON.parse and
		// ignore the unknown keys, producing byte-identical output to pre-3.6.0 since the base
		// four keys never change. Only pv>=2 connections get the negotiation block.
		// 握手合并（§2.3/§5.1）：auth 响应恒为 JSON 文本（useProtobuf 只会被 ClientInfo/pv2 提前升级设置，
		// 在此之前绝不会被设置），因此这里加 key 只是普通的 JSON object 加字段——旧客户端（无 pv 或
		// pv<2）JSON.parse 后忽略未知 key，因为基础四个 key 从不变化，输出与 3.6.0 前逐字节相同。
		// 只有 pv>=2 的连接才会拿到协商块。
		authData := map[string]interface{}{
			"version":   versionInfo.Version,
			"gitTag":    versionInfo.GitTag,
			"buildTime": versionInfo.BuildTime,
			"changelog": versionInfo.Changelog,
		}

		protobufAck := c.Protocol == "protobuf" && c.PbEnabled

		if c.ProtoVersion >= 2 {
			syncUpChunkNum, syncDownChunkNum := w.app.SyncChunkNums()
			pipelineWindowUp, pipelineWindowDown := w.app.PipelineWindows()
			authData["syncUpChunkNum"] = syncUpChunkNum
			authData["syncDownChunkNum"] = syncDownChunkNum
			authData["pipelineWindowUp"] = pipelineWindowUp
			authData["pipelineWindowDown"] = pipelineWindowDown
			authData["protobufAck"] = protobufAck
		}

		c.ToResponse(code.Success.WithData(authData), "Authorization")

		// pb 提前升级：必须在 auth 响应发出之后才切换，确保该响应本身仍以 JSON 文本帧发送
		// (early pb upgrade must happen strictly after the auth response is sent, so that
		// response itself is always transmitted as a JSON text frame).
		if c.ProtoVersion >= 2 {
			c.setUseProtobuf(protobufAck)
		}

		log(LogInfo, "WS User Enter", zap.String("uid", c.User.ID), zap.String("Nickname", c.User.Nickname), zap.Int("Count", len(c.UserClients)))
		go c.PingLoop(w.config.PingInterval)
	}
}

func (w *WebsocketServer) ClientInfo(c *WebsocketClient, msg *WebSocketMessage) {
	var info ClientInfoMessage
	if ok, errs := c.BindAndValidWithAction(msg.Type, msg.Data, &info); !ok {
		log(LogError, "WS ClientInfo Unmarshal FAILD", zap.Error(fmt.Errorf("%s", errs.ErrorsToString())))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()))
		return
	}

	platform := map[string]bool{
		"isDesktop": info.IsDesktop,
		"isMobile":  info.IsMobile,
		"isPhone":   info.IsPhone,
		"isTablet":  info.IsTablet,
		"isMacOS":   info.IsMacOS,
		"isWin":     info.IsWin,
		"isLinux":   info.IsLinux,
	}
	// Enable Protobuf if query param protocol=protobuf and ClientInfo protobuf=true
	useProtobuf := c.Protocol == "protobuf" && info.Protobuf

	// 原子更新全部连接元数据，避免并发读方看到只更新了一部分字段的中间状态
	// Atomically update all connection metadata, avoiding concurrent readers observing a partially-updated state
	c.setClientInfo(info.Name, info.Type, info.Version, platform, info.OfflineSyncStrategy, useProtobuf)
	c.DiffMergePaths = make(map[string]DiffMergeEntry)

	if useProtobuf {
		log(LogInfo, "WS Client upgraded to Protobuf successfully", zap.String("uid", func() string {
			if c.User != nil {
				return c.User.ID
			}
			return "Guest"
		}()))
	} else {
		log(LogInfo, "WS Client downgraded/disabled Protobuf successfully", zap.String("uid", func() string {
			if c.User != nil {
				return c.User.ID
			}
			return "Guest"
		}()))
	}

	log(LogInfo, "WS ClientInfo", zap.String("uid", func() string {
		if c.User != nil {
			return c.User.ID
		}
		return "Guest"
	}()), zap.String("name", c.ClientName()), zap.String("version", c.ClientVersion()), zap.String("offlineSyncStrategy", c.OfflineSyncStrategy()))

	checkVersionInfo := w.app.CheckVersion(c.ClientVersion())

	c.ToResponse(code.Success.WithData(checkVersionInfo), "ClientInfo")
}

// BroadcastClientInfo broadcasts version information to all connected clients
// BroadcastClientInfo 向所有连接的客户端广播版本信息
func (w *WebsocketServer) BroadcastClientInfo() {
	w.mu.RLock()
	clients := make([]*WebsocketClient, 0, len(w.clients))
	for _, c := range w.clients {
		clients = append(clients, c)
	}
	w.mu.RUnlock()

	for _, c := range clients {
		if c.User == nil {
			continue
		}
		checkVersionInfo := w.app.CheckVersion(c.ClientVersion())
		// Only push if there's a new version (server or plugin)
		// 只有当有新版本（服务端或插件）时才推送
		if checkVersionInfo.VersionIsNew || checkVersionInfo.PluginVersionIsNew {
			c.ToResponse(code.Success.WithData(checkVersionInfo), "ClientInfo")
		}
	}
}

func (w *WebsocketServer) GetClient(conn *gws.Conn) *WebsocketClient {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.clients[conn]
}

func (w *WebsocketServer) AddClient(c *WebsocketClient) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.clients[c.conn] = c
}

func (w *WebsocketServer) RemoveClient(conn *gws.Conn) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.clients, conn)
}

func (w *WebsocketServer) AddUserClient(c *WebsocketClient) ConnStorage {
	w.mu.Lock()
	defer w.mu.Unlock()
	uid := c.User.ID
	if _, ok := w.userClients[uid]; !ok {
		w.userClients[uid] = make(ConnStorage)
	}
	w.userClients[uid][c.conn] = c
	return w.userClients[uid]
}

// GetActiveTokenIDs gets all active token IDs for a specific user
// GetActiveTokenIDs 获取特定用户的所有活动令牌 ID
func (w *WebsocketServer) GetActiveTokenIDs(uid int64) map[int64]bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	activeTokens := make(map[int64]bool)
	uidStr := strconv.FormatInt(uid, 10)
	if clients, ok := w.userClients[uidStr]; ok {
		for _, client := range clients {
			if client.TokenID > 0 {
				activeTokens[client.TokenID] = true
			}
		}
	}
	return activeTokens
}

// GetActiveTokenClients gets all active token IDs and their client names for a specific user
// GetActiveTokenClients 获取特定用户的所有活动令牌 ID 及其对应的客户端名称
func (w *WebsocketServer) GetActiveTokenClients(uid int64) map[int64][]string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	activeClients := make(map[int64][]string)
	uidStr := strconv.FormatInt(uid, 10)
	if clients, ok := w.userClients[uidStr]; ok {
		for _, client := range clients {
			if client.TokenID > 0 {
				if _, exists := activeClients[client.TokenID]; !exists {
					activeClients[client.TokenID] = []string{}
				}
				names := activeClients[client.TokenID]
				clientName := client.ClientName()
				nameExists := false
				for _, name := range names {
					if name == clientName {
						nameExists = true
						break
					}
				}
				if !nameExists && clientName != "" {
					activeClients[client.TokenID] = append(names, clientName)
				}
			}
		}
	}
	return activeClients
}

// UpdateTokenScope updates the scope of all active connections for a specific token
// UpdateTokenScope 更新特定令牌所有活动连接的权限范围
func (w *WebsocketServer) UpdateTokenScope(uid int64, tokenID int64, newScope string) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	uidStr := strconv.FormatInt(uid, 10)
	if clients, ok := w.userClients[uidStr]; ok {
		for _, client := range clients {
			if client.TokenID == tokenID {
				log(LogInfo, "WS UpdateTokenScope", zap.Int64("uid", uid), zap.Int64("tokenID", tokenID), zap.String("newScope", newScope))
				client.Scope = newScope
			}
		}
	}
}

// KickToken closes all connections for a specific token
// KickToken 关闭特定令牌的所有连接
func (w *WebsocketServer) KickToken(uid int64, tokenID int64) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	uidStr := strconv.FormatInt(uid, 10)
	if clients, ok := w.userClients[uidStr]; ok {
		for _, client := range clients {
			if client.TokenID == tokenID {
				log(LogInfo, "WS KickToken", zap.Int64("uid", uid), zap.Int64("tokenID", tokenID))
				client.conn.WriteClose(1000, []byte("TokenRotatedOrRevoked"))
			}
		}
	}
}

// CloseAllConnections sends a close frame to all active WebSocket connections.
// This must be called before shutting down the Worker Pool and Write Queue Manager
// to ensure hijacked WebSocket connections are properly terminated.
// 向所有活跃的 WebSocket 连接发送关闭帧。
// 必须在关闭 Worker Pool 和 Write Queue Manager 之前调用，以确保被劫持的 WebSocket 连接被正确终止。
func (w *WebsocketServer) CloseAllConnections() {
	w.mu.RLock()
	clients := make([]*WebsocketClient, 0, len(w.clients))
	for _, c := range w.clients {
		clients = append(clients, c)
	}
	w.mu.RUnlock()

	for _, c := range clients {
		if c.conn != nil {
			_ = c.conn.WriteClose(1001, []byte("server shutting down"))
		}
	}
}

// WaitAllClosed waits for all WebSocket connections to be fully closed (OnClose completed).
// Returns when all connections are closed or timeout is reached.
// 等待所有 WebSocket 连接完全关闭（OnClose 回调执行完毕）。
// 在所有连接关闭或超时后返回。
func (w *WebsocketServer) WaitAllClosed(timeout time.Duration) {
	done := make(chan struct{})
	go func() {
		w.connWg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		log(LogWarn, "WaitAllClosed: timeout waiting for WebSocket connections to close",
			zap.Int("remaining", len(w.clients)))
	}
}

func (w *WebsocketServer) RemoveUserClient(c *WebsocketClient) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if clients, ok := w.userClients[c.User.ID]; ok {
		delete(clients, c.conn)
		if len(clients) == 0 {
			delete(w.userClients, c.User.ID)
		}
	}
	log(LogInfo, "WS Client Remove", zap.Int("userCount", len(w.clients)))
}

// SetSession sets global binary upload session
// SetSession 设置全局二进制上传会话
func (w *WebsocketServer) SetSession(uid string, sessionID string, session any) {
	w.sessionsMu.Lock()
	defer w.sessionsMu.Unlock()
	if w.binaryChunkSessions[uid] == nil {
		w.binaryChunkSessions[uid] = make(map[string]any)
	}
	w.binaryChunkSessions[uid][sessionID] = session
}

// GetSession gets global binary upload session
// GetSession 获取全局二进制上传会话
func (w *WebsocketServer) GetSession(uid string, sessionID string) any {
	w.sessionsMu.RLock()
	defer w.sessionsMu.RUnlock()
	if userSessions, ok := w.binaryChunkSessions[uid]; ok {
		return userSessions[sessionID]
	}
	return nil
}

// GetSessionByPathHash gets global binary upload session by path hash
// GetSessionByPathHash 通过路径哈希获取全局二进制上传会话
//go:noinline
func (w *WebsocketServer) GetSessionByPathHash(uid string, pathHash string) any {
	w.sessionsMu.RLock()
	defer w.sessionsMu.RUnlock()
	if userSessions, ok := w.binaryChunkSessions[uid]; ok {
		for _, session := range userSessions {
			if getter, ok := session.(PathHashGetter); ok {
				if getter.GetPathHash() == pathHash {
					return session
				}
			}
		}
	}
	return nil
}

// RemoveSession removes global binary upload session
// RemoveSession 移除全局二进制上传会话
func (w *WebsocketServer) RemoveSession(uid string, sessionID string) {
	w.sessionsMu.Lock()
	defer w.sessionsMu.Unlock()
	if userSessions, ok := w.binaryChunkSessions[uid]; ok {
		delete(userSessions, sessionID)
		if len(userSessions) == 0 {
			delete(w.binaryChunkSessions, uid)
		}
	}
}

// CleanSessionsByPathHash cleans up existing sessions for a specific path hash of a user
// CleanSessionsByPathHash 清理用户特定路径哈希的现有会话
func (w *WebsocketServer) CleanSessionsByPathHash(uid string, pathHash string) {
	w.sessionsMu.Lock()
	defer w.sessionsMu.Unlock()

	userSessions, ok := w.binaryChunkSessions[uid]
	if !ok {
		return
	}

	var sessionIDsToRemove []string
	for sessionID, session := range userSessions {
		if getter, ok := session.(PathHashGetter); ok {
			if getter.GetPathHash() == pathHash {
				sessionIDsToRemove = append(sessionIDsToRemove, sessionID)
			}
		}
	}

	for _, sessionID := range sessionIDsToRemove {
		session := userSessions[sessionID]
		delete(userSessions, sessionID)

		if cleaner, ok := session.(SessionCleaner); ok {
			go cleaner.Cleanup()
		}
	}

	if len(userSessions) == 0 {
		delete(w.binaryChunkSessions, uid)
	}
}

func (w *WebsocketServer) OnOpen(conn *gws.Conn) {
	log(LogInfo, "WS Client Connect", zap.Int("Count", len(w.clients)))
	_ = conn.SetDeadline(time.Now().Add(w.config.PingWait * time.Second))
}

// isNormalDisconnectError 检查给定错误是否为正常的断开连接或网络中断错误
// isNormalDisconnectError checks if the given error is a normal disconnect or network interruption error
func isNormalDisconnectError(err error) bool {
	if err == nil {
		return true
	}
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return true
	}
	errStr := err.Error()
	// 检查常见的网络关闭、重置或超时错误消息
	// Check common network closed, reset, or timeout error messages
	if strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "i/o timeout") ||
		strings.Contains(errStr, "closed network connection") ||
		strings.Contains(errStr, "unexpected EOF") {
		return true
	}
	return false
}

func (w *WebsocketServer) OnClose(conn *gws.Conn, err error) {
	defer w.connWg.Done()

	c := w.GetClient(conn)
	if c == nil {
		return
	}

	// First cancel the context of the WebSocket connection to notify all ongoing operations to stop
	// 首先取消 WebSocket 连接的 context，通知所有正在进行的操作停止
	// This must be performed before cleaning up other resources to ensure that all operations dependent on the context can receive the cancellation signal
	// 这必须在清理其他 resource 之前执行，以确保所有依赖 context 的操作能够收到取消信号
	c.cancelContext()

	w.RemoveClient(conn)

	if c.User != nil {
		select {
		case c.done <- struct{}{}:
		default:
		}
		logLevel := LogInfo
		if err != nil && !isNormalDisconnectError(err) {
			logLevel = LogError
		}
		log(logLevel, "WS User Leave", zap.String("uid", c.User.ID), zap.String("traceID", c.TraceID), zap.Error(err))
		w.RemoveUserClient(c)
	} else {
		logLevel := LogInfo
		if err != nil && !isNormalDisconnectError(err) {
			logLevel = LogError
		}
		log(logLevel, "WS Client Leave (Unauth)", zap.String("traceID", c.TraceID), zap.Error(err))
	}

	// No longer clean up BinaryChunkSessions in OnClose, rely on the timeout mechanism for automatic cleanup instead
	// 不再在 OnClose 中清理 BinaryChunkSessions，改为依赖超时机制自动清理
	// However, clean up stale sessions (older than 10 minutes) to prevent memory leaks from zombie connections.
	// 但是清理超过 10 分钟的过期会话，防止僵尸连接导致内存泄漏。
	// Recent sessions are kept to support reconnection during network fluctuations.
	// 保留近期会话以支持网络波动期间的重连。
	if c.User != nil {
		w.cleanupStaleSessions(c.User.ID, 10*time.Minute)
	}

	// Clean up all DiffMergePaths entries
	// 清理所有 DiffMergePaths 条目
	if diffMergeCount := c.ClearAllDiffMergePaths(); diffMergeCount > 0 {
		log(LogInfo, "OnClose: cleared DiffMergePaths on disconnect",
			zap.Int("count", diffMergeCount),
			zap.String("traceID", c.TraceID))
	}

	log(LogInfo, "WS Client Leave", zap.Int("Count", len(w.clients)), zap.String("traceID", c.TraceID))

}

func (w *WebsocketServer) OnPing(socket *gws.Conn, payload []byte) {
	_ = socket.SetDeadline(time.Now().Add(w.config.PingWait * time.Second))
	_ = socket.WritePong(nil)
}

func (w *WebsocketServer) OnPong(socket *gws.Conn, payload []byte) {
	_ = socket.SetDeadline(time.Now().Add(w.config.PingWait * time.Second))
	if c := w.GetClient(socket); c != nil {
		c.lastPongAt.Store(time.Now().Unix())
	}
}

func (w *WebsocketServer) OnMessage(conn *gws.Conn, message *gws.Message) {
	defer message.Close()
	if message.Opcode != gws.OpcodeText && message.Opcode != gws.OpcodeBinary {
		return
	}
	if message.Data.String() == "close" {
		conn.WriteClose(1000, []byte("ClientClose"))
		return
	}

	c := w.GetClient(conn)
	if c == nil {
		return
	}

	// Set deadline
	// 设置延时
	_ = conn.SetDeadline(time.Now().Add(w.config.PingWait * time.Second))

	if message.Opcode == gws.OpcodeBinary {
		data := message.Data.Bytes()
		if len(data) < 2 {
			log(LogError, "WS OnMessage Binary too short", zap.String("uid", c.User.ID))
			return
		}
		prefix := string(data[:2])
		payload := data[2:]

		// Create a deep copy of the payload to prevent gws from recycling or reusing the underlying buffer during asynchronous processing
		// 创建 payload 的深拷贝，防止异步处理时底层缓冲区被 gws 回收或重用
		payloadCopy := make([]byte, len(payload))
		copy(payloadCopy, payload)

		if handler, ok := w.binaryHandlers[prefix]; ok {
			// Submit task through Worker Pool
			// 通过 Worker Pool 提交任务
			err := w.app.SubmitTaskAsync(c.Context(), func(ctx context.Context) error {
				// Check if context is cancelled
				// 检查 context 是否已取消
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
				// Verify binary message permission (currently only "00" for file chunk upload)
				if !VerifyPermissions(c.Scope, "ws", c.ClientType(), "file_w") {
					log(LogWarn, "WS OnMessage Binary Permission Denied", zap.String("prefix", prefix), zap.String("uid", c.User.ID))
					c.ToResponse(code.ErrorAuthTokenScopeRestricted.WithDetails("Permission denied: binary " + prefix))
					return nil
				}
				handler(c, payloadCopy)
				return nil
			})
			if err != nil {
				// Worker Pool is full or closed, record error and return error response
				// Worker Pool 满载或已关闭，记录错误并返回错误响应
				log(LogError, "WS OnMessage Worker Pool error",
					zap.String("prefix", prefix),
					zap.String("uid", c.User.ID),
					zap.Error(err))
				c.ToResponse(code.ErrorServerBusy)
				return
			}
		} else if prefix == "pb" {
			if !c.UseProtobuf() {
				log(LogWarn, "WS OnMessage received Protobuf but UseProtobuf is false", zap.String("uid", c.User.ID))
				return
			}
			if w.EnvelopeDecoder == nil {
				log(LogError, "WS OnMessage EnvelopeDecoder is nil", zap.String("uid", c.User.ID))
				return
			}

			action, innerPayload, err := w.EnvelopeDecoder(payloadCopy)
			if err != nil {
				log(LogError, "WS OnMessage Protobuf Envelope decode failed", zap.Error(err), zap.String("uid", c.User.ID))
				return
			}

			msg := WebSocketMessage{
				Type: action,
				Data: innerPayload,
			}

			if noAuthHandler, exists := w.noAuthHandlers[msg.Type]; exists {
				noAuthHandler(c, &msg)
				return
			}

			for _, interceptor := range w.interceptors {
				if !interceptor(c, &msg) {
					return
				}
			}

			handler, exists := w.handlers[msg.Type]
			if exists {
				handler(c, &msg)
			} else {
				log(LogError, "WS Unknown Message (Protobuf)", zap.String("Type", msg.Type), zap.String("uid", func() string {
					if c.User != nil {
						return c.User.ID
					}
					return "Guest"
				}()))
			}
		} else {
			log(LogWarn, "WS OnMessage Unknown Binary Prefix", zap.String("prefix", prefix))
		}
		return
	}

	messageStr := message.Data.String()
	// Use strings.Index to find the position of the separator
	// 使用 strings.Index 找到分隔符的位置
	index := strings.Index(messageStr, "|")

	//log(LogInfo, "WS OnMessage", zap.String("data", messageStr))

	var msg WebSocketMessage
	if index != -1 {
		msg.Type = messageStr[:index]           // Extract the part before the separator // 提取分隔符之前的部分
		msg.Data = []byte(messageStr[index+1:]) // Extract the part after the separator // 提取分隔符之后的部分
	} else {
		log(LogError, "WS OnMessage", zap.String("type", "Illegal message type"), zap.String("uid", c.User.ID))
		return
	}

	// Prioritize matching and executing unauthenticated handlers
	// 优先匹配并执行免登录鉴权的消息处理器
	if noAuthHandler, exists := w.noAuthHandlers[msg.Type]; exists {
		noAuthHandler(c, &msg)
		return
	}

	// Execute all registered pre-handler interceptors
	// 执行所有注册的前置拦截器。若任何一个返回 false，则中断后续执行
	for _, interceptor := range w.interceptors {
		if !interceptor(c, &msg) {
			return
		}
	}

	// Execute operation
	// 执行操作
	// Execute core business handler
	// 执行核心业务处理器
	handler, exists := w.handlers[msg.Type]
	if exists {
		handler(c, &msg)
	} else {
		log(LogError, "WS Unknown Message", zap.String("Type", msg.Type), zap.String("uid", func() string {
			if c.User != nil {
				return c.User.ID
			}
			return "Guest"
		}()))
	}
}

func (w *WebsocketServer) BroadcastToUser(uid int64, code *code.Code, action string) {
	uidStr := strconv.FormatInt(uid, 10)
	w.mu.RLock()
	defer w.mu.RUnlock()

	userClients, ok := w.userClients[uidStr]
	if !ok || len(userClients) == 0 {
		return
	}

	var responseBytes []byte
	content := Res{
		Code:    code.Code(),
		Status:  code.Status(),
		Message: code.Lang.GetMessage(),
		Data:    code.Data(),
	}

	if code.HaveDetails() {
		content.Details = strings.Join(code.Details(), ",")
	}

	if code.HaveVault() {
		content.Vault = code.Vault()
	}

	responseBytes, _ = json.Marshal(content)

	if action != "" {
		responseBytes = []byte(fmt.Sprintf(`%s|%s`, action, string(responseBytes)))
	}

	var b = gws.NewBroadcaster(gws.OpcodeText, responseBytes)
	defer b.Close()

	for _, uc := range userClients {
		if uc.conn == nil {
			continue
		}
		if err := b.Broadcast(uc.conn); err != nil {
			if uc.failCount.Add(1) == 4 {
				uc.conn.WriteClose(1000, []byte("broadcast failed"))
			}
		} else {
			uc.failCount.Store(0)
		}
	}
}

// cleanupStaleSessions removes BinaryChunkSessions older than maxAge for a given user.
// This prevents memory leaks from zombie connections whose timeout goroutines never fired.
// cleanupStaleSessions 清理指定用户超过 maxAge 的 BinaryChunkSessions。
// 防止僵尸连接的超时 goroutine 未触发时导致的内存泄漏。
func (w *WebsocketServer) cleanupStaleSessions(uid string, maxAge time.Duration) {
	w.sessionsMu.Lock()
	defer w.sessionsMu.Unlock()

	userSessions, ok := w.binaryChunkSessions[uid]
	if !ok {
		return
	}

	var staleIDs []string
	for sessionID, session := range userSessions {
		if getter, ok := session.(SessionCreatedAtGetter); ok {
			if time.Since(getter.GetCreatedAt()) > maxAge {
				staleIDs = append(staleIDs, sessionID)
			}
		}
	}

	for _, sessionID := range staleIDs {
		session := userSessions[sessionID]
		delete(userSessions, sessionID)
		if cleaner, ok := session.(SessionCleaner); ok {
			go cleaner.Cleanup()
		}
		log(LogInfo, "cleanupStaleSessions: removed stale session",
			zap.String("uid", uid),
			zap.String("sessionID", sessionID))
	}

	if len(userSessions) == 0 {
		delete(w.binaryChunkSessions, uid)
	}
}

