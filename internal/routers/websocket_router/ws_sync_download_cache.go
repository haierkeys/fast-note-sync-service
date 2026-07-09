package websocket_router

import (
	"context"
	"sync"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/logger"
	"github.com/haierkeys/fast-note-sync-service/pkg/workerpool"
	"go.uber.org/zap"
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

	// SentPage/AckedPage/Window 是下行窗口流水线的状态机（同步流水线设计 §4.2），
	// 替代旧版 CurrentPage 的"发送游标"单一角色。
	// SentPage/AckedPage/Window form the download window pipeline state machine
	// (sync pipeline design §4.2), replacing the old CurrentPage's single "send cursor" role.
	SentPage  int // 已发出的页数，即下一待发页码 (0-indexed) // number of pages sent so far = next page index to send
	AckedPage int // 已确认处理完的最高连续页 + 1，初始 0 // highest confirmed contiguous page + 1, initial 0
	Window    int // 本连接协商的 W_down；0 = stop-and-wait（回滚开关） // negotiated W_down for this connection; 0 = stop-and-wait (rollback switch)

	UpdatedAt time.Time // 更新时间

	// FillContent 可选：当消息队列中携带了未填充正文的笔记（WSQueuedMessage.NoteID != 0）时，
	// 由构造 entry 的一方（ws_note.go）注入的按需读取函数。sendSyncPage 在发送某一页前，
	// 仅为该页内需要填充正文的消息并发调用它，避免把全部待发笔记正文一次性物化进内存。
	// FillContent optional: when the queue carries notes whose content hasn't been filled in
	// yet (WSQueuedMessage.NoteID != 0), the entry constructor (ws_note.go) injects this
	// on-demand reader. sendSyncPage calls it concurrently, but only for the page about to be
	// sent, avoiding materializing every pending note's content in memory at once.
	FillContent func(ctx context.Context, noteID int64) (string, error)
}

// totalPages returns the number of pages MessageQueue splits into at PageSize.
// totalPages 返回 MessageQueue 按 PageSize 切分后的总页数。
func (e *syncDownloadEntry) totalPages() int {
	if e.PageSize <= 0 {
		return 0
	}
	return (len(e.MessageQueue) + e.PageSize - 1) / e.PageSize
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

// pump 推进下行窗口（同步流水线设计 §4.2）：在窗口允许的范围内尽量多发送尚未发出的页。
// Window=0（stop-and-wait）时 max(Window,1)=1，循环每次调用最多发一页，与 3.5.x 前逐页行为
// 逐消息等价。isLast 页发送完成后立即销毁 entry 并返回，不再继续推进。
// 调用前必须持有 entry.mu。
// pump advances the download window (sync pipeline design §4.2): sends as many not-yet-sent
// pages as the window allows. Window=0 (stop-and-wait) makes max(Window,1)=1, so each call sends
// at most one page — message-for-message equivalent to pre-3.6.0 per-page behavior. Once the
// isLast page is sent, the entry is destroyed immediately and pump stops advancing. Caller must
// hold entry.mu.
func pump(c *pkgapp.WebsocketClient, entry *syncDownloadEntry) {
	window := entry.Window
	if window < 1 {
		window = 1
	}
	total := entry.totalPages()
	for entry.SentPage < total && entry.SentPage < entry.AckedPage+window {
		isLast := sendSyncPageFunc(c, entry)
		entry.SentPage++
		entry.UpdatedAt = time.Now()
		if isLast {
			// 发完即毁：客户端不会为最后一页发 ack（现状保留），也无需再持有 entry
			// Destroy on completion: the client never acks the last page (unchanged behavior), no need to keep the entry
			syncDownloadDelete(entry.Context, entry.TypeName)
			return
		}
	}
}

// sendSyncPageFunc is a seam over sendSyncPage: pump() calls through this package-level var
// instead of the function directly, so tests can swap in a fake that records SentPage/isLast
// without needing a live WebsocketClient/gws.Conn to actually write frames to (this codebase's
// existing tests never exercise the real conn write path either, see
// pkg/app/websocket_client_test.go). Production code never reassigns it.
// sendSyncPageFunc 是 sendSyncPage 上的一个可替换钩子：pump() 通过这个包级变量调用而非直接
// 调用函数，使测试能替换成记录 SentPage/isLast 的假实现，而不需要一个能真正写帧的
// WebsocketClient/gws.Conn（本仓库既有测试也从不触碰真实 conn 写入路径，
// 参见 pkg/app/websocket_client_test.go）。生产代码从不重新赋值它。
var sendSyncPageFunc = sendSyncPage

// sendSyncPage 发送 entry.SentPage 指向的页（页元数据 + 该页全部明细），返回该页是否为最后一页。
// 不负责推进 SentPage、不负责判断是否销毁 entry —— 这两件事由调用方 pump 统一处理，
// 使"发送"与"窗口推进"职责分离。调用前必须持有 entry.mu（由 pump 的调用者持有）。
// sendSyncPage sends the page at entry.SentPage (page metadata + all of that page's detail
// messages) and reports whether it was the last page. It does not advance SentPage or decide
// whether to destroy the entry — pump owns both, keeping "send" and "window advance"
// responsibilities separate. Caller must hold entry.mu (held by pump's caller).
func sendSyncPage(c *pkgapp.WebsocketClient, entry *syncDownloadEntry) (isLast bool) {
	page := entry.SentPage
	start := page * entry.PageSize
	end := start + entry.PageSize
	if end > len(entry.MessageQueue) {
		end = len(entry.MessageQueue)
	}

	chunk := entry.MessageQueue[start:end]
	isLast = end == len(entry.MessageQueue)

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
		return isLast
	}

	// 明细信封 pageIndex（§2.4）：仅 pv>=2 连接填写，旧连接维持零值/缺省
	// Detail envelope pageIndex (§2.4): only filled for pv>=2 connections, old connections keep the zero value/omitted
	withPageIndex := func(c2 *code.Code) *code.Code {
		if c.ProtoVersion >= 2 {
			return c2.WithPageIndex(page)
		}
		return c2
	}

	// 1. 发送 Page 页面控制指示元数据 (不含消息体，仅作元数据声明)
	c.ToResponse(withPageIndex(code.Success.WithData(dto.SyncPageMessage{
		PageIndex:  page,
		PageSize:   entry.PageSize,
		TotalCount: len(chunk),
		IsLast:     isLast,
	})).WithVault(entry.Vault).WithContext(entry.Context), string(pageAction))

	// 2. 紧接着逐个发送本页的所有明细消息
	for _, msg := range chunk {
		c.ToResponse(withPageIndex(code.Success.WithData(msg.Data)).WithVault(entry.Vault).WithContext(msg.Context), msg.Action)
	}

	return isLast
}

