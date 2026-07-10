package dao

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/blevesearch/bleve/v2"
	_ "github.com/blevesearch/bleve/v2/analysis/lang/cjk"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/haierkeys/fast-note-sync-service/pkg/safego"
	"go.uber.org/zap"
)

// ftsBatchMaxSize is the max number of pending ops per index before an immediate flush
// ftsBatchMaxSize 是单个索引待处理操作数触发立即刷新的上限
const ftsBatchMaxSize = 200

// ftsBatchFlushInterval is the max time a pending batch may wait before being flushed
// ftsBatchFlushInterval 是待处理批次刷新前可等待的最长时间
const ftsBatchFlushInterval = 200 * time.Millisecond

// ftsQueueSize is the buffer size of the async FTS op channel
// ftsQueueSize 是异步 FTS 操作 channel 的缓冲区大小
const ftsQueueSize = 4096

// ftsOp represents a single queued asynchronous FTS index mutation.
// A nil doc means the op is a delete; a non-nil barrier forces an immediate
// flush of all pending batches and signals completion once done (used by
// graceful shutdown and tests to make async writes observable).
// ftsOp 表示一个排队等待的异步 FTS 索引变更。doc 为 nil 表示删除；
// barrier 非 nil 时表示这是一个强制刷新屏障，刷新完成后关闭该 channel
// （用于优雅关闭与测试场景下让异步写入变为可观察）。
type ftsOp struct {
	uid, vaultID int64
	docID        string
	doc          *BleveNoteDoc
	barrier      chan struct{}
}

// ftsBatchKey identifies the per-vault pending batch a queued op belongs to
// ftsBatchKey 标识一个排队操作所属的按仓库划分的待处理批次
type ftsBatchKey struct {
	uid, vaultID int64
}

// ftsPendingBatch accumulates Bleve batch operations for a single vault index
// ftsPendingBatch 为单个仓库索引累积 Bleve 批次操作
type ftsPendingBatch struct {
	index bleve.Index
	batch *bleve.Batch
	count int
}

// BleveMeta metadata stored alongside the index to detect configuration changes
// BleveMeta 存储在索引旁的元数据，用于检测配置变化
type BleveMeta struct {
	FtsBleveStoreRaw bool `json:"fts-bleve-store-raw"` // Config value for store raw content // 是否存储原始文本配置值
	Version          int  `json:"version"`             // Metadata schema version // 元数据版本号
}

// BleveNoteDoc defines the document structure indexed in Bleve
// BleveNoteDoc 定义在 Bleve 中索引的文档结构
type BleveNoteDoc struct {
	ID      string  `json:"id"`       // Note ID // 笔记 ID
	Path    string  `json:"path"`     // Path for tokenized search // 路径（分词搜索）
	PathRaw string  `json:"path_raw"` // Raw Path untokenized for sorting // 原始路径（不分词，用于字母排序）
	Content string  `json:"content"`  // Note content // 笔记内容
	Action  string  `json:"action"`   // Action (e.g. "delete" for soft delete) // 操作类型（如软删除的 "delete"）
	Rename  float64 `json:"rename"`   // Rename flag // 重命名标志
	Ctime   float64 `json:"ctime"`    // Creation time // 创建时间
	Mtime   float64 `json:"mtime"`    // Modification time // 修改时间
}

// BleveManager manages the lifecycle of Bleve index instances per vault
// BleveManager 管理每个仓库 of Bleve 索引实例的生命周期
type BleveManager struct {
	enabled  bool        // Whether Bleve FTS is enabled // 是否启用 Bleve 全文搜索
	storeRaw bool        // Whether to store raw content in search index // 是否在搜索索引中存储原始内容
	logger   *zap.Logger // Logger instance // 日志记录器实例
	indexes  sync.Map    // Cached open bleve.Index instances, keyed by "uid_vaultID" // 已打开的 bleve.Index 实例缓存，键为 "uid_vaultID"
	mu       sync.Mutex  // Mutex protecting open/create operations on index files // 保护索引文件打开/创建操作的互斥锁

	ftsQueue    chan ftsOp     // Async FTS mutation queue consumed by ftsWorker // 由 ftsWorker 消费的异步 FTS 变更队列
	ftsWorkerWG sync.WaitGroup // Tracks the background ftsWorker goroutine // 跟踪后台 ftsWorker goroutine
	ftsStopOnce sync.Once      // Ensures the queue is closed at most once // 保证队列只被关闭一次
}

