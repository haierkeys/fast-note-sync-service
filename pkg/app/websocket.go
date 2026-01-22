package app

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/logger"
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
	LogInfo        LogType = "info"
	LogError       LogType = "error"
	LogWarn        LogType = "warn"
	LogDebug       LogType = "debug"
)

// traceIDKeyType 用于在 context 中存储 Trace ID
type traceIDKeyType struct{}

// TraceIDKey 是 context 中存储 Trace ID 的 key
var TraceIDKey = traceIDKeyType{}

// GetTraceID 从 context 中获取 Trace ID
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// generateTraceID 生成新的 Trace ID
func generateTraceID() string {
	return uuid.New().String()
}

// extractOrGenerateTraceID 从 HTTP 请求中提取或生成 Trace ID
func extractOrGenerateTraceID(c *gin.Context) string {
	// 尝试从 Header 中获取
	if traceID := c.GetHeader("X-Trace-ID"); traceID != "" {
		return traceID
	}
	if traceID := c.GetHeader("X-Request-ID"); traceID != "" {
		return traceID
	}
	// 生成新的 Trace ID
	return generateTraceID()
}

// wsLogger 是 WebSocket 模块使用的日志器（通过 App Container 注入）
var wsLogger *zap.Logger

// wsProductionMode 标记是否为生产模式（通过 App Container 注入）
var wsProductionMode bool

// SetWSLogger 设置 WebSocket 模块的日志器
func SetWSLogger(logger *zap.Logger) {
	wsLogger = logger
}

// SetWSProductionMode 设置 WebSocket 模块的生产模式标记
func SetWSProductionMode(production bool) {
	wsProductionMode = production
}

// isDevelopmentMode 检查是否为开发环境
// 开发环境下会输出彩色控制台日志
func isDevelopmentMode() bool {
	return !wsProductionMode
}

// log 记录日志
// t: 日志类型
// msg: 日志消息
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

// logWithTraceID 记录日志，包含 Trace ID
func logWithTraceID(t LogType, traceID string, msg string, fields ...zap.Field) {
	if traceID != "" {
		fields = append([]zap.Field{zap.String("traceId", traceID)}, fields...)
	}
	log(t, msg, fields...)
}

// NoteModifyLog 记录 WebSocket 操作日志
// 同时支持结构化日志和开发环境彩色输出
// traceID: 追踪 ID
// uid: 用户 ID
// action: 执行的操作名称
// params: 可变参数，通常第一个为 Path，第二个为 Vault
func NoteModifyLog(traceID string, uid int64, action string, params ...string) {
	var path, vault string

	if len(params) > 0 {
		path = params[0]
	}

	if len(params) > 1 {
		vault = params[1]
	}

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

	// 开发环境保留彩色控制台输出，便于本地调试
	if isDevelopmentMode() {
		printColoredLog(uid, action, traceID, vault, path)
	}
}