// handlePageAck 实现四处 XxxSyncPageAck handler 共用的 §4.2 分支表，替代原先四处各自维护的
// mismatch 兜底代码。抽成共用函数是为了让这张最容易出错的状态表只有一份实现——四处调用点
// （Note/File/Setting/FolderSyncPageAck）各自只负责绑参、按类型取 entry、传入自己的类型名用于
// 日志，核心分支逻辑单点维护，避免四份手抄代码在边界条件上不知不觉分叉。
// 调用前 entry 必须已通过 syncDownloadGet 找到；本函数内部获取并释放 entry.mu。
// handlePageAck implements the §4.2 branch table shared by all four XxxSyncPageAck handlers,
// replacing the old per-file mismatch fallback code. It's factored out so this — the single
// most error-prone piece of state logic in the whole design — has exactly one implementation:
// the four call sites (Note/File/Setting/FolderSyncPageAck) only bind params, look up their
// type's entry, and pass their type name for logging; the branch logic itself is maintained in
// one place so four hand-copied versions can't quietly drift apart at the edge cases.
// entry must already have been resolved via syncDownloadGet before calling; this function
// acquires and releases entry.mu internally.
func handlePageAck(c *pkgapp.WebsocketClient, entry *syncDownloadEntry, pageIndex int, typeName string, log *zap.Logger, traceID string) {
	entry.mu.Lock()
	defer entry.mu.Unlock()

	if pageIndex == -1 {
		// 首拉：Window=0 时 pump 恰好只发 1 页，与 3.5.x 前 stop-and-wait 现状逐消息等价
		// First pull: with Window=0, pump sends exactly 1 page, message-for-message
		// equivalent to pre-3.6.0 stop-and-wait
		entry.AckedPage = 0
		pump(c, entry)
		return
	}

	switch {
	case pageIndex < entry.AckedPage-1:
		// 过期重复 ack：忽略
		// Expired duplicate ack: ignore
		log.Debug("SyncPageAck: expired duplicate ack, ignoring",
			zap.String(logger.FieldTraceID, traceID),
			zap.String("type", typeName),
			zap.String("context", entry.Context),
			zap.Int("ackedPage", entry.AckedPage),
			zap.Int("got", pageIndex))
	case pageIndex == entry.AckedPage-1:
		// 客户端重发了上一 ack，说明它没收到后续页：回退 SentPage=AckedPage，重发窗口
		// （64be9cbc 的"重发当前页"兜底意图在此完整保留，见设计 §4.2）
		// Client resent its previous ack, meaning it never received the subsequent pages:
		// rewind SentPage=AckedPage and resend the window (fully preserves 64be9cbc's
		// "resend current page" fallback intent, see design §4.2)
		log.Warn("SyncPageAck: received retransmitted ack for previous page, rewinding window and resending",
			zap.String(logger.FieldTraceID, traceID),
			zap.String("type", typeName),
			zap.String("context", entry.Context),
			zap.Int("ackedPage", entry.AckedPage),
			zap.Int("sentPage", entry.SentPage),
			zap.Int("got", pageIndex))
		entry.SentPage = entry.AckedPage
		pump(c, entry)
	case pageIndex < entry.SentPage:
		// AckedPage-1 < pageIndex < SentPage：正常推进水位（允许乱序 ack，取最高水位）
		// AckedPage-1 < pageIndex < SentPage: normal watermark advance (out-of-order acks
		// allowed, highest watermark wins)
		entry.AckedPage = pageIndex + 1
		entry.UpdatedAt = time.Now()
		pump(c, entry)
	default:
		// pageIndex >= SentPage：客户端 ack 了一个从未发过的页，异常，仅告警不处理
		// pageIndex >= SentPage: client acked a page that was never sent; abnormal, warn only
		log.Warn("SyncPageAck: client acked a page never sent, ignoring",
			zap.String(logger.FieldTraceID, traceID),
			zap.String("type", typeName),
			zap.String("context", entry.Context),
			zap.Int("sentPage", entry.SentPage),
			zap.Int("got", pageIndex))
	}
}
