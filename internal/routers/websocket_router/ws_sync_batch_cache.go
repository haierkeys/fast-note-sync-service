package websocket_router

import (
	"sync"
	"time"
)

// syncBatchEntry 单个类型的分批缓存条目
// key 格式："{context}_{type}"，例如 "uuid-xxx_note"、"uuid-xxx_file"
// 这样 NoteSync / FileSync / SettingSync / FolderSync 即使共用同一 context 也不会互相污染
// Single-type batch cache entry. Key format: "{context}_{type}" to prevent cross-type pollution.
type syncBatchEntry struct {
	mu              sync.Mutex       // 保护并发 append（Guards concurrent appends）
	Items           []interface{}    // 累积的分批数据（Accumulated batch items）
	ReceivedCount   int              // 已收到批次数（Received batch count）
	TotalBatches    int              // 期望总批次数（Expected total batches）
	ReceivedIndexes map[int]struct{} // 已收到的 BatchIndex 集合，用于去重重复上传（Received BatchIndex set, dedups retransmitted batches）

	DelItems     []interface{} // 删除列表（Delete list）
	MissingItems []interface{} // 缺失列表（Missing list）

	UpdatedAt time.Time // 最近一次更新时间，用于 TTL 清理（Last update time for TTL cleanup）
}

// markBatchReceived 检查给定 BatchIndex 是否已经收到过；若是首次收到则记录并返回 false，
// 若是重复收到（客户端因未收到 ack 而重传）则返回 true，调用方应跳过 append/计数、只重发 ack。
// 调用前必须持有 entry.mu。
// markBatchReceived checks whether batchIndex has already been received; on first receipt
// it records it and returns false, on a duplicate (client retransmitted after missing the
// ack) it returns true and the caller should skip append/count and just resend the ack.
// Caller must hold entry.mu.
func (e *syncBatchEntry) markBatchReceived(batchIndex int) (isDuplicate bool) {
	if e.ReceivedIndexes == nil {
		e.ReceivedIndexes = make(map[int]struct{})
	}
	if _, ok := e.ReceivedIndexes[batchIndex]; ok {
		return true
	}
	e.ReceivedIndexes[batchIndex] = struct{}{}
	return false
}

// syncBatchCacheMap 全局分批缓存 Map（Global batch cache map）
var syncBatchCacheMap sync.Map

const syncBatchCacheTTL = 5 * time.Minute

func init() {
	// 后台协程定时清理过期缓存，防止客户端异常离线导致内存泄漏
	// Background goroutine periodically cleans expired entries to prevent memory leaks on client disconnect
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			now := time.Now()
			syncBatchCacheMap.Range(func(k, v interface{}) bool {
				if now.Sub(v.(*syncBatchEntry).UpdatedAt) > syncBatchCacheTTL {
					syncBatchCacheMap.Delete(k)
				}
				return true
			})
		}
	}()
}

// syncBatchKey 构造缓存 key（Build cache key from context + type name）
func syncBatchKey(context, typeName string) string {
	return context + "_" + typeName
}

// syncBatchGetOrCreate 获取或创建指定 context+type 的缓存条目，created 表示本次调用是否新建了条目
// （用于观测：若某个 context+type 早已 doSync 完成并被 syncBatchDelete 清理，随后又出现一次
// created==true 的调用，说明这是客户端的迟到批次重传，重建出的 entry 将成为等待 5 分钟 TTL
// 回收的孤儿，见同步流水线设计 §3.3 第 2 点）。
// Get or create a batch cache entry for the given context + type; created reports whether this
// call created a new entry (observability: if a context+type has already completed doSync and
// been syncBatchDelete'd, and a later call here again returns created==true, that's a late batch
// retransmit from the client rebuilding an orphan entry that will sit until the 5-minute TTL
// reclaims it — see sync pipeline design §3.3 point 2).
func syncBatchGetOrCreate(context, typeName string, totalBatches int) (entry *syncBatchEntry, created bool) {
	key := syncBatchKey(context, typeName)
	val, loaded := syncBatchCacheMap.LoadOrStore(key, &syncBatchEntry{
		Items:           make([]interface{}, 0),
		DelItems:        make([]interface{}, 0),
		MissingItems:    make([]interface{}, 0),
		TotalBatches:    totalBatches,
		ReceivedIndexes: make(map[int]struct{}),
		UpdatedAt:       time.Now(),
	})
	return val.(*syncBatchEntry), !loaded
}

// syncBatchDelete 清理指定 context+type 的缓存条目
// Delete the batch cache entry for the given context + type
func syncBatchDelete(context, typeName string) {
	syncBatchCacheMap.Delete(syncBatchKey(context, typeName))
}
