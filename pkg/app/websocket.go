package app

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/obsidian-better-sync-service/global"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/code"
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

type WebSocketMessage struct {
	Type     string `json:"type"`     // 操作类型，例如 "upload", "update", "delete"
	Filename string `json:"filename"` // 文件名
	Data     []byte `json:"data"`     // 文件数据（仅在上传和更新时使用）
}

// WebsocketClient 结构体来存储每个 WebSocket 连接及其相关状态
type WebsocketClient struct {
	conn *gws.Conn
	done chan struct{}
	user any
}

type WebsocketServerConfig struct {
	GWSOption    gws.ServerOption
	PingInterval time.Duration
	PingWait     time.Duration
}

type WebsocketServer struct {
	handlers map[string]func(*WebsocketClient, *WebSocketMessage)
	clients  map[*gws.Conn]*WebsocketClient
	mu       sync.Mutex
	up       *gws.Upgrader
	config   *WebsocketServerConfig
}

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

func NewWebsocketServer(c WebsocketServerConfig) *WebsocketServer {
	if c.PingInterval == 0 {
		c.PingInterval = WebSocketServerPingInterval
	}

	if c.PingWait == 0 {
		c.PingWait = WebSocketServerPingWait
	}

	return &WebsocketServer{
		handlers: make(map[string]func(*WebsocketClient, *WebSocketMessage)),
		clients:  make(map[*gws.Conn]*WebsocketClient),
		config:   &c,
	}
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
		log(LogInfo, "WebsocketServer Start", zap.String("type", "ReadLoop"))
		go socket.ReadLoop()
	}
}

func (w *WebsocketServer) Use(action string, handler func(*WebsocketClient, *WebSocketMessage)) {
	w.handlers[action] = handler
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

func (w *WebsocketServer) OnOpen(conn *gws.Conn) {
	log(LogInfo, "WebsocketServer OnOpen", zap.String("msg", "user join"))
	client := &WebsocketClient{conn: conn, done: make(chan struct{})}
	w.AddClient(client)
	go client.PingLoop(w.config.PingInterval)
	_ = conn.SetDeadline(time.Now().Add(w.config.PingWait * time.Second))
}

func (w *WebsocketServer) OnClose(conn *gws.Conn, err error) {

	log(LogInfo, "WebsocketServer OnClose", zap.String("msg", "user leave"))
	c := w.GetClient(conn)
	c.done <- struct{}{}
	w.RemoveClient(conn)
}

func (w *WebsocketServer) OnPing(socket *gws.Conn, payload []byte) {
	_ = socket.SetDeadline(time.Now().Add(w.config.PingWait * time.Second))
	_ = socket.WritePong(nil)
}

func (w *WebsocketServer) OnPong(socket *gws.Conn, payload []byte) {
	_ = socket.SetDeadline(time.Now().Add(w.config.PingWait * time.Second))
	log(LogInfo, "WebsocketServer PingLoop", zap.String("msg", "from user pong"))

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

	var msg WebSocketMessage
	err := json.Unmarshal(message.Data.Bytes(), &msg)
	if err != nil {
		log(LogError, "WebsocketServer OnMessage", zap.String("type", "Failed to unmarshal message"))
		return
	}

	// 执行操作
	handler, exists := w.handlers[msg.Type]
	if exists {
		log(LogInfo, "WebsocketServer OnMessage", zap.String("type", msg.Type))
		c := w.GetClient(conn)
		handler(c, &msg)
	} else {
		log(LogError, "WebsocketServer OnMessage", zap.String("msg", "Unknown message type"))
	}
}

// 定期发送 Ping 消息
func (c *WebsocketClient) PingLoop(PingInterval time.Duration) {
	ticker := time.NewTicker(PingInterval * time.Second) // 每 25 秒发送一次 Ping
	defer ticker.Stop()
	for {
		select {
		case <-c.done:
			log(LogInfo, "WebsocketServer PingLoop", zap.String("msg", "to close"))
			return
		case <-ticker.C:
			if err := c.conn.WritePing(nil); err != nil {
				log(LogError, "WebsocketServer PingLoop err ", zap.Error(err))
				return
			}
			log(LogInfo, "WebsocketServer PingLoop", zap.String("msg", "to user ping"))
		}
	}
}

// ToResponse 输出到浏览器
func (c *WebsocketClient) ToResponse(code *code.Code) {
	if code.HaveDetails() {
		details := strings.Join(code.Details(), ",")
		c.SendResponse(ResDetailsResult{
			Code:    code.Code(),
			Status:  code.Status(),
			Msg:     code.Lang.GetMessage(),
			Data:    code.Data(),
			Details: details,
		})
	} else {
		c.SendResponse(ResResult{
			Code:   code.Code(),
			Status: code.Status(),
			Msg:    code.Lang.GetMessage(),
			Data:   code.Data(),
		})
	}
}

func (c *WebsocketClient) SendResponse(content any) {
	responseBytes, _ := json.Marshal(content)
	c.conn.WriteMessage(gws.OpcodeText, responseBytes)
}
