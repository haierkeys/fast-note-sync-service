package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/config"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/storage"
	pkgstorage "github.com/haierkeys/fast-note-sync-service/pkg/storage"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

var errNoUpdates = errors.New("no updates found")

// BackupService defines the business service interface for Backup
// 定义备份业务服务接口
type BackupService interface {
	GetConfigs(ctx context.Context, uid int64) ([]*dto.BackupConfigDTO, error)
	DeleteConfig(ctx context.Context, uid int64, configID int64) error
	UpdateConfig(ctx context.Context, uid int64, req *dto.BackupConfigRequest) (*dto.BackupConfigDTO, error)
	ListHistory(ctx context.Context, uid int64, configID int64, pager *app.Pager) ([]*dto.BackupHistoryDTO, int64, error)
	ExecuteUserBackup(ctx context.Context, uid int64, configID int64) error
	ExecuteTaskBackups(ctx context.Context) error
	NotifyUpdated(uid int64)
	Shutdown(ctx context.Context) error
}

type backupService struct {
	backupRepo     domain.BackupRepository
	noteRepo       domain.NoteRepository
	folderRepo     domain.FolderRepository
	fileRepo       domain.FileRepository
	vaultRepo      domain.VaultRepository
	storageService StorageService
	storageConfig  *config.StorageConfig
	logger         *zap.Logger
	syncTimers     map[int64]*time.Timer
	timerMu        sync.Mutex
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	pendingSyncs   sync.Map                     // key: uid (int64), value: bool
	runningTasks   map[int64]context.CancelFunc // key: configID
	runningMu      sync.Mutex
}

// NewBackupService creates BackupService instance
// 创建 BackupService 实例
func NewBackupService(
	backupRepo domain.BackupRepository,
	noteRepo domain.NoteRepository,
	folderRepo domain.FolderRepository,
	fileRepo domain.FileRepository,
	vaultRepo domain.VaultRepository,
	storageService StorageService,
	storageConfig *config.StorageConfig,
	logger *zap.Logger,
) BackupService {
	ctx, cancel := context.WithCancel(context.Background())
	return &backupService{
		backupRepo:     backupRepo,
		noteRepo:       noteRepo,
		folderRepo:     folderRepo,
		fileRepo:       fileRepo,
		vaultRepo:      vaultRepo,
		storageService: storageService,
		storageConfig:  storageConfig,
		logger:         logger,
		syncTimers:     make(map[int64]*time.Timer),
		runningTasks:   make(map[int64]context.CancelFunc),
		ctx:            ctx,
		cancel:         cancel,
	}
}

// GetConfigs Get user's backup configurations
// 获取用户的备份配置列表
func (s *backupService) GetConfigs(ctx context.Context, uid int64) ([]*dto.BackupConfigDTO, error) {
	configs, err := s.backupRepo.ListConfigs(ctx, uid)
	if err != nil {
		return nil, err
	}
	var results []*dto.BackupConfigDTO
	for _, c := range configs {
		results = append(results, s.configToDTO(ctx, c))
	}
	return results, nil
}

// UpdateConfig Update or create backup configuration
// 更新或创建备份配置
func (s *backupService) UpdateConfig(ctx context.Context, uid int64, req *dto.BackupConfigRequest) (*dto.BackupConfigDTO, error) {
	var vaultID int64
	if req.Vault != "" && req.Vault != "0" && req.Vault != "all" {
		v, err := s.vaultRepo.GetByName(ctx, req.Vault, uid)
		if err != nil {
			return nil, err
		}
		if v == nil {
			return nil, code.ErrorVaultNotFound
		}
		vaultID = v.ID
	}

	// Validate Storage IDs
	var storageIds []int64
	if err := json.Unmarshal([]byte(req.StorageIds), &storageIds); err != nil {
		return nil, code.ErrorBackupStorageIDInvalid
	}
	for _, sid := range storageIds {
		if _, err := s.storageService.Get(ctx, uid, sid); err != nil {
			return nil, code.ErrorStorageNotFound
		}
	}

	config := &domain.BackupConfig{
		ID:             req.ID,
		UID:            uid,
		VaultID:        vaultID,
		Type:           req.Type,
		StorageIds:     req.StorageIds,
		IsEnabled:      req.IsEnabled,
		CronStrategy:   req.CronStrategy,
		CronExpression: req.CronExpression,
		RetentionDays:  req.RetentionDays,
	}

	// Preserve state fields if updating existing config
	if req.ID > 0 {
		if old, err := s.backupRepo.GetByID(ctx, req.ID, uid); err == nil && old != nil {
			config.LastRunTime = old.LastRunTime
			config.LastStatus = old.LastStatus
			config.LastMessage = old.LastMessage
		}
	}

	// Calculate NextRunTime based on Cron Strategy
	s.calculateNextRunTime(config)

	updated, err := s.backupRepo.SaveConfig(ctx, config, uid)
	if err != nil {
		return nil, err
	}

	// Trigger sync check immediately if enabled and type is sync
	if updated.IsEnabled && (updated.Type == "sync") {

		fmt.Println("Trigger sync check immediately if enabled and type is sync")
		s.pendingSyncs.Store(uid, true)
	}

	return s.configToDTO(ctx, updated), nil
}