// printColoredLog 输出彩色日志（仅开发环境）
// 使用 ANSI 转义码实现彩色输出
func printColoredLog(uid int64, action, traceID, vault, path string) {
	str := fmt.Sprintf("[WS] | \033[30;43m %d \033[0m\033[97;44m %s \033[0m", uid, action)

	if traceID != "" && len(traceID) >= 8 {
		str += fmt.Sprintf("\033[90m[%s]\033[0m ", traceID[:8]) // 只显示前8位以保持简洁
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
	Type string `json:"type"` // 操作类型，例如 "upload", "update", "delete"
	Data []byte `json:"data"` // 文件数据（仅在上传和更新时使用）
}

type ClientInfoMessage struct {
	Name                string `json:"name"`                // 客户端名称
	Version             string `json:"version"`             // 客户端版本
	OfflineSyncStrategy string `json:"offlineSyncStrategy"` // 离线设备同步策略 "newTimeMerge" | "ignoreTimeMerge"
}

type WSConfig struct {
	GWSOption    gws.ServerOption
	PingInterval time.Duration
	PingWait     time.Duration
}

// SessionCleaner 接口，用于在连接断开时清理会话资源
type SessionCleaner interface {
	Cleanup()
}

// DiffMergeEntry 表示 DiffMergePaths 中的条目
// 包含创建时间戳，用于超时清理机制
type DiffMergeEntry struct {
	CreatedAt time.Time // 条目创建时间
}

// WebsocketClient 结构体来存储每个 WebSocket 连接及其相关状态
type WebsocketClient struct {
	conn                *gws.Conn                 // WebSocket 底层连接句柄
	done                chan struct{}             // 关闭信号通道，用于优雅关闭读/写协程
	app                 AppContainer              // App Container 引用
	Server              *WebsocketServer          // WebSocket 服务器引用，用于访问全局状态（如会话）
	Ctx                 *gin.Context              // 原始 HTTP 升级请求的上下文（可用于获取 Header、Query 等）
	WsCtx               context.Context           // WebSocket 连接的长生命周期 context
	WsCancel            context.CancelFunc        // 用于取消 WsCtx
	TraceID             string                    // 连接的追踪 ID
	User                *UserEntity               // 已认证用户信息，通常在握手阶段绑定
	UserClients         *ConnStorage              // 用户连接池，支持多设备在线时广播或单点通信
	SF                  *singleflight.Group       // 并发控制：相同 key 的请求只执行一次，其余等待结果
	BinaryMu            sync.Mutex                // 用于读写数据时的同步锁 (不再保护 map 存储)
	ClientName          string                    // 客户端名称 (例如 "Mac", "Windows", "iPhone")
	ClientVersion       string                    // 客户端版本号 (例如 "1.2.4")
	IsFirstSync         bool                      // 是否是第一次同步过
	DiffMergePaths      map[string]DiffMergeEntry // 需要合并的文件路径，包含创建时间用于超时清理
	DiffMergePathsMu    sync.RWMutex              // 互斥锁，防止并发冲突
	OfflineSyncStrategy string                    // 离线设备同步策略 "newTimeMerge" | "ignoreTimeMerge"
}

// initContext 初始化 WebSocket 连接的 context
// 在连接建立时调用
func (c *WebsocketClient) initContext(traceID string) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, TraceIDKey, traceID)
	c.WsCtx, c.WsCancel = context.WithCancel(ctx)
	c.TraceID = traceID
}

// cancelContext 取消 WebSocket 连接的 context
// 在连接关闭时调用
func (c *WebsocketClient) cancelContext() {
	if c.WsCancel != nil {
		c.WsCancel()
	}
}

// Context 返回 WebSocket 连接的 context
// 用于所有需要 context 的操作（数据库查询、外部调用等）
func (c *WebsocketClient) Context() context.Context {
	if c.WsCtx == nil {
		panic("WebsocketClient.WsCtx is not initialized")
	}
	return c.WsCtx
}

// WithTimeout 创建带超时的子 context
// 用于需要超时控制的操作
func (c *WebsocketClient) WithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.WsCtx, timeout)
}

// CleanupExpiredDiffMergePaths 清理过期的 DiffMergePaths 条目
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

// ClearAllDiffMergePaths 清理所有 DiffMergePaths 条目
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

// 基于全局验证器的 WebSocket 版本参数绑定和验证工具函数
func (c *WebsocketClient) BindAndValid(data []byte, obj any) (bool, ValidErrors) {
	var errs ValidErrors

	// Step 1: JSON 反序列化（可替换成其他格式）
	if err := sonic.Unmarshal(data, obj); err != nil {
		// 解码错误处理
		errs = append(errs, &ValidError{
			Key:     "body",
			Message: "Invalid message format",
		})
		return false, errs
	}

	// Step 2: 参数验证
	validator := c.app.Validator()
	if validator == nil {
		return true, nil
	}
	if err := validator.ValidateStruct(obj); err != nil {
		// 如果验证失败，检查错误类型
		if validationErrors, ok := err.(validatorV10.ValidationErrors); ok {
			// 获取翻译器
			v := c.Ctx.Value("trans")
			trans := v.(ut.Translator)

			// 遍历验证错误并进行翻译
			for _, validationErr := range validationErrors {
				translatedMsg := validationErr.Translate(trans) // 翻译错误消息
				errs = append(errs, &ValidError{
					Key:     validationErr.Field(),
					Message: translatedMsg,
				})
			}
		}
		return false, errs // 返回验证错误
	}
	return true, nil
}

