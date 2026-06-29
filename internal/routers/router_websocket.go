package routers

import (
	"context"
	"fmt"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/routers/websocket_router"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
)

func initWebSocketRoutes(wss *pkgapp.WebsocketServer, appContainer *app.App) {
	// Register Protobuf Hooks
	// 注册 Protobuf 编解码钩子
	websocket_router.RegisterProtobufHooks(wss)

	// Create WebSocket Handlers (injected App Container)
	// 创建 WebSocket Handlers（注入 App Container）
	noteWSHandler := websocket_router.NewNoteWSHandler(appContainer)
	folderWSHandler := websocket_router.NewFolderWSHandler(appContainer)
	fileWSHandler := websocket_router.NewFileWSHandler(appContainer)
	settingWSHandler := websocket_router.NewSettingWSHandler(appContainer)

	// Note
	wss.Use(websocket_router.NoteReceiveModify, noteWSHandler.NoteModify)
	wss.Use(websocket_router.NoteReceiveDelete, noteWSHandler.NoteDelete)
	wss.Use(websocket_router.NoteReceiveRename, noteWSHandler.NoteRename)
	wss.Use(websocket_router.NoteReceiveRePush, noteWSHandler.NoteRePush)
	wss.Use(websocket_router.NoteReceiveCheck, noteWSHandler.NoteModifyCheck)
	wss.Use(websocket_router.NoteReceiveSync, noteWSHandler.NoteSync)
	wss.Use(websocket_router.NoteSyncPageAck, noteWSHandler.NoteSyncPageAck)

	// Folder
	wss.Use(websocket_router.FolderReceiveSync, folderWSHandler.FolderSync)
	wss.Use(websocket_router.FolderReceiveModify, folderWSHandler.FolderModify)
	wss.Use(websocket_router.FolderReceiveDelete, folderWSHandler.FolderDelete)
	wss.Use(websocket_router.FolderReceiveRename, folderWSHandler.FolderRename)
	wss.Use(websocket_router.FolderSyncPageAck, folderWSHandler.FolderSyncPageAck)

	// Setting
	wss.Use(websocket_router.SettingReceiveModify, settingWSHandler.SettingModify)
	wss.Use(websocket_router.SettingReceiveDelete, settingWSHandler.SettingDelete)
	wss.Use(websocket_router.SettingReceiveCheck, settingWSHandler.SettingModifyCheck)
	wss.Use(websocket_router.SettingReceiveSync, settingWSHandler.SettingSync)
	wss.Use(websocket_router.SettingReceiveClear, settingWSHandler.SettingClear)
	wss.Use(websocket_router.SettingReceiveRePush, settingWSHandler.SettingRePush)
	wss.Use(websocket_router.SettingSyncPageAck, settingWSHandler.SettingSyncPageAck)

	// Attachment
	wss.Use(websocket_router.FileReceiveSync, fileWSHandler.FileSync)
	wss.Use(websocket_router.FileReceiveUploadCheck, fileWSHandler.FileUploadCheck)
	wss.Use(websocket_router.FileReceiveRename, fileWSHandler.FileRename)
	wss.Use(websocket_router.FileReceiveDelete, fileWSHandler.FileDelete)
	wss.Use(websocket_router.FileReceiveChunkDownload, fileWSHandler.FileChunkDownload)
	wss.Use(websocket_router.FileReceiveRePush, fileWSHandler.FileRePush)
	wss.Use(websocket_router.FileSyncPageAck, fileWSHandler.FileSyncPageAck)

	// Attachment chunk upload
	wss.UseBinary(websocket_router.VaultFileMsgType, fileWSHandler.FileUploadChunkBinary)

	// Inject Message Interceptor to handle unauthenticated checks, Vault restrictions, RBAC checks, and error rollbacks
	// 注入消息拦截器，处理未登录验证、Vault笔记库限制校验、RBAC权限检查以及写失败回滚机制
	wss.UseInterceptor(websocket_router.NewMessageInterceptor(appContainer))

	wss.UseUserVerify(noteWSHandler.UserInfo)

	// Inject Token Verification to decouple pkg/app from internal/service
	wss.UseTokenVerify(func(ctx context.Context, uid, tokenID int64, nonce string, reqClientType, reqClientName, reqClientVersion, reqUserAgent, reqIP string) (string, string, error) {
		dbToken, err := appContainer.TokenService.GetActiveToken(ctx, uid, tokenID)
		if err != nil || dbToken == nil {
			fmt.Printf("[WSDebug] Token not found or invalid in DB: uid=%d, tokenId=%d, err=%v\n", uid, tokenID, err)
			if err != nil {
				return "", "", err
			}
			return "", "", code.ErrorInvalidUserAuthToken
		}

		// 0. Verify Nonce (Generation Check)
		// 校验 Nonce（世代校验），如果数据库中有记录且不匹配，说明该令牌已被轮换或失效
		if dbToken.TokenString != "" && nonce != dbToken.TokenString {
			fmt.Printf("[WSDebug] Token rotated: req_nonce=%s, db_nonce=%s\n", nonce, dbToken.TokenString)
			return "", "", code.ErrorInvalidUserAuthToken.WithDetails("Token has been rotated")
		}

		// 1. Verify Scope Permissions (Protocol: ws)
		if !pkgapp.VerifyPermissions(dbToken.Scope, "ws", reqClientType, "") {
			fmt.Printf("[WSDebug] Permission denied: scope=%s, protocol=%s, client=%s\n", dbToken.Scope, "ws", reqClientType)
			return "", "", code.ErrorAuthTokenScopeRestricted.WithDetails("Permission denied: Handshake")
		}

		// 2. Verify Client Type (Only for login tokens where ClientType is used for restriction)
		// 仅对登录令牌执行严格客户端匹配，手动令牌通过 Scope 校验
		if dbToken.IssueType == 1 && dbToken.ClientType != "" && !pkgapp.MatchWildcard(dbToken.ClientType, reqClientType) {
			fmt.Printf("[WSDebug] ClientType mismatch: req=%s, db=%s\n", reqClientType, dbToken.ClientType)
			return "", "", code.ErrorAuthTokenClientRestricted.WithDetails("Client mismatch")
		}

		// 3. Verify User-Agent (Only if bound)
		if dbToken.UserAgent != "" && !pkgapp.MatchWildcard(dbToken.UserAgent, reqUserAgent) {
			fmt.Printf("[WSDebug] User-Agent mismatch: req=%s, db=%s\n", reqUserAgent, dbToken.UserAgent)
			return "", "", code.ErrorAuthTokenUARestricted
		}

		// 4. Verify IP (Only if bound)
		if dbToken.BoundIP != "" && !pkgapp.MatchWildcard(dbToken.BoundIP, reqIP) {
			fmt.Printf("[WSDebug] IP mismatch: req=%s, db=%s\n", reqIP, dbToken.BoundIP)
			return "", "", code.ErrorAuthTokenIPRestricted
		}

		_ = appContainer.TokenService.RecordAccessLog(ctx, &domain.AuthTokenLog{
			TokenID:       tokenID,
			UID:           uid,
			Protocol:      "ws",
			Client:        reqClientType,
			ClientName:    reqClientName,
			ClientVersion: reqClientVersion,
			IP:            reqIP,
			UA:            reqUserAgent,
			StatusCode:    101, // Switching Protocols
		})

		return dbToken.Scope, dbToken.Vaults, nil
	})
}
