// Package service implements the business logic layer
// Package service 实现业务逻辑层
package service

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/shortlink"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"go.uber.org/zap"
)

var (
	attachmentRegex    = regexp.MustCompile(`!\[\[(.*?)\]\]`)
	markdownImageRegex = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	htmlImageRegex     = regexp.MustCompile(`(?i)<img\b([^>]*?)\bsrc\s*=\s*(['"])(.*?)['"]([^>]*)>`)
)

// ShareService defines the share business service interface
// ShareService 定义分享业务服务接口
type ShareService interface {
	// ShareGenerate generates and stores share token
	// ShareGenerate 生成并存储分享 Token
	ShareGenerate(ctx context.Context, uid int64, vaultName string, path string, pathHash string, password string) (*dto.ShareCreateResponse, error)

	// VerifyShare verifies share token and its status
	// VerifyShare 验证分享 Token 及其状态
	VerifyShare(ctx context.Context, token string, rid string, rtp string, password string) (*pkgapp.ShareEntity, error)

	// GetSharedNote retrieves shared note details
	// GetSharedNote 获取分享的单条笔记详情
	GetSharedNote(ctx context.Context, shareToken string, noteID int64, password string) (*dto.NoteDTO, error)

	// GetSharedFile retrieves shared file content
	// GetSharedFile 获取分享的文件内容
	GetSharedFile(ctx context.Context, shareToken string, fileID int64, password string) (content []byte, contentType string, mtime int64, etag string, fileName string, err error)

	// GetSharedFileInfo retrieves shared file metadata and path for zero-copy download
	// GetSharedFileInfo 获取分享文件的元数据和路径，用于零拷贝下载
	GetSharedFileInfo(ctx context.Context, shareToken string, fileID int64, password string) (savePath string, contentType string, mtime int64, etag string, fileName string, err error)

	// RecordView aggregates access statistics in memory
	// RecordView 在内存中聚合访问统计
	RecordView(uid int64, id int64)

	// StopShare revokes a share
	// StopShare 撤销分享
	StopShare(ctx context.Context, uid int64, id int64) error

	// UpdateSharePassword updates password for a share based on resource info
	// UpdateSharePassword 根据资源信息更新分享密码
	UpdateSharePassword(ctx context.Context, uid int64, vaultName string, path string, pathHash string, password string) error

	// CreateShortLink generates a short link for a share
	// CreateShortLink 为分享生成短链
	CreateShortLink(ctx context.Context, uid int64, vaultName string, path string, pathHash string, baseURL string, longURL string, isForce bool) (string, error)

	// ListShares lists all shares of a user with sorting and pagination
	// ListShares 列出用户的所有分享（支持排序和分页）
	ListShares(ctx context.Context, uid int64, sortBy string, sortOrder string, pager *pkgapp.Pager) ([]*dto.ShareListItem, int, error)

	// GetShareByPath retrieves share info by path
	// GetShareByPath 根据路径获取分享信息
	GetShareByPath(ctx context.Context, uid int64, vaultName string, pathHash string) (*domain.UserShare, error)

	// StopShareByPath revokes a share by path
	// StopShareByPath 根据路径撤销分享
	StopShareByPath(ctx context.Context, uid int64, vaultName string, pathHash string) error

	// GetActiveNotePathsByVault returns active shared note paths for a vault
	// GetActiveNotePathsByVault 返回指定 vault 下所有有效分享的笔记路径列表
	GetActiveNotePathsByVault(ctx context.Context, uid int64, vaultName string) ([]string, error)

	// GetNoteShareChangesByVault returns share path changes since sinceMs for a vault
	// GetNoteShareChangesByVault 返回指定 vault 下 sinceMs 之后的分享路径变更
	GetNoteShareChangesByVault(ctx context.Context, uid int64, vaultName string, sinceMs int64) (*dto.ShareChangesData, error)

	// Shutdown shuts down the service and flushes remaining data
	// Shutdown 关闭服务并同步最后的数据
	Shutdown(ctx context.Context) error
}

// aggStats aggregated statistics
// aggStats 聚合统计
type aggStats struct {
	uid          int64     // User ID // 用户 ID
	viewCount    int64     // View count // 访问计数
	lastViewedAt time.Time // Last viewed at // 最后访问时间
}

// shareService implementation of ShareService interface
// shareService 实现 ShareService 接口
type shareService struct {
	repo         domain.UserShareRepository // Share repository // 分享仓库
	tokenManager pkgapp.TokenManager        // Token manager // Token 管理器
	noteRepo     domain.NoteRepository      // Note repository // 笔记仓库
	fileRepo     domain.FileRepository      // File repository // 文件仓库
	vaultRepo    domain.VaultRepository     // Vault repository // 仓库仓库
	logger       *zap.Logger                // Logger // 日志器
	config       *ServiceConfig             // Service configuration // 服务配置

	// Statistics buffer
	// 统计缓冲区
	bufferMu    sync.Mutex          // Buffer mutex // 缓冲区互斥锁
	statsBuffer map[int64]*aggStats // Stats buffer // 统计缓冲区
	ticker      *time.Ticker        // Sync ticker // 同步定时器
	stopCh      chan struct{}       // Stop channel // 停止信号
	doneCh      chan struct{}       // Done channel // 完成信号
}

