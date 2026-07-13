// Package service implements the business logic layer
// Package service 实现业务逻辑层
package service

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/pkg/safego"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"go.uber.org/zap"
)

// syncLogChannelBuffer is the size of the bounded channel Log() pushes into; once full,
// new entries are dropped (with a warning) rather than blocking the caller or growing
// unbounded, since audit-log entries are allowed to degrade under load.
// syncLogChannelBuffer 是 Log() 写入的有界 channel 容量；写满后新条目会被丢弃（并记录
// warning），而不是阻塞调用方或无限增长，因为审计日志允许在高负载下降级。
const syncLogChannelBuffer = 4096

// syncLogBatchMaxSize is the number of buffered entries that triggers an immediate flush.
// syncLogBatchMaxSize 触发立即 flush 的缓冲条目数量。
const syncLogBatchMaxSize = 100

// syncLogBatchFlushInterval is the maximum time buffered entries wait before being flushed.
// syncLogBatchFlushInterval 缓冲条目在被 flush 前等待的最长时间。
const syncLogBatchFlushInterval = 500 * time.Millisecond

// SyncLogService defines the sync log business service interface
// SyncLogService 定义同步日志业务服务接口
type SyncLogService interface {
	// Log asynchronously records a sync log entry, does not block the caller
	// Log 异步记录一条同步日志，不阻塞调用方
	Log(
		uid int64,
		vaultID int64,
		logType domain.SyncLogType,
		action domain.SyncLogAction,
		changedFields string, // e.g. "content,mtime" / "mtime" / "path" / "" // 如 "content,mtime" / "mtime" / "path" / ""
		path string,
		pathHash string,
		clientType string,
		clientName string,
		clientVersion string,
		size int64,
	)

	// List retrieves sync logs with pagination
	// List 分页查询同步日志
	List(ctx context.Context, uid int64, vaultID int64, logType, action string, page, pageSize int) ([]*dto.SyncLogDTO, int64, error)

	// CleanupByTime removes sync logs older than the given cutoff time for all users
	// CleanupByTime 清理所有用户在指定截止时间之前的同步日志
	CleanupByTime(ctx context.Context, cutoffTime int64) error

	// Shutdown stops the background batch worker, flushing any buffered entries first.
	// Shutdown 停止后台批处理 worker，退出前先 flush 所有缓冲中的条目。
	Shutdown(ctx context.Context) error
}

// syncLogQueueItem pairs a buffered entry with the uid it must be written to, since each
// user's sync logs live in a separate per-user database.
// syncLogQueueItem 把缓冲条目和它所属的 uid 绑定，因为每个用户的同步日志存在独立的分库中。
type syncLogQueueItem struct {
	uid   int64
	entry *domain.SyncLog
}

// syncLogService implements SyncLogService
// syncLogService 实现 SyncLogService 接口
type syncLogService struct {
	repo   domain.SyncLogRepository // Sync log repository // 同步日志仓储
	logger *zap.Logger
	ch     chan syncLogQueueItem
	stopCh chan struct{}
	doneCh chan struct{}
}

