// Package dao 实现数据访问层
package dao

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/logger"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// noteHistoryRepository 实现 domain.NoteHistoryRepository 接口
type noteHistoryRepository struct {
	dao             *Dao
	customPrefixKey string
}

// NewNoteHistoryRepository 创建 NoteHistoryRepository 实例
func NewNoteHistoryRepository(dao *Dao) domain.NoteHistoryRepository {
	return &noteHistoryRepository{dao: dao, customPrefixKey: "user_note_history_"}
}

func (r *noteHistoryRepository) GetKey(uid int64) string {
	return r.customPrefixKey + strconv.FormatInt(uid, 10)
}

// noteHistory 获取笔记历史查询对象
func (r *noteHistoryRepository) noteHistory(uid int64) *query.Query {
	return r.dao.UseQueryWithOnceFunc(func(g *gorm.DB) {
		model.AutoMigrate(g, "NoteHistory")
	}, r.GetKey(uid)+"#noteHistory", r.GetKey(uid))
}

// toDomain 将数据库模型转换为领域模型
func (r *noteHistoryRepository) toDomain(m *model.NoteHistory, uid int64) (*domain.NoteHistory, error) {
	if m == nil {
		return nil, nil
	}
	h := &domain.NoteHistory{
		ID:          m.ID,
		NoteID:      m.NoteID,
		VaultID:     m.VaultID,
		Path:        m.Path,
		DiffPatch:   m.DiffPatch,
		Content:     m.Content,
		ContentHash: m.ContentHash,
		ClientName:  m.ClientName,
		Version:     m.Version,
		CreatedAt:   time.Time(m.CreatedAt),
		UpdatedAt:   time.Time(m.UpdatedAt),
	}
	if err := r.fillHistoryContent(uid, h); err != nil {
		return nil, err
	}
	return h, nil
}

// fillHistoryContent 填充历史记录内容及补丁
func (r *noteHistoryRepository) fillHistoryContent(uid int64, h *domain.NoteHistory) error {
	if h == nil {
		return nil
	}
	folder := r.dao.GetNoteHistoryFolderPath(uid, h.ID)

	// 加载补丁
	patch, exists, err := r.dao.LoadContentFromFile(folder, "diff.patch")
	if err != nil {
		return err
	}
	if exists {
		h.DiffPatch = patch
	} else if h.DiffPatch != "" {
		if err := r.dao.SaveContentToFile(folder, "diff.patch", h.DiffPatch); err != nil {
			r.dao.Logger().Warn("lazy migration: SaveContentToFile failed for history diff patch",
				zap.Int64(logger.FieldUID, uid),
				zap.Int64("historyId", h.ID),
				zap.String(logger.FieldMethod, "noteHistoryRepository.fillHistoryContent"),
				zap.Error(err),
			)
		}
	} else {
		return fmt.Errorf("history diff patch file not found: %w", os.ErrNotExist)
	}

	// 加载内容
	content, exists, err := r.dao.LoadContentFromFile(folder, "content.txt")
	if err != nil {
		return err
	}
	if exists {
		h.Content = content
	} else if h.Content != "" {
		if err := r.dao.SaveContentToFile(folder, "content.txt", h.Content); err != nil {
			r.dao.Logger().Warn("lazy migration: SaveContentToFile failed for history content",
				zap.Int64(logger.FieldUID, uid),
				zap.Int64("historyId", h.ID),
				zap.String(logger.FieldMethod, "noteHistoryRepository.fillHistoryContent"),
				zap.Error(err),
			)
		}
	} else {
		return fmt.Errorf("history content file not found: %w", os.ErrNotExist)
	}
	return nil
}

// GetByID 根据ID获取历史记录
func (r *noteHistoryRepository) GetByID(ctx context.Context, id, uid int64) (*domain.NoteHistory, error) {
	u := r.noteHistory(uid).NoteHistory
	m, err := u.WithContext(ctx).Where(u.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m, uid)
}

// GetByNoteIDAndHash 根据笔记ID和内容哈希获取历史记录
func (r *noteHistoryRepository) GetByNoteIDAndHash(ctx context.Context, noteID int64, contentHash string, uid int64) (*domain.NoteHistory, error) {
	u := r.noteHistory(uid).NoteHistory
	m, err := u.WithContext(ctx).Where(u.NoteID.Eq(noteID), u.ContentHash.Eq(contentHash)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m, uid)
}

