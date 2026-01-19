package middleware

import (
	"strings"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"

	"github.com/gin-gonic/gin"
)

// ShareAuthToken 分享 Token 认证中间件
// 按优先级尝试获取 Token：Header -> Query -> PostForm
func ShareAuthToken(shareService service.ShareService) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := app.NewResponse(c)
		var token string

		token = c.GetHeader("Share-Token") // 支持自定义头

		// 2. 尝试从 URL 参数解析 (GET)
		if token == "" {
			token = c.Query("shareToken")
		}

		if token == "" {
			token = c.Query("share_token")
		}

		// 3. 尝试从表单参数解析 (POST)
		if token == "" {
			token = c.PostForm("shareToken")
		}
		if token == "" {
			token = c.PostForm("share_token")
		}

		if token == "" {
			response.ToResponse(code.ErrorInvalidAuthToken)
			c.Abort()
			return
		}

		// 确定当前请求想要访问的资源 ID 和类型
		rid := c.Query("id")
		if rid == "" {
			rid = c.PostForm("id")
		}

		// 简单的资源类型判定逻辑：根据路由路径区分
		rtp := "note"
		if strings.Contains(c.Request.URL.Path, "/file") {
			rtp = "file"
		}

		if rid == "" {
			response.ToResponse(code.ErrorInvalidParams)
			c.Abort()
			return
		}

		// 验证 Token 及其在数据库中的生效状态
		entity, err := shareService.VerifyShare(c.Request.Context(), token, rid, rtp)
		if err != nil {
			switch err {
			case domain.ErrShareCancelled:
				response.ToResponse(code.ErrorShareRevoked)
			case domain.ErrShareExpired:
				response.ToResponse(code.ErrorShareExpired)
			default:
				response.ToResponse(code.ErrorShareNotFound)
			}
			c.Abort()
			return
		}

		c.Set("share_entity", entity)
		c.Set("share_token", token)
		c.Next()
	}
}
