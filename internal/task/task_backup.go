package task

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"go.uber.org/zap"
)

// BackupTask handles scheduled backups
type BackupTask struct {
	app    *app.App
	logger *zap.Logger
}

// Name returns the task name
func (t *BackupTask) Name() string {
	return "BackupScheduled"
}

// LoopInterval returns the execution interval (every minute)
func (t *BackupTask) LoopInterval() time.Duration {
	return 1 * time.Minute
}

// IsStartupRun returns whether to run on startup
func (t *BackupTask) IsStartupRun() bool {
	return true
}

// Run executes the backup processing
func (t *BackupTask) Run(ctx context.Context) error {
	if t.app.BackupService == nil {
		return nil
	}
	return t.app.BackupService.ExecuteTaskBackups(ctx)
}

// NewBackupTask creates a new BackupTask instance
func NewBackupTask(appContainer *app.App) (Task, error) {
	return &BackupTask{
		app:    appContainer,
		logger: appContainer.Logger(),
	}, nil
}

// init registers the backup task
func init() {
	RegisterWithApp(func(appContainer *app.App) (Task, error) {
		return NewBackupTask(appContainer)
	})
}