// NewShareService creates ShareService instance
// NewShareService 创建 ShareService 实例
func NewShareService(repo domain.UserShareRepository, tokenManager pkgapp.TokenManager, noteRepo domain.NoteRepository, fileRepo domain.FileRepository, vaultRepo domain.VaultRepository, logger *zap.Logger, config *ServiceConfig) ShareService {
	s := &shareService{
		repo:         repo,
		tokenManager: tokenManager,
		noteRepo:     noteRepo,
		fileRepo:     fileRepo,
		vaultRepo:    vaultRepo,
		logger:       logger,
		config:       config,
		statsBuffer:  make(map[int64]*aggStats),
		ticker:       time.NewTicker(5 * time.Minute),
		stopCh:       make(chan struct{}),
		doneCh:       make(chan struct{}),
	}

	go s.startFlushLoop()

	return s
}

// ShareGenerate generates and stores share token
// ShareGenerate 生成并存储分享 Token
func (s *shareService) ShareGenerate(ctx context.Context, uid int64, vaultName string, path string, pathHash string, password string) (*dto.ShareCreateResponse, error) {
	// 1. Get VaultID
	// 1. 获取 VaultID
	vault, err := s.vaultRepo.GetByName(ctx, vaultName, uid)
	if err != nil {
		return nil, err
	}
	vaultID := vault.ID

	var resolvedResources = make(map[string][]string)
	var mainID int64
	var mainType string

	// 2. Determine type based on suffix
	// 2. 根据后缀判定类型
	isNote := strings.HasSuffix(strings.ToLower(path), ".md")

	if isNote {
		// Try looking up as Note
		// 尝试作为 Note 查找
		note, err := s.noteRepo.GetByPathHash(ctx, pathHash, vaultID, uid)
		if err == nil && note != nil && note.Action != domain.NoteActionDelete {
			mainID = note.ID
			mainType = "note"
			noteIDStr := strconv.FormatInt(note.ID, 10)
			resolvedResources["note"] = []string{noteIDStr}
			fileRefs, err := s.resolveSharedNoteFiles(ctx, uid, note.VaultID, note.Path, note.Content)
			if err != nil {
				s.logger.Warn("ShareGenerate resolveSharedNoteFiles failed", zap.Error(err), zap.String("notePath", note.Path))
			} else {
				for _, file := range fileRefs {
					fileIDStr := strconv.FormatInt(file.ID, 10)
					// Avoid duplicate authorization
					// 避免重复授权
					if !util.Inarray(resolvedResources["file"], fileIDStr) {
						resolvedResources["file"] = append(resolvedResources["file"], fileIDStr)
					}
				}
			}
		} else {
			return nil, code.ErrorNoteNotFound.WithDetails("note not found: " + path)
		}
	} else {
		// Try looking up as File
		// 尝试作为 File 查找
		file, err := s.fileRepo.GetByPathHash(ctx, pathHash, vaultID, uid)
		if err == nil && file != nil && file.Action != domain.FileActionDelete {
			mainID = file.ID
			mainType = "file"
			fileIDStr := strconv.FormatInt(file.ID, 10)
			resolvedResources["file"] = []string{fileIDStr}
		} else {
			return nil, code.ErrorFileNotFound.WithDetails("file not found: " + path)
		}
	}

	// 3. Determine expiration time
	// 3. 确定过期时间
	expiry := 30 * 24 * time.Hour // Default 30 days // 默认 30 天
	if s.config != nil && s.config.App.ShareTokenExpiry != "" {
		if d, err := util.ParseDuration(s.config.App.ShareTokenExpiry); err == nil {
			expiry = d
		}
	}
	expiresAt := time.Now().Add(expiry)

	pwdMd5 := ""
	if password != "" {
		pwdMd5 = util.EncodeMD5(password)
	}

	share := &domain.UserShare{
		UID:       uid,
		ResType:   mainType,
		ResID:     mainID,
		Resources: resolvedResources,
		Status:    1,
		ExpiresAt: expiresAt,
		Password:  pwdMd5,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 幂等：若该资源已有 active 分享，先撤销，避免重复计数
	// Idempotent: revoke any existing active share before creating a new one
	if existing, err := s.repo.GetByRes(ctx, uid, mainType, mainID); err == nil && existing != nil {
		_ = s.StopShare(ctx, uid, existing.ID)
	}

	if err := s.repo.Create(ctx, uid, share); err != nil {
		return nil, err
	}

	// 4. Generate Token (using underlying SID encryption scheme)
	// 4. 生成 Token (使用底层 SID 加密方案)
	token, err := s.tokenManager.ShareGenerate(share.ID, uid, resolvedResources)
	if err != nil {
		return nil, err
	}

	return &dto.ShareCreateResponse{
		ID:         mainID,
		Type:       mainType,
		Token:      token,
		IsPassword: pwdMd5 != "",
		ExpiresAt:  expiresAt,
		ShortLink:  share.ShortLink,
	}, nil
}

// VerifyShare verifies share token and its status
// VerifyShare 验证分享 Token 及其状态
func (s *shareService) VerifyShare(ctx context.Context, token string, rid string, rtp string, password string) (*pkgapp.ShareEntity, error) {
	entity, err := s.tokenManager.ShareParse(token)

	if err != nil {
		return nil, err
	}

	share, err := s.repo.GetByID(ctx, entity.UID, entity.SID)

	if err != nil {
		return nil, err
	}

	if share.Status != 1 {
		return nil, domain.ErrShareCancelled
	}

	// 增加密码校验逻辑
	if share.Password != "" {
		if password == "" {
			return nil, domain.ErrSharePasswordRequired
		}
		if util.EncodeMD5(password) != share.Password {
			return nil, domain.ErrSharePasswordInvalid
		}
	}

	entity.Resources = share.Resources

	ids, ok := share.Resources[rtp]
	if !ok {
		return nil, domain.ErrShareCancelled // Match type mismatch // 资源类型不匹配
	}

	authorized := false
	for _, id := range ids {
		if id == rid {
			authorized = true
			break
		}
	}

	if !authorized {
		return nil, domain.ErrShareCancelled // Resource not authorized // 资源未授权
	}

	// In-memory record access statistics (delayed 5 minutes update)
	// 内存记录访问统计 (延迟 5 分钟更新)
	s.RecordView(share.UID, share.ID)

	return entity, nil
}

// RecordView aggregates access statistics in memory
// RecordView 在内存中聚合访问统计
func (s *shareService) RecordView(uid int64, id int64) {
	s.bufferMu.Lock()
	defer s.bufferMu.Unlock()

	stats, ok := s.statsBuffer[id]
	if !ok {
		stats = &aggStats{
			uid: uid,
		}
		s.statsBuffer[id] = stats
	}
	stats.viewCount++
	stats.lastViewedAt = time.Now()
}

// startFlushLoop starts periodic synchronization goroutine
// startFlushLoop 启动定时同步协程
func (s *shareService) startFlushLoop() {
	defer close(s.doneCh)
	for {
		select {
		case <-s.ticker.C:
			s.flush()
		case <-s.stopCh:
			s.flush()
			return
		}
	}
}

// flush synchronizes incremental totals in memory to database
// flush 将内存中的增量合计同步到数据库
func (s *shareService) flush() {
	s.bufferMu.Lock()
	if len(s.statsBuffer) == 0 {
		s.bufferMu.Unlock()
		return
	}
	tempBuffer := s.statsBuffer
	s.statsBuffer = make(map[int64]*aggStats)
	s.bufferMu.Unlock()

	ctx := context.Background()
	for id, stats := range tempBuffer {
		if err := s.repo.UpdateViewStats(ctx, stats.uid, id, stats.viewCount, stats.lastViewedAt); err != nil {
			s.logger.Error("failed to flush user_share stats", zap.Int64("id", id), zap.Error(err))
		}
	}
}

// StopShare revokes a share
// StopShare 撤销分享
func (s *shareService) StopShare(ctx context.Context, uid int64, id int64) error {
	return s.repo.UpdateStatus(ctx, uid, id, domain.UserShareStatusRevoked)
}

// UpdateSharePassword updates password for a share
// UpdateSharePassword 更新分享密码
func (s *shareService) UpdateSharePassword(ctx context.Context, uid int64, vaultName string, path string, pathHash string, password string) error {
	// 1. Get VaultID
	vault, err := s.vaultRepo.GetByName(ctx, vaultName, uid)
	if err != nil {
		return err
	}
	if vault == nil {
		return code.ErrorVaultNotFound
	}

	// 2. Get UserShare by resource
	share, err := s.repo.GetByPath(ctx, uid, vault.ID, pathHash)
	if err != nil {
		return err
	}
	if share == nil {
		return code.ErrorShareNotFound
	}

	pwdMd5 := ""
	if password != "" {
		pwdMd5 = util.EncodeMD5(password)
	}
	return s.repo.UpdatePassword(ctx, uid, share.ID, pwdMd5)
}

// ListShares lists all shares of a user with sorting, pagination and fills in resource titles
// ListShares 列出用户的所有分享，支持排序、分页，并填充资源标题
func (s *shareService) ListShares(ctx context.Context, uid int64, sortBy string, sortOrder string, pager *pkgapp.Pager) ([]*dto.ShareListItem, int, error) {
	// 1. 获取总数
	count64, err := s.repo.CountByUID(ctx, uid)
	if err != nil {
		return nil, 0, err
	}
	count := int(count64)

	if count == 0 {
		return []*dto.ShareListItem{}, 0, nil
	}

	// 2. 获取分页数据
	shares, err := s.repo.ListByUID(ctx, uid, sortBy, sortOrder, pkgapp.GetPageOffset(pager.Page, pager.PageSize), pager.PageSize)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*dto.ShareListItem, 0, len(shares))

	// 批量收集 noteIDs 和 fileIDs，避免 N+1 查询
	// Collect noteIDs and fileIDs in bulk to avoid N+1 queries
	var noteIDs, fileIDs []int64
	for _, share := range shares {
		switch share.ResType {
		case "note":
			noteIDs = append(noteIDs, share.ResID)
		case "file":
			fileIDs = append(fileIDs, share.ResID)
		}
	}

	// 批量查询 notes，建立 id→note 映射
	// Batch query notes and build id→note map
	noteMap := make(map[int64]*domain.Note)
	if len(noteIDs) > 0 {
		notes, err := s.noteRepo.ListByIDs(ctx, noteIDs, uid)
		if err != nil {
			s.logger.Warn("ListShares: batch query notes failed", zap.Error(err))
		} else {
			for _, n := range notes {
				noteMap[n.ID] = n
			}
		}
	}

	fileMap := make(map[int64]*domain.File)
	if len(fileIDs) > 0 {
		files, err := s.fileRepo.ListByIDs(ctx, fileIDs, uid)
		if err != nil {
			s.logger.Warn("ListShares: batch query files failed", zap.Error(err))
		} else {
			for _, f := range files {
				fileMap[f.ID] = f
			}
		}
	}

	// 查询 vault 名称（vault 数量极少，按需缓存）
	// Query vault names on demand with local cache (vault count is always small)
	vaultNameCache := make(map[int64]string)

	for _, share := range shares {
		item := &dto.ShareListItem{
			ID:           share.ID,
			UID:          share.UID,
			Resources:    share.Resources,
			Status:       share.Status,
			ViewCount:    share.ViewCount,
			LastViewedAt: share.LastViewedAt,
			ExpiresAt:    share.ExpiresAt,
			ShortLink:    share.ShortLink,
			CreatedAt:    share.CreatedAt,
			UpdatedAt:    share.UpdatedAt,
			IsPassword:   share.Password != "",
		}

		// 生成 Token 并拼接 URL /ResID/token
		// Generate Token and concatenate URL /ResID/token
		token, err := s.tokenManager.ShareGenerate(share.ID, uid, share.Resources)
		if err == nil {
			item.URL = "/share/" + strconv.FormatInt(share.ResID, 10) + "/" + token
		}

		// 从预加载的 map 中回填标题，无额外查询
		// Fill title from preloaded maps, no extra DB queries
		switch share.ResType {
		case "note":
			if note, ok := noteMap[share.ResID]; ok && note.Action != domain.NoteActionDelete {
				item.Title = strings.TrimSuffix(filepath.Base(note.Path), ".md")
				item.NotePath = note.Path
				if name, ok := vaultNameCache[note.VaultID]; ok {
					item.VaultName = name
				} else if v, err := s.vaultRepo.GetByID(ctx, note.VaultID, uid); err == nil && v != nil {
					vaultNameCache[note.VaultID] = v.Name
					item.VaultName = v.Name
				}
			}
		case "file":
			if file, ok := fileMap[share.ResID]; ok {
				item.Title = filepath.Base(file.Path)
			}
		}

		items = append(items, item)
	}

	return items, count, nil
}