// 定期发送 Ping 消息
func (c *WebsocketClient) PingLoop(PingInterval time.Duration) {
	ticker := time.NewTicker(PingInterval * time.Second) // 每 25 秒发送一次 Ping
	defer ticker.Stop()
	for {
		select {
		case <-c.done:
			log(LogInfo, "WS Client Close Ping")
			return
		case <-ticker.C:
			if c.conn == nil {
				return
			}
			if err := c.conn.WritePing(nil); err != nil {
				// 连接关闭时的正常错误，降低日志级别
				if strings.Contains(err.Error(), "use of closed network connection") {
					log(LogDebug, "WS Client Ping: connection closed")
				} else {
					log(LogError, "WS Client Ping err ", zap.Error(err))
				}
				return
			}
			// log(LogInfo, "WS Client Ping", zap.String("uid", c.User.ID))
		}
	}
}

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
		Message: code.Lang.GetMessage(),
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

	responseBytes, _ = sonic.Marshal(content)

	if actionType != "" {
		responseBytes = []byte(fmt.Sprintf(`%s|%s`, actionType, string(responseBytes)))
	}

	if c.app.IsReturnSuccess() || actionType != "" || code.Code() > 200 || code.HaveData() || code.HaveDetails() {
		c.send(responseBytes, false, false)
	}
}

