// Package api_router provides HTTP API router handlers
// Package api_router 提供 HTTP API 路由处理器
package api_router

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	apperrors "github.com/haierkeys/fast-note-sync-service/pkg/errors"
)

// Handler basic Handler struct, encapsulates App Container
// All API Handlers should embed this struct to gain dependency injection capability
// Handler 基础 Handler 结构体，封装 App Container
// 所有 API Handler 都应该嵌入此结构体以获得依赖注入能力
type Handler struct {
	App *app.App
	WSS *pkgapp.WebsocketServer
}

// NewHandler creates a new base Handler instance
// NewHandler 创建新的基础 Handler 实例
func NewHandler(a *app.App) *Handler {
	return &Handler{App: a}
}

// respondNoteError renders a note-service error.
//
// Application code errors that carry a payload (notably the conditional-write conflicts 452/453,
// which return the note's current revision and content so the caller can reconcile) go through
// the standard response envelope, which preserves `data` and the HTTP status attached to the
// code. Everything else keeps using the generic error renderer.
// respondNoteError 渲染笔记服务返回的错误。携带数据的应用错误码（尤其是条件写入冲突 452/453，
// 会返回笔记当前的版本号与内容以便调用方比对）走标准响应封装，保留 data 与错误码对应的 HTTP
// 状态；其他错误仍使用通用错误渲染。
func (h *Handler) respondNoteError(c *gin.Context, err error) {
	var codeErr *code.Code
	if errors.As(err, &codeErr) && (codeErr.HaveData() || codeErr.StatusCode() != http.StatusOK) {
		pkgapp.NewResponse(c).ToResponse(codeErr)
		return
	}
	apperrors.ErrorResponse(c, err)
}

// NewHandlerWithWSS creates Handler instance with WebSocket service
// NewHandlerWithWSS 创建带 WebSocket 服务的 Handler 实例
func NewHandlerWithWSS(a *app.App, wss *pkgapp.WebsocketServer) *Handler {
	return &Handler{App: a, WSS: wss}
}

// getClientInfo extracts client type, name and version from request headers
// getClientInfo 从请求头中提取客户端类型、名称和版本
func (h *Handler) getClientInfo(c *gin.Context) (string, string, string) {
	clientType := c.GetHeader("X-Client")
	clientName := c.GetHeader("X-Client-Name")
	clientVersion := c.GetHeader("X-Client-Version")

	// Decode clientName if it's URL-encoded
	// 如果 clientName 是 URL 编码的，则进行解码
	if clientName != "" {
		if decoded, err := url.QueryUnescape(clientName); err == nil {
			clientName = decoded
		}
	}

	return clientType, clientName, clientVersion
}