// DeleteConfig Deletes a backup configuration
// 删除备份配置
func (s *backupService) DeleteConfig(ctx context.Context, uid int64, configID int64) error {
	// First check if config exists and belongs to user
	config, err := s.backupRepo.GetByID(ctx, configID, uid)
	if err != nil {
		return err
	}
	if config == nil {
		return code.ErrorBackupConfigNotFound
	}

	return s.backupRepo.DeleteConfig(ctx, configID, uid)
}

// ListHistory List backup history with pagination
// 分页查询备份历史记录
func (s *backupService) ListHistory(ctx context.Context, uid int64, configID int64, pager *app.Pager) ([]*dto.BackupHistoryDTO, int64, error) {
	histories, count, err := s.backupRepo.ListHistory(ctx, uid, configID, pager.Page, pager.PageSize)
	if err != nil {
		return nil, 0, err
	}

	var results []*dto.BackupHistoryDTO
	for _, h := range histories {
		results = append(results, s.historyToDTO(h))
	}
	return results, count, nil
}

func (s *backupService) configToDTO(ctx context.Context, d *domain.BackupConfig) *dto.BackupConfigDTO {
	if d == nil {
		return nil
	}
	vaultName := "all"
	if d.VaultID > 0 {
		if v, err := s.vaultRepo.GetByID(ctx, d.VaultID, d.UID); err == nil && v != nil {
			vaultName = v.Name
		}
	}
	return &dto.BackupConfigDTO{
		ID:             d.ID,
		UID:            d.UID,
		Vault:          vaultName,
		Type:           d.Type,
		StorageIds:     d.StorageIds,
		IsEnabled:      d.IsEnabled,
		CronStrategy:   d.CronStrategy,
		CronExpression: d.CronExpression,
		RetentionDays:  d.RetentionDays,
		LastRunTime:    timex.Time(d.LastRunTime),
		NextRunTime:    timex.Time(d.NextRunTime),
		LastStatus:     d.LastStatus,
		LastMessage:    d.LastMessage,
		CreatedAt:      timex.Time(d.CreatedAt),
		UpdatedAt:      timex.Time(d.UpdatedAt),
	}
}

func (s *backupService) historyToDTO(d *domain.BackupHistory) *dto.BackupHistoryDTO {
	if d == nil {
		return nil
	}
	return &dto.BackupHistoryDTO{
		ID:        d.ID,
		UID:       d.UID,
		ConfigID:  d.ConfigID,
		StorageID: d.StorageID,
		Type:      d.Type,
		StartTime: timex.Time(d.StartTime),
		EndTime:   timex.Time(d.EndTime),
		Status:    d.Status,
		FileSize:  d.FileSize,
		FileCount: d.FileCount,
		Message:   d.Message,
		FilePath:  d.FilePath,
		CreatedAt: timex.Time(d.CreatedAt),
		UpdatedAt: timex.Time(d.UpdatedAt),
	}
}

// ExecuteUserBackup Manually execute user backup task
// 手动执行用户备份任务
// ExecuteUserBackup Manually execute user backup task
// 手动执行用户备份任务
func (s *backupService) ExecuteUserBackup(ctx context.Context, uid int64, configID int64) error {
	// If configID is specified, execute specific task
	if configID <= 0 {
		return code.ErrorBackupExecuteIDReq
	}

	config, err := s.backupRepo.GetByID(ctx, configID, uid)
	if err != nil {
		return err
	}
	if config == nil {
		return code.ErrorBackupConfigNotFound
	}
	if !config.IsEnabled {
		return code.ErrorBackupConfigDisabled
	}
	// 记录错误
	if err := s.handleBackupSync(ctx, config, true); err != nil {
		// Service shutdown errors bypass finishTask and are not persisted to history
		if s.ctx.Err() != nil {
			return err
		}
		s.logger.Warn("Manual backup completed with errors",
			zap.Int64("uid", uid),
			zap.Int64("configID", configID),
			zap.Error(err),
		)
	}
	return nil
}

