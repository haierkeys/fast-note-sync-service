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

func init() {
	Register(NewNoteHistoryTask)
}

// NewNoteHistoryTask 创建一个新的笔记历史记录任务实例
func NewNoteHistoryTask() (Task, error) {
	return &NoteHistoryTask{
		timers: make(map[string]*time.Timer),
	}, nil
}

// Name 返回任务名称
func (t *NoteHistoryTask) Name() string {
	return "NoteHistoryTask"
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
	global.Logger.Info("NoteHistoryTask started")

	// 恢复中断的任务
	go t.resumeInterruptedTasks(ctx)

	for {
		select {
		case msg := <-service.NoteHistoryChannel:
			t.handleMsg(msg)
		case <-ctx.Done():
			t.cleanup()
			global.Logger.Info("NoteHistoryTask stopped")
			return nil
		}
	}
}

// resumeInterruptedTasks 扫描并恢复中断的任务
func (t *NoteHistoryTask) resumeInterruptedTasks(ctx context.Context) {
	svc := service.NewBackground(ctx)
	uids, err := svc.GetAllUserUIDs()
	if err != nil {
		global.Logger.Error("NoteHistoryTask resume failed to get uids", zap.Error(err))
		return
	}

	for _, uid := range uids {
		notes, err := svc.NoteListNeedSnapshot(uid)
		if err != nil {
			global.Logger.Error("NoteHistoryTask resume failed to get notes", zap.Int64("uid", uid), zap.Error(err))
			continue
		}
		for _, note := range notes {
			t.handleMsg(service.NoteHistoryMsg{
				NoteID: note.ID,
				UID:    uid,
			})
		}
	}
}

// handleMsg 处理单条消息并管理定时器
func (t *NoteHistoryTask) handleMsg(msg service.NoteHistoryMsg) {
	t.mu.Lock()
	defer t.mu.Unlock()

	key := fmt.Sprintf("%d_%d", msg.UID, msg.NoteID)

	// 如果已存在定时器，先停止它（重置倒计时）
	if timer, ok := t.timers[key]; ok {
		timer.Stop()
	}

	// 创建新的30秒定时器
	t.timers[key] = time.AfterFunc(30*time.Second, func() {
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
		global.Logger.Error("NoteHistoryTask process failed",
			zap.Int64("noteID", noteID),
			zap.Int64("uid", uid),
			zap.Error(err))
	} else {
		global.Logger.Debug("NoteHistoryTask process success",
			zap.Int64("noteID", noteID),
			zap.Int64("uid", uid))
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
