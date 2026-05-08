package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	appconfig "github.com/haierkeys/fast-note-sync-service/internal/config"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// errNoChanges indicates that no changes were found after the Git sync check
// errNoChanges 表示 Git 同步检查后没有发现任何需要提交的变更
var errNoChanges = errors.New("no changes found")

const gitSyncBatchSize = 100

// GitSyncService defines the Git synchronization business service interface
// GitSyncService 定义 Git 同步业务服务接口
type GitSyncService interface {
	GetConfigs(ctx context.Context, uid int64) ([]*dto.GitSyncConfigDTO, error)
	GetConfig(ctx context.Context, uid int64, vaultID int64) (*dto.GitSyncConfigDTO, error)
	UpdateConfig(ctx context.Context, uid int64, params *dto.GitSyncConfigRequest) (*dto.GitSyncConfigDTO, error)
	DeleteConfig(ctx context.Context, uid int64, id int64) error
	Validate(ctx context.Context, params *dto.GitSyncValidateRequest) error
	ExecuteSync(ctx context.Context, uid int64, id int64) error
	CleanWorkspace(ctx context.Context, uid int64, configID int64) error
	ListHistory(ctx context.Context, uid int64, configID int64, pager *pkgapp.Pager) ([]*dto.GitSyncHistoryDTO, int64, error)
	NotifyUpdated(uid int64, vaultID int64)
	Shutdown(ctx context.Context) error
}

type gitSyncService struct {
	repo        domain.GitSyncRepository
	noteRepo    domain.NoteRepository
	folderRepo  domain.FolderRepository
	fileRepo    domain.FileRepository
	vaultRepo   domain.VaultRepository
	settingRepo domain.SettingRepository
	gitConf     *appconfig.GitConfig
	logger      *zap.Logger
	mu          sync.Mutex
	running     map[int64]context.CancelFunc // configID -> cancelFunc
	timers      map[int64]*time.Timer        // configID -> timer
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	gcTimer     *time.Timer // Timer for delayed GC // 延迟 GC 定时器
	gcMu        sync.Mutex  // Mutex for gcTimer // 保护 gcTimer 的互斥锁
}

