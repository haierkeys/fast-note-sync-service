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
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// GitSyncService 定义 Git 同步业务服务接口
type GitSyncService interface {
	GetConfigs(ctx context.Context, uid int64) ([]*dto.GitSyncConfigDTO, error)
	GetConfig(ctx context.Context, uid int64, vaultID int64) (*dto.GitSyncConfigDTO, error)
	UpdateConfig(ctx context.Context, uid int64, params *dto.GitSyncConfigRequest) (*dto.GitSyncConfigDTO, error)
	DeleteConfig(ctx context.Context, uid int64, id int64) error
	Validate(ctx context.Context, params *dto.GitSyncValidateRequest) error
	ExecuteSync(ctx context.Context, uid int64, id int64) error
	CleanWorkspace(ctx context.Context, uid int64, configID int64) error
	NotifyUpdated(uid int64)
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
}

// NewGitSyncService 创建 GitSyncService 实例
func NewGitSyncService(repo domain.GitSyncRepository, noteRepo domain.NoteRepository, folderRepo domain.FolderRepository, fileRepo domain.FileRepository, vaultRepo domain.VaultRepository, logger *zap.Logger) GitSyncService {
	return &gitSyncService{
		repo:       repo,
		noteRepo:   noteRepo,
		folderRepo: folderRepo,
		fileRepo:   fileRepo,
		vaultRepo:  vaultRepo,
		logger:     logger,
		running:    make(map[int64]bool),
		timers:     make(map[int64]*time.Timer),
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
	// If vault name is provided, find vault ID
	var vaultID int64
	if params.Vault != "" {
		v, err := s.vaultRepo.GetByName(ctx, params.Vault, uid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, code.ErrorVaultNotFound
			}
			return nil, code.ErrorDBQuery.WithDetails(err.Error())
		}
		vaultID = v.ID
	}

	conf := &domain.GitSyncConfig{
		ID:        params.ID,
		UID:       uid,
		VaultID:   vaultID,
		RepoURL:   params.RepoURL,
		Username:  params.Username,
		Password:  params.Password,
		Branch:    params.Branch,
		IsEnabled: params.IsEnabled,
		Delay:     params.Delay,
	}
	if conf.Branch == "" {
		conf.Branch = "main"
	}

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
		return code.ErrorInvalidParams.WithDetails("Git validation failed: " + err.Error())
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
		return code.ErrorInvalidParams.WithDetails("Branch not found in remote")
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

	// Run in background
	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.running, id)
			s.mu.Unlock()
		}()

		bgCtx := context.Background()
		s.syncTask(bgCtx, conf)
	}()

	return nil
}

func (s *gitSyncService) CleanWorkspace(ctx context.Context, uid int64, configID int64) error {
	path := s.getWorkspacePath(uid, configID)
	err := os.RemoveAll(path)
	if err != nil {
		return code.ErrorServerInternal.WithDetails("Failed to cleanup workspace: " + err.Error())
	}
	return nil
}

// Internal methods

func (s *gitSyncService) getWorkspacePath(uid, configID int64) string {
	return filepath.Join("storage", "git_workspace", fmt.Sprintf("%d", uid), fmt.Sprintf("%d", configID))
}

func (s *gitSyncService) syncTask(ctx context.Context, conf *domain.GitSyncConfig) {
	s.logger.Info("Starting Git sync task", zap.Int64("configId", conf.ID), zap.Int64("uid", conf.UID))

	conf.LastStatus = 1 // Running
	_, _ = s.repo.Save(ctx, conf, conf.UID)

	err := s.doSync(ctx, conf)

	if err != nil {
		s.logger.Error("Git sync task failed", zap.Int64("configId", conf.ID), zap.Error(err))
		conf.LastStatus = 3 // Failed
		conf.LastMessage = err.Error()
	} else {
		s.logger.Info("Git sync task success", zap.Int64("configId", conf.ID))
		conf.LastStatus = 2 // Success
		conf.LastMessage = "Sync completed at " + time.Now().Format("2006-01-02 15:04:05")
		now := time.Now()
		conf.LastSyncTime = &now
	}

	_, _ = s.repo.Save(ctx, conf, conf.UID)
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
	err = s.mirrorNotesToWorkspace(ctx, conf, wsPath)
	if err != nil {
		return fmt.Errorf("mirror to workspace failed: %w", err)
	}

	// 4. Commit and Push
	status, err := wt.Status()
	if err != nil {
		return err
	}
	if status.IsClean() {
		s.logger.Info("No changes to commit", zap.Int64("configId", conf.ID))
		return nil
	}

	err = wt.AddWithOptions(&git.AddOptions{All: true})
	if err != nil {
		return err
	}

	_, err = wt.Commit("Update from Sync Service", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Sync Service",
			Email: "sync@haierkeys.com",
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

func (s *gitSyncService) mirrorNotesToWorkspace(ctx context.Context, conf *domain.GitSyncConfig, wsPath string) error {
	v, err := s.vaultRepo.GetByID(ctx, conf.VaultID, conf.UID)
	if err != nil {
		return err
	}
	if v == nil {
		return fmt.Errorf("vault not found")
	}

	// 1. Mirror Notes
	notes, err := s.noteRepo.List(ctx, v.ID, 1, 1000000, conf.UID, "", false, "", false, "", "")
	if err != nil {
		return err
	}
	for _, n := range notes {
		targetPath := n.Path
		if filepath.Ext(targetPath) == "" {
			targetPath += ".md"
		}
		fullPath := filepath.Join(wsPath, targetPath)

		if n.Action == "delete" {
			_ = os.Remove(fullPath)
			continue
		}

		_ = os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err := os.WriteFile(fullPath, []byte(n.Content), 0644); err != nil {
			s.logger.Warn("Failed to write note to workspace", zap.String("path", n.Path), zap.Error(err))
		}
	}

	// 2. Mirror Files (Attachments)
	files, err := s.fileRepo.List(ctx, v.ID, 1, 1000000, conf.UID, "", false, "", "")
	if err != nil {
		return err
	}
	for _, f := range files {
		fullPath := filepath.Join(wsPath, f.Path)

		if f.Action == domain.FileActionDelete {
			_ = os.Remove(fullPath)
			continue
		}

		_ = os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err := s.copyFile(f.SavePath, fullPath); err != nil {
			s.logger.Warn("Failed to copy attachment to workspace", zap.String("path", f.Path), zap.Error(err))
		}
	}

	return nil
}

func (s *gitSyncService) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func (s *gitSyncService) NotifyUpdated(uid int64) {
	configs, err := s.repo.List(context.Background(), uid)
	if err != nil {
		return
	}

	for _, conf := range configs {
		if !conf.IsEnabled || conf.Delay <= 0 {
			continue
		}

		s.mu.Lock()
		if timer, ok := s.timers[conf.ID]; ok {
			timer.Stop()
		}

		id := conf.ID
		configUid := uid
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
