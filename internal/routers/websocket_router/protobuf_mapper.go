package websocket_router

import (
	"fmt"

	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	v1 "github.com/haierkeys/fast-note-sync-service/internal/proto/v1"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/json"
	"google.golang.org/protobuf/proto"
)

// Hook registrations
// 注册 Protobuf 相关的编解码钩子函数
func RegisterProtobufHooks(wss *pkgapp.WebsocketServer) {
	wss.EnvelopeDecoder = DeReceivePacket
	wss.ProtobufDecoder = DeReceiveProtobufToDTO
	wss.ProtobufEncoder = EnSendDTOToProtobuf
}

// DeReceivePacket unpacks the outer WSMessage packet
// DeReceivePacket 解包最外层的 WSMessage 网络数据包
func DeReceivePacket(data []byte) (WebSocketReceiveAction, []byte, error) {
	var env v1.WSMessage
	if err := proto.Unmarshal(data, &env); err != nil {
		return "", nil, err
	}
	return WebSocketReceiveAction(env.Type), env.Data, nil
}

// DeReceiveProtobufToDTO maps protobuf message data to target DTO object
// DeReceiveProtobufToDTO 将接收到的 Protobuf 消息数据映射到目标 DTO 对象
func DeReceiveProtobufToDTO(action WebSocketReceiveAction, data []byte, obj any) (bool, error) {
	switch action {
	// "ClientInfo"
	case ClientReceiveInfo:
		var pbMsg v1.ClientInfoMessage
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*pkgapp.ClientInfoMessage); ok {
			dest.Name = pbMsg.Name
			dest.Version = pbMsg.Version
			dest.Type = pbMsg.Type
			dest.IsDesktop = pbMsg.IsDesktop
			dest.IsMobile = pbMsg.IsMobile
			dest.IsPhone = pbMsg.IsPhone
			dest.IsTablet = pbMsg.IsTablet
			dest.IsMacOS = pbMsg.IsMacOS
			dest.IsWin = pbMsg.IsWin
			dest.IsLinux = pbMsg.IsLinux
			dest.OfflineSyncStrategy = pbMsg.OfflineSyncStrategy
			dest.Protobuf = pbMsg.Protobuf
			return true, nil
		}

	// 1. Note Messages
	// "NoteSync"
	case NoteReceiveSync:
		var pbMsg v1.NoteSyncRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.NoteSyncRequest); ok {
			dest.Context = pbMsg.Context
			dest.Vault = pbMsg.Vault
			dest.LastTime = pbMsg.LastTime
			dest.BatchIndex = int(pbMsg.BatchIndex)
			dest.TotalBatches = int(pbMsg.TotalBatches)
			dest.Notes = make([]dto.NoteSyncCheckRequest, len(pbMsg.Notes))
			for i, v := range pbMsg.Notes {
				dest.Notes[i] = dto.NoteSyncCheckRequest{
					Path:        v.Path,
					PathHash:    v.PathHash,
					ContentHash: v.ContentHash,
					Mtime:       v.Mtime,
					Ctime:       v.Ctime,
				}
			}
			dest.DelNotes = make([]dto.NoteSyncDelNote, len(pbMsg.DelNotes))
			for i, v := range pbMsg.DelNotes {
				dest.DelNotes[i] = dto.NoteSyncDelNote{
					Path:     v.Path,
					PathHash: v.PathHash,
				}
			}
			dest.MissingNotes = make([]dto.NoteSyncDelNote, len(pbMsg.MissingNotes))
			for i, v := range pbMsg.MissingNotes {
				dest.MissingNotes[i] = dto.NoteSyncDelNote{
					Path:     v.Path,
					PathHash: v.PathHash,
				}
			}
			return true, nil
		}
	// "NoteModify"
	case NoteReceiveModify:
		var pbMsg v1.NoteModifyOrCreateRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.NoteModifyOrCreateRequest); ok {
			dest.Vault = pbMsg.Vault
			dest.Path = pbMsg.Path
			dest.PathHash = pbMsg.PathHash
			dest.BaseHash = pbMsg.BaseHash
			dest.BaseHashMissing = pbMsg.BaseHashMissing
			dest.Content = pbMsg.Content
			dest.ContentHash = pbMsg.ContentHash
			dest.Ctime = pbMsg.Ctime
			dest.Mtime = pbMsg.Mtime
			dest.CreateOnly = pbMsg.CreateOnly
			dest.Context = pbMsg.Context
			return true, nil
		}
	// "NoteCheck"
	case NoteReceiveCheck:
		var pbMsg v1.NoteUpdateCheckRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.NoteUpdateCheckRequest); ok {
			dest.Vault = pbMsg.Vault
			dest.Path = pbMsg.Path
			dest.PathHash = pbMsg.PathHash
			dest.ContentHash = pbMsg.ContentHash
			dest.Ctime = pbMsg.Ctime
			dest.Mtime = pbMsg.Mtime
			return true, nil
		}
	// "NoteDelete"
	case NoteReceiveDelete:
		var pbMsg v1.NoteDeleteRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.NoteDeleteRequest); ok {
			dest.Vault = pbMsg.Vault
			dest.Path = pbMsg.Path
			dest.PathHash = pbMsg.PathHash
			dest.Context = pbMsg.Context
			return true, nil
		}
	// "NoteRename"
	case NoteReceiveRename:
		var pbMsg v1.NoteRenameRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.NoteRenameRequest); ok {
			dest.Vault = pbMsg.Vault
			dest.Path = pbMsg.Path
			dest.PathHash = pbMsg.PathHash
			dest.OldPath = pbMsg.OldPath
			dest.OldPathHash = pbMsg.OldPathHash
			dest.Context = pbMsg.Context
			return true, nil
		}
	// "NoteRePush"
	case NoteReceiveRePush:
		var pbMsg v1.NoteGetRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.NoteGetRequest); ok {
			dest.Vault = pbMsg.Vault
			dest.Path = pbMsg.Path
			dest.PathHash = pbMsg.PathHash
			dest.IsRecycle = pbMsg.IsRecycle
			return true, nil
		}

	// 2. File Messages
	// "FileSync"
	case FileReceiveSync:
		var pbMsg v1.FileSyncRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.FileSyncRequest); ok {
			dest.Context = pbMsg.Context
			dest.Vault = pbMsg.Vault
			dest.LastTime = pbMsg.LastTime
			dest.BatchIndex = int(pbMsg.BatchIndex)
			dest.TotalBatches = int(pbMsg.TotalBatches)
			dest.Files = make([]dto.FileSyncCheckRequest, len(pbMsg.Files))
			for i, v := range pbMsg.Files {
				dest.Files[i] = dto.FileSyncCheckRequest{
					Path:        v.Path,
					PathHash:    v.PathHash,
					ContentHash: v.ContentHash,
					Size:        v.Size,
					Mtime:       v.Mtime,
					Ctime:       v.Ctime,
				}
			}
			dest.DelFiles = make([]dto.FileSyncDelFile, len(pbMsg.DelFiles))
			for i, v := range pbMsg.DelFiles {
				dest.DelFiles[i] = dto.FileSyncDelFile{
					Path:     v.Path,
					PathHash: v.PathHash,
				}
			}
			dest.MissingFiles = make([]dto.FileSyncDelFile, len(pbMsg.MissingFiles))
			for i, v := range pbMsg.MissingFiles {
				dest.MissingFiles[i] = dto.FileSyncDelFile{
					Path:     v.Path,
					PathHash: v.PathHash,
				}
			}
			return true, nil
		}
	// "FileUploadCheck"
	case FileReceiveUploadCheck:
		var pbMsg v1.FileUploadCheckRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.FileUpdateCheckRequest); ok {
			dest.Vault = pbMsg.Vault
			dest.Path = pbMsg.Path
			dest.PathHash = pbMsg.PathHash
			dest.ContentHash = pbMsg.ContentHash
			dest.Size = pbMsg.Size
			dest.Ctime = pbMsg.Ctime
			dest.Mtime = pbMsg.Mtime
			dest.Context = pbMsg.Context
			return true, nil
		}
	// "FileDelete"
	case FileReceiveDelete:
		var pbMsg v1.FileDeleteRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.FileDeleteRequest); ok {
			dest.Vault = pbMsg.Vault
			dest.Path = pbMsg.Path
			dest.PathHash = pbMsg.PathHash
			dest.Context = pbMsg.Context
			return true, nil
		}
	// "FileRename"
	case FileReceiveRename:
		var pbMsg v1.FileRenameRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.FileRenameRequest); ok {
			dest.Vault = pbMsg.Vault
			dest.Path = pbMsg.Path
			dest.PathHash = pbMsg.PathHash
			dest.OldPath = pbMsg.OldPath
			dest.OldPathHash = pbMsg.OldPathHash
			dest.Context = pbMsg.Context
			return true, nil
		}
	// "FileChunkDownload"
	case FileReceiveChunkDownload:
		var pbMsg v1.FileChunkDownloadRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.FileGetRequest); ok {
			dest.Vault = pbMsg.Vault
			dest.Path = pbMsg.Path
			dest.PathHash = pbMsg.PathHash
			dest.Context = pbMsg.Context
			return true, nil
		}
	// "FileRePush"
	case FileReceiveRePush:
		var pbMsg v1.FileGetRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.FileGetRequest); ok {
			dest.Vault = pbMsg.Vault
			dest.Path = pbMsg.Path
			dest.PathHash = pbMsg.PathHash
			dest.Context = pbMsg.Context
			return true, nil
		}

	// 3. Setting Messages
	// "SettingSync"
	case SettingReceiveSync:
		var pbMsg v1.SettingSyncRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.SettingSyncRequest); ok {
			dest.Context = pbMsg.Context
			dest.Vault = pbMsg.Vault
			dest.LastTime = pbMsg.LastTime
			dest.BatchIndex = int(pbMsg.BatchIndex)
			dest.TotalBatches = int(pbMsg.TotalBatches)
			dest.Settings = make([]dto.SettingSyncCheckRequest, len(pbMsg.Settings))
			for i, v := range pbMsg.Settings {
				dest.Settings[i] = dto.SettingSyncCheckRequest{
					Path:        v.Path,
					PathHash:    v.PathHash,
					ContentHash: v.ContentHash,
					Mtime:       v.Mtime,
				}
			}
			dest.DelSettings = make([]dto.SettingSyncDelSetting, len(pbMsg.DelSettings))
			for i, v := range pbMsg.DelSettings {
				dest.DelSettings[i] = dto.SettingSyncDelSetting{
					Path:     v.Path,
					PathHash: v.PathHash,
				}
			}
			dest.MissingSettings = make([]dto.SettingSyncDelSetting, len(pbMsg.MissingSettings))
			for i, v := range pbMsg.MissingSettings {
				dest.MissingSettings[i] = dto.SettingSyncDelSetting{
					Path:     v.Path,
					PathHash: v.PathHash,
				}
			}
			return true, nil
		}
	// "SettingModify"
	case SettingReceiveModify:
		var pbMsg v1.SettingModifyOrCreateRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.SettingModifyOrCreateRequest); ok {
			dest.Vault = pbMsg.Vault
			dest.Path = pbMsg.Path
			dest.PathHash = pbMsg.PathHash
			dest.Content = pbMsg.Content
			dest.ContentHash = pbMsg.ContentHash
			dest.Ctime = pbMsg.Ctime
			dest.Mtime = pbMsg.Mtime
			dest.Context = pbMsg.Context
			return true, nil
		}
	// "SettingCheck"
	case SettingReceiveCheck:
		var pbMsg v1.SettingUpdateCheckRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.SettingUpdateCheckRequest); ok {
			dest.Vault = pbMsg.Vault
			dest.Path = pbMsg.Path
			dest.PathHash = pbMsg.PathHash
			dest.ContentHash = pbMsg.ContentHash
			dest.Ctime = pbMsg.Ctime
			dest.Mtime = pbMsg.Mtime
			return true, nil
		}
	// "SettingDelete"
	case SettingReceiveDelete:
		var pbMsg v1.SettingDeleteRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.SettingDeleteRequest); ok {
			dest.Vault = pbMsg.Vault
			dest.Path = pbMsg.Path
			dest.PathHash = pbMsg.PathHash
			dest.Context = pbMsg.Context
			return true, nil
		}
	// "SettingClear"
	case SettingReceiveClear:
		var pbMsg v1.SettingClearRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.SettingClearRequest); ok {
			dest.Vault = pbMsg.Vault
			return true, nil
		}
	// "SettingRePush"
	case SettingReceiveRePush:
		var pbMsg v1.SettingGetRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.SettingGetRequest); ok {
			dest.Vault = pbMsg.Vault
			dest.Path = pbMsg.Path
			dest.PathHash = pbMsg.PathHash
			return true, nil
		}
	// "FolderSync"
	case FolderReceiveSync:
		var pbMsg v1.FolderSyncRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.FolderSyncRequest); ok {
			dest.Context = pbMsg.Context
			dest.Vault = pbMsg.Vault
			dest.LastTime = pbMsg.LastTime
			dest.BatchIndex = int(pbMsg.BatchIndex)
			dest.TotalBatches = int(pbMsg.TotalBatches)
			dest.Folders = make([]dto.FolderSyncCheckRequest, len(pbMsg.Folders))
			for i, f := range pbMsg.Folders {
				dest.Folders[i] = dto.FolderSyncCheckRequest{
					Path:     f.Path,
					PathHash: f.PathHash,
					Mtime:    f.Mtime,
				}
			}
			dest.DelFolders = make([]dto.FolderSyncDelFolder, len(pbMsg.DelFolders))
			for i, f := range pbMsg.DelFolders {
				dest.DelFolders[i] = dto.FolderSyncDelFolder{
					Path:     f.Path,
					PathHash: f.PathHash,
				}
			}
			dest.MissingFolders = make([]dto.FolderSyncDelFolder, len(pbMsg.MissingFolders))
			for i, f := range pbMsg.MissingFolders {
				dest.MissingFolders[i] = dto.FolderSyncDelFolder{
					Path:     f.Path,
					PathHash: f.PathHash,
				}
			}
			return true, nil
		}
	// "FolderModify"
	case FolderReceiveModify:
		var pbMsg v1.FolderCreateRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.FolderCreateRequest); ok {
			dest.Vault = pbMsg.Vault
			dest.Path = pbMsg.Path
			dest.PathHash = pbMsg.PathHash
			dest.Context = pbMsg.Context
			return true, nil
		}
	// "FolderDelete"
	case FolderReceiveDelete:
		var pbMsg v1.FolderDeleteRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.FolderDeleteRequest); ok {
			dest.Vault = pbMsg.Vault
			dest.Path = pbMsg.Path
			dest.PathHash = pbMsg.PathHash
			dest.Context = pbMsg.Context
			return true, nil
		}
	// "FolderRename"
	case FolderReceiveRename:
		var pbMsg v1.FolderRenameRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.FolderRenameRequest); ok {
			dest.Vault = pbMsg.Vault
			dest.Path = pbMsg.Path
			dest.PathHash = pbMsg.PathHash
			dest.OldPath = pbMsg.OldPath
			dest.OldPathHash = pbMsg.OldPathHash
			dest.Context = pbMsg.Context
			return true, nil
		}
	case NoteSyncPageAck:
		var pbMsg v1.NoteSyncPageAckRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.SyncPageAckRequest); ok {
			dest.Context = pbMsg.Context
			dest.Vault = pbMsg.Vault
			dest.PageIndex = int(pbMsg.PageIndex)
			return true, nil
		}
	case FileSyncPageAck:
		var pbMsg v1.FileSyncPageAckRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.SyncPageAckRequest); ok {
			dest.Context = pbMsg.Context
			dest.Vault = pbMsg.Vault
			dest.PageIndex = int(pbMsg.PageIndex)
			return true, nil
		}
	case SettingSyncPageAck:
		var pbMsg v1.SettingSyncPageAckRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.SyncPageAckRequest); ok {
			dest.Context = pbMsg.Context
			dest.Vault = pbMsg.Vault
			dest.PageIndex = int(pbMsg.PageIndex)
			return true, nil
		}
	case FolderSyncPageAck:
		var pbMsg v1.FolderSyncPageAckRequest
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return false, err
		}
		if dest, ok := obj.(*dto.SyncPageAckRequest); ok {
			dest.Context = pbMsg.Context
			dest.Vault = pbMsg.Vault
			dest.PageIndex = int(pbMsg.PageIndex)
			return true, nil
		}
	}
	return false, fmt.Errorf("unknown action: %s", action)
}

