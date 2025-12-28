package app

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"golang.org/x/sync/singleflight"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/lxzan/gws"
	"go.uber.org/zap"
)

type LogType string

const (
	WSPingInterval         = 25
	WSPingWait             = 40
	LogInfo        LogType = "info"
	LogError       LogType = "error"
	LogWarn        LogType = "warn"
)

func log(t LogType, msg string, fields ...zap.Field) {

	if t == "error" {
		global.Logger.Error(msg, fields...)
	} else if t == "warn" {
		global.Logger.Warn(msg, fields...)
	} else if t == "info" {
		global.Logger.Info(msg, fields...)

	}
}

type WebSocketMessage struct {
	Type string `json:"type"` // 操作类型，例如 "upload", "update", "delete"
	Data []byte `json:"data"` // 文件数据（仅在上传和更新时使用）
}

type ClientInfoMessage struct {
	Name    string `json:"name"`    // 客户端名称
	Version string `json:"version"` // 客户端版本
}

type WSConfig struct {
	GWSOption    gws.ServerOption
	PingInterval time.Duration
	PingWait     time.Duration
}

// WebsocketClient 结构体来存储每个 WebSocket 连接及其相关状态
type WebsocketClient struct {
	conn                *gws.Conn           // WebSocket 底层连接句柄
	done                chan struct{}       // 关闭信号通道，用于优雅关闭读/写协程
	Ctx                 *gin.Context        // 原始 HTTP 升级请求的上下文（可用于获取 Header、Query 等）
	User                *UserEntity         // 已认证用户信息，通常在握手阶段绑定
	UserClients         *ConnStorage        // 用户连接池，支持多设备在线时广播或单点通信
	SF                  *singleflight.Group // 并发控制：相同 key 的请求只执行一次，其余等待结果
	BinaryMu            sync.Mutex          // 二进制分块会话的互斥锁，防止并发冲突
	BinaryChunkSessions map[string]any      // 临时存储分块上传/下载的中间状态，例如文件重组缓冲区
	ClientName          string              // 客户端名称 (例如 "Mac", "Windows", "iPhone")
	ClientVersion       string              // 客户端版本号 (例如 "1.2.4")
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
	if err := global.Validator.Validate.Struct(obj); err != nil {

		// 如果验证失败，检查错误类型
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
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
				log(LogError, "WS Client Ping err ", zap.Error(err))
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
	if code.HaveDetails() {
		details := strings.Join(code.Details(), ",")
		c.send(actionType, ResDetailsResult{
			Code:    code.Code(),
			Status:  code.Status(),
			Msg:     code.Lang.GetMessage(),
			Data:    code.Data(),
			Details: details,
		}, false, false)
	} else {
		if global.Config.App.IsReturnSussess || actionType != "" || code.Code() > 200 || code.HaveData() {
			c.send(actionType, ResResult{
				Code:   code.Code(),
				Status: code.Status(),
				Msg:    code.Lang.GetMessage(),
				Data:   code.Data(),
			}, false, false)
		}
	}
	code.Reset()
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

	if code.HaveDetails() {
		details := strings.Join(code.Details(), ",")
		if code.HaveVault() {
			c.send(actionType, ResVaultDetailsResult{
				Code:    code.Code(),
				Status:  code.Status(),
				Msg:     code.Lang.GetMessage(),
				Data:    code.Data(),
				Details: details,
				Vault:   code.Vault(),
			}, true, options[0].(bool))
		} else {
			c.send(actionType, ResDetailsResult{
				Code:    code.Code(),
				Status:  code.Status(),
				Msg:     code.Lang.GetMessage(),
				Data:    code.Data(),
				Details: details,
			}, true, options[0].(bool))
		}
	} else {
		if code.HaveVault() {
			c.send(actionType, ResVaultResult{
				Code:   code.Code(),
				Status: code.Status(),
				Msg:    code.Lang.GetMessage(),
				Data:   code.Data(),
				Vault:  code.Vault(),
			}, true, options[0].(bool))
		} else {
			c.send(actionType, ResResult{
				Code:   code.Code(),
				Status: code.Status(),
				Msg:    code.Lang.GetMessage(),
				Data:   code.Data(),
			}, true, options[0].(bool))
		}
	}

	code.Reset()
}

func (c *WebsocketClient) send(actionType string, content any, isBroadcast bool, isExcludeSelf bool) {
	responseBytes, _ := sonic.Marshal(content)
	if actionType != "" {
		responseBytes = []byte(fmt.Sprintf(`%s|%s`, actionType, string(responseBytes)))
	}
	if isBroadcast {
		c.broadcast(responseBytes, isExcludeSelf)
	} else {
		c.message(responseBytes)
	}
}

func (c *WebsocketClient) message(payload []byte) {
	c.conn.WriteMessage(gws.OpcodeText, payload)
}

// SendBinary 发送二进制消息
// prefix: 2字节前缀
func (c *WebsocketClient) SendBinary(prefix string, payload []byte) error {
	if len(prefix) != 2 {
		return fmt.Errorf("prefix must be 2 bytes")
	}
	// 拼接前缀和数据
	data := make([]byte, 2+len(payload))
	copy(data[0:2], prefix)
	copy(data[2:], payload)
	return c.conn.WriteMessage(gws.OpcodeBinary, data)
}

func (c *WebsocketClient) broadcast(payload []byte, isExcludeSelf bool) {
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

// ------------------------------------> WS

type ConnStorage = map[*gws.Conn]*WebsocketClient

type WS struct {
	handlers          map[string]func(*WebsocketClient, *WebSocketMessage)
	userVerifyHandler func(*WebsocketClient, int64) (*UserSelectEntity, error)
	// binaryHandler     func(*WebsocketClient, []byte) // Deprecated: replaced by binaryHandlers
	binaryHandlers map[string]func(*WebsocketClient, []byte) // 二进制消息处理器映射 prefix -> handler
	clients        ConnStorage
	userClients    map[string]ConnStorage
	mu             sync.Mutex
	up             *gws.Upgrader
	config         *WSConfig
}

func NewWS(c WSConfig) *WS {
	if c.PingInterval == 0 {
		c.PingInterval = WSPingInterval
	}
	if c.PingWait == 0 {
		c.PingWait = WSPingWait
	}
	wss := WS{
		handlers:       make(map[string]func(*WebsocketClient, *WebSocketMessage)),
		binaryHandlers: make(map[string]func(*WebsocketClient, []byte)),
		clients:        make(ConnStorage),
		userClients:    make(map[string]ConnStorage),
		config:         &c,
	}
	return &wss
}

func (w *WS) Upgrade() {
	w.up = gws.NewUpgrader(w, &w.config.GWSOption)
}

func (w *WS) Run() gin.HandlerFunc {

	return func(c *gin.Context) {

		w.Upgrade()
		socket, err := w.up.Upgrade(c.Writer, c.Request)
		if err != nil {
			log(LogError, "WS Start err", zap.Error(err))
			return
		}
		client := &WebsocketClient{conn: socket, done: make(chan struct{}), Ctx: c, SF: new(singleflight.Group)}
		w.AddClient(client)
		log(LogInfo, "WS Start", zap.String("type", "ReadLoop"))
		go socket.ReadLoop()
	}
}

func (w *WS) Use(action string, handler func(*WebsocketClient, *WebSocketMessage)) {
	w.handlers[action] = handler
}

func (w *WS) UseUserVerify(handler func(*WebsocketClient, int64) (*UserSelectEntity, error)) {
	w.userVerifyHandler = handler
}

func (w *WS) UseBinary(prefix string, handler func(*WebsocketClient, []byte)) {
	if len(prefix) != 2 {
		panic("binary message prefix must be 2 characters")
	}
	w.binaryHandlers[prefix] = handler
}

func (w *WS) Authorization(c *WebsocketClient, msg *WebSocketMessage) {

	if user, err := ParseToken(string(msg.Data)); err != nil {
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
		// 清空所有会话(具体的文件清理由超时机制或 cleanupSession 处理)
		c.BinaryChunkSessions = make(map[string]any)
		c.BinaryMu.Unlock()

		c.UserClients = &userClients

		versionInfo := map[string]string{
			"version":   global.Version,
			"gitTag":    global.GitTag,
			"buildTime": global.BuildTime,
		}

		c.ToResponse(code.Success.WithData(versionInfo), "Authorization")
		log(LogInfo, "WS User Enter", zap.String("uid", c.User.ID), zap.String("Nickname", c.User.Nickname), zap.Int("Count", len(userClients)))
		go c.PingLoop(w.config.PingInterval)
	}
}

func (w *WS) ClientInfo(c *WebsocketClient, msg *WebSocketMessage) {
	var info ClientInfoMessage
	if err := sonic.Unmarshal(msg.Data, &info); err != nil {
		log(LogError, "WS ClientInfo Unmarshal FAILD", zap.Error(err))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(err.Error()))
		return
	}

	c.ClientName = info.Name
	c.ClientVersion = info.Version

	log(LogInfo, "WS ClientInfo", zap.String("uid", func() string {
		if c.User != nil {
			return c.User.ID
		}
		return "Guest"
	}()), zap.String("name", c.ClientName), zap.String("version", c.ClientVersion))

	c.ToResponse(code.Success, "ClientInfo")
}

func (w *WS) GetClient(conn *gws.Conn) *WebsocketClient {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.clients[conn]
}

func (w *WS) AddClient(c *WebsocketClient) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.clients[c.conn] = c
}

