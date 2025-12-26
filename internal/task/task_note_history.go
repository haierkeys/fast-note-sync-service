package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"go.uber.org/zap"
)

// NoteHistoryTask 负责处理笔记历史记录的异步延时任务
type NoteHistoryTask struct {
	timers map[string]*time.Timer
	mu     sync.Mutex
}

// Name 返回任务名称
func (t *NoteHistoryTask) Name() string {
	return "NoteHistory"
}

// LoopInterval 返回执行间隔，此处为0，因为由 Run 内部循环控制
func (t *NoteHistoryTask) LoopInterval() time.Duration {
	return 0
}

// IsStartupRun 返回 true，使任务启动后立即开始执行 Run 循环
func (t *NoteHistoryTask) IsStartupRun() bool {
	return true
}

// Run 启动任务主循环，处理通道中的消息
func (t *NoteHistoryTask) Run(ctx context.Context) error {

	// 恢复中断的任务
	go t.resumeInterruptedTasks(ctx)

	for {
		select {
		case msg := <-service.NoteHistoryChannel:
			t.handleMsg(msg)
		case msg := <-service.NoteMigrateChannel:
			t.processMigrate(msg.OldNoteID, msg.NewNoteID, msg.UID)
		case <-ctx.Done():
			t.cleanup()
			global.Logger.Info("task log",
				zap.String("task", t.Name()),
				zap.String("type", "startupRun"),
				zap.String("event", "stopped"),
				zap.String("msg", "success"))
			return nil
		}
	}
}

// resumeInterruptedTasks 扫描并恢复中断的任务
func (t *NoteHistoryTask) resumeInterruptedTasks(ctx context.Context) {
	svc := service.NewBackground(ctx)
	uids, err := svc.GetAllUserUIDs()
	if err != nil {
		global.Logger.Error("task log",
			zap.String("task", t.Name()),
			zap.String("type", "startupRun"),
			zap.String("reason", "failed to get uids"),
			zap.String("msg", "failed"),
			zap.Error(err))
		return
	}

	for _, uid := range uids {
		notes, err := svc.NoteListNeedSnapshot(uid)
		if err != nil {
			global.Logger.Error("task log",
				zap.String("task", t.Name()),
				zap.String("type", "startupRun"),
				zap.String("msg", "failed"),
				zap.Int64("uid", uid),
				zap.Error(err))
			continue
		}
		for i, note := range notes {
			// 增加微小的错峰延迟，避免瞬间触发大量写事务
			delay := time.Duration(i%100) * 10 * time.Millisecond
			t.handleMsgWithDelay(service.NoteHistoryMsg{
				NoteID: note.ID,
				UID:    uid,
			}, delay)
		}
	}
}

// handleMsg 处理单条消息并管理定时器
func (t *NoteHistoryTask) handleMsg(msg service.NoteHistoryMsg) {
	t.handleMsgWithDelay(msg, 10*time.Second)
}

// handleMsgWithDelay 处理单条消息并设置自定义定时器延迟
func (t *NoteHistoryTask) handleMsgWithDelay(msg service.NoteHistoryMsg, baseDelay time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	key := fmt.Sprintf("%d_%d", msg.UID, msg.NoteID)

	// 如果已存在定时器，先停止它（重置倒计时）
	if timer, ok := t.timers[key]; ok {
		timer.Stop()
	}

	// 实际延迟为基础延迟 + 30秒固定业务延迟（或者根据需求调整）
	// 这里我们保持原有的 10 秒逻辑，但在 baseDelay 基础上叠加
	totalDelay := 10*time.Second + baseDelay

	// 创建定时器
	t.timers[key] = time.AfterFunc(totalDelay, func() {
		t.process(msg.NoteID, msg.UID, key)
	})
}

// process 执行实际的历史记录保存逻辑
func (t *NoteHistoryTask) process(noteID, uid int64, key string) {
	t.mu.Lock()
	delete(t.timers, key)
	t.mu.Unlock()

	// 使用背景上下文创建服务
	svc := service.NewBackground(context.Background())
	err := svc.NoteHistoryProcessDelay(noteID, uid)
	if err != nil {
		global.Logger.Error("task log",
			zap.String("task", "NoteHistory"),
			zap.String("type", "startupRun"),
			zap.Int64("noteID", noteID),
			zap.Int64("uid", uid),
			zap.String("reason", "process failed"),
			zap.String("msg", "failed"),
			zap.Error(err))
	} else {
		global.Logger.Debug("task log",
			zap.String("task", "NoteHistory"),
			zap.String("type", "startupRun"),
			zap.Int64("noteID", noteID),
			zap.Int64("uid", uid),
			zap.String("msg", "success"))
	}
}

// cleanup 在任务停止时清理所有定时器
func (t *NoteHistoryTask) cleanup() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, timer := range t.timers {
		timer.Stop()
	}
	t.timers = make(map[string]*time.Timer)
}

// processMigrate 执行历史记录迁移
func (t *NoteHistoryTask) processMigrate(oldNoteID, newNoteID, uid int64) {

	svc := service.NewBackground(context.Background())

	err := svc.NoteMigrate(oldNoteID, newNoteID, uid)
	if err != nil {
		global.Logger.Error("task log",
			zap.String("task", "NoteHistory"),
			zap.String("type", "startupRun"),
			zap.Int64("oldNoteID", oldNoteID),
			zap.Int64("newNoteID", newNoteID),
			zap.Int64("uid", uid),
			zap.String("reason", "NoteMigrate failed"),
			zap.String("msg", "failed"),
			zap.Error(err))
	} else {
		global.Logger.Debug("task log",
			zap.String("task", "NoteHistory"),
			zap.String("type", "startupRun"),
			zap.Int64("oldNoteID", oldNoteID),
			zap.Int64("newNoteID", newNoteID),
			zap.Int64("uid", uid),
			zap.String("event", "HistoryMigrate success"),
			zap.String("msg", "success"))
	}

	err = svc.NoteHistoryMigrate(oldNoteID, newNoteID, uid)
	if err != nil {
		global.Logger.Error("task log",
			zap.String("task", "NoteHistory"),
			zap.String("type", "startupRun"),
			zap.Int64("oldNoteID", oldNoteID),
			zap.Int64("newNoteID", newNoteID),
			zap.Int64("uid", uid),
			zap.String("reason", "processMigrate failed"),
			zap.String("msg", "failed"),
			zap.Error(err))
	} else {
		global.Logger.Debug("task log",
			zap.String("task", "NoteHistory"),
			zap.String("type", "startupRun"),
			zap.Int64("oldNoteID", oldNoteID),
			zap.Int64("newNoteID", newNoteID),
			zap.Int64("uid", uid),
			zap.String("event", "processMigrate success"),
			zap.String("msg", "success"))
	}
}

// NewNoteHistoryTask 创建一个新的笔记历史记录任务实例
func NewNoteHistoryTask() (Task, error) {
	return &NoteHistoryTask{
		timers: make(map[string]*time.Timer),
	}, nil
}

func init() {
	Register(NewNoteHistoryTask)
}