// ExecuteTaskBackups Poll and process all scheduled backup tasks
// 轮询处理所有待执行的定时备份任务
func (s *backupService) ExecuteTaskBackups(ctx context.Context) error {
	configs, err := s.backupRepo.ListEnabledConfigs(ctx)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, config := range configs {
		if !config.IsEnabled {
			continue
		}

		// Check if user has pending changes
		_, pending := s.pendingSyncs.LoadAndDelete(config.UID)

		isScheduled := config.NextRunTime.Before(now)
		shouldTrigger := false

		if isScheduled {
			shouldTrigger = true
		} else if pending && config.Type == "sync" {
			// Only "sync" type tasks are allowed to be triggered directly by changes (debounced)
			shouldTrigger = true
		}

		if shouldTrigger {
			s.logger.Info("Triggering backup task",
				zap.Int64("uid", config.UID),
				zap.String("type", config.Type),
				zap.Bool("isScheduled", isScheduled),
				zap.Bool("isPending", pending),
			)
			go func(cfg *domain.BackupConfig, p bool) {
				// Use service context to support graceful shutdown
				if err := s.handleBackupSync(s.ctx, cfg, p); err != nil {
					s.logger.Error("Backup execution failed", zap.Int64("uid", cfg.UID), zap.Error(err))
				}
			}(config, pending)
		}
	}

	return nil
}

// calculateNextRunTime Calculate next run time based on Cron strategy
// 根据 Cron 策略计算下次运行时间
func (s *backupService) calculateNextRunTime(config *domain.BackupConfig) {
	if !config.IsEnabled {
		return
	}

	if config.Type == "sync" {
		config.NextRunTime = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
		return
	}

	expr := ""
	switch config.CronStrategy {
	case "daily":
		expr = "0 0 * * *" // Midnight daily
	case "weekly":
		expr = "0 0 * * 0" // Midnight Sunday
	case "monthly":
		expr = "0 0 1 * *" // Midnight 1st of month
	case "custom":
		expr = config.CronExpression
	}

	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(expr)
	if err != nil {
		s.logger.Error("Failed to parse cron expression", zap.String("expr", expr), zap.Error(err))
		return
	}

	config.NextRunTime = schedule.Next(time.Now())
}

// handleBackupSync Core entry point for performing backup/sync
// 执行备份/同步的核心入口
func (s *backupService) handleBackupSync(ctx context.Context, config *domain.BackupConfig, isPending bool) error {
	uid := config.UID
	configID := config.ID

	// 1. 并发冲突处理策略
	s.runningMu.Lock()
	if cancel, running := s.runningTasks[configID]; running {
		if config.Type == "sync" {
			// 同步任务策略：取消旧任务，执行新任务
			s.logger.Info("Cancelling existing sync task to start a newer one", zap.Int64("uid", uid), zap.Int64("configID", configID))
			cancel()
			delete(s.runningTasks, configID)
		} else {
			// 全量/增量备份策略：保留旧任务，忽略新任务
			s.runningMu.Unlock()
			s.logger.Info("Backup task already running, skipping this trigger", zap.Int64("uid", uid), zap.Int64("configID", configID), zap.String("type", config.Type))
			return nil
		}
	}

	// 创建带取消功能的 context
	taskCtx, taskCancel := context.WithCancel(ctx)
	s.runningTasks[configID] = taskCancel
	s.runningMu.Unlock()

	// 任务结束时的清理
	defer func() {
		s.runningMu.Lock()
		if _, ok := s.runningTasks[configID]; ok {
			// 确保清理当前的 cancel 记录
			delete(s.runningTasks, configID)
		}
		s.runningMu.Unlock()
		taskCancel() // 释放资源
	}()

	s.wg.Add(1)
	defer s.wg.Done()

	// Check if context is already done
	select {
	case <-taskCtx.Done():
		return taskCtx.Err()
	default:
	}

	startTime := time.Now()
	prevRunTime := config.LastRunTime // 记录本次执行前的上一次执行时间

	shouldRun := false
	switch config.Type {
	case "full":
		shouldRun = true
	case "incremental", "sync":
		// Exception: If it's the first run (prevRunTime is zero), we must execute to create a base backup.
		if isPending || prevRunTime.IsZero() {
			shouldRun = true
		}
	default:
		return code.ErrorBackupTypeUnknown
	}

	if !shouldRun {
		s.logger.Info("Skipping backup: no pending changes", zap.Int64("uid", uid), zap.String("type", config.Type))
		s.recordNoUpdateHistory(taskCtx, config, startTime)
		return s.finishTask(taskCtx, config, errNoUpdates, 0, 0, startTime)
	}

	s.logger.Info("handleBackupSync start", zap.Int64("uid", uid), zap.String("type", config.Type))

	// 2. 设置运行状态 (Running)
	config.LastStatus = domain.BackupStatusRunning
	s.backupRepo.SaveConfig(taskCtx, config, uid)

	// 3. 准备临时工作目录
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("backup_%d_", uid))
	if err != nil {
		return s.finishTask(taskCtx, config, err, 0, 0, startTime)
	}
	defer os.RemoveAll(tempDir)

	var fileCount, fileSize int64
	var backupErr error

	// 4. 执行核心逻辑
	switch config.Type {
	case "full":
		fileCount, fileSize, backupErr = s.runArchive(taskCtx, config, tempDir, startTime, prevRunTime)
	case "incremental":
		fileCount, fileSize, backupErr = s.runArchive(taskCtx, config, tempDir, startTime, prevRunTime)
	case "sync":
		backupErr = s.runSync(taskCtx, config, startTime, prevRunTime)
	}

	// 5. 更新最终状态与清理
	return s.finishTask(taskCtx, config, backupErr, fileCount, fileSize, startTime)
}

