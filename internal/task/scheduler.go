package task

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/pkg/safe_close"
	"go.uber.org/zap"
)

// Task 定义任务接口
type Task interface {
	Name() string                  // 任务名称
	Run(ctx context.Context) error // 执行任务
	LoopInterval() time.Duration   // 执行间隔
	IsStartupRun() bool            // 是否立即执行一次
}

// Scheduler 任务调度器
type Scheduler struct {
	logger *zap.Logger
	tasks  []Task
	sc     *safe_close.SafeClose
}

// NewScheduler 创建任务调度器
func NewScheduler(logger *zap.Logger, sc *safe_close.SafeClose) *Scheduler {
	return &Scheduler{
		logger: logger,
		tasks:  make([]Task, 0),
		sc:     sc,
	}
}

// AddTask 添加任务
func (s *Scheduler) AddTask(task Task) {
	s.tasks = append(s.tasks, task)
}

// Start 启动所有任务
func (s *Scheduler) Start() {
	if len(s.tasks) == 0 {
		s.logger.Info("no tasks to schedule")
		return
	}

	s.logger.Info("tasks starting ", zap.Int("count", len(s.tasks)))

	for _, task := range s.tasks {
		s.startTask(task)
	}
}

// startTask 启动单个任务
func (s *Scheduler) startTask(task Task) {

	s.sc.Attach(func(done func(), closeSignal <-chan struct{}) {
		defer done()

		// 如果任务需要立即执行
		if task.IsStartupRun() {
			s.logger.Info("task running", zap.String("name", task.Name()), zap.Bool("startupRun", true))
			go func() {
				defer func() {
					if r := recover(); r != nil {
						s.logger.Error("task startupRun panic",
							zap.String("name", task.Name()),
							zap.Any("panic", r),
							zap.Stack("stack"))
					}
				}()
				if err := task.Run(context.Background()); err != nil {
					s.logger.Error("task running error",
						zap.String("name", task.Name()),
						zap.Bool("startupRun", true),
						zap.Error(err))
				}
			}()
		}

		if task.LoopInterval() <= 0 {
			return
		}

		ticker := time.NewTicker(task.LoopInterval())
		defer ticker.Stop()

		// 定时执行
		for {
			select {
			case <-ticker.C:
				func() {
					defer func() {
						if r := recover(); r != nil {
							s.logger.Error("task loopRun panic",
								zap.String("name", task.Name()),
								zap.Any("panic", r),
								zap.Stack("stack"))
						}
					}()
					s.logger.Info("task running", zap.String("name", task.Name()), zap.Bool("loopRun", true))
					if err := task.Run(context.Background()); err != nil {
						s.logger.Error("task running error",
							zap.String("name", task.Name()),
							zap.Bool("loopRun", true),
							zap.Error(err))
					}
				}()
			case <-closeSignal:
				s.logger.Info("task stopped", zap.String("name", task.Name()), zap.Bool("loopRun", true))
				return
			}
		}
	})
}
