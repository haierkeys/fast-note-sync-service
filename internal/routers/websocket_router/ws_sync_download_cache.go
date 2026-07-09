package websocket_router

import (
	"context"
	"sync"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/workerpool"
)

// syncDownloadEntry 单个类型的分页下载缓存条目
// key 格式："{context}_{type}"，例如 "uuid-xxx_note"、"uuid-xxx_file"
type syncDownloadEntry struct {
	mu           sync.Mutex            // 互斥锁保护并发读取与翻页
	Context      string                // 同步上下文
	TypeName     string                // 同步类型 "note" | "file" | "setting" | "folder"
	Vault        string                // 仓库名称
	MessageQueue []dto.WSQueuedMessage // 待发送的全部明细队列
	PageSize     int                   // 每页大小
	CurrentPage  int                   // 当前已发送的页码 (0-indexed)
	UpdatedAt    time.Time             // 更新时间

	// FillContent 可选：当消息队列中携带了未填充正文的笔记（WSQueuedMessage.NoteID != 0）时，
	// 由构造 entry 的一方（ws_note.go）注入的按需读取函数。sendSyncPage 在发送某一页前，
	// 仅为该页内需要填充正文的消息并发调用它，避免把全部待发笔记正文一次性物化进内存。
	// FillContent optional: when the queue carries notes whose content hasn't been filled in
	// yet (WSQueuedMessage.NoteID != 0), the entry constructor (ws_note.go) injects this
	// on-demand reader. sendSyncPage calls it concurrently, but only for the page about to be
	// sent, avoiding materializing every pending note's content in memory at once.
	FillContent func(ctx context.Context, noteID int64) (string, error)
}

// noteContentFillPool 是用于分页发送前按需回填笔记正文的小并发 worker pool，
// 并发度限制在个位数，避免大批量笔记回填时把磁盘 IO 打成突发洪峰。
// noteContentFillPool is a small worker pool used to lazily fill note content right
// before a page is sent. Concurrency is capped in the low single digits to avoid
// bursting disk IO when a large batch of notes needs to be filled.
var noteContentFillPool = workerpool.New(&workerpool.Config{MaxWorkers: 6, QueueSize: 256}, nil)

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

	// 按需回填本页内尚未填充正文的笔记消息，用后即可被后续 GC 回收，
	// 而不是在 doNoteSync 阶段就把全部待发笔记正文一次性物化进内存。
	// Lazily fill content for messages in this page that still need it, so memory
	// stays bounded to one page's worth of content instead of the whole queue.
	if entry.FillContent != nil {
		var wg sync.WaitGroup
		for i := range chunk {
			if chunk[i].NoteID == 0 {
				continue
			}
			idx := i
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = noteContentFillPool.Submit(c.Context(), func(ctx context.Context) error {
					content, err := entry.FillContent(ctx, chunk[idx].NoteID)
					if err != nil {
						return err
					}
					if m, ok := chunk[idx].Data.(dto.NoteSyncModifyMessage); ok {
						m.Content = content
						chunk[idx].Data = m
					}
					return nil
				})
			}()
		}
		wg.Wait()
	}

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

	// 1. 发送 Page 页面控制指示元数据 (不含消息体，仅作元数据声明)
	c.ToResponse(code.Success.WithData(dto.SyncPageMessage{
		PageIndex:  entry.CurrentPage,
		PageSize:   entry.PageSize,
		TotalCount: len(chunk),
		IsLast:     isLast,
	}).WithVault(entry.Vault).WithContext(entry.Context), string(pageAction))

	// 2. 紧接着逐个发送本页的所有明细消息
	for _, msg := range chunk {
		c.ToResponse(code.Success.WithData(msg.Data).WithVault(entry.Vault).WithContext(msg.Context), msg.Action)
	}

	// 3. 如果是最后一页，主动销毁服务端下载缓存，客户端无需发送最终确认 PageAck
	// 3. If it's the last page, actively delete download cache on server, client doesn't need to send final PageAck
	if isLast {
		syncDownloadDelete(entry.Context, entry.TypeName)
	}
}