// getVaultName Get vault name by ID
// 根据 ID 获取 Vault 名称
func (s *backupService) getVaultName(ctx context.Context, vaultID, uid int64) string {
	if vaultID > 0 {
		if v, err := s.vaultRepo.GetByID(ctx, vaultID, uid); err == nil && v != nil {
			return v.Name
		}
	}
	return "all"
}

// runArchive Execute archive backup (full/incremental)
// 1. Export notes and attachments to temp directory
// 2. Archive to ZIP
// 3. Upload to all configured storage targets
// 执行压缩归档备份 (全量/增量)
// 1. 导出笔记和附件到临时目录
// 2. 打包为 ZIP
// 3. 上传到配置的所有存储目标
func (s *backupService) runArchive(ctx context.Context, config *domain.BackupConfig, tempDir string, startTime time.Time, lastRun time.Time) (int64, int64, error) {
	uid := config.UID
	vaultName := s.getVaultName(ctx, config.VaultID, uid)
	zipName := fmt.Sprintf("backup_%s_%d_%s_%s.zip", config.Type, uid, vaultName, startTime.Format("20060102_150405"))
	zipPath := filepath.Join(os.TempDir(), zipName)

	defer os.Remove(zipPath)

	// 1. 收集资源 (包含笔记和附件)
	count, size, err := s.exportArchiveFiles(ctx, uid, config.VaultID, tempDir, config.Type == "incremental", lastRun)
	if err != nil {
		return 0, 0, err
	}

	if count == 0 {
		s.recordNoUpdateHistory(ctx, config, startTime)
		return 0, 0, errNoUpdates
	}

	// 2. 压缩打包
	if err := util.Zip(tempDir, zipPath); err != nil {
		return 0, 0, err
	}

	// 3. 上传到所有存储目标
	var storageIds []int64
	if err := json.Unmarshal([]byte(config.StorageIds), &storageIds); err != nil {
		return count, size, code.ErrorBackupStorageIDInvalid
	}

	for _, sid := range storageIds {
		st, err := s.storageService.Get(ctx, uid, sid)
		if err != nil {
			s.logger.Warn("Failed to get storage config, skipping", zap.Int64("sid", sid), zap.Error(err))
			continue
		}
		if !st.IsEnabled {
			s.logger.Info("Storage is disabled, skipping", zap.Int64("sid", sid))
			continue
		}
		s.uploadArchive(ctx, uid, config.ID, st, zipPath, zipName, config.Type, startTime, count, size)
	}

	return count, size, nil
}