// NewBleveManager creates a new BleveManager instance
// NewBleveManager 创建一个新的 BleveManager 实例
func NewBleveManager(enabled *bool, storeRaw *bool, logger *zap.Logger) *BleveManager {
	en := true
	if enabled != nil {
		en = *enabled
	}
	raw := true
	if storeRaw != nil {
		raw = *storeRaw
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	m := &BleveManager{
		enabled:  en,
		storeRaw: raw,
		logger:   logger,
		ftsQueue: make(chan ftsOp, ftsQueueSize),
	}

	m.ftsWorkerWG.Add(1)
	safego.Go(logger, func() {
		defer m.ftsWorkerWG.Done()
		m.ftsWorker()
	})

	return m
}

// IsEnabled returns whether Bleve FTS is enabled
// IsEnabled 返回是否启用 Bleve 全文搜索
func (m *BleveManager) IsEnabled() bool {
	return m.enabled
}

// EnqueueUpsert asynchronously queues a note upsert into the Bleve FTS index.
// The write-path caller does not wait for the index write to complete; the
// background ftsWorker batches ops per vault via Bleve's native Batch API.
// EnqueueUpsert 异步投递一次笔记的 FTS 新增/更新。写路径调用方不等待索引写入完成，
// 后台 ftsWorker 使用 Bleve 原生 Batch API 按仓库攒批写入。
func (m *BleveManager) EnqueueUpsert(uid, vaultID int64, doc BleveNoteDoc) {
	m.ftsQueue <- ftsOp{uid: uid, vaultID: vaultID, docID: doc.ID, doc: &doc}
}

// EnqueueDelete asynchronously queues a note delete from the Bleve FTS index.
// EnqueueDelete 异步投递一次笔记的 FTS 删除。
func (m *BleveManager) EnqueueDelete(uid, vaultID int64, docID string) {
	m.ftsQueue <- ftsOp{uid: uid, vaultID: vaultID, docID: docID}
}

// FlushSync forces the background worker to immediately flush all pending
// batches and blocks until the flush completes. Used by graceful shutdown
// (before closing indexes) and by tests that need to observe async writes.
// FlushSync 强制后台 worker 立即刷新所有待处理批次，并阻塞等待刷新完成。
// 用于优雅关闭前的排空（关闭索引之前）以及需要观察异步写入结果的测试。
func (m *BleveManager) FlushSync() {
	done := make(chan struct{})
	m.ftsQueue <- ftsOp{barrier: done}
	<-done
}

// Shutdown stops accepting new async FTS ops, flushes all pending batches and
// waits for the background worker to exit. It must be called before CloseAll
// so no batch write races with an index being closed.
// Shutdown 停止接收新的异步 FTS 操作，刷新所有待处理批次并等待后台 worker 退出。
// 必须在 CloseAll 之前调用，避免批次写入与索引关闭发生竞争。
func (m *BleveManager) Shutdown() {
	m.ftsStopOnce.Do(func() {
		close(m.ftsQueue)
	})
	m.ftsWorkerWG.Wait()
}

// ftsWorker consumes queued FTS ops, accumulating them into per-vault Bleve
// batches that are flushed when a batch reaches ftsBatchMaxSize or
// ftsBatchFlushInterval elapses, whichever comes first. Ops for the same
// docID are appended to the batch in arrival order, so upsert/delete on the
// same note stay ordered within a flush.
// ftsWorker 消费排队的 FTS 操作，将其累积到按仓库划分的 Bleve 批次中，
// 批次达到 ftsBatchMaxSize 或等待超过 ftsBatchFlushInterval（先到者为准）时刷新。
// 同一 docID 的操作按到达顺序追加进批次，保证同一笔记的 upsert/delete 在一次刷新内保序。
func (m *BleveManager) ftsWorker() {
	pending := make(map[ftsBatchKey]*ftsPendingBatch)
	ticker := time.NewTicker(ftsBatchFlushInterval)
	defer ticker.Stop()

	flush := func(key ftsBatchKey) {
		pb, ok := pending[key]
		if !ok {
			return
		}
		delete(pending, key)
		if pb.count == 0 {
			return
		}
		if err := pb.index.Batch(pb.batch); err != nil {
			m.logger.Error("failed to flush async Bleve FTS batch",
				zap.Int64("uid", key.uid),
				zap.Int64("vaultID", key.vaultID),
				zap.Int("count", pb.count),
				zap.Error(err))
		}
	}

	flushAll := func() {
		for key := range pending {
			flush(key)
		}
	}

	for {
		select {
		case op, ok := <-m.ftsQueue:
			if !ok {
				flushAll()
				return
			}
			if op.barrier != nil {
				flushAll()
				close(op.barrier)
				continue
			}

			key := ftsBatchKey{uid: op.uid, vaultID: op.vaultID}
			pb, ok := pending[key]
			if !ok {
				index, err := m.GetIndex(op.uid, op.vaultID)
				if err != nil {
					m.logger.Error("failed to get Bleve index for async FTS op",
						zap.Int64("uid", op.uid),
						zap.Int64("vaultID", op.vaultID),
						zap.Error(err))
					continue
				}
				pb = &ftsPendingBatch{index: index, batch: index.NewBatch()}
				pending[key] = pb
			}

			if op.doc != nil {
				if err := pb.batch.Index(op.docID, op.doc); err != nil {
					m.logger.Error("failed to stage Bleve FTS index op", zap.String("docID", op.docID), zap.Error(err))
					continue
				}
			} else {
				pb.batch.Delete(op.docID)
			}
			pb.count++

			if pb.count >= ftsBatchMaxSize {
				flush(key)
			}
		case <-ticker.C:
			flushAll()
		}
	}
}

// GetIndexPath gets the path to the Bleve index folder for a specific vault
// GetIndexPath 获取特定仓库的 Bleve 索引文件夹路径
func (m *BleveManager) GetIndexPath(uid, vaultID int64) string {
	return filepath.Join("storage", "vault_fts", fmt.Sprintf("u_%d", uid), fmt.Sprintf("v_%d", vaultID))
}

// GetIndex gets or opens a Bleve index for a specific vault
// GetIndex 获取或打开特定仓库的 Bleve 索引
func (m *BleveManager) GetIndex(uid, vaultID int64) (bleve.Index, error) {
	key := fmt.Sprintf("%d_%d", uid, vaultID)
	if val, ok := m.indexes.Load(key); ok {
		return val.(bleve.Index), nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double check cache
	// 双重检查缓存
	if val, ok := m.indexes.Load(key); ok {
		return val.(bleve.Index), nil
	}

	path := m.GetIndexPath(uid, vaultID)
	metaPath := filepath.Join(path, "meta.json")

	// Check if metadata exists and config matches
	// 检查元数据是否存在，以及配置是否匹配
	rebuildNeeded := false
	if _, err := os.Stat(path); err == nil {
		metaData, err := os.ReadFile(metaPath)
		if err != nil {
			rebuildNeeded = true // Metadata missing/unreadable, rebuild to be safe // 元数据丢失/不可读，为安全起见执行重建
		} else {
			var meta BleveMeta
			if err := json.Unmarshal(metaData, &meta); err != nil {
				rebuildNeeded = true
			} else if meta.FtsBleveStoreRaw != m.storeRaw {
				m.logger.Info("FtsBleveStoreRaw config changed, rebuilding FTS index",
					zap.Int64("uid", uid),
					zap.Int64("vaultID", vaultID),
					zap.Bool("old", meta.FtsBleveStoreRaw),
					zap.Bool("new", m.storeRaw))
				rebuildNeeded = true
			}
		}
	}

	if rebuildNeeded {
		_ = m.closeAndClean(uid, vaultID)
	}

	var index bleve.Index
	var openErr error

	// If index path doesn't exist, create it and write meta.json
	// 如果索引路径不存在，创建它并写入 meta.json
	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, fmt.Errorf("failed to create index directory: %w", err)
		}

		// Write meta.json
		// 写入 meta.json
		meta := BleveMeta{
			FtsBleveStoreRaw: m.storeRaw,
			Version:          1,
		}
		metaData, marshalErr := json.Marshal(meta)
		if marshalErr == nil {
			_ = os.WriteFile(metaPath, metaData, 0644)
		}

		// Create a new Bleve index mapping
		// 创建全新 Bleve 索引映射
		mapping := m.createIndexMapping()
		index, openErr = bleve.New(path, mapping)
		if openErr != nil {
			return nil, fmt.Errorf("failed to create new bleve index: %w", openErr)
		}
	} else {
		// Open existing Bleve index
		// 打开已存在的 Bleve 索引
		index, openErr = bleve.Open(path)
		if openErr != nil {
			// If failed to open, might be corrupted, delete and recreate
			// 如果打开失败可能是索引损坏，尝试删除重建
			m.logger.Warn("failed to open bleve index, trying to recreate", zap.String("path", path), zap.Error(openErr))
			_ = m.closeAndClean(uid, vaultID)
			return m.GetIndex(uid, vaultID)
		}
	}

	m.indexes.Store(key, index)
	return index, nil
}

