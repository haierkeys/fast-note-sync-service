package service

import (
	"context"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"go.uber.org/zap"
)

var (
	attachmentRegex = regexp.MustCompile(`!\[\[(.*?)\]\]`)
)

type ShareService interface {
	ShareGenerate(ctx context.Context, uid int64, vaultName string, path string, pathHash string) (*dto.ShareCreateResponse, error)
	VerifyShare(ctx context.Context, token string, rid string, rtp string) (*pkgapp.ShareEntity, error)
	GetSharedNote(ctx context.Context, shareToken string, noteID int64) (*dto.NoteDTO, error)
	GetSharedFile(ctx context.Context, shareToken string, fileID int64) (content []byte, contentType string, mtime int64, etag string, fileName string, err error)
	RecordView(uid int64, id int64)
	StopShare(ctx context.Context, uid int64, id int64) error
	ListShares(ctx context.Context, uid int64) ([]*domain.UserShare, error)
	Shutdown(ctx context.Context) error
}

type aggStats struct {
	uid          int64
	viewCount    int64
	lastViewedAt time.Time
}

type shareService struct {
	repo         domain.UserShareRepository
	tokenManager pkgapp.TokenManager
	noteRepo     domain.NoteRepository
	fileRepo     domain.FileRepository
	vaultRepo    domain.VaultRepository
	logger       *zap.Logger
	config       *ServiceConfig

	// 统计缓冲区
	bufferMu    sync.Mutex
	statsBuffer map[int64]*aggStats
	ticker      *time.Ticker
	stopCh      chan struct{}
	doneCh      chan struct{}
}

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