// runSync Execute real-time file sync
// Iterate through file changes and mirror sync to all storage targets (no archiving)
// 执行实时文件同步
// 遍历文件变更，直接镜像同步到所有存储目标 (不打包)
func (s *backupService) runSync(ctx context.Context, config *domain.BackupConfig, startTime time.Time, lastRun time.Time) error {
	var storageIds []int64
	if err := json.Unmarshal([]byte(config.StorageIds), &storageIds); err != nil {
		return code.ErrorBackupStorageIDInvalid
	}

	// First, check if there are any updates across all storages
	// Note: syncFiles will check all resources and return true if any changes found
	hasUpdates, err := s.syncFiles(ctx, config.UID, config.VaultID, config.ID, nil, startTime, lastRun)
	if err != nil {
		return err
	}
	if !hasUpdates {
		s.recordNoUpdateHistory(ctx, config, startTime)
		return errNoUpdates
	}

	var syncErrors []string
	for _, sid := range storageIds {
		st, err := s.storageService.Get(ctx, config.UID, sid)
		if err != nil {
			s.logger.Warn("Failed to get storage config, skipping", zap.Int64("sid", sid), zap.Error(err))
			syncErrors = append(syncErrors, fmt.Sprintf("storage %d: config error: %v", sid, err))
			continue
		}
		if !st.IsEnabled {
			s.logger.Info("Storage is disabled, skipping", zap.Int64("sid", sid))
			continue
		}
		if st.Type == storage.LOCAL {
			st.CustomPath = filepath.Join(strconv.FormatInt(config.UID, 10), strconv.FormatInt(config.VaultID, 10), st.CustomPath)
		}
		if _, err := s.syncFiles(ctx, config.UID, config.VaultID, config.ID, st, startTime, lastRun); err != nil {
			s.logger.Warn("Sync to storage failed", zap.Int64("sid", sid), zap.String("type", st.Type), zap.Error(err))
			syncErrors = append(syncErrors, fmt.Sprintf("storage %d (%s): %v", sid, st.Type, err))
		}
	}
	if len(syncErrors) > 0 {
		return fmt.Errorf("sync errors: %s", strings.Join(syncErrors, "; "))
	}
	return nil
}

// finishTask Update final status and cleanup after task completion
// 任务完成后的状态更新与清理
// finishTask Update final status and cleanup after task completion
// 任务完成后的状态更新与清理
func (s *backupService) finishTask(ctx context.Context, config *domain.BackupConfig, err error, fileCount, fileSize int64, startTime time.Time) error {
	config.LastRunTime = startTime // 更新最后执行时间

	if s.ctx.Err() != nil {
		// Service shutdown or context cancelled
		config.LastStatus = domain.BackupStatusStopped // 4: Stopped
		config.LastMessage = "Backup stopped by system"
		if err != nil {
			config.LastMessage += fmt.Sprintf(": %v", err)
		}
	} else if err == nil {
		config.LastStatus = domain.BackupStatusSuccess // 2: Success
		config.LastMessage = "Backup completed successfully"
	} else if errors.Is(err, errNoUpdates) {
		config.LastStatus = domain.BackupStatusNoUpdate // 5: No update
		config.LastMessage = "Backup success, no updates found"
		err = nil // Clear error for return
	} else {
		config.LastStatus = domain.BackupStatusFailed // 3: Failed
		config.LastMessage = fmt.Sprintf("Backup failed: %v", err)
	}

	// Use a new context for status update to ensure it persists even if the task context is cancelled
	saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute) // Increased timeout for file deletion
	defer cancel()

	s.calculateNextRunTime(config)
	s.backupRepo.SaveConfig(saveCtx, config, config.UID)

	if config.RetentionDays != 0 {
		var cutoffTime time.Time
		if config.RetentionDays == -1 {
			// -1 means clean up all history except the current one
			cutoffTime = startTime
		} else if config.RetentionDays > 0 {
			// > 0 means clean up history older than RetentionDays
			cutoffTime = time.Now().AddDate(0, 0, -config.RetentionDays)
		}

		if !cutoffTime.IsZero() {
			// 1. Fetch old history before deleting from DB
			oldHistories, err := s.backupRepo.ListOldHistory(saveCtx, config.UID, config.ID, cutoffTime)
			if err != nil {
				s.logger.Error("Failed to list old backup history for cleanup", zap.Error(err))
			} else {
				// 2. Delete corresponding files in storage for non-sync backups
				for _, history := range oldHistories {
					if history.Type != "sync" && history.FilePath != "" {
						st, err := s.storageService.Get(saveCtx, history.UID, history.StorageID)
						if err != nil || st == nil || !st.IsEnabled {
							s.logger.Warn("Could not get storage client for cleanup or storage disabled", zap.Int64("sid", history.StorageID), zap.Error(err))
							continue
						}

						client, err := s.getStorageClient(saveCtx, history.UID, st)
						if err != nil {
							s.logger.Warn("Failed to initialize storage client for cleanup", zap.Error(err))
							continue
						}

						if err := client.Delete(history.FilePath); err != nil {
							s.logger.Warn("Failed to delete old backup file", zap.String("file", history.FilePath), zap.Error(err))
						} else {
							s.logger.Info("Successfully deleted old backup file", zap.String("file", history.FilePath))
						}
					}
				}
			}

			// 3. Delete records from database
			if err := s.backupRepo.DeleteOldHistory(saveCtx, config.UID, config.ID, cutoffTime); err != nil {
				s.logger.Error("Failed to delete old backup history records from database", zap.Error(err))
			}
		}
	}

	return err
}