// BroadcastResponse 将结果转换为 JSON 格式并广播给当前用户的所有连接客户端
//
// 参数:
//
//	code: 业务响应状态码对象，包含状态码、消息和数据
//	options: 可选参数列表
//	  - options[0] (bool):   必填，是否排除当前客户端 (true: 排除自己, false: 广播给所有端)
//	  - options[1] (string): 选填，动作类型标识 (ActionType)，用于客户端区分消息类型
func (c *WebsocketClient) BroadcastResponse(code *code.Code, options ...any) {

	var actionType string
	if len(options) > 1 {
		actionType = options[1].(string)
	}

	if len(*c.UserClients) <= 0 {
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

	if code.HaveContext() {
		content.Context = code.Context()
	}

	responseBytes, _ = sonic.Marshal(content)

	if actionType != "" {
		responseBytes = []byte(fmt.Sprintf(`%s|%s`, actionType, string(responseBytes)))
	}

	c.send(responseBytes, true, options[0].(bool))
}

func (c *WebsocketClient) send(responseBytes []byte, isBroadcast bool, isExcludeSelf bool) {
	if isBroadcast {
		c.sendBroadcast(responseBytes, isExcludeSelf)
	} else {
		c.sendMessage(responseBytes)
	}
}

func (c *WebsocketClient) sendMessage(payload []byte) {
	c.conn.WriteMessage(gws.OpcodeText, payload)
}

func (c *WebsocketClient) sendBroadcast(payload []byte, isExcludeSelf bool) {
	var b = gws.NewBroadcaster(gws.OpcodeText, payload)
	defer b.Close()

	for _, uc := range *c.UserClients {
		if uc.conn == nil {
			continue
		}
		if isExcludeSelf && uc.conn == c.conn {
			continue
		}

		_ = b.Broadcast(uc.conn)
	}
}

// SendBinary 发送二进制消息
// prefix: 2字节前缀
func (c *WebsocketClient) SendBinary(prefix string, payload []byte) error {
	if c.conn == nil {
		return fmt.Errorf("connection is nil")
	}
	if len(prefix) != 2 {
		return fmt.Errorf("prefix must be 2 bytes")
	}
	// 拼接前缀和数据
	data := make([]byte, 2+len(payload))
	copy(data[0:2], prefix)
	copy(data[2:], payload)
	return c.conn.WriteMessage(gws.OpcodeBinary, data)
}

// ------------------------------------> WebsocketServer

type ConnStorage = map[*gws.Conn]*WebsocketClient

// AppContainer 定义 App Container 接口，用于解耦 pkg/app 和 internal/app
// 这个接口允许 WebsocketServer 使用 App Container 的功能而不产生循环依赖
type AppContainer interface {
	// Logger 获取日志器
	Logger() *zap.Logger
	// SubmitTask 提交任务到 Worker Pool
	SubmitTask(ctx context.Context, task func(context.Context) error) error
	// SubmitTaskAsync 异步提交任务到 Worker Pool（不等待结果）
	SubmitTaskAsync(ctx context.Context, task func(context.Context) error) error
	// Version 获取版本信息
	Version() VersionInfo
	// Validator 获取验证器（可能为 nil）
	Validator() ValidatorInterface
	// IsReturnSuccess 是否返回成功响应
	IsReturnSuccess() bool
	// GetAuthTokenKey 获取 Token 密钥
	GetAuthTokenKey() string
	// IsProductionMode 是否为生产模式
	IsProductionMode() bool
}

// VersionInfo 版本信息
type VersionInfo struct {
	Version   string
	GitTag    string
	BuildTime string
}

// ValidatorInterface 验证器接口
type ValidatorInterface interface {
	ValidateStruct(obj interface{}) error
}

type WebsocketServer struct {
	app               AppContainer // App Container（必须）
	handlers          map[string]func(*WebsocketClient, *WebSocketMessage)
	userVerifyHandler func(*WebsocketClient, int64) (*UserSelectEntity, error)
	binaryHandlers    map[string]func(*WebsocketClient, []byte) // 二进制消息处理器映射 prefix -> handler
	clients           ConnStorage
	userClients       map[string]ConnStorage
	mu                sync.Mutex
	up                *gws.Upgrader
	config            *WSConfig
	// 全局会话管理 (UID -> SessionID -> Session)
	binaryChunkSessions map[string]map[string]any
	sessionsMu          sync.RWMutex
}

// NewWebsocketServer 创建 WebSocket 服务器实例
// c: WebSocket 配置
// app: App Container（必须）
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

	// 设置 WebSocket 模块的日志器
	SetWSLogger(app.Logger())
	// 设置 WebSocket 模块的生产模式标记
	SetWSProductionMode(app.IsProductionMode())

	return &WebsocketServer{
		app:                 app,
		handlers:            make(map[string]func(*WebsocketClient, *WebSocketMessage)),
		binaryHandlers:      make(map[string]func(*WebsocketClient, []byte)),
		clients:             make(ConnStorage),
		userClients:         make(map[string]ConnStorage),
		config:              &c,
		binaryChunkSessions: make(map[string]map[string]any),
	}
}

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

		// 从 HTTP 请求中提取或生成 Trace ID
		traceID := extractOrGenerateTraceID(c)

		client := &WebsocketClient{
			conn:   socket,
			done:   make(chan struct{}),
			app:    w.app,
			Server: w,
			Ctx:    c,
			SF:     new(singleflight.Group),
		}

		// 初始化 WebSocket 连接的长生命周期 context
		client.initContext(traceID)

		w.AddClient(client)
		log(LogInfo, "WS Start", zap.String("type", "ReadLoop"), zap.String("traceID", traceID))
		go socket.ReadLoop()
	}
}

func (w *WebsocketServer) Use(action string, handler func(*WebsocketClient, *WebSocketMessage)) {
	w.handlers[action] = handler
}

