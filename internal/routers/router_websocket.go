package routers

import (
	"context"
	"strings"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/routers/websocket_router"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"fmt"
)

func initWebSocketRoutes(wss *pkgapp.WebsocketServer, appContainer *app.App) {
	// Create WebSocket Handlers (injected App Container)
	// 创建 WebSocket Handlers（注入 App Container）
	noteWSHandler := websocket_router.NewNoteWSHandler(appContainer)
	folderWSHandler := websocket_router.NewFolderWSHandler(appContainer)
	fileWSHandler := websocket_router.NewFileWSHandler(appContainer)
	settingWSHandler := websocket_router.NewSettingWSHandler(appContainer)

	// Note
	wss.Use(dto.NoteReceiveModify, noteWSHandler.NoteModify)
	wss.Use(dto.NoteReceiveDelete, noteWSHandler.NoteDelete)
	wss.Use(dto.NoteReceiveRename, noteWSHandler.NoteRename)
	wss.Use(dto.NoteReceiveRePush, noteWSHandler.NoteRePush)
	wss.Use(dto.NoteReceiveCheck, noteWSHandler.NoteModifyCheck)
	wss.Use(dto.NoteReceiveSync, noteWSHandler.NoteSync)

	// Folder
	wss.Use(dto.FolderReceiveSync, folderWSHandler.FolderSync)
	wss.Use(dto.FolderReceiveModify, folderWSHandler.FolderModify)
	wss.Use(dto.FolderReceiveDelete, folderWSHandler.FolderDelete)
	wss.Use(dto.FolderReceiveRename, folderWSHandler.FolderRename)

	// Setting
	wss.Use(dto.SettingReceiveModify, settingWSHandler.SettingModify)
	wss.Use(dto.SettingReceiveDelete, settingWSHandler.SettingDelete)
	wss.Use(dto.SettingReceiveCheck, settingWSHandler.SettingModifyCheck)
	wss.Use(dto.SettingReceiveSync, settingWSHandler.SettingSync)
	wss.Use(dto.SettingReceiveClear, settingWSHandler.SettingClear)

	// Attachment
	wss.Use(dto.FileReceiveSync, fileWSHandler.FileSync)
	wss.Use(dto.FileReceiveUploadCheck, fileWSHandler.FileUploadCheck)
	wss.Use(dto.FileReceiveRename, fileWSHandler.FileRename)
	wss.Use(dto.FileReceiveDelete, fileWSHandler.FileDelete)
	wss.Use(dto.FileReceiveChunkDownload, fileWSHandler.FileChunkDownload)
	wss.Use(dto.FileReceiveRePush, fileWSHandler.FileRePush)

	// Attachment chunk upload
	wss.UseBinary(dto.VaultFileMsgType, fileWSHandler.FileUploadChunkBinary)

	wss.UseUserVerify(noteWSHandler.UserInfo)

	// Inject Token Verification to decouple pkg/app from internal/service
	wss.UseTokenVerify(func(ctx context.Context, uid, tokenID int64, reqClientType, reqClientName, reqClientVersion, reqUserAgent, reqIP string) error {
		dbToken, err := appContainer.TokenService.GetActiveToken(ctx, uid, tokenID)
		if err != nil || dbToken == nil {
			fmt.Printf("[WSDebug] Token not found or invalid in DB: uid=%d, tokenId=%d, err=%v\n", uid, tokenID, err)
			if err != nil {
				return err
			}
			return code.ErrorInvalidUserAuthToken
		}

		// 1. Verify Scope Permissions (Protocol: ws)
		if !pkgapp.VerifyPermissions(dbToken.Scope, "ws", reqClientType, "") {
			fmt.Printf("[WSDebug] Permission denied: scope=%s, protocol=%s, client=%s\n", dbToken.Scope, "ws", reqClientType)
			return code.ErrorInvalidAuthToken.WithDetails("Permission denied")
		}

		// 2. Verify Client Type
		if reqClientType != "" && !strings.EqualFold(reqClientType, dbToken.ClientType) {
			fmt.Printf("[WSDebug] ClientType mismatch: req=%s, db=%s\n", reqClientType, dbToken.ClientType)
			return code.ErrorAuthTokenClientRestricted
		}

		// 3. Verify User-Agent (Only if bound)
		if dbToken.UserAgent != "" && !pkgapp.MatchWildcard(dbToken.UserAgent, reqUserAgent) {
			fmt.Printf("[WSDebug] User-Agent mismatch: req=%s, db=%s\n", reqUserAgent, dbToken.UserAgent)
			return code.ErrorAuthTokenUARestricted
		}

		// 4. Verify IP (Only if bound)
		if dbToken.BoundIP != "" && !pkgapp.MatchWildcard(dbToken.BoundIP, reqIP) {
			fmt.Printf("[WSDebug] IP mismatch: req=%s, db=%s\n", reqIP, dbToken.BoundIP)
			return code.ErrorAuthTokenIPRestricted
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
		
		return nil
	})
}