// exportArchiveFiles Export files to be backed up to temp directory for subsequent archiving
// 将需要备份的文件导出到临时目录，用于后续打包
func (s *backupService) exportArchiveFiles(ctx context.Context, uid, vaultID int64, targetDir string, incremental bool, lastRun time.Time) (int64, int64, error) {
	if vaultID <= 0 {
		return 0, 0, code.ErrorBackupVaultRequired
	}

	vault, err := s.vaultRepo.GetByID(ctx, vaultID, uid)
	if err != nil {
		return 0, 0, err
	}
	if vault == nil {
		return 0, 0, code.ErrorVaultNotFound
	}

	totalCount := int64(0)
	totalSize := int64(0)

	err = s.forEachResource(ctx, uid, vault, incremental, lastRun, func(v *domain.Vault, path string, isNote bool, content []byte, localSize int64, localPath string, mtime time.Time, isDeleted bool) error {
		if isDeleted {
			return nil
		}

		destPath := filepath.Join(targetDir, path)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		if isNote {
			if err := os.WriteFile(destPath, content, 0644); err != nil {
				return err
			}
		} else {
			if err := util.CopyFile(localPath, destPath); err != nil {
				return err
			}
		}
		totalCount++
		totalSize += localSize
		return nil
	})

	return totalCount, totalSize, err
}

// uploadArchive Upload the archived ZIP file to specified storage target
// 将打包好的 ZIP 文件上传到指定的存储目标
func (s *backupService) uploadArchive(ctx context.Context, uid, configId int64, stDTO *dto.StorageDTO, filePath, fileName, bType string, startTime time.Time, count, size int64) {
	h := &domain.BackupHistory{
		UID:       uid,
		ConfigID:  configId,
		StorageID: stDTO.ID,
		Type:      bType,
		StartTime: startTime,
		Status:    domain.BackupStatusRunning,
		FileCount: count,
		FileSize:  size,
		FilePath:  fileName,
	}

	h, err := s.backupRepo.CreateHistory(ctx, h, uid)
	if err != nil {
		s.logger.Error("Failed to create backup history", zap.Error(err))
		return
	}

	client, err := s.getStorageClient(ctx, uid, stDTO)
	if err != nil {
		s.updateHistory(ctx, h, domain.BackupStatusFailed, err.Error())
		return
	}

	f, err := os.Open(filePath)
	if err != nil {
		s.updateHistory(ctx, h, domain.BackupStatusFailed, fmt.Sprintf("Failed to open backup file: %v", err))
		return
	}
	defer f.Close()

	_, err = client.SendFile(fileName, f, "application/zip", startTime)
	if err != nil {
		s.updateHistory(ctx, h, domain.BackupStatusFailed, fmt.Sprintf("Upload failed: %v", err))
		return
	}

	s.updateHistory(ctx, h, domain.BackupStatusSuccess, "Success")
}

