/**
  @author: haierkeys
  @since: 2026/06/20
  @desc: WebGUI access control middleware with multi-factor verification // WebGUI 访问控制中间件（多因子联合校验）
**/

package middleware

import (
	"github.com/gin-gonic/gin"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
)

// RequireWebGUI is a Gin middleware that enforces WebGUI client access control
// with multi-factor verification for authenticated routes.
// RequireWebGUI 是一个 Gin 中间件，对已认证路由执行多因子 WebGUI 访问控制校验：
//   - 未认证路由（如登录/注册）：仅校验请求头 x-client/client 是否为 "webgui"
//   - 已认证路由（有 Token 上下文时）：额外校验服务端 Token 属性：
//     IssueType 必须为 1（Login Token），ClientType 必须匹配 "webgui"
//     从根本上防止手动 API 令牌伪造请求头绕过 WebGUI 专属管理接口
func RequireWebGUI() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Check disabled for all modes except ReleaseMode for Swagger/tool testing.
		if gin.Mode() != gin.ReleaseMode {
			c.Next()
			return
		}

		response := pkgapp.NewResponse(c)

		// Layer 1: Header/query parameter check (applies to all routes)
		// 第一层：请求头 / 参数校验（所有路由均须通过）
		if !pkgapp.IsWebGUI(c) {
			response.ToResponse(code.ErrorAuthTokenClientRestricted.WithDetails(
				"This action is restricted to webgui client only"))
			c.Abort()
			return
		}

		// Layer 2: Token attribute joint check (only applies when token context is present, i.e. authenticated routes)
		// 第二层：Token 属性联合校验（仅在 Token 上下文存在时生效，即已认证路由）
		// The data comes from the server-side database and cannot be forged by the client.
		// 这些数据来自服务端数据库，客户端无法伪造
		issueTypeVal, issueTypeExists := c.Get("token_issue_type")
		clientTypeVal, _ := c.Get("token_client_type")

		if issueTypeExists {
			issueType, _ := issueTypeVal.(int)
			clientType, _ := clientTypeVal.(string)

			// IssueType must be 1 (Login Token); manual API tokens (IssueType=2) are not allowed
			// IssueType 必须为 1（登录令牌），手动创建的 API 令牌（IssueType=2）不允许访问
			if issueType != 1 {
				response.ToResponse(code.ErrorAuthTokenClientRestricted.WithDetails(
					"Manual API tokens are not allowed to access WebGUI-only endpoints"))
				c.Abort()
				return
			}

			// Token's bound ClientType must match "webgui" (server-side binding, not client-controllable)
			// Token 绑定的 ClientType 必须匹配 "webgui"（服务端签发时绑定，客户端不可篡改）
			if !pkgapp.MatchWildcard("webgui", clientType) {
				response.ToResponse(code.ErrorAuthTokenClientRestricted.WithDetails(
					"Token client type does not match webgui"))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