// GetShareByPath retrieves share info by path
// GetShareByPath 根据路径获取分享信息
func (s *shareService) GetShareByPath(ctx context.Context, uid int64, vaultName string, pathHash string) (*domain.UserShare, error) {
	vault, err := s.vaultRepo.GetByName(ctx, vaultName, uid)
	if err != nil {
		return nil, err
	}

	resID := int64(0)
	resType := ""

	// Check if it's a note or file
	note, err := s.noteRepo.GetByPathHash(ctx, pathHash, vault.ID, uid)
	if err == nil && note != nil {
		resID = note.ID
		resType = "note"
	} else {
		file, err := s.fileRepo.GetByPathHash(ctx, pathHash, vault.ID, uid)
		if err == nil && file != nil {
			resID = file.ID
			resType = "file"
		}
	}

	if resID == 0 {
		return nil, code.ErrorFileNotFound
	}

	// Use precision index query instead of iterating list
	// 使用精确索引查询，替代遍历列表
	share, err := s.repo.GetByRes(ctx, uid, resType, resID)
	if err != nil {
		return nil, code.ErrorFileNotFound // No active share found
	}

	return share, nil
}

// CreateShortLink generates a short link for a share record
// CreateShortLink 为分享记录生成短链
func (s *shareService) CreateShortLink(ctx context.Context, uid int64, vaultName string, path string, pathHash string, baseURL string, longURL string, isForce bool) (string, error) {
	// Find vault first to get ID
	vault, err := s.vaultRepo.GetByName(ctx, vaultName, uid)
	if err != nil {
		return "", err
	}
	if vault == nil {
		return "", code.ErrorVaultNotFound
	}

	// Find existing share record by path using vault.ID
	share, err := s.repo.GetByPath(ctx, uid, vault.ID, pathHash)
	if err != nil {
		return "", err
	}

	if share == nil {
		return "", code.ErrorShareNotFound
	}

	// If short link already exists and not forcing regeneration, return it
	if !isForce && share.ShortLink != "" {
		return share.ShortLink, nil
	}

	// Prepare short link creation parameters from service config
	sinkBaseURL := s.config.App.ShortLink.BaseURL
	apiKey := s.config.App.ShortLink.APIKey
	password := s.config.App.ShortLink.Password
	cloaking := s.config.App.ShortLink.Cloaking

	// expiration matches the share record
	expiresAt := share.ExpiresAt

	// Use client-provided URL if available; otherwise fall back to generating one
	// 优先使用客户端传入的完整分享 URL，避免因重新生成 token 导致 URL 不一致
	if longURL == "" {
		token, err := s.tokenManager.ShareGenerate(share.ID, uid, share.Resources)
		if err != nil {
			return "", err
		}
		longURL = fmt.Sprintf("%s/share/%d/%s", strings.TrimRight(baseURL, "/"), share.ResID, token)
	}

	client := shortlink.NewSinkCoolClient(sinkBaseURL, apiKey)

	title := ""
	switch share.ResType {
	case "note":
		if note, err := s.noteRepo.GetByID(ctx, share.ResID, uid); err == nil && note != nil {
			title = strings.TrimSuffix(filepath.Base(note.Path), ".md")
		}
	case "file":
		if file, err := s.fileRepo.GetByID(ctx, share.ResID, uid); err == nil && file != nil {
			title = filepath.Base(file.Path)
		}
	}

	shortURL, err := client.Create(longURL, expiresAt, password, cloaking, title)

	if err != nil {
		return "", err
	}

	// Save the generated short link back to database
	if err := s.repo.UpdateShortLink(ctx, uid, share.ID, shortURL); err != nil {
		return "", err
	}

	return shortURL, nil
}