// NewGitSyncService creates a GitSyncService instance
// NewGitSyncService 创建 GitSyncService 实例
func NewGitSyncService(repo domain.GitSyncRepository, noteRepo domain.NoteRepository, folderRepo domain.FolderRepository, fileRepo domain.FileRepository, vaultRepo domain.VaultRepository, settingRepo domain.SettingRepository, gitConf *appconfig.GitConfig, logger *zap.Logger) GitSyncService {
	ctx, cancel := context.WithCancel(context.Background())
	return &gitSyncService{
		repo:        repo,
		noteRepo:    noteRepo,
		folderRepo:  folderRepo,
		fileRepo:    fileRepo,
		vaultRepo:   vaultRepo,
		settingRepo: settingRepo,
		gitConf:     gitConf,
		logger:      logger,
		running:     make(map[int64]context.CancelFunc),
		timers:      make(map[int64]*time.Timer),
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (s *gitSyncService) domainToDTO(conf *domain.GitSyncConfig) *dto.GitSyncConfigDTO {
	if conf == nil {
		return nil
	}
	res := &dto.GitSyncConfigDTO{
		ID:              conf.ID,
		UID:             conf.UID,
		RepoURL:         conf.RepoURL,
		Username:        conf.Username,
		Password:        conf.Password,
		Branch:          conf.Branch,
		IsEnabled:       conf.IsEnabled,
		Delay:           conf.Delay,
		RetentionDays:   conf.RetentionDays,
		LastStatus:      conf.LastStatus,
		LastMessage:     conf.LastMessage,
		IncludeConfig:   conf.IncludeConfig,
		ConfigSyncRules: conf.ConfigSyncRules,
		CreatedAt:       timex.Time(conf.CreatedAt),
		UpdatedAt:       timex.Time(conf.UpdatedAt),
	}
	if conf.LastSyncTime != nil {
		res.LastSyncTime = timex.Time(*conf.LastSyncTime)
	}

	// Fetch vault name if possible
	if conf.VaultID > 0 {
		v, err := s.vaultRepo.GetByID(context.Background(), conf.VaultID, conf.UID)
		if err == nil {
			res.Vault = v.Name
		}
	}

	return res
}

func (s *gitSyncService) GetConfigs(ctx context.Context, uid int64) ([]*dto.GitSyncConfigDTO, error) {
	configs, err := s.repo.List(ctx, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	var res []*dto.GitSyncConfigDTO
	for _, c := range configs {
		res = append(res, s.domainToDTO(c))
	}
	return res, nil
}

func (s *gitSyncService) GetConfig(ctx context.Context, uid int64, vaultID int64) (*dto.GitSyncConfigDTO, error) {
	conf, err := s.repo.GetByVaultID(ctx, vaultID, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	if conf == nil {
		return nil, code.ErrorVaultNotFound
	}
	return s.domainToDTO(conf), nil
}

func (s *gitSyncService) UpdateConfig(ctx context.Context, uid int64, params *dto.GitSyncConfigRequest) (*dto.GitSyncConfigDTO, error) {
	var conf *domain.GitSyncConfig
	var err error

	if params.ID > 0 {
		conf, err = s.repo.GetByID(ctx, params.ID, uid)
		if err != nil {
			return nil, code.ErrorDBQuery.WithDetails(err.Error())
		}
		if conf == nil {
			return nil, code.ErrorGitSyncNotFound
		}
	} else {
		conf = &domain.GitSyncConfig{
			UID: uid,
		}
	}

	if params.Vault != "" {
		v, err := s.vaultRepo.GetByName(ctx, params.Vault, uid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, code.ErrorVaultNotFound
			}
			return nil, code.ErrorDBQuery.WithDetails(err.Error())
		}
		conf.VaultID = v.ID
	}

	conf.RepoURL = params.RepoURL
	conf.Username = params.Username
	conf.Password = params.Password
	conf.Branch = params.Branch
	if conf.Branch == "" {
		conf.Branch = "main"
	}
	conf.IsEnabled = params.IsEnabled
	conf.Delay = params.Delay
	conf.RetentionDays = params.RetentionDays
	conf.IncludeConfig = params.IncludeConfig
	conf.ConfigSyncRules = params.ConfigSyncRules

	saved, err := s.repo.Save(ctx, conf, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	return s.domainToDTO(saved), nil
}

func (s *gitSyncService) DeleteConfig(ctx context.Context, uid int64, id int64) error {
	// 1. Verification identity // 1. 验证身份
	conf, err := s.repo.GetByID(ctx, id, uid)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}
	if conf == nil {
		return code.ErrorGitSyncNotFound
	}

	err = s.repo.Delete(ctx, id, uid)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Clean workspace as well? User request said "Cleanup API". Delete config doesn't necessarily mean delete workspace.
	// But usually it's better to cleanup. However, I'll follow the plan and keep them separate as per the "Cleanup API" request.

	return nil
}

func (s *gitSyncService) Validate(ctx context.Context, params *dto.GitSyncValidateRequest) error {
	branch := params.Branch
	if branch == "" {
		branch = "main"
	}

	auth := &githttp.BasicAuth{
		Username: params.Username,
		Password: params.Password,
	}

	// Try LsRemote to validate credentials and repo visibility
	rem := git.NewRemote(nil, &config.RemoteConfig{
		Name: "origin",
		URLs: []string{params.RepoURL},
	})

	refs, err := rem.List(&git.ListOptions{
		Auth: auth,
	})
	if err != nil {
		if errors.Is(err, transport.ErrEmptyRemoteRepository) {
			// Remote is empty, validation success
			return nil
		}
		return code.ErrorGitSyncValidateFailed.WithDetails(err.Error())
	}

	// Check if branch exists
	branchRef := plumbing.NewBranchReferenceName(branch)
	found := false
	for _, ref := range refs {
		if ref.Name() == branchRef || ref.Name() == plumbing.HEAD {
			found = true
			break
		}
	}

	if !found {
		// Even if branch not found, if it's an empty repo (though List usually returns ErrEmptyRemoteRepository),
		// we should have caught it above. If we are here, it means refs is not empty but branch not found.
		return code.ErrorGitSyncValidateFailed.WithDetails("Branch not found in remote")
	}

	return nil
}

func (s *gitSyncService) ExecuteSync(ctx context.Context, uid int64, id int64) error {
	conf, err := s.repo.GetByID(ctx, id, uid)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}
	if conf == nil {
		return code.ErrorGitSyncNotFound
	}

	// Strategy: For sync/mirror sync, directly cancel the old task and start a new one
	// 策略：同步/镜像同步直接取消旧任务，启动新任务
	s.mu.Lock()
	if oldCancel, running := s.running[id]; running {
		s.logger.Info("Cancelling existing Git sync task to start a newer one", zap.Int64("uid", uid), zap.Int64("configId", id))
		oldCancel()
		delete(s.running, id)
	}

	// Create context for new task
	// 为新任务创建 context
	taskCtx, taskCancel := context.WithCancel(s.ctx)
	s.running[id] = taskCancel
	s.mu.Unlock()

	s.wg.Add(1)
	// Run in background
	go func() {
		defer func() {
			s.mu.Lock()
			// Ensure only the current cancel function is cleaned up
			// 确保只清理当前的 cancel 函数
			if _, ok := s.running[id]; ok {
				// 虽然 sync 策略下会先 cancel 再 set，但为了闭包内引用的严谨
				delete(s.running, id)
			}
			s.mu.Unlock()
			taskCancel()
			s.wg.Done()
		}()

		// Use the newly created task context
		// 使用新创建的任务 context
		s.syncTask(taskCtx, conf)
	}()

	return nil
}

func (s *gitSyncService) CleanWorkspace(ctx context.Context, uid int64, configID int64) error {
	if configID > 0 {
		// 1. Reset database fields // 1. 重置数据库字段
		conf, err := s.repo.GetByID(ctx, configID, uid)
		if err != nil {
			return code.ErrorDBQuery.WithDetails(err.Error())
		}
		if conf == nil {
			return code.ErrorGitSyncNotFound
		}

		conf.LastSyncTime = nil
		conf.LastStatus = domain.GitSyncStatusIdle
		conf.LastMessage = ""

		_, err = s.repo.Save(ctx, conf, uid)
		if err != nil {
			return code.ErrorDBQuery.WithDetails(err.Error())
		}

		// 2. Delete History // 2. 删除历史记录
		_ = s.repo.DeleteHistory(ctx, uid, configID)

		// 3. Remove physical workspace // 3. 删除物理工作区
		path := s.getWorkspacePath(uid, configID)
		err = os.RemoveAll(path)
		if err != nil {
			s.logger.Warn("Failed to cleanup physical workspace", zap.String("path", path), zap.Error(err))
		}
	} else {
		// 1. Reset all database fields for user // 1. 重置用户的所有数据库字段
		configs, err := s.repo.List(ctx, uid)
		if err != nil {
			return code.ErrorDBQuery.WithDetails(err.Error())
		}
		for _, conf := range configs {
			conf.LastSyncTime = nil
			conf.LastStatus = domain.GitSyncStatusIdle
			conf.LastMessage = ""
			_, _ = s.repo.Save(ctx, conf, uid)
		}

		// 2. Delete All History for user // 2. 删除用户的所有历史记录
		_ = s.repo.DeleteHistory(ctx, uid, 0)

		// 3. Remove all physical workspaces for user // 3. 删除用户的所有物理工作区
		path := s.getUserWorkspacePath(uid)
		err = os.RemoveAll(path)
		if err != nil {
			s.logger.Warn("Failed to cleanup user workspaces", zap.String("path", path), zap.Error(err))
		}
	}

	return nil
}

func (s *gitSyncService) ListHistory(ctx context.Context, uid int64, configID int64, pager *pkgapp.Pager) ([]*dto.GitSyncHistoryDTO, int64, error) {
	histories, count, err := s.repo.ListHistory(ctx, uid, configID, pager.Page, pager.PageSize)
	if err != nil {
		return nil, 0, code.ErrorDBQuery.WithDetails(err.Error())
	}
	var res []*dto.GitSyncHistoryDTO
	for _, h := range histories {
		res = append(res, s.historyToDTO(h))
	}
	return res, count, nil
}

func (s *gitSyncService) Shutdown(ctx context.Context) error {
	s.cancel()

	s.mu.Lock()
	for _, cancel := range s.running {
		cancel()
	}
	for _, timer := range s.timers {
		timer.Stop()
	}
	s.mu.Unlock()

	s.gcMu.Lock()
	if s.gcTimer != nil {
		s.gcTimer.Stop()
	}
	s.gcMu.Unlock()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Internal methods // 内部方法

func (s *gitSyncService) historyToDTO(h *domain.GitSyncHistory) *dto.GitSyncHistoryDTO {
	if h == nil {
		return nil
	}
	return &dto.GitSyncHistoryDTO{
		ID:        h.ID,
		ConfigID:  h.ConfigID,
		StartTime: timex.Time(h.StartTime),
		EndTime:   timex.Time(h.EndTime),
		Status:    h.Status,
		Message:   h.Message,
		CreatedAt: timex.Time(h.CreatedAt),
	}
}

func (s *gitSyncService) getWorkspacePath(uid, configID int64) string {
	return filepath.Join(s.getUserWorkspacePath(uid), fmt.Sprintf("%d", configID))
}

func (s *gitSyncService) getUserWorkspacePath(uid int64) string {
	return filepath.Join("storage", "git_workspace", fmt.Sprintf("%d", uid))
}

func (s *gitSyncService) syncTask(ctx context.Context, conf *domain.GitSyncConfig) {
	startTime := time.Now()
	s.logger.Info("Starting Git sync task", zap.Int64("configId", conf.ID), zap.Int64("uid", conf.UID))

	// Record status before running to recover if there are no changes
	// 记录运行前的状态，以便无变更时恢复
	prevStatus := conf.LastStatus

	// Update Config Status to Running
	conf.LastStatus = domain.GitSyncStatusRunning
	_, _ = s.repo.Save(ctx, conf, conf.UID)

	err := s.doSync(ctx, conf)

	// No changes: restore original status, trigger Save only to update updated_at
	// Without writing history, without changing last_sync_time / last_status / last_message
	// 无变更：恢复原始状态，只触发 Save 更新 updated_at
	// 不写 history，不改 last_sync_time / last_status / last_message
	if errors.Is(err, errNoChanges) {
		s.logger.Info("No changes found, skipping history and status update", zap.Int64("configId", conf.ID))
		conf.LastStatus = prevStatus
		_, _ = s.repo.Save(context.Background(), conf, conf.UID)
		return
	}

	endTime := time.Now()
	var finalStatus int64
	var message string

	if ctx.Err() != nil {
		finalStatus = domain.GitSyncStatusShutdown
		message = "Sync stopped by system shutdown"
		if err != nil {
			message += ": " + err.Error()
		}
	} else if err != nil {
		s.logger.Error("Git sync task failed", zap.Int64("configId", conf.ID), zap.Error(err))
		finalStatus = domain.GitSyncStatusFailed
		message = err.Error()
	} else {
		s.logger.Info("Git sync task success", zap.Int64("configId", conf.ID))
		finalStatus = domain.GitSyncStatusSuccess
		message = "Sync completed at " + endTime.Format("2006-01-02 15:04:05")
		conf.LastSyncTime = &endTime
	}

	// Update Config Final Status
	conf.LastStatus = finalStatus
	conf.LastMessage = message
	_, _ = s.repo.Save(context.Background(), conf, conf.UID)

	// Create History Record
	h := &domain.GitSyncHistory{
		ConfigID:  conf.ID,
		UID:       conf.UID,
		StartTime: startTime,
		EndTime:   endTime,
		Status:    finalStatus,
		Message:   message,
	}
	_, _ = s.repo.CreateHistory(context.Background(), h, conf.UID)

	// Automatically clean up expired history records
	// 自动清理过期历史记录
	if conf.RetentionDays != 0 {
		var cutoffTime time.Time
		if conf.RetentionDays == -1 {
			// -1 means retain only the current latest record
			// -1 表示仅保留当前最新的一条记录
			cutoffTime = startTime
		} else if conf.RetentionDays > 0 {
			// > 0 means clean up records exceeding specified days
			// > 0 表示清理超过指定天数的记录
			cutoffTime = time.Now().AddDate(0, 0, -int(conf.RetentionDays))
		}

		if !cutoffTime.IsZero() {
			if err := s.repo.DeleteOldHistory(context.Background(), conf.UID, conf.ID, cutoffTime); err != nil {
				s.logger.Error("Failed to delete old git sync history", zap.Error(err), zap.Int64("configId", conf.ID))
			}
		}
	}

	// Schedule delayed memory release after task completion (addressing Issue #113)
	// Return virtual memory to OS 30 minutes after high-pressure sync ends to avoid frequent operations
	// 任务结束后调度延迟内存释放 (针对 Issue #113)
	// 高压同步结束后 30 分钟再归还虚拟内存给操作系统，避免频繁操作
	s.scheduleGC()
}

// scheduleGC schedules a delayed GC and FreeOSMemory call (debounced)
// scheduleGC 调度一个延迟的 GC 和内存释放操作（防抖）
func (s *gitSyncService) scheduleGC() {
	s.gcMu.Lock()
	defer s.gcMu.Unlock()

	if s.gcTimer != nil {
		s.gcTimer.Stop()
	}

	s.gcTimer = time.AfterFunc(30*time.Minute, func() {
		s.logger.Info("Triggering delayed background GC and memory release to OS")
		runtime.GC()
		debug.FreeOSMemory()
	})
}

func (s *gitSyncService) doSync(ctx context.Context, conf *domain.GitSyncConfig) error {
	wsPath := s.getWorkspacePath(conf.UID, conf.ID)
	auth := &githttp.BasicAuth{
		Username: conf.Username,
		Password: conf.Password,
	}

	var r *git.Repository
	var err error

	// 1. Check/Init Local Repo // 1. 检查/初始化本地仓库
	if _, err := os.Stat(filepath.Join(wsPath, ".git")); os.IsNotExist(err) {
		s.logger.Info("Initializing local git repo", zap.String("path", wsPath))
		_ = os.RemoveAll(wsPath)
		r, err = git.PlainClone(wsPath, false, &git.CloneOptions{
			URL:           conf.RepoURL,
			Auth:          auth,
			ReferenceName: plumbing.NewBranchReferenceName(conf.Branch),
			SingleBranch:  true,
			Depth:         1,
			Tags:          git.NoTags,
		})
		if err != nil {
			if errors.Is(err, transport.ErrEmptyRemoteRepository) {
				s.logger.Info("Remote repository is empty, initializing locally", zap.String("path", wsPath))
				r, err = git.PlainInit(wsPath, false)
				if err != nil {
					return fmt.Errorf("git init failed: %w", err)
				}
				_, err = r.CreateRemote(&config.RemoteConfig{
					Name: "origin",
					URLs: []string{conf.RepoURL},
				})
				if err != nil {
					return fmt.Errorf("create remote failed: %w", err)
				}
			} else {
				return fmt.Errorf("git clone failed: %w", err)
			}
		}
	} else {
		r, err = git.PlainOpen(wsPath)
		if err != nil {
			// Try to re-init if open fails
			_ = os.RemoveAll(wsPath)
			return s.doSync(ctx, conf)
		}
	}

	wt, err := r.Worktree()
	if err != nil {
		return err
	}

	// 2. Pull latest // 2. 拉取最新变更
	s.logger.Info("Pulling latest changes", zap.Int64("configId", conf.ID))
	err = wt.Pull(&git.PullOptions{
		Auth:          auth,
		ReferenceName: plumbing.NewBranchReferenceName(conf.Branch),
		SingleBranch:  true,
		Force:         true,
		Depth:         1,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		if errors.Is(err, transport.ErrEmptyRemoteRepository) || errors.Is(err, git.ErrRemoteNotFound) || errors.Is(err, plumbing.ErrReferenceNotFound) {
			s.logger.Info("Remote is empty or branch not found, skipping pull", zap.Int64("configId", conf.ID))
		} else {
			return fmt.Errorf("git pull failed: %w", err)
		}
	}

	// 3. Extract DB content to Workspace // 3. 提取 DB 内容到工作区
	// We need to mirror files from DB to this workspace // 我们需要将 DB 中的文件镜像到此工作区
	changed, err := s.mirrorNotesToWorkspace(ctx, conf, wsPath, conf.LastSyncTime)
	if err != nil {
		return fmt.Errorf("mirror to workspace failed: %w", err)
	}

	if !changed {
		s.logger.Info("No notes or attachments updated, skipping Git operations", zap.Int64("configId", conf.ID))
		return errNoChanges
	}

	// 4. Commit and Push // 4. 提交并推送
	status, err := wt.Status()
	if err != nil {
		return err
	}
	if status.IsClean() {
		s.logger.Info("No changes to commit", zap.Int64("configId", conf.ID))
		return errNoChanges
	}

	err = wt.AddWithOptions(&git.AddOptions{All: true})
	if err != nil {
		return err
	}

	name := s.gitConf.Name
	if name == "" {
		name = "FNS Service"
	}
	email := s.gitConf.Email
	if email == "" {
		email = "fns@email.com"
	}

	_, err = wt.Commit("Update from Sync Service", &git.CommitOptions{
		Author: &object.Signature{
			Name:  name,
			Email: email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	s.logger.Info("Pushing changes", zap.Int64("configId", conf.ID))
	err = r.Push(&git.PushOptions{
		Auth: auth,
	})
	if err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}

	return nil
}

func (s *gitSyncService) mirrorNotesToWorkspace(ctx context.Context, conf *domain.GitSyncConfig, wsPath string, lastSyncTime *time.Time) (bool, error) {
	v, err := s.vaultRepo.GetByID(ctx, conf.VaultID, conf.UID)
	if err != nil {
		return false, err
	}
	if v == nil {
		return false, fmt.Errorf("vault not found")
	}

	var ts int64
	if lastSyncTime != nil {
		ts = lastSyncTime.UnixMilli()
		s.logger.Info("Performing incremental sync to workspace", zap.Int64("configId", conf.ID), zap.Int64("sinceTs", ts))
	} else {
		s.logger.Info("Performing initial full sync to workspace (using unified incremental method)", zap.Int64("configId", conf.ID))
	}

	var actuallyChanged bool

	for offset := 0; ; offset += gitSyncBatchSize {
		notes, err := s.noteRepo.ListByUpdatedTimestampPage(ctx, ts, v.ID, conf.UID, offset, gitSyncBatchSize)
		if err != nil {
			return false, err
		}
		if len(notes) == 0 {
			break
		}

		batchChanged, err := s.processNotesBatch(notes, wsPath)
		if err != nil {
			return false, err
		}
		if batchChanged {
			actuallyChanged = true
		}

		if len(notes) < gitSyncBatchSize {
			break
		}
	}

	for offset := 0; ; offset += gitSyncBatchSize {
		files, err := s.fileRepo.ListByUpdatedTimestampPage(ctx, ts, v.ID, conf.UID, offset, gitSyncBatchSize)
		if err != nil {
			return false, err
		}
		if len(files) == 0 {
			break
		}

		batchChanged, err := s.processFilesBatch(files, conf, wsPath)
		if err != nil {
			return false, err
		}
		if batchChanged {
			actuallyChanged = true
		}

		if len(files) < gitSyncBatchSize {
			break
		}
	}

	if conf.IncludeConfig && len(conf.ConfigSyncRules) > 0 {
		settingsChanged, err := s.mirrorSettingsToWorkspace(ctx, conf, wsPath, lastSyncTime)
		if err != nil {
			return false, fmt.Errorf("mirror settings to workspace failed: %w", err)
		}
		if settingsChanged {
			actuallyChanged = true
		}
	}

	return actuallyChanged, nil
}

func (s *gitSyncService) mirrorSettingsToWorkspace(ctx context.Context, conf *domain.GitSyncConfig, wsPath string, lastSyncTime *time.Time) (bool, error) {
	v, err := s.vaultRepo.GetByID(ctx, conf.VaultID, conf.UID)
	if err != nil {
		return false, err
	}
	if v == nil {
		return false, fmt.Errorf("vault not found")
	}

	var ts int64
	if lastSyncTime != nil {
		ts = lastSyncTime.UnixMilli()
	}

	var actuallyChanged bool

	// Fetch all settings for the vault
	settings, err := s.settingRepo.ListByUpdatedTimestamp(ctx, ts, v.ID, conf.UID)
	if err != nil {
		return false, err
	}

	if len(settings) == 0 {
		return false, nil
	}

	batchChanged, err := s.processSettingsBatch(settings, conf, wsPath)
	if err != nil {
		return false, err
	}
	if batchChanged {
		actuallyChanged = true
	}

	return actuallyChanged, nil
}

func (s *gitSyncService) processSettingsBatch(settings []*domain.Setting, conf *domain.GitSyncConfig, wsPath string) (bool, error) {
	var actuallyChanged bool

	for _, st := range settings {
		// Filter by rules
		matched := false
		for _, rule := range conf.ConfigSyncRules {
			if rule == "" {
				continue
			}
			// Prefix match for directories, exact match for files
			if st.Path == rule || (len(st.Path) > len(rule) && st.Path[:len(rule)] == rule && (rule[len(rule)-1] == '/' || st.Path[len(rule)] == '/')) {
				matched = true
				break
			}
		}

		if !matched {
			continue
		}

		fullPath := filepath.Join(wsPath, st.Path)

		if st.Action == domain.SettingActionDelete {
			if _, err := os.Stat(fullPath); err == nil {
				_ = os.Remove(fullPath)
				actuallyChanged = true
			}
			continue
		}

		_ = os.MkdirAll(filepath.Dir(fullPath), 0755)

		// Check if content is different before writing
		if oldFile, err := os.Open(fullPath); err == nil {
			defer oldFile.Close()
			if oldContent, err := io.ReadAll(oldFile); err == nil {
				if string(oldContent) == st.Content {
					continue // Skip writing if content is identical
				}
			}
		}

		if err := os.WriteFile(fullPath, []byte(st.Content), 0644); err != nil {
			return false, fmt.Errorf("failed to write setting to workspace: %w", err)
		} else {
			actuallyChanged = true
			if st.Mtime > 0 {
				mt := time.UnixMilli(st.Mtime)
				_ = os.Chtimes(fullPath, mt, mt)
			}
		}
	}

	return actuallyChanged, nil
}

func (s *gitSyncService) processNotesBatch(notes []*domain.Note, wsPath string) (bool, error) {
	var actuallyChanged bool

	for _, n := range notes {
		targetPath := n.Path
		if filepath.Ext(targetPath) == "" {
			targetPath += ".md"
		}
		fullPath := filepath.Join(wsPath, targetPath)

		if n.Action == domain.NoteActionDelete {
			if _, err := os.Stat(fullPath); err == nil {
				_ = os.Remove(fullPath)
				actuallyChanged = true
			}
			continue
		}

		_ = os.MkdirAll(filepath.Dir(fullPath), 0755)

		// Check if content is different before writing
		if oldFile, err := os.Open(fullPath); err == nil {
			defer oldFile.Close()
			// For notes, we still compare strings as they are typically small
			// and this maintains simple logic for .md files.
			if oldContent, err := io.ReadAll(oldFile); err == nil {
				if string(oldContent) == n.Content {
					continue // Skip writing if content is identical
				}
			}
		}

		if err := os.WriteFile(fullPath, []byte(n.Content), 0644); err != nil {
			return false, fmt.Errorf("failed to write note to workspace: %w", err)
		} else {
			actuallyChanged = true
			if n.Mtime > 0 {
				mt := time.UnixMilli(n.Mtime)
				_ = os.Chtimes(fullPath, mt, mt)
			}
		}
	}

	return actuallyChanged, nil
}

func (s *gitSyncService) processFilesBatch(files []*domain.File, conf *domain.GitSyncConfig, wsPath string) (bool, error) {
	var actuallyChanged bool

	for _, f := range files {
		fullPath := filepath.Join(wsPath, f.Path)

		if f.Action == domain.FileActionDelete {
			if _, err := os.Stat(fullPath); err == nil {
				_ = os.Remove(fullPath)
				actuallyChanged = true
			}
			continue
		}

		_ = os.MkdirAll(filepath.Dir(fullPath), 0755)

		// Add physical file existence check to prevent source not existing from causing copyFileIfDifferent error interruption
		// 增加物理文件存在性检查，防止 src 不存在导致 copyFileIfDifferent 报错中断
		if _, err := os.Stat(f.SavePath); os.IsNotExist(err) {
			s.logger.Warn("Attachment file not found in storage, skipping mirror for this file",
				zap.Int64("uid", conf.UID),
				zap.Int64("vaultId", conf.VaultID),
				zap.String("path", f.Path),
				zap.String("savePath", f.SavePath))
			continue
		}

		copyChanged, err := s.copyFileIfDifferent(f.SavePath, fullPath)
		if err != nil {
			return false, fmt.Errorf("failed to copy attachment to workspace: %w", err)
		} else if copyChanged {
			actuallyChanged = true
			if f.Mtime > 0 {
				mt := time.UnixMilli(f.Mtime)
				_ = os.Chtimes(fullPath, mt, mt)
			}
		}
	}

	return actuallyChanged, nil
}

func (s *gitSyncService) copyFileIfDifferent(src, dst string) (bool, error) {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return false, err
	}

	dstInfo, err := os.Stat(dst)
	if err == nil {
		if srcInfo.Size() == dstInfo.Size() {
			// Sizes match, we could do deep comparison, but for sync service
			// relying on size and potentially mtime/hash in DB is safer and faster.
			// Here we assume if size matches, it's likely same (simplification to avoid full read).
			// If we really need deep compare, we should use streaming hash.
			return false, nil
		}
	}

	// Streaming copy
	srcFile, err := os.Open(src)
	if err != nil {
		return false, err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return false, err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *gitSyncService) NotifyUpdated(uid int64, vaultID int64) {
	s.logger.Debug("NotifyUpdated called", zap.Int64("uid", uid), zap.Int64("vaultID", vaultID))

	configs, err := s.repo.ListByVaultID(context.Background(), vaultID, uid)
	if err != nil {
		s.logger.Error("NotifyUpdated: failed to list configs by vaultID", zap.Int64("uid", uid), zap.Int64("vaultID", vaultID), zap.Error(err))
		return
	}

	s.logger.Debug("NotifyUpdated: found configs", zap.Int64("uid", uid), zap.Int64("vaultID", vaultID), zap.Int("count", len(configs)))

	for _, conf := range configs {
		if !conf.IsEnabled || conf.Delay <= 0 {
			s.logger.Debug("NotifyUpdated: skipping config", zap.Int64("configId", conf.ID), zap.Bool("isEnabled", conf.IsEnabled), zap.Int64("delay", conf.Delay))
			continue
		}

		s.mu.Lock()
		if timer, ok := s.timers[conf.ID]; ok {
			timer.Stop()
			s.logger.Debug("NotifyUpdated: reset existing timer", zap.Int64("configId", conf.ID))
		}

		id := conf.ID
		configUid := uid
		s.logger.Info("NotifyUpdated: scheduling delayed sync", zap.Int64("configId", id), zap.Int64("delay", conf.Delay))
		s.timers[id] = time.AfterFunc(time.Duration(conf.Delay)*time.Second, func() {
			s.mu.Lock()
			delete(s.timers, id)
			s.mu.Unlock()

			ctx := context.Background()
			_ = s.ExecuteSync(ctx, configUid, id)
		})
		s.mu.Unlock()
	}
}
