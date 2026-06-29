package websocket_router

import (
	"sync"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
)

// syncDownloadEntry 单个类型的分页下载缓存条目
// key 格式："{context}_{type}"，例如 "uuid-xxx_note"、"uuid-xxx_file"
type syncDownloadEntry struct {
	mu           sync.Mutex        // 互斥锁保护并发读取与翻页
	Context      string            // 同步上下文
	TypeName     string            // 同步类型 "note" | "file" | "setting" | "folder"
	Vault        string            // 仓库名称
	MessageQueue []WSQueuedMessage // 待发送的全部明细队列
	PageSize     int               // 每页大小
	CurrentPage  int               // 当前已发送的页码 (0-indexed)
	UpdatedAt    time.Time         // 更新时间
}

var syncDownloadCacheMap sync.Map

const syncDownloadCacheTTL = 10 * time.Minute

func init() {
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			now := time.Now()
			syncDownloadCacheMap.Range(func(k, v interface{}) bool {
				if now.Sub(v.(*syncDownloadEntry).UpdatedAt) > syncDownloadCacheTTL {
					syncDownloadCacheMap.Delete(k)
				}
				return true
			})
		}
	}()
}

func syncDownloadGet(context, typeName string) (*syncDownloadEntry, bool) {
	key := context + "_" + typeName
	val, ok := syncDownloadCacheMap.Load(key)
	if !ok {
		return nil, false
	}
	return val.(*syncDownloadEntry), true
}

func syncDownloadStore(context, typeName string, entry *syncDownloadEntry) {
	key := context + "_" + typeName
	entry.UpdatedAt = time.Now()
	syncDownloadCacheMap.Store(key, entry)
}

func syncDownloadDelete(context, typeName string) {
	key := context + "_" + typeName
	syncDownloadCacheMap.Delete(key)
}

// sendSyncPage 发送指定页码的明细
func sendSyncPage(c *pkgapp.WebsocketClient, entry *syncDownloadEntry) {
	start := entry.CurrentPage * entry.PageSize
	end := start + entry.PageSize
	if end > len(entry.MessageQueue) {
		end = len(entry.MessageQueue)
	}

	chunk := entry.MessageQueue[start:end]
	isLast := end == len(entry.MessageQueue)

	var pageAction WebSocketSendAction
	switch entry.TypeName {
	case "note":
		pageAction = NoteSyncPage
	case "file":
		pageAction = FileSyncPage
	case "setting":
		pageAction = SettingSyncPage
	case "folder":
		pageAction = FolderSyncPage
	default:
		return
	}

	// 1. 发送 Page 页面控制指示元数据
	c.ToResponse(code.Success.WithData(dto.SyncPageMessage{
		Context:    entry.Context,
		PageIndex:  entry.CurrentPage,
		PageSize:   entry.PageSize,
		TotalCount: len(chunk),
		IsLast:     isLast,
	}).WithVault(entry.Vault).WithContext(entry.Context), pageAction)

	// 2. 发送本页的所有明细消息
	for _, item := range chunk {
		c.ToResponse(code.Success.WithData(item.Data).WithVault(entry.Vault).WithContext(entry.Context), item.Action)
	}
}