func (w *WebsocketServer) UseUserVerify(handler func(*WebsocketClient, int64) (*UserSelectEntity, error)) {
	w.userVerifyHandler = handler
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
		c.ToResponse(code.ErrorInvalidUserAuthToken, "Authorization")
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

		// 用户有效性强制验证
		userSelect, err := w.userVerifyHandler(c, uid)
		if userSelect == nil || err != nil {
			log(LogError, "WS Authorization FAILD USER Not Exist", zap.Error(err))
			c.ToResponse(code.ErrorInvalidUserAuthToken, "Authorization")
			time.Sleep(2 * time.Second)
			c.conn.WriteClose(1000, []byte("AuthorizationFaild"))
			return
		}

		user.Nickname = userSelect.Nickname

		log(LogInfo, "WS Authorization", zap.String("uid", user.ID), zap.String("Nickname", user.Nickname))
		c.User = user
		w.AddUserClient(c)

		userClients := w.userClients[user.ID]

		c.BinaryMu.Lock()
		// 登录时不清理全局会话，确保重连后会话依然存在
		c.BinaryMu.Unlock()

		c.UserClients = &userClients

		versionInfo := w.app.Version()

		c.ToResponse(code.Success.WithData(map[string]string{
			"version":   versionInfo.Version,
			"gitTag":    versionInfo.GitTag,
			"buildTime": versionInfo.BuildTime,
		}), "Authorization")
		log(LogInfo, "WS User Enter", zap.String("uid", c.User.ID), zap.String("Nickname", c.User.Nickname), zap.Int("Count", len(userClients)))
		go c.PingLoop(w.config.PingInterval)
	}
}

func (w *WebsocketServer) ClientInfo(c *WebsocketClient, msg *WebSocketMessage) {
	var info ClientInfoMessage
	if err := sonic.Unmarshal(msg.Data, &info); err != nil {
		log(LogError, "WS ClientInfo Unmarshal FAILD", zap.Error(err))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(err.Error()))
		return
	}

	c.ClientName = info.Name
	c.ClientVersion = info.Version
	c.OfflineSyncStrategy = info.OfflineSyncStrategy
	c.DiffMergePaths = make(map[string]DiffMergeEntry)

	log(LogInfo, "WS ClientInfo", zap.String("uid", func() string {
		if c.User != nil {
			return c.User.ID
		}
		return "Guest"
	}()), zap.String("name", c.ClientName), zap.String("version", c.ClientVersion), zap.String("offlineSyncStrategy", c.OfflineSyncStrategy))

	c.ToResponse(code.Success, "ClientInfo")
}

func (w *WebsocketServer) GetClient(conn *gws.Conn) *WebsocketClient {
	w.mu.Lock()
	defer w.mu.Unlock()
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

func (w *WebsocketServer) AddUserClient(c *WebsocketClient) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.userClients[c.User.ID] == nil {
		w.userClients[c.User.ID] = make(ConnStorage)
	}
	w.userClients[c.User.ID][c.conn] = c
}

func (w *WebsocketServer) RemoveUserClient(c *WebsocketClient) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.userClients[c.User.ID], c.conn)
	log(LogInfo, "WS Client Remove", zap.Int("userCount", len(w.clients)))
}

// SetSession 设置全局二进制上传会话
func (w *WebsocketServer) SetSession(uid string, sessionID string, session any) {
	w.sessionsMu.Lock()
	defer w.sessionsMu.Unlock()
	if w.binaryChunkSessions[uid] == nil {
		w.binaryChunkSessions[uid] = make(map[string]any)
	}
	w.binaryChunkSessions[uid][sessionID] = session
}

// GetSession 获取全局二进制上传会话
func (w *WebsocketServer) GetSession(uid string, sessionID string) any {
	w.sessionsMu.RLock()
	defer w.sessionsMu.RUnlock()
	if userSessions, ok := w.binaryChunkSessions[uid]; ok {
		return userSessions[sessionID]
	}
	return nil
}

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

func (w *WebsocketServer) OnOpen(conn *gws.Conn) {
	log(LogInfo, "WS Client Connect", zap.Int("Count", len(w.clients)))
	_ = conn.SetDeadline(time.Now().Add(w.config.PingWait * time.Second))
}

