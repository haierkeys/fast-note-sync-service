package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// errNoChanges 表示 Git 同步检查后没有发现任何需要提交的变更
var errNoChanges = errors.New("no changes found")

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
	repo       domain.GitSyncRepository
	noteRepo   domain.NoteRepository
	folderRepo domain.FolderRepository
	fileRepo   domain.FileRepository
	vaultRepo  domain.VaultRepository
	logger     *zap.Logger
	mu         sync.Mutex
	running    map[int64]bool        // configID -> isRunning
	timers     map[int64]*time.Timer // configID -> timer
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// NewGitSyncService 创建 GitSyncService 实例
func NewGitSyncService(repo domain.GitSyncRepository, noteRepo domain.NoteRepository, folderRepo domain.FolderRepository, fileRepo domain.FileRepository, vaultRepo domain.VaultRepository, logger *zap.Logger) GitSyncService {
	ctx, cancel := context.WithCancel(context.Background())
	return &gitSyncService{
		repo:       repo,
		noteRepo:   noteRepo,
		folderRepo: folderRepo,
		fileRepo:   fileRepo,
		vaultRepo:  vaultRepo,
		logger:     logger,
		running:    make(map[int64]bool),
		timers:     make(map[int64]*time.Timer),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (s *gitSyncService) domainToDTO(conf *domain.GitSyncConfig) *dto.GitSyncConfigDTO {
	if conf == nil {
		return nil
	}
	res := &dto.GitSyncConfigDTO{
		ID:          conf.ID,
		UID:         conf.UID,
		RepoURL:     conf.RepoURL,
		Username:    conf.Username,
		Password:    conf.Password,
		Branch:      conf.Branch,
		IsEnabled:   conf.IsEnabled,
		Delay:       conf.Delay,
		LastStatus:  conf.LastStatus,
		LastMessage: conf.LastMessage,
		CreatedAt:   timex.Time(conf.CreatedAt),
		UpdatedAt:   timex.Time(conf.UpdatedAt),
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

	saved, err := s.repo.Save(ctx, conf, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	return s.domainToDTO(saved), nil
}

func (s *gitSyncService) DeleteConfig(ctx context.Context, uid int64, id int64) error {
	// Check identity
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

	auth := &http.BasicAuth{
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

	// Mark as running
	s.mu.Lock()
	if s.running[id] {
		s.mu.Unlock()
		return code.ErrorGitSyncTaskRunning
	}
	s.running[id] = true
	s.mu.Unlock()

	s.wg.Add(1)
	// Run in background
	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.running, id)
			s.mu.Unlock()
			s.wg.Done()
		}()

		// Use service context for background task
		s.syncTask(s.ctx, conf)
	}()

	return nil
}

func (s *gitSyncService) CleanWorkspace(ctx context.Context, uid int64, configID int64) error {
	if configID > 0 {
		// 1. Reset database fields
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

		// 2. Delete History
		_ = s.repo.DeleteHistory(ctx, uid, configID)

		// 3. Remove physical workspace
		path := s.getWorkspacePath(uid, configID)
		err = os.RemoveAll(path)
		if err != nil {
			s.logger.Warn("Failed to cleanup physical workspace", zap.String("path", path), zap.Error(err))
		}
	} else {
		// 1. Reset all database fields for user
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

		// 2. Delete All History for user
		_ = s.repo.DeleteHistory(ctx, uid, 0)

		// 3. Remove all physical workspaces for user
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

// Internal methods

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

	// 记录运行前的状态，以便无变更时恢复
	prevStatus := conf.LastStatus

	// Update Config Status to Running
	conf.LastStatus = domain.GitSyncStatusRunning
	_, _ = s.repo.Save(ctx, conf, conf.UID)

	err := s.doSync(ctx, conf)

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

	// 自动清理过期历史记录
	if conf.RetentionDays != 0 {
		var cutoffTime time.Time
		if conf.RetentionDays == -1 {
			// -1 表示仅保留当前最新的一条记录
			cutoffTime = startTime
		} else if conf.RetentionDays > 0 {
			// > 0 表示清理超过指定天数的记录
			cutoffTime = time.Now().AddDate(0, 0, -int(conf.RetentionDays))
		}

		if !cutoffTime.IsZero() {
			if err := s.repo.DeleteOldHistory(context.Background(), conf.UID, conf.ID, cutoffTime); err != nil {
				s.logger.Error("Failed to delete old git sync history", zap.Error(err), zap.Int64("configId", conf.ID))
			}
		}
	}
}

func (s *gitSyncService) doSync(ctx context.Context, conf *domain.GitSyncConfig) error {
	wsPath := s.getWorkspacePath(conf.UID, conf.ID)
	auth := &http.BasicAuth{
		Username: conf.Username,
		Password: conf.Password,
	}

	var r *git.Repository
	var err error

	// 1. Check/Init Local Repo
	if _, err := os.Stat(filepath.Join(wsPath, ".git")); os.IsNotExist(err) {
		s.logger.Info("Initializing local git repo", zap.String("path", wsPath))
		_ = os.RemoveAll(wsPath)
		r, err = git.PlainClone(wsPath, false, &git.CloneOptions{
			URL:           conf.RepoURL,
			Auth:          auth,
			ReferenceName: plumbing.NewBranchReferenceName(conf.Branch),
			SingleBranch:  true,
		})
		if err != nil {
			return fmt.Errorf("git clone failed: %w", err)
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

	// 2. Pull latest
	s.logger.Info("Pulling latest changes", zap.Int64("configId", conf.ID))
	err = wt.Pull(&git.PullOptions{
		Auth:          auth,
		ReferenceName: plumbing.NewBranchReferenceName(conf.Branch),
		SingleBranch:  true,
		Force:         true,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("git pull failed: %w", err)
	}

	// 3. Extract DB content to Workspace
	// We need to mirror files from DB to this workspace
	changed, err := s.mirrorNotesToWorkspace(ctx, conf, wsPath, conf.LastSyncTime)
	if err != nil {
		return fmt.Errorf("mirror to workspace failed: %w", err)
	}

	if !changed {
		s.logger.Info("No notes or attachments updated, skipping Git operations", zap.Int64("configId", conf.ID))
		return errNoChanges
	}

	// 4. Commit and Push
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

	_, err = wt.Commit("Update from Sync Service", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "FNS Service",
			Email: "fns@email.com",
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

	notes, err := s.noteRepo.ListByUpdatedTimestamp(ctx, ts, v.ID, conf.UID)
	if err != nil {
		return false, err
	}

	files, err := s.fileRepo.ListByUpdatedTimestamp(ctx, ts, v.ID, conf.UID)
	if err != nil {
		return false, err
	}

	if len(notes) == 0 && len(files) == 0 {
		return false, nil
	}

	var actuallyChanged bool

	// 1. Process Notes
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
		if oldContent, err := os.ReadFile(fullPath); err == nil {
			if string(oldContent) == n.Content {
				continue // Skip writing if content is identical
			}
		}

		if err := os.WriteFile(fullPath, []byte(n.Content), 0644); err != nil {
			s.logger.Warn("Failed to write note to workspace", zap.String("path", n.Path), zap.Error(err))
		} else {
			actuallyChanged = true
			if n.Mtime > 0 {
				mt := time.UnixMilli(n.Mtime)
				_ = os.Chtimes(fullPath, mt, mt)
			}
		}
	}

	// 2. Process Files
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

		copyChanged, err := s.copyFileIfDifferent(f.SavePath, fullPath)
		if err != nil {
			s.logger.Warn("Failed to copy attachment to workspace", zap.String("path", f.Path), zap.Error(err))
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
			// Sizes match, do a byte-by-byte comparison or checksum if needed.
			// For simplicity and speed in this context, byte-by-byte for small files or just relying on size + mtime might be okay,
			// but to be sure we'll do a quick content check.
			srcData, err := os.ReadFile(src)
			if err != nil {
				return false, err
			}
			dstData, err := os.ReadFile(dst)
			if err != nil {
				// If dst is unreadable, treat as different
				goto doCopy
			}
			if string(srcData) == string(dstData) {
				return false, nil
			}
		}
	}

doCopy:
	data, err := os.ReadFile(src)
	if err != nil {
		return false, err
	}
	err = os.WriteFile(dst, data, 0644)
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
