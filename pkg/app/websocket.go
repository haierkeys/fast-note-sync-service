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
	conn        *gws.Conn
	done        chan struct{}
	Ctx         *gin.Context
	User        *UserEntity
	UserClients *ConnStorage
	SF          *singleflight.Group // 用于处理并发请求的缓存
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

// BroadcastResponse 将结果转换为 JSON 格式并广播给所有客户端
// 第二个options参数为是否排除自己 第三个options参数为动作类型
func (c *WebsocketClient) BroadcastResponse(code *code.Code, options ...any) {

	var actionType string
	if len(options) > 1 {
		actionType = options[1].(string)
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
	handlers        map[string]func(*WebsocketClient, *WebSocketMessage)
	userDataHandler func(*WebsocketClient, int64) (*UserSelectEntity, error)
	clients         ConnStorage
	userClients     map[string]ConnStorage
	mu              sync.Mutex
	up              *gws.Upgrader
	config          *WebsocketServerConfig
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

func (w *WebsocketServer) UserDataSelectUse(handler func(*WebsocketClient, int64) (*UserSelectEntity, error)) {
	w.userDataHandler = handler
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
		userSelect, err := w.userDataHandler(c, uid)
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
	if message.Opcode != gws.OpcodeText {
		return
	}
	if message.Data.String() == "close" {
		conn.WriteClose(1000, []byte("ClientClose"))
		return
	}

	c := w.GetClient(conn)

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