func (w *WebsocketServer) OnClose(conn *gws.Conn, err error) {

	c := w.GetClient(conn)
	if c == nil {
		return
	}

	// 首先取消 WebSocket 连接的 context，通知所有正在进行的操作停止
	// 这必须在清理其他资源之前执行，以确保所有依赖 context 的操作能够收到取消信号
	c.cancelContext()

	w.RemoveClient(conn)

	if c.User != nil {
		select {
		case c.done <- struct{}{}:
		default:
		}
		log(LogInfo, "WS User Leave", zap.String("uid", c.User.ID), zap.String("traceID", c.TraceID), zap.Error(err))
		w.RemoveUserClient(c)
	} else {
		log(LogInfo, "WS Client Leave (Unauth)", zap.String("traceID", c.TraceID), zap.Error(err))
	}

	// 不再在 OnClose 中清理 BinaryChunkSessions，改为依赖超时机制自动清理
	// 这样可以支持在大文件上传过程中网络波动导致重连时，继续使用原有会话

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

	//设置延时
	_ = conn.SetDeadline(time.Now().Add(w.config.PingWait * time.Second))

	if message.Opcode == gws.OpcodeBinary {
		data := message.Data.Bytes()
		if len(data) < 2 {
			log(LogError, "WS OnMessage Binary too short", zap.String("uid", c.User.ID))
			return
		}
		prefix := string(data[:2])
		payload := data[2:]

		// 创建 payload 的深拷贝，防止异步处理时底层缓冲区被 gws 回收或重用
		payloadCopy := make([]byte, len(payload))
		copy(payloadCopy, payload)

		if handler, ok := w.binaryHandlers[prefix]; ok {
			// 通过 Worker Pool 提交任务
			err := w.app.SubmitTaskAsync(c.Context(), func(ctx context.Context) error {
				// 检查 context 是否已取消
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
				handler(c, payloadCopy)
				return nil
			})
			if err != nil {
				// Worker Pool 满载或已关闭，记录错误并返回错误响应
				log(LogError, "WS OnMessage Worker Pool error",
					zap.String("prefix", prefix),
					zap.String("uid", c.User.ID),
					zap.Error(err))
				c.ToResponse(code.ErrorServerBusy)
				return
			}
		} else {
			log(LogWarn, "WS OnMessage Unknown Binary Prefix", zap.String("prefix", prefix))
		}
		return
	}

	messageStr := message.Data.String()
	// 使用 strings.Index 找到分隔符的位置
	index := strings.Index(messageStr, "|")

	//log(LogInfo, "WS OnMessage", zap.String("data", messageStr))

	var msg WebSocketMessage
	if index != -1 {
		msg.Type = messageStr[:index]           // 提取分隔符之前的部分
		msg.Data = []byte(messageStr[index+1:]) // 提取分隔符之后的部分
	} else {
		log(LogError, "WS OnMessage", zap.String("type", "Illegal message type"), zap.String("uid", c.User.ID))
		return
	}

	if msg.Type == "Authorization" {
		w.Authorization(c, &msg)
		return
	}

	if msg.Type == "ClientInfo" {
		w.ClientInfo(c, &msg)
		return
	}

	// 验证用户是否登录
	if c.User == nil {
		log(LogWarn, "WS User not authenticated",
			zap.String("msgType", msg.Type),
			zap.String("traceId", c.TraceID))
		c.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	// 执行操作
	handler, exists := w.handlers[msg.Type]
	if exists {
		// Use the client object retrieved at the beginning of the function
		handler(c, &msg)
	} else {
		log(LogError, "WS Unknown Message", zap.String("Type", msg.Type), zap.String("uid", c.User.ID))
	}
}

func (w *WebsocketServer) BroadcastToUser(uid int64, code *code.Code, action string) {
	uidStr := strconv.FormatInt(uid, 10)
	w.mu.Lock()
	userClients, ok := w.userClients[uidStr]
	w.mu.Unlock()

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

	responseBytes, _ = sonic.Marshal(content)

	if action != "" {
		responseBytes = []byte(fmt.Sprintf(`%s|%s`, action, string(responseBytes)))
	}

	var b = gws.NewBroadcaster(gws.OpcodeText, responseBytes)
	defer b.Close()

	for _, uc := range userClients {
		if uc.conn == nil {
			continue
		}
		_ = b.Broadcast(uc.conn)
	}
}