// StopShareByPath revokes a share by path
// StopShareByPath 根据路径撤销分享
func (s *shareService) StopShareByPath(ctx context.Context, uid int64, vaultName string, pathHash string) error {
	share, err := s.GetShareByPath(ctx, uid, vaultName, pathHash)
	if err != nil {
		return err
	}
	return s.StopShare(ctx, uid, share.ID)
}

// GetActiveNotePathsByVault returns active shared note paths for a vault
// GetActiveNotePathsByVault 返回指定 vault 下所有有效分享的笔记路径（两步查询，避免跨库 JOIN）
func (s *shareService) GetActiveNotePathsByVault(ctx context.Context, uid int64, vaultName string) ([]string, error) {
	vault, err := s.vaultRepo.GetByName(ctx, vaultName, uid)
	if err != nil {
		return nil, err
	}

	// Step 1: query active note res_ids from user_shares DB (no cross-DB JOIN)
	// 步骤1：从 user_shares 库查出有效分享的 note res_id 列表（不做跨库 JOIN）
	noteIDs, err := s.repo.ListActiveNoteResIDs(ctx, uid)
	if err != nil {
		return nil, err
	}
	if len(noteIDs) == 0 {
		return []string{}, nil
	}

	// Step 2: batch query notes from notes DB, filter by vault and non-deleted action
	// 步骤2：从 notes 库批量查笔记，按 vault 和非删除状态过滤
	notes, err := s.noteRepo.ListByIDs(ctx, noteIDs, uid)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(notes))
	for _, n := range notes {
		if n.VaultID == vault.ID && n.Action != domain.NoteActionDelete {
			paths = append(paths, n.Path)
		}
	}
	return paths, nil
}