// EnSendDTOToProtobuf serializes the Res struct to Protobuf WSResponse and wraps it in WSMessage
// EnSendDTOToProtobuf 将要发送的 DTO 响应体序列化为 Protobuf 的 WSResponse 格式，并封装进 WSMessage 外壳中
func EnSendDTOToProtobuf(action WebSocketSendAction, res *pkgapp.Res) ([]byte, error) {
	var innerData []byte
	var err error

	if res.Data != nil {
		innerData, err = enSendDataPayload(action, res.Data)
		if err != nil {
			return nil, err
		}
	}

	wsResp := &v1.WSResponse{
		Code:    int32(res.Code),
		Status:  res.Status,
		Message: formatString(res.Message),
		Data:    innerData,
		Details: formatString(res.Details),
		Vault:   formatString(res.Vault),
		Context: formatString(res.Context),
	}

	wsRespBytes, err := proto.Marshal(wsResp)
	if err != nil {
		return nil, err
	}

	envelope := &v1.WSMessage{
		Type: string(action),
		Data: wsRespBytes,
	}

	envelopeBytes, err := proto.Marshal(envelope)
	if err != nil {
		return nil, err
	}

	// Prepend "pb" prefix to the serialized protobuf packet
	// 为序列化后的 Protobuf 报文添加 "pb" 前缀
	result := make([]byte, 2+len(envelopeBytes))
	result[0] = 'p'
	result[1] = 'b'
	copy(result[2:], envelopeBytes)
	return result, nil
}