// Create 创建历史记录
func (r *noteHistoryRepository) Create(ctx context.Context, history *domain.NoteHistory, uid int64) (*domain.NoteHistory, error) {
	var result *domain.NoteHistory
	var createErr error

	err := r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.noteHistory(uid).NoteHistory
		m := &model.NoteHistory{
			NoteID:      history.NoteID,
			VaultID:     history.VaultID,
			Path:        history.Path,
			ContentHash: history.ContentHash,
			ClientName:  history.ClientName,
			Version:     history.Version,
			CreatedAt:   timex.Time(history.CreatedAt),
			UpdatedAt:   timex.Time(history.UpdatedAt),
		}

		// 暂存内容用于写文件
		diffPatch := history.DiffPatch
		content := history.Content

		// 不在数据库中保存大数据
		m.DiffPatch = ""
		m.Content = ""

		createErr = u.WithContext(ctx).Create(m)
		if createErr != nil {
			return createErr
		}

		// 保存到文件
		folder := r.dao.GetNoteHistoryFolderPath(uid, m.ID)
		if err := r.dao.SaveContentToFile(folder, "diff.patch", diffPatch); err != nil {
			return err
		}
		if err := r.dao.SaveContentToFile(folder, "content.txt", content); err != nil {
			return err
		}

		hRes, err := r.toDomain(m, uid)
		if err != nil {
			return err
		}
		result = hRes
		result.DiffPatch = diffPatch
		result.Content = content
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// ListByNoteID 根据笔记ID获取历史记录列表
func (r *noteHistoryRepository) ListByNoteID(ctx context.Context, noteID int64, page, pageSize int, uid int64) ([]*domain.NoteHistory, int64, error) {
	u := r.noteHistory(uid).NoteHistory
	q := u.WithContext(ctx).Where(u.NoteID.Eq(noteID))

	count, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	modelList, err := q.Order(u.Version.Desc()).
		Limit(pageSize).
		Offset(app.GetPageOffset(page, pageSize)).
		Find()
	if err != nil {
		return nil, 0, err
	}

	var results []*domain.NoteHistory
	for _, m := range modelList {
		h, err := r.toDomain(m, uid)
		if err != nil {
			return nil, 0, err
		}
		results = append(results, h)
	}
	return results, count, nil
}

// GetLatestVersion 获取笔记的最新版本号
func (r *noteHistoryRepository) GetLatestVersion(ctx context.Context, noteID, uid int64) (int64, error) {
	u := r.noteHistory(uid).NoteHistory
	m, err := u.WithContext(ctx).Where(u.NoteID.Eq(noteID)).Order(u.Version.Desc()).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return m.Version, nil
}

// Migrate 迁移历史记录（更新 NoteID）
func (r *noteHistoryRepository) Migrate(ctx context.Context, oldNoteID, newNoteID, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.noteHistory(uid).NoteHistory
		_, err := u.WithContext(ctx).Where(u.NoteID.Eq(oldNoteID)).Update(u.NoteID, newNoteID)
		return err
	})
}

// GetNoteIDsWithOldHistory 获取有旧历史记录的笔记ID列表
func (r *noteHistoryRepository) GetNoteIDsWithOldHistory(ctx context.Context, cutoffTime int64, uid int64) ([]int64, error) {
	u := r.noteHistory(uid).NoteHistory
	cutoffTimeValue := timex.Time(time.UnixMilli(cutoffTime))
	var noteIDs []int64
	err := u.WithContext(ctx).
		Where(u.CreatedAt.Lt(cutoffTimeValue)).
		Distinct(u.NoteID).
		Pluck(u.NoteID, &noteIDs)
	if err != nil {
		return nil, err
	}
	return noteIDs, nil
}

// DeleteOldVersions 删除旧版本历史记录，保留最近 N 个版本
func (r *noteHistoryRepository) DeleteOldVersions(ctx context.Context, noteID int64, cutoffTime int64, keepVersions int, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.noteHistory(uid).NoteHistory

		// 先获取需要保留的最近 N 个版本的最小版本号
		var minKeepVersion int64 = 0
		if keepVersions > 0 {
			histories, err := u.WithContext(ctx).
				Where(u.NoteID.Eq(noteID)).
				Order(u.Version.Desc()).
				Limit(keepVersions).
				Find()
			if err != nil {
				return err
			}
			if len(histories) > 0 {
				minKeepVersion = histories[len(histories)-1].Version
			}
		}

		cutoffTimeValue := timex.Time(time.UnixMilli(cutoffTime))

		// 查询需要删除的历史记录ID
		var toDeleteIDs []int64
		q := u.WithContext(ctx).
			Where(u.NoteID.Eq(noteID)).
			Where(u.CreatedAt.Lt(cutoffTimeValue))

		if minKeepVersion > 0 {
			q = q.Where(u.Version.Lt(minKeepVersion))
		}

		histories, err := q.Find()
		if err != nil {
			return err
		}

		for _, h := range histories {
			toDeleteIDs = append(toDeleteIDs, h.ID)
		}

		if len(toDeleteIDs) == 0 {
			return nil
		}

		// 删除数据库记录
		_, err = u.WithContext(ctx).Where(u.ID.In(toDeleteIDs...)).Delete()
		if err != nil {
			return err
		}

		// 删除关联的文件
		for _, id := range toDeleteIDs {
			folder := r.dao.GetNoteHistoryFolderPath(uid, id)
			if err := r.dao.RemoveContentFolder(folder); err != nil {
				r.dao.Logger().Warn("failed to delete history folder",
					zap.Int64(logger.FieldUID, uid),
					zap.Int64("historyId", id),
					zap.String("folder", folder),
					zap.Error(err),
				)
			}
		}

		return nil
	})
}

// Delete 删除指定ID的历史记录
func (r *noteHistoryRepository) Delete(ctx context.Context, id, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.noteHistory(uid).NoteHistory

		// 删除数据库记录
		_, err := u.WithContext(ctx).Where(u.ID.Eq(id)).Delete()
		if err != nil {
			return err
		}

		// 删除关联的文件
		folder := r.dao.GetNoteHistoryFolderPath(uid, id)
		if err := r.dao.RemoveContentFolder(folder); err != nil {
			r.dao.Logger().Warn("failed to delete history folder",
				zap.Int64(logger.FieldUID, uid),
				zap.Int64("historyId", id),
				zap.String("folder", folder),
				zap.Error(err),
			)
		}

		return nil
	})
}

// 确保 noteHistoryRepository 实现了 domain.NoteHistoryRepository 接口
var _ domain.NoteHistoryRepository = (*noteHistoryRepository)(nil)