func (w *WS) RemoveClient(conn *gws.Conn) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.clients, conn)
}

func (w *WS) AddUserClient(c *WebsocketClient) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.userClients[c.User.ID] == nil {
		w.userClients[c.User.ID] = make(ConnStorage)
	}
	w.userClients[c.User.ID][c.conn] = c
}

func (w *WS) RemoveUserClient(c *WebsocketClient) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.userClients[c.User.ID], c.conn)
	log(LogInfo, "WS Client Remove", zap.Int("userCount", len(w.clients)))
}

func (w *WS) OnOpen(conn *gws.Conn) {
	log(LogInfo, "WS Client Connect", zap.Int("Count", len(w.clients)))
	_ = conn.SetDeadline(time.Now().Add(w.config.PingWait * time.Second))
}

func (w *WS) OnClose(conn *gws.Conn, err error) {

	c := w.GetClient(conn)
	if c == nil {
		return
	}

	w.RemoveClient(conn)

	if c.User != nil {
		c.done <- struct{}{}
		log(LogInfo, "WS User Leave", zap.String("uid", c.User.ID))
		w.RemoveUserClient(c)
	}

	// 清理所有未完成的上传会话
	if len(c.BinaryChunkSessions) > 0 {
		c.BinaryMu.Lock()
		sessionCount := len(c.BinaryChunkSessions)
		// 清空所有会话(具体的文件清理由超时机制或 cleanupSession 处理)
		c.BinaryChunkSessions = make(map[string]any)
		c.BinaryMu.Unlock()

		if sessionCount > 0 {
			log(LogWarn, "OnClose: cleared upload sessions on disconnect",
				zap.Int("sessionCount", sessionCount))
		}
	}

	log(LogInfo, "WS Client Leave", zap.Int("Count", len(w.clients)))

}

