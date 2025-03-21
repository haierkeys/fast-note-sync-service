package app

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/haierkeys/obsidian-better-sync-service/global"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/code"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/lxzan/gws"
	"go.uber.org/zap"
)

type LogType string

const (
	WebSocketServerPingInterval         = 25
	WebSocketServerPingWait             = 50
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
		if global.Config.Server.RunMode == "debug" {
			global.Logger.Info(msg, fields...)
		}
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
	conn        *gws.Conn
	done        chan struct{}
	Ctx         *gin.Context
	user        *UserEntity
	userClients *ConnStorage
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
			if err := c.conn.WritePing(nil); err != nil {
				log(LogError, "WebsocketServer Client Ping err ", zap.Error(err))
				return
			}
			log(LogInfo, "WebsocketServer Client Ping", zap.String("uid", c.user.ID))
		}
	}
}

// ToResponse 将结果转换为 JSON 格式并发送给客户端
func (c *WebsocketClient) ToResponse(code *code.Code) {
	if code.HaveDetails() {
		details := strings.Join(code.Details(), ",")
		c.send(ResDetailsResult{
			Code:    code.Code(),
			Status:  code.Status(),
			Msg:     code.Lang.GetMessage(),
			Data:    code.Data(),
			Details: details,
		}, false, false)
	} else {
		c.send(ResResult{
			Code:   code.Code(),
			Status: code.Status(),
			Msg:    code.Lang.GetMessage(),
			Data:   code.Data(),
		}, false, false)
	}
}

func (c *WebsocketClient) BroadcastResponse(code *code.Code, options ...bool) {

	if code.HaveDetails() {
		details := strings.Join(code.Details(), ",")
		c.send(ResDetailsResult{
			Code:    code.Code(),
			Status:  code.Status(),
			Msg:     code.Lang.GetMessage(),
			Data:    code.Data(),
			Details: details,
		}, true, options[0])
	} else {
		c.send(ResResult{
			Code:   code.Code(),
			Status: code.Status(),
			Msg:    code.Lang.GetMessage(),
			Data:   code.Data(),
		}, true, options[0])
	}
}

func (c *WebsocketClient) send(content any, isBroadcast bool, isExcludeSelf bool) {
	responseBytes, _ := json.Marshal(content)
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

	for _, uc := range *c.userClients {
		if uc.conn == nil {
			continue
		}
		if isExcludeSelf && uc.conn == c.conn {
			continue
		}

		_ = b.Broadcast(uc.conn)
	}
}

//------------------------------------> WebsocketServer

type ConnStorage = map[*gws.Conn]*WebsocketClient
type WebsocketServer struct {
	handlers    map[string]func(*WebsocketClient, *WebSocketMessage)
	clients     ConnStorage
	userClients map[string]ConnStorage
	mu          sync.Mutex
	up          *gws.Upgrader
	config      *WebsocketServerConfig
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
		client := &WebsocketClient{conn: socket, done: make(chan struct{}), Ctx: c}
		w.AddClient(client)
		log(LogInfo, "WebsocketServer Start", zap.String("type", "ReadLoop"))
		go socket.ReadLoop()
	}
}

func (w *WebsocketServer) Use(action string, handler func(*WebsocketClient, *WebSocketMessage)) {
	w.handlers[action] = handler
}

func (w *WebsocketServer) Authorization(c *WebsocketClient, msg *WebSocketMessage) {
	if user, err := ParseToken(string(msg.Data)); err != nil {
		log(LogError, "WebsocketServer Authorization", zap.Error(err))
		c.ToResponse(code.ErrorInvalidUserAuthToken)
		c.conn.WriteMessage(gws.OpcodeCloseConnection, nil)
	} else {
		log(LogInfo, "WebsocketServer Authorization", zap.String("uid", user.ID))
		c.user = user
		w.AddUserClient(c)

		userClients := w.userClients[user.ID]
		c.userClients = &userClients
		c.ToResponse(code.Success)
		log(LogInfo, "WebsocketServer User enters", zap.String("uid", c.user.ID))
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
	if w.userClients[c.user.ID] == nil {
		w.userClients[c.user.ID] = make(ConnStorage)
	}
	w.userClients[c.user.ID][c.conn] = c
}

func (w *WebsocketServer) RemoveUserClient(c *WebsocketClient) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.userClients[c.user.IP], c.conn)
	log(LogInfo, "WebsocketServer Client Remove", zap.Int("userCount", len(w.clients)))
}

func (w *WebsocketServer) OnOpen(conn *gws.Conn) {
	log(LogInfo, "WebsocketServer Client Connect")
	log(LogInfo, "WebsocketServer Client Online", zap.Int("Count", len(w.clients)))
	_ = conn.SetDeadline(time.Now().Add(w.config.PingWait * time.Second))
}

func (w *WebsocketServer) OnClose(conn *gws.Conn, err error) {
	log(LogInfo, "WebsocketServer Client OnClose")
	c := w.GetClient(conn)

	w.RemoveClient(conn)
	// dump.P(c)
	if c.user != nil {
		c.done <- struct{}{}
		log(LogInfo, "WebsocketServer User Leave", zap.String("uid", c.user.ID))
		w.RemoveUserClient(c)
	}

	log(LogInfo, "WebsocketServer Client Online", zap.Int("Count", len(w.clients)))

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
	if message.Opcode != gws.OpcodeText {
		return
	}
	if message.Data.String() == "close" {
		conn.WriteMessage(gws.OpcodeCloseConnection, nil)
		return
	}

	c := w.GetClient(conn)

	messageStr := message.Data.String()
	// 使用 strings.Index 找到分隔符的位置
	index := strings.Index(messageStr, "|")

	var msg WebSocketMessage
	if index != -1 {
		msg.Type = messageStr[:index]           // 提取分隔符之前的部分
		msg.Data = []byte(messageStr[index+1:]) // 提取分隔符之后的部分
	} else {
		log(LogError, "WebsocketServer OnMessage", zap.String("type", "Illegal message type"), zap.String("uid", c.user.ID))
		return
	}

	if msg.Type == "Authorization" {
		w.Authorization(c, &msg)
		return
	}

	// 验证用户是否登录
	if c.user == nil {
		fmt.Println(msg.Type, c.user)
		c.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	// 执行操作
	handler, exists := w.handlers[msg.Type]
	if exists {
		log(LogInfo, "WebsocketServer OnMessage", zap.String("type", msg.Type))
		c := w.GetClient(conn)
		handler(c, &msg)
	} else {
		fmt.Println(msg.Type, string(msg.Data))
		log(LogError, "WebsocketServer OnMessage", zap.String("msg", "Unknown message type"))
	}
}
