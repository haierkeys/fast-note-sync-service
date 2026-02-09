package task

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"go.uber.org/zap"
)

// SyncFIDTask FID 同步任务
type SyncFIDTask struct {
	app    *app.App
	logger *zap.Logger
}

// Name 返回任务名称
func (t *SyncFIDTask) Name() string {
	return "SyncFID"
}

// LoopInterval 返回执行间隔（每天执行一次）
func (t *SyncFIDTask) LoopInterval() time.Duration {
	return 24 * time.Hour
}

// IsStartupRun 是否立即执行一次
func (t *SyncFIDTask) IsStartupRun() bool {
	return true
}

// Run 执行同步任务
func (t *SyncFIDTask) Run(ctx context.Context) error {
	t.logger.Info("starting SyncFID startup task")

	// 1. 获取所有用户 UID
	uids, err := t.app.UserRepo.GetAllUIDs(ctx)
	if err != nil {
		t.logger.Error("SyncFIDTask: failed to get all user UIDs", zap.Error(err))
		return err
	}

	for _, uid := range uids {
		// 2. 获取该用户的所有 Vault
		vaults, err := t.app.VaultService.List(ctx, uid)
		if err != nil {
			t.logger.Warn("SyncFIDTask: failed to list vaults for user", zap.Int64("uid", uid), zap.Error(err))
			continue
		}

		for _, vault := range vaults {
			// 3. 执行全量 FID 同步
			t.logger.Info("SyncFIDTask: syncing FID for vault", zap.Int64("uid", uid), zap.Int64("vaultID", vault.ID), zap.String("vaultName", vault.Name))

			// 先清理重复目录
			if err := t.app.FolderService.CleanDuplicateFolders(ctx, uid, vault.ID); err != nil {
				t.logger.Error("SyncFIDTask: failed to clean duplicate folders", zap.Int64("uid", uid), zap.Int64("vaultID", vault.ID), zap.Error(err))
			}

			if err := t.app.FolderService.SyncResourceFID(ctx, uid, vault.ID, nil, nil); err != nil {
				t.logger.Error("SyncFIDTask: failed to sync FID for vault",
					zap.Int64("uid", uid),
					zap.Int64("vaultID", vault.ID),
					zap.Error(err))
			}
		}
	}

	t.logger.Info("SyncFIDTask: startup sync completed")
	return nil
}

// NewSyncFIDTask 创建同步任务
func NewSyncFIDTask(appContainer *app.App) (Task, error) {
	return &SyncFIDTask{
		app:    appContainer,
		logger: appContainer.Logger(),
	}, nil
}

// init 自动注册同步任务
func init() {
	RegisterWithApp(func(appContainer *app.App) (Task, error) {
		return NewSyncFIDTask(appContainer)
	})
}