// syncFiles Sync file changes to specified storage target (supports add, modify, delete)
// returns (hasChanges, error)
// 将文件变更同步到指定的存储目标 (支持新增、修改和删除)
func (s *backupService) syncFiles(ctx context.Context, uid, vaultID, configId int64, stDTO *dto.StorageDTO, startTime time.Time, lastRun time.Time) (bool, error) {
	var h *domain.BackupHistory
	var client pkgstorage.Storager

	if stDTO != nil {
		h = &domain.BackupHistory{
			UID:       uid,
			ConfigID:  configId,
			StorageID: stDTO.ID,
			Type:      "sync",
			StartTime: startTime,
			Status:    domain.BackupStatusRunning,
		}

		var err error
		h, err = s.backupRepo.CreateHistory(ctx, h, uid)
		if err != nil {
			s.logger.Error("Failed to create sync history", zap.Error(err))
			return false, err
		}

		client, err = s.getStorageClient(ctx, uid, stDTO)
		if err != nil {
			s.updateHistory(ctx, h, domain.BackupStatusFailed, err.Error())
			return false, err
		}
	}

	if vaultID <= 0 {
		if h != nil {
			s.updateHistory(ctx, h, domain.BackupStatusFailed, code.ErrorBackupVaultRequired.Msg())
		}
		return false, code.ErrorBackupVaultRequired
	}

	vault, err := s.vaultRepo.GetByID(ctx, vaultID, uid)
	if err != nil {
		if h != nil {
			s.updateHistory(ctx, h, domain.BackupStatusFailed, err.Error())
		}
		return false, err
	}
	if vault == nil {
		if h != nil {
			s.updateHistory(ctx, h, domain.BackupStatusFailed, code.ErrorVaultNotFound.Msg())
		}
		return false, code.ErrorVaultNotFound
	}

	totalCount, totalSize := int64(0), int64(0)
	failedCount := int64(0)
	var lastSendErr error
	hasChanges := false
	err = s.forEachResource(ctx, uid, vault, !lastRun.IsZero(), lastRun, func(v *domain.Vault, path string, isNote bool, content []byte, localSize int64, localPath string, mtime time.Time, isDeleted bool) error {
		hasChanges = true
		if client == nil {
			return nil // Just checking for changes
		}

		// Remove VaultName from path as per user request
		// 根据用户要求移除 VaultName
		objName := path
		if isDeleted {
			if delErr := client.Delete(objName); delErr != nil {
				failedCount++
				lastSendErr = delErr
				s.logger.Warn("Sync delete failed", zap.String("path", objName), zap.Error(delErr))
			}
			return nil
		}

		var sendErr error
		if isNote {
			_, sendErr = client.SendContent(objName, content, mtime)
		} else {
			if f, err := os.Open(localPath); err == nil {
				_, sendErr = client.SendFile(objName, f, "application/octet-stream", mtime)
				f.Close()
			} else {
				sendErr = err
			}
		}

		if sendErr != nil {
			failedCount++
			lastSendErr = sendErr
			s.logger.Warn("Sync upload failed", zap.String("path", objName), zap.Error(sendErr))
		} else {
			totalCount++
			totalSize += localSize
		}
		return nil
	})

	if err != nil {
		if h != nil {
			s.updateHistory(ctx, h, domain.BackupStatusFailed, err.Error())
		}
		return hasChanges, err
	}

	if h != nil {
		h.FileCount = totalCount
		h.FileSize = totalSize
		if !hasChanges {
			s.updateHistory(ctx, h, domain.BackupStatusNoUpdate, "No updates")
		} else if failedCount > 0 {
			msg := fmt.Sprintf("Partial failure: %d files synced, %d files failed. Last error: %v", totalCount, failedCount, lastSendErr)
			s.updateHistory(ctx, h, domain.BackupStatusFailed, msg)
		} else {
			s.updateHistory(ctx, h, domain.BackupStatusSuccess, "Success")
		}
	}

	if failedCount > 0 {
		return hasChanges, fmt.Errorf("sync completed with %d failures, last error: %w", failedCount, lastSendErr)
	}
	return hasChanges, nil
}

type resourceAction func(v *domain.Vault, path string, isNote bool, content []byte, localSize int64, localPath string, mtime time.Time, isDeleted bool) error

// forEachResource Iterate through all resources (notes and attachments) in the specified vault
// 遍历指定 Vault 中的所有资源 (笔记和附件)
func (s *backupService) forEachResource(ctx context.Context, uid int64, v *domain.Vault, incremental bool, lastRun time.Time, action resourceAction) error {
	// Check context before processing
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// 1. 处理笔记
	var notes []*domain.Note
	var err error
	if incremental && !lastRun.IsZero() {
		notes, err = s.noteRepo.ListByUpdatedTimestamp(ctx, lastRun.UnixMilli(), v.ID, uid)
	} else {
		// List(ctx, vaultID, page, pageSize, uid, keyword, isDeleted, sort, isAsc, tag, folder)
		notes, err = s.noteRepo.List(ctx, v.ID, 1, 1000000, uid, "", false, "", false, "", "")
	}

	if err != nil {
		return err
	}
	for _, n := range notes {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		path := n.Path
		if filepath.Ext(path) != ".md" {
			path += ".md"
		}
		if err := action(v, path, true, []byte(n.Content), int64(len(n.Content)), "", time.UnixMilli(n.Mtime), n.IsDeleted()); err != nil {
			return err
		}
	}

	// 2. 处理附件
	var files []*domain.File
	if incremental && !lastRun.IsZero() {
		files, err = s.fileRepo.ListByUpdatedTimestamp(ctx, lastRun.UnixMilli(), v.ID, uid)
	} else {
		files, err = s.fileRepo.List(ctx, v.ID, 1, 1000000, uid, "", false, "", "")
	}

	if err != nil {
		return err
	}
	for _, f := range files {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		var size int64
		// Check file existence/size if not deleted
		if !f.IsDeleted() {
			if info, _ := os.Stat(f.SavePath); info != nil {
				size = info.Size()
			}
		}
		if err := action(v, f.Path, false, nil, size, f.SavePath, time.UnixMilli(f.Mtime), f.IsDeleted()); err != nil {
			return err
		}
	}

	return nil
}

