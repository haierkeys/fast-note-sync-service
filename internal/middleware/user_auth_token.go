package middleware

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"

	"github.com/gin-gonic/gin"
)

func UserAuthTokenWithConfig(secretKey string, tokenService service.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := app.NewResponse(c)
		user, scope, vaults, appErr := AuthenticateUserToken(c, secretKey, tokenService)
		if appErr != nil {
			response.ToResponse(appErr)
			c.Abort()
			return
		}
		c.Set("user_token", user)
		c.Set("scope", scope)
		c.Set("vaults", vaults)
		c.Next()
	}
}

func ExtractUserAuthToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	if authHeader := c.GetHeader("Token"); authHeader != "" {
		return authHeader
	}
	if authHeader := c.GetHeader("token"); authHeader != "" {
		return authHeader
	}

	return c.Query("token")
}

func AuthenticateUserToken(c *gin.Context, secretKey string, tokenService service.TokenService) (*app.UserEntity, string, string, *code.Code) {
	token := ExtractUserAuthToken(c)
	if token == "" {
		return nil, "", "", code.ErrorNotUserAuthToken
	}

	user, err := app.ParseTokenWithKey(token, secretKey)
	if err != nil {
		if appErr, ok := err.(*code.Code); ok {
			return nil, "", "", appErr
		}
		return nil, "", "", code.ErrorInvalidUserAuthToken
	}

	ctx := c.Request.Context()
	dbToken, err := tokenService.GetActiveToken(ctx, user.UID, user.TokenID)
	if err != nil || dbToken == nil {
		if appErr, ok := err.(*code.Code); ok {
			return nil, "", "", appErr
		}
		return nil, "", "", code.ErrorInvalidUserAuthToken
	}

	if dbToken.TokenString != "" && user.Nonce != dbToken.TokenString {
		return nil, "", "", code.ErrorInvalidUserAuthToken.WithDetails("Token has been rotated")
	}

	reqClientType := c.GetHeader("x-client")
	if reqClientType == "" {
		reqClientType = c.Query("client")
	}
	if reqClientType == "" && dbToken.IssueType == 1 && isHeaderlessLoginTokenResourceRead(c) {
		reqClientType = dbToken.ClientType
	}

	if dbToken.IssueType == 1 && !app.MatchWildcard(dbToken.ClientType, reqClientType) {
		return nil, "", "", code.ErrorAuthTokenClientRestricted.WithDetails("Client mismatch")
	}

	if dbToken.UserAgent != "" {
		if reqUserAgent := c.GetHeader("User-Agent"); !app.MatchWildcard(dbToken.UserAgent, reqUserAgent) {
			return nil, "", "", code.ErrorAuthTokenUARestricted.WithDetails("User-Agent mismatch")
		}
	}

	if dbToken.BoundIP != "" {
		if reqIP := c.ClientIP(); !app.MatchWildcard(dbToken.BoundIP, reqIP) {
			return nil, "", "", code.ErrorAuthTokenIPRestricted.WithDetails("IP mismatch")
		}
	}

	path := c.Request.URL.Path
	method := c.Request.Method
	var function string

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

	protocol := "rest"
	if strings.HasPrefix(path, "/api/mcp") {
		protocol = "mcp"
	}

	if path != "/api/health" && !app.VerifyPermissions(dbToken.Scope, protocol, reqClientType, function) {
		resPath := c.Query("path")
		if resPath == "" {
			resPath = c.Query("name")
		}
		if resPath == "" {
			resPath = c.Query("file")
		}
		if resPath == "" {
			resPath = path
		}
		return nil, "", "", code.ErrorAuthTokenScopeRestricted.WithDetails("Permission denied: " + resPath)
	}

	if dbToken.Vaults != "" {
		targetVault := app.RequestParam(c, "vault")
		if targetVault != "" && !util.VerifyVaultAccess(dbToken.Vaults, targetVault) {
			return nil, "", "", code.ErrorAuthTokenScopeRestricted.WithDetails("Vault access restricted: " + targetVault)
		}
	}

	go func() {
		clientName := c.GetHeader("x-client-name")
		if clientName != "" {
			if decoded, err := url.QueryUnescape(clientName); err == nil {
				clientName = decoded
			}
		}

		log := &domain.AuthTokenLog{
			TokenID:       dbToken.ID,
			UID:           dbToken.UID,
			Protocol:      protocol,
			Client:        reqClientType,
			ClientName:    clientName,
			ClientVersion: c.GetHeader("x-client-version"),
			IP:            c.ClientIP(),
			UA:            c.GetHeader("User-Agent"),
			StatusCode:    int64(c.Writer.Status()),
		}
		_ = tokenService.RecordAccessLog(context.Background(), log)
	}()

	return user, dbToken.Scope, dbToken.Vaults, nil
}

func isHeaderlessLoginTokenResourceRead(c *gin.Context) bool {
	return c.Request.URL.Path == "/api/file" &&
		(c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead)
}

// UserAuthToken user Token authentication middleware (no secret key, always fails)
// UserAuthToken 用户 Token 认证中间件（无密钥，始终失败）
// Deprecated: Use UserAuthTokenWithConfig instead
// Deprecated: 推荐使用 UserAuthTokenWithConfig
func UserAuthToken() gin.HandlerFunc {
	// Without token service this cannot work properly in 3D RBAC
	return UserAuthTokenWithConfig("", nil)
}
