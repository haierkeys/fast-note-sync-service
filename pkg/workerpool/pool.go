// Package workerpool 提供 goroutine 生命周期管理的 Worker Pool 实现
// 用于限制并发 goroutine 数量，防止资源泄漏
package workerpool

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
	// ErrWorkerPoolFull 当任务队列已满时返回
	ErrWorkerPoolFull = errors.New("worker pool queue is full")
	// ErrWorkerPoolClosed 当 Worker Pool 已关闭时返回
	ErrWorkerPoolClosed = errors.New("worker pool is closed")
	// ErrTaskCancelled 当任务被取消时返回
	ErrTaskCancelled = errors.New("task was cancelled")
	// ErrTaskTimeout 当任务执行超时时返回
	ErrTaskTimeout = errors.New("task execution timeout")
)

// Config Worker Pool 配置
type Config struct {
	// MaxWorkers 最大并发 worker 数量，默认 100
	MaxWorkers int
	// QueueSize 任务队列大小，默认 1000
	QueueSize int
	// WarningPercent 告警阈值百分比，默认 0.8 (80%)
	WarningPercent float64
	// IdleTimeout 空闲超时时间，默认 5 分钟
	IdleTimeout time.Duration
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		MaxWorkers:     100,
		QueueSize:      1000,
		WarningPercent: 0.8,
		IdleTimeout:    5 * time.Minute,
	}
}

// taskWrapper 任务包装器
type taskWrapper struct {
	ctx  context.Context
	fn   func(context.Context) error
	done chan error
}

// Pool 管理 goroutine 生命周期的 Worker Pool
type Pool struct {
	config Config
	logger *zap.Logger

	taskCh   chan taskWrapper
	workerWg sync.WaitGroup

	activeCount atomic.Int64

	ctx    context.Context
	cancel context.CancelFunc

	mu     sync.RWMutex
	closed bool
}

// New 创建新的 Worker Pool
// cfg: 配置，如果为 nil 则使用默认配置
// logger: zap 日志器，如果为 nil 则使用 nop logger
func New(cfg *Config, logger *zap.Logger) *Pool {
	if cfg == nil {
		defaultCfg := DefaultConfig()
		cfg = &defaultCfg
	}

	// 应用默认值
	if cfg.MaxWorkers <= 0 {
		cfg.MaxWorkers = 100
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 1000
	}
	if cfg.WarningPercent <= 0 || cfg.WarningPercent > 1 {
		cfg.WarningPercent = 0.8
	}
	if cfg.IdleTimeout <= 0 {
		cfg.IdleTimeout = 5 * time.Minute
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	ctx, cancel := context.WithCancel(context.Background())

	p := &Pool{
		config: *cfg,
		logger: logger,
		taskCh: make(chan taskWrapper, cfg.QueueSize),
		ctx:    ctx,
		cancel: cancel,
		closed: false,
	}

	// 启动 worker goroutines
	for i := 0; i < cfg.MaxWorkers; i++ {
		p.workerWg.Add(1)
		go p.worker(i)
	}

	p.logger.Info("worker pool started",
		zap.Int("maxWorkers", cfg.MaxWorkers),
		zap.Int("queueSize", cfg.QueueSize),
		zap.Float64("warningPercent", cfg.WarningPercent))

	return p
}

// worker 工作协程，从任务通道获取任务并执行
func (p *Pool) worker(id int) {
	defer p.workerWg.Done()

	for {
		select {
		case <-p.ctx.Done():
			return
		case task, ok := <-p.taskCh:
			if !ok {
				return
			}
			p.executeTask(task)
		}
	}
}

// executeTask 执行单个任务
func (p *Pool) executeTask(task taskWrapper) {
	p.activeCount.Add(1)
	defer p.activeCount.Add(-1)

	// 检查告警阈值
	p.checkWarningThreshold()

	// 执行任务
	var err error
	select {
	case <-task.ctx.Done():
		err = ErrTaskCancelled
	default:
		err = task.fn(task.ctx)
	}

	// 发送结果
	if task.done != nil {
		select {
		case task.done <- err:
		default:
			// done channel 已关闭或已满，忽略
		}
	}
}

// checkWarningThreshold 检查是否超过告警阈值
func (p *Pool) checkWarningThreshold() {
	active := p.activeCount.Load()
	threshold := int64(float64(p.config.MaxWorkers) * p.config.WarningPercent)

	if active >= threshold {
		p.logger.Warn("worker pool approaching capacity",
			zap.Int64("activeCount", active),
			zap.Int("maxWorkers", p.config.MaxWorkers),
			zap.Float64("usagePercent", float64(active)/float64(p.config.MaxWorkers)*100))
	}
}

// Submit 提交任务并等待完成
// 返回任务执行结果或错误（池满/已关闭）
func (p *Pool) Submit(ctx context.Context, fn func(context.Context) error) error {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return ErrWorkerPoolClosed
	}
	p.mu.RUnlock()

	done := make(chan error, 1)
	task := taskWrapper{
		ctx:  ctx,
		fn:   fn,
		done: done,
	}

	// 尝试提交任务
	select {
	case p.taskCh <- task:
		// 任务已提交，等待结果
	default:
		// 队列已满
		return ErrWorkerPoolFull
	}

	// 等待任务完成或 context 取消
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-p.ctx.Done():
		return ErrWorkerPoolClosed
	}
}