// GetNoteShareChangesByVault 返回指定 vault 下 sinceMs 之后的分享路径变更（两步查询，避免跨库 JOIN）
// GetNoteShareChangesByVault returns note share path changes since sinceMs for a vault (two-step query, no cross-DB JOIN)
func (s *shareService) GetNoteShareChangesByVault(ctx context.Context, uid int64, vaultName string, sinceMs int64) (*dto.ShareChangesData, error) {
	// since=0 表示客户端无本地缓存，需要全量刷新 / since=0 means client has no cache, require full refresh
	if sinceMs == 0 {
		return &dto.ShareChangesData{
			FullRefreshRequired: true,
			Added:               []string{},
			Removed:             []string{},
			LastTime:            time.Now().UnixMilli(),
		}, nil
	}

	vault, err := s.vaultRepo.GetByName(ctx, vaultName, uid)
	if err != nil || vault == nil {
		return nil, code.ErrorVaultNotFound
	}

	since := time.UnixMilli(sinceMs)

	// 步骤1：查询 user_shares 中变更的 note res_id（无跨库 JOIN）
	// Step 1: query changed note res_ids from user_shares (no cross-DB JOIN)
	activeIDs, revokedIDs, err := s.repo.ListChangedNoteResIDs(ctx, uid, since)
	if err != nil {
		return nil, err
	}

	result := &dto.ShareChangesData{
		Added:    []string{},
		Removed:  []string{},
		LastTime: time.Now().UnixMilli(),
	}

	// 步骤2：合并 ID 列表，一次批量查询 notes，按 vault 过滤后按来源分拆
	// Step 2: merge ID lists, single batch query for notes, filter by vault and split by source
	allIDs := make([]int64, 0, len(activeIDs)+len(revokedIDs))
	allIDs = append(allIDs, activeIDs...)
	allIDs = append(allIDs, revokedIDs...)

	if len(allIDs) > 0 {
		activeSet := make(map[int64]struct{}, len(activeIDs))
		for _, id := range activeIDs {
			activeSet[id] = struct{}{}
		}

		notes, err := s.noteRepo.ListByIDs(ctx, allIDs, uid)
		if err == nil {
			for _, n := range notes {
				if n.VaultID != vault.ID {
					continue
				}
				if _, ok := activeSet[n.ID]; ok {
					if n.Action != domain.NoteActionDelete {
						result.Added = append(result.Added, n.Path)
					}
				} else {
					result.Removed = append(result.Removed, n.Path)
				}
			}
		}
	}

	return result, nil
}