// Close closes a specific vault's index
// Close 关闭特定仓库的索引
func (m *BleveManager) Close(uid, vaultID int64) error {
	key := fmt.Sprintf("%d_%d", uid, vaultID)
	if val, ok := m.indexes.Load(key); ok {
		index := val.(bleve.Index)
		m.indexes.Delete(key)
		return index.Close()
	}
	return nil
}

// CloseAll closes all open index instances (used on graceful shutdown)
// CloseAll 关闭所有已打开的索引实例（用于优雅关闭）
func (m *BleveManager) CloseAll() error {
	// Flush and stop the async FTS worker first so no pending batch write
	// races with the index Close() calls below.
	// 先刷新并停止异步 FTS worker，避免待处理的批次写入与下方的索引 Close() 竞争。
	m.Shutdown()

	var lastErr error
	m.indexes.Range(func(key, value interface{}) bool {
		index := value.(bleve.Index)
		m.indexes.Delete(key)
		if err := index.Close(); err != nil {
			lastErr = err
		}
		return true
	})
	return lastErr
}

// DeleteIndex closes and physically removes index files for a specific vault
// DeleteIndex 关闭并物理删除特定仓库的索引文件
func (m *BleveManager) DeleteIndex(uid, vaultID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closeAndClean(uid, vaultID)
}

