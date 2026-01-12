// Package writequeue 提供 Per-User Write Queue 实现
// 用于串行化同一用户的 SQLite 写操作，解决 "database is locked" 问题
package writequeue

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// 错误定义
var (
	// ErrWriteQueueFull 当用户写队列已满时返回
	ErrWriteQueueFull = errors.New("write queue is full")
	// ErrWriteQueueClosed 当写队列管理器已关闭时返回
	ErrWriteQueueClosed = errors.New("write queue is closed")
	// ErrWriteTimeout 当写操作超时时返回
	ErrWriteTimeout = errors.New("write operation timeout")
)

// Config 写队列配置
type Config struct {
	// QueueCapacity 每用户队列容量，默认 100
	QueueCapacity int
	// WriteTimeout 写操作超时时间，默认 30 秒
	WriteTimeout time.Duration
	// IdleTimeout 空闲清理超时时间，默认 10 分钟
	IdleTimeout time.Duration
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		QueueCapacity: 100,
		WriteTimeout:  30 * time.Second,
		IdleTimeout:   10 * time.Minute,
	}
}

// writeOp 写操作
type writeOp struct {
	ctx    context.Context
	fn     func() error
	result chan error
}

// userWriteQueue 单用户写队列
type userWriteQueue struct {
	uid      int64
	ch       chan writeOp
	lastUsed atomic.Int64
	closed   atomic.Bool
	workerWg sync.WaitGroup

	// 用于通知 worker 停止
	stopCh chan struct{}
}


// Manager 管理所有用户的写队列
type Manager struct {
	config Config
	logger *zap.Logger

	queues sync.Map // map[int64]*userWriteQueue

	ctx    context.Context
	cancel context.CancelFunc

	mu     sync.RWMutex
	closed bool

	// 清理 goroutine 控制
	cleanupWg   sync.WaitGroup
	cleanupDone chan struct{}
}

// New 创建写队列管理器
// cfg: 配置，如果为 nil 则使用默认配置
// logger: zap 日志器，如果为 nil 则使用 nop logger
func New(cfg *Config, logger *zap.Logger) *Manager {
	if cfg == nil {
		defaultCfg := DefaultConfig()
		cfg = &defaultCfg
	}

	// 应用默认值
	if cfg.QueueCapacity <= 0 {
		cfg.QueueCapacity = 100
	}
	if cfg.WriteTimeout <= 0 {
		cfg.WriteTimeout = 30 * time.Second
	}
	if cfg.IdleTimeout <= 0 {
		cfg.IdleTimeout = 10 * time.Minute
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		config:      *cfg,
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
		closed:      false,
		cleanupDone: make(chan struct{}),
	}

	// 启动空闲队列清理 goroutine
	m.cleanupWg.Add(1)
	go m.cleanupIdleQueues()

	m.logger.Info("write queue manager started",
		zap.Int("queueCapacity", cfg.QueueCapacity),
		zap.Duration("writeTimeout", cfg.WriteTimeout),
		zap.Duration("idleTimeout", cfg.IdleTimeout))

	return m
}

// Execute 执行写操作
// 写操作会被串行化执行，同一用户的写操作按 FIFO 顺序处理
func (m *Manager) Execute(ctx context.Context, uid int64, fn func() error) error {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return ErrWriteQueueClosed
	}
	m.mu.RUnlock()

	// 获取或创建用户队列
	queue := m.getOrCreateQueue(uid)
	if queue == nil {
		return ErrWriteQueueClosed
	}

	// 创建写操作
	result := make(chan error, 1)
	op := writeOp{
		ctx:    ctx,
		fn:     fn,
		result: result,
	}

	// 尝试提交到队列
	select {
	case queue.ch <- op:
		// 操作已提交
	default:
		// 队列已满
		return ErrWriteQueueFull
	}

	// 等待结果或超时
	timeout := m.config.WriteTimeout
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining < timeout {
			timeout = remaining
		}
	}

	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(timeout):
		return ErrWriteTimeout
	case <-m.ctx.Done():
		return ErrWriteQueueClosed
	}
}


// getOrCreateQueue 获取或创建用户写队列（懒加载）
func (m *Manager) getOrCreateQueue(uid int64) *userWriteQueue {
	// 先尝试获取已存在的队列
	if v, ok := m.queues.Load(uid); ok {
		queue := v.(*userWriteQueue)
		if !queue.closed.Load() {
			queue.lastUsed.Store(time.Now().UnixNano())
			return queue
		}
	}

	// 检查是否已关闭
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return nil
	}
	m.mu.RUnlock()

	// 创建新队列
	queue := &userWriteQueue{
		uid:    uid,
		ch:     make(chan writeOp, m.config.QueueCapacity),
		stopCh: make(chan struct{}),
	}
	queue.lastUsed.Store(time.Now().UnixNano())

	// 使用 LoadOrStore 确保只有一个队列被创建
	actual, loaded := m.queues.LoadOrStore(uid, queue)
	if loaded {
		// 已存在队列，关闭新创建的
		close(queue.stopCh)
		existingQueue := actual.(*userWriteQueue)
		if !existingQueue.closed.Load() {
			existingQueue.lastUsed.Store(time.Now().UnixNano())
			return existingQueue
		}
		// 已存在的队列已关闭，需要替换
		m.queues.Store(uid, queue)
	}

	// 启动 worker goroutine（懒加载）
	queue.workerWg.Add(1)
	go m.worker(queue)

	m.logger.Debug("created write queue for user",
		zap.Int64("uid", uid),
		zap.Int("capacity", m.config.QueueCapacity))

	return queue
}