// GetSharedNote retrieves specific shared note details
// GetSharedNote 获取分享的单条笔记详情
func (s *shareService) GetSharedNote(ctx context.Context, shareToken string, noteID int64, password string) (*dto.NoteDTO, error) {
	ridStr := strconv.FormatInt(noteID, 10)
	shareEntity, err := s.VerifyShare(ctx, shareToken, ridStr, "note", password)
	if err != nil {
		if cObj, ok := err.(*code.Code); ok {
			return nil, cObj
		}
		return nil, code.ErrorInvalidAuthToken.WithDetails(err.Error())
	}

	// Retrieve note directly via ID (using resource owner's UID)
	// 直接通过 ID 获取笔记 (使用资源所有者的 UID)
	note, err := s.noteRepo.GetByID(ctx, noteID, shareEntity.UID)
	if err != nil {
		return nil, code.ErrorNoteNotFound
	}

	noteDTO := &dto.NoteDTO{
		ID:               note.ID,
		Path:             note.Path,
		PathHash:         note.PathHash,
		Content:          note.Content,
		ContentHash:      note.ContentHash,
		Version:          note.Version,
		Ctime:            note.Ctime,
		Mtime:            note.Mtime,
		UpdatedTimestamp: note.UpdatedTimestamp,
		UpdatedAt:        timex.Time(note.UpdatedAt),
		CreatedAt:        timex.Time(note.CreatedAt),
	}

	fileRefs, err := s.resolveSharedNoteFiles(ctx, shareEntity.UID, note.VaultID, note.Path, noteDTO.Content)
	if err != nil {
		s.logger.Warn("GetSharedNote resolveSharedNoteFiles failed", zap.Error(err), zap.String("notePath", note.Path))
	}

	if len(fileRefs) > 0 {
		updatedResources, changed := mergeShareFileResources(shareEntity.Resources, fileRefs)
		if changed {
			if err := s.repo.UpdateResources(ctx, shareEntity.UID, shareEntity.SID, updatedResources); err != nil {
				s.logger.Warn("GetSharedNote UpdateResources failed", zap.Error(err), zap.Int64("shareID", shareEntity.SID))
			} else {
				shareEntity.Resources = updatedResources
			}
		}
	}

	// Handle Obsidian attachment embedded tags ![[...]]
	// 处理 Obsidian 附件嵌入标签 ![[...]]
	newContent := attachmentRegex.ReplaceAllStringFunc(noteDTO.Content, func(match string) string {
		submatches := attachmentRegex.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}

		inner := submatches[1]
		rawPath := inner
		options := ""

		// 提取资源路径（移除别名 | 和锚点 # 之后的部分）
		if idx := strings.IndexAny(inner, "|#"); idx != -1 {
			rawPath = inner[:idx]
			if inner[idx] == '|' {
				options = strings.TrimSpace(inner[idx+1:])
			}
		}
		rawPath = strings.TrimSpace(rawPath)
		if rawPath == "" {
			return match
		}

		file := fileRefs[rawPath]
		if file == nil {
			return match
		}

		apiUrl := buildSharedFileAPIURL(file.ID, shareToken, password)
		lowerPath := strings.ToLower(file.Path)
		ext := filepath.Ext(lowerPath)

		isImage := strings.Contains(".png.jpg.jpeg.gif.svg.webp.bmp", ext) && ext != ""
		isVideo := strings.Contains(".mp4.webm.ogg.mov", ext) && ext != ""
		isAudio := strings.Contains(".mp3.wav.ogg.m4a.flac", ext) && ext != ""

		if isImage {
			width := ""
			height := ""
			if options != "" {
				sizeRe := regexp.MustCompile(`^(\d+)(?:x(\d+))?`)
				sizeMatch := sizeRe.FindStringSubmatch(options)
				if len(sizeMatch) > 1 && sizeMatch[1] != "" {
					width = ` width="` + sizeMatch[1] + `"`
				}
				if len(sizeMatch) > 2 && sizeMatch[2] != "" {
					height = ` height="` + sizeMatch[2] + `"`
				}
			}
			return `<img src="` + apiUrl + `" alt="` + rawPath + `"` + width + height + ` />`
		} else if isVideo {
			return `<video src="` + apiUrl + `" controls style="max-width:100%"></video>`
		} else if isAudio {
			return `<audio src="` + apiUrl + `" controls></audio>`
		} else {
			return `<a href="` + apiUrl + `" target="_blank">📎 ` + rawPath + `</a>`
		}
	})
	newContent = rewriteMarkdownImageLinks(newContent, fileRefs, shareToken, password)
	newContent = rewriteHTMLImageSources(newContent, fileRefs, shareToken, password)
	noteDTO.Content = newContent

	return noteDTO, nil
}

func (s *shareService) resolveSharedNoteFiles(ctx context.Context, uid int64, vaultID int64, notePath string, content string) (map[string]*domain.File, error) {
	rawRefs := extractSharedNoteFileRefs(content)
	if len(rawRefs) == 0 {
		return map[string]*domain.File{}, nil
	}

	result := make(map[string]*domain.File, len(rawRefs))
	for _, rawRef := range rawRefs {
		file, err := s.resolveSharedFileReference(ctx, uid, vaultID, notePath, rawRef)
		if err != nil {
			return nil, err
		}
		if file != nil {
			result[rawRef] = file
		}
	}
	return result, nil
}

func (s *shareService) resolveSharedFileReference(ctx context.Context, uid int64, vaultID int64, notePath string, rawRef string) (*domain.File, error) {
	ref := strings.TrimSpace(rawRef)
	if !isLocalSharePath(ref) {
		return nil, nil
	}

	for _, candidate := range buildSharePathCandidates(notePath, ref) {
		file, err := s.fileRepo.GetByPath(ctx, candidate, vaultID, uid)
		if err == nil && file != nil && file.Action != domain.FileActionDelete {
			return file, nil
		}
	}

	normalizedRef := normalizeShareVaultPath(ref)
	if normalizedRef != "" && !strings.Contains(normalizedRef, "/") {
		file, err := s.fileRepo.GetByPathLike(ctx, normalizedRef, vaultID, uid)
		if err == nil && file != nil && file.Action != domain.FileActionDelete {
			return file, nil
		}
	}

	return nil, nil
}

