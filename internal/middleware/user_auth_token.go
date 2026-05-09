package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"

	"github.com/gin-gonic/gin"
)

// UserAuthTokenWithConfig user Token authentication middleware (using injected secret key and token service)
// UserAuthTokenWithConfig 用户 Token 认证中间件（使用注入的密钥和 Token 服务）
func UserAuthTokenWithConfig(secretKey string, tokenService service.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := app.NewResponse(c)
		var token string

		// Prioritize getting from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}

		if token == "" {
			authHeader := c.GetHeader("Token")
			if authHeader != "" {
				token = authHeader
			}
			authHeader = c.GetHeader("token")
			if authHeader != "" {
				token = authHeader
			}
		}

		// If not in header, try getting from URL parameter
		if token == "" {
			token = c.Query("token")
		}

		if token == "" {
			response.ToResponse(code.ErrorNotUserAuthToken)
			c.Abort()
			return
		}

		user, err := app.ParseTokenWithKey(token, secretKey)
		if err != nil {
			response.ToResponse(code.ErrorInvalidUserAuthToken)
			c.Abort()
			return
		}

		// 2. Fetch and Validate Stateful Token from DB
		// 从数据库获取并验证状态化 Token
		ctx := c.Request.Context()
		dbToken, err := tokenService.GetActiveToken(ctx, user.UID, user.TokenID)
		if err != nil || dbToken == nil {
			if appErr, ok := err.(*code.Code); ok {
				response.ToResponse(appErr)
			} else {
				response.ToResponse(code.ErrorInvalidUserAuthToken)
			}
			c.Abort()
			return
		}

		// 3. Verify Client, IP and User-Agent Binding
		// 验证客户端类型、IP 和浏览器 User-Agent 的严格绑定
		reqClientType := c.GetHeader("x-client")
		if reqClientType != dbToken.ClientType {
			response.ToResponse(code.ErrorInvalidUserAuthToken.WithDetails("Client mismatch"))
			c.Abort()
			return
		}

		// 检查 User-Agent 防篡改/防盗用
		if reqUserAgent := c.GetHeader("User-Agent"); reqUserAgent != dbToken.UserAgent {
			response.ToResponse(code.ErrorInvalidUserAuthToken.WithDetails("User-Agent mismatch"))
			c.Abort()
			return
		}

		// 检查 IP 防盗用
		if reqIP := c.ClientIP(); reqIP != dbToken.BoundIP {
			response.ToResponse(code.ErrorInvalidUserAuthToken.WithDetails("IP mismatch"))
			c.Abort()
			return
		}

		// 4. Determine Function Dimension for RBAC
		// 确定 RBAC 的功能维度
		path := c.Request.URL.Path
		method := c.Request.Method
		var function string
		
		// Map path to resource
		var resource string
		if strings.HasPrefix(path, "/api/note") || strings.HasPrefix(path, "/api/folder") {
			resource = "note"
		} else if strings.HasPrefix(path, "/api/file") || strings.HasPrefix(path, "/api/storage") {
			resource = "file"
		} else if strings.HasPrefix(path, "/api/setting") || strings.HasPrefix(path, "/api/admin/config") {
			resource = "config"
		}
		
		if resource != "" {
			if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
				function = resource + "_r"
			} else {
				function = resource + "_w"
			}
		}

		// Protocol is rest for this middleware
		protocol := "rest"

		// 5. Verify Permissions
		if !app.VerifyPermissions(dbToken.Scope, protocol, reqClientType, function) {
			response.ToResponse(code.ErrorInvalidAuthToken.WithDetails("Permission denied"))
			c.Abort()
			return
		}

		// 6. Asynchronously record access log
		// 异步记录访问日志
		go func() {
			log := &domain.AuthTokenLog{
				TokenID:    dbToken.ID,
				UID:        dbToken.UID,
				Path:       path,
				Method:     method,
				IP:         c.ClientIP(),
				StatusCode: int64(c.Writer.Status()),
			}
			// Use background context for async operation
			_ = tokenService.RecordAccessLog(context.Background(), log)
		}()

		c.Set("user_token", user)
		c.Next()
	}
}

// UserAuthToken user Token authentication middleware (no secret key, always fails)
// UserAuthToken 用户 Token 认证中间件（无密钥，始终失败）
// Deprecated: Use UserAuthTokenWithConfig instead
// Deprecated: 推荐使用 UserAuthTokenWithConfig
func UserAuthToken() gin.HandlerFunc {
	// Without token service this cannot work properly in 3D RBAC
	return UserAuthTokenWithConfig("", nil)
}