func (w *WS) OnPing(socket *gws.Conn, payload []byte) {
	_ = socket.SetDeadline(time.Now().Add(w.config.PingWait * time.Second))
	_ = socket.WritePong(nil)
}

func (w *WS) OnPong(socket *gws.Conn, payload []byte) {
	_ = socket.SetDeadline(time.Now().Add(w.config.PingWait * time.Second))
}

func (w *WS) OnMessage(conn *gws.Conn, message *gws.Message) {
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

	if message.Opcode == gws.OpcodeBinary {
		data := message.Data.Bytes()
		if len(data) < 2 {
			log(LogError, "WS OnMessage Binary too short", zap.String("uid", c.User.ID))
			return
		}
		prefix := string(data[:2])
		payload := data[2:]

		if handler, ok := w.binaryHandlers[prefix]; ok {
			handler(c, payload)
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
		fmt.Println(msg.Type, c.User)
		c.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	// 执行操作
	handler, exists := w.handlers[msg.Type]
	if exists {
		log(LogInfo, "WS Message "+msg.Type, zap.String("uid", c.User.ID))
		// Use the client object retrieved at the beginning of the function
		handler(c, &msg)
	} else {
		log(LogError, "WS Unknown Message", zap.String("Type", msg.Type), zap.String("uid", c.User.ID))
	}
}

func (w *WS) BroadcastToUser(uid int64, code *code.Code, action string) {
	uidStr := strconv.FormatInt(uid, 10)
	w.mu.Lock()
	userClients, ok := w.userClients[uidStr]
	w.mu.Unlock()

	if !ok || len(userClients) == 0 {
		return
	}

	var responseBytes []byte
	if code.HaveDetails() {
		details := strings.Join(code.Details(), ",")
		if code.HaveVault() {
			content := ResVaultDetailsResult{
				Code:    code.Code(),
				Status:  code.Status(),
				Msg:     code.Lang.GetMessage(),
				Data:    code.Data(),
				Details: details,
				Vault:   code.Vault(),
			}
			responseBytes, _ = sonic.Marshal(content)
		} else {
			content := ResDetailsResult{
				Code:    code.Code(),
				Status:  code.Status(),
				Msg:     code.Lang.GetMessage(),
				Data:    code.Data(),
				Details: details,
			}
			responseBytes, _ = sonic.Marshal(content)
		}

	} else {
		if code.HaveVault() {
			content := ResVaultResult{
				Code:   code.Code(),
				Status: code.Status(),
				Msg:    code.Lang.GetMessage(),
				Data:   code.Data(),
				Vault:  code.Vault(),
			}
			responseBytes, _ = sonic.Marshal(content)
		} else {
			content := ResResult{
				Code:   code.Code(),
				Status: code.Status(),
				Msg:    code.Lang.GetMessage(),
				Data:   code.Data(),
			}
			responseBytes, _ = sonic.Marshal(content)
		}
	}

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
	code.Reset()
}