// getStorageClient Get and initialize storage client
// 获取并初始化存储客户端
func (s *backupService) getStorageClient(ctx context.Context, uid int64, stDTO *dto.StorageDTO) (pkgstorage.Storager, error) {
	sConfig := &pkgstorage.Config{
		Type:            stDTO.Type,
		CustomPath:      stDTO.CustomPath,
		Endpoint:        stDTO.Endpoint,
		Region:          stDTO.Region,
		BucketName:      stDTO.BucketName,
		AccessKeyID:     stDTO.AccessKeyID,
		AccessKeySecret: stDTO.AccessKeySecret,
		AccountID:       stDTO.AccountID,
		User:            stDTO.User,
		Password:        stDTO.Password,
		SavePath:        s.storageConfig.LocalFS.SavePath,
	}

	return pkgstorage.NewClient(sConfig)
}

func (s *backupService) updateHistory(ctx context.Context, h *domain.BackupHistory, status int, message string) {
	h.Status = status
	h.Message = message
	h.EndTime = time.Now()

	// Use a new context for history update to ensure it persists even if the task context is cancelled
	saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.backupRepo.CreateHistory(saveCtx, h, h.UID)
}

func (s *backupService) recordNoUpdateHistory(ctx context.Context, config *domain.BackupConfig, startTime time.Time) {
	var storageIds []int64
	if err := json.Unmarshal([]byte(config.StorageIds), &storageIds); err != nil {
		return
	}

	for _, sid := range storageIds {
		h := &domain.BackupHistory{
			UID:       config.UID,
			ConfigID:  config.ID,
			StorageID: sid,
			Type:      config.Type,
			StartTime: startTime,
			Status:    domain.BackupStatusNoUpdate,
			Message:   "No updates",
			EndTime:   time.Now(),
		}
		// Use a new context for history update
		saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		s.backupRepo.CreateHistory(saveCtx, h, config.UID)
		cancel()
	}
}

const syncDebounceDelay = 30 * time.Second

// NotifyUpdated Trigger debounced incremental sync task
// Called when note/file/folder changes, executes ExecuteUserBackup after syncDebounceDelay
// 触发防抖的增量同步任务
// 当笔记/文件/目录发生变更时调用，会在延迟 syncDebounceDelay 后执行 ExecuteUserBackup
func (s *backupService) NotifyUpdated(uid int64) {
	s.timerMu.Lock()
	defer s.timerMu.Unlock()

	if timer, ok := s.syncTimers[uid]; ok {
		timer.Stop()
	}

	s.syncTimers[uid] = time.AfterFunc(syncDebounceDelay, func() {
		s.logger.Info("Triggering debounced sync (memory flag)", zap.Int64("uid", uid))

		// Set in-memory flag instead of DB write
		s.pendingSyncs.Store(uid, true)

		s.timerMu.Lock()
		delete(s.syncTimers, uid)
		s.timerMu.Unlock()
	})
}

// Shutdown Clean up resources and handle state changes during shutdown
// 停止服务，清理资源并处理关闭时的状态变更
func (s *backupService) Shutdown(ctx context.Context) error {
	// 1. Signal all background tasks to stop
	s.cancel()

	s.timerMu.Lock()
	// Stop all pending sync timers
	for uid, timer := range s.syncTimers {
		if timer.Stop() {
			s.logger.Info("Stopped pending sync timer during shutdown", zap.Int64("uid", uid))
		}
	}
	// Clear the map
	s.syncTimers = make(map[int64]*time.Timer)
	s.timerMu.Unlock()

	// 2. Wait for active backup/sync tasks to finish or abort
	// We use a channel to support timeout if needed, though ctx passed to Shutdown usually handles timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("All backup tasks finished successfully during shutdown")
	case <-ctx.Done():
		s.logger.Warn("Shutdown context expired before all backup tasks finished")
		return ctx.Err()
	}

	return nil
}

var _ BackupService = (*backupService)(nil)