// SubmitAsync 异步提交任务（不等待结果）
// 返回错误如果池已满或已关闭
func (p *Pool) SubmitAsync(ctx context.Context, fn func(context.Context) error) error {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return ErrWorkerPoolClosed
	}
	p.mu.RUnlock()

	task := taskWrapper{
		ctx:  ctx,
		fn:   fn,
		done: nil, // 异步任务不需要 done channel
	}

	// 尝试提交任务
	select {
	case p.taskCh <- task:
		return nil
	default:
		return ErrWorkerPoolFull
	}
}

// ActiveCount 返回当前活跃任务数
func (p *Pool) ActiveCount() int64 {
	return p.activeCount.Load()
}

// QueuedCount 返回当前队列中等待的任务数
func (p *Pool) QueuedCount() int {
	return len(p.taskCh)
}

// IsClosed 返回 Worker Pool 是否已关闭
func (p *Pool) IsClosed() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.closed
}

// Shutdown 关闭 Worker Pool，等待所有任务完成
// ctx 用于控制关闭超时
func (p *Pool) Shutdown(ctx context.Context) error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	p.mu.Unlock()

	p.logger.Info("worker pool shutting down",
		zap.Int64("activeCount", p.activeCount.Load()),
		zap.Int("queuedCount", len(p.taskCh)))

	// 关闭任务通道，不再接受新任务
	close(p.taskCh)

	// 等待所有 worker 完成
	done := make(chan struct{})
	go func() {
		p.workerWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		p.logger.Info("worker pool shutdown completed")
		return nil
	case <-ctx.Done():
		// 超时，强制取消所有任务
		p.cancel()
		p.logger.Warn("worker pool shutdown timeout, forcing cancellation")
		return ctx.Err()
	}
}

// Metrics 返回 Worker Pool 的指标
type Metrics struct {
	MaxWorkers    int
	ActiveCount   int64
	QueuedCount   int
	QueueCapacity int
	IsClosed      bool
}

// GetMetrics 获取当前指标
func (p *Pool) GetMetrics() Metrics {
	p.mu.RLock()
	closed := p.closed
	p.mu.RUnlock()

	return Metrics{
		MaxWorkers:    p.config.MaxWorkers,
		ActiveCount:   p.activeCount.Load(),
		QueuedCount:   len(p.taskCh),
		QueueCapacity: p.config.QueueSize,
		IsClosed:      closed,
	}
}