// closeAndClean helper to close index and remove directory (must hold mu lock or ensure exclusive execution)
// closeAndClean 关闭索引并清理物理目录的辅助函数
func (m *BleveManager) closeAndClean(uid, vaultID int64) error {
	key := fmt.Sprintf("%d_%d", uid, vaultID)
	if val, ok := m.indexes.Load(key); ok {
		index := val.(bleve.Index)
		_ = index.Close()
		m.indexes.Delete(key)
	}

	path := m.GetIndexPath(uid, vaultID)
	return os.RemoveAll(path)
}

// createIndexMapping configures default field analyzers and mapping
// createIndexMapping 配置默认字段分析器和映射规则
func (m *BleveManager) createIndexMapping() mapping.IndexMapping {
	indexMapping := bleve.NewIndexMapping()

	// Text field mapping using "cjk" analyzer
	// 文本字段映射，使用内置的 "cjk" 中日韩分词器
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = "cjk"
	textFieldMapping.Store = m.storeRaw
	textFieldMapping.Index = true

	// Keyword mapping for exact matching (e.g. action) and sorting (path_raw)
	// 关键字映射，用于精确匹配（如 action）和排序（path_raw）
	keywordMapping := bleve.NewTextFieldMapping()
	keywordMapping.Analyzer = "keyword"
	keywordMapping.Store = false
	keywordMapping.Index = true

	// Numeric field mapping for filters or sorting (ctime, mtime, rename)
	// 数值字段映射，用于过滤或排序（ctime, mtime, rename）
	numericMapping := bleve.NewNumericFieldMapping()
	numericMapping.Store = false
	numericMapping.Index = true

	// Define document mapping
	// 定义文档映射关系
	docMapping := bleve.NewDocumentMapping()

	// Add fields
	// 添加各字段映射规则
	docMapping.AddFieldMappingsAt("path", textFieldMapping)
	docMapping.AddFieldMappingsAt("path_raw", keywordMapping)
	docMapping.AddFieldMappingsAt("content", textFieldMapping)
	docMapping.AddFieldMappingsAt("action", keywordMapping)
	docMapping.AddFieldMappingsAt("rename", numericMapping)
	docMapping.AddFieldMappingsAt("ctime", numericMapping)
	docMapping.AddFieldMappingsAt("mtime", numericMapping)

	indexMapping.DefaultMapping = docMapping

	return indexMapping
}