// NewSyncLogService creates a new SyncLogService instance and starts its background
// batch-flush worker.
// NewSyncLogService 创建 SyncLogService 实例，并启动其后台批量 flush worker。
func NewSyncLogService(repo domain.SyncLogRepository, logger *zap.Logger) SyncLogService {
	if logger == nil {
		logger = zap.L()
	}
	s := &syncLogService{
		repo:   repo,
		logger: logger,
		ch:     make(chan syncLogQueueItem, syncLogChannelBuffer),
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
	safego.Go(logger, s.runBatchWorker)
	return s
}

// Log enqueues a sync log entry for asynchronous, batched persistence; it never blocks the
// caller. If the internal channel is full, the entry is dropped and a warning is logged
// (acceptable degradation for an audit log under heavy write load).
// Log 将一条同步日志条目加入队列，异步批量落库；不会阻塞调用方。若内部 channel 已满，
// 条目会被丢弃并记录 warning（审计日志在高写入负载下允许这种降级）。
func (s *syncLogService) Log(
	uid int64,
	vaultID int64,
	logType domain.SyncLogType,
	action domain.SyncLogAction,
	changedFields string,
	path string,
	pathHash string,
	clientType string,
	clientName string,
	clientVersion string,
	size int64,
) {
	entry := &domain.SyncLog{
		UID:           uid,
		VaultID:       vaultID,
		Type:          logType,
		Action:        action,
		ChangedFields: changedFields,
		Path:          path,
		PathHash:      pathHash,
		Size:          size,
		ClientType:    clientType,
		ClientName:    clientName,
		ClientVersion: clientVersion,
		Status:        1, // success // 成功
		CreatedAt:     timex.Now(),
	}

	select {
	case s.ch <- syncLogQueueItem{uid: uid, entry: entry}:
	default:
		s.logger.Warn("SyncLogService.Log: queue full, dropping sync log entry",
			zap.Int64("uid", uid),
			zap.Int64("vaultID", vaultID),
			zap.String("type", string(logType)),
			zap.String("action", string(action)),
			zap.String("path", path),
		)
	}
}

// runBatchWorker drains the queue, grouping entries per uid (each user's logs live in a
// separate database), and flushes each group with a single batched write once syncLogBatchMaxSize
// entries have accumulated or syncLogBatchFlushInterval has elapsed, whichever comes first.
// runBatchWorker 消费队列，按 uid 分组（每个用户的日志存在独立数据库中），
// 累计到 syncLogBatchMaxSize 条或每隔 syncLogBatchFlushInterval（以先到者为准）批量 flush 一次。
func (s *syncLogService) runBatchWorker() {
	defer close(s.doneCh)

	buf := make(map[int64][]*domain.SyncLog)
	count := 0
	ticker := time.NewTicker(syncLogBatchFlushInterval)
	defer ticker.Stop()

	flush := func() {
		if count == 0 {
			return
		}
		for uid, entries := range buf {
			if err := s.repo.CreateBatch(context.Background(), entries, uid); err != nil {
				s.logger.Warn("SyncLogService: failed to batch create sync logs",
					zap.Int64("uid", uid),
					zap.Int("count", len(entries)),
					zap.Error(err),
				)
			}
		}
		buf = make(map[int64][]*domain.SyncLog)
		count = 0
	}

	for {
		select {
		case item := <-s.ch:
			buf[item.uid] = append(buf[item.uid], item.entry)
			count++
			if count >= syncLogBatchMaxSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-s.stopCh:
			// Drain whatever is already queued (best-effort, non-blocking) before the final flush.
			// 尽力（非阻塞）排空已入队的条目后再做最后一次 flush。
			for drained := true; drained; {
				select {
				case item := <-s.ch:
					buf[item.uid] = append(buf[item.uid], item.entry)
					count++
				default:
					drained = false
				}
			}
			flush()
			return
		}
	}
}

// Shutdown stops the background batch worker and waits for it to flush any buffered entries,
// or until ctx is done.
// Shutdown 停止后台批处理 worker，等待其 flush 完所有缓冲条目，或等到 ctx 结束。
func (s *syncLogService) Shutdown(ctx context.Context) error {
	select {
	case <-s.stopCh:
		// Already shutting down // 已在关闭中
	default:
		close(s.stopCh)
	}

	select {
	case <-s.doneCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// List retrieves sync logs with optional filters and pagination
// List 按条件分页查询同步日志
func (s *syncLogService) List(ctx context.Context, uid int64, vaultID int64, logType, action string, page, pageSize int) ([]*dto.SyncLogDTO, int64, error) {
	logs, total, err := s.repo.List(ctx, uid, logType, action, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]*dto.SyncLogDTO, 0, len(logs))
	for _, l := range logs {
		if vaultID > 0 && l.VaultID != vaultID {
			continue
		}
		result = append(result, s.domainToDTO(l))
	}
	return result, total, nil
}

// CleanupByTime removes sync logs older than the given cutoff time for all users
// CleanupByTime 清理所有用户在指定截止时间之前的同步日志
func (s *syncLogService) CleanupByTime(ctx context.Context, cutoffTime int64) error {
	return s.repo.CleanupByTimeAll(ctx, cutoffTime)
}

// domainToDTO converts domain SyncLog to DTO
// domainToDTO 将领域模型转换为 DTO
func (s *syncLogService) domainToDTO(l *domain.SyncLog) *dto.SyncLogDTO {
	return &dto.SyncLogDTO{
		ID:            l.ID,
		VaultID:       l.VaultID,
		Type:          string(l.Type),
		Action:        string(l.Action),
		ChangedFields: l.ChangedFields,
		Path:          l.Path,
		PathHash:      l.PathHash,
		Size:          l.Size,
		ClientName:    l.ClientName,
		ClientType:    l.ClientType,
		ClientVersion: l.ClientVersion,
		Status:        l.Status,
		Message:       l.Message,
		CreatedAt:     l.CreatedAt,
	}
}

// Ensure syncLogService implements SyncLogService
// 确保 syncLogService 实现了 SyncLogService 接口
var _ SyncLogService = (*syncLogService)(nil)
