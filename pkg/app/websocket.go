package app

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

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
	WebSocketServerPingInterval         = 25
	WebSocketServerPingWait             = 40
	LogInfo                     LogType = "info"
	LogError                    LogType = "error"
	LogWarn                     LogType = "warn"
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

type WebsocketServerConfig struct {
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
}

// 基于全局验证器的 WebSocket 版本参数绑定和验证工具函数
func (c *WebsocketClient) BindAndValid(data []byte, obj any) (bool, ValidErrors) {
	var errs ValidErrors

	// Step 1: JSON 反序列化（可替换成其他格式）
	if err := json.Unmarshal(data, obj); err != nil {
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
			log(LogInfo, "WebsocketServer Client Close Ping")
			return
		case <-ticker.C:
			if c.conn == nil {
				return
			}
			if err := c.conn.WritePing(nil); err != nil {
				log(LogError, "WebsocketServer Client Ping err ", zap.Error(err))
				return
			}
			// log(LogInfo, "WebsocketServer Client Ping", zap.String("uid", c.User.ID))
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
		c.send(actionType, ResDetailsResult{
			Code:    code.Code(),
			Status:  code.Status(),
			Msg:     code.Lang.GetMessage(),
			Data:    code.Data(),
			Details: details,
		}, true, options[0].(bool))
	} else {
		c.send(actionType, ResResult{
			Code:   code.Code(),
			Status: code.Status(),
			Msg:    code.Lang.GetMessage(),
			Data:   code.Data(),
		}, true, options[0].(bool))
	}

	code.Reset()
}

