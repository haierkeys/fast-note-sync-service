package routers

import (
	"context"
	"strings"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/routers/websocket_router"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
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
	wss.UseTokenVerify(func(ctx context.Context, uid, tokenID int64, reqClientType, reqUserAgent, reqIP string) error {
		dbToken, err := appContainer.TokenService.GetActiveToken(ctx, uid, tokenID)
		if err != nil || dbToken == nil {
			return code.ErrorInvalidUserAuthToken
		}

		// 1. Verify Scope Permissions (Protocol: ws)
		if !pkgapp.VerifyPermissions(dbToken.Scope, "ws", reqClientType, "") {
			return code.ErrorInvalidUserAuthToken
		}

		// 2. Verify Client Type
		if reqClientType != "" && !strings.EqualFold(reqClientType, dbToken.ClientType) {
			return code.ErrorInvalidUserAuthToken
		}

		// 3. Verify User-Agent
		if reqUserAgent != dbToken.UserAgent {
			return code.ErrorInvalidUserAuthToken
		}

		// 4. Verify IP
		if reqIP != dbToken.BoundIP {
			return code.ErrorInvalidUserAuthToken
		}

		return nil
	})
}