func extractSharedNoteFileRefs(content string) []string {
	seen := make(map[string]struct{})
	refs := make([]string, 0)

	for _, match := range attachmentRegex.FindAllStringSubmatch(content, -1) {
		if len(match) < 2 {
			continue
		}
		ref := extractObsidianEmbedPath(match[1])
		if ref == "" || !isLocalSharePath(ref) {
			continue
		}
		if _, ok := seen[ref]; ok {
			continue
		}
		seen[ref] = struct{}{}
		refs = append(refs, ref)
	}

	for _, match := range markdownImageRegex.FindAllStringSubmatch(content, -1) {
		if len(match) < 3 {
			continue
		}
		ref, _, _ := parseMarkdownLinkTarget(match[2])
		ref = strings.TrimSpace(ref)
		if ref == "" || !isLocalSharePath(ref) {
			continue
		}
		if _, ok := seen[ref]; ok {
			continue
		}
		seen[ref] = struct{}{}
		refs = append(refs, ref)
	}

	for _, match := range htmlImageRegex.FindAllStringSubmatch(content, -1) {
		if len(match) < 4 {
			continue
		}
		ref := strings.TrimSpace(match[3])
		if ref == "" || !isLocalSharePath(ref) {
			continue
		}
		if _, ok := seen[ref]; ok {
			continue
		}
		seen[ref] = struct{}{}
		refs = append(refs, ref)
	}

	return refs
}

func extractObsidianEmbedPath(inner string) string {
	rawPath := inner
	if idx := strings.IndexAny(inner, "|#"); idx != -1 {
		rawPath = inner[:idx]
	}
	return strings.TrimSpace(rawPath)
}

func parseMarkdownLinkTarget(raw string) (target string, start int, end int) {
	start = 0
	for start < len(raw) {
		switch raw[start] {
		case ' ', '\t', '\n':
			start++
		default:
			goto targetStart
		}
	}
	return "", -1, -1

targetStart:
	if raw[start] == '<' {
		end = strings.IndexByte(raw[start+1:], '>')
		if end == -1 {
			return "", -1, -1
		}
		end += start + 2
		return raw[start+1 : end-1], start, end
	}

	end = len(raw)
	escaped := false
	for i := start; i < len(raw); i++ {
		if escaped {
			escaped = false
			continue
		}
		switch raw[i] {
		case '\\':
			escaped = true
		case ' ', '\t', '\n':
			end = i
			return raw[start:end], start, end
		}
	}

	return raw[start:end], start, end
}

func rewriteMarkdownImageLinks(content string, fileRefs map[string]*domain.File, shareToken string, password string) string {
	return markdownImageRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatches := markdownImageRegex.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}

		target, start, end := parseMarkdownLinkTarget(submatches[2])
		if target == "" || start < 0 || end < 0 {
			return match
		}

		file := fileRefs[strings.TrimSpace(target)]
		if file == nil {
			return match
		}

		replacementTarget := buildSharedFileAPIURL(file.ID, shareToken, password)
		rawTarget := submatches[2][start:end]
		if strings.HasPrefix(rawTarget, "<") && strings.HasSuffix(rawTarget, ">") {
			replacementTarget = "<" + replacementTarget + ">"
		}

		return "![" + submatches[1] + "](" + submatches[2][:start] + replacementTarget + submatches[2][end:] + ")"
	})
}

func rewriteHTMLImageSources(content string, fileRefs map[string]*domain.File, shareToken string, password string) string {
	return htmlImageRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatches := htmlImageRegex.FindStringSubmatch(match)
		if len(submatches) < 5 {
			return match
		}

		file := fileRefs[strings.TrimSpace(submatches[3])]
		if file == nil {
			return match
		}

		return "<img" + submatches[1] + "src=" + submatches[2] + buildSharedFileAPIURL(file.ID, shareToken, password) + submatches[2] + submatches[4] + ">"
	})
}

func buildSharedFileAPIURL(fileID int64, shareToken string, password string) string {
	apiURL := "/api/share/file?id=" + strconv.FormatInt(fileID, 10) + "&share_token=" + shareToken
	if password != "" {
		apiURL += "&password=" + password
	}
	return apiURL
}

func mergeShareFileResources(resources map[string][]string, fileRefs map[string]*domain.File) (map[string][]string, bool) {
	merged := cloneShareResources(resources)
	allowed := make(map[string]struct{}, len(merged["file"]))
	for _, id := range merged["file"] {
		allowed[id] = struct{}{}
	}

	changed := false
	for _, file := range fileRefs {
		id := strconv.FormatInt(file.ID, 10)
		if _, ok := allowed[id]; ok {
			continue
		}
		merged["file"] = append(merged["file"], id)
		allowed[id] = struct{}{}
		changed = true
	}

	if changed {
		sort.Strings(merged["file"])
	}

	return merged, changed
}