func (c *WebsocketClient) send(actionType string, content any, isBroadcast bool, isExcludeSelf bool) {
	responseBytes, _ := json.Marshal(content)
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

// ------------------------------------> WebsocketServer

type ConnStorage = map[*gws.Conn]*WebsocketClient

type WebsocketServer struct {
	handlers          map[string]func(*WebsocketClient, *WebSocketMessage)
	userVerifyHandler func(*WebsocketClient, int64) (*UserSelectEntity, error)
	binaryHandler     func(*WebsocketClient, []byte)
	clients           ConnStorage
	userClients       map[string]ConnStorage
	mu                sync.Mutex
	up                *gws.Upgrader
	config            *WebsocketServerConfig
}

func NewWebsocketServer(c WebsocketServerConfig) *WebsocketServer {
	if c.PingInterval == 0 {
		c.PingInterval = WebSocketServerPingInterval
	}
	if c.PingWait == 0 {
		c.PingWait = WebSocketServerPingWait
	}
	wss := WebsocketServer{
		handlers:    make(map[string]func(*WebsocketClient, *WebSocketMessage)),
		clients:     make(ConnStorage),
		userClients: make(map[string]ConnStorage),
		config:      &c,
	}
	return &wss
}

func (w *WebsocketServer) Upgrade() {
	w.up = gws.NewUpgrader(w, &w.config.GWSOption)
}

func (w *WebsocketServer) Run() gin.HandlerFunc {

	return func(c *gin.Context) {

		w.Upgrade()
		socket, err := w.up.Upgrade(c.Writer, c.Request)
		if err != nil {
			log(LogError, "WebsocketServer Start err", zap.Error(err))
			return
		}
		client := &WebsocketClient{conn: socket, done: make(chan struct{}), Ctx: c, SF: new(singleflight.Group)}
		w.AddClient(client)
		log(LogInfo, "WebsocketServer Start", zap.String("type", "ReadLoop"))
		go socket.ReadLoop()
	}
}

func (w *WebsocketServer) Use(action string, handler func(*WebsocketClient, *WebSocketMessage)) {
	w.handlers[action] = handler
}

func (w *WebsocketServer) UseUserVerify(handler func(*WebsocketClient, int64) (*UserSelectEntity, error)) {
	w.userVerifyHandler = handler
}

func (w *WebsocketServer) UseBinary(handler func(*WebsocketClient, []byte)) {
	w.binaryHandler = handler
}

func (w *WebsocketServer) Authorization(c *WebsocketClient, msg *WebSocketMessage) {

	if user, err := ParseToken(string(msg.Data)); err != nil {
		log(LogError, "WebsocketServer Authorization FAILD", zap.Error(err))
		c.ToResponse(code.ErrorInvalidUserAuthToken, "Authorization")
		time.Sleep(2 * time.Second)
		c.conn.WriteClose(1000, []byte("AuthorizationFaild"))
	} else {

		uid, err := strconv.ParseInt(user.ID, 10, 64)
		if err != nil {
			log(LogError, "WebsocketServer Authorization FAILD", zap.Error(err))
			c.ToResponse(code.ErrorInvalidUserAuthToken, "Authorization")
			time.Sleep(2 * time.Second)
			c.conn.WriteClose(1000, []byte("AuthorizationFaild"))
			return
		}

		// 用户有效性强制验证
		userSelect, err := w.userVerifyHandler(c, uid)
		if userSelect == nil || err != nil {
			log(LogError, "WebsocketServer Authorization FAILD USER Not Exist", zap.Error(err))
			c.ToResponse(code.ErrorInvalidUserAuthToken, "Authorization")
			time.Sleep(2 * time.Second)
			c.conn.WriteClose(1000, []byte("AuthorizationFaild"))
			return
		}

		user.Nickname = userSelect.Nickname

		log(LogInfo, "WebsocketServer Authorization", zap.String("uid", user.ID), zap.String("Nickname", user.Nickname))
		c.User = user
		w.AddUserClient(c)

		userClients := w.userClients[user.ID]

		c.UserClients = &userClients
		c.ToResponse(code.Success, "Authorization")
		log(LogInfo, "WebsocketServer User Enters", zap.String("uid", c.User.ID), zap.String("Nickname", c.User.Nickname), zap.Int("Count", len(userClients)))
		go c.PingLoop(w.config.PingInterval)
	}
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
	log(LogInfo, "WebsocketServer Client Remove", zap.Int("userCount", len(w.clients)))
}

func (w *WebsocketServer) OnOpen(conn *gws.Conn) {
	log(LogInfo, "WebsocketServer Client Connect", zap.Int("Count", len(w.clients)))
	_ = conn.SetDeadline(time.Now().Add(w.config.PingWait * time.Second))
}

func (w *WebsocketServer) OnClose(conn *gws.Conn, err error) {

	c := w.GetClient(conn)

	w.RemoveClient(conn)

	if c.User != nil {
		c.done <- struct{}{}
		log(LogInfo, "WebsocketServer User Leave", zap.String("uid", c.User.ID))
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

	log(LogInfo, "WebsocketServer Client Leave", zap.Int("Count", len(w.clients)))

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

	if message.Opcode == gws.OpcodeBinary {
		if w.binaryHandler != nil {
			w.binaryHandler(c, message.Data.Bytes())
		}
		return
	}

	messageStr := message.Data.String()
	// 使用 strings.Index 找到分隔符的位置
	index := strings.Index(messageStr, "|")

	//log(LogInfo, "WebsocketServer OnMessage", zap.String("data", messageStr))

	var msg WebSocketMessage
	if index != -1 {
		msg.Type = messageStr[:index]           // 提取分隔符之前的部分
		msg.Data = []byte(messageStr[index+1:]) // 提取分隔符之后的部分
	} else {
		log(LogError, "WebsocketServer OnMessage", zap.String("type", "Illegal message type"), zap.String("uid", c.User.ID))
		return
	}

	if msg.Type == "Authorization" {
		w.Authorization(c, &msg)
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
		log(LogInfo, "WebsocketServer OnMessage", zap.String("Type", msg.Type))
		c := w.GetClient(conn)
		handler(c, &msg)
	} else {
		log(LogError, "WebsocketServer OnMessage", zap.String("msg", "Unknown message type"))
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
	if code.HaveDetails() {
		details := strings.Join(code.Details(), ",")
		content := ResDetailsResult{
			Code:    code.Code(),
			Status:  code.Status(),
			Msg:     code.Lang.GetMessage(),
			Data:    code.Data(),
			Details: details,
		}
		responseBytes, _ = json.Marshal(content)
	} else {
		content := ResResult{
			Code:   code.Code(),
			Status: code.Status(),
			Msg:    code.Lang.GetMessage(),
			Data:   code.Data(),
		}
		responseBytes, _ = json.Marshal(content)
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