// ShareGenerate 生成并存储分享 Token
func (s *shareService) ShareGenerate(ctx context.Context, uid int64, vaultName string, path string, pathHash string) (*dto.ShareCreateResponse, error) {
	// 1. 获取 VaultID
	vault, err := s.vaultRepo.GetByName(ctx, vaultName, uid)
	if err != nil {
		return nil, err
	}
	vaultID := vault.ID

	var resolvedResources = make(map[string][]string)
	var mainID int64
	var mainType string

	// 2. 根据后缀判定类型
	isNote := strings.HasSuffix(strings.ToLower(path), ".md")

	if isNote {
		// 尝试作为 Note 查找
		note, err := s.noteRepo.GetByPathHash(ctx, pathHash, vaultID, uid)
		if err == nil && note != nil && note.Action != domain.NoteActionDelete {
			mainID = note.ID
			mainType = "note"
			noteIDStr := strconv.FormatInt(note.ID, 10)
			resolvedResources["note"] = []string{noteIDStr}

			// 解析内容中的附件 ![[附件路径]]
			matches := attachmentRegex.FindAllStringSubmatch(note.Content, -1)
			for _, match := range matches {
				if len(match) > 1 {
					inner := match[1]
					// 提取资源路径（移除别名 | 和锚点 # 之后的部分）
					attPath := inner
					if idx := strings.IndexAny(inner, "|#"); idx != -1 {
						attPath = inner[:idx]
					}
					attPath = strings.TrimSpace(attPath)
					if attPath == "" {
						continue
					}

					var file *domain.File
					var ferr error

					// 策略 1: 尝试精确匹配（完整路径哈希）
					attHash := util.EncodeHash32(attPath)
					file, ferr = s.fileRepo.GetByPathHash(ctx, attHash, vaultID, uid)

					// 策略 2: 尝试后缀匹配（处理 Obsidian 简写路径）
					if (ferr != nil || file == nil) && !strings.Contains(attPath, "/") {
						file, ferr = s.fileRepo.GetByPathLike(ctx, attPath, vaultID, uid)
					}

					if ferr == nil && file != nil && file.Action != domain.FileActionDelete {
						fileIDStr := strconv.FormatInt(file.ID, 10)
						// 避免重复授权
						if !util.Inarray(resolvedResources["file"], fileIDStr) {
							resolvedResources["file"] = append(resolvedResources["file"], fileIDStr)
						}
					}
				}
			}
		} else {
			return nil, code.ErrorNoteNotFound.WithDetails("note not found: " + path)
		}
	} else {
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

	// 3. 确定过期时间
	expiry := 30 * 24 * time.Hour // 默认 30 天
	if s.config != nil && s.config.App.ShareTokenExpiry != "" {
		if d, err := util.ParseDuration(s.config.App.ShareTokenExpiry); err == nil {
			expiry = d
		}
	}
	expiresAt := time.Now().Add(expiry)

	share := &domain.UserShare{
		UID:       uid,
		Resources: resolvedResources,
		Status:    1,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, uid, share); err != nil {
		return nil, err
	}

	// 4. 生成 Token (使用底层 SID 加密方案)
	token, err := s.tokenManager.ShareGenerate(share.ID, uid, resolvedResources)
	if err != nil {
		return nil, err
	}

	return &dto.ShareCreateResponse{
		ID:        mainID,
		Type:      mainType,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// VerifyShare 验证分享 Token 及其状态
func (s *shareService) VerifyShare(ctx context.Context, token string, rid string, rtp string) (*pkgapp.ShareEntity, error) {
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

	entity.Resources = share.Resources

	ids, ok := share.Resources[rtp]
	if !ok {
		return nil, domain.ErrShareCancelled // 资源类型不匹配
	}

	authorized := false
	for _, id := range ids {
		if id == rid {
			authorized = true
			break
		}
	}

	if !authorized {
		return nil, domain.ErrShareCancelled // 资源未授权
	}

	// 内存记录访问统计 (延迟 5 分钟更新)
	s.RecordView(share.UID, share.ID)

	return entity, nil
}

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

// StopShare 撤销分享
func (s *shareService) StopShare(ctx context.Context, uid int64, id int64) error {
	return s.repo.UpdateStatus(ctx, uid, id, 2)
}

// ListShares 列出用户的所有分享
func (s *shareService) ListShares(ctx context.Context, uid int64) ([]*domain.UserShare, error) {
	return s.repo.ListByUID(ctx, uid)
}

// GetSharedNote 获取分享的单条笔记详情
func (s *shareService) GetSharedNote(ctx context.Context, shareToken string, noteID int64) (*dto.NoteDTO, error) {
	ridStr := strconv.FormatInt(noteID, 10)
	shareEntity, err := s.VerifyShare(ctx, shareToken, ridStr, "note")
	if err != nil {
		return nil, code.ErrorInvalidAuthToken
	}

	// 直接通过 ID 获取笔记 (使用资源所有者的 UID)
	note, err := s.noteRepo.GetByID(ctx, noteID, shareEntity.UID)
	if err != nil {
		return nil, code.ErrorNoteNotFound
	}

	noteDTO := &dto.NoteDTO{
		ID:               note.ID,
		Path:             note.Path,
		Content:          note.Content,
		ContentHash:      note.ContentHash,
		Version:          note.Version,
		Ctime:            note.Ctime,
		Mtime:            note.Mtime,
		UpdatedTimestamp: note.UpdatedTimestamp,
		UpdatedAt:        timex.Time(note.UpdatedAt),
		CreatedAt:        timex.Time(note.CreatedAt),
	}

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

		// 查找文件 ID
		file, err := s.fileRepo.GetByPathLike(ctx, rawPath, note.VaultID, shareEntity.UID)
		if err != nil || file == nil {
			return match
		}

		apiUrl := "/api/share/file?id=" + strconv.FormatInt(file.ID, 10) + "&share_token=" + shareToken
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
	noteDTO.Content = newContent

	return noteDTO, nil
}

// GetSharedFile 获取分享的文件内容
func (s *shareService) GetSharedFile(ctx context.Context, shareToken string, fileID int64) (content []byte, contentType string, mtime int64, etag string, fileName string, err error) {
	ridStr := strconv.FormatInt(fileID, 10)
	shareEntity, err := s.VerifyShare(ctx, shareToken, ridStr, "file")
	if err != nil {
		return nil, "", 0, "", "", code.ErrorInvalidAuthToken
	}

	// 1. 获取资源所有者的 UID
	ownerUID := shareEntity.UID

	// 2. 确认路径哈希 (从 fileRepo 获取文件元数据)
	file, err := s.fileRepo.GetByID(ctx, fileID, ownerUID)
	if err != nil {
		return nil, "", 0, "", "", code.ErrorFileNotFound
	}

	if file.Action == domain.FileActionDelete {
		return nil, "", 0, "", "", code.ErrorFileNotFound
	}

	// 读取物理文件内容
	content, err = os.ReadFile(file.SavePath)
	if err != nil {
		return nil, "", 0, "", "", code.ErrorFileReadFailed.WithDetails(err.Error())
	}

	// 识别文件 MIME 类型
	ext := filepath.Ext(file.Path)
	contentType = mime.TypeByExtension(ext)
	if contentType == "" {
		// 如果扩展名识别不到, 进行内容嗅探
		contentType = http.DetectContentType(content)
	}

	// 实时计算 etag
	etag = util.EncodeHash32(string(content))

	return content, contentType, file.Mtime, etag, file.Path, nil

}

// Shutdown 关闭服务并同步最后的数据
func (s *shareService) Shutdown(ctx context.Context) error {
	s.ticker.Stop()
	close(s.stopCh)

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