func cloneShareResources(resources map[string][]string) map[string][]string {
	cloned := make(map[string][]string, len(resources))
	for key, values := range resources {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
}

func buildSharePathCandidates(notePath string, rawRef string) []string {
	ref := strings.TrimSpace(strings.ReplaceAll(rawRef, "\\", "/"))
	if ref == "" || !isLocalSharePath(ref) {
		return nil
	}

	candidates := make([]string, 0, 2)
	addCandidate := func(candidate string) {
		candidate = normalizeShareVaultPath(candidate)
		if candidate == "" {
			return
		}
		for _, existing := range candidates {
			if existing == candidate {
				return
			}
		}
		candidates = append(candidates, candidate)
	}

	noteDir := normalizeShareVaultPath(path.Dir(strings.ReplaceAll(notePath, "\\", "/")))
	if noteDir == "." {
		noteDir = ""
	}
	if noteDir != "" {
		addCandidate(path.Join(noteDir, ref))
	} else {
		addCandidate(ref)
	}
	if strings.Contains(ref, "/") && !strings.HasPrefix(ref, "./") && !strings.HasPrefix(ref, "../") {
		addCandidate(ref)
	}

	return candidates
}

func normalizeShareVaultPath(p string) string {
	p = strings.TrimSpace(strings.ReplaceAll(p, "\\", "/"))
	p = strings.TrimPrefix(p, "./")
	p = strings.TrimPrefix(p, "/")
	if p == "" {
		return ""
	}

	cleaned := path.Clean(p)
	if cleaned == "." || cleaned == "/" || strings.HasPrefix(cleaned, "../") {
		return ""
	}
	return strings.TrimPrefix(cleaned, "/")
}

func isLocalSharePath(ref string) bool {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return false
	}

	lowerRef := strings.ToLower(ref)
	switch {
	case strings.HasPrefix(ref, "#"):
		return false
	case strings.HasPrefix(ref, "/"):
		return false
	case strings.HasPrefix(lowerRef, "http://"):
		return false
	case strings.HasPrefix(lowerRef, "https://"):
		return false
	case strings.HasPrefix(lowerRef, "//"):
		return false
	case strings.HasPrefix(lowerRef, "data:"):
		return false
	case strings.HasPrefix(lowerRef, "mailto:"):
		return false
	case strings.HasPrefix(lowerRef, "tel:"):
		return false
	case strings.HasPrefix(lowerRef, "obsidian://"):
		return false
	default:
		return true
	}
}

// GetSharedFile retrieves shared file content
// GetSharedFile 获取分享的文件内容
func (s *shareService) GetSharedFile(ctx context.Context, shareToken string, fileID int64, password string) (content []byte, contentType string, mtime int64, etag string, fileName string, err error) {
	ridStr := strconv.FormatInt(fileID, 10)
	shareEntity, err := s.VerifyShare(ctx, shareToken, ridStr, "file", password)
	if err != nil {
		if cObj, ok := err.(*code.Code); ok {
			return nil, "", 0, "", "", cObj
		}
		return nil, "", 0, "", "", code.ErrorInvalidAuthToken.WithDetails(err.Error())
	}

	// 1. Get resource owner's UID
	// 1. 获取资源所有者的 UID
	ownerUID := shareEntity.UID

	// 2. Confirm path hash (get file metadata from fileRepo)
	// 2. 确认路径哈希 (从 fileRepo 获取文件元数据)
	file, err := s.fileRepo.GetByID(ctx, fileID, ownerUID)
	if err != nil {
		return nil, "", 0, "", "", code.ErrorFileNotFound
	}

	if file.Action == domain.FileActionDelete {
		return nil, "", 0, "", "", code.ErrorFileNotFound
	}

	// Read physical file content
	// 读取物理文件内容
	content, err = os.ReadFile(file.SavePath)
	if err != nil {
		return nil, "", 0, "", "", code.ErrorFileReadFailed.WithDetails(err.Error())
	}

	// Identify file MIME type
	// 识别文件 MIME 类型
	ext := filepath.Ext(file.Path)
	contentType = mime.TypeByExtension(ext)
	if contentType == "" {
		// If extension cannot be identified, perform content sniffing
		// 如果扩展名识别不到, 进行内容嗅探
		contentType = http.DetectContentType(content)
	}

	// Compute etag in real-time
	// 实时计算 etag
	etag = util.EncodeHash32(string(content))

	return content, contentType, file.Mtime, etag, file.Path, nil

}

// GetSharedFileInfo retrieves shared file metadata and path for zero-copy download
// GetSharedFileInfo 获取分享文件的元数据和路径，用于零拷贝下载
func (s *shareService) GetSharedFileInfo(ctx context.Context, shareToken string, fileID int64, password string) (savePath string, contentType string, mtime int64, etag string, fileName string, err error) {
	ridStr := strconv.FormatInt(fileID, 10)
	shareEntity, err := s.VerifyShare(ctx, shareToken, ridStr, "file", password)
	if err != nil {
		if cObj, ok := err.(*code.Code); ok {
			return "", "", 0, "", "", cObj
		}
		return "", "", 0, "", "", code.ErrorInvalidAuthToken.WithDetails(err.Error())
	}

	// 1. Get resource owner's UID
	ownerUID := shareEntity.UID

	// 2. Confirm path hash (get file metadata from fileRepo)
	file, err := s.fileRepo.GetByID(ctx, fileID, ownerUID)
	if err != nil {
		return "", "", 0, "", "", code.ErrorFileNotFound
	}

	if file.Action == domain.FileActionDelete {
		return "", "", 0, "", "", code.ErrorFileNotFound
	}

	// Identify file MIME type
	ext := filepath.Ext(file.Path)
	contentType = mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Use file's content hash as ETag
	etag = file.ContentHash
	if etag == "" {
		etag = file.PathHash
	}

	return file.SavePath, contentType, file.Mtime, etag, filepath.Base(file.Path), nil
}

// Shutdown shuts down the service and flushes remaining data
// Shutdown 关闭服务并同步最后的数据
func (s *shareService) Shutdown(ctx context.Context) error {
	s.ticker.Stop()
	close(s.stopCh)

	// Wait for periodic synchronization goroutine to end (i.e., last flush completed)
	// 等待定时同步协程结束（即最后一次 flush 完成）
	select {
	case <-s.doneCh:
		s.logger.Info("ShareService background flush loop stopped")
		return nil
	case <-ctx.Done():
		s.logger.Warn("ShareService shutdown timeout, some data might not be flushed")
		return ctx.Err()
	}
}