// worker 处理单用户写队列的 worker goroutine
func (m *Manager) worker(queue *userWriteQueue) {
	defer queue.workerWg.Done()
	defer func() {
		queue.closed.Store(true)
		m.logger.Debug("write queue worker stopped",
			zap.Int64("uid", queue.uid))
	}()

	for {
		select {
		case <-m.ctx.Done():
			// 管理器关闭，处理剩余操作
			m.drainQueue(queue)
			return
		case <-queue.stopCh:
			// 队列被停止
			m.drainQueue(queue)
			return
		case op, ok := <-queue.ch:
			if !ok {
				return
			}
			m.executeOp(queue, op)
		}
	}
}

// executeOp 执行单个写操作
func (m *Manager) executeOp(queue *userWriteQueue, op writeOp) {
	queue.lastUsed.Store(time.Now().UnixNano())

	// 检查 context 是否已取消
	select {
	case <-op.ctx.Done():
		op.result <- op.ctx.Err()
		return
	default:
	}

	// 执行写操作
	err := op.fn()

	// 发送结果
	select {
	case op.result <- err:
	default:
		// result channel 已关闭或已满
	}
}

// drainQueue 排空队列中的剩余操作
func (m *Manager) drainQueue(queue *userWriteQueue) {
	for {
		select {
		case op, ok := <-queue.ch:
			if !ok {
				return
			}
			m.executeOp(queue, op)
		default:
			return
		}
	}
}


// cleanupIdleQueues 定期清理空闲队列
func (m *Manager) cleanupIdleQueues() {
	defer m.cleanupWg.Done()

	ticker := time.NewTicker(m.config.IdleTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-m.cleanupDone:
			return
		case <-ticker.C:
			m.doCleanup()
		}
	}
}

// doCleanup 执行一次清理
func (m *Manager) doCleanup() {
	now := time.Now().UnixNano()
	idleThreshold := m.config.IdleTimeout.Nanoseconds()

	m.queues.Range(func(key, value interface{}) bool {
		uid := key.(int64)
		queue := value.(*userWriteQueue)

		// 检查是否空闲超时
		lastUsed := queue.lastUsed.Load()
		if now-lastUsed > idleThreshold {
			// 检查队列是否为空
			if len(queue.ch) == 0 && !queue.closed.Load() {
				m.logger.Debug("cleaning up idle write queue",
					zap.Int64("uid", uid),
					zap.Duration("idleTime", time.Duration(now-lastUsed)))

				// 标记关闭并通知 worker 停止
				queue.closed.Store(true)
				close(queue.stopCh)

				// 从 map 中删除
				m.queues.Delete(uid)
			}
		}
		return true
	})
}

// Shutdown 关闭写队列管理器，等待所有操作完成
// ctx 用于控制关闭超时
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil
	}
	m.closed = true
	m.mu.Unlock()

	m.logger.Info("write queue manager shutting down")

	// 停止清理 goroutine
	close(m.cleanupDone)

	// 等待所有队列的 worker 完成
	done := make(chan struct{})
	go func() {
		// 通知所有队列停止
		m.queues.Range(func(key, value interface{}) bool {
			queue := value.(*userWriteQueue)
			if !queue.closed.Load() {
				queue.closed.Store(true)
				select {
				case <-queue.stopCh:
					// 已关闭
				default:
					close(queue.stopCh)
				}
			}
			return true
		})

		// 等待所有 worker 完成
		m.queues.Range(func(key, value interface{}) bool {
			queue := value.(*userWriteQueue)
			queue.workerWg.Wait()
			return true
		})

		// 等待清理 goroutine 完成
		m.cleanupWg.Wait()

		close(done)
	}()

	select {
	case <-done:
		m.logger.Info("write queue manager shutdown completed")
		m.cancel()
		return nil
	case <-ctx.Done():
		m.logger.Warn("write queue manager shutdown timeout, forcing cancellation")
		m.cancel()
		return ctx.Err()
	}
}

// QueueCount 返回当前活跃队列数量
func (m *Manager) QueueCount() int {
	count := 0
	m.queues.Range(func(key, value interface{}) bool {
		queue := value.(*userWriteQueue)
		if !queue.closed.Load() {
			count++
		}
		return true
	})
	return count
}

// QueuedCount 返回指定用户队列中等待的操作数
func (m *Manager) QueuedCount(uid int64) int {
	if v, ok := m.queues.Load(uid); ok {
		queue := v.(*userWriteQueue)
		return len(queue.ch)
	}
	return 0
}

// IsClosed 返回管理器是否已关闭
func (m *Manager) IsClosed() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.closed
}

// Metrics 写队列管理器指标
type Metrics struct {
	QueueCapacity int
	ActiveQueues  int
	IsClosed      bool
}

// GetMetrics 获取当前指标
func (m *Manager) GetMetrics() Metrics {
	m.mu.RLock()
	closed := m.closed
	m.mu.RUnlock()

	return Metrics{
		QueueCapacity: m.config.QueueCapacity,
		ActiveQueues:  m.QueueCount(),
		IsClosed:      closed,
	}
}
