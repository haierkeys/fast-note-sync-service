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
	mu            sync.Mutex    // 保护并发 append（Guards concurrent appends）
	Items         []interface{} // 累积的分批数据（Accumulated batch items）
	ReceivedCount int           // 已收到批次数（Received batch count）
	TotalBatches  int           // 期望总批次数（Expected total batches）

	DelItems     []interface{} // 删除列表（Delete list）
	MissingItems []interface{} // 缺失列表（Missing list）

	UpdatedAt time.Time // 最近一次更新时间，用于 TTL 清理（Last update time for TTL cleanup）
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

// syncBatchGetOrCreate 获取或创建指定 context+type 的缓存条目
// Get or create a batch cache entry for the given context + type
func syncBatchGetOrCreate(context, typeName string, totalBatches int) *syncBatchEntry {
	key := syncBatchKey(context, typeName)
	val, _ := syncBatchCacheMap.LoadOrStore(key, &syncBatchEntry{
		Items:        make([]interface{}, 0),
		DelItems:     make([]interface{}, 0),
		MissingItems: make([]interface{}, 0),
		TotalBatches: totalBatches,
		UpdatedAt:    time.Now(),
	})
	return val.(*syncBatchEntry)
}

// syncBatchDelete 清理指定 context+type 的缓存条目
// Delete the batch cache entry for the given context + type
func syncBatchDelete(context, typeName string) {
	syncBatchCacheMap.Delete(syncBatchKey(context, typeName))
}