// enSendDataPayload serializes data payload by action type
// enSendDataPayload 根据动作类型序列化要发送的数据荷载
func enSendDataPayload(action WebSocketSendAction, data any) ([]byte, error) {
	switch action {
	case NoteSyncPage:
		if src, ok := data.(dto.SyncPageMessage); ok {
			pbMsg := &v1.NoteSyncPageMessage{
				PageIndex:  int32(src.PageIndex),
				PageSize:   int32(src.PageSize),
				TotalCount: int32(src.TotalCount),
				IsLast:     src.IsLast,
			}
			return proto.Marshal(pbMsg)
		}
	case FileSyncPage:
		if src, ok := data.(dto.SyncPageMessage); ok {
			pbMsg := &v1.FileSyncPageMessage{
				PageIndex:  int32(src.PageIndex),
				PageSize:   int32(src.PageSize),
				TotalCount: int32(src.TotalCount),
				IsLast:     src.IsLast,
			}
			return proto.Marshal(pbMsg)
		}
	case SettingSyncPage:
		if src, ok := data.(dto.SyncPageMessage); ok {
			pbMsg := &v1.SettingSyncPageMessage{
				PageIndex:  int32(src.PageIndex),
				PageSize:   int32(src.PageSize),
				TotalCount: int32(src.TotalCount),
				IsLast:     src.IsLast,
			}
			return proto.Marshal(pbMsg)
		}
	case FolderSyncPage:
		if src, ok := data.(dto.SyncPageMessage); ok {
			pbMsg := &v1.FolderSyncPageMessage{
				PageIndex:  int32(src.PageIndex),
				PageSize:   int32(src.PageSize),
				TotalCount: int32(src.TotalCount),
				IsLast:     src.IsLast,
			}
			return proto.Marshal(pbMsg)
		}
	// "ClientInfo"
	case ClientInfo:
		if src, ok := data.(pkgapp.CheckVersionInfo); ok {
			pbMsg := &v1.CheckVersionInfo{
				GithubAvailable:                  src.GithubAvailable,
				VersionIsNew:                     src.VersionIsNew,
				VersionNewName:                   src.VersionNewName,
				VersionNewLink:                   src.VersionNewLink,
				VersionNewChangelog:              src.VersionNewChangelog,
				VersionNewChangelogContent:       src.VersionNewChangelogContent,
				PluginVersionIsNew:               src.PluginVersionIsNew,
				PluginVersionNewName:             src.PluginVersionNewName,
				PluginVersionNewLink:             src.PluginVersionNewLink,
				PluginVersionNewChangelog:        src.PluginVersionNewChangelog,
				PluginVersionNewChangelogContent: src.PluginVersionNewChangelogContent,
				SyncUpChunkNum:                   int32(src.SyncUpChunkNum),
				SyncDownChunkNum:                 int32(src.SyncDownChunkNum),
			}
			pbMsg.VersionHistory = make([]*v1.HistoricalVersion, len(src.VersionHistory))
			for i, v := range src.VersionHistory {
				pbMsg.VersionHistory[i] = &v1.HistoricalVersion{
					Version:          v.Version,
					ChangelogContent: v.ChangelogContent,
				}
			}
			pbMsg.PluginVersionHistory = make([]*v1.HistoricalVersion, len(src.PluginVersionHistory))
			for i, v := range src.PluginVersionHistory {
				pbMsg.PluginVersionHistory[i] = &v1.HistoricalVersion{
					Version:          v.Version,
					ChangelogContent: v.ChangelogContent,
				}
			}
			return proto.Marshal(pbMsg)
		}

	// 1. Note responses
	// "NoteSyncModify"
	case NoteSyncModify:
		if src, ok := data.(dto.NoteSyncModifyMessage); ok {
			pbMsg := &v1.NoteSyncModifyMessage{
				Path:        src.Path,
				PathHash:    src.PathHash,
				Content:     src.Content,
				ContentHash: src.ContentHash,
				Ctime:       src.Ctime,
				Mtime:       src.Mtime,
				LastTime:    src.UpdatedTimestamp,
			}
			return proto.Marshal(pbMsg)
		}
	// "NoteSyncDelete"
	case NoteSyncDelete:
		if src, ok := data.(dto.NoteSyncDeleteMessage); ok {
			pbMsg := &v1.NoteSyncDeleteMessage{
				Path:     src.Path,
				PathHash: src.PathHash,
				Ctime:    src.Ctime,
				Mtime:    src.Mtime,
				Size:     src.Size,
				LastTime: src.UpdatedTimestamp,
			}
			return proto.Marshal(pbMsg)
		}
	// "NoteSyncRename"
	case NoteSyncRename:
		if src, ok := data.(dto.NoteSyncRenameMessage); ok {
			pbMsg := &v1.NoteSyncRenameMessage{
				Path:        src.Path,
				PathHash:    src.PathHash,
				ContentHash: src.ContentHash,
				Ctime:       src.Ctime,
				Mtime:       src.Mtime,
				Size:        src.Size,
				OldPath:     src.OldPath,
				OldPathHash: src.OldPathHash,
				LastTime:    src.UpdatedTimestamp,
			}
			return proto.Marshal(pbMsg)
		} else if m, ok := data.(map[string]interface{}); ok {
			pbMsg := &v1.NoteSyncRenameMessage{
				Path:        formatString(m["path"]),
				PathHash:    formatString(m["pathHash"]),
				OldPath:     formatString(m["oldPath"]),
				OldPathHash: formatString(m["oldPathHash"]),
			}
			return proto.Marshal(pbMsg)
		} else if m, ok := data.(map[string]any); ok {
			pbMsg := &v1.NoteSyncRenameMessage{
				Path:        formatString(m["path"]),
				PathHash:    formatString(m["pathHash"]),
				OldPath:     formatString(m["oldPath"]),
				OldPathHash: formatString(m["oldPathHash"]),
			}
			return proto.Marshal(pbMsg)
		}
	// "NoteSyncMtime"
	case NoteSyncMtime:
		if src, ok := data.(dto.NoteSyncMtimeMessage); ok {
			pbMsg := &v1.NoteSyncMtimeMessage{
				Path:     src.Path,
				Ctime:    src.Ctime,
				Mtime:    src.Mtime,
				LastTime: src.UpdatedTimestamp,
			}
			return proto.Marshal(pbMsg)
		}
	// "NoteSyncEnd"
	case NoteSyncEnd:
		if src, ok := data.(dto.NoteSyncEndMessage); ok {
			pbMsg := &v1.NoteSyncEndMessage{
				LastTime:           src.LastTime,
				NeedUploadCount:    src.NeedUploadCount,
				NeedModifyCount:    src.NeedModifyCount,
				NeedSyncMtimeCount: src.NeedSyncMtimeCount,
				NeedDeleteCount:    src.NeedDeleteCount,
			}
			return proto.Marshal(pbMsg)
		}
	// "NoteSyncNeedPush"
	case NoteSyncNeedPush:
		if src, ok := data.(dto.NoteSyncNeedPushMessage); ok {
			pbMsg := &v1.NoteSyncNeedPushMessage{
				Path:     src.Path,
				PathHash: src.PathHash,
			}
			return proto.Marshal(pbMsg)
		}
	// "NoteModifyAck"
	case NoteModifyAck:
		if src, ok := data.(dto.NoteModifyAckMessage); ok {
			pbMsg := &v1.NoteModifyAckMessage{
				LastTime: src.LastTime,
				Path:     src.Path,
				PathHash: src.PathHash,
			}
			return proto.Marshal(pbMsg)
		}
	// "NoteRenameAck"
	case NoteRenameAck:
		if src, ok := data.(dto.NoteRenameAckMessage); ok {
			pbMsg := &v1.NoteRenameAckMessage{
				LastTime: src.LastTime,
				Path:     src.Path,
				PathHash: src.PathHash,
			}
			return proto.Marshal(pbMsg)
		}
	// "NoteDeleteAck"
	case NoteDeleteAck:
		if src, ok := data.(dto.NoteDeleteAckMessage); ok {
			pbMsg := &v1.NoteDeleteAckMessage{
				LastTime: src.LastTime,
				Path:     src.Path,
				PathHash: src.PathHash,
			}
			return proto.Marshal(pbMsg)
		}

	// 2. File responses
	// "FileSyncUpdate"
	case FileSyncUpdate:
		if src, ok := data.(dto.FileSyncModifyMessage); ok {
			pbMsg := &v1.FileSyncModifyMessage{
				Path:        src.Path,
				PathHash:    src.PathHash,
				ContentHash: src.ContentHash,
				Size:        src.Size,
				Ctime:       src.Ctime,
				Mtime:       src.Mtime,
				LastTime:    src.UpdatedTimestamp,
			}
			return proto.Marshal(pbMsg)
		}
	// "FileSyncDelete"
	case FileSyncDelete:
		if src, ok := data.(dto.FileSyncDeleteMessage); ok {
			pbMsg := &v1.FileSyncDeleteMessage{
				Path:     src.Path,
				PathHash: src.PathHash,
				Ctime:    src.Ctime,
				Mtime:    src.Mtime,
				Size:     src.Size,
				LastTime: src.UpdatedTimestamp,
			}
			return proto.Marshal(pbMsg)
		}
	// "FileSyncRename"
	case FileSyncRename:
		if src, ok := data.(dto.FileSyncRenameMessage); ok {
			pbMsg := &v1.FileSyncRenameMessage{
				Path:        src.Path,
				PathHash:    src.PathHash,
				ContentHash: src.ContentHash,
				Ctime:       src.Ctime,
				Mtime:       src.Mtime,
				Size:        src.Size,
				LastTime:    src.UpdatedTimestamp,
				OldPath:     src.OldPath,
				OldPathHash: src.OldPathHash,
			}
			return proto.Marshal(pbMsg)
		} else if m, ok := data.(map[string]interface{}); ok {
			pbMsg := &v1.FileSyncRenameMessage{
				Path:        formatString(m["path"]),
				PathHash:    formatString(m["pathHash"]),
				OldPath:     formatString(m["oldPath"]),
				OldPathHash: formatString(m["oldPathHash"]),
			}
			return proto.Marshal(pbMsg)
		} else if m, ok := data.(map[string]any); ok {
			pbMsg := &v1.FileSyncRenameMessage{
				Path:        formatString(m["path"]),
				PathHash:    formatString(m["pathHash"]),
				OldPath:     formatString(m["oldPath"]),
				OldPathHash: formatString(m["oldPathHash"]),
			}
			return proto.Marshal(pbMsg)
		}
	// "FileSyncMtime"
	case FileSyncMtime:
		if src, ok := data.(dto.FileSyncMtimeMessage); ok {
			pbMsg := &v1.FileSyncMtimeMessage{
				Path:     src.Path,
				Ctime:    src.Ctime,
				Mtime:    src.Mtime,
				LastTime: src.UpdatedTimestamp,
			}
			return proto.Marshal(pbMsg)
		}
	// "FileSyncEnd"
	case FileSyncEnd:
		if src, ok := data.(dto.FileSyncEndMessage); ok {
			pbMsg := &v1.FileSyncEndMessage{
				LastTime:           src.LastTime,
				NeedUploadCount:    src.NeedUploadCount,
				NeedModifyCount:    src.NeedModifyCount,
				NeedSyncMtimeCount: src.NeedSyncMtimeCount,
				NeedDeleteCount:    src.NeedDeleteCount,
			}
			return proto.Marshal(pbMsg)
		}
	// "FileUpload"
	case FileUpload:
		if src, ok := data.(dto.FileSyncUploadMessage); ok {
			pbMsg := &v1.FileSyncUploadMessage{
				Path:      src.Path,
				PathHash:  src.PathHash,
				SessionId: src.SessionID,
				ChunkSize: src.ChunkSize,
			}
			return proto.Marshal(pbMsg)
		}
	// "FileSyncChunkDownload"
	case FileSyncChunkDownload:
		if src, ok := data.(dto.FileSyncDownloadMessage); ok {
			pbMsg := &v1.FileSyncDownloadMessage{
				Path:        src.Path,
				ContentHash: src.ContentHash,
				Ctime:       src.Ctime,
				Mtime:       src.Mtime,
				SessionId:   src.SessionID,
				ChunkSize:   src.ChunkSize,
				TotalChunks: src.TotalChunks,
				Size:        src.Size,
			}
			return proto.Marshal(pbMsg)
		}
	// "FileRenameAck"
	case FileRenameAck:
		if src, ok := data.(dto.FileRenameAckMessage); ok {
			pbMsg := &v1.FileRenameAckMessage{
				LastTime: src.LastTime,
				Path:     src.Path,
				PathHash: src.PathHash,
			}
			return proto.Marshal(pbMsg)
		}
	// "FileUploadAck"
	case FileUploadAck:
		if src, ok := data.(dto.FileUploadAckMessage); ok {
			pbMsg := &v1.FileUploadAckMessage{
				LastTime: src.LastTime,
				Path:     src.Path,
				PathHash: src.PathHash,
			}
			return proto.Marshal(pbMsg)
		}
	// "FileDeleteAck"
	case FileDeleteAck:
		if src, ok := data.(dto.FileDeleteAckMessage); ok {
			pbMsg := &v1.FileDeleteAckMessage{
				LastTime: src.LastTime,
				Path:     src.Path,
				PathHash: src.PathHash,
			}
			return proto.Marshal(pbMsg)
		}

	// 3. Setting responses
	// "SettingSyncModify"
	case SettingSyncModify:
		if src, ok := data.(dto.SettingSyncModifyMessage); ok {
			pbMsg := &v1.SettingSyncModifyMessage{
				Vault:       src.Vault,
				Path:        src.Path,
				PathHash:    src.PathHash,
				Content:     src.Content,
				ContentHash: src.ContentHash,
				Ctime:       src.Ctime,
				Mtime:       src.Mtime,
				LastTime:    src.UpdatedTimestamp,
			}
			return proto.Marshal(pbMsg)
		}
	// "SettingSyncDelete"
	case SettingSyncDelete:
		if src, ok := data.(dto.SettingSyncDeleteMessage); ok {
			pbMsg := &v1.SettingSyncDeleteMessage{
				Path:     src.Path,
				PathHash: src.PathHash,
				Ctime:    src.Ctime,
				Mtime:    src.Mtime,
				LastTime: src.UpdatedTimestamp,
			}
			return proto.Marshal(pbMsg)
		}
	// "SettingSyncMtime"
	case SettingSyncMtime:
		if src, ok := data.(dto.SettingSyncMtimeMessage); ok {
			pbMsg := &v1.SettingSyncMtimeMessage{
				Path:     src.Path,
				Ctime:    src.Ctime,
				Mtime:    src.Mtime,
				LastTime: src.UpdatedTimestamp,
			}
			return proto.Marshal(pbMsg)
		}
	// "SettingSyncEnd"
	case SettingSyncEnd:
		if src, ok := data.(dto.SettingSyncEndMessage); ok {
			pbMsg := &v1.SettingSyncEndMessage{
				LastTime:           src.LastTime,
				NeedUploadCount:    src.NeedUploadCount,
				NeedModifyCount:    src.NeedModifyCount,
				NeedSyncMtimeCount: src.NeedSyncMtimeCount,
				NeedDeleteCount:    src.NeedDeleteCount,
			}
			return proto.Marshal(pbMsg)
		}
	// "SettingSyncNeedUpload"
	case SettingSyncNeedUpload:
		if src, ok := data.(dto.SettingSyncNeedUploadMessage); ok {
			pbMsg := &v1.SettingSyncNeedUploadMessage{
				Path: src.Path,
			}
			return proto.Marshal(pbMsg)
		}
	// "SettingSyncClear"
	case SettingSyncClear:
		return []byte{}, nil
	// "SettingModifyAck"
	case SettingModifyAck:
		if src, ok := data.(dto.SettingModifyAckMessage); ok {
			pbMsg := &v1.SettingModifyAckMessage{
				LastTime: src.LastTime,
				Path:     src.Path,
				PathHash: src.PathHash,
			}
			return proto.Marshal(pbMsg)
		}
	// "SettingDeleteAck"
	case SettingDeleteAck:
		if src, ok := data.(dto.SettingDeleteAckMessage); ok {
			pbMsg := &v1.SettingDeleteAckMessage{
				LastTime: src.LastTime,
				Path:     src.Path,
				PathHash: src.PathHash,
			}
			return proto.Marshal(pbMsg)
		}
	// "FolderSyncModify"
	case FolderSyncModify:
		var src dto.FolderSyncModifyMessage
		var ok bool
		if src, ok = data.(dto.FolderSyncModifyMessage); !ok {
			var p *dto.FolderSyncModifyMessage
			if p, ok = data.(*dto.FolderSyncModifyMessage); ok && p != nil {
				src = *p
			}
		}
		if ok {
			pbMsg := &v1.FolderSyncModifyMessage{
				Path:     src.Path,
				PathHash: src.PathHash,
				Ctime:    src.Ctime,
				Mtime:    src.Mtime,
				LastTime: src.UpdatedTimestamp,
			}
			return proto.Marshal(pbMsg)
		}
	// "FolderSyncDelete"
	case FolderSyncDelete:
		var src dto.FolderSyncDeleteMessage
		var ok bool
		if src, ok = data.(dto.FolderSyncDeleteMessage); !ok {
			var p *dto.FolderSyncDeleteMessage
			if p, ok = data.(*dto.FolderSyncDeleteMessage); ok && p != nil {
				src = *p
			}
		}
		if ok {
			pbMsg := &v1.FolderSyncDeleteMessage{
				Path:     src.Path,
				PathHash: src.PathHash,
				Ctime:    src.Ctime,
				Mtime:    src.Mtime,
				LastTime: src.UpdatedTimestamp,
			}
			return proto.Marshal(pbMsg)
		}
	// "FolderSyncRename"
	case FolderSyncRename:
		var src dto.FolderSyncRenameMessage
		var ok bool
		if src, ok = data.(dto.FolderSyncRenameMessage); !ok {
			var p *dto.FolderSyncRenameMessage
			if p, ok = data.(*dto.FolderSyncRenameMessage); ok && p != nil {
				src = *p
			}
		}
		if ok {
			pbMsg := &v1.FolderSyncRenameMessage{
				Path:        src.Path,
				PathHash:    src.PathHash,
				Ctime:       src.Ctime,
				Mtime:       src.Mtime,
				OldPath:     src.OldPath,
				OldPathHash: src.OldPathHash,
				LastTime:    src.UpdatedTimestamp,
			}
			return proto.Marshal(pbMsg)
		} else if m, ok := data.(map[string]interface{}); ok {
			pbMsg := &v1.FolderSyncRenameMessage{
				Path:        formatString(m["path"]),
				PathHash:    formatString(m["pathHash"]),
				OldPath:     formatString(m["oldPath"]),
				OldPathHash: formatString(m["oldPathHash"]),
			}
			return proto.Marshal(pbMsg)
		} else if m, ok := data.(map[string]any); ok {
			pbMsg := &v1.FolderSyncRenameMessage{
				Path:        formatString(m["path"]),
				PathHash:    formatString(m["pathHash"]),
				OldPath:     formatString(m["oldPath"]),
				OldPathHash: formatString(m["oldPathHash"]),
			}
			return proto.Marshal(pbMsg)
		}
	// "FolderSyncEnd"
	case FolderSyncEnd:
		var src dto.FolderSyncEndMessage
		var ok bool
		if src, ok = data.(dto.FolderSyncEndMessage); !ok {
			var p *dto.FolderSyncEndMessage
			if p, ok = data.(*dto.FolderSyncEndMessage); ok && p != nil {
				src = *p
			}
		}
		if ok {
			pbMsg := &v1.FolderSyncEndMessage{
				LastTime:        src.LastTime,
				NeedModifyCount: src.NeedModifyCount,
				NeedDeleteCount: src.NeedDeleteCount,
			}
			return proto.Marshal(pbMsg)
		}
	// "FolderModifyAck"
	case FolderModifyAck:
		var src dto.FolderModifyAckMessage
		var ok bool
		if src, ok = data.(dto.FolderModifyAckMessage); !ok {
			var p *dto.FolderModifyAckMessage
			if p, ok = data.(*dto.FolderModifyAckMessage); ok && p != nil {
				src = *p
			}
		}
		if ok {
			pbMsg := &v1.FolderModifyAckMessage{
				LastTime: src.LastTime,
				Path:     src.Path,
				PathHash: src.PathHash,
			}
			return proto.Marshal(pbMsg)
		}
	// "FolderDeleteAck"
	case FolderDeleteAck:
		var src dto.FolderDeleteAckMessage
		var ok bool
		if src, ok = data.(dto.FolderDeleteAckMessage); !ok {
			var p *dto.FolderDeleteAckMessage
			if p, ok = data.(*dto.FolderDeleteAckMessage); ok && p != nil {
				src = *p
			}
		}
		if ok {
			pbMsg := &v1.FolderDeleteAckMessage{
				LastTime: src.LastTime,
				Path:     src.Path,
				PathHash: src.PathHash,
			}
			return proto.Marshal(pbMsg)
		}
	// "FolderRenameAck"
	case FolderRenameAck:
		var src dto.FolderRenameAckMessage
		var ok bool
		if src, ok = data.(dto.FolderRenameAckMessage); !ok {
			var p *dto.FolderRenameAckMessage
			if p, ok = data.(*dto.FolderRenameAckMessage); ok && p != nil {
				src = *p
			}
		}
		if ok {
			pbMsg := &v1.FolderRenameAckMessage{
				LastTime: src.LastTime,
				Path:     src.Path,
				PathHash: src.PathHash,
			}
			return proto.Marshal(pbMsg)
		}
	}

	// For unhandled message types, fallback to JSON encoding
	return json.Marshal(data)
}

func formatString(v any) string {
	if v == nil {
		return ""
	}
	if str, ok := v.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", v)
}
